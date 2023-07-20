// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package cmd

import (
	"errors"
	"fmt"
	iofs "io/fs"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/matteoarella/pedersen/internal/io"
	"github.com/matteoarella/pedersen/internal/io/json"
	"github.com/matteoarella/pedersen/internal/io/xml"
	"github.com/matteoarella/pedersen/internal/io/yaml"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type FileFmt string
type FilePerm uint32

const (
	YAML FileFmt = "yaml"
	JSON FileFmt = "json"
	XML  FileFmt = "xml"
)

var (
	fmts = map[string]struct{}{
		string(YAML): {},
		string(JSON): {},
		string(XML):  {},
	}
)

func (f *FileFmt) String() string {
	return string(*f)
}

func keys(m map[string]struct{}) []string {
	r := make([]string, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	return r
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (f *FileFmt) Set(v string) error {
	v = strings.ToLower(v)
	if _, ok := fmts[v]; ok {
		*f = FileFmt(v)
		return nil
	}

	return fmt.Errorf("must be one of \"%s\"", strings.Join(keys(fmts), "\", \""))
}

// Type is only used in help text
func (f *FileFmt) Type() string {
	return "FileFmt"
}

func (f *FilePerm) String() string {
	return fmt.Sprintf("%#o", *f)
}

func (f *FilePerm) Set(v string) error {
	val, err := strconv.ParseUint(v, 8, 0)
	if err != nil {
		return err
	}

	*f = FilePerm(val)

	return nil
}

func (f *FilePerm) Type() string {
	return "FilePerm"
}

type fileFmtFlags struct {
	fileFmt  FileFmt
	filePerm FilePerm
}

func (f *fileFmtFlags) register(cmd *cobra.Command) {
	cmd.PersistentFlags().Var(&f.fileFmt, "format", fmt.Sprintf("file format. allowed: %s", strings.Join(keys(fmts), ", ")))
	cmd.PersistentFlags().AddFlag(&pflag.Flag{
		Name:      "perm",
		Value:     &f.filePerm,
		DefValue:  "400",
		Shorthand: "",
		Usage:     "output file permissions (default 400)",
	})
}

func readFileAutofmt(fs afero.Fs, name string, v interface{}) error {
	bios := []io.IO{yaml.New(fs), json.New(fs), xml.New(fs)}
	var err error

	for _, b := range bios {
		err = b.ReadFile(name, v)
		if err == nil {
			return nil
		}
	}

	return err
}

func writeFileAutofmt(fs afero.Fs, fileFmt FileFmt, name string, v interface{}, perm iofs.FileMode) error {
	bios := []io.IO{}

	switch fileFmt {
	case YAML:
		bios = append(bios, yaml.New(fs))
	case JSON:
		bios = append(bios, json.New(fs))
	case XML:
		bios = append(bios, xml.New(fs))
	default:
		ext := filepath.Ext(name)
		if ext == "" {
			bios = append(bios, yaml.New(fs), json.New(fs), xml.New(fs))
		} else {
			switch ext {
			case yaml.Ext():
				bios = append(bios, yaml.New(fs))
			case json.Ext():
				bios = append(bios, json.New(fs))
			case xml.Ext():
				bios = append(bios, xml.New(fs))
			}
		}
	}

	var err error

	for _, b := range bios {
		err = b.WriteFile(name, v, perm)
		if err == nil {
			return nil
		} else if errors.Is(err, iofs.ErrPermission) {
			return err
		}

	}

	return err
}
