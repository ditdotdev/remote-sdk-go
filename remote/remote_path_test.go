// Copyright Dit 2026
// SPDX-License-Identifier: BUSL-1.1

package remote

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPluginBinaryPath verifies that the helper resolves the plugin binary path
// using the OS-appropriate executable suffix. On Windows, Go's `go build` emits
// binaries with a .exe suffix, and exec.Command does not auto-append .exe when
// the path contains a separator. Without this helper, Load() would never find
// its plugin binary on Windows.
func TestPluginBinaryPath(t *testing.T) {
	got := pluginBinaryPath("/some/dir", "echo")

	var want string
	if runtime.GOOS == windowsGOOS {
		want = filepath.Join("/some/dir", "echo.exe")
	} else {
		want = filepath.Join("/some/dir", "echo")
	}

	assert.Equal(t, want, got)
}

// TestPluginBinaryPathPreservesExistingExtension ensures that callers who
// already pass a fully-qualified binary name (e.g. one that already ends in
// .exe on Windows) get back the same name without a doubled suffix.
func TestPluginBinaryPathPreservesExistingExtension(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only behavior")
	}

	got := pluginBinaryPath("/some/dir", "echo.exe")
	assert.Equal(t, filepath.Join("/some/dir", "echo.exe"), got)
}
