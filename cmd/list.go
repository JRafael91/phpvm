package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
	"github.com/yourusername/phpvm/data"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available PHP versions",
	Long: `List all PHP versions that are available for installation.
This fetches the list of available versions from the official PHP.net website.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listAvailableVersions()
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
}

func listAvailableVersions() error {
	versions := data.AvailableVersions

	// Sort versions by release date (newest first)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Released.After(versions[j].Released)
	})

	fmt.Println("Available PHP versions:")
	fmt.Printf("%-12s %-12s\n", "Version", "Released")
	fmt.Println("----------------------------------")
	
	for _, v := range versions {
		status := " "
		if isVersionInstalled(v.Version) {
			status = "*"
		}
		
		fmt.Printf("%-12s %-12s\n",
			status + v.Version, 
			v.Released.Format("2006-01-02"))
	}
	
	fmt.Println("\nUse 'phpvm install <version>' to install a specific version")
	fmt.Println("* = Already installed")

	return nil
}

// isVersionInstalled checks if a specific PHP version is installed
func isVersionInstalled(version string) bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	
	versionDir := filepath.Join(homeDir, ".phpvm", "versions", version)
	phpBinary := filepath.Join(versionDir, "php")
	
	if _, err := os.Stat(phpBinary); os.IsNotExist(err) {
		return false
	}
	
	return true
}
