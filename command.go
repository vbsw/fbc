/*
 *          Copyright 2020, Vitali Baumtrok.
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
	cmd_COUNT = 0
	cmd_CP    = 1
	cmd_MV    = 2
	cmd_PRINT = 3
	cmd_RM    = 4
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
	return commandFromArgs(os.Args[1:])
}

func commandFromArgs(args []string) (*command, error) {
	var cmd *command
	params, err := parametersFromArgs(args)

	if err == nil {
		if params == nil {
			cmd = new(command)
			cmd.info = true
			cmd.infoMessage = messageShortInfo()

		} else if params.incompatibleArguments() {
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

	if len(params.help) > 0 {
		cmd.info = true
		cmd.infoMessage = messageHelp()

	} else if len(params.version) > 0 {
		cmd.info = true
		cmd.infoMessage = messageVersion()

	} else if len(params.example) > 0 {
		cmd.info = true
		cmd.infoMessage = messageExample()

	} else if len(params.copyright) > 0 {
		cmd.info = true
		cmd.infoMessage = messageCopyright()

	} else {
		cmd.recursive = len(params.recursive) > 0
		cmd.or = len(params.or) > 0
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
		if len(params.commandId) > 0 {
			switch params.commandId[0].Key {
			case param_COUNT:
				cmd.commandId = cmd_COUNT
			case param_CP:
				cmd.commandId = cmd_CP
			case param_MV:
				cmd.commandId = cmd_MV
			case param_PRINT:
				cmd.commandId = cmd_PRINT
			case param_RM:
				cmd.commandId = cmd_RM
			default:
				err = errors.New("command " + params.commandId[0].Key + " is not implemented")
			}
		} else {
			err = errors.New("wrong command")
		}
	}
	return err
}

func (cmd *command) interpretFileNameFilter(params *parameters, err error) error {
	if err == nil && len(params.input) > 0 {
		input := ensurePathWithSlash(params.input[0].Value)
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
			params.input[0].Value = input

		} else {
			cmd.inputFilter = input[inputSlashEnd:]
			params.input[0].Value = input[:inputSlashEnd]

			if len(cmd.inputFilter) == 0 {
				cmd.inputFilter = "*"
			}
		}
	}
	return err
}

func (cmd *command) interpretInput(params *parameters, err error) error {
	if err == nil {
		if len(params.input) > 0 {
			input := params.input[0].Value
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
		if len(params.output) > 0 {
			output := params.output[0].Value
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
