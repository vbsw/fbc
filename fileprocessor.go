/*
 *          Copyright 2020, Vitali Baumtrok.
 * Distributed under the Boost Software License, Version 1.0.
 *     (See accompanying file LICENSE or copy at
 *        http://www.boost.org/LICENSE_1_0.txt)
 */

package main

import (
	"errors"
	"fmt"
	"github.com/vbsw/checkfile"
	"io"
	"os"
	"path/filepath"
)

// fileProcessor holds callbacks for iteration over files.
type fileProcessor interface {
	// processFile is called for every file matching the criteria.
	// If returned value != nil, iteration is stopped.
	processFile(*command, os.FileInfo) error

	// summary is called after all files have been processed or
	// an error occurred. count is the number of files processed.
	summary(count int, err error)
}

type fileProcessorDefault struct {
}

type fileProcessorCount struct {
}

type fileProcessorCP struct {
	fileProcessorDefault
}

type fileProcessorMV struct {
	fileProcessorDefault
}

type fileProcessorPrint struct {
}

type fileProcessorRM struct {
	fileProcessorDefault
}

func newFileProcessor(commandId int) fileProcessor {
	var processor fileProcessor
	switch commandId {
	case cmdCOUNT:
		processor = new(fileProcessorCount)
	case cmdCP:
		processor = new(fileProcessorCP)
	case cmdMV:
		processor = new(fileProcessorMV)
	case cmdPRINT:
		processor = new(fileProcessorPrint)
	case cmdRM:
		processor = new(fileProcessorRM)
	}
	return processor
}

func (fileProc *fileProcessorDefault) processFile(cmd *command, fileInfo os.FileInfo) error {
	fmt.Println(cmd.inputDir + fileInfo.Name())
	return nil
}

func (fileProc *fileProcessorDefault) summary(count int, err error) {
	if err == nil {
		fmt.Println(messageFinished(count))
	} else {
		fmt.Println(messageError(err))
	}
}

func (fileProc *fileProcessorCount) processFile(cmd *command, fileInfo os.FileInfo) error {
	return nil
}

func (fileProc *fileProcessorCount) summary(count int, err error) {
	fmt.Println(count)
}

func (fileProc *fileProcessorCP) processFile(cmd *command, fileInfo os.FileInfo) error {
	var err error
	inputPath := filepath.Join(cmd.inputDir, fileInfo.Name())

	if err == nil {
		var inputFile *os.File
		inputFile, err = os.Open(inputPath)

		if err == nil {
			var outputFile *os.File
			defer inputFile.Close()
			outputPath := filepath.Join(cmd.outputDir, fileInfo.Name())

			if !checkfile.Exists(outputPath) {
				outputFile, err = os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)

				if err == nil {
					defer outputFile.Close()
					_, err = io.Copy(outputFile, inputFile)
				}
			} else {
				err = errors.New("target file or directory already exists: " + fileInfo.Name())
			}
		}
	}
	return err
}

func (fileProc *fileProcessorMV) processFile(cmd *command, fileInfo os.FileInfo) error {
	var err error
	inputPath := filepath.Join(cmd.inputDir, fileInfo.Name())
	outputPath := filepath.Join(cmd.outputDir, fileInfo.Name())

	if !checkfile.Exists(outputPath) {
		err = os.Rename(inputPath, outputPath)
	} else {
		err = errors.New("target file or directory already exists: " + fileInfo.Name())
	}
	return err
}

func (fileProc *fileProcessorPrint) processFile(cmd *command, fileInfo os.FileInfo) error {
	fmt.Println(fileInfo.Name())
	return nil
}

func (fileProc *fileProcessorPrint) summary(count int, err error) {
}

func (fileProc *fileProcessorRM) processFile(cmd *command, fileInfo os.FileInfo) error {
	inputPath := filepath.Join(cmd.inputDir, fileInfo.Name())
	err := os.Remove(inputPath)
	return err
}
