package main

import (
	"fmt"
	"github.com/binxio/cru/ref"
	"github.com/docopt/docopt-go"
	"io/ioutil"
	"log"
	"os"
)

type Cru struct {
	Paths          []string
	List           bool
	Update         bool
	Bump           bool
	NoFilename     bool
	DryRun         bool
	Verbose        bool
	ResolveDigest  bool
	ResolveTag     bool
	CommitMessage  string
	All			   bool
	ImageReferences []string
	Level          string
	imageRefs      ref.ContainerImageReferences
	updatedFiles   []string
	committedFiles []string
}

func (c *Cru) AssertPathsExists() {
	if len(c.Paths) == 0 {
		c.Paths = append(c.Paths, ".")
	}

	for _, path := range c.Paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Fatalf("ERROR: %s is not a file or directory\n", path)
		}
	}
}

func CollectReferences(c *Cru, filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read %s, %s", filename, err)
	}
	c.imageRefs = append(c.imageRefs, ref.FindAllContainerImageReference(content)...)
	return nil
}

func List(c *Cru, filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read %s, %s", filename, err)
	}
	for _, ref := range ref.FindAllContainerImageReference(content) {
		if c.NoFilename {
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
	content, updated := ref.UpdateReferences(content, c.imageRefs, filename, c.Verbose)
	if updated {
		if !c.DryRun {
			err := ioutil.WriteFile(filename, content, 0644)
			if err != nil {
				return fmt.Errorf("failed to overwrite %s with updated references, %s", filename, err)
			}
			c.updatedFiles = append(c.updatedFiles, filename)
		}
	}
	return nil
}

func main() {
	usage := `cru - container image reference updater

Usage:
  cru list   [--verbose] [--no-filename] [PATH] ...
  cru update [--verbose] [--dry-run] [(--resolve-digest|--resolve-tag|--bump)] [--commit=MESSAGE] (--all | --image-reference=REFERENCE ...) [PATH] ...
  cru -h | --help

Options:
--no-filename	    do not print the filename.
--resolve-digest 	change the image reference tag to a reference of the digest of the image.
--resolve-tag		change the image reference tag to the first alternate tag of the reference.
--bump              change the image reference tag to the next available version                
--image-reference=REFERENCE to update.
--dry-run			pretend to run the update, make no changes.
--all               replace all container image reference tags with "latest"
--verbose			show more output.
--commit=MESSAGE	commit the changes with the specified message.

`
	cru := Cru{}

	args, err := docopt.ParseDoc(usage)
	if err != nil {
		log.Fatal(err)
	}

	if err = args.Bind(&cru); err != nil {
		log.Fatal(err)
	}
	cru.imageRefs = make(ref.ContainerImageReferences, 0)

	cru.AssertPathsExists()

	if cru.All {
		if cru.Verbose {
			log.Println("INFO: collecting all container references")
		}
		err = cru.Walk(CollectReferences)
		if err != nil {
			log.Fatalf("%s\n", err)
		}
		for i, _ := range cru.imageRefs {
			cru.imageRefs[i].SetTag("latest")
		}
		cru.imageRefs = cru.imageRefs.RemoveDuplicates()
		log.Printf("INFO: %d image references found\n", len(cru.imageRefs))
		if len(cru.imageRefs) == 0 {
			os.Exit(0)
		}
	}

	for _, r := range cru.ImageReferences {
		r, err := ref.NewContainerImageReference(r)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}
		cru.imageRefs = append(cru.imageRefs, *r)
	}

	if cru.ResolveDigest {
		var err error
		cru.imageRefs, err = cru.imageRefs.ResolveDigest()
		if err != nil {
			log.Fatal(err)
		}
	}

	if cru.ResolveTag {
		var err error
		cru.imageRefs, err = cru.imageRefs.ResolveTag()
		if err != nil {
			log.Fatal(err)
		}
	}

	if cru.Bump {
		// todo
	}

	if cru.List {
		if err = cru.Walk(List); err != nil {
			log.Fatal(err)
		}
	} else if cru.Update {
		if err = cru.Walk(Update); err != nil {
			log.Fatal(err)
		}
		if len(cru.updatedFiles) > 0 {
			if cru.CommitMessage != "" {
				if err = cru.Commit(); err != nil {
					log.Fatal(err)
				}
				log.Printf("INFO: %d out of %d updated files committed", len(cru.committedFiles), len(cru.updatedFiles))
			} else {
				log.Printf("INFO: %d files updated", len(cru.updatedFiles))
			}
		} else {
			log.Println("INFO: no files were updated by cru")
		}

	}
}
