package cmd

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/geekxflood/common/logging"
	"github.com/geekxflood/gxf-discord-bot/internal/bot"
	"github.com/geekxflood/gxf-discord-bot/internal/config"
	"github.com/spf13/cobra"
)

//go:embed schema/config.cue
var embeddedSchema []byte

var (
	cfgFile string
	debug   bool

	rootCmd = &cobra.Command{
		Use:   "gxf-discord-bot",
		Short: "A configurable Discord bot for Kubernetes",
		Long: `GXF Discord Bot is a highly configurable Discord bot that can be deployed
in Kubernetes clusters. It supports YAML-based configuration with CUE schema
validation, Vault/OpenBao integration for secrets, and OAuth-based authentication.`,
		RunE: runBot,
	}
)

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yaml", "config file path")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")
}

func runBot(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize logger
	logger, err := initLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	logger.Info("Starting GXF Discord Bot")

	// Load configuration
	cfgProvider, err := loadConfig(logger)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logger.Info("Configuration loaded successfully")

	// Create and start bot
	discordBot, err := bot.New(ctx, cfgProvider, logger)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	if err := discordBot.Start(ctx); err != nil {
		return fmt.Errorf("failed to start bot: %w", err)
	}

	logger.Info("Bot is now running. Press CTRL-C to exit.")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	select {
	case sig := <-sigChan:
		logger.Info("Received shutdown signal", "signal", sig)
	case <-ctx.Done():
		logger.Info("Context cancelled")
	}

	// Graceful shutdown
	logger.Info("Shutting down bot...")
	if err := discordBot.Stop(); err != nil {
		logger.Error("Error during shutdown", "error", err)
		return err
	}

	logger.Info("Bot shutdown complete")
	return nil
}

func initLogger() (logging.Logger, error) {
	level := "info"
	if debug {
		level = "debug"
	}

	logger, _, err := logging.NewLogger(logging.Config{
		Level:  level,
		Format: "json",
	})
	if err != nil {
		return nil, err
	}

	return logger, nil
}

func loadConfig(logger logging.Logger) (config.Provider, error) {
	// Create config manager with embedded schema
	cfgManager, err := config.NewManager(config.Options{
		SchemaContent: embeddedSchema,
		ConfigPath:    cfgFile,
		Logger:        logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create config manager: %w", err)
	}

	// Validate configuration
	if err := cfgManager.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfgManager, nil
}
