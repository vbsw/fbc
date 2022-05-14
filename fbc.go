/*
 *       Copyright 2020 - 2022, Vitali Baumtrok.
 * Distributed under the Boost Software License, Version 1.0.
 *     (See accompanying file LICENSE or copy at
 *        http://www.boost.org/LICENSE_1_0.txt)
 */

// Package main is compiled to an executable. It allows various commands on files filtered by their content.
package main

import (
	"errors"
	"fmt"
	"github.com/vbsw/golib/check"
	"github.com/vbsw/golib/iter"
	"github.com/vbsw/golib/osargs"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

const (
	argCOUNT = "count"
	argCP    = "cp"
	argMV    = "mv"
	argPRINT = "print"
	argRM    = "rm"
)

type tParameters struct {
	help           *osargs.Result
	version        *osargs.Result
	example        *osargs.Result
	copyright      *osargs.Result
	or             *osargs.Result
	silent         *osargs.Result
	command        *osargs.Result
	recursive      *osargs.Result
	input          *osargs.Result
	output         *osargs.Result
	contentFilter  []string
	fileNameFilter string
}

type tFileProcessor interface {
	iter.FileProcessor
	printSummary(err error)
}

type tFileProcessorDefault struct {
	count          int
	silent         bool
	or             bool
	contentFilter  [][]byte
	fileNameFilter [][]byte
	buffer         []byte
}

type tFileProcessorCount struct {
	tFileProcessorDefault
}

type tFileProcessorCP struct {
	tFileProcessorDefault
	existingDirs   []string
	inputDirLength int
	outputDir      string
}

type tFileProcessorMV struct {
	tFileProcessorCP
}

type tFileProcessorPrint struct {
	tFileProcessorDefault
}

type tFileProcessorRM struct {
	tFileProcessorDefault
}

func main() {
	var params tParameters
	err := params.initFromOSArgs()
	if err == nil {
		if params.infoAvailable() {
			printInfo(&params)
		} else {
			proc := newFileProcessor(&params)
			if params.recursive.Available() {
				err = iter.IterateFilesRecr(params.inputDir(), proc)
			} else {
				err = iter.IterateFilesFlat(params.inputDir(), proc)
			}
			proc.printSummary(err)
		}
	} else {
		printError(err)
	}
}

func (params *tParameters) initFromOSArgs() error {
	args := osargs.New()
	err := params.initFromArgs(args)
	return err
}

func (params *tParameters) inputDir() string {
	return params.input.Values[0]
}

func (params *tParameters) infoAvailable() bool {
	if params.help == nil || !params.command.Available() {
		return true
	}
	return false
}

// initFromArgs is for test purposes.
func (params *tParameters) initFromArgs(args *osargs.Arguments) error {
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
		params.input = new(osargs.Result)
		params.output = new(osargs.Result)

		unparsedArgs := args.UnparsedArgs()
		unparsedArgs = params.parseInput(unparsedArgs)
		unparsedArgs = params.parseOutput(unparsedArgs)
		params.parseContentFilter(unparsedArgs)
		params.parseFileNameFilter()

		err = params.validateParameters()
	}
	return err
}

func (params *tParameters) parseInput(unparsedArgs []string) []string {
	// just accept the first unparsed argument
	if len(unparsedArgs) > 0 {
		inputPath, err := filepath.Abs(unparsedArgs[0])
		if err != nil {
			panic(err.Error())
		}
		params.input.Values = append(params.input.Values, inputPath)
		return unparsedArgs[1:]
	}
	return unparsedArgs
}

func (params *tParameters) parseOutput(unparsedArgs []string) []string {
	// just accept the first unparsed argument
	if params.outputDirNeeded() && len(unparsedArgs) > 0 {
		outputPath, err := filepath.Abs(unparsedArgs[0])
		if err != nil {
			panic(err.Error())
		}
		params.output.Values = append(params.output.Values, outputPath)
		return unparsedArgs[1:]
	}
	return unparsedArgs
}

func (params *tParameters) parseContentFilter(unparsedArgs []string) {
	for _, unparsedArg := range unparsedArgs {
		if len(unparsedArg) > 0 {
			params.contentFilter = append(params.contentFilter, unparsedArg)
		}
	}
}

func (params *tParameters) parseFileNameFilter() {
	if params.input.Available() {
		input := params.input.Values[0]
		separator := pathSeparator(input)
		fileNameBegin := rindex(input, separator) + 1
		fileName := input[fileNameBegin:]
		if rindex(fileName, '*') >= 0 {
			// directory; remove ending separator, eventually
			input = filepath.Join(input[:fileNameBegin], ".")
			params.fileNameFilter = fileName
			params.input.Values[0] = input
		} else {
			params.fileNameFilter = "*"
		}
	}
}

func (params *tParameters) validateParameters() error {
	var err error
	paramsInfo := params.infoParameters()
	paramsCmd := params.commandParameters()
	if parametersIncompatible(paramsInfo, paramsCmd) || params.isMultiple() {
		err = errors.New("wrong argument usage")
	} else if anyAvailable(paramsCmd) {
		if params.command.Available() {
			err = params.validateIODirectories()
		} else {
			err = errors.New("command missing")
		}
	}
	return err
}

func (params *tParameters) infoParameters() []*osargs.Result {
	paramsInfo := make([]*osargs.Result, 4)
	paramsInfo[0] = params.help
	paramsInfo[1] = params.version
	paramsInfo[2] = params.example
	paramsInfo[3] = params.copyright
	return paramsInfo
}

func (params *tParameters) commandParameters() []*osargs.Result {
	paramsCmd := make([]*osargs.Result, 5)
	paramsCmd[0] = params.command
	paramsCmd[1] = params.input
	paramsCmd[2] = params.or
	paramsCmd[3] = params.output
	paramsCmd[4] = params.recursive
	return paramsCmd
}

func (params *tParameters) isMultiple() bool {
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

func (params *tParameters) validateIODirectories() error {
	var err error
	if !params.input.Available() {
		err = errors.New("input directory is not specified")
	} else if params.outputDirNeeded() && !params.output.Available() {
		err = errors.New("output directory is not specified")
	} else {
		err = validateDirectory(params.input.Values[0], "input")
		if err == nil && params.output.Available() {
			err = validateDirectory(params.output.Values[0], "output")
			if err == nil {
				if params.input.Values[0] == params.output.Values[0] {
					err = errors.New("input and output directories are the same")
				}
			}
		}
	}
	return err
}

func (params *tParameters) outputDirNeeded() bool {
	if params.command.Available() {
		command := params.command.Values[0]
		if command == argCP || command == argMV {
			return true
		}
	}
	return false
}

func parametersIncompatible(paramsInfo, paramsCmd []*osargs.Result) bool {
	// either info or command
	if anyAvailable(paramsInfo) && anyAvailable(paramsCmd) {
		return true
	}
	// only one info parameter is allowed
	for i, paramInfoA := range paramsInfo {
		if paramInfoA.Available() {
			for _, paramInfoB := range paramsInfo[i+1:] {
				if paramInfoB.Available() {
					return true
				}
			}
		}
	}
	return false
}

func anyAvailable(results []*osargs.Result) bool {
	for _, result := range results {
		if result.Available() {
			return true
		}
	}
	return false
}

func validateDirectory(path, dirType string) error {
	var err error
	info, errInfo := os.Stat(path)
	if errInfo == nil || !os.IsNotExist(errInfo) {
		if info != nil {
			if !info.IsDir() {
				err = errors.New(dirType + " path is a file, but must be a directory")
			}
		} else {
			err = errors.New("wrong " + dirType + " path syntax")
		}
	} else {
		err = errors.New(dirType + " directory does not exist")
	}
	return err
}

func newFileProcessor(params *tParameters) tFileProcessor {
	switch params.command.Values[0] {
	case argCOUNT:
		processorCount := new(tFileProcessorCount)
		processorCount.init(params)
		return processorCount
	case argCP:
		processorCP := new(tFileProcessorCP)
		processorCP.init(params)
		return processorCP
	case argMV:
		processorMV := new(tFileProcessorMV)
		processorMV.init(params)
		return processorMV
	case argPRINT:
		processorPrint := new(tFileProcessorPrint)
		processorPrint.init(params)
		return processorPrint
	case argRM:
		processorRM := new(tFileProcessorRM)
		processorRM.init(params)
		return processorRM
	}
	processorDefault := new(tFileProcessorDefault)
	processorDefault.init(params)
	return processorDefault
}

func (proc *tFileProcessorDefault) init(params *tParameters) {
	proc.silent = params.silent.Available()
	proc.or = params.or.Available()
	proc.contentFilter = toBytes(params.contentFilter)
	proc.fileNameFilter = splitStringByStar(params.fileNameFilter)
	proc.buffer = make([]byte, 1024*1024*4)
}

func (proc *tFileProcessorDefault) ProcessFile(path string, info os.FileInfo, err error) error {
	var match bool
	if err == nil && proc.isFileNameMatch(info.Name()) {
		match, err = proc.isContentMatch(path)
	}
	return proc.postProcess(match, err)
}

func (proc *tFileProcessorDefault) postProcess(match bool, err error) error {
	if err == nil {
		if match {
			proc.count++
		}
	} else if !proc.silent && !os.IsNotExist(err) {
		printWarning(err)
	}
	// ignore errors
	return nil
}

func (proc *tFileProcessorDefault) printSummary(err error) {
	if err == nil {
		printFinished(proc.count)
	} else {
		printError(err)
	}
}

func (proc *tFileProcessorDefault) isFileNameMatch(name string) bool {
	if len(proc.fileNameFilter) > 0 {
		if hasPrefix(name, proc.fileNameFilter[0], 0) {
			offset := len(proc.fileNameFilter[0])
			for _, part := range proc.fileNameFilter[1:] {
				if len(part) > 0 {
					offsetPrev, limit := offset, len(name)-len(part)+1
					for i := offset; i < limit; i++ {
						if hasPrefix(name, part, i) {
							offset = i + len(part)
							break
						}
					}
					if offset == offsetPrev {
						return false
					}
				} else {
					// last part can be empty; this matches rest of string
					return true
				}
			}
			return offset == len(name)
		}
		return false
	}
	return true
}

func (proc *tFileProcessorDefault) isContentMatch(path string) (bool, error) {
	if len(proc.contentFilter) > 0 {
		if proc.or {
			return check.FileHasAny(path, proc.buffer, proc.contentFilter)
		}
		return check.FileHasAll(path, proc.buffer, proc.contentFilter)
	}
	return true, nil
}

func (proc *tFileProcessorCount) printSummary(err error) {
	if err == nil {
		printCount(proc.count)
	} else {
		printError(err)
	}
}

func (proc *tFileProcessorCP) init(params *tParameters) {
	proc.tFileProcessorDefault.init(params)
	proc.existingDirs = make([]string, 0, 16)
	proc.inputDirLength = dirLengthWOEndingSeparator(params.input.Values[0]) + 1
	proc.outputDir = params.output.Values[0]
}

func (proc *tFileProcessorCP) ProcessFile(path string, info os.FileInfo, err error) error {
	var match bool
	if err == nil && proc.isFileNameMatch(info.Name()) {
		match, err = proc.isContentMatch(path)
		if err == nil && match {
			var inputFile *os.File
			inputFile, err = os.Open(path)
			if err == nil {
				var outputFile *os.File
				defer inputFile.Close()
				subDir := path[proc.inputDirLength : len(path)-len(info.Name())]
				outputPath := filepath.Join(proc.outputDir, subDir)
				err = proc.ensureDir(outputPath, subDir)
				if err == nil {
					outputPath = filepath.Join(outputPath, info.Name())
					if !check.FileExists(outputPath) {
						outputFile, err = os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
						if err == nil {
							defer outputFile.Close()
							_, err = io.Copy(outputFile, inputFile)
						}
					} else {
						err = errors.New("target file already exists: " + filepath.Join(subDir, info.Name()))
					}
				}
			}
		}
	}
	return proc.postProcess(match, err)
}

func (proc *tFileProcessorCP) ensureDir(dir, subDir string) error {
	for _, existingDir := range proc.existingDirs {
		if existingDir == dir {
			return nil
		}
	}
	info, err := os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0666)
		if err == nil || check.FileExists(dir) {
			proc.existingDirs = append(proc.existingDirs, dir)
			err = nil
		}
	} else if info != nil && err == nil {
		if info.IsDir() {
			proc.existingDirs = append(proc.existingDirs, dir)
		} else {
			err = errors.New("can't create directory (already exists as file): " + filepath.Join(subDir, info.Name()))
		}
	}
	return err
}

func (proc *tFileProcessorMV) ProcessFile(path string, info os.FileInfo, err error) error {
	var match bool
	if err == nil && proc.isFileNameMatch(info.Name()) {
		match, err = proc.isContentMatch(path)
		if err == nil && match {
			subDir := path[proc.inputDirLength : len(path)-len(info.Name())]
			outputPath := filepath.Join(proc.outputDir, subDir)
			err = proc.ensureDir(outputPath, subDir)
			if err == nil {
				outputPath = filepath.Join(outputPath, info.Name())
				if !check.FileExists(outputPath) {
					err = os.Rename(path, outputPath)
				} else {
					err = errors.New("target file already exists: " + filepath.Join(subDir, info.Name()))
				}
			}
		}
	}
	return proc.postProcess(match, err)
}

func (proc *tFileProcessorPrint) ProcessFile(path string, info os.FileInfo, err error) error {
	var match bool
	if err == nil && proc.isFileNameMatch(info.Name()) {
		match, err = proc.isContentMatch(path)
		if err == nil && match {
			printName(info.Name())
		}
	}
	return proc.postProcess(match, err)
}

func (proc *tFileProcessorPrint) printSummary(err error) {
	if err != nil {
		printError(err)
	}
}

func (proc *tFileProcessorRM) ProcessFile(path string, info os.FileInfo, err error) error {
	var match bool
	if err == nil && proc.isFileNameMatch(info.Name()) {
		match, err = proc.isContentMatch(path)
		if err == nil && match {
			err = os.Remove(path)
		}
	}
	return proc.postProcess(match, err)
}

func hasPrefix(str string, prefix []byte, offset int) bool {
	if len(str) >= len(prefix) {
		for i := 0; i + offset <= len(str); i++ {
			if i < len(prefix) {
				if str[i+offset] != prefix[i] {
					return false
				}
			} else {
				return true
			}
		}
	}
	return false
}

func pathSeparator(path string) byte {
	for i := len(path) - 1; i >= 0; i-- {
		b := path[i]
		if b == '/' || b == '\\' {
			return b
		}
	}
	return filepath.Separator
}

func rindex(str string, b byte) int {
	for i := len(str) - 1; i >= 0; i-- {
		if str[i] == b {
			return i
		}
	}
	return -1
}

func dirLengthWOEndingSeparator(path string) int {
	if b := path[len(path)-1]; b == '/' || b == '\\' {
		return len(path) - 1
	}
	return len(path)
}

func toBytes(strings []string) [][]byte {
	terms := make([][]byte, len(strings))
	for i, term := range strings {
		terms[i] = []byte(term)
	}
	return terms
}

func splitStringByStar(str string) [][]byte {
	parts := make([][]byte, 0, 2)
	bytes := trim([]byte(str))
	if len(bytes) > 0 {
		contentBegin := 0
		for contentBegin < len(bytes) {
			// this is correkt: starBegin 0, then part ""
			starBegin := seekStar(bytes, contentBegin)
			part := bytes[contentBegin:starBegin]
			contentBegin = seekContent(bytes, starBegin)
			parts = append(parts, part)
		}
	}
	return parts
}

func trim(bytes []byte) []byte {
	begin := 0
	end := 0
	for i, b := range bytes {
		if b > 32 {
			begin = i
			break
		}
	}
	for i := len(bytes) - 1; i > 0; i-- {
		b := bytes[i]
		if b > 32 {
			end = i + 1
			break
		}
	}
	return bytes[begin:end]
}

func seekStar(bytes []byte, offset int) int {
	for offset < len(bytes) {
		if bytes[offset] == '*' {
			break
		}
		offset++
	}
	return offset
}

func seekContent(bytes []byte, offset int) int {
	for offset < len(bytes) {
		if bytes[offset] != '*' {
			break
		}
		offset++
	}
	return offset
}

func printInfo(params *tParameters) {
	if params.help == nil {
		printShortInfo()
	} else if params.help.Available() {
		printHelp()
	} else if params.version.Available() {
		printVersion()
	} else if params.example.Available() {
		printExample()
	} else if params.copyright.Available() {
		printCopyright()
	} else {
		printShortInfo()
	}
}

func printShortInfo() {
	fmt.Println("Run 'fbc --help' for usage.")
}

func printHelp() {
	message := "\nUSAGE\n"
	message += "  fbc (INFO | ( COMMAND INPUT-DIR {OUTPUT-DIR FILTER OPTION} ))\n\n"
	message += "INFO\n"
	message += "  -h, --help       print this help\n"
	message += "  -v, --version    print version\n"
	message += "  -e, --example    print example\n"
	message += "  -c, --copyright  print copyright\n"
	message += "COMMAND\n"
	message += "  count            count files\n"
	message += "  cp               copy files\n"
	message += "  mv               move files\n"
	message += "  print            print file names\n"
	message += "  rm               delete files\n"
	message += "OPTION\n"
	message += "  -o, --or         filter is OR (not AND)\n"
	message += "  -r, --recursive  recursive file iteration\n"
	message += "  -s, --silent     don't output errors to screen when reading files"
	fmt.Println(message)
}

func printVersion() {
	fmt.Println("1.1.1")
}

func printExample() {
	message := "\nEXAMPLES\n"
	message += "   fbc cp ./ ../bak bob alice\n"
	message += "   fbc mv \"./*.txt\" ../bak bob alice\n"
	message += "   fbc rm \"./*.txt\" bob alice"
	fmt.Println(message)
}

func printCopyright() {
	message := "Copyright 2020 - 2022, Vitali Baumtrok (vbsw@mailbox.org).\n"
	message += "Distributed under the Boost Software License, Version 1.0."
	fmt.Println(message)
}

func printCount(count int) {
	fmt.Println(count)
}

func printName(name string) {
	fmt.Println(name)
}

func printFinished(count int) {
	var fileStr string
	if count == 1 {
		fileStr = " file"
	} else {
		fileStr = " files"
	}
	fmt.Println("finished: " + strconv.Itoa(count) + fileStr)
}

func printError(err error) {
	fmt.Println("error: " + err.Error())
}

func printWarning(err error) {
	fmt.Println("warning: " + err.Error())
}
