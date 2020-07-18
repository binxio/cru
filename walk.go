package main

import (
	"fmt"
	"golang.org/x/tools/godoc/util"
	"golang.org/x/tools/godoc/vfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/format/gitignore"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Visitor func(cru *Cru, path string) error

type GitIgnorePatterns struct {
	root     string
	patterns []gitignore.Pattern
}

func GitIgnores(baseDir string) GitIgnorePatterns {

	root := filepath.Clean(baseDir)
	repository, err := git.PlainOpenWithOptions(root, &git.PlainOpenOptions{DetectDotGit: true})

	if err == nil {
		wt, err := repository.Worktree()
		if err == nil {
			result, err := gitignore.ReadPatterns(wt.Filesystem, []string{"."})
			if err == nil {
				return GitIgnorePatterns{wt.Filesystem.Root(), result}
			}
		}
	}
	return GitIgnorePatterns{root, []gitignore.Pattern{}}
}

func (i GitIgnorePatterns) Ignore(filename string, isDir bool) bool {
	filename, err := filepath.Abs(filename)
	if err != nil {
		log.Fatalf("could not determine absolute path of filename, %s", err)
	}

	filename = filepath.Clean(filename)
	if strings.HasPrefix(filename, i.root) {
		filename = fmt.Sprintf(".%s", filename[len(i.root):])
	}
	path := strings.Split(filename, "/")
	for _, pattern := range i.patterns {
		r := pattern.Match(path, isDir)
		if r == gitignore.Exclude {
			return true
		}
	}
	return false
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

					if gitIgnorePattern.Ignore(p, info.IsDir()) {
						if info.IsDir() {
							return filepath.SkipDir
						}
						return nil
					}

					if info.IsDir() && filepath.Base(p) == ".git" {
						return filepath.SkipDir
					}

					if util.IsTextFile(vfs.OS(filepath.Dir(p)), filepath.Base(p)) {
						err = visitor(c, p)
						if err != nil {
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
