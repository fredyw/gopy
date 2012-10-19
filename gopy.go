// Copyright 2012 Fredy Wijaya
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package main

import (
    "bufio"
    "flag"
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
)

type fileInfo struct {
    file string
    size int64
}

func getSize(dir string) int64 {
    size := int64(0)
    filepath.Walk(dir,
        func(path string, info os.FileInfo, err error) error {
            size += info.Size()
            return nil
        })
    return size
}

func listFiles(dir string, noFile, noDir bool) ([]fileInfo, error) {
    result := []fileInfo{}
    if fi, e := ioutil.ReadDir(dir); e != nil {
        return result, e
    } else {
        for _, info := range fi {
            if (info.IsDir() && !noDir) || (!info.IsDir() && !noFile) {
                filePath, _ := filepath.Abs(filepath.Join(dir, info.Name()))
                size := getSize(filePath)
                result = append(result, fileInfo{filePath, size})
            }
        }
    }
    return result, nil
}

func listFilesRecursively(dir string, noFile, noDir bool) ([]fileInfo, error) {
    result := []fileInfo{}
    e := filepath.Walk(dir,
        func(path string, info os.FileInfo, err error) error {
            if (info.IsDir() && !noDir) || (!info.IsDir() && !noFile) {
                filePath, _ := filepath.Abs(path)
                size := getSize(filePath)
                result = append(result, fileInfo{filePath, size})
            }
            return nil
        })
    if e != nil {
        return result, e
    }
    return result, nil
}

func printUsage() {
    fmt.Println("Usage:", os.Args[0])
    flag.PrintDefaults()
}

func printUsageAndExit(exitCode int) {
    printUsage()
    os.Exit(exitCode)
}

func printError(msg interface{}) {
    fmt.Println("Error:", msg)
}

func printErrorAndExit(msg interface{}, exitCode int) {
    printError(msg)
    os.Exit(exitCode)
}

func isDirectory(path string) bool {
    f, e := os.Open(path)
    if e != nil {
        return false
    }
    if fi, e := f.Stat(); e != nil {
        return false
    } else {
        if fi.IsDir() {
            return true
        }
    }
    return false
}

func fileExists(path string) bool {
    f, e := os.Open(path)
    if f == nil && e != nil {
        return false
    }
    return true
}

var copyFlag *bool
var inputFile *string
var listFlag *bool
var directoryPath *string
var outputFile *string
var noDirFlag *bool
var noFileFlag *bool
var recursiveFlag *bool

func init() {
    copyFlag = flag.Bool("copy", false, "copy operation")
    inputFile = flag.String("input", "", "input file (for copy) - mandatory")
    listFlag = flag.Bool("list", false, "list operation")
    directoryPath = flag.String("directory", "", "directory (for list & copy) - mandatory")
    outputFile = flag.String("output", "", "output file (for list) - mandatory")
    noDirFlag = flag.Bool("nodir", false, "don't include directories (for list) - optional")
    noFileFlag = flag.Bool("nofile", false, "don't include files (for list) - optional")
    recursiveFlag = flag.Bool("recursive", false, "recursive (for list) - optional")
    helpFlag := flag.Bool("help", false, "help")

    flag.Parse()

    if *helpFlag {
        printUsageAndExit(0)
    }

    if *copyFlag && *listFlag {
        printUsageAndExit(1)
    }

    if !*copyFlag && !*listFlag {
        printUsageAndExit(1)
    }

    if *copyFlag {
        if *inputFile == "" || *directoryPath == "" {
            printUsageAndExit(1)
        }
        if !fileExists(*inputFile) {
            printErrorAndExit(*inputFile + " does not exist", 1)
        }
    } else if *listFlag {
        if *outputFile == "" || *directoryPath == "" {
            printUsageAndExit(1)
        }
        if !isDirectory(*directoryPath) {
            printErrorAndExit(*directoryPath + " does not exist or is not a directory", 1)
        }
    }
}

func List(directoryPath, outputFile string, noFileFlag, noDirFlag, recursiveFlag bool) {
    var info []fileInfo
    var e error
    if recursiveFlag {
        info, e = listFilesRecursively(directoryPath, noFileFlag, noDirFlag)
    } else {
        info, e = listFiles(directoryPath, noFileFlag, noDirFlag)
    }
    if e != nil {
        printErrorAndExit(e, 1)
    }
    f, e := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
    if e != nil {
        printErrorAndExit(e, 1)
    }
    for _, i := range info {
        // TODO: make a more human-readable size, e.g. KB, MB, GB, TB, and not just MB
        fmt.Fprintf(f, "%s - %.2fMB\n", i.file, float64(i.size) / float64(1024000))
    }
}

func copyFile(src, dest string) error {
    srcFile, e := os.Open(src)
    if e != nil {
        return e
    }
    defer srcFile.Close()

    destFile, e := os.Create(dest)
    if e != nil {
        return e
    }
    defer destFile.Close()

    io.Copy(destFile, srcFile)
    return nil
}

func readTextFile(inputFile string) []string {
    result := []string{}
    f, _ := os.Open(inputFile)
    defer f.Close()
    r := bufio.NewReader(f)
    line, e := r.ReadString('\n')
    for e == nil {
        trimmedLine := strings.TrimSpace(line)
        endIdx := strings.LastIndex(trimmedLine, "-") - 1
        result = append(result, trimmedLine[0:endIdx])
        line, e = r.ReadString('\n')
    }
    return result
}

func Copy(directoryPath, inputPath string) {
    os.MkdirAll(directoryPath, 0755)
    for _, dir := range readTextFile(inputPath) {
        baseDir := filepath.Base(dir)
        filepath.Walk(dir,
            func(path string, info os.FileInfo, err error) error {
                dest := filepath.Join(directoryPath, path[strings.Index(path, baseDir):])
                if info.IsDir() {
                    os.MkdirAll(dest, 0755)
                } else {
                    copyFile(path, dest)
                }
                return nil
        })
    }
}

func main() {
    if *listFlag {
        List(*directoryPath, *outputFile, *noFileFlag, *noDirFlag, *recursiveFlag)
    } else if *copyFlag {
        Copy(*directoryPath, *inputFile)
    }
}

