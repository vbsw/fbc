/*
 *      Copyright 2021, 2022 Vitali Baumtrok.
 * Distributed under the Boost Software License, Version 1.0.
 *     (See accompanying file LICENSE or copy at
 *        http://www.boost.org/LICENSE_1_0.txt)
 */

package main

import (
	"errors"
	"fmt"
	"github.com/vbsw/golib/check"
	"io"
	"os"
	"path/filepath"
)

// fileProcessor holds callbacks for iteration over files.
type fileProcessor interface {
	// processFile is called for every file matching the criteria.
	// If returned value != nil, iteration is stopped.
	processFile(*parameters, string, os.FileInfo) error

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

func (fileProc *fileProcessorDefault) processFile(params *parameters, path string, info os.FileInfo) error {
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

func (fileProc *fileProcessorCount) processFile(params *parameters, path string, info os.FileInfo) error {
	return nil
}

func (fileProc *fileProcessorCount) summary(count int, err error) {
	if err == nil {
		fmt.Println(count)
	} else {
		fmt.Println(messageError(err))
	}
}

func (fileProc *fileProcessorCP) processFile(params *parameters, path string, info os.FileInfo) error {
	var err error
	var inputFile *os.File
	inputFile, err = os.Open(path)
	if err == nil {
		var outputFile *os.File
		defer inputFile.Close()
		inputDir := params.input.Values[0]
		outputDir := params.output.Values[0]
		subDir := path[len(inputDir) : len(path)-len(info.Name())]
		outputPath := filepath.Join(outputDir, subDir)
		err = fileProc.ensureDir(outputPath)
		if err == nil {
			outputPath = filepath.Join(outputPath, info.Name())
			if !check.FileExists(outputPath) {
				outputFile, err = os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
				if err == nil {
					defer outputFile.Close()
					_, err = io.Copy(outputFile, inputFile)
				}
			} else {
				err = errors.New("target file already exists: " + info.Name())
			}
		}
	}
	return err
}

func (fileProc *fileProcessorCP) ensureDir(dir string) error {
	for _, existingDir := range fileProc.existingDirs {
		if existingDir == dir {
			return nil
		}
	}
	info, err := os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0666)
		if err == nil {
			fileProc.existingDirs = append(fileProc.existingDirs, dir)
		}
	} else if info != nil && err == nil {
		if info.IsDir() {
			fileProc.existingDirs = append(fileProc.existingDirs, dir)
		} else {
			err = errors.New("can't create directory (already exists as file): " + info.Name())
		}
	}
	return err
}

func (fileProc *fileProcessorMV) processFile(params *parameters, path string, info os.FileInfo) error {
	var err error
	inputDir := params.input.Values[0]
	outputDir := params.output.Values[0]
	subDir := path[len(inputDir) : len(path)-len(info.Name())]
	outputPath := filepath.Join(outputDir, subDir)
	err = fileProc.ensureDir(outputPath)
	if err == nil {
		outputPath = filepath.Join(outputPath, info.Name())
		if !check.FileExists(outputPath) {
			err = os.Rename(path, outputPath)
		} else {
			err = errors.New("target file already exists: " + info.Name())
		}
	}
	return err
}

func (fileProc *fileProcessorPrint) processFile(params *parameters, path string, info os.FileInfo) error {
	fmt.Println(info.Name())
	return nil
}

func (fileProc *fileProcessorPrint) summary(count int, err error) {
	if err != nil {
		fmt.Println(messageError(err))
	}
}

func (fileProc *fileProcessorRM) processFile(params *parameters, path string, info os.FileInfo) error {
	err := os.Remove(path)
	return err
}
