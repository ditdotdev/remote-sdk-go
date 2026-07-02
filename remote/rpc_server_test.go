// Copyright Dit 2026
// SPDX-License-Identifier: BUSL-1.1
package remote

import (
	"context"
	"errors"
	"testing"

	proto "github.com/ditdotdev/remote-sdk-go/internal/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestServerGetType(t *testing.T) {
	r := new(MockRemote)
	r.On("Type").Return("test", nil)

	server := &remoteRPCServer{Impl: r}
	resp, err := server.GetType(context.Background(), &proto.GetTypeRequest{})

	assert.NoError(t, err)
	assert.Equal(t, "test", resp.Type)
	r.AssertExpectations(t)
}

func TestServerGetTypeError(t *testing.T) {
	r := new(MockRemote)
	r.On("Type").Return("", errors.New("type error"))

	server := &remoteRPCServer{Impl: r}
	resp, err := server.GetType(context.Background(), &proto.GetTypeRequest{})

	assert.Error(t, err)
	assert.Nil(t, resp)
	r.AssertExpectations(t)
}

func TestServerFromURL(t *testing.T) {
	r := new(MockRemote)
	r.On("FromURL", "test://url", map[string]string{"key": "value"}).Return(
		map[string]interface{}{"prop": "val"}, nil)

	server := &remoteRPCServer{Impl: r}
	resp, err := server.FromURL(context.Background(), &proto.FromURLRequest{
		Url:        "test://url",
		Properties: map[string]string{"key": "value"},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp.Remote)
	r.AssertExpectations(t)
}

func TestServerFromURLError(t *testing.T) {
	r := new(MockRemote)
	r.On("FromURL", "bad://url", mock.Anything).Return(
		map[string]interface{}{}, errors.New("fromurl error"))

	server := &remoteRPCServer{Impl: r}
	resp, err := server.FromURL(context.Background(), &proto.FromURLRequest{
		Url:        "bad://url",
		Properties: map[string]string{},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	r.AssertExpectations(t)
}

func TestServerToURL(t *testing.T) {
	r := new(MockRemote)
	r.On("ToURL", mock.Anything).Return("test://url", map[string]string{"key": "value"}, nil)

	propsStruct, _ := structpb.NewStruct(map[string]interface{}{"prop": "val"})
	server := &remoteRPCServer{Impl: r}
	resp, err := server.ToURL(context.Background(), &proto.ToURLRequest{
		Remote: propsStruct,
	})

	assert.NoError(t, err)
	assert.Equal(t, "test://url", resp.Url)
	assert.Equal(t, "value", resp.Properties["key"])
	r.AssertExpectations(t)
}

func TestServerToURLError(t *testing.T) {
	r := new(MockRemote)
	r.On("ToURL", mock.Anything).Return("", map[string]string{}, errors.New("tourl error"))

	propsStruct, _ := structpb.NewStruct(map[string]interface{}{})
	server := &remoteRPCServer{Impl: r}
	resp, err := server.ToURL(context.Background(), &proto.ToURLRequest{
		Remote: propsStruct,
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	r.AssertExpectations(t)
}

func TestServerGetParameters(t *testing.T) {
	r := new(MockRemote)
	r.On("GetParameters", mock.Anything).Return(map[string]interface{}{"param": "value"}, nil)

	propsStruct, _ := structpb.NewStruct(map[string]interface{}{"remote": "prop"})
	server := &remoteRPCServer{Impl: r}
	resp, err := server.GetParameters(context.Background(), &proto.GetParametersRequest{
		Remote: propsStruct,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp.Parameters)
	r.AssertExpectations(t)
}

func TestServerGetParametersError(t *testing.T) {
	r := new(MockRemote)
	r.On("GetParameters", mock.Anything).Return(
		map[string]interface{}{}, errors.New("getparams error"))

	propsStruct, _ := structpb.NewStruct(map[string]interface{}{})
	server := &remoteRPCServer{Impl: r}
	resp, err := server.GetParameters(context.Background(), &proto.GetParametersRequest{
		Remote: propsStruct,
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	r.AssertExpectations(t)
}

func TestServerValidateRemote(t *testing.T) {
	r := new(MockRemote)
	r.On("ValidateRemote", mock.Anything).Return(nil)

	propsStruct, _ := structpb.NewStruct(map[string]interface{}{"valid": "remote"})
	server := &remoteRPCServer{Impl: r}
	resp, err := server.ValidateRemote(context.Background(), &proto.ValidateRemoteRequest{
		Remote: propsStruct,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	r.AssertExpectations(t)
}

func TestServerValidateRemoteError(t *testing.T) {
	r := new(MockRemote)
	r.On("ValidateRemote", mock.Anything).Return(errors.New("validation failed"))

	propsStruct, _ := structpb.NewStruct(map[string]interface{}{})
	server := &remoteRPCServer{Impl: r}
	resp, err := server.ValidateRemote(context.Background(), &proto.ValidateRemoteRequest{
		Remote: propsStruct,
	})

	assert.Error(t, err)
	assert.NotNil(t, resp)
	r.AssertExpectations(t)
}

func TestServerValidateParameters(t *testing.T) {
	r := new(MockRemote)
	r.On("ValidateParameters", mock.Anything).Return(nil)

	paramsStruct, _ := structpb.NewStruct(map[string]interface{}{"valid": "params"})
	server := &remoteRPCServer{Impl: r}
	resp, err := server.ValidateParameters(context.Background(), &proto.ValidateParametersRequest{
		Parameters: paramsStruct,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	r.AssertExpectations(t)
}

func TestServerValidateParametersError(t *testing.T) {
	r := new(MockRemote)
	r.On("ValidateParameters", mock.Anything).Return(errors.New("param validation failed"))

	paramsStruct, _ := structpb.NewStruct(map[string]interface{}{})
	server := &remoteRPCServer{Impl: r}
	resp, err := server.ValidateParameters(context.Background(), &proto.ValidateParametersRequest{
		Parameters: paramsStruct,
	})

	assert.Error(t, err)
	assert.NotNil(t, resp)
	r.AssertExpectations(t)
}

func TestServerListCommits(t *testing.T) {
	r := new(MockRemote)
	expectedCommits := []Commit{
		{ID: "commit1", Properties: map[string]interface{}{"timestamp": "2024-01-01T00:00:00Z"}},
		{ID: "commit2", Properties: map[string]interface{}{"timestamp": "2024-01-02T00:00:00Z"}},
	}
	r.On("ListCommits", mock.Anything, mock.Anything, mock.Anything).Return(expectedCommits, nil)

	remoteStruct, _ := structpb.NewStruct(map[string]interface{}{})
	paramsStruct, _ := structpb.NewStruct(map[string]interface{}{})
	server := &remoteRPCServer{Impl: r}
	resp, err := server.ListCommits(context.Background(), &proto.ListCommitRequest{
		Remote:     remoteStruct,
		Parameters: paramsStruct,
		Tags:       []*proto.Tag{},
	})

	assert.NoError(t, err)
	assert.Len(t, resp.Commits, 2)
	assert.Equal(t, "commit1", resp.Commits[0].Id)
	assert.Equal(t, "commit2", resp.Commits[1].Id)
	r.AssertExpectations(t)
}

func TestServerListCommitsWithTags(t *testing.T) {
	r := new(MockRemote)
	expectedCommits := []Commit{
		{ID: "commit1", Properties: map[string]interface{}{"timestamp": "2024-01-01T00:00:00Z"}},
	}
	r.On("ListCommits", mock.Anything, mock.Anything, mock.MatchedBy(func(tags []Tag) bool {
		return len(tags) == 2 && tags[0].Key == "key1" && tags[1].Key == "key2" && *tags[1].Value == "val2"
	})).Return(expectedCommits, nil)

	remoteStruct, _ := structpb.NewStruct(map[string]interface{}{})
	paramsStruct, _ := structpb.NewStruct(map[string]interface{}{})
	server := &remoteRPCServer{Impl: r}
	resp, err := server.ListCommits(context.Background(), &proto.ListCommitRequest{
		Remote:     remoteStruct,
		Parameters: paramsStruct,
		Tags: []*proto.Tag{
			{Key: "key1", Value: &proto.Tag_ValueNull{ValueNull: true}},
			{Key: "key2", Value: &proto.Tag_ValueString{ValueString: "val2"}},
		},
	})

	assert.NoError(t, err)
	assert.Len(t, resp.Commits, 1)
	r.AssertExpectations(t)
}

func TestServerListCommitsError(t *testing.T) {
	r := new(MockRemote)
	r.On("ListCommits", mock.Anything, mock.Anything, mock.Anything).Return(
		[]Commit{}, errors.New("list error"))

	remoteStruct, _ := structpb.NewStruct(map[string]interface{}{})
	paramsStruct, _ := structpb.NewStruct(map[string]interface{}{})
	server := &remoteRPCServer{Impl: r}
	resp, err := server.ListCommits(context.Background(), &proto.ListCommitRequest{
		Remote:     remoteStruct,
		Parameters: paramsStruct,
		Tags:       []*proto.Tag{},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	r.AssertExpectations(t)
}

func TestServerGetCommit(t *testing.T) {
	r := new(MockRemote)
	expectedCommit := &Commit{
		ID:         "commit123",
		Properties: map[string]interface{}{"timestamp": "2024-01-01T00:00:00Z"},
	}
	r.On("GetCommit", mock.Anything, mock.Anything, "commit123").Return(expectedCommit, nil)

	remoteStruct, _ := structpb.NewStruct(map[string]interface{}{})
	paramsStruct, _ := structpb.NewStruct(map[string]interface{}{})
	server := &remoteRPCServer{Impl: r}
	resp, err := server.GetCommit(context.Background(), &proto.GetCommitRequest{
		Remote:     remoteStruct,
		Parameters: paramsStruct,
		CommitId:   "commit123",
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp.GetCommitValue())
	assert.Equal(t, "commit123", resp.GetCommitValue().Id)
	r.AssertExpectations(t)
}

func TestServerGetCommitNull(t *testing.T) {
	r := new(MockRemote)
	r.On("GetCommit", mock.Anything, mock.Anything, "missing").Return((*Commit)(nil), nil)

	remoteStruct, _ := structpb.NewStruct(map[string]interface{}{})
	paramsStruct, _ := structpb.NewStruct(map[string]interface{}{})
	server := &remoteRPCServer{Impl: r}
	resp, err := server.GetCommit(context.Background(), &proto.GetCommitRequest{
		Remote:     remoteStruct,
		Parameters: paramsStruct,
		CommitId:   "missing",
	})

	assert.NoError(t, err)
	assert.True(t, resp.GetCommitNull())
	r.AssertExpectations(t)
}

func TestServerGetCommitError(t *testing.T) {
	r := new(MockRemote)
	r.On("GetCommit", mock.Anything, mock.Anything, "error").Return(
		(*Commit)(nil), errors.New("get error"))

	remoteStruct, _ := structpb.NewStruct(map[string]interface{}{})
	paramsStruct, _ := structpb.NewStruct(map[string]interface{}{})
	server := &remoteRPCServer{Impl: r}
	resp, err := server.GetCommit(context.Background(), &proto.GetCommitRequest{
		Remote:     remoteStruct,
		Parameters: paramsStruct,
		CommitId:   "error",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	r.AssertExpectations(t)
}

// === Request-side Struct2Map error paths in rpc_server (now reachable) ===
//
// These exercise the conversion.Struct2Map(req.Remote/Parameters) error
// returns in each handler. Before the util.go:63 fix the conversion
// silently returned the error in the value position, so these branches
// were unreachable; after the fix they are correctly reported as errors.

func TestServerToURLConversionError(t *testing.T) {
	server := &remoteRPCServer{Impl: new(MockRemote)}
	resp, err := server.ToURL(context.Background(), &proto.ToURLRequest{Remote: badStruct()})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestServerGetParametersRequestConversionError(t *testing.T) {
	server := &remoteRPCServer{Impl: new(MockRemote)}
	resp, err := server.GetParameters(context.Background(), &proto.GetParametersRequest{Remote: badStruct()})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestServerValidateRemoteConversionError(t *testing.T) {
	server := &remoteRPCServer{Impl: new(MockRemote)}
	resp, err := server.ValidateRemote(context.Background(), &proto.ValidateRemoteRequest{Remote: badStruct()})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestServerValidateParametersConversionError(t *testing.T) {
	server := &remoteRPCServer{Impl: new(MockRemote)}
	resp, err := server.ValidateParameters(context.Background(), &proto.ValidateParametersRequest{Parameters: badStruct()})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestServerListCommitsRemoteConversionError(t *testing.T) {
	server := &remoteRPCServer{Impl: new(MockRemote)}
	good, _ := structpb.NewStruct(map[string]interface{}{})
	resp, err := server.ListCommits(context.Background(), &proto.ListCommitRequest{
		Remote: badStruct(), Parameters: good,
	})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestServerListCommitsParametersConversionError(t *testing.T) {
	server := &remoteRPCServer{Impl: new(MockRemote)}
	good, _ := structpb.NewStruct(map[string]interface{}{})
	resp, err := server.ListCommits(context.Background(), &proto.ListCommitRequest{
		Remote: good, Parameters: badStruct(),
	})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestServerGetCommitRemoteConversionError(t *testing.T) {
	server := &remoteRPCServer{Impl: new(MockRemote)}
	good, _ := structpb.NewStruct(map[string]interface{}{})
	resp, err := server.GetCommit(context.Background(), &proto.GetCommitRequest{
		Remote: badStruct(), Parameters: good, CommitId: "x",
	})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestServerGetCommitParametersConversionError(t *testing.T) {
	server := &remoteRPCServer{Impl: new(MockRemote)}
	good, _ := structpb.NewStruct(map[string]interface{}{})
	resp, err := server.GetCommit(context.Background(), &proto.GetCommitRequest{
		Remote: good, Parameters: badStruct(), CommitId: "x",
	})
	assert.Error(t, err)
	assert.Nil(t, resp)
}
