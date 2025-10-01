package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// originalHelpFunc holds Cobra's original help renderer so we can
// call it without the ASCII art when `phpvm` runs without arguments.
var originalHelpFunc func(cmd *cobra.Command, args []string)

// displayASCIIArt displays the phpvm ASCII art with purple gradient
func displayASCIIArt() {
	lines := []string{
		"\033[38;2;128;0;128m           █████                                           ",
		"\033[38;2;139;0;139m          ░░███                                            ",
		"\033[38;2;148;0;211m ████████  ░███████   ████████  █████ █████ █████████████  ",
		"\033[38;2;160;32;240m░░███░░███ ░███░░███ ░░███░░███░░███ ░░███ ░░███░░███░░███ ",
		"\033[38;2;186;85;211m ░███ ░███ ░███ ░███  ░███ ░███ ░███  ░███  ░███ ░███ ░███ ",
		"\033[38;2;200;100;220m ░███ ░███ ░███ ░███  ░███ ░███ ░░███ ███   ░███ ░███ ░███ ",
		"\033[38;2;221;160;221m ░███████  ████ █████ ░███████   ░░█████    █████░███ █████",
		"\033[38;2;230;190;255m ░███░░░  ░░░░ ░░░░░  ░███░░░     ░░░░░    ░░░░░ ░░░ ░░░░░ ",
		"\033[38;2;240;210;255m ░███                 ░███                                 ",
		"\033[38;2;250;230;255m █████                █████                                ",
		"\033[38;2;255;240;255m░░░░░                ░░░░░                                 ",
		"\033[0m", // Reset color
	}

	for _, line := range lines {
		fmt.Println(line)
	}
}

var RootCmd = &cobra.Command{
	Use:   "phpvm",
	Short: "PHP Version Manager",
	Long: `PHP Version Manager (phpvm) is a command-line tool that helps you manage
multiple PHP versions on your system. You can install, switch between, and
manage different PHP versions with ease.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Show ASCII art, then the standard help/commands list when running just `phpvm`
		displayASCIIArt()
		fmt.Println()
		if originalHelpFunc != nil {
			originalHelpFunc(cmd, args)
			return
		}
		_ = cmd.Help()
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Disable the auto-generated completion command
	RootCmd.CompletionOptions.DisableDefaultCmd = false
	
	// Wrap the default help function to prepend the ASCII art when help is shown
	originalHelpFunc = RootCmd.HelpFunc()
	RootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		displayASCIIArt()
		fmt.Println()
		originalHelpFunc(cmd, args)
	})
	
	// Add global flags here
	// RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.phpvm.yaml)")
}
