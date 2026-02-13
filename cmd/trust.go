package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/paulrose/hatch/internal/certs"
	"github.com/paulrose/hatch/internal/config"
)

var trustCmd = &cobra.Command{
	Use:   "trust",
	Short: "Trust the Hatch root CA in the macOS Keychain",
	Long:  `Re-trusts the Hatch root CA certificate in the system Keychain. Useful if the certificate was removed or the Keychain was reset.`,
	RunE:  runTrust,
}

func init() {
	rootCmd.AddCommand(trustCmd)
}

func runTrust(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	caPaths := certs.NewCAPaths(config.CertsDir())

	// Ensure the root CA exists before attempting to trust it.
	if !certs.CAExists(caPaths) {
		if err := config.EnsureConfigDir(); err != nil {
			fmt.Printf("  %s Failed to create config directory: %v\n", red("✗"), err)
			os.Exit(1)
		}
		if err := certs.GenerateCA(caPaths); err != nil {
			fmt.Printf("  %s Failed to generate root CA: %v\n", red("✗"), err)
			os.Exit(1)
		}
		fmt.Printf("  %s Root CA generated\n", green("✓"))
	}

	if certs.IsCATrusted(caPaths.Cert) {
		fmt.Printf("  %s Root CA is already trusted\n", green("✓"))
		return nil
	}

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
	return nil
}
