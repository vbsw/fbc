/*
 *          Copyright 2020, Vitali Baumtrok.
 * Distributed under the Boost Software License, Version 1.0.
 *     (See accompanying file LICENSE or copy at
 *        http://www.boost.org/LICENSE_1_0.txt)
 */

package main

import (
	"github.com/vbsw/osargs"
)

const (
	PARAM_COUNT = "count"
	PARAM_CP    = "cp"
	PARAM_MV    = "mv"
	PARAM_PRINT = "print"
	PARAM_RM    = "rm"
)

type parameters struct {
	commandId []osargs.Parameter
	copyright []osargs.Parameter
	example   []osargs.Parameter
	filter    []osargs.Parameter
	help      []osargs.Parameter
	input     []osargs.Parameter
	or        []osargs.Parameter
	output    []osargs.Parameter
	recursive []osargs.Parameter
	version   []osargs.Parameter
}

func parametersFromArgs(args []string) (*parameters, error) {
	var params *parameters
	var err error
	osArgs := osargs.NewFromArgs(args, " ", "=", "")

	if len(osArgs.Str) > 0 {
		params = new(parameters)
		params.help = osArgs.Parse("-h", "--help", "-help", "help")
		params.version = osArgs.Parse("-v", "--version", "-version", "version")
		params.example = osArgs.Parse("-e", "--example", "-example", "example")
		params.copyright = osArgs.Parse("-c", "--copyright", "-copyright", "copyright")
		params.or = osArgs.Parse("-o", "--or", "-or", "or")
		params.commandId = osArgs.Parse(PARAM_COUNT, PARAM_CP, PARAM_MV, PARAM_PRINT, PARAM_RM)
		params.recursive = osArgs.Parse("-r", "--recursive", "-recursive", "recursive")

		// no need for it (only noise? also conflict with "-o" parameter)
		// params.filter = osArgs.ParsePairs("-f", "--filter", "-filter", "filter")
		// params.input = osArgs.ParsePairs("-i", "--input", "-input", "input")
		// params.output = osArgs.ParsePairs("-o", "--output", "-output", "output")

		unparsedArgs := osArgs.Rest(params.allParams()...)
		unparsedArgs = params.parseInput(osArgs, unparsedArgs)
		unparsedArgs = params.parseOutput(osArgs, unparsedArgs)
		params.parseFilter(osArgs, unparsedArgs)
	}
	return params, err
}

func (params *parameters) parseInput(osArgs *osargs.Arguments, unparsedArgs []int) []int {
	if len(params.input) == 0 {
		if len(unparsedArgs) > 0 {
			index := unparsedArgs[0]
			value := osArgs.Str[index]
			params.input = append(params.input, osargs.Parameter{"<none>", value, "", index})
			unparsedArgs = unparsedArgs[1:]
		}
	}
	return unparsedArgs
}

func (params *parameters) parseOutput(osArgs *osargs.Arguments, unparsedArgs []int) []int {
	if len(params.output) == 0 {
		if params.isOutputDirNeeded() && len(unparsedArgs) > 0 {
			index := unparsedArgs[0]
			value := osArgs.Str[index]
			params.output = append(params.output, osargs.Parameter{"<none>", value, "", index})
			unparsedArgs = unparsedArgs[1:]
		}
	}
	return unparsedArgs
}

func (params *parameters) parseFilter(osArgs *osargs.Arguments, unparsedArgs []int) []int {
	key := "<none>"
	op := ""
	for _, i := range unparsedArgs {
		value := osArgs.Str[i]
		params.filter = append(params.filter, osargs.Parameter{key, value, op, i})
	}
	return unparsedArgs[:0]
}

func (params *parameters) incompatibleArguments() bool {
	nfoParams := params.infoParams()
	cmdAvailable := len(params.commandId) > 0

	for i, nfoParamA := range nfoParams {
		if len(nfoParamA) > 0 {
			// either info or command
			if cmdAvailable {
				return true
				// only one info parameter is allowed
			} else {
				for j, nfoParamB := range nfoParams {
					if j != i && len(nfoParamB) > 0 {
						return true
					}
				}
			}
		}
	}
	return false
}

func (params *parameters) oneParamHasMultipleResults() bool {
	for _, param := range params.singleValueParams() {
		if len(param) > 1 {
			return true
		}
	}
	return false
}

func (params *parameters) isOutputDirNeeded() bool {
	if len(params.commandId) > 0 {
		commandId := params.commandId[0].Key
		if commandId == PARAM_CP || commandId == PARAM_MV {
			return true
		}
	}
	return false
}

func (params *parameters) infoParams() [][]osargs.Parameter {
	nfoParams := make([][]osargs.Parameter, 4)
	nfoParams[0] = params.help
	nfoParams[1] = params.version
	nfoParams[2] = params.example
	nfoParams[3] = params.copyright
	return nfoParams
}

func (params *parameters) singleValueParams() [][]osargs.Parameter {
	prms := make([][]osargs.Parameter, 9)
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

func (params *parameters) allParams() [][]osargs.Parameter {
	prms := make([][]osargs.Parameter, 10)
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

func paramsToStringArray(params []osargs.Parameter) []string {
	strings := make([]string, 0, len(params))
	for _, param := range params {
		if len(param.Value) > 0 {
			strings = append(strings, param.Value)
		}
	}
	return strings
}
