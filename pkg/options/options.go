// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package options

import (
	"sort"
	"strings"

	"github.com/DataDog/pupernetes/pkg/util/sets"
	"github.com/fatih/structs"
	"github.com/golang/glog"
)

var commonOptions = sets.NewString("all", "none")

type common struct {
	All  bool `json:"all,omitempty"`
	None bool `json:"none,omitempty"`
}

// GetCleanOptionsString returns a string representation of available "--clean" options.
func GetCleanOptionsString() string {
	return getOptionsString(Clean{})
}

// GetDrainOptionsString returns a string representation of available "--drain" options.
func GetDrainOptionsString() string {
	return getOptionsString(Drain{})
}

func getOptionsString(opt interface{}) string {
	names := getOptionNames(opt)
	sort.Strings(names)
	names = append(names, "all", "none")
	return strings.Join(names, ",")
}

func getOptionNames(v interface{}) []string {
	var names []string
	for _, name := range structs.Names(v) {
		name = strings.ToLower(name)
		if name == "common" {
			continue
		}
		names = append(names, name)
	}
	return names
}

func newOptions(optionsString string, availableOptions sets.String) sets.String {
	opts := sets.NewStringFromString(optionsString, ",")

	diff := opts.Difference(availableOptions.Union(commonOptions))
	if diff.Len() > 0 {
		glog.Warningf("%q are not in available options %q", diff, availableOptions)
	}

	if opts.Has("all") && opts.Has("none") {
		glog.Warningf("\"all\" and \"none\" are mutually exclusive options. Using \"all\"")
		opts.Delete("none") // "all" has precedence over "none"
	}

	if opts.Has("all") {
		opts = opts.Union(availableOptions)
	}
	if opts.Has("none") {
		opts = sets.NewString("none")
	}

	return opts
}
