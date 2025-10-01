package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/phpvm/data"
)

var switchCmd = &cobra.Command{
	Use:   "switch [version]",
	Short: "Show or switch to a specific PHP version",
	Long: `Show the current PHP version or switch to a specific version as active.
If no version is specified, it shows the current active PHP version.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return showCurrentVersion()
		}
		return setVersion(args[0])
	},
}

func init() {
	RootCmd.AddCommand(switchCmd)
}

func showCurrentVersion() error {
	out, err := exec.Command("php", "-v").CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get PHP version: %v\n%s", err, out)
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) > 0 {
		fmt.Println("Current PHP version:")
		fmt.Println(lines[0])
	}
	return nil
}

func setVersion(version string) error {
	// Check if version is installed
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	versionDir := filepath.Join(homeDir, ".phpvm", "versions", version)
	phpBinary := filepath.Join(versionDir, "php")

	if _, err := os.Stat(phpBinary); os.IsNotExist(err) {
		return fmt.Errorf("PHP version %s is not installed. Use 'phpvm install %s' first", version, version)
	}

	// Create symlink directory
	binDir := filepath.Join(homeDir, ".phpvm", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %v", err)
	}

	// Create or update symlink
	symlinkPath := filepath.Join(binDir, "php")

	// Remove existing symlink if it exists
	if _, err := os.Lstat(symlinkPath); err == nil {
		if err := os.Remove(symlinkPath); err != nil {
			return fmt.Errorf("failed to remove existing symlink: %v", err)
		}
	}

	// Create new symlink
	if err := os.Symlink(phpBinary, symlinkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %v", err)
	}

	// Create or update Composer symlink
	if err := createComposerSymlink(version, binDir); err != nil {
		fmt.Printf("⚠️  Warning: Failed to create Composer symlink: %v\n", err)
	}

	fmt.Printf("✅ Switched to PHP %s\n", version)

	// Automatically add to PATH for Linux
	added, err := addToPath(binDir)
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not automatically add to PATH: %v\n", err)
		fmt.Printf("Please manually add %s to your PATH\n", binDir)
	} else if added {
		fmt.Printf("✅ Added %s to your PATH\n", binDir)

		// Generate and execute auto-source script
		if err := createAndExecuteSourceScript(binDir); err != nil {
			fmt.Printf("⚠️  Warning: Could not auto-apply PATH changes: %v\n", err)
			fmt.Printf("Please restart your terminal or run 'source ~/.bashrc' (or ~/.zshrc) to apply changes\n")
		} else {
			fmt.Printf("✅ PATH changes applied automatically\n")
		}
	} else {
		fmt.Printf("ℹ️  %s is already in your PATH\n", binDir)
	}

	return nil
}

// createComposerSymlink creates a symlink for the Composer version compatible with the PHP version
func createComposerSymlink(phpVersion, binDir string) error {
	// Find compatible Composer version
	composerVersion := data.GetCompatibleComposerVersion(phpVersion)
	if composerVersion == nil {
		return fmt.Errorf("no compatible Composer version found for PHP %s", phpVersion)
	}

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	// Path to the Composer phar file
	composerPharPath := filepath.Join(homeDir, ".phpvm", "composer", composerVersion.Version, "composer.phar")
	
	// Check if Composer phar exists
	if _, err := os.Stat(composerPharPath); os.IsNotExist(err) {
		return fmt.Errorf("Composer %s not found at %s", composerVersion.Version, composerPharPath)
	}

	// Create Composer symlink in bin directory
	composerSymlinkPath := filepath.Join(binDir, "composer")

	// Remove existing symlink if it exists
	if _, err := os.Lstat(composerSymlinkPath); err == nil {
		if err := os.Remove(composerSymlinkPath); err != nil {
			return fmt.Errorf("failed to remove existing Composer symlink: %v", err)
		}
	}

	// Get PHP binary path for the wrapper script
	phpBinaryPath := filepath.Join(binDir, "php")

	// Create wrapper script content
	scriptContent := fmt.Sprintf("#!/bin/bash\n%s %s \"$@\"\n", phpBinaryPath, composerPharPath)

	// Write the wrapper script
	if err := os.WriteFile(composerSymlinkPath, []byte(scriptContent), 0755); err != nil {
		return fmt.Errorf("failed to create Composer wrapper script: %v", err)
	}

	fmt.Printf("✅ Composer %s linked to PHP %s\n", composerVersion.Version, phpVersion)
	return nil
}

// createAndExecuteSourceScript creates and executes a script to apply PATH changes immediately
func createAndExecuteSourceScript(binDir string) error {
	println("createAndExecuteSourceScript")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	// Determine which shell config file to source
	var configFile string
	shellConfigs := []string{
		filepath.Join(homeDir, ".zshrc"),
		filepath.Join(homeDir, ".bashrc"),
		filepath.Join(homeDir, ".profile"),
	}

	for _, config := range shellConfigs {
		if _, err := os.Stat(config); err == nil {
			println("Found shell config file: " + config)
			configFile = config
			break
		}
	}

	if configFile == "" {
		return fmt.Errorf("no shell configuration file found")
	}

	// Create a temporary script that sources the config and runs php --version
	tempDir := os.TempDir()
	scriptPath := filepath.Join(tempDir, "phpvm_source.sh")

	scriptContent := fmt.Sprintf(`#!/bin/bash
# Auto-generated script by phpvm to apply PATH changes
source "%s"
export PATH="%s:$PATH"
echo "PHP version after applying changes:"
php --version 2>/dev/null || echo "PHP not found in PATH"
`, configFile, binDir)

	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		return fmt.Errorf("failed to create source script: %v", err)
	}

	// Execute the script
	cmd := exec.Command("/bin/bash", scriptPath)
	output, err := cmd.CombinedOutput()

	// Clean up the temporary script
	os.Remove(scriptPath)

	if err != nil {
		return fmt.Errorf("failed to execute source script: %v", err)
	}

	// Show the output to confirm it worked
	fmt.Printf("Verification:\n%s", string(output))

	return nil
}

// addToPath adds the phpvm bin directory to the user's PATH by modifying shell configuration files
// Returns (wasAdded, error) where wasAdded indicates if any files were actually modified
func addToPath(binDir string) (bool, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("failed to get home directory: %v", err)
	}

	// List of shell configuration files to check and update
	shellConfigs := []string{
		filepath.Join(homeDir, ".bashrc"),
		filepath.Join(homeDir, ".zshrc"),
		filepath.Join(homeDir, ".profile"),
	}

	pathExport := fmt.Sprintf("export PATH=\"%s:$PATH\"", binDir)
	phpvmComment := "# Added by phpvm"

	updated := false
	pathAlreadyExists := false

	for _, configFile := range shellConfigs {
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			continue // Skip if file doesn't exist
		}

		// Check if PATH is already added
		pathExists := isPathAlreadyAdded(configFile, binDir)
		if pathExists {
			pathAlreadyExists = true
			continue
		}

		// Add PATH to the configuration file
		if err := appendToFile(configFile, phpvmComment, pathExport); err != nil {
			return false, fmt.Errorf("failed to update %s: %v", configFile, err)
		}

		updated = true
	}

	// If PATH already exists in any config file, don't create a new one
	if pathAlreadyExists {
		return false, nil
	}

	// If no existing shell config files were found, create .bashrc
	if !updated {
		defaultConfig := filepath.Join(homeDir, ".bashrc")

		if err := appendToFile(defaultConfig, phpvmComment, pathExport); err != nil {
			return false, fmt.Errorf("failed to create %s: %v", defaultConfig, err)
		}
		updated = true
	}

	return updated, nil
}

// isPathAlreadyAdded checks if the phpvm bin directory is already in the shell configuration file
func isPathAlreadyAdded(configFile, binDir string) bool {
	file, err := os.Open(configFile)
	if err != nil {
		return false
	}
	defer file.Close()

	// The exact export line we're looking for
	expectedExport := fmt.Sprintf("export PATH=\"%s:$PATH\"", binDir)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Check for exact match or common variations
		if line == expectedExport ||
			line == fmt.Sprintf("export PATH='%s:$PATH'", binDir) ||
			line == fmt.Sprintf("export PATH=%s:$PATH", binDir) {
			return true
		}
	}

	return false
}

// appendToFile appends lines to a file, creating it if it doesn't exist
func appendToFile(filename, comment, content string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Add a newline before our content if the file is not empty
	if stat, err := file.Stat(); err == nil && stat.Size() > 0 {
		if _, err := file.WriteString("\n"); err != nil {
			return err
		}
	}

	// Write comment and PATH export
	if _, err := file.WriteString(fmt.Sprintf("%s\n%s\n", comment, content)); err != nil {
		return err
	}

	return nil
}
