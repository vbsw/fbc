/*
 *          Copyright 2020, Vitali Baumtrok.
 * Distributed under the Boost Software License, Version 1.0.
 *     (See accompanying file LICENSE or copy at
 *        http://www.boost.org/LICENSE_1_0.txt)
 */

// Package fbc is compiled to an executable. It allows various commands on files filtered by their content.
package main

import (
	pkgbytes "bytes"
	"fmt"
	"github.com/vbsw/golib/check"
	"os"
	"path/filepath"
)

func main() {
	var params parameters
	err := params.initFromOSArgs()

	if err == nil {
		if params.infoAvailable() {
			params.printInfo()

		} else {
			fileProc := newFileProcessor(params.command.Values[0])
			iterate(&params, fileProc)
		}
	} else {
		fmt.Println(messageError(err))
	}
}

// iterate iterates over files calling fileProc.processFile for each file
// matching the criteria. If fileProc.processFile returns error != nil,
// then processing is stopped.
func iterate(params *parameters, fileProc fileProcessor) error {
	var err error
	if fileProc == nil {
		fileProc = new(fileProcessorDefault)
	}
	if params.recursive.Available() {
		err = iterateRecursive(params, fileProc)
	} else {
		err = iterateFlat(params, fileProc)
	}
	return err
}

func iterateRecursive(params *parameters, fileProc fileProcessor) error {
	inputDir := params.input.Values[0]
	byOr := params.or.Available()
	silent := params.silent.Available()
	fileNameFilter := splitStringByStar(params.inputFilter)
	buffer := make([]byte, 1024*1024*4)
	count := 0
	terms := toBytes(params.contentFilter)
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() {
			// avoid input directory as input file
			if len(path) > len(inputDir) {
				var hasTerms bool
				hasTerms, err = fileHasTerms(byOr, path, info.Name(), buffer, fileNameFilter, terms)
				if hasTerms && err == nil {
					err = fileProc.processFile(params, path, info)
					if err == nil {
						count++
					}
				}
			}
		}
		// ignore errors
		if err != nil {
			if !silent && !os.IsNotExist(err) {
				fmt.Println(messageWarning(err))
			}
			err = nil
		}
		return err
	})
	fileProc.summary(count, err)
	return err
}

func iterateFlat(params *parameters, fileProc fileProcessor) error {
	inputDir := params.input.Values[0]
	byOr := params.or.Available()
	silent := params.silent.Available()
	fileNameFilter := splitStringByStar(params.inputFilter)
	buffer := make([]byte, 1024*1024*4)
	count := 0
	terms := toBytes(params.contentFilter)
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() {
			// avoid input directory as input file; parent must be input directory
			if len(path) > len(inputDir) && len(filepath.Dir(path)) == len(inputDir) {
				var hasTerms bool
				hasTerms, err = fileHasTerms(byOr, path, info.Name(), buffer, fileNameFilter, terms)
				if hasTerms && err == nil {
					err = fileProc.processFile(params, path, info)
					if err == nil {
						count++
					}
				}
			}
		}
		// ignore errors
		if err != nil {
			if !silent && !os.IsNotExist(err) {
				fmt.Println(messageWarning(err))
			}
			err = nil
		}
		return err
	})
	fileProc.summary(count, err)

	return err
}

func fileHasTerms(byOr bool, path, fileName string, buffer []byte, fileNameFilter, terms [][]byte) (bool, error) {
	if isFilterMatch(fileName, fileNameFilter) {
		if len(terms) > 0 {
			if byOr {
				return check.FileHasAny(path, buffer, terms)
			}
			return check.FileHasAll(path, buffer, terms)
		}
		return true, nil
	}
	return false, nil
}

func isFilterMatch(str string, filter [][]byte) bool {
	bytes := []byte(str)
	if len(filter) > 0 {
		if pkgbytes.HasPrefix(bytes, filter[0]) {
			offset := len(filter[0])
			for _, part := range filter[1:] {
				if len(part) > 0 {
					offsetPrev, limit := offset, len(bytes)-len(part)+1
					for i := offset; i < limit; i++ {
						if pkgbytes.HasPrefix(bytes[i:], part) {
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
			return offset == len(bytes)
		}
		return false
	}
	return true
}

func toBytes(contentFilter []string) [][]byte {
	terms := make([][]byte, len(contentFilter))
	for i, term := range contentFilter {
		terms[i] = []byte(term)
	}
	return terms
}
