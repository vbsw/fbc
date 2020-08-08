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
	return commandFromArgs(os.Args[1:])
}

func commandFromArgs(osArgs []string) (*command, error) {
	var cmd *command
	args, err := argumentsFromCL(osArgs)

	if err == nil {
		if args == nil {
			cmd = new(command)
			cmd.info = true
			cmd.infoMessage = messageShortInfo()

		} else if args.incompatibleArguments() {
			err = errors.New("wrong argument usage")

		} else if args.oneParamHasMultipleResults() {
			err = errors.New("wrong argument usage")

		} else {
			cmd = new(command)
			err = cmd.initFromParams(args)
		}
	}
	return cmd, err
}

func (cmd *command) initFromParams(args *arguments) error {
	var err error

	if len(args.help) > 0 {
		cmd.info = true
		cmd.infoMessage = messageHelp()

	} else if len(args.version) > 0 {
		cmd.info = true
		cmd.infoMessage = messageVersion()

	} else if len(args.example) > 0 {
		cmd.info = true
		cmd.infoMessage = messageExample()

	} else if len(args.copyright) > 0 {
		cmd.info = true
		cmd.infoMessage = messageCopyright()

	} else {
		cmd.recursive = len(args.recursive) > 0
		cmd.or = len(args.or) > 0
		cmd.contentFilter = argsToStringArray(args.filter)
		err = cmd.interpretCommand(args, err)
		err = cmd.interpretFileNameFilter(args, err)
		err = cmd.interpretInput(args, err)
		err = cmd.interpretOutput(args, err)
		err = cmd.checkIODirectories(err)
	}
	return err
}

func (cmd *command) interpretCommand(args *arguments, err error) error {
	if err == nil {
		if len(args.commandId) > 0 {
			switch args.commandId[0].Key {
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
				err = errors.New("command " + args.commandId[0].Key + " is not implemented")
			}
		} else {
			err = errors.New("wrong command")
		}
	}
	return err
}

func (cmd *command) interpretFileNameFilter(args *arguments, err error) error {
	if err == nil && len(args.input) > 0 {
		input := ensurePathWithSlash(args.input[0].Value)
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
			args.input[0].Value = input

		} else {
			cmd.inputFilter = input[inputSlashEnd:]
			args.input[0].Value = input[:inputSlashEnd]

			if len(cmd.inputFilter) == 0 {
				cmd.inputFilter = "*"
			}
		}
	}
	return err
}

func (cmd *command) interpretInput(args *arguments, err error) error {
	if err == nil {
		if len(args.input) > 0 {
			input := args.input[0].Value
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

func (cmd *command) interpretOutput(args *arguments, err error) error {
	if err == nil {
		if len(args.output) > 0 {
			output := args.output[0].Value
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
		} else if args.isOutputDirNeeded() {
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
