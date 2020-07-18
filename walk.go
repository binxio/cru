package main

import (
	"os"
	"fmt"
	"golang.org/x/tools/godoc/util"
	"golang.org/x/tools/godoc/vfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/format/gitignore"
	"path/filepath"
)

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
