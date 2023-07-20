// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package cmd

import (
	iofs "io/fs"

	"github.com/matteoarella/pedersen"
	"github.com/matteoarella/pedersen/internal/schema"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const (
	defaultPedersenParts     = 5
	defaultPedersenThreshold = 3
)

type SplitCommand struct {
	cobra.Command

	fileFmtFlags
	pedersenFlags
	secretSharesFlags
	inFile string
	fs     afero.Fs
}

func NewSplitCommand(fs afero.Fs) (*SplitCommand, error) {
	splitCmd := &SplitCommand{
		fs: fs,
		secretSharesFlags: secretSharesFlags{
			fs: fs,
		},
	}

	splitCmd.Command = cobra.Command{
		Use:   "split",
		Short: "Split secret into Pedersen shares",
		RunE: func(*cobra.Command, []string) error {
			return splitCmd.execute()
		},
	}

	err := splitCmd.pedersenFlags.register(&splitCmd.Command)
	if err != nil {
		return nil, err
	}

	err = splitCmd.secretSharesFlags.register(&splitCmd.Command)
	if err != nil {
		return nil, err
	}

	splitCmd.fileFmtFlags.register(&splitCmd.Command)

	splitCmd.PersistentFlags().StringVarP(&splitCmd.inFile, "in", "i", "", "input file")

	err = splitCmd.MarkPersistentFlagRequired("in")
	if err != nil {
		return nil, err
	}

	return splitCmd, nil
}

func (s *SplitCommand) execute() error {
	group := schema.Group{}
	if err := readFileAutofmt(s.fs, s.groupFile, &group); err != nil {
		return err
	}

	// read secret file
	inFile, err := afero.ReadFile(s.fs, s.inFile)
	if err != nil {
		return err
	}

	p, err := pedersen.NewPedersen(s.parts,
		s.threshold,
		pedersen.CyclicGroup(&pedersen.Group{
			P: group.P,
			Q: group.Q,
			G: group.G,
			H: group.H,
		}),
	)
	if err != nil {
		return err
	}

	shares, err := p.Split(inFile, nil)
	if err != nil {
		return err
	}

	for i := 0; i < s.parts; i++ {
		parts := schema.Shares{
			Abscissa: shares.Abscissae[i],
			Parts:    shares.Parts[i],
		}

		if err := writeFileAutofmt(s.fs, s.fileFmt, s.share(i), parts, iofs.FileMode(s.filePerm)); err != nil {
			return err
		}
	}

	commitments := schema.Commitments{
		Commitments: shares.Commitments,
	}

	return writeFileAutofmt(s.fs,
		s.fileFmt,
		s.commitmentsFile,
		commitments,
		iofs.FileMode(s.filePerm),
	)
}
