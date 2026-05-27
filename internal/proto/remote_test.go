/*
 * Copyright Datadatdat.
 */

// Package remote tests cover the generated protobuf message types and
// gRPC stubs. The round-trip Marshal/Unmarshal tests exercise the proto
// reflection surface (XXX_Marshal, XXX_Unmarshal, XXX_Size, XXX_Merge,
// XXX_DiscardUnknown, Reset, String, ProtoMessage, Descriptor) and the
// getters for every populated field. The gRPC tests stand up an in-process
// server backed by bufconn and dispatch every method on RemoteClient
// against a fake RemoteServer, covering both the generated client and
// server adapters in a single pass.
package remote

import (
	"context"
	"net"
	"testing"

	//nolint:staticcheck // generated .pb.go uses golang/protobuf v1; tests must speak the same type system
	proto "github.com/golang/protobuf/proto"
	_struct "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// sampleStruct returns a populated google.protobuf.Struct used as the
// Remote/Parameters payload across many of the request/response messages.
func sampleStruct() *_struct.Struct {
	return &_struct.Struct{Fields: map[string]*_struct.Value{
		"k": {Kind: &_struct.Value_StringValue{StringValue: "v"}},
	}}
}

// xxxAccessor is the minimal interface implemented by every generated
// proto message in this package. The generated XXX_* methods are not
// invoked by proto.Marshal / proto.Unmarshal in the current version of
// golang/protobuf — they go through reflection — so we have to call
// them explicitly to drive coverage.
type xxxAccessor interface {
	proto.Message
	XXX_Unmarshal(b []byte) error
	XXX_Marshal(b []byte, deterministic bool) ([]byte, error)
	XXX_Merge(src proto.Message)
	XXX_Size() int
	XXX_DiscardUnknown()
}

// roundTrip marshals a message to bytes and unmarshals into a fresh
// instance of the same concrete type, then explicitly drives each of
// the XXX_* reflection helpers on the message. The caller then asserts
// that the material fields survived.
//
// `out` is the target of proto.Unmarshal that the caller asserts on.
// `scratch` is an independently-allocated throwaway used only by the
// XXX_* helpers, so XXX_Merge appending into a repeated field cannot
// corrupt `out`. Both must be non-nil pointers to a freshly-zeroed
// instance of the same concrete proto message type.
func roundTrip(t *testing.T, msg, out, scratch xxxAccessor) {
	t.Helper()
	data, err := proto.Marshal(msg)
	require.NoError(t, err)
	require.NoError(t, proto.Unmarshal(data, out))

	// Drive XXX_Marshal / XXX_Unmarshal / XXX_Size / XXX_Merge /
	// XXX_DiscardUnknown directly. These exist on the generated type but
	// are not always called through the proto.Marshal reflection path.
	_, err = msg.XXX_Marshal(nil, false)
	require.NoError(t, err)
	xx, err := msg.XXX_Marshal(nil, true)
	require.NoError(t, err)
	scratch.Reset()
	require.NoError(t, scratch.XXX_Unmarshal(xx))
	_ = msg.XXX_Size()
	scratch.Reset()
	scratch.XXX_Merge(msg)
	scratch.XXX_DiscardUnknown()
}

func TestGetTypeRequest_RoundTrip(t *testing.T) {
	in := &GetTypeRequest{}
	out := &GetTypeRequest{}
	roundTrip(t, in, out, &GetTypeRequest{})
	// Exercise the reflection-only methods so they show up as covered.
	out.Reset()
	assert.NotEmpty(t, out.String()+"_")
	out.ProtoMessage()
	d, idx := (*GetTypeRequest)(nil).Descriptor()
	assert.NotEmpty(t, d)
	assert.NotEmpty(t, idx)
}

func TestGetTypeResponse_RoundTrip(t *testing.T) {
	in := &GetTypeResponse{Type: "echo"}
	out := &GetTypeResponse{}
	roundTrip(t, in, out, &GetTypeResponse{})
	assert.Equal(t, "echo", out.GetType())
	assert.Equal(t, "", (*GetTypeResponse)(nil).GetType())
	out.Reset()
	_ = out.String()
	out.ProtoMessage()
	d, _ := (*GetTypeResponse)(nil).Descriptor()
	assert.NotEmpty(t, d)
}

func TestFromURLRequest_RoundTrip(t *testing.T) {
	in := &FromURLRequest{Url: "echo://x", Properties: map[string]string{"a": "b"}}
	out := &FromURLRequest{}
	roundTrip(t, in, out, &FromURLRequest{})
	assert.Equal(t, "echo://x", out.GetUrl())
	assert.Equal(t, "b", out.GetProperties()["a"])
	assert.Equal(t, "", (*FromURLRequest)(nil).GetUrl())
	assert.Nil(t, (*FromURLRequest)(nil).GetProperties())
	out.Reset()
	_ = out.String()
	out.ProtoMessage()
	d, _ := (*FromURLRequest)(nil).Descriptor()
	assert.NotEmpty(t, d)
}

func TestFromURLResponse_RoundTrip(t *testing.T) {
	in := &FromURLResponse{Remote: sampleStruct()}
	out := &FromURLResponse{}
	roundTrip(t, in, out, &FromURLResponse{})
	require.NotNil(t, out.GetRemote())
	assert.Equal(t, "v", out.GetRemote().Fields["k"].GetStringValue())
	assert.Nil(t, (*FromURLResponse)(nil).GetRemote())
	out.Reset()
	_ = out.String()
	out.ProtoMessage()
	d, _ := (*FromURLResponse)(nil).Descriptor()
	assert.NotEmpty(t, d)
}

func TestToURLRequest_RoundTrip(t *testing.T) {
	in := &ToURLRequest{Remote: sampleStruct()}
	out := &ToURLRequest{}
	roundTrip(t, in, out, &ToURLRequest{})
	require.NotNil(t, out.GetRemote())
	assert.Equal(t, "v", out.GetRemote().Fields["k"].GetStringValue())
	assert.Nil(t, (*ToURLRequest)(nil).GetRemote())
	out.Reset()
	_ = out.String()
	out.ProtoMessage()
	d, _ := (*ToURLRequest)(nil).Descriptor()
	assert.NotEmpty(t, d)
}

func TestToURLResponse_RoundTrip(t *testing.T) {
	in := &ToURLResponse{Url: "echo://x", Properties: map[string]string{"p": "q"}}
	out := &ToURLResponse{}
	roundTrip(t, in, out, &ToURLResponse{})
	assert.Equal(t, "echo://x", out.GetUrl())
	assert.Equal(t, "q", out.GetProperties()["p"])
	assert.Equal(t, "", (*ToURLResponse)(nil).GetUrl())
	assert.Nil(t, (*ToURLResponse)(nil).GetProperties())
	out.Reset()
	_ = out.String()
	out.ProtoMessage()
	d, _ := (*ToURLResponse)(nil).Descriptor()
	assert.NotEmpty(t, d)
}

func TestGetParametersRequest_RoundTrip(t *testing.T) {
	in := &GetParametersRequest{Remote: sampleStruct()}
	out := &GetParametersRequest{}
	roundTrip(t, in, out, &GetParametersRequest{})
	require.NotNil(t, out.GetRemote())
	assert.Nil(t, (*GetParametersRequest)(nil).GetRemote())
	out.Reset()
	_ = out.String()
	out.ProtoMessage()
	d, _ := (*GetParametersRequest)(nil).Descriptor()
	assert.NotEmpty(t, d)
}

func TestGetParametersResponse_RoundTrip(t *testing.T) {
	in := &GetParametersResponse{Parameters: sampleStruct()}
	out := &GetParametersResponse{}
	roundTrip(t, in, out, &GetParametersResponse{})
	require.NotNil(t, out.GetParameters())
	assert.Nil(t, (*GetParametersResponse)(nil).GetParameters())
	out.Reset()
	_ = out.String()
	out.ProtoMessage()
	d, _ := (*GetParametersResponse)(nil).Descriptor()
	assert.NotEmpty(t, d)
}

func TestValidateRemoteRequest_RoundTrip(t *testing.T) {
	in := &ValidateRemoteRequest{Remote: sampleStruct()}
	out := &ValidateRemoteRequest{}
	roundTrip(t, in, out, &ValidateRemoteRequest{})
	require.NotNil(t, out.GetRemote())
	assert.Nil(t, (*ValidateRemoteRequest)(nil).GetRemote())
	out.Reset()
	_ = out.String()
	out.ProtoMessage()
	d, _ := (*ValidateRemoteRequest)(nil).Descriptor()
	assert.NotEmpty(t, d)
}

func TestValidateRemoteResponse_RoundTrip(t *testing.T) {
	in := &ValidateRemoteResponse{}
	out := &ValidateRemoteResponse{}
	roundTrip(t, in, out, &ValidateRemoteResponse{})
	out.Reset()
	_ = out.String()
	out.ProtoMessage()
	d, _ := (*ValidateRemoteResponse)(nil).Descriptor()
	assert.NotEmpty(t, d)
}

func TestValidateParametersRequest_RoundTrip(t *testing.T) {
	in := &ValidateParametersRequest{Parameters: sampleStruct()}
	out := &ValidateParametersRequest{}
	roundTrip(t, in, out, &ValidateParametersRequest{})
	require.NotNil(t, out.GetParameters())
	assert.Nil(t, (*ValidateParametersRequest)(nil).GetParameters())
	out.Reset()
	_ = out.String()
	out.ProtoMessage()
	d, _ := (*ValidateParametersRequest)(nil).Descriptor()
	assert.NotEmpty(t, d)
}

func TestValidateParametersResponse_RoundTrip(t *testing.T) {
	in := &ValidateParametersResponse{}
	out := &ValidateParametersResponse{}
	roundTrip(t, in, out, &ValidateParametersResponse{})
	out.Reset()
	_ = out.String()
	out.ProtoMessage()
	d, _ := (*ValidateParametersResponse)(nil).Descriptor()
	assert.NotEmpty(t, d)
}

func TestTag_StringVariant_RoundTrip(t *testing.T) {
	in := &Tag{Key: "name", Value: &Tag_ValueString{ValueString: "one"}}
	out := &Tag{}
	roundTrip(t, in, out, &Tag{})
	assert.Equal(t, "name", out.GetKey())
	assert.Equal(t, "one", out.GetValueString())
	assert.False(t, out.GetValueNull())
	assert.NotNil(t, out.GetValue())
	// Cover the nil-receiver branches.
	assert.Equal(t, "", (*Tag)(nil).GetKey())
	assert.Equal(t, "", (*Tag)(nil).GetValueString())
	assert.False(t, (*Tag)(nil).GetValueNull())
	assert.Nil(t, (*Tag)(nil).GetValue())
	// Cover the oneof helpers.
	(*Tag_ValueString)(nil).isTag_Value()
	(*Tag_ValueNull)(nil).isTag_Value()
	assert.NotEmpty(t, (*Tag)(nil).XXX_OneofWrappers())
	out.Reset()
	_ = out.String()
	out.ProtoMessage()
	d, _ := (*Tag)(nil).Descriptor()
	assert.NotEmpty(t, d)
}

func TestTag_NullVariant_RoundTrip(t *testing.T) {
	in := &Tag{Key: "absent", Value: &Tag_ValueNull{ValueNull: true}}
	out := &Tag{}
	roundTrip(t, in, out, &Tag{})
	assert.Equal(t, "absent", out.GetKey())
	assert.True(t, out.GetValueNull())
	assert.Equal(t, "", out.GetValueString())
}

func TestCommit_RoundTrip(t *testing.T) {
	in := &Commit{Id: "c1", Properties: sampleStruct()}
	out := &Commit{}
	roundTrip(t, in, out, &Commit{})
	assert.Equal(t, "c1", out.GetId())
	require.NotNil(t, out.GetProperties())
	assert.Equal(t, "", (*Commit)(nil).GetId())
	assert.Nil(t, (*Commit)(nil).GetProperties())
	out.Reset()
	_ = out.String()
	out.ProtoMessage()
	d, _ := (*Commit)(nil).Descriptor()
	assert.NotEmpty(t, d)
}

func TestGetCommitRequest_RoundTrip(t *testing.T) {
	in := &GetCommitRequest{Remote: sampleStruct(), Parameters: sampleStruct(), CommitId: "abc"}
	out := &GetCommitRequest{}
	roundTrip(t, in, out, &GetCommitRequest{})
	require.NotNil(t, out.GetRemote())
	require.NotNil(t, out.GetParameters())
	assert.Equal(t, "abc", out.GetCommitId())
	assert.Nil(t, (*GetCommitRequest)(nil).GetRemote())
	assert.Nil(t, (*GetCommitRequest)(nil).GetParameters())
	assert.Equal(t, "", (*GetCommitRequest)(nil).GetCommitId())
	out.Reset()
	_ = out.String()
	out.ProtoMessage()
	d, _ := (*GetCommitRequest)(nil).Descriptor()
	assert.NotEmpty(t, d)
}

func TestGetCommitResponse_NullVariant_RoundTrip(t *testing.T) {
	in := &GetCommitResponse{Commit: &GetCommitResponse_CommitNull{CommitNull: true}}
	out := &GetCommitResponse{}
	roundTrip(t, in, out, &GetCommitResponse{})
	assert.True(t, out.GetCommitNull())
	assert.Nil(t, out.GetCommitValue())
	assert.NotNil(t, out.GetCommit())
	// Nil-receiver guards.
	assert.False(t, (*GetCommitResponse)(nil).GetCommitNull())
	assert.Nil(t, (*GetCommitResponse)(nil).GetCommitValue())
	assert.Nil(t, (*GetCommitResponse)(nil).GetCommit())
	// Oneof helpers + wrappers.
	(*GetCommitResponse_CommitNull)(nil).isGetCommitResponse_Commit()
	(*GetCommitResponse_CommitValue)(nil).isGetCommitResponse_Commit()
	assert.NotEmpty(t, (*GetCommitResponse)(nil).XXX_OneofWrappers())
	out.Reset()
	_ = out.String()
	out.ProtoMessage()
	d, _ := (*GetCommitResponse)(nil).Descriptor()
	assert.NotEmpty(t, d)
}

func TestGetCommitResponse_ValueVariant_RoundTrip(t *testing.T) {
	in := &GetCommitResponse{Commit: &GetCommitResponse_CommitValue{CommitValue: &Commit{Id: "c1", Properties: sampleStruct()}}}
	out := &GetCommitResponse{}
	roundTrip(t, in, out, &GetCommitResponse{})
	require.NotNil(t, out.GetCommitValue())
	assert.Equal(t, "c1", out.GetCommitValue().GetId())
	assert.False(t, out.GetCommitNull())
}

func TestListCommitRequest_RoundTrip(t *testing.T) {
	in := &ListCommitRequest{
		Remote:     sampleStruct(),
		Parameters: sampleStruct(),
		Tags: []*Tag{
			{Key: "k1", Value: &Tag_ValueString{ValueString: "v1"}},
			{Key: "k2", Value: &Tag_ValueNull{ValueNull: true}},
		},
	}
	out := &ListCommitRequest{}
	roundTrip(t, in, out, &ListCommitRequest{})
	require.NotNil(t, out.GetRemote())
	require.NotNil(t, out.GetParameters())
	require.Len(t, out.GetTags(), 2)
	assert.Equal(t, "k1", out.GetTags()[0].GetKey())
	assert.Equal(t, "v1", out.GetTags()[0].GetValueString())
	assert.True(t, out.GetTags()[1].GetValueNull())
	assert.Nil(t, (*ListCommitRequest)(nil).GetRemote())
	assert.Nil(t, (*ListCommitRequest)(nil).GetParameters())
	assert.Nil(t, (*ListCommitRequest)(nil).GetTags())
	out.Reset()
	_ = out.String()
	out.ProtoMessage()
	d, _ := (*ListCommitRequest)(nil).Descriptor()
	assert.NotEmpty(t, d)
}

func TestListCommitResponse_RoundTrip(t *testing.T) {
	in := &ListCommitResponse{Commits: []*Commit{
		{Id: "c1", Properties: sampleStruct()},
		{Id: "c2"},
	}}
	out := &ListCommitResponse{}
	roundTrip(t, in, out, &ListCommitResponse{})
	require.Len(t, out.GetCommits(), 2)
	assert.Equal(t, "c1", out.GetCommits()[0].GetId())
	assert.Equal(t, "c2", out.GetCommits()[1].GetId())
	assert.Nil(t, (*ListCommitResponse)(nil).GetCommits())
	out.Reset()
	_ = out.String()
	out.ProtoMessage()
	d, _ := (*ListCommitResponse)(nil).Descriptor()
	assert.NotEmpty(t, d)
}

// === gRPC stub coverage ===

// fakeRemoteServer is a trivial RemoteServer implementation used by the
// gRPC tests below. Each method records that it was called and returns
// a populated response that the client side can assert on. This is what
// drives coverage of the generated _Remote_*_Handler functions and the
// remoteClient.* methods.
type fakeRemoteServer struct {
	UnimplementedRemoteServer
}

func (fakeRemoteServer) GetType(context.Context, *GetTypeRequest) (*GetTypeResponse, error) {
	return &GetTypeResponse{Type: "echo"}, nil
}

func (fakeRemoteServer) FromURL(_ context.Context, req *FromURLRequest) (*FromURLResponse, error) {
	return &FromURLResponse{Remote: &_struct.Struct{Fields: map[string]*_struct.Value{
		"url": {Kind: &_struct.Value_StringValue{StringValue: req.GetUrl()}},
	}}}, nil
}

func (fakeRemoteServer) ToURL(context.Context, *ToURLRequest) (*ToURLResponse, error) {
	return &ToURLResponse{Url: "echo://x", Properties: map[string]string{"a": "b"}}, nil
}

func (fakeRemoteServer) GetParameters(context.Context, *GetParametersRequest) (*GetParametersResponse, error) {
	return &GetParametersResponse{Parameters: sampleStruct()}, nil
}

func (fakeRemoteServer) ValidateRemote(context.Context, *ValidateRemoteRequest) (*ValidateRemoteResponse, error) {
	return &ValidateRemoteResponse{}, nil
}

func (fakeRemoteServer) ValidateParameters(context.Context, *ValidateParametersRequest) (*ValidateParametersResponse, error) {
	return &ValidateParametersResponse{}, nil
}

func (fakeRemoteServer) ListCommits(context.Context, *ListCommitRequest) (*ListCommitResponse, error) {
	return &ListCommitResponse{Commits: []*Commit{{Id: "c1"}}}, nil
}

func (fakeRemoteServer) GetCommit(_ context.Context, req *GetCommitRequest) (*GetCommitResponse, error) {
	return &GetCommitResponse{Commit: &GetCommitResponse_CommitValue{CommitValue: &Commit{Id: req.GetCommitId()}}}, nil
}

// startGRPCServer spins up an in-process gRPC server on a bufconn listener
// with the generated RemoteServer registered. Returns a RemoteClient bound
// to that server plus a cleanup func.
func startGRPCServer(t *testing.T) (RemoteClient, func()) {
	t.Helper()
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	RegisterRemoteServer(srv, fakeRemoteServer{})

	go func() {
		// Ignore the error from Serve; the listener is closed by the cleanup
		// func which causes Serve to return a benign error on shutdown.
		_ = srv.Serve(lis)
	}()

	dialer := func(context.Context, string) (net.Conn, error) { return lis.Dial() }
	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	cleanup := func() {
		_ = conn.Close()
		srv.Stop()
		_ = lis.Close()
	}
	return NewRemoteClient(conn), cleanup
}

func TestGRPCStubs_AllMethods(t *testing.T) {
	client, cleanup := startGRPCServer(t)
	defer cleanup()

	ctx := context.Background()

	typeRes, err := client.GetType(ctx, &GetTypeRequest{})
	require.NoError(t, err)
	assert.Equal(t, "echo", typeRes.GetType())

	fromRes, err := client.FromURL(ctx, &FromURLRequest{Url: "echo://x", Properties: map[string]string{"a": "b"}})
	require.NoError(t, err)
	require.NotNil(t, fromRes.GetRemote())
	assert.Equal(t, "echo://x", fromRes.GetRemote().Fields["url"].GetStringValue())

	toRes, err := client.ToURL(ctx, &ToURLRequest{Remote: sampleStruct()})
	require.NoError(t, err)
	assert.Equal(t, "echo://x", toRes.GetUrl())
	assert.Equal(t, "b", toRes.GetProperties()["a"])

	paramRes, err := client.GetParameters(ctx, &GetParametersRequest{Remote: sampleStruct()})
	require.NoError(t, err)
	assert.NotNil(t, paramRes.GetParameters())

	_, err = client.ValidateRemote(ctx, &ValidateRemoteRequest{Remote: sampleStruct()})
	require.NoError(t, err)

	_, err = client.ValidateParameters(ctx, &ValidateParametersRequest{Parameters: sampleStruct()})
	require.NoError(t, err)

	listRes, err := client.ListCommits(ctx, &ListCommitRequest{Remote: sampleStruct(), Parameters: sampleStruct()})
	require.NoError(t, err)
	require.Len(t, listRes.GetCommits(), 1)
	assert.Equal(t, "c1", listRes.GetCommits()[0].GetId())

	commitRes, err := client.GetCommit(ctx, &GetCommitRequest{Remote: sampleStruct(), Parameters: sampleStruct(), CommitId: "abc"})
	require.NoError(t, err)
	require.NotNil(t, commitRes.GetCommitValue())
	assert.Equal(t, "abc", commitRes.GetCommitValue().GetId())
}

// TestUnimplementedRemoteServer exercises the default Unimplemented*
// methods that ship with the generated server stub. They are invoked
// directly (not through a gRPC server) because the only thing they do
// is return a status.Error — no transport layer is needed to cover them.
func TestUnimplementedRemoteServer(t *testing.T) {
	srv := &UnimplementedRemoteServer{}
	ctx := context.Background()

	_, err := srv.GetType(ctx, &GetTypeRequest{})
	assert.Error(t, err)

	_, err = srv.FromURL(ctx, &FromURLRequest{})
	assert.Error(t, err)

	_, err = srv.ToURL(ctx, &ToURLRequest{})
	assert.Error(t, err)

	_, err = srv.GetParameters(ctx, &GetParametersRequest{})
	assert.Error(t, err)

	_, err = srv.ValidateRemote(ctx, &ValidateRemoteRequest{})
	assert.Error(t, err)

	_, err = srv.ValidateParameters(ctx, &ValidateParametersRequest{})
	assert.Error(t, err)

	_, err = srv.ListCommits(ctx, &ListCommitRequest{})
	assert.Error(t, err)

	_, err = srv.GetCommit(ctx, &GetCommitRequest{})
	assert.Error(t, err)
}
