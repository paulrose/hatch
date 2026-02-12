package dns

import "testing"

// MockRunner records commands instead of executing them.
type MockRunner struct {
	Commands []string
	Err      error // error to return from Run, if any
}

func (m *MockRunner) Run(command string) error {
	m.Commands = append(m.Commands, command)
	return m.Err
}

func TestMockRunner_RecordsCommands(t *testing.T) {
	runner := &MockRunner{}

	runner.Run("echo hello")
	runner.Run("echo world")

	if len(runner.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(runner.Commands))
	}
	if runner.Commands[0] != "echo hello" {
		t.Errorf("expected 'echo hello', got %q", runner.Commands[0])
	}
	if runner.Commands[1] != "echo world" {
		t.Errorf("expected 'echo world', got %q", runner.Commands[1])
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
			input: "echo hello",
			want:  "echo hello",
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
		{
			name:  "mixed special chars",
			input: `echo "it's a \"test\""`,
			want:  `echo \"it's a \\\"test\\\"\"`,
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
