package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func calculateFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("Error while opening file %q: %v", path, err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("Error while copying file %q: %v", path, err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func findDuplicatesFilesInfolder(path string) (map[string][]string, error) {
	duplicates:= make(map[string][]string)
	hashes := make(map[string]string)
	
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("Error while reading directory %q: %v", path, err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(path, file.Name())
		hash, err := calculateFileHash(filePath)

		if err != nil {
			return nil, fmt.Errorf("Error while calculating hash for file %q: %v", filePath, err)
		}

		if _, ok := hashes[hash]; ok {
			fmt.Printf("Found duplicate file: %q\n", filePath)
			duplicates[hash] = append(duplicates[hash], filePath)
		} else {
			fmt.Printf("New file: %q\n", filePath)
			hashes[hash] = filePath
		}
	}

	return duplicates, nil
}

func removeDplicateFiles(path string) error {
	duplicates, err := findDuplicatesFilesInfolder(path)

	if err != nil {
		return fmt.Errorf("Error while finding duplicates in %q: %v\n", path, err)
	}

	for hash, files := range duplicates {
		fmt.Printf("Hash: %q\n", hash)
		for _, file := range files {
			err := os.Remove(file)
			if err != nil {
				return fmt.Errorf("Error while removing file %q: %v\n", file, err)
			}
			
			fmt.Printf("\t%q\n", file)
		}
	}

	return nil
}


func main() {
	// open download file
	downloadsPath := filepath.Join(os.Getenv("HOME"), "Downloads")

	
	folders := map[string][]string{
		"Books": {".epub", ".mobi"},
		"Images": {".jpg", ".jpeg", ".png", ".gif", ".svg"},
		"Videos": {".mov", ".avi", ".mp4"},
		"Documents": {".doc", ".docx", ".txt", ".pdf", ".xlsx", ".xls", ".ppt", ".pptx", ".csv"},
		"Applications": {".exe", ".pkg", ".deb", ".dmg", ".apk"},
		"Compressed": {".zip", ".tar", ".bz2", ".rar"},
		"Music": {".mp3", ".wav", ".ogg", ".midi"},
		"Code": {".go", ".py", ".ts",".tsx", ".html", ".css", ".java", ".cpp", ".h", ".c", ".rb", ".php", ".json", ".sql"},
	}

	for folder := range folders {
		dir := filepath.Join(downloadsPath, folder)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.Mkdir(dir, 0755)
		}
	}

	categorizeAndMoveFile := func (path string, info os.FileInfo)  {
			if info.IsDir() {
				return
			}

			ext := strings.ToLower(filepath.Ext(path))

			moved := false

			for folder, exts := range folders {
				for _, e := range exts {
					if e == ext {
						// new path
						target := filepath.Join(downloadsPath, folder, info.Name())
						os.Rename(path, target)
						moved = true
						break
					}
				}
				if moved{
					break
				}
			}

			if !moved {
				target := filepath.Join(downloadsPath, "Others", info.Name())
				os.Rename(path, target)
			}
	}

	err := filepath.Walk(downloadsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		categorizeAndMoveFile(path, info)



		return nil
	})

	if err != nil {
		fmt.Printf("Error while walking the path %q: %v\n", downloadsPath, err)
	}

	// check for duplicates
	for folder := range folders {
		dir := filepath.Join(downloadsPath, folder)
		err := removeDplicateFiles(dir)
		if err != nil {
			fmt.Printf("Error while removing duplicates in %q: %v\n", dir, err)
		}
	}	
		 
}
