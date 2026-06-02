/*
 * Copyright Dit.
 */
package remote

import (
	"context"
	"errors"
	"testing"

	proto "github.com/ditdotdev/remote-sdk-go/internal/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

// mockProtoClient implements proto.RemoteClient for testing error paths
type mockProtoClient struct {
	typeResp      *proto.GetTypeResponse
	typeErr       error
	fromURLResp   *proto.FromURLResponse
	fromURLErr    error
	toURLResp     *proto.ToURLResponse
	toURLErr      error
	getParamsResp *proto.GetParametersResponse
	getParamsErr  error
	valRemoteResp *proto.ValidateRemoteResponse
	valRemoteErr  error
	valParamsResp *proto.ValidateParametersResponse
	valParamsErr  error
	listResp      *proto.ListCommitResponse
	listErr       error
	getCommitResp *proto.GetCommitResponse
	getCommitErr  error
}

func (m *mockProtoClient) GetType(_ context.Context, _ *proto.GetTypeRequest, _ ...grpc.CallOption) (*proto.GetTypeResponse, error) {
	return m.typeResp, m.typeErr
}

func (m *mockProtoClient) FromURL(_ context.Context, _ *proto.FromURLRequest, _ ...grpc.CallOption) (*proto.FromURLResponse, error) {
	return m.fromURLResp, m.fromURLErr
}

func (m *mockProtoClient) ToURL(_ context.Context, _ *proto.ToURLRequest, _ ...grpc.CallOption) (*proto.ToURLResponse, error) {
	return m.toURLResp, m.toURLErr
}

func (m *mockProtoClient) GetParameters(_ context.Context, _ *proto.GetParametersRequest, _ ...grpc.CallOption) (*proto.GetParametersResponse, error) {
	return m.getParamsResp, m.getParamsErr
}

func (m *mockProtoClient) ValidateRemote(_ context.Context, _ *proto.ValidateRemoteRequest, _ ...grpc.CallOption) (*proto.ValidateRemoteResponse, error) {
	return m.valRemoteResp, m.valRemoteErr
}

func (m *mockProtoClient) ValidateParameters(_ context.Context, _ *proto.ValidateParametersRequest, _ ...grpc.CallOption) (*proto.ValidateParametersResponse, error) {
	return m.valParamsResp, m.valParamsErr
}

func (m *mockProtoClient) ListCommits(_ context.Context, _ *proto.ListCommitRequest, _ ...grpc.CallOption) (*proto.ListCommitResponse, error) {
	return m.listResp, m.listErr
}

func (m *mockProtoClient) GetCommit(_ context.Context, _ *proto.GetCommitRequest, _ ...grpc.CallOption) (*proto.GetCommitResponse, error) {
	return m.getCommitResp, m.getCommitErr
}

func TestClientTypeError(t *testing.T) {
	client := remoteRPCClient{Client: &mockProtoClient{typeErr: errors.New("grpc error")}}
	_, err := client.Type()
	assert.Error(t, err)
}

func TestClientFromURLGRPCError(t *testing.T) {
	client := remoteRPCClient{Client: &mockProtoClient{fromURLErr: errors.New("grpc error")}}
	_, err := client.FromURL("test://url", nil)
	assert.Error(t, err)
}

func TestClientToURLGRPCError(t *testing.T) {
	client := remoteRPCClient{Client: &mockProtoClient{toURLErr: errors.New("grpc error")}}
	_, _, err := client.ToURL(map[string]interface{}{"a": "b"})
	assert.Error(t, err)
}

func TestClientGetParametersGRPCError(t *testing.T) {
	client := remoteRPCClient{Client: &mockProtoClient{getParamsErr: errors.New("grpc error")}}
	_, err := client.GetParameters(map[string]interface{}{"a": "b"})
	assert.Error(t, err)
}

func TestClientValidateRemoteGRPCError(t *testing.T) {
	client := remoteRPCClient{Client: &mockProtoClient{valRemoteErr: errors.New("grpc error")}}
	err := client.ValidateRemote(map[string]interface{}{"a": "b"})
	assert.Error(t, err)
}

func TestClientValidateParametersGRPCError(t *testing.T) {
	client := remoteRPCClient{Client: &mockProtoClient{valParamsErr: errors.New("grpc error")}}
	err := client.ValidateParameters(map[string]interface{}{"a": "b"})
	assert.Error(t, err)
}

func TestClientListCommitsGRPCError(t *testing.T) {
	client := remoteRPCClient{Client: &mockProtoClient{listErr: errors.New("grpc error")}}
	_, err := client.ListCommits(map[string]interface{}{}, map[string]interface{}{}, []Tag{})
	assert.Error(t, err)
}

func TestClientListCommitsWithTags(t *testing.T) {
	s, _ := structpb.NewStruct(map[string]interface{}{"ts": "2024-01-01T00:00:00Z"})
	client := remoteRPCClient{Client: &mockProtoClient{
		listResp: &proto.ListCommitResponse{
			Commits: []*proto.Commit{{Id: "c1", Properties: s}},
		},
	}}
	v := "prod"
	commits, err := client.ListCommits(map[string]interface{}{}, map[string]interface{}{}, []Tag{
		{Key: "env", Value: &v},
		{Key: "all", Value: nil},
	})
	assert.NoError(t, err)
	assert.Len(t, commits, 1)
}

func TestClientGetCommitGRPCError(t *testing.T) {
	client := remoteRPCClient{Client: &mockProtoClient{getCommitErr: errors.New("grpc error")}}
	_, err := client.GetCommit(map[string]interface{}{}, map[string]interface{}{}, "id")
	assert.Error(t, err)
}

func TestClientGetCommitNull(t *testing.T) {
	client := remoteRPCClient{Client: &mockProtoClient{
		getCommitResp: &proto.GetCommitResponse{
			Commit: &proto.GetCommitResponse_CommitNull{CommitNull: true},
		},
	}}
	commit, err := client.GetCommit(map[string]interface{}{}, map[string]interface{}{}, "id")
	assert.NoError(t, err)
	assert.Nil(t, commit)
}

func TestClientGetCommitSuccess(t *testing.T) {
	s, _ := structpb.NewStruct(map[string]interface{}{"key": "val"})
	client := remoteRPCClient{Client: &mockProtoClient{
		getCommitResp: &proto.GetCommitResponse{
			Commit: &proto.GetCommitResponse_CommitValue{
				CommitValue: &proto.Commit{Id: "abc", Properties: s},
			},
		},
	}}
	commit, err := client.GetCommit(map[string]interface{}{}, map[string]interface{}{}, "abc")
	assert.NoError(t, err)
	assert.Equal(t, "abc", commit.ID)
}

// Test client-side Map2Struct conversion errors (unconvertible input)
func TestClientToURLConversionError(t *testing.T) {
	client := remoteRPCClient{Client: &mockProtoClient{}}
	_, _, err := client.ToURL(map[string]interface{}{"bad": make(chan int)})
	assert.Error(t, err)
}

func TestClientGetParametersConversionError(t *testing.T) {
	client := remoteRPCClient{Client: &mockProtoClient{}}
	_, err := client.GetParameters(map[string]interface{}{"bad": make(chan int)})
	assert.Error(t, err)
}

func TestClientValidateRemoteConversionError(t *testing.T) {
	client := remoteRPCClient{Client: &mockProtoClient{}}
	err := client.ValidateRemote(map[string]interface{}{"bad": make(chan int)})
	assert.Error(t, err)
}

func TestClientValidateParametersConversionError(t *testing.T) {
	client := remoteRPCClient{Client: &mockProtoClient{}}
	err := client.ValidateParameters(map[string]interface{}{"bad": make(chan int)})
	assert.Error(t, err)
}

func TestClientListCommitsPropsConversionError(t *testing.T) {
	client := remoteRPCClient{Client: &mockProtoClient{}}
	_, err := client.ListCommits(map[string]interface{}{"bad": make(chan int)}, map[string]interface{}{}, []Tag{})
	assert.Error(t, err)
}

func TestClientListCommitsParamsConversionError(t *testing.T) {
	client := remoteRPCClient{Client: &mockProtoClient{}}
	_, err := client.ListCommits(map[string]interface{}{}, map[string]interface{}{"bad": make(chan int)}, []Tag{})
	assert.Error(t, err)
}

func TestClientGetCommitPropsConversionError(t *testing.T) {
	client := remoteRPCClient{Client: &mockProtoClient{}}
	_, err := client.GetCommit(map[string]interface{}{"bad": make(chan int)}, map[string]interface{}{}, "id")
	assert.Error(t, err)
}

func TestClientGetCommitParamsConversionError(t *testing.T) {
	client := remoteRPCClient{Client: &mockProtoClient{}}
	_, err := client.GetCommit(map[string]interface{}{}, map[string]interface{}{"bad": make(chan int)}, "id")
	assert.Error(t, err)
}

// badStruct returns a *structpb.Struct whose Fields contain a Value with no
// Kind set. Struct2Map errors on this input (after the util.go:63 fix);
// before the fix, the error was silently returned in the value position.
// Used to exercise the response-side Struct2Map error returns in rpc_client.
func badStruct() *structpb.Struct {
	return &structpb.Struct{Fields: map[string]*structpb.Value{
		"x": {}, // nil Kind triggers the elabValue fallthrough
	}}
}

// === Response-side Struct2Map error paths in rpc_client (now reachable) ===

func TestClientFromURLResponseConversionError(t *testing.T) {
	client := remoteRPCClient{Client: &mockProtoClient{
		fromURLResp: &proto.FromURLResponse{Remote: badStruct()},
	}}
	_, err := client.FromURL("test://url", nil)
	assert.Error(t, err)
}

func TestClientGetParametersResponseConversionError(t *testing.T) {
	client := remoteRPCClient{Client: &mockProtoClient{
		getParamsResp: &proto.GetParametersResponse{Parameters: badStruct()},
	}}
	_, err := client.GetParameters(map[string]interface{}{"a": "b"})
	assert.Error(t, err)
}

func TestClientListCommitsResponseConversionError(t *testing.T) {
	client := remoteRPCClient{Client: &mockProtoClient{
		listResp: &proto.ListCommitResponse{
			Commits: []*proto.Commit{{Id: "c1", Properties: badStruct()}},
		},
	}}
	_, err := client.ListCommits(map[string]interface{}{}, map[string]interface{}{}, []Tag{})
	assert.Error(t, err)
}

func TestClientGetCommitResponseConversionError(t *testing.T) {
	client := remoteRPCClient{Client: &mockProtoClient{
		getCommitResp: &proto.GetCommitResponse{
			Commit: &proto.GetCommitResponse_CommitValue{
				CommitValue: &proto.Commit{Id: "abc", Properties: badStruct()},
			},
		},
	}}
	_, err := client.GetCommit(map[string]interface{}{}, map[string]interface{}{}, "abc")
	assert.Error(t, err)
}

// Test server-side conversion errors by passing unconvertible types from Impl
func TestServerFromURLConversionError(t *testing.T) {
	r := new(MockRemote)
	// Return a map with a channel value which Map2Struct cannot convert
	r.On("FromURL", "test://url", map[string]string(nil)).Return(
		map[string]interface{}{"bad": make(chan int)}, nil)

	server := &remoteRPCServer{Impl: r}
	resp, err := server.FromURL(context.Background(), &proto.FromURLRequest{Url: "test://url"})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestServerGetParametersConversionError(t *testing.T) {
	r := new(MockRemote)
	r.On("GetParameters", mock.Anything).Return(
		map[string]interface{}{"bad": make(chan int)}, nil)

	s, _ := structpb.NewStruct(map[string]interface{}{})
	server := &remoteRPCServer{Impl: r}
	resp, err := server.GetParameters(context.Background(), &proto.GetParametersRequest{Remote: s})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestServerListCommitsConversionError(t *testing.T) {
	r := new(MockRemote)
	// Return commits with unconvertible properties
	r.On("ListCommits", mock.Anything, mock.Anything, mock.Anything).Return(
		[]Commit{{ID: "c1", Properties: map[string]interface{}{"bad": make(chan int)}}}, nil)

	s, _ := structpb.NewStruct(map[string]interface{}{})
	server := &remoteRPCServer{Impl: r}
	resp, err := server.ListCommits(context.Background(), &proto.ListCommitRequest{
		Remote: s, Parameters: s, Tags: []*proto.Tag{},
	})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestServerGetCommitConversionError(t *testing.T) {
	r := new(MockRemote)
	r.On("GetCommit", mock.Anything, mock.Anything, "id").Return(
		&Commit{ID: "id", Properties: map[string]interface{}{"bad": make(chan int)}}, nil)

	s, _ := structpb.NewStruct(map[string]interface{}{})
	server := &remoteRPCServer{Impl: r}
	resp, err := server.GetCommit(context.Background(), &proto.GetCommitRequest{
		Remote: s, Parameters: s, CommitId: "id",
	})
	assert.Error(t, err)
	assert.Nil(t, resp)
}
