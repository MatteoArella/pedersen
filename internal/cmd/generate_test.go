// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package cmd_test

import (
	"bytes"
	"fmt"
	iofs "io/fs"
	"testing"

	"github.com/matteoarella/pedersen/internal/cmd"
	"github.com/matteoarella/pedersen/internal/io/json"
	"github.com/matteoarella/pedersen/internal/io/xml"
	"github.com/matteoarella/pedersen/internal/io/yaml"
	"github.com/matteoarella/pedersen/internal/schema"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestGenerateCmd(t *testing.T) {
	testCases := []CliCmdTestCase{
		{
			scenario: "invalid command with no output arg",
			args:     []string{},
			err:      fmt.Errorf("required flag(s) \"out\" not set"),
		},
		{
			scenario: "valid group file with JSON encoding from file extension",
			args:     []string{"-o", "group.json"},
			validateFn: func(t *testing.T, fs afero.Fs) {
				bio := json.New(fs)
				group := schema.Group{}
				err := bio.ReadFile("group.json", &group)
				require.NoError(t, err)
			},
		},
		{
			scenario: "valid group file with YAML encoding from file extension",
			args:     []string{"-o", "group.yaml"},
			err:      nil,
			validateFn: func(t *testing.T, fs afero.Fs) {
				bio := yaml.New(fs)
				group := schema.Group{}
				err := bio.ReadFile("group.yaml", &group)
				require.NoError(t, err)
			},
		},
		{
			scenario: "valid group file with XML encoding from file extension",
			args:     []string{"-o", "group.xml"},
			err:      nil,
			validateFn: func(t *testing.T, fs afero.Fs) {
				bio := xml.New(fs)
				group := schema.Group{}
				err := bio.ReadFile("group.xml", &group)
				require.NoError(t, err)
			},
		},
		{
			scenario: "valid group file with JSON encoding from args format",
			args:     []string{"-o", "group", "--format", "json"},
			err:      nil,
			validateFn: func(t *testing.T, fs afero.Fs) {
				bio := json.New(fs)
				group := schema.Group{}
				err := bio.ReadFile("group.json", &group)
				require.NoError(t, err)
			},
		},
		{
			scenario: "valid group file with YAML encoding from args format",
			args:     []string{"-o", "group", "--format", "yaml"},
			err:      nil,
			validateFn: func(t *testing.T, fs afero.Fs) {
				bio := yaml.New(fs)
				group := schema.Group{}
				err := bio.ReadFile("group.yaml", &group)
				require.NoError(t, err)
			},
		},
		{
			scenario: "valid group file with XML encoding from args format",
			args:     []string{"-o", "group", "--format", "xml"},
			err:      nil,
			validateFn: func(t *testing.T, fs afero.Fs) {
				bio := xml.New(fs)
				group := schema.Group{}
				err := bio.ReadFile("group.xml", &group)
				require.NoError(t, err)
			},
		},
		{
			scenario: "valid group file with encoding override from args format",
			args:     []string{"-o", "group.xml", "--format", "json"},
			err:      nil,
			validateFn: func(t *testing.T, fs afero.Fs) {
				bio := json.New(fs)
				group := schema.Group{}
				err := bio.ReadFile("group.xml.json", &group)
				require.NoError(t, err)
			},
		},
		{
			scenario: "group file with specified file permission",
			args:     []string{"-o", "group.json", "--perm", "666"},
			err:      nil,
			validateFn: func(t *testing.T, fs afero.Fs) {
				info, err := fs.Stat("group.json")
				require.NoError(t, err)
				require.EqualValues(t, iofs.FileMode(0o666), info.Mode())
			},
		},
		{
			scenario: "group file with invalid file permission",
			args:     []string{"-o", "group.json", "--perm", "66s"},
			err:      fmt.Errorf("invalid argument \"66s\" for \"--perm\" flag: strconv.ParseUint: parsing \"66s\": invalid syntax"),
		},
		{
			scenario: "group with specified prime size (short option)",
			args:     []string{"-o", "group.json", "-b", "256"},
			err:      nil,
			validateFn: func(t *testing.T, fs afero.Fs) {
				bio := json.New(fs)
				group := schema.Group{}
				err := bio.ReadFile("group.json", &group)
				require.NoError(t, err)

				require.GreaterOrEqual(t, group.P.BitLen(), 256)
			},
		},
		{
			scenario: "group with specified prime size (long option)",
			args:     []string{"-o", "group.json", "--bits", "256"},
			err:      nil,
			validateFn: func(t *testing.T, fs afero.Fs) {
				bio := json.New(fs)
				group := schema.Group{}
				err := bio.ReadFile("group.json", &group)
				require.NoError(t, err)

				require.GreaterOrEqual(t, group.P.BitLen(), 256)
			},
		},
	}

	for _, scenario := range testCases {
		t.Run(scenario.scenario, func(t *testing.T) {
			fs := afero.NewMemMapFs()

			generateCmd, err := cmd.NewGenerateCommand(fs)
			require.NoError(t, err)

			buf := new(bytes.Buffer)
			generateCmd.SetOut(buf)
			generateCmd.SetErr(buf)

			generateCmd.SetArgs(scenario.args)

			err = generateCmd.Execute()
			require.EqualValues(t, scenario.err, err)

			if scenario.validateFn != nil {
				scenario.validateFn(t, fs)
			}
		})
	}
}
