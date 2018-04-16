// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package setup

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckHostname(t *testing.T) {
	cases := []struct {
		hostname string
		valid    bool
	}{
		{
			"ok",
			true,
		},
		{
			"KO",
			false,
		},
		{
			"ok-host.com",
			true,
		},
		{
			"ko-HOST.com",
			false,
		},
		{
			"plop-VirtualBox",
			false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.hostname, func(t *testing.T) {
			assert.Equal(t, tc.valid, isValidHostname(tc.hostname))
		})
	}
}
