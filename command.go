/*
 *        Copyright 2020, 2021 Vitali Baumtrok.
 * Distributed under the Boost Software License, Version 1.0.
 *     (See accompanying file LICENSE or copy at
 *        http://www.boost.org/LICENSE_1_0.txt)
 */

package main

import (
	"errors"
	"github.com/vbsw/checkfile"
	"os"
	"path/filepath"
)

const (
	cmdCOUNT = 0
	cmdCP    = 1
	cmdMV    = 2
	cmdPRINT = 3
	cmdRM    = 4
)

type command struct {
	commandId     int
	contentFilter []string
	info          bool
	infoMessage   string
	inputDir      string
	inputFilter   string
	or            bool
	outputDir     string
	recursive     bool
}

func commandFromOSArgs() (*command, error) {
	var cmd *command
	params, err := parametersFromCL()

	if err == nil {
		if params == nil {
			cmd = new(command)
			cmd.info = true
			cmd.infoMessage = messageShortInfo()

		} else if params.incompatibleParameters() {
			err = errors.New("wrong argument usage")

		} else if params.oneParamHasMultipleResults() {
			err = errors.New("wrong argument usage")

		} else {
			cmd = new(command)
			err = cmd.initFromParams(params)
		}
	}
	return cmd, err
}

func (cmd *command) initFromParams(params *parameters) error {
	var err error

	if params.help.Available() {
		cmd.info = true
		cmd.infoMessage = messageHelp()

	} else if params.version.Available() {
		cmd.info = true
		cmd.infoMessage = messageVersion()

	} else if params.example.Available() {
		cmd.info = true
		cmd.infoMessage = messageExample()

	} else if params.copyright.Available() {
		cmd.info = true
		cmd.infoMessage = messageCopyright()

	} else {
		cmd.recursive = params.recursive.Available()
		cmd.or = params.or.Available()
		cmd.contentFilter = paramsToStringArray(params.filter)
		err = cmd.interpretCommand(params, err)
		err = cmd.interpretFileNameFilter(params, err)
		err = cmd.interpretInput(params, err)
		err = cmd.interpretOutput(params, err)
		err = cmd.checkIODirectories(err)
	}
	return err
}

func (cmd *command) interpretCommand(params *parameters, err error) error {
	if err == nil {
		if params.commandId.Available() {
			switch params.commandId.Keys()[0] {
			case argCOUNT:
				cmd.commandId = cmdCOUNT
			case argCP:
				cmd.commandId = cmdCP
			case argMV:
				cmd.commandId = cmdMV
			case argPRINT:
				cmd.commandId = cmdPRINT
			case argRM:
				cmd.commandId = cmdRM
			default:
				err = errors.New("command " + params.commandId.Keys()[0] + " is not implemented")
			}
		} else {
			err = errors.New("wrong command")
		}
	}
	return err
}

func (cmd *command) interpretFileNameFilter(params *parameters, err error) error {
	if err == nil && params.input.Available() {
		input := ensurePathWithSlash(params.input.Values()[0])
		inputSlash := firstSlash(input)

		// assume it's current directory
		if inputSlash == 0 {
			inputSlash = os.PathSeparator
			input = "." + string(inputSlash) + input
		}
		inputSlashEnd := rindex(input, inputSlash) + 1

		// input file is a directory
		if len(input) > inputSlashEnd && checkfile.IsDirectory(input) {
			input += string(inputSlash)
			cmd.inputFilter = "*"
			params.input.Values()[0] = input

		} else {
			cmd.inputFilter = input[inputSlashEnd:]
			params.input.Values()[0] = input[:inputSlashEnd]

			if len(cmd.inputFilter) == 0 {
				cmd.inputFilter = "*"
			}
		}
	}
	return err
}

func (cmd *command) interpretInput(params *parameters, err error) error {
	if err == nil {
		if params.input.Available() {
			input := params.input.Values()[0]
			fileInfo, statErr := os.Stat(input)

			if statErr == nil || !os.IsNotExist(statErr) {

				if fileInfo != nil {
					if fileInfo.IsDir() {
						cmd.inputDir, err = filepath.Abs(input)

					} else {
						err = errors.New("input path is a file, but must be a directory")
					}
				} else {
					err = errors.New("wrong input path syntax")
				}
			} else {
				err = errors.New("input directory does not exist")
			}
		} else {
			err = errors.New("input directory is not specified")
		}
	}
	return err
}

func (cmd *command) interpretOutput(params *parameters, err error) error {
	if err == nil {
		if params.output.Available() {
			output := params.output.Values()[0]
			fileInfo, statErr := os.Stat(output)

			if statErr == nil || !os.IsNotExist(statErr) {
				if fileInfo != nil {
					if fileInfo.IsDir() {
						cmd.outputDir, err = filepath.Abs(output)

					} else {
						err = errors.New("output path is a file, but must be a directory")
					}
				} else {
					err = errors.New("wrong output path syntax")
				}
			} else {
				err = errors.New("output directory does not exist")
			}
		} else if params.isOutputDirNeeded() {
			err = errors.New("output directory is not specified")
		}
	}
	return err
}

func (cmd *command) checkIODirectories(err error) error {
	if err == nil {
		// TODO: better comparison
		if cmd.inputDir == cmd.outputDir {
			err = errors.New("input and output directories are the same")
		}
	}
	return err
}

func firstSlash(str string) byte {
	for _, b := range []byte(str) {
		if b == '/' || b == '\\' {
			return b
		}
	}
	return 0
}

func ensurePathWithSlash(path string) string {
	if path == "." || path == ".." {
		path += string(os.PathSeparator)
	}
	return path
}

func rindex(str string, b byte) int {
	bytes := []byte(str)
	for i := len(bytes) - 1; i >= 0; i-- {
		if bytes[i] == b {
			return i
		}
	}
	return -1
}
