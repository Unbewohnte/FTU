package fsys

import (
	"fmt"
	"os"
	"path/filepath"
)

// A struct that represents the main information about a directory
type Directory struct {
	Name               string
	Path               string
	Size               uint64
	RelativeParentPath string // Relative path to the directory, where the highest point in the hierarchy is the upmost parent dir. Set manually
	Files              []*File
	Directories        []*Directory
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

	directory := Directory{
		Name:        stats.Name(),
		Path:        absPath,
		Directories: nil,
		Files:       nil,
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

				directory.Size += innerDir.Size

				innerDirs = append(innerDirs, innerDir)
			}
			// if not - skip the directory and only work with the files

		} else {
			innerFilePath := filepath.Join(absPath, entryInfo.Name())

			innerFile, err := GetFile(innerFilePath)
			if err != nil {
				return nil, err
			}

			directory.Size += innerFile.Size

			innerFiles = append(innerFiles, innerFile)
		}
	}

	directory.Directories = innerDirs
	directory.Files = innerFiles

	return &directory, nil
}

// Returns every file in that directory
func (dir *Directory) GetAllFiles(recursive bool) []*File {
	var files []*File = dir.Files

	if recursive {
		if len(dir.Directories) == 0 {
			return files
		}

		for _, innerDir := range dir.Directories {
			innerFiles := innerDir.GetAllFiles(recursive)
			files = append(files, innerFiles...)
		}

	} else {
		files = dir.Files
	}

	return files
}

// Sets `RelativeParentPath` relative to the given base path.
// file with such path:
// /home/user/directory/somefile.txt
// had a relative path like that:
// /directory/somefile.txt
// (where base path is /home/user/directory)
func (dir *Directory) SetRelativePaths(base string, recursive bool) error {
	for _, file := range dir.GetAllFiles(recursive) {
		relPath, err := filepath.Rel(base, file.Path)
		if err != nil {
			return err
		}

		file.RelativeParentPath = relPath

	}
	return nil
}
