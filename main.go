/*
 *          Copyright 2020, Vitali Baumtrok.
 * Distributed under the Boost Software License, Version 1.0.
 *     (See accompanying file LICENSE or copy at
 *        http://www.boost.org/LICENSE_1_0.txt)
 */

// Package fbc is compiled to an executable. It allows various commands on files filtered by their content.
package main

import (
	"fmt"
	"github.com/vbsw/checkfile"
	"os"
	"path/filepath"
	"unsafe"
)

func main() {
	args := new(arguments)
	err := args.parseCommandLine(os.Args[1:])

	if err == nil {
		if args.infoAvailable() {
			args.printInfo()

		} else {
			fileProc := newFileProcessor(args.command.Values[0])
			iterate(args, fileProc)
		}
	} else {
		fmt.Println(messageError(err))
	}
}

// iterate iterates over files calling fileProc.processFile for each file
// matching the criteria. If fileProc.processFile returns error != nil,
// then processing is stopped.
func iterate(args *arguments, fileProc fileProcessor) error {
	var err error

	if fileProc == nil {
		fileProc = new(fileProcessorDefault)
	}
	if args.recursive.Available() {
		err = iterateRecursive(args, fileProc)

	} else {
		err = iterateFlat(args, fileProc)
	}
	return err
}

func iterateRecursive(args *arguments, fileProc fileProcessor) error {
	inputDir := args.input.Values[0]
	byOr := args.or.Available()
	silent := args.silent.Available()
	filterParts := splitStringByStar(args.inputFilter)
	buffer := checkfile.NewTermsBuffer(1024*1024*4, args.contentFilter)
	count := 0
	err := filepath.Walk(inputDir, func(path string, fileInfo os.FileInfo, err error) error {
		if err == nil {
			// avoid input directory as input file
			if len(path) > len(inputDir) {
				var match bool
				match, err = isFileMatch(byOr, path, fileInfo, filterParts, buffer)

				if match && err == nil {
					err = fileProc.processFile(args, path, fileInfo)

					if err == nil {
						count++
					}
				}
			}
		}
		// ignore errors
		if err != nil {
			if !silent {
				fmt.Println(messageWarning(err))
			}
			err = nil
		}
		return err
	})
	fileProc.summary(count, err)

	return err
}

func iterateFlat(args *arguments, fileProc fileProcessor) error {
	inputDir := args.input.Values[0]
	byOr := args.or.Available()
	silent := args.silent.Available()
	filterParts := splitStringByStar(args.inputFilter)
	buffer := checkfile.NewTermsBuffer(1024*1024*4, args.contentFilter)
	count := 0
	err := filepath.Walk(inputDir, func(path string, fileInfo os.FileInfo, err error) error {
		if err == nil {
			// avoid input directory as input file; parent must be input directory
			if len(path) > len(inputDir) && len(filepath.Dir(path)) == len(inputDir) {
				var match bool
				match, err = isFileMatch(byOr, path, fileInfo, filterParts, buffer)

				if match && err == nil {
					err = fileProc.processFile(args, path, fileInfo)

					if err == nil {
						count++
					}
				}
			}
		}
		// ignore errors
		if err != nil {
			if !silent {
				fmt.Println(messageWarning(err))
			}
			err = nil
		}
		return err
	})
	fileProc.summary(count, err)

	return err
}

func isFileMatch(byOr bool, path string, fileInfo os.FileInfo, filterParts [][]byte, buffer *checkfile.TermsBuffer) (bool, error) {
	var match bool
	var err error

	if !fileInfo.IsDir() && isNameMatch(fileInfo.Name(), filterParts) {
		if byOr {
			match, err = checkfile.ContainsAny(path, buffer)
		} else {
			match, err = checkfile.ContainsAll(path, buffer)
		}
	}
	return match, err
}

func isNameMatch(fileName string, filterParts [][]byte) bool {
	// avoid copying
	name := *(*[]byte)(unsafe.Pointer(&fileName))
	if len(filterParts) > 0 {
		if hasPrefix(name, filterParts[0]) {
			offset := len(filterParts[0])

			for _, filterPart := range filterParts[1:] {
				offsetNew := matchEndIndex(name, filterPart, offset)

				if offset != offsetNew {
					offset = offsetNew
				} else {
					return false
				}
			}
			if offset == len(name) {
				return true
			}
			return false
		}
		return false
	}
	return true
}

func hasPrefix(bytes, prefix []byte) bool {
	if len(bytes) >= len(prefix) {
		for i, b := range prefix {
			if bytes[i] != b {
				return false
			}
		}
		return true
	}
	return false
}

func matchEndIndex(bytes, part []byte, offset int) int {
	if len(part) > 0 {
		for i := offset; i < len(bytes)-len(part)+1; i++ {
			match := true

			for j, b := range part {
				if bytes[i+j] != b {
					match = false
					break
				}
			}
			if match {
				offset = i + len(part)
				break
			}
		}
		// only last part is empty
	} else {
		offset = len(bytes)
	}
	return offset
}
