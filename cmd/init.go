package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/paulrose/hatch/internal/certs"
	"github.com/paulrose/hatch/internal/config"
	"github.com/paulrose/hatch/internal/dns"
)

// sudoRunner executes commands via sudo, inheriting the terminal's TTY
// so the password prompt works. Used as a fallback when osascript cannot
// complete Keychain authorization.
type sudoRunner struct{}

func (r *sudoRunner) Run(command string) error {
	cmd := exec.Command("sudo", "sh", "-c", command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sudo command failed: %s: %w", strings.TrimSpace(command), err)
	}
	return nil
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Hatch for first-time use",
	Long:  `Sets up the Hatch config directory, generates a root CA, trusts it in Keychain, and installs the DNS resolver. Does not start the daemon.`,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()

	fmt.Printf("Initializing %s...\n\n", cyan("Hatch"))

	anyCreated := false

	// Step 1: Config directory
	dirExists := dirExistsAt(config.Dir())
	if err := config.EnsureConfigDir(); err != nil {
		fmt.Printf("  %s Failed to create config directory: %v\n", red("✗"), err)
		os.Exit(1)
	}
	if dirExists {
		fmt.Printf("  %s Config directory exists\n", green("✓"))
	} else {
		fmt.Printf("  %s Config directory created (~/.hatch)\n", green("✓"))
		anyCreated = true
	}

	// Step 2: Config file
	fileExists := fileExistsAt(config.ConfigFile())
	if err := config.EnsureConfigFile(); err != nil {
		fmt.Printf("  %s Failed to write default config: %v\n", red("✗"), err)
		os.Exit(1)
	}
	if fileExists {
		fmt.Printf("  %s Config file exists\n", green("✓"))
	} else {
		fmt.Printf("  %s Default config written (~/.hatch/config.yml)\n", green("✓"))
		anyCreated = true
	}

	// Load config to get TLD for later steps
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("  %s Failed to load config: %v\n", red("✗"), err)
		os.Exit(1)
	}

	// Step 3: Root CA
	caPaths := certs.NewCAPaths(config.CertsDir())
	if certs.CAExists(caPaths) {
		fmt.Printf("  %s Root CA exists\n", green("✓"))
	} else {
		if err := certs.GenerateCA(caPaths); err != nil {
			fmt.Printf("  %s Failed to generate root CA: %v\n", red("✗"), err)
			os.Exit(1)
		}
		fmt.Printf("  %s Root CA generated\n", green("✓"))
		anyCreated = true
	}

	// Step 4: Trust CA
	if certs.IsCATrusted(caPaths.Cert) {
		fmt.Printf("  %s Root CA already trusted\n", green("✓"))
	} else {
		// Try osascript first; fall back to sudo if it fails (macOS can
		// block the Keychain authorization dialog inside osascript).
		var trustErr error
		trustErr = certs.TrustCA(&certs.OSAScriptRunner{}, caPaths.Cert)
		if trustErr != nil {
			fmt.Printf("  %s osascript trust failed, trying sudo...\n", yellow("!"))
			trustErr = certs.TrustCA(&sudoRunner{}, caPaths.Cert)
		}
		if trustErr != nil {
			fmt.Printf("  %s Failed to trust root CA: %v\n", red("✗"), trustErr)
			os.Exit(1)
		}
		fmt.Printf("  %s Root CA trusted in Keychain\n", green("✓"))
		anyCreated = true
	}

	// Step 5: DNS resolver
	tld := cfg.Settings.TLD
	if dns.IsResolverInstalled(tld) {
		fmt.Printf("  %s DNS resolver already installed\n", green("✓"))
	} else {
		runner := &dns.OSAScriptRunner{}
		if err := dns.InstallResolverFile(runner, tld, dns.DefaultListenIP, dns.DefaultPort); err != nil {
			fmt.Printf("  %s Failed to install DNS resolver: %v\n", red("✗"), err)
			os.Exit(1)
		}
		fmt.Printf("  %s DNS resolver installed for .%s\n", green("✓"), tld)
		anyCreated = true
	}

	fmt.Println()
	if anyCreated {
		fmt.Printf("%s initialized! Next steps:\n", cyan("Hatch"))
		fmt.Println("  hatch up       Start the daemon")
		fmt.Println("  hatch version  Check installed version")
	} else {
		fmt.Printf("Already initialized. Run '%s' to start the daemon.\n", color.New(color.Bold).Sprint("hatch up"))
	}

	return nil
}

func dirExistsAt(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExistsAt(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
