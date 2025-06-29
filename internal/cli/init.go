// internal/cli/init.go
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/briandowns/spinner"
	"github.com/localcloud/localcloud/internal/config"
	"github.com/spf13/cobra"
)

var (
	interactive bool
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new LocalCloud project",
	Long:  `Initialize a new LocalCloud project in the current directory or with the specified name.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runInit,
	Example: `  lc init                    # Initialize in current directory
  lc init my-project         # Create new project directory
  lc init --interactive      # Interactive setup wizard`,
}

func init() {
	initCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Run interactive setup wizard")
}

func runInit(cmd *cobra.Command, args []string) error {
	projectName := "my-project"
	if len(args) > 0 {
		projectName = args[0]
	}

	// Run interactive wizard if requested
	if interactive {
		return RunInteractiveInit(projectName)
	}

	// Create project directory if specified
	projectDir := projectPath
	if len(args) > 0 {
		projectDir = filepath.Join(projectPath, projectName)
		if err := os.MkdirAll(projectDir, 0755); err != nil {
			return fmt.Errorf("failed to create project directory: %w", err)
		}
	}

	// Check if already initialized
	configPath := filepath.Join(projectDir, ".localcloud")
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		return fmt.Errorf("project already initialized in %s", projectDir)
	}

	// Start initialization
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Initializing LocalCloud project..."
	s.Start()

	// Create .localcloud directory
	if err := os.MkdirAll(configPath, 0755); err != nil {
		s.Stop()
		return fmt.Errorf("failed to create .localcloud directory: %w", err)
	}

	// Create config file
	configFile := filepath.Join(configPath, "config.yaml")
	configContent, err := config.GenerateDefault(projectName, "custom")
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to generate config: %w", err)
	}

	if err := os.WriteFile(configFile, configContent, 0644); err != nil {
		s.Stop()
		return fmt.Errorf("failed to create config file: %w", err)
	}

	// Create .gitignore
	gitignore := filepath.Join(projectDir, ".gitignore")
	gitignoreContent := `.localcloud/data/
.localcloud/logs/
.localcloud/tunnels/
.env.local
*.log
`
	if err := os.WriteFile(gitignore, []byte(gitignoreContent), 0644); err != nil {
		s.Stop()
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}

	s.Stop()

	// Print success message
	printSuccess(fmt.Sprintf("Initialized LocalCloud project: %s", projectName))
	fmt.Println()
	fmt.Println("Next steps:")
	if len(args) > 0 {
		fmt.Printf("  1. cd %s\n", projectName)
		fmt.Println("  2. lc setup        # Configure components")
		fmt.Println("  3. lc start        # Start services")
	} else {
		fmt.Println("  1. lc setup        # Configure components")
		fmt.Println("  2. lc start        # Start services")
	}
	fmt.Println()
	fmt.Println("For more information, run: lc --help")

	return nil
}
