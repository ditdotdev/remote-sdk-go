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
	r.On("Type").Return("mock")
	Register(r)

	res := Get("mock")
	typ, _ := res.Type()
	assert.Equal(t, "mock", typ)
	r.AssertExpectations(t)
}
