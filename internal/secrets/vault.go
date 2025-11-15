package secrets

import (
	"context"
	"fmt"
	"os"

	vault "github.com/hashicorp/vault/api"
	"github.com/geekxflood/common/logging"
	"github.com/geekxflood/gxf-discord-bot/internal/config"
)

// Provider defines the interface for secret providers
type Provider interface {
	GetSecret(ctx context.Context, path string) (map[string]interface{}, error)
	GetSecretValue(ctx context.Context, path, key string) (string, error)
	Close() error
}

// VaultProvider implements secret provider for Vault/OpenBao
type VaultProvider struct {
	client    *vault.Client
	cfg       config.Provider
	logger    logging.Logger
	mountPath string
}

// NewVaultProvider creates a new Vault/OpenBao secret provider
func NewVaultProvider(ctx context.Context, cfg config.Provider, logger logging.Logger) (*VaultProvider, error) {
	if !cfg.Exists("secrets") {
		return nil, fmt.Errorf("secrets configuration not found")
	}

	address, err := cfg.GetString("secrets.address", "")
	if err != nil || address == "" {
		return nil, fmt.Errorf("vault address is required")
	}

	// Create Vault config
	vaultCfg := vault.DefaultConfig()
	vaultCfg.Address = address

	// Configure TLS
	tlsVerify, _ := cfg.GetBool("secrets.tlsVerify", true)
	if !tlsVerify {
		tlsConfig := &vault.TLSConfig{
			Insecure: true,
		}
		if err := vaultCfg.ConfigureTLS(tlsConfig); err != nil {
			return nil, fmt.Errorf("failed to configure TLS: %w", err)
		}
	} else if cfg.Exists("secrets.caCert") {
		caCert, _ := cfg.GetString("secrets.caCert", "")
		tlsConfig := &vault.TLSConfig{
			CACert: caCert,
		}
		if err := vaultCfg.ConfigureTLS(tlsConfig); err != nil {
			return nil, fmt.Errorf("failed to configure TLS: %w", err)
		}
	}

	// Create client
	client, err := vault.NewClient(vaultCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	provider := &VaultProvider{
		client: client,
		cfg:    cfg,
		logger: logger,
	}

	// Get mount path
	provider.mountPath, _ = cfg.GetString("secrets.mountPath", "secret")

	// Authenticate
	if err := provider.authenticate(ctx); err != nil {
		return nil, fmt.Errorf("failed to authenticate with vault: %w", err)
	}

	logger.Info("Vault provider initialized", "address", address)

	return provider, nil
}

// authenticate authenticates with Vault using configured method
func (v *VaultProvider) authenticate(ctx context.Context) error {
	authMethod, _ := v.cfg.GetString("secrets.authMethod", "token")

	switch authMethod {
	case "token":
		return v.authenticateToken()
	case "approle":
		return v.authenticateAppRole(ctx)
	case "kubernetes":
		return v.authenticateKubernetes(ctx)
	default:
		return fmt.Errorf("unsupported auth method: %s", authMethod)
	}
}

// authenticateToken authenticates using a token
func (v *VaultProvider) authenticateToken() error {
	var token string

	// Try to get token from environment variable
	if v.cfg.Exists("secrets.tokenEnvVar") {
		tokenEnv, _ := v.cfg.GetString("secrets.tokenEnvVar", "")
		token = os.Getenv(tokenEnv)
	}

	// Fall back to direct token
	if token == "" && v.cfg.Exists("secrets.token") {
		token, _ = v.cfg.GetString("secrets.token", "")
	}

	if token == "" {
		return fmt.Errorf("vault token not found")
	}

	v.client.SetToken(token)
	v.logger.Debug("Authenticated with vault using token")
	return nil
}

// authenticateAppRole authenticates using AppRole
func (v *VaultProvider) authenticateAppRole(ctx context.Context) error {
	if !v.cfg.Exists("secrets.appRole") {
		return fmt.Errorf("appRole configuration not found")
	}

	roleIDMap, err := v.cfg.GetMap("secrets.appRole")
	if err != nil {
		return fmt.Errorf("failed to get appRole config: %w", err)
	}

	roleID, ok := roleIDMap["roleId"].(string)
	if !ok || roleID == "" {
		return fmt.Errorf("roleId is required for appRole auth")
	}

	secretID, ok := roleIDMap["secretId"].(string)
	if !ok || secretID == "" {
		return fmt.Errorf("secretId is required for appRole auth")
	}

	// Authenticate
	data := map[string]interface{}{
		"role_id":   roleID,
		"secret_id": secretID,
	}

	secret, err := v.client.Logical().WriteWithContext(ctx, "auth/approle/login", data)
	if err != nil {
		return fmt.Errorf("appRole login failed: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("appRole login returned no auth data")
	}

	v.client.SetToken(secret.Auth.ClientToken)
	v.logger.Info("Authenticated with vault using appRole")
	return nil
}

// authenticateKubernetes authenticates using Kubernetes service account
func (v *VaultProvider) authenticateKubernetes(ctx context.Context) error {
	if !v.cfg.Exists("secrets.kubernetes") {
		return fmt.Errorf("kubernetes configuration not found")
	}

	k8sMap, err := v.cfg.GetMap("secrets.kubernetes")
	if err != nil {
		return fmt.Errorf("failed to get kubernetes config: %w", err)
	}

	role, ok := k8sMap["role"].(string)
	if !ok || role == "" {
		return fmt.Errorf("role is required for kubernetes auth")
	}

	// Get JWT token from service account
	saTokenPath := "/var/run/secrets/kubernetes.io/serviceaccount/token" //nolint:gosec // This is the standard Kubernetes SA path
	if path, ok := k8sMap["serviceAccount"].(string); ok && path != "" {
		saTokenPath = path
	}

	jwt, err := os.ReadFile(saTokenPath)
	if err != nil {
		return fmt.Errorf("failed to read service account token: %w", err)
	}

	// Authenticate
	data := map[string]interface{}{
		"role": role,
		"jwt":  string(jwt),
	}

	secret, err := v.client.Logical().WriteWithContext(ctx, "auth/kubernetes/login", data)
	if err != nil {
		return fmt.Errorf("kubernetes login failed: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("kubernetes login returned no auth data")
	}

	v.client.SetToken(secret.Auth.ClientToken)
	v.logger.Info("Authenticated with vault using kubernetes")
	return nil
}

// GetSecret retrieves a secret from Vault
func (v *VaultProvider) GetSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	// Construct full path (KV v2 uses /data/ in the path)
	fullPath := fmt.Sprintf("%s/data/%s", v.mountPath, path)

	secret, err := v.client.Logical().ReadWithContext(ctx, fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret: %w", err)
	}

	if secret == nil {
		return nil, fmt.Errorf("secret not found: %s", path)
	}

	// KV v2 stores data under "data" key
	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		// Try KV v1 format
		return secret.Data, nil
	}

	return data, nil
}

// GetSecretValue retrieves a specific value from a secret
func (v *VaultProvider) GetSecretValue(ctx context.Context, path, key string) (string, error) {
	data, err := v.GetSecret(ctx, path)
	if err != nil {
		return "", err
	}

	value, ok := data[key]
	if !ok {
		return "", fmt.Errorf("key %s not found in secret %s", key, path)
	}

	strValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("value at %s.%s is not a string", path, key)
	}

	return strValue, nil
}

// Close closes the Vault client
func (v *VaultProvider) Close() error {
	// Vault client doesn't need explicit cleanup
	return nil
}
