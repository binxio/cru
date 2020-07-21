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
	paths           []string
	noFilename      bool
	dryRun          bool
	verbose         bool
	resolveDigest   bool
	resolveTag      bool
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

func CollectReferences(c *Cru, filename string) error {
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
	content, updated := ref.UpdateReferences(content, c.imageReferences, filename)
	if updated {
		if !c.dryRun {
			err := ioutil.WriteFile(filename, content, 0644)
			if err != nil {
				return fmt.Errorf("failed to overwrite %s with updated references, %s", filename, err)
			}
		}
	}
	return nil
}

func main() {
	usage := `cru - container image reference updater

Usage:
  cru list   [--verbose] [--no-filename] [PATH] ...
  cru update [--verbose] [--dry-run] [(--resolve-digest|--resolve-tag)] (--all | --image-reference=REFERENCE ...) [PATH] ...
  cru -h | --help

Options:
--no-filename	    do not print the filename.
--resolve-digest 	change the image reference tag to a reference of the digest of the image.
--resolve-tag		change the image reference tag to the first alternate tag of the reference.
--image-reference=REFERENCE to update.
--dry-run			pretend to run the update, make no changes.
--all               replace all container image reference tags with "latest"
--verbose			show more output.

`
	cru := Cru{}

	args, err := docopt.ParseDoc(usage)
	if err != nil {
		log.Fatal(err)
	}
	cru.paths = args["PATH"].([]string)
	cru.dryRun = args["--dry-run"].(bool)
	cru.verbose = args["--verbose"].(bool)
	cru.noFilename = args["--no-filename"].(bool)
	cru.resolveDigest = args["--resolve-digest"].(bool)
	cru.resolveTag = args["--resolve-tag"].(bool)
	cru.imageReferences = make(ref.ContainerImageReferences, 0)

	cru.AssertPathsExists()

	if args["--all"].(bool) {
		if cru.verbose {
			log.Println("INFO: collecting all container references")
		}
		err = cru.Walk(CollectReferences)
		if err != nil {
			log.Fatalf("%s\n", err)
		}
		for i, _ := range cru.imageReferences {
			cru.imageReferences[i].SetTag("latest")
		}
		cru.imageReferences = cru.imageReferences.RemoveDuplicates()
		log.Printf("INFO: %d image references found\n", len(cru.imageReferences))
		if len(cru.imageReferences) == 0 {
			os.Exit(0)
		}
	}

	for _, r := range args["--image-reference"].([]string) {
		r, err := ref.NewContainerImageReference(r)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}
		cru.imageReferences = append(cru.imageReferences, *r)
	}

	if cru.resolveDigest {
		var err error
		cru.imageReferences, err = cru.imageReferences.ResolveDigest()
		if err != nil {
			log.Fatal(err)
		}
	}

	if cru.resolveTag {
		var err error
		cru.imageReferences, err = cru.imageReferences.ResolveTag()
		if err != nil {
			log.Fatal(err)
		}
	}

	if args["list"].(bool) {
		cru.Walk(List)
	} else if args["update"].(bool) {
		if cru.verbose {
		}
		cru.Walk(Update)
	}
}
