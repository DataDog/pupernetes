// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestNewOptions(t *testing.T) {
	testCases := []struct {
		optionsString    string
		availableOptions sets.String
		expected         sets.String
	}{
		{
			"kubectl,foo,iptables",
			sets.NewString("kubectl", "iptables"),
			sets.NewString("kubectl", "iptables"),
		},
		{
			"foo,bar,all",
			sets.NewString("kubectl", "iptables"),
			sets.NewString("kubectl", "iptables", "all"),
		},
		{
			"foo,bar",
			sets.NewString("kubectl", "iptables"),
			sets.NewString(),
		},
		{
			"all,none", // "all" has precedence over "none"
			sets.NewString("kubectl", "iptables"),
			sets.NewString("kubectl", "iptables", "all"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.optionsString, func(t *testing.T) {
			actual := newOptions(testCase.optionsString, testCase.availableOptions)
			assert.Equal(t, testCase.expected, actual)
		})
	}
}
