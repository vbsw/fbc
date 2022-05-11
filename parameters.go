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
	"github.com/vbsw/golib/osargs"
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
)

type parameters struct {
	help          *osargs.Result
	version       *osargs.Result
	example       *osargs.Result
	copyright     *osargs.Result
	or            *osargs.Result
	silent        *osargs.Result
	command       *osargs.Result
	recursive     *osargs.Result
	filter        *osargs.Result
	input         *osargs.Result
	output        *osargs.Result
	contentFilter []string
	inputFilter   string
}

func (params *parameters) initFromOSArgs() error {
	args := osargs.New()
	err := params.initFromArgs(args)
	return err
}

// initFromArgs is for test purposes.
func (params *parameters) initFromArgs(args *osargs.Arguments) error {
	var err error
	if len(args.Values) > 0 {
		params.help = args.Parse("-h", "--help", "-help", "help")
		params.version = args.Parse("-v", "--version", "-version", "version")
		params.example = args.Parse("-e", "--example", "-example", "example")
		params.copyright = args.Parse("-c", "--copyright", "-copyright", "copyright")
		params.or = args.Parse("-o", "--or", "-or", "or")
		params.silent = args.Parse("-s", "--silent", "-silent", "silent")
		params.command = args.Parse(argCOUNT, argCP, argMV, argPRINT, argRM)
		params.recursive = args.Parse("-r", "--recursive", "-recursive", "recursive")
		params.filter = new(osargs.Result)
		params.input = new(osargs.Result)
		params.output = new(osargs.Result)

		unparsedArgs := args.UnparsedArgs()
		unparsedArgs = parseInput(params, unparsedArgs)
		unparsedArgs = parseOutput(params, unparsedArgs)
		parseFilter(params, unparsedArgs)
		parseInputFilter(params)

		err = validateParameters(params)
	}
	return err
}

func (params *parameters) infoParameters() []*osargs.Result {
	paramsInfo := make([]*osargs.Result, 4)
	paramsInfo[0] = params.help
	paramsInfo[1] = params.version
	paramsInfo[2] = params.example
	paramsInfo[3] = params.copyright
	return paramsInfo
}

func (params *parameters) commandParameters() []*osargs.Result {
	paramsCmd := make([]*osargs.Result, 6)
	paramsCmd[0] = params.command
	paramsCmd[1] = params.filter
	paramsCmd[2] = params.input
	paramsCmd[3] = params.or
	paramsCmd[4] = params.output
	paramsCmd[5] = params.recursive
	return paramsCmd
}

func (params *parameters) infoAvailable() bool {
	if params.help == nil || !params.command.Available() {
		return true
	}
	return false
}

func (params *parameters) printInfo() {
	if params.help == nil {
		fmt.Println(messageShortInfo())
	} else if params.help.Available() {
		fmt.Println(messageHelp())
	} else if params.version.Available() {
		fmt.Println(messageVersion())
	} else if params.example.Available() {
		fmt.Println(messageExample())
	} else if params.copyright.Available() {
		fmt.Println(messageCopyright())
	} else {
		fmt.Println(messageShortInfo())
	}
}

func parseInput(params *parameters, unparsedArgs []string) []string {
	// just accept the first unparsed argument
	if len(unparsedArgs) > 0 {
		inputPath := unparsedArgs[0]
		if !filepath.IsAbs(inputPath) {
			wd, err := os.Getwd()
			if err == nil {
				inputPath = filepath.Join(wd, inputPath)
			} else {
				panic(err.Error())
			}
		}
		params.input.Values = append(params.input.Values, inputPath)
		return unparsedArgs[1:]
	}
	return unparsedArgs
}

func parseOutput(params *parameters, unparsedArgs []string) []string {
	// just accept the first unparsed argument
	if outputDirNeeded(params) && len(unparsedArgs) > 0 {
		outputPath := unparsedArgs[0]
		if !filepath.IsAbs(outputPath) {
			wd, err := os.Getwd()
			if err == nil {
				outputPath = wd + string(filepath.Separator) + outputPath
			} else {
				panic(err.Error())
			}
		}
		params.output.Values = append(params.output.Values, outputPath)
		return unparsedArgs[1:]
	}
	return unparsedArgs
}

func outputDirNeeded(params *parameters) bool {
	if params.command.Available() {
		command := params.command.Values[0]
		if command == argCP || command == argMV {
			return true
		}
	}
	return false
}

func parseFilter(params *parameters, unparsedArgs []string) {
	for _, unparsedCLArg := range unparsedArgs {
		params.filter.Values = append(params.filter.Values, unparsedCLArg)
	}
	params.contentFilter = make([]string, 0, params.filter.Count())
	for _, value := range params.filter.Values {
		if len(value) > 0 {
			params.contentFilter = append(params.contentFilter, value)
		}
	}
}

func parseInputFilter(params *parameters) {
	if params.input.Available() {
		input := params.input.Values[0]
		separator := pathSeparator(input)

		if separator == 0 {
			separator = filepath.Separator
		}
		fileNameBegin := rindex(input, separator) + 1

		fileName := input[fileNameBegin:]
		wildCardUsed := rindex(fileName, '*') >= 0

		if wildCardUsed {
			input = input[:fileNameBegin]
			params.inputFilter = fileName
			if len(input) > 0 {
				params.input.Values[0] = input
			} else {
				params.input.Values[0] = "."
			}
		} else {
			params.inputFilter = "*"
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

func validateParameters(params *parameters) error {
	var err error
	paramsInfo := params.infoParameters()
	paramsCmd := params.commandParameters()

	if parametersIncompatible(paramsInfo, paramsCmd) || parametersMultiple(params) {
		err = errors.New("wrong argument usage")

	} else if !anyAvailable(paramsInfo) && anyAvailable(paramsCmd) {
		if params.command.Available() {
			err = validateIODirectories(params)
		} else {
			err = errors.New("command missing")
		}
	}
	return err
}

func validateIODirectories(params *parameters) error {
	var err error
	if !params.input.Available() {
		err = errors.New("input directory is not specified")
	} else if outputDirNeeded(params) && !params.output.Available() {
		err = errors.New("output directory is not specified")
	} else {
		err = validateDirectory(params.input.Values[0], "input")
		if err == nil {
			var input string
			input, err = filepath.Abs(params.input.Values[0])
			if err == nil {
				params.input.Values[0] = input
				if params.output.Available() {
					err = validateDirectory(params.output.Values[0], "output")
					if err == nil {
						var output string
						output, err = filepath.Abs(params.output.Values[0])
						if err == nil {
							if input != output {
								params.output.Values[0] = output
							} else {
								err = errors.New("input and output directories are the same")
							}
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
	info, errInfo := os.Stat(path)
	if errInfo == nil || !os.IsNotExist(errInfo) {
		if info != nil {
			if !info.IsDir() {
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

func parametersIncompatible(paramsInfo, paramsCmd []*osargs.Result) bool {
	// either info or command
	for _, paramCmd := range paramsCmd {
		if paramCmd.Available() {
			for _, paramInfo := range paramsInfo {
				if paramInfo.Available() {
					return true
				}
			}
			break
		}
	}
	// only one info parameter is allowed
	for i, paramInfoA := range paramsInfo {
		if paramInfoA.Available() {
			for j, paramInfoB := range paramsInfo {
				if j != i && paramInfoB.Available() {
					return true
				}
			}
		}
	}
	return false
}

func parametersMultiple(params *parameters) bool {
	paramsMult := make([]*osargs.Result, 10)
	paramsMult[0] = params.command
	paramsMult[1] = params.copyright
	paramsMult[2] = params.example
	paramsMult[3] = params.help
	paramsMult[4] = params.input
	paramsMult[5] = params.or
	paramsMult[6] = params.silent
	paramsMult[7] = params.output
	paramsMult[8] = params.recursive
	paramsMult[9] = params.version
	for _, param := range paramsMult {
		if param.Count() > 1 {
			return true
		}
	}
	return false
}

func anyAvailable(params []*osargs.Result) bool {
	for _, param := range params {
		if param.Available() {
			return true
		}
	}
	return false
}
