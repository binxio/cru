package main

import (
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"log"
	"time"
)

func (c *Cru) Commit() error {
	if c.CommitMsg == "" {
		return nil
	}

	for _, path := range c.updatedFiles {

		_, err := c.workTree.Add(path)
		if err != nil {
			return err
		}

		if c.Verbose {
			log.Printf("INFO:%s added to commit\n", c.RelPath(path))
		}
	}

	if !c.DryRun {
		hash, err := c.workTree.Commit(c.CommitMsg, &git.CommitOptions{
			Author: &object.Signature{
				Name:  "cru",
				Email: "cru@binx.io",
				When:  time.Now(),
			},
		})
		if err != nil {
			log.Printf("ERROR: failed to commit changes, %s\n", err)
			c.workTree.Reset(nil)
			return err
		}

		if c.Verbose {
			log.Printf("INFO: changes committed with %s", hash.String()[0:7])
		}
	}
	return nil
}
