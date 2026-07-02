// Copyright Dit 2026
// SPDX-License-Identifier: BUSL-1.1

package remote

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	assert.True(t, contains([]string{"a", "b", "c"}, "b"))
	assert.False(t, contains([]string{"a", "b", "c"}, "d"))
	assert.False(t, contains([]string{}, "a"))
}

func TestGetTimestamp(t *testing.T) {
	// Valid RFC3339 timestamp
	ts := getTimestamp("2024-01-15T10:30:00Z")
	expected, _ := time.Parse(time.RFC3339, "2024-01-15T10:30:00Z")
	assert.Equal(t, expected, ts)

	// Empty string returns epoch
	assert.Equal(t, epoch, getTimestamp(""))

	// Invalid format returns epoch
	assert.Equal(t, epoch, getTimestamp("not-a-timestamp"))

	// Non-string returns epoch
	assert.Equal(t, epoch, getTimestamp(12345))
	assert.Equal(t, epoch, getTimestamp(nil))
}

func TestSortCommits(t *testing.T) {
	commits := []Commit{
		{ID: "old", Properties: map[string]interface{}{"timestamp": "2024-01-01T00:00:00Z"}},
		{ID: "new", Properties: map[string]interface{}{"timestamp": "2024-06-01T00:00:00Z"}},
		{ID: "mid", Properties: map[string]interface{}{"timestamp": "2024-03-01T00:00:00Z"}},
	}

	SortCommits(commits)

	assert.Equal(t, "new", commits[0].ID)
	assert.Equal(t, "mid", commits[1].ID)
	assert.Equal(t, "old", commits[2].ID)
}

func TestSortCommitsEmpty(t *testing.T) {
	commits := []Commit{}
	SortCommits(commits)
	assert.Empty(t, commits)
}

func TestSortCommitsMissingTimestamp(t *testing.T) {
	commits := []Commit{
		{ID: "a", Properties: map[string]interface{}{}},
		{ID: "b", Properties: map[string]interface{}{"timestamp": "2024-01-01T00:00:00Z"}},
	}

	SortCommits(commits)

	assert.Equal(t, "b", commits[0].ID)
	assert.Equal(t, "a", commits[1].ID)
}

func TestValidateFields(t *testing.T) {
	props := map[string]interface{}{
		"bucket": "my-bucket",
		"region": "us-east-1",
	}

	// All required present, optional present
	err := ValidateFields(props, []string{"bucket"}, []string{"region"})
	assert.NoError(t, err)

	// Missing required
	err = ValidateFields(props, []string{"bucket", "missing"}, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required property 'missing'")

	// Unknown property
	err = ValidateFields(props, []string{"bucket"}, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid property 'region'")

	// Empty properties with no requirements
	err = ValidateFields(map[string]interface{}{}, []string{}, []string{})
	assert.NoError(t, err)
}

func TestMatchTags(t *testing.T) {
	// No tags always matches
	assert.True(t, MatchTags(map[string]interface{}{}, []Tag{}))

	// Commit with map[string]interface{} tags
	commit := map[string]interface{}{
		"tags": map[string]interface{}{
			"env": "prod",
			"ver": "1.0",
		},
	}

	v := "prod"
	assert.True(t, MatchTags(commit, []Tag{{Key: "env", Value: &v}}))

	wrong := "dev"
	assert.False(t, MatchTags(commit, []Tag{{Key: "env", Value: &wrong}}))

	// Key-only match (nil value)
	assert.True(t, MatchTags(commit, []Tag{{Key: "env", Value: nil}}))

	// Missing key
	assert.False(t, MatchTags(commit, []Tag{{Key: "missing", Value: nil}}))

	// Commit with map[string]string tags
	commit2 := map[string]interface{}{
		"tags": map[string]string{
			"env": "staging",
		},
	}
	staging := "staging"
	assert.True(t, MatchTags(commit2, []Tag{{Key: "env", Value: &staging}}))

	// Commit with no tags field
	assert.False(t, MatchTags(map[string]interface{}{}, []Tag{{Key: "env", Value: nil}}))

	// Commit with wrong type for tags
	assert.False(t, MatchTags(map[string]interface{}{"tags": 123}, []Tag{{Key: "env", Value: nil}}))
}

func TestParseURLNoProviders(t *testing.T) {
	ClearForTesting()
	_, err := ParseURL("s3://bucket/path", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no remote provider found")
}

func TestParseURLInvalidQueryParam(t *testing.T) {
	ClearForTesting()
	_, err := ParseURL("s3://bucket/path?invalid=value", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid query parameter")
}

func TestParseURLInvalidURL(t *testing.T) {
	ClearForTesting()
	_, err := ParseURL("://invalid", nil)
	assert.Error(t, err)
}

func TestParseURLWithTagsAndFragment(t *testing.T) {
	ClearForTesting()

	r := new(MockRemote)
	r.On("Type").Return("mock", nil)
	r.On("FromURL", "mock://bucket/path", map[string]string(nil)).Return(map[string]interface{}{"bucket": "bucket"}, nil)
	Register(r)

	result, err := ParseURL("mock://bucket/path?tag=v1&tag=v2#commit-abc", nil)
	assert.NoError(t, err)
	assert.Equal(t, "mock", result.Provider)
	assert.Equal(t, map[string]interface{}{"bucket": "bucket"}, result.Properties)
	assert.Equal(t, []string{"v1", "v2"}, result.Tags)
	assert.Equal(t, "commit-abc", result.Commit)
}

func TestParseURLNoFragment(t *testing.T) {
	ClearForTesting()

	r := new(MockRemote)
	r.On("Type").Return("mock", nil)
	r.On("FromURL", "mock://bucket/path", map[string]string(nil)).Return(map[string]interface{}{"bucket": "bucket"}, nil)
	Register(r)

	result, err := ParseURL("mock://bucket/path", nil)
	assert.NoError(t, err)
	assert.Equal(t, "mock", result.Provider)
	assert.Empty(t, result.Tags)
	assert.Equal(t, "", result.Commit)
}

func TestParseURLSchemeOnly(t *testing.T) {
	ClearForTesting()
	_, err := ParseURL("justpath", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no remote provider found")
}

func TestParseURLProviderTypeError(t *testing.T) {
	ClearForTesting()

	r := new(MockRemote)
	r.On("Type").Return("mock", nil).Once()
	r.On("FromURL", "mock://bucket", map[string]string(nil)).Return(map[string]interface{}{"bucket": "bucket"}, nil)
	Register(r)

	// After registration, make Type() return an error for the ParseURL call
	r.On("Type").Return("", assert.AnError).Maybe()
	_, err := ParseURL("mock://bucket", nil)
	// Should fall through to "no remote provider found" since Type() errors
	assert.Error(t, err)
}

func TestRegisterPanicsOnTypeError(t *testing.T) {
	ClearForTesting()

	r := new(MockRemote)
	r.On("Type").Return("", assert.AnError)

	assert.Panics(t, func() {
		Register(r)
	})
}
