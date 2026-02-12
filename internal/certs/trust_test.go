package certs

import (
	"errors"
	"strings"
	"testing"
)

// MockRunner records commands instead of executing them.
type MockRunner struct {
	Commands []string
	Err      error // error to return from Run, if any
}

func (m *MockRunner) Run(command string) error {
	m.Commands = append(m.Commands, command)
	return m.Err
}

func TestTrustCA(t *testing.T) {
	runner := &MockRunner{}
	certPath := "/tmp/certs/rootCA.pem"

	if err := TrustCA(runner, certPath); err != nil {
		t.Fatalf("TrustCA: %v", err)
	}

	if len(runner.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(runner.Commands))
	}

	want := "security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain /tmp/certs/rootCA.pem"
	if runner.Commands[0] != want {
		t.Errorf("unexpected command:\n  got:  %s\n  want: %s", runner.Commands[0], want)
	}
}

func TestUntrustCA(t *testing.T) {
	runner := &MockRunner{}
	certPath := "/tmp/certs/rootCA.pem"

	if err := UntrustCA(runner, certPath); err != nil {
		t.Fatalf("UntrustCA: %v", err)
	}

	if len(runner.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(runner.Commands))
	}

	want := "security remove-trusted-cert -d /tmp/certs/rootCA.pem"
	if runner.Commands[0] != want {
		t.Errorf("unexpected command:\n  got:  %s\n  want: %s", runner.Commands[0], want)
	}
}

func TestTrustCA_Error(t *testing.T) {
	runner := &MockRunner{Err: errors.New("permission denied")}

	err := TrustCA(runner, "/tmp/certs/rootCA.pem")
	if err == nil {
		t.Fatal("expected error from TrustCA")
	}
	if !strings.Contains(err.Error(), "trusting CA certificate") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestOSAScriptEscape(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no special chars",
			input: "security verify-cert -c /tmp/cert.pem",
			want:  "security verify-cert -c /tmp/cert.pem",
		},
		{
			name:  "single quotes",
			input: "echo 'hello world'",
			want:  `echo '\\''hello world'\\''`,
		},
		{
			name:  "double quotes",
			input: `echo "hello"`,
			want:  `echo \"hello\"`,
		},
		{
			name:  "backslashes",
			input: `path\to\file`,
			want:  `path\\to\\file`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeAppleScript(tt.input)
			if got != tt.want {
				t.Errorf("escapeAppleScript(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
