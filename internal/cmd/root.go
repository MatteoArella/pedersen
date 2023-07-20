// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package cmd

import (
	"os"

	perrors "github.com/matteoarella/pedersen/internal/errors"
	"github.com/matteoarella/pedersen/internal/logger"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type RootCommand struct {
	cobra.Command

	logLevel  string
	logOutput string
}

func NewRootCommand(fs afero.Fs) (*RootCommand, error) {
	rootCmd := &RootCommand{}

	rootCmd.Command = cobra.Command{
		PersistentPreRunE: func(c *cobra.Command, args []string) error {
			return logger.Init(rootCmd.logLevel, rootCmd.logOutput)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&rootCmd.logLevel, "loglevel", "", "INFO", "logging level")
	rootCmd.PersistentFlags().StringVarP(&rootCmd.logOutput, "logfile", "", "", "logging file")

	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		cmd.PrintErrln(err)
		cmd.PrintErrln()
		cmd.PrintErrln(cmd.UsageString())

		os.Exit(1)

		return nil
	})

	versionCmd, err := NewVersionCommand()
	if err != nil {
		return nil, err
	}

	generateCmd, err := NewGenerateCommand(fs)
	if err != nil {
		return nil, err
	}

	splitCmd, err := NewSplitCommand(fs)
	if err != nil {
		return nil, err
	}

	verifyCmd, err := NewVerifyCommand(fs)
	if err != nil {
		return nil, err
	}

	combineCmd, err := NewCombineCommand(fs)
	if err != nil {
		return nil, err
	}

	rootCmd.AddCommand(&versionCmd.Command,
		&generateCmd.Command,
		&splitCmd.Command,
		&verifyCmd.Command,
		&combineCmd.Command,
	)

	return rootCmd, nil
}

func Execute() error {
	fs := afero.NewOsFs()

	rootCmd, err := NewRootCommand(fs)
	if err != nil {
		return err
	}

	err = rootCmd.Execute()
	if err != nil {
		logrus.Error(perrors.UnwrapAll(err))
	}

	return err
}
