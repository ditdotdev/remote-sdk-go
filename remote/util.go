/*
 * Copyright Datadatdat.
 */

// Package remote provides the core remote plugin infrastructure for Datadatdat.
package remote

import (
	"fmt"
	"sort"
	"time"
)

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
	var (
		ts string
		ok bool
	)

	if ts, ok = raw.(string); ok {
		if ts == "" {
			return epoch
		}

		t, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			return epoch
		}

		return t
	}

	return epoch
}

// SortCommits sorts a list of commits in reverse descending order, based on timestamp.
func SortCommits(commits []Commit) {
	sort.Slice(commits, func(i, j int) bool {
		t1 := getTimestamp(commits[i].Properties["timestamp"])
		t2 := getTimestamp(commits[j].Properties["timestamp"])

		return t1.After(t2)
	})
}

// ValidateFields validates a set of properties (as with remotes and parameters) for required and optional fields.
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

// MatchTags matches a commit against a set of tags. Returns true if the commit matches the given tags, false otherwise.
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
	} else if tags, ok = commit["tags"].(map[string]string); !ok {
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
