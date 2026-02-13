package daemon

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// PortInfo describes a process listening on a port.
type PortInfo struct {
	Process string
	PID     int
}

// String returns a human-readable description like "nginx (PID 1234)".
func (p *PortInfo) String() string {
	if p.PID > 0 {
		return fmt.Sprintf("%s (PID %d)", p.Process, p.PID)
	}
	return p.Process
}

// CheckPort returns info about the process listening on the given TCP port,
// or nil if the port is free.
func CheckPort(port int) (*PortInfo, error) {
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("invalid port number: %d", port)
	}
	out, err := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port), "-sTCP:LISTEN", "-P", "-n").Output()
	if err != nil {
		// lsof exits 1 when no matching files found — port is free.
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("running lsof: %w", err)
	}
	return parseLsofOutput(string(out), port), nil
}

// parseLsofOutput extracts process name and PID from lsof output.
// Returns nil if no matching line is found.
func parseLsofOutput(output string, port int) *PortInfo {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil // header only or empty
	}

	// lsof output columns: COMMAND PID USER FD TYPE DEVICE SIZE/OFF NODE NAME
	// Take the first data line (skip header).
	fields := strings.Fields(lines[1])
	if len(fields) < 2 {
		return nil
	}

	name := fields[0]
	pid, err := strconv.Atoi(fields[1])
	if err != nil {
		return &PortInfo{Process: name}
	}
	return &PortInfo{Process: name, PID: pid}
}
