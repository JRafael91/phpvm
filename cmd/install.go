package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/yourusername/phpvm/data"
)

var installCmd = &cobra.Command{
	Use:   "install [version]",
	Short: "Install a specific PHP version",
	Long: `Download and install a specific version of PHP.
This will download the source code for the specified version, compile it,
and install it to the PHPVM directory.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return installPHP(args[0])
	},
}

func init() {
	RootCmd.AddCommand(installCmd)
}

func installPHP(version string) error {
	fmt.Printf("Preparing to install PHP %s...\n", version)

	// Find the version in our data
	var phpVersion *data.PHPVersion
	for _, v := range data.AvailableVersions {
		if v.Version == version {
			phpVersion = &v
			break
		}
	}

	if phpVersion == nil {
		return fmt.Errorf("PHP version %s not found. Use 'phpvm list' to see available versions", version)
	}

	fmt.Printf("Found PHP %s (released: %s)\n", phpVersion.Version, phpVersion.Released.Format("2006-01-02"))

	// Determine architecture and binary URL
	var binaryURL string
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		binaryURL = phpVersion.BinaryURLx64
	case "arm64":
		binaryURL = phpVersion.BinaryURLarm64
	default:
		return fmt.Errorf("unsupported architecture: %s. Only amd64 and arm64 are supported", arch)
	}

	// Create installation directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	installDir := filepath.Join(homeDir, ".phpvm", "versions", version)
	phpBinary := filepath.Join(installDir, "php")
	
	// Check if PHP version is already installed
	if _, err := os.Stat(phpBinary); err == nil {
		fmt.Printf("✅ PHP %s is already installed at %s\n", version, installDir)
		fmt.Printf("Skipping download. Use 'phpvm switch %s' to use this version\n", version)
		
		// Still try to install/link Composer if not present
		composerScript := filepath.Join(installDir, "composer")
		if _, err := os.Stat(composerScript); err != nil {
			fmt.Printf("Installing Composer for existing PHP %s...\n", version)
			if err := installComposer(phpVersion.Version, installDir); err != nil {
				fmt.Printf("⚠️  Warning: Failed to install Composer: %v\n", err)
			} else {
				fmt.Printf("✅ Composer installed successfully\n")
			}
		} else {
			fmt.Printf("✅ Composer is already configured for this PHP version\n")
		}
		
		return nil
	}

	// Create installation directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %v", err)
	}

	// Download and install PHP binary
	fmt.Printf("Downloading PHP binary from %s...\n", binaryURL)

	if err := downloadFile(binaryURL, phpBinary); err != nil {
		return fmt.Errorf("failed to download PHP binary: %v", err)
	}

	// Make binary executable
	if err := os.Chmod(phpBinary, 0755); err != nil {
		return fmt.Errorf("failed to make PHP binary executable: %v", err)
	}

	fmt.Printf("✅ PHP %s installed successfully to %s\n", version, installDir)

	// Install Composer
	if err := installComposer(phpVersion.Version, installDir); err != nil {
		fmt.Printf("⚠️  Warning: Failed to install Composer: %v\n", err)
		fmt.Printf("You can install Composer manually later\n")
	} else {
		fmt.Printf("✅ Composer installed successfully\n")
	}

	fmt.Printf("Use 'phpvm switch %s' to switch to this version\n", version)

	return nil
}

// installComposer downloads and installs Composer for the PHP version
func installComposer(phpVersion string, phpInstallDir string) error {
	// Find compatible Composer version
	composerVersion := data.GetCompatibleComposerVersion(phpVersion)
	if composerVersion == nil {
		return fmt.Errorf("no compatible Composer version found for PHP %s", phpVersion)
	}

	// Create Composer installation directory structure
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	composerBaseDir := filepath.Join(homeDir, ".phpvm", "composer")
	composerVersionDir := filepath.Join(composerBaseDir, composerVersion.Version)
	
	// Create directories if they don't exist
	if err := os.MkdirAll(composerVersionDir, 0755); err != nil {
		return fmt.Errorf("failed to create Composer directory: %v", err)
	}

	composerPharPath := filepath.Join(composerVersionDir, "composer.phar")
	
	// Check if Composer is already installed
	if _, err := os.Stat(composerPharPath); err == nil {
		fmt.Printf("Composer %s already installed, creating symlink...\n", composerVersion.Version)
	} else {
		// Download Composer
		fmt.Printf("Downloading Composer %s from %s...\n", composerVersion.Version, composerVersion.URL)
		if err := downloadFile(composerVersion.URL, composerPharPath); err != nil {
			return fmt.Errorf("failed to download Composer: %v", err)
		}

		// Make Composer executable
		if err := os.Chmod(composerPharPath, 0755); err != nil {
			return fmt.Errorf("failed to make Composer executable: %v", err)
		}
	}

	// Create a composer wrapper script in the PHP installation directory
	composerScript := filepath.Join(phpInstallDir, "composer")
	scriptContent := fmt.Sprintf("#!/bin/bash\n%s %s \"$@\"\n", 
		filepath.Join(phpInstallDir, "php"), 
		composerPharPath)

	if err := os.WriteFile(composerScript, []byte(scriptContent), 0755); err != nil {
		return fmt.Errorf("failed to create Composer script: %v", err)
	}

	fmt.Printf("Composer %s linked to PHP %s\n", composerVersion.Version, phpVersion)
	return nil
}

// downloadFile downloads a file from URL to the specified path
func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
