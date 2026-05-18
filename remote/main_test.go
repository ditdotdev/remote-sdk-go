/*
 * Copyright Datadatdat.
 */
package remote

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestMain builds the echo plugin binary that the RPC integration tests spawn via Load(). Building it here makes the
// test suite self-contained: developers and CI no longer need a separate `go build -o build/echo ./cmd/echo` step,
// and the binary is named with the OS-appropriate suffix (echo on Linux/macOS, echo.exe on Windows) so Load() can
// actually find it.
func TestMain(m *testing.M) {
	if err := buildEchoPlugin(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build echo plugin for tests: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func buildEchoPlugin() error {
	buildDir := filepath.Join("..", "build")
	if err := os.MkdirAll(buildDir, 0o750); err != nil {
		return err
	}

	outPath := pluginBinaryPath(buildDir, "echo")
	cmd := exec.Command("go", "build", "-o", outPath, "../cmd/echo") // #nosec G204 -- test-only, paths are constants
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
