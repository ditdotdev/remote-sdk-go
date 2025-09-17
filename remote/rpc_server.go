/*
 * Copyright The Titan Project Contributors.
 */
package remote

import (
	"context"
	proto "github.com/datadatdat/remote-sdk-go/internal/proto"
	"github.com/datadatdat/remote-sdk-go/internal/util"
)

type remoteRPCServer struct {
	Impl Remote
}

func (r *remoteRPCServer) GetType(context.Context, *proto.GetTypeRequest) (*proto.GetTypeResponse, error) {
	typ, err := r.Impl.Type()
	if err != nil {
		return nil, err
	}
	return &proto.GetTypeResponse{Type: typ}, nil
}

func (r *remoteRPCServer) FromURL(ctx context.Context, req *proto.FromURLRequest) (*proto.FromURLResponse, error) {
	props, err := r.Impl.FromURL(req.Url, req.Properties)
	if err != nil {
		return nil, err
	}
	output, err := util.Map2Struct(props)
	if err != nil {
		return nil, err
	}
	return &proto.FromURLResponse{Remote: output}, nil
}

func (r *remoteRPCServer) ToURL(ctx context.Context, req *proto.ToURLRequest) (*proto.ToURLResponse, error) {
	input, err := util.Struct2Map(req.Remote)
	if err != nil {
		return nil, err
	}
	url, props, err := r.Impl.ToURL(input)
	if err != nil {
		return nil, err
	}
	ret := &proto.ToURLResponse{Url: url, Properties: props}
	return ret, nil
}

func (r *remoteRPCServer) GetParameters(ctx context.Context, req *proto.GetParametersRequest) (*proto.GetParametersResponse, error) {
	input, err := util.Struct2Map(req.Remote)
	if err != nil {
		return nil, err
	}
	props, err := r.Impl.GetParameters(input)
	if err != nil {
		return nil, err
	}
	output, err := util.Map2Struct(props)
	if err != nil {
		return nil, err
	}
	return &proto.GetParametersResponse{Parameters: output}, nil
}

func (r *remoteRPCServer) ValidateRemote(ctx context.Context, req *proto.ValidateRemoteRequest) (*proto.ValidateRemoteResponse, error) {
	remote, err := util.Struct2Map(req.Remote)
	if err != nil {
		return nil, err
	}
	err = r.Impl.ValidateRemote(remote)
	return &proto.ValidateRemoteResponse{}, err
}

func (r *remoteRPCServer) ValidateParameters(ctx context.Context, req *proto.ValidateParametersRequest) (*proto.ValidateParametersResponse, error) {
	params, err := util.Struct2Map(req.Parameters)
	if err != nil {
		return nil, err
	}
	err = r.Impl.ValidateParameters(params)
	return &proto.ValidateParametersResponse{}, err
}

func (r *remoteRPCServer) ListCommits(ctx context.Context, req *proto.ListCommitRequest) (*proto.ListCommitResponse, error) {
	remote, err := util.Struct2Map(req.Remote)
	if err != nil {
		return nil, err
	}
	params, err := util.Struct2Map(req.Parameters)
	if err != nil {
		return nil, err
	}
	nativeTags := make([]Tag, len(req.Tags))
	for i, t := range req.Tags {
		if t.GetValueNull() {
			nativeTags[i] = Tag{Key: t.Key}
		} else {
			val := t.GetValueString()
			nativeTags[i] = Tag{Key: t.Key, Value: &val}
		}
	}
	commits, err := r.Impl.ListCommits(remote, params, nativeTags)
	if err != nil {
		return nil, err
	}

	rpcCommits := make([]*proto.Commit, len(commits))
	for i, c := range commits {
		props, err := util.Map2Struct(c.Properties)
		if err != nil {
			return nil, err
		}
		rpcCommits[i] = &proto.Commit{
			Id:         c.Id,
			Properties: props,
		}
	}

	return &proto.ListCommitResponse{Commits: rpcCommits}, nil
}

func (r *remoteRPCServer) GetCommit(ctx context.Context, req *proto.GetCommitRequest) (*proto.GetCommitResponse, error) {
	remote, err := util.Struct2Map(req.Remote)
	if err != nil {
		return nil, err
	}
	params, err := util.Struct2Map(req.Parameters)
	if err != nil {
		return nil, err
	}
	commit, err := r.Impl.GetCommit(remote, params, req.CommitId)
	if err != nil {
		return nil, err
	}
	if commit == nil {
		return &proto.GetCommitResponse{
			Commit: &proto.GetCommitResponse_CommitNull{CommitNull: true},
		}, nil
	} else {
		s, err := util.Map2Struct(commit.Properties)
		if err != nil {
			return nil, err
		}
		rpcCommit := proto.Commit{
			Id:         commit.Id,
			Properties: s,
		}
		return &proto.GetCommitResponse{
			Commit: &proto.GetCommitResponse_CommitValue{CommitValue: &rpcCommit},
		}, nil
	}
}
