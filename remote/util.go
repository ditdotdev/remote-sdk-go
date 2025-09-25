/*
 * Copyright The Titan Project Contributors.
 */
package remote

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"time"
)

/*
 * Wrap remote URL parsing in an easier-to use function that will handle converting to the intermediate URL format,
 * processing any query parameters (for tags) and fragment (for commit IDs).
 */
func ParseURL(input string, properties map[string]string) (string, map[string]interface{}, []string, string, error) {
	u, err := url.Parse(input)
	if err != nil {
		return "", nil, nil, "", err
	}

	var provider string
	if u.Scheme != "" {
		provider = u.Scheme
	} else {
		provider = u.Path
	}

	var r = Get(provider)
	if r == nil {
		return "", nil, nil, "", errors.New(fmt.Sprintf("unknown remote provider '%s'", provider))
	}

	commit := u.Fragment
	tags := []string{}
	for k := range u.Query() {
		if k != "tag" {
			return "", nil, nil, "", errors.New(fmt.Sprintf("invalid query parameter '%s'", k))
		}
	}
	if u.Query()["tag"] != nil {
		tags = u.Query()["tag"]
	}

	u.RawQuery = ""
	u.Fragment = ""
	props, err := r.FromURL(u.String(), properties)
	if err != nil {
		return "", nil, nil, "", err
	}

	return provider, props, tags, commit, nil
}

func contains(arr []string, search string) bool {
	for _, v := range arr {
		if v == search {
			return true
		}
	}
	return false
}

var epoch = time.Unix(0, 0)

func getTimestamp(raw interface{}) time.Time {
	var ts string
	var ok bool
	if ts, ok = raw.(string); ok {
		if ts == "" {
			return epoch
		} else {
			t, err := time.Parse(time.RFC3339, ts)
			if err != nil {
				return epoch
			} else {
				return t
			}
		}
	} else {
		return epoch
	}
}

/**
 * Sorts a list of commits in reverse descending order, based on timestamp.
 */
func SortCommits(commits []Commit) {
	sort.Slice(commits, func(i, j int) bool {
		t1 := getTimestamp(commits[i].Properties["timestamp"])
		t2 := getTimestamp(commits[j].Properties["timestamp"])
		return t1.After(t2)
	})
}

/*
 * Validate a set of properties (as with remotes and parameters) for required and optional fields.
 */
func ValidateFields(properties map[string]interface{}, required []string, optional []string) error {
	for _, p := range required {
		if _, ok := properties[p]; !ok {
			return fmt.Errorf("missing required property '%s'", p)
		}
	}

	for p := range properties {
		if !contains(required, p) && !contains(optional, p) {
			return fmt.Errorf("invalid property '%s'", p)
		}
	}

	return nil
}

/*
 * Match a commit against a set of tags. Returns true if the commit matches the given tags, false otherwise.
 */
func MatchTags(commit map[string]interface{}, query []Tag) bool {
	// No tags always matches
	if len(query) == 0 {
		return true
	}

	var ok bool
	tags := map[string]string{}
	if raw, ok := commit["tags"].(map[string]interface{}); ok {
		for k, v := range raw {
			tags[k] = v.(string)
		}
	}  else if tags, ok = commit["tags"].(map[string]string); !ok {
		return false
	}

	for _, t := range query {
		var v interface{}
		if v, ok = tags[t.Key]; !ok {
			return false
		}

		if t.Value != nil && v != *t.Value {
			return false
		}
	}

	return true
}
