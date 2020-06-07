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
)

func main() {
	cmd, err := commandFromOSArgs()

	if err == nil {
		if cmd.info {
			fmt.Println(cmd.infoMessage)

		} else {
			fileProc := newFileProcessor(cmd.commandId)
			iterate(cmd, fileProc)
		}
	} else {
		fmt.Println(messageError(err))
	}
}

// iterate iterates over files calling fileProc.processFile for each file
// matching the criteria. If fileProc.processFile returns error != nil,
// then processing is stopped.
func iterate(cmd *command, fileProc fileProcessor) error {
	var err error

	if fileProc == nil {
		fileProc = new(fileProcessorDefault)
	}
	if cmd.recursive {
		err = iterateRecursive(cmd, fileProc)

	} else {
		err = iterateFlat(cmd, fileProc)
	}
	return err
}

func iterateRecursive(cmd *command, fileProc fileProcessor) error {
	filterParts := splitStringByStar(cmd.inputFilter)
	inputDirWOPrefix := removePathPrefix(cmd.inputDir)
	buffer := checkfile.NewTermsBuffer(1024*1024*4, cmd.contentFilter)
	count := 0

	err := filepath.Walk(cmd.inputDir, func(path string, fileInfo os.FileInfo, err error) error {
		if err == nil {
			// don't process subdirectories
			if len(inputDirWOPrefix) == len(path)-len(fileInfo.Name()) {
				var match bool
				match, err = isFileMatch(cmd, fileInfo, filterParts, buffer)

				if match && err == nil {
					err = fileProc.processFile(cmd, fileInfo)

					if err == nil {
						count++
					}
				}
			}
		}
		return err
	})
	fileProc.summary(count, err)

	return err
}

func iterateFlat(cmd *command, fileProc fileProcessor) error {
	dir, err := os.Open(cmd.inputDir)
	count := 0

	if err == nil {
		var fileInfos []os.FileInfo
		fileInfos, err = dir.Readdir(0)
		dir.Close()

		if err == nil {
			filterParts := splitStringByStar(cmd.inputFilter)
			buffer := checkfile.NewTermsBuffer(1024*1024*4, cmd.contentFilter)

			for _, fileInfo := range fileInfos {
				var match bool
				match, err = isFileMatch(cmd, fileInfo, filterParts, buffer)

				if err == nil && match {
					err = fileProc.processFile(cmd, fileInfo)

					if err == nil {
						count++
					}
				}
				if err != nil {
					break
				}
			}
		}
	}
	fileProc.summary(count, err)

	return err
}

func isFileMatch(cmd *command, fileInfo os.FileInfo, filterParts [][]byte, buffer *checkfile.TermsBuffer) (bool, error) {
	var match bool
	var err error

	if !fileInfo.IsDir() && isNameMatch([]byte(fileInfo.Name()), filterParts) {
		path := filepath.Join(cmd.inputDir, fileInfo.Name())

		if cmd.or {
			match, err = checkfile.ContainsAny(path, buffer)
		} else {
			match, err = checkfile.ContainsAll(path, buffer)
		}
	}
	return match, err
}

func isNameMatch(fileName []byte, filterParts [][]byte) bool {
	if len(filterParts) > 0 {
		if hasPrefix(fileName, filterParts[0]) {
			offset := len(filterParts[0])

			for _, filterPart := range filterParts[1:] {
				offsetNew := matchEndIndex(fileName, filterPart, offset)

				if offset != offsetNew {
					offset = offsetNew
				} else {
					return false
				}
			}
			if offset == len(fileName) {
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

func removePathPrefix(path string) string {
	bytes := []byte(path)

	if len(bytes) > 0 && bytes[0] == '.' {
		if len(bytes) > 1 {
			if bytes[1] == '/' || bytes[1] == '\\' {
				path = path[2:]
			}
		} else {
			path = ""
		}
	}
	return path
}
