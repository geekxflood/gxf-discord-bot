package cmd

import (
	"fmt"

	"github.com/geekxflood/gxf-discord-bot/internal/config"
	"github.com/spf13/cobra"
)

var (
	validateCmd = &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration file",
		Long: `Validate a configuration file against the embedded CUE schema.
This checks for syntax errors, missing required fields, and type mismatches.`,
		RunE: validateConfig,
	}
)

func init() {
	rootCmd.AddCommand(validateCmd)
}

func validateConfig(_ *cobra.Command, _ []string) error {
	fmt.Printf("Validating configuration file: %s\n", cfgFile)

	// Create config manager
	cfgManager, err := config.NewManager(config.Options{
		SchemaContent: embeddedSchema,
		ConfigPath:    cfgFile,
	})
	if err != nil {
		return fmt.Errorf("❌ Failed to load config: %w", err)
	}

	// Validate
	if err := cfgManager.Validate(); err != nil {
		return fmt.Errorf("❌ Validation failed: %w", err)
	}

	fmt.Println("✅ Configuration is valid!")

	// Print configuration summary
	if err := printConfigSummary(cfgManager); err != nil {
		return err
	}

	return nil
}

func printConfigSummary(cfg config.Provider) error {
	fmt.Println("\n📋 Configuration Summary:")

	// Bot settings
	prefix, _ := cfg.GetString("bot.prefix", "!")
	fmt.Printf("  Bot Prefix: %s\n", prefix)

	// Check token source
	if cfg.Exists("bot.tokenEnvVar") {
		tokenEnv, _ := cfg.GetString("bot.tokenEnvVar", "")
		fmt.Printf("  Token Source: Environment variable (%s)\n", tokenEnv)
	} else if cfg.Exists("bot.tokenVaultPath") {
		vaultPath, _ := cfg.GetString("bot.tokenVaultPath", "")
		fmt.Printf("  Token Source: Vault (%s)\n", vaultPath)
	} else {
		fmt.Println("  Token Source: Direct (⚠️  Not recommended for production)")
	}

	// Secrets configuration
	if cfg.Exists("secrets") {
		provider, _ := cfg.GetString("secrets.provider", "")
		address, _ := cfg.GetString("secrets.address", "")
		fmt.Printf("  Secrets: %s (%s)\n", provider, address)
	}

	// Auth configuration
	if cfg.Exists("auth") {
		enabled, _ := cfg.GetBool("auth.enabled", false)
		if enabled {
			provider, _ := cfg.GetString("auth.provider", "")
			fmt.Printf("  Authentication: Enabled (%s)\n", provider)
		}
	}

	// Count actions
	actionCount := 0
	for i := 0; ; i++ {
		key := fmt.Sprintf("actions[%d]", i)
		if !cfg.Exists(key) {
			break
		}
		actionCount++
	}
	fmt.Printf("  Actions: %d configured\n", actionCount)

	// Logging
	logLevel, _ := cfg.GetString("logging.level", "info")
	logFormat, _ := cfg.GetString("logging.format", "json")
	fmt.Printf("  Logging: %s (%s)\n", logLevel, logFormat)

	return nil
}
