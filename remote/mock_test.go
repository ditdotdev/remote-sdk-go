// Copyright Dit 2026
// SPDX-License-Identifier: BUSL-1.1

package remote

import (
	"github.com/stretchr/testify/mock"
)

type MockRemote struct {
	mock.Mock

	u     string
	props map[string]string
}

func (r *MockRemote) Type() (string, error) {
	args := r.Called()
	return args.String(0), args.Error(1)
}

func (r *MockRemote) FromURL(url string, additionalProperties map[string]string) (map[string]interface{}, error) {
	r.u = url
	r.props = additionalProperties

	args := r.Called(url, additionalProperties)

	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (r *MockRemote) ToURL(properties map[string]interface{}) (string, map[string]string, error) {
	args := r.Called(properties)
	return args.String(0), args.Get(1).(map[string]string), args.Error(2)
}

func (r *MockRemote) GetParameters(remoteProperties map[string]interface{}) (map[string]interface{}, error) {
	args := r.Called(remoteProperties)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (r *MockRemote) ValidateRemote(properties map[string]interface{}) error {
	args := r.Called(properties)
	return args.Error(0)
}

func (r *MockRemote) ValidateParameters(parameters map[string]interface{}) error {
	args := r.Called(parameters)
	return args.Error(0)
}

func (r *MockRemote) ListCommits(properties map[string]interface{}, parameters map[string]interface{}, tags []Tag) ([]Commit, error) {
	args := r.Called(properties, parameters, tags)
	return args.Get(0).([]Commit), args.Error(1)
}

func (r *MockRemote) GetCommit(properties map[string]interface{}, parameters map[string]interface{}, commitID string) (*Commit, error) {
	args := r.Called(properties, parameters, commitID)
	return args.Get(0).(*Commit), args.Error(1)
}
