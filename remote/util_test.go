/*
 * Copyright The Titan Project Contributors.
 */
package remote

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func registerDefaultRemote() *MockRemote {
	Clear()

	r := new(MockRemote)
	r.On("Type").Return("mock")
	r.On("FromURL", mock.Anything, mock.Anything).Return(map[string]interface{}{}, nil)
	Register(r)

	return r
}

func TestProvider(t *testing.T) {
	r := registerDefaultRemote()
	provider, _, _, _, _ := ParseURL("mock", map[string]string{})
	assert.Equal(t, "mock", provider)
	r.AssertExpectations(t)
}

func TestBadProvider(t *testing.T) {
	_ = registerDefaultRemote()
	_, _, _, _, err := ParseURL("notmock", map[string]string{})
	assert.NotNil(t, err)
}

func TestBadURL(t *testing.T) {
	_ = registerDefaultRemote()
	_, _, _, _, err := ParseURL("a\000b", map[string]string{})
	assert.NotNil(t, err)
}

func TestProviderScheme(t *testing.T) {
	r := registerDefaultRemote()
	provider, _, _, _, _ := ParseURL("mock://foo", map[string]string{})
	assert.Equal(t, "mock", provider)
	r.AssertExpectations(t)
}

func TestFragment(t *testing.T) {
	r := registerDefaultRemote()
	_, _, _, commit, _ := ParseURL("mock://foo#commit", map[string]string{})
	assert.Equal(t, "commit", commit)
	assert.Equal(t, "mock://foo", r.u)
	r.AssertExpectations(t)
}

func TestNoFragment(t *testing.T) {
	r := registerDefaultRemote()
	_, _, _, commit, _ := ParseURL("mock://foo", map[string]string{})
	assert.Empty(t, commit)
	r.AssertExpectations(t)
}

func TestQueryParams(t *testing.T) {
	r := registerDefaultRemote()
	_, _, tags, _, _ := ParseURL("mock://foo?tag=one&tag=two=three", map[string]string{})
	assert.Len(t, tags, 2)
	assert.Equal(t, "one", tags[0])
	assert.Equal(t, "two=three", tags[1])
	assert.Equal(t, "mock://foo", r.u)
	r.AssertExpectations(t)
}

func TestBadQueryParams(t *testing.T) {
	_ = registerDefaultRemote()
	_, _, _, _, err := ParseURL("mock://foo?nottag=one", map[string]string{})
	assert.NotNil(t, err)
}

func TestEmptyQueryParams(t *testing.T) {
	r := registerDefaultRemote()
	_, _, tags, _, _ := ParseURL("mock://foo", map[string]string{})
	assert.Empty(t, tags)
	r.AssertExpectations(t)
}

func TestProperties(t *testing.T) {
	Clear()

	r := new(MockRemote)
	r.On("Type").Return("mock")
	r.On("FromURL", mock.Anything, mock.Anything).Return(map[string]interface{}{"a": "b"}, nil)
	Register(r)

	_, props, _, _, _ := ParseURL("mock://foo", map[string]string{})
	assert.Len(t, props, 1)
	assert.Equal(t, "b", props["a"])
	r.AssertExpectations(t)
}

func TestBadProperties(t *testing.T) {
	Clear()

	r := new(MockRemote)
	r.On("Type").Return("mock")
	r.On("FromURL", mock.Anything, mock.Anything).Return(map[string]interface{}{}, errors.New("error"))
	Register(r)

	_, _, _, _, err := ParseURL("mock://foo", map[string]string{})
	assert.NotNil(t, err)
	r.AssertExpectations(t)
}

func TestProviderArguments(t *testing.T) {
	Clear()

	r := new(MockRemote)
	r.On("Type").Return("mock")
	r.On("FromURL", mock.Anything, mock.Anything).Return(map[string]interface{}{}, nil)
	Register(r)

	_, _, _, _, _ = ParseURL("mock://user:pass@host:80/path", map[string]string{"a": "b"})

	assert.Len(t, r.props, 1)
	assert.Equal(t, "b", r.props["a"])
	assert.Equal(t, "mock://user:pass@host:80/path", r.u)
	r.AssertExpectations(t)
}

func makeCommit(props map[string]string) map[string]interface{} {
	if len(props) == 0 {
		return map[string]interface{}{}
	}
	
	return map[string]interface{}{"tags": props}
}

func existTag(key string) Tag {
	return Tag{Key: key, Value: nil}
}

func exactTag(key string, value string) Tag {
	return Tag{Key: key, Value: &value}
}

func TestMatchEmpty(t *testing.T) {
	assert.True(t, MatchTags(makeCommit(map[string]string{}), []Tag{}))
	assert.True(t, MatchTags(makeCommit(map[string]string{"a": "b"}), []Tag{}))
	assert.True(t, MatchTags(makeCommit(map[string]string{"c": "d"}), []Tag{}))
}

func TestMatchExact(t *testing.T) {
	tags := []Tag{exactTag("a", "b")}
	assert.False(t, MatchTags(makeCommit(map[string]string{}), tags))
	assert.True(t, MatchTags(makeCommit(map[string]string{"a": "b"}), tags))
	assert.False(t, MatchTags(makeCommit(map[string]string{"a": "B"}), tags))
	assert.False(t, MatchTags(makeCommit(map[string]string{"c": "d"}), tags))
}

func TestMatchInterface(t *testing.T) {
	tags := []Tag{exactTag("a", "b")}
	assert.True(t, MatchTags(map[string]interface{}{"tags": map[string]interface{}{"a": "b"}}, tags))
}

func TestMatchExistence(t *testing.T) {
	tags := []Tag{existTag("a")}
	assert.False(t, MatchTags(makeCommit(map[string]string{}), tags))
	assert.True(t, MatchTags(makeCommit(map[string]string{"a": "b"}), tags))
	assert.False(t, MatchTags(makeCommit(map[string]string{"c": "d"}), tags))
}

func TestMatchMultiple(t *testing.T) {
	tags := []Tag{existTag("a"), exactTag("c", "d")}
	assert.False(t, MatchTags(makeCommit(map[string]string{}), tags))
	assert.False(t, MatchTags(makeCommit(map[string]string{"a": "b"}), tags))
	assert.True(t, MatchTags(makeCommit(map[string]string{"a": "b", "c": "d"}), tags))
}

func TestValidateFields(t *testing.T) {
	assert.NoError(t, ValidateFields(map[string]interface{}{"a": "A", "b": "B"}, []string{"a"}, []string{"b"}))
}

func TestValidateMissingRequired(t *testing.T) {
	assert.Error(t, ValidateFields(map[string]interface{}{}, []string{"a"}, []string{}))
}

func TestValidateInvalid(t *testing.T) {
	assert.Error(t, ValidateFields(map[string]interface{}{"c": "C"}, []string{}, []string{"b"}))
}

func TestSortDescending(t *testing.T) {
	commits := []Commit{
		{ID: "four", Properties: map[string]interface{}{"timestamp": "2019-09-21T13:45:30Z"}},
		{ID: "one", Properties: map[string]interface{}{"timestamp": "2019-09-20T13:45:36Z"}},
		{ID: "three", Properties: map[string]interface{}{"timestamp": "2019-09-20T13:45:38Z"}},
		{ID: "two", Properties: map[string]interface{}{"timestamp": "2019-09-20T13:45:37Z"}},
	}
	SortCommits(commits)
	assert.Len(t, commits, 4)
	assert.Equal(t, "four", commits[0].ID)
	assert.Equal(t, "three", commits[1].ID)
	assert.Equal(t, "two", commits[2].ID)
	assert.Equal(t, "one", commits[3].ID)
}

func TestSortMissingTimestamp(t *testing.T) {
	commits := []Commit{
		{ID: "four", Properties: map[string]interface{}{}},
		{ID: "one", Properties: map[string]interface{}{"timestamp": "2019-09-20T13:45:36Z"}},
		{ID: "three", Properties: map[string]interface{}{"timestamp": "2019-09-20T13:45:38Z"}},
		{ID: "two", Properties: map[string]interface{}{"timestamp": "2019-09-20T13:45:37Z"}},
	}
	SortCommits(commits)
	assert.Len(t, commits, 4)
	assert.Equal(t, "three", commits[0].ID)
	assert.Equal(t, "two", commits[1].ID)
	assert.Equal(t, "one", commits[2].ID)
	assert.Equal(t, "four", commits[3].ID)
}

func TestSortBadTimestampFormat(t *testing.T) {
	commits := []Commit{
		{ID: "four", Properties: map[string]interface{}{"timestamp": "foo"}},
		{ID: "one", Properties: map[string]interface{}{"timestamp": "2019-09-20T13:45:36Z"}},
		{ID: "three", Properties: map[string]interface{}{"timestamp": "2019-09-20T13:45:38Z"}},
		{ID: "two", Properties: map[string]interface{}{"timestamp": "2019-09-20T13:45:37Z"}},
	}
	SortCommits(commits)
	assert.Len(t, commits, 4)
	assert.Equal(t, "three", commits[0].ID)
	assert.Equal(t, "two", commits[1].ID)
	assert.Equal(t, "one", commits[2].ID)
	assert.Equal(t, "four", commits[3].ID)
}

func TestSortEmptyTimestamp(t *testing.T) {
	commits := []Commit{
		{ID: "four", Properties: map[string]interface{}{"timestamp": ""}},
		{ID: "one", Properties: map[string]interface{}{"timestamp": "2019-09-20T13:45:36Z"}},
		{ID: "three", Properties: map[string]interface{}{"timestamp": "2019-09-20T13:45:38Z"}},
		{ID: "two", Properties: map[string]interface{}{"timestamp": "2019-09-20T13:45:37Z"}},
	}
	SortCommits(commits)
	assert.Len(t, commits, 4)
	assert.Equal(t, "three", commits[0].ID)
	assert.Equal(t, "two", commits[1].ID)
	assert.Equal(t, "one", commits[2].ID)
	assert.Equal(t, "four", commits[3].ID)
}

func TestSortBadTimestamp(t *testing.T) {
	commits := []Commit{
		{ID: "four", Properties: map[string]interface{}{"timestamp": 4}},
		{ID: "one", Properties: map[string]interface{}{"timestamp": "2019-09-20T13:45:36Z"}},
		{ID: "three", Properties: map[string]interface{}{"timestamp": "2019-09-20T13:45:38Z"}},
		{ID: "two", Properties: map[string]interface{}{"timestamp": "2019-09-20T13:45:37Z"}},
	}
	SortCommits(commits)
	assert.Len(t, commits, 4)
	assert.Equal(t, "three", commits[0].ID)
	assert.Equal(t, "two", commits[1].ID)
	assert.Equal(t, "one", commits[2].ID)
	assert.Equal(t, "four", commits[3].ID)
}
