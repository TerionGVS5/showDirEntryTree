package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	prefixLastEntry    = "└───"
	prefixHasNextEntry = "├───"
)

type DirEntry struct {
	FileName string
	Path     string
	IsDir    bool
	Prefix   string
	FileSize string
}

type Stack []DirEntry

func (stack *Stack) push(dirEntry DirEntry) {
	*stack = append(*stack, dirEntry)
}

func (stack *Stack) pop() DirEntry {
	n := len(*stack)
	if n == 0 {
		panic("Stack empty!")
	}
	dirEntry := (*stack)[n-1]
	*stack = (*stack)[:n-1]
	return dirEntry
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := showDirEntryTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func showDirEntryTree(out io.Writer, path string, withFiles bool) error {
	var stack Stack
	var pathChanged bool
	var closestPrefix, PrevPrefix, currPrefix, prettySize, prettyFileName string
	var countFilesSkipped int

	currPath := path

	for {
		pathChanged = false
		dirEntries, err := os.ReadDir(currPath)
		if err != nil {
			return err
		}
		countFilesSkipped = 0
		for i := len(dirEntries) - 1; i >= 0; i-- {
			isDir := dirEntries[i].IsDir()
			dirName := dirEntries[i].Name()

			if !isDir && !withFiles {
				countFilesSkipped += 1
				continue
			}

			if i == len(dirEntries)-1-countFilesSkipped {
				closestPrefix = prefixLastEntry
			} else {
				closestPrefix = prefixHasNextEntry
			}
			currPrefix = strings.ReplaceAll(PrevPrefix, prefixLastEntry, "\t")
			currPrefix = strings.ReplaceAll(currPrefix, prefixHasNextEntry, "│\t")
			currPrefix = currPrefix + closestPrefix
			if !isDir {
				fi, err := os.Stat(filepath.Join(currPath, dirName))
				if err != nil {
					return err
				}
				size := fi.Size()
				if size == 0 {
					prettySize = "(empty)"
				} else {
					prettySize = fmt.Sprintf("(%db)", size)
				}
			} else {
				prettySize = ""
			}
			stack.push(DirEntry{
				FileName: dirName,
				Path:     currPath,
				IsDir:    isDir,
				Prefix:   currPrefix,
				FileSize: prettySize,
			})
		}

		for {
			if len(stack) == 0 {
				break
			}
			dirEntry := stack.pop()

			if !dirEntry.IsDir {
				prettyFileName = dirEntry.FileName + " " + dirEntry.FileSize
			} else {
				prettyFileName = dirEntry.FileName
			}

			_, err = fmt.Fprintf(out, "%s%s\n", dirEntry.Prefix, prettyFileName)
			if err != nil {
				return err
			}

			if dirEntry.IsDir {
				pathChanged = true
				currPath = filepath.Join(dirEntry.Path, dirEntry.FileName)
				PrevPrefix = dirEntry.Prefix
				break
			}
		}

		if !pathChanged && len(stack) == 0 {
			return nil
		}

	}
}
