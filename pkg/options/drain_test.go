// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDrainOptions(t *testing.T) {
	testCases := []struct {
		input     string
		expected  *Drain
		cliString string
	}{
		{
			"all",
			&Drain{
				common{
					true,
					false,
				},
				true,
				true,
				true,
			},
			"all",
		},
		{
			"none",
			&Drain{
				common{
					false,
					true,
				},
				false,
				false,
				false,
			},
			"",
		},
		{
			"none,all",
			&Drain{
				common{
					false,
					true,
				},
				false,
				false,
				false,
			},
			"",
		},
		{
			"all,none",
			&Drain{
				common{
					true,
					false,
				},
				true,
				true,
				true,
			},
			"all",
		},
		{
			"pods",
			&Drain{
				common{
					false,
					false,
				},
				true,
				false,
				false,
			},
			"pods",
		},
		{
			"all,pods",
			&Drain{
				common{
					true,
					false,
				},
				true,
				true,
				true,
			},
			"all",
		},
		{
			"none,pods",
			&Drain{
				common{
					false,
					true,
				},
				false,
				false,
				false,
			},
			"",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			actual := NewDrainOptions(testCase.input)
			assert.Equal(t, testCase.expected, actual)
			assert.Equal(t, testCase.cliString, actual.StringCLI())
		})
	}
}
