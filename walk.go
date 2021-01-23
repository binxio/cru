package main

import (
	"fmt"
	"golang.org/x/tools/godoc/util"
	"golang.org/x/tools/godoc/vfs"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/format/gitignore"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Visitor func(cru *Cru, path string) error

type GitIgnorePatterns []gitignore.Pattern

func (i GitIgnorePatterns) Ignore(filename string, isDir bool) bool {
	path := strings.Split(filename, "/")
	for _, pattern := range i {
		r := pattern.Match(path, isDir)
		if r == gitignore.Exclude {
			return true
		}
	}
	return false
}

type BillyVFS struct {
	fs billy.Filesystem
}

func (this BillyVFS) Open(name string) (vfs.ReadSeekCloser, error) {
	return this.fs.Open(name)
}

func (c *Cru) walkFn(visitor Visitor, filename string, info os.FileInfo) error {

	if info.IsDir() {
		var patterns GitIgnorePatterns
		patterns, err := gitignore.ReadPatterns(*c.filesystem, strings.Split(filename, "/"))
		if err != nil {
			return err
		}

		if patterns.Ignore(filename, info.IsDir()) {
			return nil
		}
		if filepath.Base(filename) == ".git" {
			return nil
		}

		dir, err := (*c.filesystem).ReadDir(filename)
		if err != nil {
			return err
		}
		for _, f := range dir {
			childPath := path.Join(filename, f.Name())
			childInfo, err := (*c.filesystem).Stat(childPath)
			if err != nil {
				return err
			}
			if !patterns.Ignore(childPath, childInfo.IsDir()) {
				err = c.walkFn(visitor, childPath, childInfo)
				if err != nil {
					return err
				}
			}
		}
	} else {
		fs := BillyVFS{*c.filesystem}
		if util.IsTextFile(fs, filename) {
			return visitor(c, filename)
		}
	}
	return nil
}

func (c *Cru) AbsPath(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Clean(filepath.Join(c.cwd, filename))
}

func (c *Cru) RelPath(filename string) string {
	if !filepath.IsAbs(filename) {
		return filename
	}
	if rel, err := filepath.Rel(c.cwd, filename); err == nil {
		return rel
	}
	return filename
}

func (c *Cru) Walk(visitor Visitor) (err error) {
	for _, p := range c.Path {
		filename := c.AbsPath(p)
		info, err := (*c.filesystem).Stat(filename)
		if err != nil {
			return fmt.Errorf("%s is not a valid path or is not readable, %s", p, err)
		}
		err = c.walkFn(visitor, filename, info)
		if err != nil {
			return err
		}
	}
	return nil
}
