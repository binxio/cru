package main

import (

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"time"

	"log"

	"path/filepath"

)


func (c *Cru) Commit() error {
	workTrees := make(map[string]*git.Worktree)
	toCommit := make(map[string]*git.Worktree)

	if c.commitMessage == "" {
		return nil
	}
	if len(c.updatedFiles) == 0 {
		log.Printf("no files updated by cru\n")
	}
	absolutePaths := make([]string, len(c.updatedFiles), len(c.updatedFiles))
	for i, path := range c.updatedFiles {

		path, err := filepath.Abs(path)
		absolutePaths[i] = path

		if err != nil {
			return err
		}
	}

	for _, path := range absolutePaths {

		path = filepath.Clean(path)
		root := filepath.Dir(path)
		repository, err := git.PlainOpenWithOptions(root, &git.PlainOpenOptions{DetectDotGit: true})
		if err != nil {
			if err == git.ErrRepositoryNotExists {
				if c.verbose {
					log.Printf("INFO: %s is not under control of git\n", path)
				}
				continue
			}
			return err
		}
		wt, err := repository.Worktree()
		if err != nil {
			return err
		}
		workTrees[wt.Filesystem.Root()] = wt
		status, err := wt.Status()
		if err != nil {
			return err
		}

		if status.IsClean() {
			log.Printf("INFO: %s was updated but not under control of git", path)
			return nil
		}
		for p, _ := range status {
			if path == filepath.Join(wt.Filesystem.Root(), p) {
				wt.Add(p)
				log.Printf("INFO: add %s to commit\n", p)
				toCommit[wt.Filesystem.Root()] = wt
			}
		}
	}

	for root, wt := range toCommit {

		hash, err := wt.Commit(c.commitMessage, &git.CommitOptions{
			Author: &object.Signature{
				Name:  "cru",
				Email: "cru@binx.io",
				When:  time.Now(),
			},
		})
		if err != nil {
			log.Printf("ERROR: failed to commit changes to %s, %s\n", root, err)
			wt.Reset(nil)
			return err
		}

		log.Printf("INFO: committed changes to %s as %s", root, (hash.String())[0:6])
	}
	return nil
}
