// Copyright Dit 2026
// SPDX-License-Identifier: BUSL-1.1

// Package remote provides the core remote plugin infrastructure for Datadatdat.
package remote

import (
	"fmt"
	"net/url"
)

// ParseResult is the structured output of ParseURL: the provider that claimed the URL, the parsed
// provider-specific properties, any `?tag=` query parameters, and the URL fragment (interpreted as a commit ID).
//
// Replaces the previous 5-value return signature of ParseURL, which was unreadable at call sites.
type ParseResult struct {
	Provider   string
	Properties map[string]interface{}
	Tags       []string
	Commit     string
}

// ParseURL parses a remote URL using the Default registry. See Registry.ParseURL for details.
func ParseURL(input string, properties map[string]string) (*ParseResult, error) {
	return Default.ParseURL(input, properties)
}

// ParseURL parses a remote URL by handing it to every registered provider in turn and returning the first one
// that accepts it. Query parameters (`?tag=...`) and the URL fragment (`#commit-id`) are stripped before the
// provider sees the URL and returned in the ParseResult instead.
//
// Returns an error if no registered provider accepts the URL, or if the URL has query parameters other than `tag`.
func (r *Registry) ParseURL(input string, properties map[string]string) (*ParseResult, error) {
	u, err := url.Parse(input)
	if err != nil {
		return nil, err
	}

	commit := u.Fragment
	tags := []string{}

	for k := range u.Query() {
		if k != "tag" {
			return nil, fmt.Errorf("invalid query parameter '%s'", k)
		}
	}

	if u.Query()["tag"] != nil {
		tags = u.Query()["tag"]
	}

	u.RawQuery = ""
	u.Fragment = ""
	urlWithoutQueryAndFragment := u.String()

	// Try to find a provider that can handle this URL by attempting to parse it with each registered provider.
	// snapshot() returns a slice copy so we don't hold the registry lock across FromURL/Type callbacks.
	for _, rem := range r.snapshot() {
		props, err := rem.FromURL(urlWithoutQueryAndFragment, properties)
		if err != nil {
			continue
		}
		provider, err := rem.Type()
		if err != nil {
			continue
		}
		return &ParseResult{
			Provider:   provider,
			Properties: props,
			Tags:       tags,
			Commit:     commit,
		}, nil
	}

	// If no provider could parse the URL, surface the scheme we tried — helps users diagnose typos.
	scheme := u.Scheme
	if scheme == "" {
		scheme = u.Path
	}
	return nil, fmt.Errorf("no remote provider found that can handle URI '%s' (tried scheme '%s')", input, scheme)
}
