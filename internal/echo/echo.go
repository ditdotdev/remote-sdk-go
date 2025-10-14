/*
 * Copyright Datadatdat.
 */

// Package echo provides a sample echo remote implementation for testing and demonstration purposes.
package echo

import "github.com/datadatdat/remote-sdk-go/remote"

// Remote is a sample remote implementation that echoes back test data for development and testing.
type Remote struct {
}

// Type returns the remote type identifier for the echo remote.
func (m Remote) Type() (string, error) {
	return "echo", nil
}

// FromURL parses a URL and additional properties to create remote properties for the echo remote.
func (m Remote) FromURL(url string, additionalProperties map[string]string) (map[string]interface{}, error) {
	ret := map[string]interface{}{
		"url": url,
	}
	for k, v := range additionalProperties {
		ret[k] = v
	}

	return ret, nil
}

// ToURL converts remote properties back to a URL and additional properties.
func (m Remote) ToURL(properties map[string]interface{}) (string, map[string]string, error) {
	ret := map[string]string{}
	for k, v := range properties {
		ret[k] = v.(string)
	}

	return "echo://echo", ret, nil
}

// GetParameters extracts operation parameters from remote properties.
func (m Remote) GetParameters(remoteProperties map[string]interface{}) (map[string]interface{}, error) {
	return remoteProperties, nil
}

// ValidateRemote validates the remote connection properties.
func (m Remote) ValidateRemote(_ map[string]interface{}) error {
	return nil
}

// ValidateParameters validates the operation parameters.
func (m Remote) ValidateParameters(_ map[string]interface{}) error {
	return nil
}

// ListCommits returns a list of available commits, optionally filtered by tags.
func (m Remote) ListCommits(_ map[string]interface{}, _ map[string]interface{}, tags []remote.Tag) ([]remote.Commit, error) {
	res := []remote.Commit{{
		ID:         "one",
		Properties: map[string]interface{}{"tags": map[string]interface{}{"name": "one"}, "timestamp": "2019-09-20T13:45:36Z"},
	}, {
		ID:         "two",
		Properties: map[string]interface{}{"tags": map[string]interface{}{"name": "two"}, "timestamp": "2019-09-20T13:45:37Z"},
	}}
	n := 0

	for _, c := range res {
		if remote.MatchTags(c.Properties, tags) {
			res[n] = c
			n++
		}
	}

	res = res[:n]
	remote.SortCommits(res)

	return res, nil
}

// GetCommit retrieves a specific commit by its identifier.
func (m Remote) GetCommit(_ map[string]interface{}, _ map[string]interface{}, commitID string) (*remote.Commit, error) {
	if commitID == "echo" {
		return &remote.Commit{
			ID:         "echo",
			Properties: map[string]interface{}{"name": "echo", "timestamp": "2019-09-20T13:45:36Z"},
		}, nil
	}

	return nil, nil
}
