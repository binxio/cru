package cmd

import (
	"fmt"
	"golang.org/x/tools/godoc/util"
	"golang.org/x/tools/godoc/vfs"
	"io/ioutil"
	"os"
	"path/filepath"
)

func searchReferences(filename string) (ContainerImageReferences, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read %s, %s", filename, err)
	}
	return FindAllContainerImageReference(content), nil
}

func SearchAllReferences(paths []string) (ContainerImageReferenceMap, error) {
	result := make(map[string]ContainerImageReferences)
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("%s is not a valid path or is not readable, %s", path, err)
		}
		if info.IsDir() {
			err := filepath.Walk(path,
				func(p string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if util.IsTextFile(vfs.OS(filepath.Dir(p)), filepath.Base(p)) {
						refs, err := searchReferences(p)
						if err != nil {
							return err
						}
						result[p] = refs
					}
					return nil
				})
			if err != nil {
				return nil, err
			}
		} else {
			if !util.IsTextFile(vfs.OS(filepath.Dir(path)), filepath.Base(path)) {
				return nil, fmt.Errorf("%s is not a text file", path)
			}
			refs, err := searchReferences(path)
			if err != nil {
				return nil, err
			}
			result[path] = refs
		}
	}
	return result, nil
}

type ContainerImageReferenceMap map[string]ContainerImageReferences

func (m ContainerImageReferenceMap) Merge() ContainerImageReferences {
	result := make(ContainerImageReferences, 0)
	for _, references := range m {
		result = append(result, references...)
	}
	return result.Unique()
}

func List(paths []string, printFilenames bool) error {
	refMap, err := SearchAllReferences(paths); if err == nil {
		if printFilenames {
			for path, references := range refMap {
				for _, ref := range references {
					fmt.Printf("%s:%s\n", path, ref)
				}
			}
		} else {
			references := refMap.Merge()
			for _, ref := range references {
				fmt.Printf("%s\n", ref)
			}
		}
	}
	return err
}
