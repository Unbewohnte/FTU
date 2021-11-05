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
	Size        uint64
	Files       []*File
	Directories []*Directory
}

var ErrorNotDirectory error = fmt.Errorf("not a directory")

// // gets a child directory
// func getDirChild(path string, parentDir *Directory, recursive bool) (*Directory, error) {

// }

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

	directory := Directory{
		Name:        stats.Name(),
		Path:        absPath,
		ParentPath:  filepath.Dir(absPath),
		Directories: nil,
		Files:       nil,
	}

	// loop through each entry in the directory
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, err
	}

	// var totalSize uint64 = 0
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

				directory.Size += innerDir.Size
			}
			// if not - skip the directory and only work with the files

		} else {
			innerFilePath := filepath.Join(absPath, entryInfo.Name())

			innerFile, err := GetFile(innerFilePath)
			if err != nil {
				return nil, err
			}

			innerFiles = append(innerFiles, innerFile)

			directory.Size += innerFile.Size
		}
	}

	directory.Directories = innerDirs
	directory.Files = innerFiles
	// directory.Size

	return &directory, nil
}
