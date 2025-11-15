package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/geekxflood/common/logging"
	"github.com/geekxflood/gxf-discord-bot/internal/config"
	"github.com/geekxflood/gxf-discord-bot/internal/secrets"
	"golang.org/x/oauth2"
)

// Manager handles OAuth authentication
type Manager struct {
	cfg            config.Provider
	secretsMgr     *secrets.Manager
	logger         logging.Logger
	oauthConfig    *oauth2.Config
	sessions       map[string]*Session
	sessionsMux    sync.RWMutex
	server         *http.Server
	enabled        bool
	authorizedUsers []string
	authorizedRoles []string
}

// Session represents an authenticated session
type Session struct {
	UserID    string
	Username  string
	Token     *oauth2.Token
	ExpiresAt time.Time
}

// NewManager creates a new auth manager
func NewManager(ctx context.Context, cfg config.Provider, secretsMgr *secrets.Manager, logger logging.Logger) (*Manager, error) {
	mgr := &Manager{
		cfg:        cfg,
		secretsMgr: secretsMgr,
		logger:     logger,
		sessions:   make(map[string]*Session),
	}

	// Check if auth is enabled
	if !cfg.Exists("auth") {
		mgr.enabled = false
		logger.Info("Authentication not configured")
		return mgr, nil
	}

	enabled, _ := cfg.GetBool("auth.enabled", false)
	if !enabled {
		mgr.enabled = false
		logger.Info("Authentication disabled")
		return mgr, nil
	}

	mgr.enabled = true

	// Get authorized users and roles
	if cfg.Exists("auth.authorizedUsers") {
		users, _ := cfg.GetStringSlice("auth.authorizedUsers")
		mgr.authorizedUsers = users
	}

	if cfg.Exists("auth.authorizedRoles") {
		roles, _ := cfg.GetStringSlice("auth.authorizedRoles")
		mgr.authorizedRoles = roles
	}

	// Initialize OAuth config
	if err := mgr.initOAuthConfig(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize OAuth: %w", err)
	}

	// Start callback server
	if err := mgr.startCallbackServer(); err != nil {
		return nil, fmt.Errorf("failed to start callback server: %w", err)
	}

	logger.Info("Authentication manager initialized")
	return mgr, nil
}

// initOAuthConfig initializes the OAuth configuration
func (m *Manager) initOAuthConfig(ctx context.Context) error {
	provider, _ := m.cfg.GetString("auth.provider", "discord")
	clientID, _ := m.cfg.GetString("auth.clientId", "")
	redirectURL, _ := m.cfg.GetString("auth.redirectUrl", "")
	scopes, _ := m.cfg.GetStringSlice("auth.scopes")

	if clientID == "" || redirectURL == "" {
		return fmt.Errorf("clientId and redirectUrl are required")
	}

	// Get client secret
	clientSecret, err := m.secretsMgr.GetOAuthClientSecret(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client secret: %w", err)
	}

	// Get OAuth endpoints
	endpoint := oauth2.Endpoint{}

	switch provider {
	case "discord":
		endpoint = oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		}
		if len(scopes) == 0 {
			scopes = []string{"identify"}
		}

	case "google":
		endpoint = oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://accounts.google.com/o/oauth2/token",
		}

	case "github":
		endpoint = oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		}

	case "custom":
		if !m.cfg.Exists("auth.endpoints") {
			return fmt.Errorf("custom provider requires endpoints configuration")
		}
		endpointsMap, _ := m.cfg.GetMap("auth.endpoints")
		authURL, _ := endpointsMap["authUrl"].(string)
		tokenURL, _ := endpointsMap["tokenUrl"].(string)
		if authURL == "" || tokenURL == "" {
			return fmt.Errorf("authUrl and tokenUrl are required for custom provider")
		}
		endpoint = oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		}

	default:
		return fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	m.oauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     endpoint,
	}

	return nil
}

// startCallbackServer starts the HTTP server for OAuth callbacks
func (m *Manager) startCallbackServer() error {
	host, _ := m.cfg.GetString("auth.callbackServer.host", "localhost")
	port, _ := m.cfg.GetInt("auth.callbackServer.port", 8080)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", m.handleCallback)
	mux.HandleFunc("/health", m.handleHealth)

	m.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: mux,
	}

	go func() {
		m.logger.Info("OAuth callback server starting", "addr", m.server.Addr)
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			m.logger.Error("Callback server error", "error", err)
		}
	}()

	return nil
}

// handleCallback handles OAuth callback
func (m *Manager) handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		http.Error(w, "Authorization code not found", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := m.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		m.logger.Error("Failed to exchange code for token", "error", err)
		http.Error(w, "Failed to authenticate", http.StatusInternalServerError)
		return
	}

	// Get user info
	userInfo, err := m.getUserInfo(r.Context(), token)
	if err != nil {
		m.logger.Error("Failed to get user info", "error", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	// Create session
	sessionDuration, _ := m.cfg.GetInt("auth.sessionDuration", 60)
	session := &Session{
		UserID:    userInfo["id"].(string),
		Username:  userInfo["username"].(string),
		Token:     token,
		ExpiresAt: time.Now().Add(time.Duration(sessionDuration) * time.Minute),
	}

	m.sessionsMux.Lock()
	m.sessions[session.UserID] = session
	m.sessionsMux.Unlock()

	m.logger.Info("User authenticated", "userId", session.UserID, "username", session.Username, "state", state)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<html><body><h1>Authentication Successful!</h1><p>You can close this window and return to Discord.</p></body></html>")
}

// handleHealth handles health check endpoint
func (m *Manager) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

// getUserInfo retrieves user information from the OAuth provider
func (m *Manager) getUserInfo(ctx context.Context, token *oauth2.Token) (map[string]interface{}, error) {
	provider, _ := m.cfg.GetString("auth.provider", "discord")

	var userURL string
	switch provider {
	case "discord":
		userURL = "https://discord.com/api/users/@me"
	case "google":
		userURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	case "github":
		userURL = "https://api.github.com/user"
	case "custom":
		endpointsMap, _ := m.cfg.GetMap("auth.endpoints")
		userURL, _ = endpointsMap["userUrl"].(string)
	}

	client := m.oauthConfig.Client(ctx, token)
	resp, err := client.Get(userURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return userInfo, nil
}

// GetAuthURL generates an OAuth authorization URL
func (m *Manager) GetAuthURL(state string) string {
	if !m.enabled {
		return ""
	}
	return m.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// IsAuthenticated checks if a user is authenticated
func (m *Manager) IsAuthenticated(userID string) bool {
	if !m.enabled {
		return true // If auth is disabled, everyone is "authenticated"
	}

	m.sessionsMux.RLock()
	session, exists := m.sessions[userID]
	m.sessionsMux.RUnlock()

	if !exists {
		return false
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		m.sessionsMux.Lock()
		delete(m.sessions, userID)
		m.sessionsMux.Unlock()
		return false
	}

	return true
}

// IsAuthorized checks if a user is authorized (has required role or is in authorized users list)
func (m *Manager) IsAuthorized(userID string, userRoles []string) bool {
	if !m.enabled {
		return true
	}

	// Check authorized users
	for _, authorizedUser := range m.authorizedUsers {
		if authorizedUser == userID {
			return true
		}
	}

	// Check authorized roles
	for _, authorizedRole := range m.authorizedRoles {
		for _, userRole := range userRoles {
			if authorizedRole == userRole {
				return true
			}
		}
	}

	return false
}

// Enabled returns whether authentication is enabled
func (m *Manager) Enabled() bool {
	return m.enabled
}

// Close shuts down the auth manager
func (m *Manager) Close() error {
	if m.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return m.server.Shutdown(ctx)
	}
	return nil
}
