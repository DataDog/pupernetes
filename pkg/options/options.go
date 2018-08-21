// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package options

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"

	"github.com/fatih/structs"
	"github.com/golang/glog"
)

type common struct {
	All  bool `json:"all,omitempty"`
	None bool `json:"none,omitempty"`
}

func containsString(slice []string, elt string) bool {
	for _, item := range slice {
		if elt == item {
			return true
		}
	}
	return false
}

func setAllOptionsTo(d interface{}, set bool) {
	for _, name := range structs.Names(d) {
		if name == "common" {
			continue
		}
		reflect.ValueOf(d).Elem().FieldByName(name).SetBool(set)
	}
}

// GetOptionNames returns the options from the given interface
// The interface must be a Clean or Drain one
func GetOptionNames(opt interface{}) string {
	var names []string
	for _, elt := range structs.Names(opt) {
		elt = strings.ToLower(elt)
		if elt == "common" {
			continue
		}
		names = append(names, elt)
	}
	sort.Strings(names)
	names = append(names, "all", "none")
	return strings.Join(names, ",")
}

func newOptions(stringOptions string, enabled bool, opt interface{}) interface{} {
	stringOptions = strings.TrimSpace(stringOptions)
	defer func() {
		b, err := json.Marshal(opt)
		if err != nil {
			glog.Errorf("Cannot display options: %v", err)
			return
		}
		t := reflect.TypeOf(opt).String()
		t = strings.TrimPrefix(t, "*options.")
		glog.V(3).Infof("%s options are: %s", t, string(b))
	}()
	setAllOptionsTo(opt, !enabled)
	availableOptions := structs.Names(opt)
	for i := range availableOptions {
		availableOptions[i] = strings.ToLower(availableOptions[i])
	}
	for _, elt := range strings.Split(stringOptions, ",") {
		switch elt {
		case "all":
			setAllOptionsTo(opt, enabled)
			reflect.ValueOf(opt).Elem().FieldByName("All").SetBool(enabled)
			reflect.ValueOf(opt).Elem().FieldByName("None").SetBool(!enabled)
			return opt

		case "none":
			setAllOptionsTo(opt, !enabled)
			reflect.ValueOf(opt).Elem().FieldByName("All").SetBool(!enabled)
			reflect.ValueOf(opt).Elem().FieldByName("None").SetBool(enabled)
			return opt

		case "common":
			glog.Warningf("Cannot use %q as option", elt)
			continue

		default:
			if !containsString(availableOptions, elt) {
				glog.Warningf("Cannot use %q as option it's not in %s", elt, availableOptions)
				continue
			}
			elt = strings.Title(elt)
			reflect.ValueOf(opt).Elem().FieldByName(elt).SetBool(enabled)
		}
	}
	return opt
}
