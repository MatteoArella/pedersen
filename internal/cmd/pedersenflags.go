// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type pedersenFlags struct {
	parts     int
	threshold int
	groupFile string
}

func (p *pedersenFlags) register(cmd *cobra.Command) error {
	cmd.PersistentFlags().IntVarP(&p.parts, "parts", "p", defaultPedersenParts, "shares parts")
	cmd.PersistentFlags().IntVarP(&p.threshold, "threshold", "t", defaultPedersenThreshold, "shares threshold")
	cmd.PersistentFlags().StringVarP(&p.groupFile, "group", "g", "", "group file")

	return cmd.MarkPersistentFlagRequired("group")
}

type secretSharesFlags struct {
	sharesFilePattern string
	commitmentsFile   string

	fs afero.Fs
}

func (s *secretSharesFlags) register(cmd *cobra.Command) error {
	cmd.PersistentFlags().StringVarP(&s.sharesFilePattern, "shares", "", "", `secret shares files pattern expression.
Use '*' as placeholder for the index of the share
(e.g. shares/shareholder-*)`)
	cmd.PersistentFlags().StringVarP(&s.commitmentsFile, "commitments", "", "", "commitments file")

	err := cmd.MarkPersistentFlagRequired("shares")
	if err != nil {
		return err
	}

	return cmd.MarkPersistentFlagRequired("commitments")
}

func (s *secretSharesFlags) share(index int) string {
	return strings.ReplaceAll(s.sharesFilePattern, "*", fmt.Sprintf("%d", index))
}

type secretShareFlags struct {
	shareFile       string
	commitmentsFile string
}

func (s *secretShareFlags) register(cmd *cobra.Command) error {
	cmd.PersistentFlags().StringVarP(&s.shareFile, "share", "", "", "secret shares file")
	cmd.PersistentFlags().StringVarP(&s.commitmentsFile, "commitments", "", "", "commitments file")

	err := cmd.MarkPersistentFlagRequired("share")
	if err != nil {
		return err
	}

	return cmd.MarkPersistentFlagRequired("commitments")
}
