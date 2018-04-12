// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCleanOptions(t *testing.T) {
	testCases := []struct {
		input    string
		expected *Clean
	}{
		{
			"all",
			&Clean{
				common{
					true,
					false,
				},
				true,
				true,
				true,
				true,
				true,
				true,
				true,
				true,
				true,
			},
		},
		{
			"none",
			&Clean{
				common{
					false,
					true,
				},
				false,
				false,
				false,
				false,
				false,
				false,
				false,
				false,
				false,
			},
		},
		{
			"none,all",
			&Clean{
				common{
					false,
					true,
				},
				false,
				false,
				false,
				false,
				false,
				false,
				false,
				false,
				false,
			},
		},
		{
			"all,none",
			&Clean{
				common{
					true,
					false,
				},
				true,
				true,
				true,
				true,
				true,
				true,
				true,
				true,
				true,
			},
		},
		{
			"etcd",
			&Clean{
				common{
					false,
					false,
				},
				true,
				false,
				false,
				false,
				false,
				false,
				false,
				false,
				false,
			},
		},
		{
			"all,etcd",
			&Clean{
				common{
					true,
					false,
				},
				true,
				true,
				true,
				true,
				true,
				true,
				true,
				true,
				true,
			},
		},
		{
			"etcd,binaries",
			&Clean{
				common{
					false,
					false,
				},
				true,
				true,
				false,
				false,
				false,
				false,
				false,
				false,
				false,
			},
		},
		{
			"etcd,binaries,secrets",
			&Clean{
				common{
					false,
					false,
				},
				true,
				true,
				false,
				false,
				true,
				false,
				false,
				false,
				false,
			},
		},
		{
			"none,etcd",
			&Clean{
				common{
					false,
					true,
				},
				false,
				false,
				false,
				false,
				false,
				false,
				false,
				false,
				false,
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			actual := NewCleanOptions(testCase.input)
			assert.Equal(t, testCase.expected, actual)
		})
	}
}
