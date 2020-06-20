package main

import (
	"fmt"
	"github.com/docopt/docopt-go"
	"golang.org/x/tools/godoc/util"
	"golang.org/x/tools/godoc/vfs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func list_references(filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read %s, %s", filename, err)
	}

	for _, ref := range FindAllContainerImageReference(content) {
		fmt.Printf("%s: %s\n", filename, ref)
	}

	return nil
}

func list(paths []string) error {
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
						list_references(p)
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
			list_references(path)
		}
	}
	return nil
}

func main() {
	usage := `cru - container image reference updater

Usage:
  cru list [PATH] ...
  cru -h | --help

Options:
  list - image references in the specified files and directories
  -h --help     Show this screen.
`
	args, _ := docopt.ParseDoc(usage)
	if _, err := args.Bool("list"); err == nil {
		paths := args["PATH"].([]string)
		if len(paths) == 0 {
			paths = append(paths, ".")
		}
		list(paths)
	}
}
