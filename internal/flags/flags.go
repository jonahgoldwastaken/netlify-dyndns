// Package flags handles CLI flags passable into netlify-dyndns
package flags

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Defaults sets defaults for all flags/environment variables
func Defaults() {
	viper.AutomaticEnv()
	viper.SetDefault("NETLIFY_TOKEN", "")
	viper.SetDefault("ND_NETLIFY_DOMAIN_NAME", "")
	viper.SetDefault("ND_RECORD_HOSTNAME", "")
	viper.SetDefault("ND_IP_API", "https://api.ipify.org")
	viper.SetDefault("ND_LOG_LEVEL", "info")
	viper.SetDefault("ND_SCHEDULE", "0 0 * * *")
	viper.SetDefault("ND_RUN_ONCE", false)
}

// Register registers all possible flags/options
func Register(rootCmd *cobra.Command) {
	flags := rootCmd.PersistentFlags()

	flags.String("token", viper.GetString("NETLIFY_TOKEN"), "The Netlify API token used to authenticate")
	flags.String("domain", viper.GetString("ND_NETLIFY_DOMAIN_NAME"), "The domain name registered at Netlify as shown on their dashboard and through their API")
	flags.String("hostname", viper.GetString("ND_RECORD_HOSTNAME"), "The hostname to be put in the A record")
	flags.String("ip-api", viper.GetString("ND_IP_API"), "The API used to retrieve public IP Address of connected network, must respond with a text string body")
	flags.String("log-level", viper.GetString("ND_LOG_LEVEL"), "Maximum level that will be written to stderr")
	flags.StringP("schedule", "S", viper.GetString("ND_SCHEDULE"), "Cron schedule the DNS check runs on")
	flags.Bool("run-once", viper.GetBool("ND_RUN_ONCE"), "Only run the updater once, immediately exiting after")
}

// TestRequired tests if any of the required flags/env variabels are absent
func TestRequired(rootCmd *cobra.Command) error {
	flags := rootCmd.PersistentFlags()

	if str, err := flags.GetString("token"); err != nil || str == "" {
		return errors.New("'token'/$NETLIFY_TOKEN not set")
	}
	if str, err := flags.GetString("domain"); err != nil || str == "" {
		return errors.New("'domain'/$ND_NETLIFY_DOMAIN_NAME not set")
	}
	if str, err := flags.GetString("hostname"); err != nil || str == "" {
		return errors.New("'hostname'/$ND_RECORD_HOSTNAME not set")
	}

	return nil
}
