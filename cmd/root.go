// Package cmd provides the command-line interface for the Discord bot.
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/geekxflood/common/logging"
	"github.com/geekxflood/gxf-discord-bot/pkg/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	debug   bool
)

// rootCmd represents the base command when called without subcommands
var rootCmd = &cobra.Command{
	Use:   "gxf-discord-bot",
	Short: "A highly configurable Discord bot",
	Long: `GXF Discord Bot - A highly configurable Discord bot designed for
Kubernetes deployments with enterprise-grade features including
Vault/OpenBao secret management and OAuth-based authentication.`,
	RunE: runBot,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yaml", "config file path")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")
}

func runBot(cmd *cobra.Command, args []string) error {
	// Initialize logger
	logger, cleanup, err := logging.NewLogger(logging.Config{
		Level:  getLogLevel(),
		Format: "json",
		Output: "stdout",
	})
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	defer cleanup.Close()

	logger.Info("Starting GXF Discord Bot")

	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	logger.Info("Configuration loaded and validated")

	// TODO: Initialize and start bot
	logger.Info("Bot initialization not yet implemented (TDD in progress)")

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = ctx // Will be used when bot is implemented

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	logger.Info("Shutdown signal received, stopping bot...")

	return nil
}

func getLogLevel() string {
	if debug {
		return "debug"
	}
	return "info"
}
