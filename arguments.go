/*
 *        Copyright 2021 Vitali Baumtrok.
 * Distributed under the Boost Software License, Version 1.0.
 *     (See accompanying file LICENSE or copy at
 *        http://www.boost.org/LICENSE_1_0.txt)
 */

package main

import (
	"errors"
	"fmt"
	"github.com/vbsw/cmdl"
	"os"
	"path/filepath"
	"unsafe"
)

const (
	argCOUNT = "count"
	argCP    = "cp"
	argMV    = "mv"
	argPRINT = "print"
	argRM    = "rm"
	keyNONE  = "<none>"
)

type arguments struct {
	help          *cmdl.Parameter
	version       *cmdl.Parameter
	example       *cmdl.Parameter
	copyright     *cmdl.Parameter
	or            *cmdl.Parameter
	silent        *cmdl.Parameter
	command       *cmdl.Parameter
	recursive     *cmdl.Parameter
	filter        *cmdl.Parameter
	input         *cmdl.Parameter
	output        *cmdl.Parameter
	contentFilter []string
	inputFilter   string
}

func (args *arguments) parseCommandLine(clArgs []string) error {
	var err error
	if len(clArgs) > 0 {
		cl := cmdl.NewFrom(clArgs)
		args.help = cl.NewParam().Parse("-h", "--help", "-help", "help")
		args.version = cl.NewParam().Parse("-v", "--version", "-version", "version")
		args.example = cl.NewParam().Parse("-e", "--example", "-example", "example")
		args.copyright = cl.NewParam().Parse("-c", "--copyright", "-copyright", "copyright")
		args.or = cl.NewParam().Parse("-o", "--or", "-or", "or")
		args.silent = cl.NewParam().Parse("-s", "--silent", "-silent", "silent")
		args.command = cl.NewParam().Parse(argCOUNT, argCP, argMV, argPRINT, argRM)
		args.recursive = cl.NewParam().Parse("-r", "--recursive", "-recursive", "recursive")
		args.filter = cl.NewParam()
		args.input = cl.NewParam()
		args.output = cl.NewParam()

		unparsedCLArgs := cl.UnparsedArgs()
		unparsedCLArgs = parseInput(args, unparsedCLArgs)
		unparsedCLArgs = parseOutput(args, unparsedCLArgs)
		parseFilter(args, unparsedCLArgs)
		parseInputFilter(args)

		err = validateParameters(args)
	}
	return err
}

func (args *arguments) infoParameters() []*cmdl.Parameter {
	infoParams := make([]*cmdl.Parameter, 4)
	infoParams[0] = args.help
	infoParams[1] = args.version
	infoParams[2] = args.example
	infoParams[3] = args.copyright
	return infoParams
}

func (args *arguments) commandParameters() []*cmdl.Parameter {
	cmdParams := make([]*cmdl.Parameter, 6)
	cmdParams[0] = args.command
	cmdParams[1] = args.filter
	cmdParams[2] = args.input
	cmdParams[3] = args.or
	cmdParams[4] = args.output
	cmdParams[5] = args.recursive
	return cmdParams
}

func (args *arguments) infoAvailable() bool {
	if args.help == nil || !args.command.Available() {
		return true
	}
	return false
}

func (args *arguments) printInfo() {
	if args.help == nil {
		fmt.Println(messageShortInfo())
	} else if args.help.Available() {
		fmt.Println(messageHelp())
	} else if args.version.Available() {
		fmt.Println(messageVersion())
	} else if args.example.Available() {
		fmt.Println(messageExample())
	} else if args.copyright.Available() {
		fmt.Println(messageCopyright())
	} else {
		fmt.Println(messageShortInfo())
	}
}

func parseInput(args *arguments, unparsedCLArgs []string) []string {
	// just accept the first unparsed argument
	if len(unparsedCLArgs) > 0 {
		inputPath := unparsedCLArgs[0]
		if !filepath.IsAbs(inputPath) {
			wd, err := os.Getwd()
			if err == nil {
				inputPath = filepath.Join(wd, inputPath)
			} else {
				panic(err.Error())
			}
		}
		args.input.Add(keyNONE, inputPath)
		return unparsedCLArgs[1:]
	}
	return unparsedCLArgs
}

func parseOutput(args *arguments, unparsedCLArgs []string) []string {
	// just accept the first unparsed argument
	if outputDirNeeded(args) && len(unparsedCLArgs) > 0 {
		outputPath := unparsedCLArgs[0]
		if !filepath.IsAbs(outputPath) {
			wd, err := os.Getwd()
			if err == nil {
				outputPath = wd + string(filepath.Separator) + outputPath
			} else {
				panic(err.Error())
			}
		}
		args.output.Add(keyNONE, outputPath)
		return unparsedCLArgs[1:]
	}
	return unparsedCLArgs
}

func outputDirNeeded(args *arguments) bool {
	if args.command.Available() {
		command := args.command.Keys()[0]
		if command == argCP || command == argMV {
			return true
		}
	}
	return false
}

func parseFilter(args *arguments, unparsedCLArgs []string) {
	for _, unparsedCLArg := range unparsedCLArgs {
		args.filter.Add(keyNONE, unparsedCLArg)
	}
	args.contentFilter = make([]string, 0, args.filter.Count())
	for _, value := range args.filter.Values() {
		if len(value) > 0 {
			args.contentFilter = append(args.contentFilter, value)
		}
	}
}

func parseInputFilter(args *arguments) {
	if args.input.Available() {
		input := args.input.Values()[0]
		separator := pathSeparator(input)

		if separator == 0 {
			separator = filepath.Separator
		}
		fileNameBegin := rindex(input, separator) + 1
		fileName := input[fileNameBegin:]
		wildCardUsed := rindex(fileName, '*') >= 0

		if wildCardUsed {
			input = input[:fileNameBegin]
			args.inputFilter = fileName
			if len(input) > 0 {
				args.input.Values()[0] = input
			} else {
				args.input.Values()[0] = "."
			}
		} else {
			args.inputFilter = "*"
		}
	}
}

func pathSeparator(str string) byte {
	bytes := *(*[]byte)(unsafe.Pointer(&str))
	for i := len(bytes) - 1; i >= 0; i-- {
		b := bytes[i]
		if b == '/' || b == '\\' {
			return b
		}
	}
	return 0
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

func validateParameters(args *arguments) error {
	var err error
	infoParams := args.infoParameters()
	cmdParams := args.commandParameters()

	if parametersIncompatible(infoParams, cmdParams) || parametersMultiple(args) {
		err = errors.New("wrong argument usage")

	} else if !anyAvailable(infoParams) && anyAvailable(cmdParams) {
		if !args.command.Available() {
			err = errors.New("command missing")
		} else {
			err = validateIODirectories(args)
		}
	}
	return err
}

func validateIODirectories(args *arguments) error {
	var err error
	if !args.input.Available() {
		err = errors.New("input directory is not specified")
	} else if outputDirNeeded(args) && !args.output.Available() {
		err = errors.New("output directory is not specified")
	} else {
		err = validateDirectory(args.input.Values()[0], "input")
		if err == nil && args.output.Available() {
			err = validateDirectory(args.output.Values()[0], "output")
			if err == nil {
				var input string
				input, err = filepath.Abs(args.input.Values()[0])
				if err == nil {
					var output string
					output, err = filepath.Abs(args.output.Values()[0])
					if err == nil {
						if input != output {
							args.input.Values()[0] = input
							args.output.Values()[0] = output
						} else {
							err = errors.New("input and output directories are the same")
						}
					}
				}
			}
		}
	}
	return err
}

func validateDirectory(path, dirTypeName string) error {
	var err error
	fileInfo, statErr := os.Stat(path)
	if statErr == nil || !os.IsNotExist(statErr) {
		if fileInfo != nil {
			if !fileInfo.IsDir() {
				err = errors.New(dirTypeName + " path is a file, but must be a directory")
			}
		} else {
			err = errors.New("wrong " + dirTypeName + " path syntax")
		}
	} else {
		err = errors.New(dirTypeName + " directory does not exist")
	}
	return err
}

func parametersIncompatible(infoParams, cmdParams []*cmdl.Parameter) bool {
	var cmdAvailable bool
	for _, cmdParam := range cmdParams {
		cmdAvailable = cmdAvailable || cmdParam.Available()
	}
	for i, infoParamA := range infoParams {
		if infoParamA.Available() {
			// either info or command
			if cmdAvailable {
				return true
			}
			// only one info parameter is allowed
			for j, infoParamB := range infoParams {
				if j != i && infoParamB.Available() {
					return true
				}
			}
		}
	}
	return false
}

func parametersMultiple(args *arguments) bool {
	params := make([]*cmdl.Parameter, 10)
	params[0] = args.command
	params[1] = args.copyright
	params[2] = args.example
	params[3] = args.help
	params[4] = args.input
	params[5] = args.or
	params[6] = args.silent
	params[7] = args.output
	params[8] = args.recursive
	params[9] = args.version
	for _, param := range params {
		if param.Count() > 1 {
			return true
		}
	}
	return false
}

func anyAvailable(params []*cmdl.Parameter) bool {
	for _, param := range params {
		if param.Available() {
			return true
		}
	}
	return false
}
