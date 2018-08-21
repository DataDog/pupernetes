package sets

import (
	"strings"
)

// String is a set of strings.
type String map[string]struct{}

// NewString returns a new set containing the items.
func NewString(items ...string) String {
	s := make(String)
	s.Insert(items...)
	return s
}

// NewStringFromString returns a new set containing the items from the split string.
func NewStringFromString(s, sep string) String {
	s = strings.TrimSpace(s)
	return NewString(strings.Split(s, sep)...)
}

// Insert adds items to the set.
func (s String) Insert(items ...string) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

// Delete removes items from the set.
func (s String) Delete(items ...string) {
	for _, item := range items {
		delete(s, item)
	}
}

// Difference returns the difference of the set with another set.
func (s String) Difference(s2 String) String {
	result := NewString()
	for key := range s {
		if !s2.Has(key) {
			result.Insert(key)
		}
	}
	return result
}

// Union returns the union of the set with another set.
func (s String) Union(s2 String) String {
	result := NewString()
	for key := range s {
		result.Insert(key)
	}
	for key := range s2 {
		result.Insert(key)
	}
	return result
}

func (s String) String() string {
	var list []string
	for item := range s {
		list = append(list, item)
	}
	return strings.Join(list, ",")
}

// Has returns true if an item is in the set.
func (s String) Has(item string) bool {
	_, ok := s[item]
	return ok
}

// Len returns the number of items in the set.
func (s String) Len() int {
	return len(s)
}
