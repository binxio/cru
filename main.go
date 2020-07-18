package main

import (
	"fmt"
	"github.com/binxio/cru/ref"
	"github.com/docopt/docopt-go"
	"golang.org/x/tools/godoc/util"
	"golang.org/x/tools/godoc/vfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/format/gitignore"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type Cru struct {
	paths           []string
	noFilename      bool
	dryRun          bool
	resolveDigest   bool
	imageReferences ref.ContainerImageReferences
}

func (c *Cru) AssertPathsExists() {
	if len(c.paths) == 0 {
		c.paths = append(c.paths, ".")
	}

	for _, path := range c.paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Fatalf("ERROR: %s is not a file or directory\n", path)
		}
	}
}

type Visitor func(cru *Cru, path string) error

type GitIgnorePatterns []gitignore.Pattern

func GitIgnores(baseDir string) GitIgnorePatterns {
	repository, err := git.PlainOpenWithOptions(baseDir, &git.PlainOpenOptions{DetectDotGit:true})
	if err == nil {
		wt, err := repository.Worktree(); if err == nil {
			return wt.Excludes
		}
	}
	return []gitignore.Pattern{}
}

func (p GitIgnorePatterns) Ignore(filename string, isDir bool) bool {
	path := filepath.SplitList(filename)
	for _, pattern := range p {
		if r := pattern.Match(path, isDir); r == gitignore.Exclude {
			return true
		}
	}
	return false
}

func CollectReferences(c *Cru, filename string)  error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read %s, %s", filename, err)
	}
	c.imageReferences = append(c.imageReferences, ref.FindAllContainerImageReference(content)...)
	return nil
}

func List(c *Cru, filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read %s, %s", filename, err)
	}
	for _, ref := range ref.FindAllContainerImageReference(content) {
		if c.noFilename {
			fmt.Printf("%s\n", ref.String())
		} else {
			fmt.Printf("%s:%s\n", filename, ref.String())
		}
	}
	return nil
}


func Update(c *Cru, filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read %s, %s", filename, err)
	}
	content, updated := ref.UpdateReferences(content, c.imageReferences)
	if updated {
		log.Printf("INFO: updating %s\n", filename)
		if !c.dryRun {
			err := ioutil.WriteFile(filename, content, 0644)
			if err != nil {
				return fmt.Errorf("failed to overwrite %s with updated references, %s", filename, err)
			}
		}
	}
	return nil
}

func (c *Cru) Walk(visitor Visitor) error {
	for _, path := range c.paths {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("%s is not a valid path or is not readable, %s", path, err)
		}
		if info.IsDir() {
			gitIgnorePattern := GitIgnores(path)
			err := filepath.Walk(path,
				func(p string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if gitIgnorePattern.Ignore(p, false) {
						return nil
					}
					if util.IsTextFile(vfs.OS(filepath.Dir(p)), filepath.Base(p)) {
						err = visitor(c, p); if err != nil {
							return err
						}
					}
					return nil
				})
			if err != nil {
				return err
			}
		} else {
			if !util.IsTextFile(vfs.OS(filepath.Dir(path)), filepath.Base(path)) {
				return fmt.Errorf("%s is not a text file", path)
			}
			if err = visitor(c, path); err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	usage := `cru - container image reference updater

Usage:
  cru list   [--no-filename] [PATH] ...
  cru update [--dry-run] [--resolve-digest] [--all | [--image-reference=REFERENCE] ...] [PATH] ...
  cru -h | --help

Options:
--no-filename	    do not print the filename.
--resolve-digest 	change the image reference tag to a reference of the digest of the image.
--image-reference=REFERENCE to update.
--dry-run			pretend to run the update, make no changes.
--all               replace all container image reference tags with "latest"

`
	cru := Cru{}

	args, err := docopt.ParseDoc(usage)
	if err != nil {
		log.Fatal(err)
	}
	cru.paths = args["PATH"].([]string)
	cru.dryRun = args["--dry-run"].(bool)
	cru.noFilename = args["--no-filename"].(bool)
	cru.resolveDigest= args["--resolve-digest"].(bool)
	cru.imageReferences = make(ref.ContainerImageReferences, 0)

	cru.AssertPathsExists()


	if args["--all"].(bool) {
		err = cru.Walk(CollectReferences)
		if err != nil {
			log.Fatal("%s\n", err)
		}
		for i, _ := range cru.imageReferences {
			cru.imageReferences [i].SetTag("latest")
		}
		cru.imageReferences = cru.imageReferences.RemoveDuplicates()
	}

	for _, r := range args["--image-reference"].([]string) {
		r, err := ref.NewContainerImageReference(r)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}
		cru.imageReferences = append(cru.imageReferences, *r)
	}

	resolveLatest := args["--resolve-digest"].(bool)
	if resolveLatest {
		var err error
		cru.imageReferences, err = cru.imageReferences.ResolveDigest()
		if err != nil {
			log.Fatal(err)
		}
	}

	if args["list"].(bool) {
		cru.Walk(List)
	} else if args["update"].(bool) {
		cru.Walk(Update)
	}
}
