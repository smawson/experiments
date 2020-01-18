package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const dryRun = false

func main() {
	// Take a directory path, read the files, print them out.
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Expected a single argument, got %v\n", args)
		os.Exit(1)
	}
	path := args[0]
	log.Printf("Reading files at path %s\n", path)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading path: %v\n", err)
		os.Exit(1)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".MP4") {
			err = renameFile(path, file)
			if err != nil {
				log.Printf("Rename file %s failed: %v", file.Name(), err)
			}
		}
	}
}

func renameFile(path string, file os.FileInfo) error {
	fullpath := filepath.Join(path, file.Name())
	cmd := exec.Command("mediainfo", fullpath)
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	date, err := getDate(out)
	if err != nil {
		return err
	}

	//	newname := fmt.Sprintf("GP_%s%s%s_%s_%s_%s.MP4",
	//		date[0:4], date[5:6], date[7:8], date[10:11], date[12:13], date[14:15])
	newname := fmt.Sprintf("GP_%s%s%s_%s_%s_%s.MP4",
		date[0:4], date[5:7], date[8:10], date[11:13], date[14:16], date[17:19])
	if newname == file.Name() {
		fmt.Printf("File %s already has correct name, skipping.\n", newname)
		return nil
	}
	newpath := filepath.Join(path, newname)
	if dryRun {
		fmt.Printf("[DRY RUN] Renaming from %s to %s\n", fullpath, newpath)
	} else {
		fmt.Printf("Renaming from %s to %s\n", fullpath, newpath)
		_, err := os.Stat(newpath)
		if os.IsNotExist(err) {
			os.Rename(fullpath, newpath)
		} else if err != nil {
			return err
		} else {
			return fmt.Errorf("file %s already exists, not overwriting", newpath)
		}
	}
	return nil
}

func getDate(out []byte) (string, error) {
	lines := bytes.Split(out, []byte("\n"))
	for _, line := range lines {
		if len(line) > 0 {
			if bytes.HasPrefix(line, []byte("Encoded date")) {
				last := bytes.LastIndex(line, []byte("UTC"))
				if last == -1 {
					return "", fmt.Errorf("Invalid Encoded date: %s", line)
				}
				return string(line[last+4:]), nil
			}
		}
	}

	return "", fmt.Errorf("unable to locate date")
}
