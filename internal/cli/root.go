package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const version = "0.7.0"

var rootCmd = &cobra.Command{
	Use:   "ait",
	Short: "AI Toolkit Package Manager",
	Long: `AIT is a package manager for AI agents, skills, prompts, and MCP servers.
	
It allows you to declare your project's AI dependencies in ait.yml and
automatically install and sync them to your AI tools (OpenCode, Cursor, Claude, etc).`,
	Version: version,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable colored output")

	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color"))
}

func initConfig() {
	// Set config name and paths
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Look for config in home directory
	home, err := os.UserHomeDir()
	if err == nil {
		viper.AddConfigPath(fmt.Sprintf("%s/.ait", home))
	}

	// Read config file
	viper.AutomaticEnv()
	viper.ReadInConfig()
}
