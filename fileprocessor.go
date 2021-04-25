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
	processFile(*arguments, string, os.FileInfo) error

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
	existingDirs []string
}

type fileProcessorMV struct {
	fileProcessorCP
}

type fileProcessorPrint struct {
}

type fileProcessorRM struct {
	fileProcessorDefault
}

func newFileProcessor(command string) fileProcessor {
	var processor fileProcessor
	switch command {
	case argCOUNT:
		processor = new(fileProcessorCount)
	case argCP:
		processorCP := new(fileProcessorCP)
		processorCP.existingDirs = make([]string, 0, 16)
		processor = processorCP
	case argMV:
		processorMV := new(fileProcessorMV)
		processorMV.existingDirs = make([]string, 0, 16)
		processor = processorMV
	case argPRINT:
		processor = new(fileProcessorPrint)
	case argRM:
		processor = new(fileProcessorRM)
	}
	return processor
}

func (fileProc *fileProcessorDefault) processFile(args *arguments, path string, fileInfo os.FileInfo) error {
	fmt.Println(path)
	return nil
}

func (fileProc *fileProcessorDefault) summary(count int, err error) {
	if err == nil {
		fmt.Println(messageFinished(count))
	} else {
		fmt.Println(messageError(err))
	}
}

func (fileProc *fileProcessorCount) processFile(args *arguments, path string, fileInfo os.FileInfo) error {
	return nil
}

func (fileProc *fileProcessorCount) summary(count int, err error) {
	if err == nil {
		fmt.Println(count)
	} else {
		fmt.Println(messageError(err))
	}
}

func (fileProc *fileProcessorCP) processFile(args *arguments, path string, fileInfo os.FileInfo) error {
	var err error
	var inputFile *os.File
	inputFile, err = os.Open(path)

	if err == nil {
		var outputFile *os.File
		defer inputFile.Close()
		inputDir := args.input.Values()[0]
		outputDir := args.output.Values()[0]
		subDir := path[len(inputDir) : len(path)-len(fileInfo.Name())]
		outputPath := filepath.Join(outputDir, subDir)
		err = fileProc.ensureDir(outputPath)

		if err == nil {
			outputPath = filepath.Join(outputPath, fileInfo.Name())

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

func (fileProc *fileProcessorCP) ensureDir(dir string) error {
	var err error
	exists := false

	for _, existingDir := range fileProc.existingDirs {
		if existingDir == dir {
			exists = true
			break
		}
	}
	if !exists {
		fileProc.existingDirs = append(fileProc.existingDirs, dir)

		if !checkfile.Exists(dir) {
			err = os.MkdirAll(dir, 0666)
		}
	}
	return err
}

func (fileProc *fileProcessorMV) processFile(args *arguments, path string, fileInfo os.FileInfo) error {
	var err error
	inputDir := args.input.Values()[0]
	outputDir := args.output.Values()[0]
	subDir := path[len(inputDir) : len(path)-len(fileInfo.Name())]
	outputPath := filepath.Join(outputDir, subDir)
	err = fileProc.ensureDir(outputPath)

	if err == nil {
		outputPath = filepath.Join(outputPath, fileInfo.Name())

		if !checkfile.Exists(outputPath) {
			err = os.Rename(path, outputPath)
		} else {
			err = errors.New("target file or directory already exists: " + fileInfo.Name())
		}
	}
	return err
}

func (fileProc *fileProcessorPrint) processFile(args *arguments, path string, fileInfo os.FileInfo) error {
	fmt.Println(fileInfo.Name())
	return nil
}

func (fileProc *fileProcessorPrint) summary(count int, err error) {
	if err != nil {
		fmt.Println(messageError(err))
	}
}

func (fileProc *fileProcessorRM) processFile(args *arguments, path string, fileInfo os.FileInfo) error {
	err := os.Remove(path)
	return err
}
