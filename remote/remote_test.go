/*
 * Copyright Datadatdat.
 */
package remote

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	ClearForTesting()

	r := new(MockRemote)
	r.On("Type").Return("mock", nil)
	Register(r)

	res, ok := Get("mock")
	assert.True(t, ok)
	typ, _ := res.Type()
	assert.Equal(t, "mock", typ)
	r.AssertExpectations(t)
}

func TestGetNonExistent(t *testing.T) {
	ClearForTesting()

	res, ok := Get("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, res)
}

func TestUnload(t *testing.T) {
	ClearForTesting()
	r, err := Load("echo", "../build")
	assert.NoError(t, err)
	assert.NotNil(t, r)

	Unload("echo")
	assert.False(t, Loaded("echo"))
}

func TestUnloadNonExistent(t *testing.T) {
	ClearForTesting()
	// Should not panic
	Unload("doesnotexist")
}

func TestLoadCached(t *testing.T) {
	ClearForTesting()
	r1, err := Load("echo", "../build")
	assert.NoError(t, err)

	r2, err := Load("echo", "../build")
	assert.NoError(t, err)
	assert.Same(t, r1, r2, "Load should return the same cached instance")
}

func TestLoadInvalidPath(t *testing.T) {
	ClearForTesting()
	_, err := Load("nonexistent", "/invalid/path")
	assert.Error(t, err)
}
