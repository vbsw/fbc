/*
 *      Copyright 2021, 2022 Vitali Baumtrok.
 * Distributed under the Boost Software License, Version 1.0.
 *     (See accompanying file LICENSE or copy at
 *        http://www.boost.org/LICENSE_1_0.txt)
 */

package main

import (
	"github.com/vbsw/golib/osargs"
	"os"
	"path/filepath"
	"testing"
)

func TestParseOSArgsA(t *testing.T) {
	args := new(osargs.Arguments)
	args.Values = []string{}
	args.Parsed = make([]bool, len(args.Values))
	params := new(tParameters)
	err := params.initFromArgs(args)
	if err != nil {
		t.Error(err.Error())
	}

	args.Values = []string{"--help", "--version"}
	args.Parsed = make([]bool, len(args.Values))
	params = new(tParameters)
	err = params.initFromArgs(args)
	if err == nil {
		t.Error("incompatible parameters not recognized")
	}

	args.Values = []string{"--help", "cp"}
	args.Parsed = make([]bool, len(args.Values))
	params = new(tParameters)
	err = params.initFromArgs(args)
	if err == nil {
		t.Error("incompatible parameters not recognized")
	}

	args.Values = []string{"asdf", "qwer"}
	args.Parsed = make([]bool, len(args.Values))
	params = new(tParameters)
	err = params.initFromArgs(args)
	if err == nil {
		t.Error("missing command not recognized")
	}
}

func TestParseOSArgsB(t *testing.T) {
	args := new(osargs.Arguments)
	args.Values = []string{"count"}
	args.Parsed = make([]bool, len(args.Values))
	params := new(tParameters)
	err := params.initFromArgs(args)
	if err == nil {
		t.Error("unspecified input directory not recognized")
	}

	args.Values = []string{"cp", "./"}
	args.Parsed = make([]bool, len(args.Values))
	params = new(tParameters)
	err = params.initFromArgs(args)
	if err == nil {
		t.Error("unspecified output directory not recognized")
	}

	args.Values = []string{"count", "a directory that hopefully does not exist"}
	args.Parsed = make([]bool, len(args.Values))
	params = new(tParameters)
	err = params.initFromArgs(args)
	if err == nil {
		t.Error("unexistent input directory not recognized")
	}

	args.Values = []string{"cp", "./", "a directory that hopefully does not exist"}
	args.Parsed = make([]bool, len(args.Values))
	params = new(tParameters)
	err = params.initFromArgs(args)
	if err == nil {
		t.Error("unexistent output directory not recognized")
	}

	args.Values = []string{"cp", "./", "./."}
	args.Parsed = make([]bool, len(args.Values))
	params = new(tParameters)
	err = params.initFromArgs(args)
	if err == nil {
		t.Error("equal input and output directory not recognized")
	}
}

func TestParseOSArgsC(t *testing.T) {
	args := new(osargs.Arguments)
	args.Values = []string{"count", "."}
	args.Parsed = make([]bool, len(args.Values))
	params := new(tParameters)
	err := params.initFromArgs(args)
	if err != nil {
		t.Error(err.Error())
	} else if params.fileNameFilter != "*" {
		t.Error(params.fileNameFilter)
	} else {
		wd, err := os.Getwd()
		if err == nil {
			inputDir := filepath.Join(wd, ".")
			if params.input.Values[0] != inputDir {
				t.Error(params.input.Values[0])
			}
		} else {
			t.Error(err.Error())
		}
	}

	args.Values = []string{"count", "./*.txt"}
	args.Parsed = make([]bool, len(args.Values))
	params = new(tParameters)
	err = params.initFromArgs(args)
	if err != nil {
		t.Error(err.Error())
	} else if params.fileNameFilter != "*.txt" {
		t.Error(params.fileNameFilter)
	}
}
