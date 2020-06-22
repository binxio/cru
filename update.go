package main

import (
	"bytes"
	"fmt"
	"golang.org/x/tools/godoc/util"
	"golang.org/x/tools/godoc/vfs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func updateReference(content []byte, reference ContainerImageReference) ([]byte, bool) {
	previous := 0
	updated := false
	result := bytes.Buffer{}
	allMatches := imageReferencePattern.FindAllIndex(content, -1)
	for _, match := range allMatches {
		s := string(content[match[0]:match[1]])
		r, err := NewContainerImageReference(s)
		if err == nil && r.name == reference.name {
			if r.String() != reference.String() {
				updated = true
				result.Write(content[previous:match[0]])
				result.Write([]byte(reference.String()))
				previous = match[1]
			}
		}
	}
	if previous < len(content) {
		result.Write(content[previous:len(content)])
	}

	return result.Bytes(), updated
}

func updateReferences(content []byte, references []ContainerImageReference) ([]byte, bool) {
	updated := false
	changed := false
	for _, ref := range references {
		if content, changed = updateReference(content, ref); changed {
			updated = true
		}
	}
	return content, updated
}

func updateFile(filename string, references []ContainerImageReference, dryRun bool) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read %s, %s", filename, err)
	}
	content, updated := updateReferences(content, references)
	if updated && !dryRun {
		err := ioutil.WriteFile(filename, content, 0644)
		if err != nil {
			return fmt.Errorf("failed to overwrite %s with updated references, %s", filename, err)
		}
		log.Printf("INFO: updated %s\n", filename)
	}
	return nil
}

func update(paths []string, references []ContainerImageReference, resolveLatest bool, dryRun bool) error {
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("%s is not a valid path or is not readable, %s", path, err)
		}
		if info.IsDir() {
			err := filepath.Walk(path,
				func(p string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if util.IsTextFile(vfs.OS(filepath.Dir(p)), filepath.Base(p)) {
						updateFile(p, references, dryRun)
					}
					return nil
				})
			if err != nil {
				return err
			}
		} else {
			log.Println(path)
			if !util.IsTextFile(vfs.OS(filepath.Dir(path)), filepath.Base(path)) {
				return fmt.Errorf("%s is not a text file", path)
			}
			updateFile(path, references, dryRun)
		}
	}
	return nil
}
