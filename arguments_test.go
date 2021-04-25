/*
 *        Copyright 2021 Vitali Baumtrok.
 * Distributed under the Boost Software License, Version 1.0.
 *     (See accompanying file LICENSE or copy at
 *        http://www.boost.org/LICENSE_1_0.txt)
 */

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCommandLineA(t *testing.T) {
	args := new(arguments)

	clArgs := []string{}
	err := args.parseCommandLine(clArgs)
	if err != nil {
		t.Error(err.Error())
	}

	clArgs = []string{"--help", "--version"}
	err = args.parseCommandLine(clArgs)
	if err == nil {
		t.Error("incompatible parameters not recognized")
	}

	clArgs = []string{"--help", "cp"}
	err = args.parseCommandLine(clArgs)
	if err == nil {
		t.Error("incompatible parameters not recognized")
	}

	clArgs = []string{"asdf", "qwer"}
	err = args.parseCommandLine(clArgs)
	if err == nil {
		t.Error("missing command not recognized")
	}
}

func TestParseCommandLineB(t *testing.T) {
	args := new(arguments)

	clArgs := []string{"count"}
	err := args.parseCommandLine(clArgs)
	if err == nil {
		t.Error("unspecified input directory not recognized")
	}

	clArgs = []string{"cp", "./"}
	err = args.parseCommandLine(clArgs)
	if err == nil {
		t.Error("unspecified output directory not recognized")
	}

	clArgs = []string{"count", "a directory that hopefully does not exist"}
	err = args.parseCommandLine(clArgs)
	if err == nil {
		t.Error("unexistent input directory not recognized")
	}

	clArgs = []string{"cp", "./", "a directory that hopefully does not exist"}
	err = args.parseCommandLine(clArgs)
	if err == nil {
		t.Error("unexistent output directory not recognized")
	}

	clArgs = []string{"cp", "./", "./."}
	err = args.parseCommandLine(clArgs)
	if err == nil {
		t.Error("equal input and output directory not recognized")
	}
}

func TestParseCommandLineC(t *testing.T) {
	args := new(arguments)

	clArgs := []string{"count", "."}
	err := args.parseCommandLine(clArgs)
	if err != nil {
		t.Error(err.Error())
	} else if args.inputFilter != "*" {
		t.Error(args.inputFilter)
	} else {
		wd, err := os.Getwd()
		if err == nil {
			inputDir := filepath.Join(wd, ".")
			if args.input.Values()[0] != inputDir {
				t.Error(args.input.Values()[0])
			}
		} else {
			t.Error(err.Error())
		}
	}

	clArgs = []string{"count", "./*.txt"}
	err = args.parseCommandLine(clArgs)
	if err != nil {
		t.Error(err.Error())
	} else if args.inputFilter != "*.txt" {
		t.Error(args.inputFilter)
	}
}
