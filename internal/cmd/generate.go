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
	defaultPrimeBits = 128
)

type GenerateCommand struct {
	cobra.Command

	fileFmtFlags
	primeBits int
	outFile   string
	fs        afero.Fs
}

func NewGenerateCommand(fs afero.Fs) (*GenerateCommand, error) {
	generateCmd := &GenerateCommand{fs: fs}

	generateCmd.Command = cobra.Command{
		Use:   "generate",
		Short: "Generate Pedersen parameters",
		RunE: func(*cobra.Command, []string) error {
			return generateCmd.execute()
		},
	}

	generateCmd.fileFmtFlags.register(&generateCmd.Command)

	generateCmd.PersistentFlags().IntVarP(&generateCmd.primeBits, "bits", "b", defaultPrimeBits, "prime bits size")
	generateCmd.PersistentFlags().StringVarP(&generateCmd.outFile, "out", "o", "", "output file")

	err := generateCmd.MarkPersistentFlagRequired("out")
	if err != nil {
		return nil, err
	}

	return generateCmd, nil
}

func (g *GenerateCommand) execute() error {
	group, err := pedersen.NewSchnorrGroup(g.primeBits)
	if err != nil {
		return err
	}

	return writeFileAutofmt(g.fs,
		g.fileFmt,
		g.outFile,
		&schema.Group{
			P: group.P,
			Q: group.Q,
			G: group.G,
			H: group.H,
		},
		iofs.FileMode(g.filePerm),
	)
}
