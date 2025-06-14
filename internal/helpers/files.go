/*
Copyright © 2025 Nissar Chababy

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package helpers

import (
	"bufio"
	"io"
	"log"
	"os"
)

// IterFile reads a file line by line and applies the provided yield function to each line - thus allowing for processing of each line.
//
// Args:
//
//	filePath: The path to the file to be read.
//	yield: A function that takes a string (the line read from the file) and processes it.
//
// Returns:
//
//	None. If an error occurs while opening or reading the file, it logs the error and exits the program.
func IterFile(filePath string, yield func(string)) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Panic("error closing file:", err)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		yield(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

// WriteFileFromIter creates a file at the specified path and writes lines to it using the provided iterator function.
//
// Args:
//
//	filePath: The path where the file will be created.
//	iter: A function that takes a function as an argument, which will be called with each line to write to the file.
//
// Returns:
//
//	None. If an error occurs while creating the file, it logs the error and exits the program.
func WriteFileFromIter(filePath string, iter func(func(string))) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Panicf("error closing file: %v", err)
		}
	}()

	iter(func(line string) {
		if _, err := file.WriteString(line + "\n"); err != nil {
			log.Panic("failed to write line to file:", err)
		}
	})
}

// CopyFile copies the contents of a source file to a destination file.
// Args:
//
//	srcFile: The path to the source file.
//	destFile: The path to the destination file.
//
// Returns:
//
//	An error if the copy operation fails, otherwise nil.
func CopyFile(srcFile string, destFile string) error {
	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer func() {
		if err := src.Close(); err != nil {
			log.Panicf("error closing source file: %v", err)
		}
	}()

	dest, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer func() {
		if err := dest.Close(); err != nil {
			log.Panicf("error closing destination file: %v", err)
		}
	}()

	if _, err := io.Copy(dest, src); err != nil {
		return err
	}

	return nil
}
