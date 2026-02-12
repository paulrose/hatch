package certs

import (
	"fmt"
	"os/exec"
	"strings"
)

// CommandRunner is the interface for executing shell commands, allowing
// injection for testing.
type CommandRunner interface {
	Run(command string) error
}

// OSAScriptRunner executes commands with administrator privileges via
// macOS osascript.
type OSAScriptRunner struct{}

// Run executes the given shell command with administrator privileges
// using osascript's "do shell script" with administrator privileges.
func (r *OSAScriptRunner) Run(command string) error {
	escaped := escapeAppleScript(command)
	script := fmt.Sprintf(`do shell script "%s" with administrator privileges`, escaped)
	out, err := exec.Command("osascript", "-e", script).CombinedOutput()
	if err != nil {
		return fmt.Errorf("privileged command failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// escapeAppleScript escapes backslashes, double quotes, and single quotes
// for embedding in an AppleScript string.
func escapeAppleScript(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, `'`, `'\\''`)
	return s
}

// TrustCA adds the CA certificate to the macOS System Keychain as a
// trusted root. Requires elevated privileges via the provided runner.
func TrustCA(runner CommandRunner, certPath string) error {
	cmd := fmt.Sprintf("security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain %s", certPath)
	if err := runner.Run(cmd); err != nil {
		return fmt.Errorf("trusting CA certificate: %w", err)
	}
	return nil
}

// UntrustCA removes the CA certificate from the macOS trusted certificates.
// Requires elevated privileges via the provided runner.
func UntrustCA(runner CommandRunner, certPath string) error {
	cmd := fmt.Sprintf("security remove-trusted-cert -d %s", certPath)
	if err := runner.Run(cmd); err != nil {
		return fmt.Errorf("untrusting CA certificate: %w", err)
	}
	return nil
}

// IsCATrusted reports whether the certificate at certPath is trusted by
// the system. It runs the non-privileged security verify-cert command.
func IsCATrusted(certPath string) bool {
	err := exec.Command("security", "verify-cert", "-c", certPath).Run()
	return err == nil
}
