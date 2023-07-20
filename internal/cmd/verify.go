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
	"golang.org/x/sync/errgroup"
)

type VerifySharesCommand struct {
	cobra.Command

	pedersenFlags
	secretSharesFlags
	fs afero.Fs
}

func (v *VerifySharesCommand) execute() error {
	group := pedersen.Group{}

	if err := readFileAutofmt(v.fs, v.groupFile, &group); err != nil {
		return err
	}

	commitments := schema.Commitments{}
	if err := readFileAutofmt(v.fs, v.commitmentsFile, &commitments); err != nil {
		return err
	}

	shares := &pedersen.Shares{
		Abscissae:   make([]*big.Int, v.parts),
		Commitments: commitments.Commitments,
		Parts:       make([][]pedersen.SecretPart, v.parts),
	}

	for i := 0; i < v.parts; i++ {
		parts := schema.Shares{}

		shareFile := v.share(i)

		if err := readFileAutofmt(v.fs, shareFile, &parts); err != nil {
			if errors.Is(err, iofs.ErrNotExist) {
				continue
			}

			return err
		}

		shares.Abscissae[i] = parts.Abscissa
		shares.Parts[i] = parts.Parts
	}

	p, err := pedersen.NewPedersen(v.parts,
		v.threshold,
		pedersen.CyclicGroup(&group),
	)
	if err != nil {
		return err
	}

	return p.VerifyShares(shares)
}

func NewVerifySharesCommand(fs afero.Fs) (*VerifySharesCommand, error) {
	verifySharesCmd := &VerifySharesCommand{
		fs: fs,
		secretSharesFlags: secretSharesFlags{
			fs: fs,
		},
	}

	verifySharesCmd.Command = cobra.Command{
		Use:   "shares",
		Short: "Verify Pedersen shares",
		RunE: func(*cobra.Command, []string) error {
			return verifySharesCmd.execute()
		},
	}

	err := verifySharesCmd.pedersenFlags.register(&verifySharesCmd.Command)
	if err != nil {
		return nil, err
	}

	err = verifySharesCmd.secretSharesFlags.register(&verifySharesCmd.Command)
	if err != nil {
		return nil, err
	}

	return verifySharesCmd, nil
}

type VerifyPartCommand struct {
	cobra.Command

	pedersenFlags
	secretShareFlags
	Fs afero.Fs
}

func (v *VerifyPartCommand) execute() error {
	group := pedersen.Group{}

	if err := readFileAutofmt(v.Fs, v.groupFile, &group); err != nil {
		return err
	}

	commitments := schema.Commitments{}
	if err := readFileAutofmt(v.Fs, v.commitmentsFile, &commitments); err != nil {
		return err
	}

	parts := schema.Shares{}
	if err := readFileAutofmt(v.Fs, v.shareFile, &parts); err != nil {
		return err
	}

	p, err := pedersen.NewPedersen(v.parts,
		v.threshold,
		pedersen.CyclicGroup(&group),
	)
	if err != nil {
		return err
	}

	g := errgroup.Group{}

	for i := 0; i < len(parts.Parts); i++ {
		i := i
		g.Go(func() error {
			return p.Verify(parts.Abscissa, parts.Parts[i], commitments.Commitments[i])
		})
	}

	return g.Wait()
}

func NewVerifyPartCommand(fs afero.Fs) (*VerifyPartCommand, error) {
	verifyPartCmd := &VerifyPartCommand{
		Fs: fs,
	}

	verifyPartCmd.Command = cobra.Command{
		Use:   "part",
		Short: "Verify Pedersen part",
		RunE: func(*cobra.Command, []string) error {
			return verifyPartCmd.execute()
		},
	}

	if err := verifyPartCmd.pedersenFlags.register(&verifyPartCmd.Command); err != nil {
		return nil, err
	}

	if err := verifyPartCmd.secretShareFlags.register(&verifyPartCmd.Command); err != nil {
		return nil, err
	}

	return verifyPartCmd, nil
}

type VerifyCommand struct {
	cobra.Command

	fs afero.Fs
}

func NewVerifyCommand(fs afero.Fs) (*VerifyCommand, error) {
	verifyCmd := &VerifyCommand{fs: fs}

	verifyCmd.Command = cobra.Command{
		Use:   "verify",
		Short: "Verify Pedersen shares or parts",
	}

	verifySharesCommand, err := NewVerifySharesCommand(verifyCmd.fs)
	if err != nil {
		return nil, err
	}

	verifyPartCommand, err := NewVerifyPartCommand(verifyCmd.fs)
	if err != nil {
		return nil, err
	}

	verifyCmd.AddCommand(&verifySharesCommand.Command, &verifyPartCommand.Command)

	return verifyCmd, nil
}
