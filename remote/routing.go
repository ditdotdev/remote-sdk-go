/*
 * Copyright Datadatdat.
 */

// Package remote provides the core remote plugin infrastructure for Datadatdat.
package remote

import (
	"fmt"
	"net/url"
)

// ParseURL wraps remote URL parsing in an easier-to use function that will handle converting to the intermediate URL format,
// processing any query parameters (for tags) and fragment (for commit IDs).
func ParseURL(input string, properties map[string]string) (string, map[string]interface{}, []string, string, error) {
	u, err := url.Parse(input)
	if err != nil {
		return "", nil, nil, "", err
	}

	commit := u.Fragment
	tags := []string{}

	for k := range u.Query() {
		if k != "tag" {
			return "", nil, nil, "", fmt.Errorf("invalid query parameter '%s'", k)
		}
	}

	if u.Query()["tag"] != nil {
		tags = u.Query()["tag"]
	}

	u.RawQuery = ""
	u.Fragment = ""
	urlWithoutQueryAndFragment := u.String()

	// Try to find a provider that can handle this URL by attempting to parse it with each registered provider
	for _, r := range registeredRemotes {
		props, err := r.FromURL(urlWithoutQueryAndFragment, properties)
		if err == nil {
			// This provider successfully parsed the URL
			provider, err := r.Type()
			if err != nil {
				continue
			}
			return provider, props, tags, commit, nil
		}
		// Continue trying other providers if this one failed
	}

	// If no provider could parse the URL, try the legacy scheme-based lookup for backwards compatibility
	var scheme string
	if u.Scheme != "" {
		scheme = u.Scheme
	} else {
		scheme = u.Path
	}

	return "", nil, nil, "", fmt.Errorf("no remote provider found that can handle URI '%s' (tried scheme '%s')", input, scheme)
}
