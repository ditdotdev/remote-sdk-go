// Copyright Dit 2026
// SPDX-License-Identifier: BUSL-1.1

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
//
// It also creates two extra copies (echo2, echo3) so LRU tests can Load() multiple distinct remote types from a single
// build of the cmd/echo source. Each copy is the same binary; the only thing that varies is the path that Load() spawns,
// which is the cache key.
func TestMain(m *testing.M) {
	if err := buildEchoPlugin(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build echo plugin for tests: %v\n", err)
		os.Exit(1)
	}
	for _, alias := range []string{"echo2", "echo3"} {
		if err := copyEchoBinary(alias); err != nil {
			fmt.Fprintf(os.Stderr, "failed to create %s alias of echo plugin: %v\n", alias, err)
			os.Exit(1)
		}
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

func copyEchoBinary(alias string) error {
	buildDir := filepath.Join("..", "build")
	src := pluginBinaryPath(buildDir, "echo")
	dst := pluginBinaryPath(buildDir, alias)

	in, err := os.ReadFile(src) // #nosec G304 -- test-only path
	if err != nil {
		return err
	}
	// Test-only fixture: the binary must be executable to be spawned by Load(); 0o700 is owner-rwx, no group/other.
	return os.WriteFile(dst, in, 0o700) // #nosec G306,G703 -- test fixture, exec bit required, paths are constants
}
