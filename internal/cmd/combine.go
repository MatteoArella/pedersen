// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package cmd

import (
	"errors"
	iofs "io/fs"

	"github.com/matteoarella/pedersen"
	"github.com/matteoarella/pedersen/big"
	"github.com/matteoarella/pedersen/internal/schema"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type CombineCommand struct {
	cobra.Command

	pedersenFlags
	fileFmtFlags
	secretSharesFlags
	outFile string
	verify  bool
	fs      afero.Fs
}

func NewCombineCommand(fs afero.Fs) (*CombineCommand, error) {
	combineCmd := &CombineCommand{
		fs: fs,
		secretSharesFlags: secretSharesFlags{
			fs: fs,
		},
	}

	combineCmd.Command = cobra.Command{
		Use:   "combine",
		Short: "Combine Pedersen shares",
		RunE: func(*cobra.Command, []string) error {
			return combineCmd.execute()
		},
	}

	err := combineCmd.pedersenFlags.register(&combineCmd.Command)
	if err != nil {
		return nil, err
	}

	err = combineCmd.secretSharesFlags.register(&combineCmd.Command)
	if err != nil {
		return nil, err
	}

	combineCmd.fileFmtFlags.register(&combineCmd.Command)

	combineCmd.PersistentFlags().StringVarP(&combineCmd.outFile, "out", "o", "", "output file")
	combineCmd.PersistentFlags().BoolVarP(&combineCmd.verify, "verify", "v", true, "verify shares before combine")

	err = combineCmd.MarkPersistentFlagRequired("out")
	if err != nil {
		return nil, err
	}

	return combineCmd, nil
}

func (c *CombineCommand) execute() error {
	group := pedersen.Group{}

	if err := readFileAutofmt(c.fs, c.groupFile, &group); err != nil {
		return err
	}

	commitments := schema.Commitments{}
	if err := readFileAutofmt(c.fs, c.commitmentsFile, &commitments); err != nil {
		return err
	}

	shares := &pedersen.Shares{
		Abscissae:   make([]*big.Int, c.parts),
		Commitments: commitments.Commitments,
		Parts:       make([][]pedersen.SecretPart, c.parts),
	}

	for i := 0; i < c.parts; i++ {
		parts := schema.Shares{}

		shareFile := c.share(i)

		if err := readFileAutofmt(c.fs, shareFile, &parts); err != nil {
			if errors.Is(err, iofs.ErrNotExist) {
				continue
			}

			return err
		}

		shares.Abscissae[i] = parts.Abscissa
		shares.Parts[i] = parts.Parts
	}

	p, err := pedersen.NewPedersen(c.parts,
		c.threshold,
		pedersen.CyclicGroup(&group),
	)
	if err != nil {
		return err
	}

	if c.verify {
		if err := p.VerifyShares(shares); err != nil {
			return err
		}
	}

	reconstructed, err := p.Combine(shares)
	if err != nil {
		return err
	}

	if err := afero.WriteFile(c.fs, c.outFile, reconstructed, iofs.FileMode(c.filePerm)); err != nil {
		return err
	}

	return nil
}
