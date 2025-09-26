/*
 * Copyright The Titan Project Contributors.
 */
package remote

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func getEcho(t *testing.T) Remote {
	Clear()

	r, err := Load("echo", "../build")
	if assert.NoError(t, err) {
		return r
	}
	
	return nil
}

func TestPluginType(t *testing.T) {
	e := getEcho(t)
	if assert.NotNil(t, e) {
		ret, err := e.Type()
		if assert.NoError(t, err) {
			assert.Equal(t, "echo", ret)
		}
	}
}

func TestFromURL(t *testing.T) {
	e := getEcho(t)
	if assert.NotNil(t, e) {
		props, err := e.FromURL("echo://echo", map[string]string{"a": "b"})
		if assert.NoError(t, err) {
			assert.Len(t, props, 2)
			assert.Equal(t, "b", props["a"])
			assert.Equal(t, "echo://echo", props["url"])
		}
	}
}

func TestToURL(t *testing.T) {
	e := getEcho(t)
	if assert.NotNil(t, e) {
		url, props, err := e.ToURL(map[string]interface{}{"a": "b"})
		if assert.NoError(t, err) {
			assert.Equal(t, "echo://echo", url)
			assert.Len(t, props, 1)
			assert.Equal(t, "b", props["a"])
		}
	}
}

func TestGetParameters(t *testing.T) {
	e := getEcho(t)
	if assert.NotNil(t, e) {
		props, err := e.GetParameters(map[string]interface{}{"a": "b", "c": 4})
		if assert.NoError(t, err) {
			assert.Len(t, props, 2)
			assert.Equal(t, "b", props["a"])
			assert.Equal(t, 4.0, props["c"])
		}
	}
}

func TestValidateRemote(t *testing.T) {
	e := getEcho(t)
	if assert.NotNil(t, e) {
		err := e.ValidateRemote(map[string]interface{}{})
		assert.NoError(t, err)
	}
}

func TestValidateParameters(t *testing.T) {
	e := getEcho(t)
	if assert.NotNil(t, e) {
		err := e.ValidateRemote(map[string]interface{}{})
		assert.NoError(t, err)
	}
}

func TestListCommits(t *testing.T) {
	e := getEcho(t)
	if assert.NotNil(t, e) {
		commits, err := e.ListCommits(map[string]interface{}{}, map[string]interface{}{}, []Tag{})
		if assert.NoError(t, err) {
			assert.Len(t, commits, 2)
			assert.Equal(t, "two", commits[0].ID)
			assert.Equal(t, "two", commits[0].Properties["tags"].(map[string]interface{})["name"])
			assert.Equal(t, "one", commits[1].ID)
			assert.Equal(t, "one", commits[1].Properties["tags"].(map[string]interface{})["name"])
		}
	}
}

func TestListCommitsFilter(t *testing.T) {
	e := getEcho(t)
	if assert.NotNil(t, e) {
		search := "one"

		commits, err := e.ListCommits(map[string]interface{}{}, map[string]interface{}{}, []Tag{{Key: "name", Value: &search}})
		if assert.NoError(t, err) {
			assert.Len(t, commits, 1)
			assert.Equal(t, "one", commits[0].ID)
		}
	}
}

func TestGetCommit(t *testing.T) {
	e := getEcho(t)
	if assert.NotNil(t, e) {
		commit, err := e.GetCommit(map[string]interface{}{}, map[string]interface{}{}, "echo")
		if assert.NoError(t, err) {
			assert.Equal(t, "echo", commit.ID)
		}
	}
}

func TestGetMissingCommit(t *testing.T) {
	e := getEcho(t)
	if assert.NotNil(t, e) {
		commit, err := e.GetCommit(map[string]interface{}{}, map[string]interface{}{}, "foo")
		if assert.NoError(t, err) {
			assert.Nil(t, commit)
		}
	}
}
