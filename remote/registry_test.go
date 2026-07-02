// Copyright Dit 2026
// SPDX-License-Identifier: BUSL-1.1

package remote

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockRemote is a minimal Remote used only by registry tests where the interface methods don't matter —
// only Type() is invoked (by Register), and the rest panic if called. Distinct from MockRemote in mock_test.go,
// which is a testify mock requiring expectations to be set up per call.
type mockRemote struct {
	typ string
}

func (m *mockRemote) Type() (string, error) { return m.typ, nil }
func (m *mockRemote) FromURL(string, map[string]string) (map[string]interface{}, error) {
	panic("not implemented")
}
func (m *mockRemote) ToURL(map[string]interface{}) (string, map[string]string, error) {
	panic("not implemented")
}
func (m *mockRemote) GetParameters(map[string]interface{}) (map[string]interface{}, error) {
	panic("not implemented")
}
func (m *mockRemote) ValidateRemote(map[string]interface{}) error     { panic("not implemented") }
func (m *mockRemote) ValidateParameters(map[string]interface{}) error { panic("not implemented") }
func (m *mockRemote) ListCommits(map[string]interface{}, map[string]interface{}, []Tag) ([]Commit, error) {
	panic("not implemented")
}
func (m *mockRemote) GetCommit(map[string]interface{}, map[string]interface{}, string) (*Commit, error) {
	panic("not implemented")
}

// TestRegistryIsolation verifies that two independently-constructed registries do not share state.
// This is the property that unblocks t.Parallel() in tests and multi-tenant scenarios — without it,
// every test that touches the registry has to serialize on the package-level Default.
func TestRegistryIsolation(t *testing.T) {
	r1 := New()
	r2 := New()

	mockA := &mockRemote{typ: "registry-isolation-a"}
	r1.Register(mockA)

	if _, ok := r2.Get("registry-isolation-a"); ok {
		t.Fatalf("expected r2 to be isolated from r1, but it returned r1's registration")
	}
	got, ok := r1.Get("registry-isolation-a")
	assert.True(t, ok)
	assert.Same(t, mockA, got)
}

// TestRegistryGetReturnsOkFlag verifies the new (Remote, bool) signature for Get. The bool lets callers
// distinguish "not registered" from "registered with nil" — under the old signature, both looked the same
// and led to NPEs in callers (the issue #4 motivation).
func TestRegistryGetReturnsOkFlag(t *testing.T) {
	r := New()

	got, ok := r.Get("never-registered")
	assert.False(t, ok)
	assert.Nil(t, got)

	mock := &mockRemote{typ: "get-ok-flag"}
	r.Register(mock)

	got, ok = r.Get("get-ok-flag")
	assert.True(t, ok)
	assert.Same(t, mock, got)
}

// TestDefaultRegistryFreeFunctions verifies that the package-level free functions (Register, Get, Loaded)
// delegate to the Default registry. This is the source-level backwards compatibility shim — existing
// callers shouldn't have to switch to Default.Method() to keep working.
func TestDefaultRegistryFreeFunctions(t *testing.T) {
	ClearForTesting()
	defer ClearForTesting()

	mock := &mockRemote{typ: "default-delegation"}
	Register(mock)

	got, ok := Get("default-delegation")
	assert.True(t, ok)
	assert.Same(t, mock, got)

	// Default.Get should see the same registration
	got2, ok2 := Default.Get("default-delegation")
	assert.True(t, ok2)
	assert.Same(t, mock, got2)
}

// TestLoaded verifies the new Loaded() query. Before this PR, Unload silently succeeded for never-loaded
// types and there was no way to ask "is this type currently loaded?" — making the Load/Unload pair
// asymmetric. Loaded() closes that gap.
func TestLoaded(t *testing.T) {
	r := New()
	assert.False(t, r.Loaded("echo"), "expected false for a never-loaded type")
}

// TestConcurrentLoad verifies the race-condition fix from issue #46 finding #1. Under the buggy
// implementation, N concurrent goroutines calling Load("echo", ...) for the first time would all see an
// empty cache, all spawn a plugin subprocess, then race to write loadedRemotes — leaving the losers'
// subprocesses leaked. We can detect this deterministically by checking that all returned Remote
// pointers are identical (the cache-deduplication invariant). The -race detector will additionally
// flag the map data race on CI.
func TestConcurrentLoad(t *testing.T) {
	ClearForTesting()
	defer ClearForTesting()

	const n = 8
	var wg sync.WaitGroup
	results := make([]Remote, n)
	errs := make([]error, n)

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			r, err := Load("echo", "../build")
			results[idx] = r
			errs[idx] = err
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if !assert.NoError(t, err, "goroutine %d", i) {
			return
		}
	}

	// All N goroutines must see the same cached Remote. If different pointers come back, multiple
	// subprocesses were spawned and the loser-subprocesses leaked.
	first := results[0]
	assert.NotNil(t, first)
	for i := 1; i < n; i++ {
		assert.Samef(t, first, results[i], "goroutine %d got a different Remote than goroutine 0", i)
	}

	// Cleanup
	Unload("echo")
}
