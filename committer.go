package main

import (
	"bytes"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"io"
	"log"
	"os"
	"time"
)

func (c *Cru) Commit() error {
	if c.CommitMsg == "" {
		return nil
	}

	for _, path := range c.updatedFiles {

		_, err := c.workTree.Add(c.RelPath(path))
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

func (c *Cru) Push() error {

	if !c.DryRun {
		var progress io.Writer = os.Stderr
		if !c.Verbose {
			progress = &bytes.Buffer{}
		}
		log.Printf("INFO: pushing changes to %s", c.Url)

		auth, _, err := GetAuth(c.Url)
		if err != nil {
			return err
		}
		return c.repository.Push(&git.PushOptions{Auth: auth, Progress: progress})
	} else {
		log.Printf("INFO: changes would be pushed to %s", c.Url)
	}
	return nil
}
