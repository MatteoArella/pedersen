// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package cmd

import (
	"strings"

	"github.com/matteoarella/pedersen/internal"
	"github.com/spf13/cobra"
)

type VersionCommand struct {
	cobra.Command
}

func NewVersionCommand() (*VersionCommand, error) {
	versionCmd := &VersionCommand{}

	versionCmd.Command = cobra.Command{
		Use:   "version",
		Short: "Show the Pedersen version information",
		RunE: func(*cobra.Command, []string) error {
			return versionCmd.execute()
		},
	}

	return versionCmd, nil
}

func (v *VersionCommand) execute() error {
	v.Println(strings.TrimPrefix(internal.Version, "v"))
	return nil
}
