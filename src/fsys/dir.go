package fsys

import (
	"fmt"
	"os"
	"path/filepath"
)

// A struct that represents the main information about a directory
type Directory struct {
	Name        string
	Path        string
	ParentPath  string
	Files       []*File
	Directories []*Directory
}

var ErrorNotDirectory error = fmt.Errorf("not a directory")

func GetDir(path string, recursive bool) (*Directory, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	stats, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	// check if it is a directory for real
	if !stats.IsDir() {
		return nil, ErrorNotDirectory
	}

	// loop through each entry in the directory
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, err
	}

	var innerDirs []*Directory
	var innerFiles []*File
	for _, entry := range entries {
		entryInfo, err := entry.Info()
		if err != nil {
			return nil, err
		}

		if entryInfo.IsDir() {
			if recursive {
				// do the recursive magic
				innerDirPath := filepath.Join(absPath, entry.Name())

				innerDir, err := GetDir(innerDirPath, true)
				if err != nil {
					return nil, err
				}

				innerDirs = append(innerDirs, innerDir)
			}
			// if not - skip the directory

		} else {
			innerFilePath := filepath.Join(absPath, entryInfo.Name())

			innerFile, err := GetFile(innerFilePath)
			if err != nil {
				return nil, err
			}

			innerFiles = append(innerFiles, innerFile)
		}
	}

	directory := Directory{
		Name:        stats.Name(),
		Path:        absPath,
		ParentPath:  filepath.Dir(absPath),
		Directories: innerDirs,
		Files:       innerFiles,
	}

	return &directory, nil
}
