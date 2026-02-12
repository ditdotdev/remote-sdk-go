/*
 * Copyright Datadatdat.
 */
package remote

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegister(t *testing.T) {
	Clear()

	r := new(MockRemote)
	r.On("Type").Return("mock", nil)
	Register(r)

	res := Get("mock")
	typ, _ := res.Type()
	assert.Equal(t, "mock", typ)
	r.AssertExpectations(t)
}

func TestGetNonExistent(t *testing.T) {
	Clear()

	res := Get("nonexistent")
	assert.Nil(t, res)
}

func TestUnload(t *testing.T) {
	Clear()
	r, err := Load("echo", "../build")
	if err != nil {
		t.Skip("Skipping test - echo plugin not built")
		return
	}

	assert.NotNil(t, r)

	// Unload should not panic
	Unload("echo")
}

func TestUnloadNonExistent(t *testing.T) {
	Clear()
	// Should not panic
	Unload("doesnotexist")
}

func TestLoadCached(t *testing.T) {
	Clear()
	r1, err := Load("echo", "../build")
	if err != nil {
		t.Skip("Skipping test - echo plugin not built")
		return
	}

	r2, err := Load("echo", "../build")
	assert.NoError(t, err)
	assert.Equal(t, r1, r2) // Should return the same cached instance
}

func TestLoadInvalidPath(t *testing.T) {
	Clear()
	_, err := Load("nonexistent", "/invalid/path")
	assert.Error(t, err)
}
