/*
 *          Copyright 2020, Vitali Baumtrok.
 * Distributed under the Boost Software License, Version 1.0.
 *     (See accompanying file LICENSE or copy at
 *        http://www.boost.org/LICENSE_1_0.txt)
 */

package main

import (
	"github.com/vbsw/cl"
)

const (
	argCOUNT = "count"
	argCP    = "cp"
	argMV    = "mv"
	argPRINT = "print"
	argRM    = "rm"
)

type arguments struct {
	commandId []cl.Argument
	copyright []cl.Argument
	example   []cl.Argument
	filter    []cl.Argument
	help      []cl.Argument
	input     []cl.Argument
	or        []cl.Argument
	output    []cl.Argument
	recursive []cl.Argument
	version   []cl.Argument
}

func argumentsFromCL(osArgs []string) (*arguments, error) {
	var args *arguments
	var err error
	cmdLine := cl.New(osArgs)

	if len(cmdLine.Args) > 0 {
		args = new(arguments)
		args.help = cmdLine.Parse("-h", "--help", "-help", "help")
		args.version = cmdLine.Parse("-v", "--version", "-version", "version")
		args.example = cmdLine.Parse("-e", "--example", "-example", "example")
		args.copyright = cmdLine.Parse("-c", "--copyright", "-copyright", "copyright")
		args.or = cmdLine.Parse("-o", "--or", "-or", "or")
		args.commandId = cmdLine.Parse(argCOUNT, argCP, argMV, argPRINT, argRM)
		args.recursive = cmdLine.Parse("-r", "--recursive", "-recursive", "recursive")

		// no need for it (only noise? also conflicts with "-o" parameter)
		// ops := []string { " ", "=", "" }
		// args.filter = cmdLine.ParsePairs(ops, "-f", "--filter", "-filter", "filter")
		// args.input = cmdLine.ParsePairs(ops, "-i", "--input", "-input", "input")
		// args.output = cmdLine.ParsePairs(ops, "-o", "--output", "-output", "output")

		unparsedArgs := cmdLine.UnparsedArgsIndices()
		unparsedArgs = args.parseInput(cmdLine, unparsedArgs)
		unparsedArgs = args.parseOutput(cmdLine, unparsedArgs)
		args.parseFilter(cmdLine, unparsedArgs)
	}
	return args, err
}

func (args *arguments) parseInput(cmdLine *cl.CommandLine, unparsedArgs []int) []int {
	if len(args.input) == 0 {
		if len(unparsedArgs) > 0 {
			index := unparsedArgs[0]
			value := cmdLine.Args[index]
			args.input = append(args.input, cl.Argument{"<none>", value, "", index})
			unparsedArgs = unparsedArgs[1:]
		}
	}
	return unparsedArgs
}

func (args *arguments) parseOutput(cmdLine *cl.CommandLine, unparsedArgs []int) []int {
	if len(args.output) == 0 {
		if args.isOutputDirNeeded() && len(unparsedArgs) > 0 {
			index := unparsedArgs[0]
			value := cmdLine.Args[index]
			args.output = append(args.output, cl.Argument{"<none>", value, "", index})
			unparsedArgs = unparsedArgs[1:]
		}
	}
	return unparsedArgs
}

// parseFilter extracts filter phrases, i.e. words to search for
func (args *arguments) parseFilter(cmdLine *cl.CommandLine, unparsedArgs []int) []int {
	key := "<none>"
	op := ""
	for _, i := range unparsedArgs {
		value := cmdLine.Args[i]
		args.filter = append(args.filter, cl.Argument{key, value, op, i})
	}
	return unparsedArgs[:0]
}

func (args *arguments) incompatibleArguments() bool {
	nfoParams := args.infoParams()
	cmdAvailable := len(args.commandId) > 0

	for i, nfoParamA := range nfoParams {
		if len(nfoParamA) > 0 {
			// either info or command
			if cmdAvailable {
				return true
			}
			// only one info parameter is allowed
			for j, nfoParamB := range nfoParams {
				if j != i && len(nfoParamB) > 0 {
					return true
				}
			}
		}
	}
	return false
}

func (args *arguments) oneParamHasMultipleResults() bool {
	for _, param := range args.singleValueParams() {
		if len(param) > 1 {
			return true
		}
	}
	return false
}

func (args *arguments) isOutputDirNeeded() bool {
	if len(args.commandId) > 0 {
		commandId := args.commandId[0].Key
		if commandId == argCP || commandId == argMV {
			return true
		}
	}
	return false
}

func (args *arguments) infoParams() [][]cl.Argument {
	nfoParams := make([][]cl.Argument, 4)
	nfoParams[0] = args.help
	nfoParams[1] = args.version
	nfoParams[2] = args.example
	nfoParams[3] = args.copyright
	return nfoParams
}

func (args *arguments) singleValueParams() [][]cl.Argument {
	prms := make([][]cl.Argument, 9)
	prms[0] = args.commandId
	prms[1] = args.copyright
	prms[2] = args.example
	prms[3] = args.help
	prms[4] = args.input
	prms[5] = args.or
	prms[6] = args.output
	prms[7] = args.recursive
	prms[8] = args.version
	return prms
}

func (args *arguments) allParams() [][]cl.Argument {
	prms := make([][]cl.Argument, 10)
	prms[0] = args.commandId
	prms[1] = args.copyright
	prms[2] = args.example
	prms[3] = args.filter
	prms[4] = args.help
	prms[5] = args.input
	prms[6] = args.or
	prms[7] = args.output
	prms[8] = args.recursive
	prms[9] = args.version
	return prms
}

func argsToStringArray(clArgs []cl.Argument) []string {
	strings := make([]string, 0, len(clArgs))
	for _, clArg := range clArgs {
		if len(clArg.Value) > 0 {
			strings = append(strings, clArg.Value)
		}
	}
	return strings
}
