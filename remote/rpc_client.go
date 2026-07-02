// Copyright Dit 2026
// SPDX-License-Identifier: BUSL-1.1

// Package remote provides the core remote plugin infrastructure for Dit.
package remote

import (
	"context"

	"github.com/ditdotdev/remote-sdk-go/internal/conversion"
	proto "github.com/ditdotdev/remote-sdk-go/internal/proto"
)

type remoteRPCClient struct {
	Client proto.RemoteClient
}

func (r remoteRPCClient) Type() (string, error) {
	req := proto.GetTypeRequest{}

	res, err := r.Client.GetType(context.Background(), &req)
	if err != nil {
		return "", err
	}

	return res.Type, nil
}

func (r remoteRPCClient) FromURL(url string, properties map[string]string) (map[string]interface{}, error) {
	req := proto.FromURLRequest{Url: url, Properties: properties}

	res, err := r.Client.FromURL(context.Background(), &req)
	if err != nil {
		return nil, err
	}

	output, err := conversion.Struct2Map(res.Remote)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (r remoteRPCClient) ToURL(properties map[string]interface{}) (string, map[string]string, error) {
	s, err := conversion.Map2Struct(properties)
	if err != nil {
		return "", nil, err
	}

	req := proto.ToURLRequest{Remote: s}

	res, err := r.Client.ToURL(context.Background(), &req)
	if err != nil {
		return "", nil, err
	}

	return res.Url, res.Properties, nil
}

func (r remoteRPCClient) GetParameters(properties map[string]interface{}) (map[string]interface{}, error) {
	p, err := conversion.Map2Struct(properties)
	if err != nil {
		return nil, err
	}

	req := proto.GetParametersRequest{Remote: p}

	res, err := r.Client.GetParameters(context.Background(), &req)
	if err != nil {
		return nil, err
	}

	return conversion.Struct2Map(res.Parameters)
}

func (r remoteRPCClient) ValidateRemote(properties map[string]interface{}) error {
	p, err := conversion.Map2Struct(properties)
	if err != nil {
		return err
	}

	req := proto.ValidateRemoteRequest{Remote: p}
	_, err = r.Client.ValidateRemote(context.Background(), &req)

	return err
}

func (r remoteRPCClient) ValidateParameters(parameters map[string]interface{}) error {
	p, err := conversion.Map2Struct(parameters)
	if err != nil {
		return err
	}

	req := proto.ValidateParametersRequest{Parameters: p}
	_, err = r.Client.ValidateParameters(context.Background(), &req)

	return err
}

func (r remoteRPCClient) ListCommits(properties map[string]interface{}, parameters map[string]interface{}, tags []Tag) ([]Commit, error) {
	remote, err := conversion.Map2Struct(properties)
	if err != nil {
		return nil, err
	}

	params, err := conversion.Map2Struct(parameters)
	if err != nil {
		return nil, err
	}

	rpcTags := make([]*proto.Tag, len(tags))
	for i, t := range tags {
		if t.Value != nil {
			rpcTags[i] = &proto.Tag{
				Key:   t.Key,
				Value: &proto.Tag_ValueString{ValueString: *t.Value},
			}
		} else {
			rpcTags[i] = &proto.Tag{
				Key:   t.Key,
				Value: &proto.Tag_ValueNull{ValueNull: true},
			}
		}
	}

	input := proto.ListCommitRequest{
		Remote:     remote,
		Parameters: params,
		Tags:       rpcTags,
	}

	res, err := r.Client.ListCommits(context.Background(), &input)
	if err != nil {
		return nil, err
	}

	nativeCommits := make([]Commit, len(res.Commits))
	for i, c := range res.Commits {
		props, err := conversion.Struct2Map(c.Properties)
		if err != nil {
			return nil, err
		}

		nativeCommits[i] = Commit{
			ID:         c.Id,
			Properties: props,
		}
	}

	return nativeCommits, nil
}

func (r remoteRPCClient) GetCommit(properties map[string]interface{}, parameters map[string]interface{}, commitID string) (*Commit, error) {
	remote, err := conversion.Map2Struct(properties)
	if err != nil {
		return nil, err
	}

	params, err := conversion.Map2Struct(parameters)
	if err != nil {
		return nil, err
	}

	input := proto.GetCommitRequest{
		Remote:     remote,
		Parameters: params,
		CommitId:   commitID,
	}

	res, err := r.Client.GetCommit(context.Background(), &input)
	if err != nil {
		return nil, err
	}

	if res.GetCommitNull() {
		return nil, nil
	}

	props, err := conversion.Struct2Map(res.GetCommitValue().Properties)
	if err != nil {
		return nil, err
	}

	return &Commit{ID: commitID, Properties: props}, nil
}
