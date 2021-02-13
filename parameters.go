/*
 *        Copyright 2020, 2021 Vitali Baumtrok.
 * Distributed under the Boost Software License, Version 1.0.
 *     (See accompanying file LICENSE or copy at
 *        http://www.boost.org/LICENSE_1_0.txt)
 */

package main

import (
	"github.com/vbsw/cmdl"
)

const (
	argCOUNT = "count"
	argCP    = "cp"
	argMV    = "mv"
	argPRINT = "print"
	argRM    = "rm"
)

type parameters struct {
	commandId *cmdl.Parameter
	copyright *cmdl.Parameter
	example   *cmdl.Parameter
	filter    *cmdl.Parameter
	help      *cmdl.Parameter
	input     *cmdl.Parameter
	or        *cmdl.Parameter
	output    *cmdl.Parameter
	recursive *cmdl.Parameter
	version   *cmdl.Parameter
}

func parametersFromCL() (*parameters, error) {
	var params *parameters
	var err error
	cl := cmdl.New()

	if len(cl.Args()) > 0 {
		params = new(parameters)
		params.help = cl.NewParam().Parse("-h", "--help", "-help", "help")
		params.version = cl.NewParam().Parse("-v", "--version", "-version", "version")
		params.example = cl.NewParam().Parse("-e", "--example", "-example", "example")
		params.copyright = cl.NewParam().Parse("-c", "--copyright", "-copyright", "copyright")
		params.or = cl.NewParam().Parse("-o", "--or", "-or", "or")
		params.commandId = cl.NewParam().Parse(argCOUNT, argCP, argMV, argPRINT, argRM)
		params.recursive = cl.NewParam().Parse("-r", "--recursive", "-recursive", "recursive")
		params.filter = cl.NewParam()
		params.input = cl.NewParam()
		params.output = cl.NewParam()

		// no need for it (only noise? also conflicts with "-o" parameter)
		// ops := []string { " ", "=", "" }
		// params.filter = cl.NewParam().ParsePairs(ops, "-f", "--filter", "-filter", "filter")
		// params.input = cl.NewParam().ParsePairs(ops, "-i", "--input", "-input", "input")
		// params.output = cl.NewParam().ParsePairs(ops, "-o", "--output", "-output", "output")

		unparsedArgs := cl.UnparsedArgs()
		unparsedArgs = params.parseInput(unparsedArgs)
		unparsedArgs = params.parseOutput(unparsedArgs)
		params.parseFilter(unparsedArgs)
	}
	return params, err
}

func (params *parameters) parseInput(unparsedArgs []string) []string {
	if !params.input.Available() {
		// just accept the first unparsed argument, if input wasn't set explicitly
		if len(unparsedArgs) > 0 {
			params.input.Add("<none>", unparsedArgs[0])
			unparsedArgs = unparsedArgs[1:]
		}
	}
	return unparsedArgs
}

func (params *parameters) parseOutput(unparsedArgs []string) []string {
	if !params.output.Available() {
		// just accept the first unparsed argument, if output wasn't set explicitly
		if params.isOutputDirNeeded() && len(unparsedArgs) > 0 {
			params.input.Add("<none>", unparsedArgs[0])
			unparsedArgs = unparsedArgs[1:]
		}
	}
	return unparsedArgs
}

// parseFilter extracts filter phrases, i.e. words to search for
func (params *parameters) parseFilter(unparsedArgs []string) {
	for _, unparsedArg := range unparsedArgs {
		params.filter.Add("<none>", unparsedArg)
	}
}

func (params *parameters) incompatibleParameters() bool {
	nfoParams := params.infoParams()
	cmdAvailable := params.commandId.Available()

	for i, nfoParamA := range nfoParams {
		if nfoParamA.Available() {
			// either info or command
			if cmdAvailable {
				return true
			}
			// only one info parameter is allowed
			for j, nfoParamB := range nfoParams {
				if j != i && nfoParamB.Available() {
					return true
				}
			}
		}
	}
	return false
}

func (params *parameters) oneParamHasMultipleResults() bool {
	for _, param := range params.singleValueParams() {
		if param.Count() > 1 {
			return true
		}
	}
	return false
}

func (params *parameters) isOutputDirNeeded() bool {
	if !params.commandId.Available() {
		commandId := params.commandId.Keys()[0]
		if commandId == argCP || commandId == argMV {
			return true
		}
	}
	return false
}

func (params *parameters) infoParams() []*cmdl.Parameter {
	nfoParams := make([]*cmdl.Parameter, 4)
	nfoParams[0] = params.help
	nfoParams[1] = params.version
	nfoParams[2] = params.example
	nfoParams[3] = params.copyright
	return nfoParams
}

func (params *parameters) singleValueParams() []*cmdl.Parameter {
	prms := make([]*cmdl.Parameter, 9)
	prms[0] = params.commandId
	prms[1] = params.copyright
	prms[2] = params.example
	prms[3] = params.help
	prms[4] = params.input
	prms[5] = params.or
	prms[6] = params.output
	prms[7] = params.recursive
	prms[8] = params.version
	return prms
}

func (params *parameters) allParams() []*cmdl.Parameter {
	prms := make([]*cmdl.Parameter, 10)
	prms[0] = params.commandId
	prms[1] = params.copyright
	prms[2] = params.example
	prms[3] = params.filter
	prms[4] = params.help
	prms[5] = params.input
	prms[6] = params.or
	prms[7] = params.output
	prms[8] = params.recursive
	prms[9] = params.version
	return prms
}

func paramsToStringArray(param *cmdl.Parameter) []string {
	strings := make([]string, 0, param.Count())
	for _, value := range param.Values() {
		if len(value) > 0 {
			strings = append(strings, value)
		}
	}
	return strings
}
