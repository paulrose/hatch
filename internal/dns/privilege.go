package dns

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

// escapeAppleScript escapes a string for embedding inside an
// AppleScript double-quoted string. Backslashes are escaped first,
// then double quotes, then single quotes (which need special handling
// in the shell layer).
func escapeAppleScript(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
