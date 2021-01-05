package main

import (
	"fmt"
	"github.com/binxio/cru/ref"
	"github.com/binxio/cru/tag"
	"github.com/docopt/docopt-go"
	"github.com/google/go-containerregistry/pkg/name"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Cru struct {
	Path            []string
	List            bool
	Update          bool
	Bump            bool
	NoFilename      bool
	DryRun          bool
	Verbose         bool
	ResolveDigest   bool
	ResolveTag      bool
	All             bool
	ImageReference  []string
	imageReferences ref.ContainerImageReferences
	bumpReferences  map[string]string
	bumpOrder       []string
}

func (c *Cru) AssertPathsExists() {
	if len(c.Path) == 0 {
		c.Path = append(c.Path, ".")
	}

	for _, path := range c.Path {
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
	content, updated := ref.UpdateReferences(content, c.imageReferences, filename)
	if updated {
		if !c.DryRun {
			err := ioutil.WriteFile(filename, content, 0644)
			if err != nil {
				return fmt.Errorf("failed to overwrite %s with updated references, %s", filename, err)
			}
		}
	}
	return nil
}

func Bump(c *Cru, filename string) error {
	var updated = false
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read %s, %s", filename, err)
	}
	for _,k := range c.bumpOrder {
		v, _ := c.bumpReferences[k]
		s := string(content)
		ns := strings.ReplaceAll(s, k, v)
		if s != ns {
			updated = true
			log.Printf("INFO: updating reference %s to %s in %s", k, v, filename)
			content = append(make([]byte, 0, len(ns)), ns...)
		}
	}
	if updated {
		if !c.DryRun {
			err := ioutil.WriteFile(filename, content, 0644)
			if err != nil {
				return fmt.Errorf("failed to overwrite %s with updated references, %s", filename, err)
			}
		}
	}
	return nil
}

func (c *Cru) bumpOrderDepth(ref string) int {
	if v, ok := c.bumpReferences[ref]; ok && v != ref {
		return c.bumpOrderDepth(v) + 1
	} else {
		return 0
	}
}

func (c *Cru) determineBumpOrder() {
	var highest = 0
	var ordered = make(map[int][]string, len(c.bumpReferences))
	for ref, _ := range c.bumpReferences {
		depth := c.bumpOrderDepth(ref)
		if depth > highest {
			highest = depth
		}
		if v, ok := ordered[depth] ; ok {
			ordered[depth] = append(v, ref)
		} else {
			ordered[depth] = []string{ref}
		}
	}

	for depth := 1; depth <= highest; depth = depth + 1 {
		if v, ok := ordered[depth]; ok {
			c.bumpOrder = append(c.bumpOrder, v...)
		}
	}
}

func main() {
	usage := `cru - container image reference updater

Usage:
  cru list   [--verbose] [--no-filename] [PATH] ...
  cru update [--verbose] [--dry-run] [(--resolve-digest|--resolve-tag)] (--all | --image-reference=REFERENCE ...) [PATH] ...
  cru bump   [--verbose] [--dry-run] (--all | --image-reference=REFERENCE ...) [PATH] ...

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

	if err := args.Bind(&cru); err != nil {
		log.Fatal(err)
	}
	cru.imageReferences = make(ref.ContainerImageReferences, 0)

	cru.AssertPathsExists()

	if cru.All  {
		if cru.Verbose {
			log.Print("INFO: collecting all container references")
		}
		err = cru.Walk(CollectReferences)
		if err != nil {
			log.Fatalf("%s\n", err)
		}
		if cru.Verbose {
			log.Printf("INFO: found %d container references\n", len(cru.imageReferences))
		}
	}

	for _, r := range cru.ImageReference {
		r, err := ref.NewContainerImageReference(r)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}
		cru.imageReferences = append(cru.imageReferences, *r)
	}

	if cru.ResolveDigest {
		var err error
		cru.imageReferences, err = cru.imageReferences.ResolveDigest()
		if err != nil {
			log.Fatal(err)
		}
	}

	if cru.ResolveTag {
		var err error
		cru.imageReferences, err = cru.imageReferences.ResolveTag()
		if err != nil {
			log.Fatal(err)
		}
	}

	if cru.List {
		cru.Walk(List)
	} else if cru.Update {
		if cru.All {
			for i, _ := range cru.imageReferences {
				cru.imageReferences[i].SetTag("latest")
			}
		}
		cru.imageReferences = cru.imageReferences.RemoveDuplicates()
		if len(cru.imageReferences) > 0 {
			cru.Walk(Update)
		}
	} else if cru.Bump {
		cru.imageReferences = cru.imageReferences.RemoveDuplicates()
		cru.bumpReferences = make(map[string]string, len(cru.ImageReference))
		for _, r := range cru.imageReferences {
			ref, err := name.ParseReference(r.String())
			if err != nil {
				log.Fatal(err)
			}
			t, ok := ref.(name.Tag)
			if ok {
				nt, err := tag.GetNextVersion(t)
				if err != nil {
					log.Fatal(err)
				}

				cru.bumpReferences[r.String()] = nt.String()
			} else {
				log.Printf("INFO: skipping %s as it is not a tag reference", r)
			}
		}

		if len(cru.bumpReferences) > 0 {
			cru.determineBumpOrder()
			cru.Walk(Bump)
		}
	}
}
