package main

import (
	"bytes"
	"fmt"
	"github.com/binxio/cru/ref"
	"github.com/docopt/docopt-go"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type Cru struct {
	Path           []string
	List           bool
	Update         bool
	Serve          bool
	Port           string
	Bump           bool
	NoFilename     bool
	DryRun         bool
	Verbose        bool
	ResolveDigest  bool
	ResolveTag     bool
	All            bool
	ImageReference []string
	Level          string
	Url            string `docopt:"--repository"`
	CommitMsg      string `docopt:"--commit"`
	Branch         string `docopt:"--branch"`
	imageRefs      ref.ContainerImageReferences
	updatedFiles   []string
	committedFiles []string
	repository     *git.Repository
	workTree       *git.Worktree
	filesystem     *billy.Filesystem
	cwd            string
}

func (c *Cru) AssertPathsExists() {
	if len(c.Path) == 0 {
		c.Path = append(c.Path, ".")
	}

	for _, path := range c.Path {
		if _, err := (*c.filesystem).Stat(c.AbsPath(path)); os.IsNotExist(err) {
			log.Fatalf("ERROR: %s is not a file or directory\n", path)
		}
	}
}

func CollectReferences(c *Cru, filename string) error {
	content, err := c.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read %s, %s", filename, err)
	}
	c.imageRefs = append(c.imageRefs, ref.FindAllContainerImageReference(content)...)
	return nil
}

func (c *Cru) ReadFile(filename string) (content []byte, err error) {
	var file billy.File
	file, err = (*c.filesystem).Open(filename)
	if err != nil {
		return
	}
	defer file.Close()
	return ioutil.ReadAll(file)
}

func (c *Cru) WriteFile(filename string, content []byte, perm os.FileMode) error {
	file, err := (*c.filesystem).OpenFile(filename, os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(content)
	return err
}

func List(c *Cru, filename string) error {

	content, err := c.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read %s, %s", filename, err)
	}
	for _, ref := range ref.FindAllContainerImageReference(content) {
		if c.NoFilename {
			fmt.Printf("%s\n", ref.String())
		} else {
			if relative, err := filepath.Rel(c.cwd, filename); err == nil {
				filename = relative
			}
			fmt.Printf("%s:%s\n", filename, ref.String())
		}
	}
	return nil
}

func Update(c *Cru, filename string) error {
	content, err := c.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read %s, %s", c.RelPath(filename), err)
	}
	content, updated := ref.UpdateReferences(content, c.imageRefs, c.RelPath(filename), c.Verbose)
	if updated {
		if !c.DryRun {
			err := c.WriteFile(filename, content, 0644)
			if err != nil {
				return fmt.Errorf("failed to overwrite %s with updated references, %s", c.RelPath(filename), err)
			}
		}
		c.updatedFiles = append(c.updatedFiles, filename)
	}
	return nil
}

func (c *Cru) ConnectToRepository() error {
	if c.Url != "" {
		var progressReporter io.Writer = os.Stderr
		if !c.Verbose {
			progressReporter = &bytes.Buffer{}
		}
		repository, err := Clone(c.Url, progressReporter)
		if err != nil {
			return err
		}
		c.repository = repository

		wt, err := repository.Worktree()
		if err != nil {
			return err
		}

		c.workTree = wt
		if c.Branch != "" {
			var branch *plumbing.Reference
			if branches, err := repository.Branches(); err == nil {
				branches.ForEach(func(ref *plumbing.Reference) error {
					if ref.Name().Short() == c.Branch {
						branch = ref
					}
					return nil
				})
			}
			if err != nil {
				return err
			}
			if branch == nil {
				return fmt.Errorf("ERROR: branch %s not found", c.Branch)
			}
			err = wt.Checkout(&git.CheckoutOptions{Branch: branch.Name()})
			if err != nil {
				return err
			}
		}
		c.filesystem = &wt.Filesystem
		c.cwd = "/"
	} else {
		cwd, err := filepath.Abs(".")
		if err != nil {
			return err
		}
		c.cwd = cwd
		fs := osfs.New("/")
		c.filesystem = &fs
	}
	return nil
}

func main() {
	usage := `cru - container image reference updater

Usage:
  cru list   [--verbose] [--no-filename] [--repository=URL [--branch=BRANCH]] [PATH] ...
  cru update [--verbose] [--dry-run] [(--resolve-digest|--resolve-tag)] [--repository=URL [--branch=BRANCH] [--commit=MESSAGE]] (--all | --image-reference=REFERENCE ...) [PATH] ...
  cru serve  [--verbose] [--dry-run] [--port=PORT] --repository=URL --branch=BRANCH [PATH] ...

Options:
--no-filename	    do not print the filename.
--resolve-digest 	change the image reference tag to a reference of the digest of the image.
--resolve-tag		change the image reference tag to the first alternate tag of the reference.
--image-reference=REFERENCE to update.
--dry-run			pretend to run the update, make no changes.
--all               replace all container image reference tags with "latest"
--verbose			show more output.
--commit=MESSAGE	commit the changes with the specified message.
--repository=URL    to read and/or update.
--branch=BRANCH     to update.
--port=PORT         to listen on, defaults to 8080 or PORT environment variable.
`
	cru := Cru{}

	args, err := docopt.ParseDoc(usage)
	if err != nil {
		log.Fatal(err)
	}

	if err = args.Bind(&cru); err != nil {
		log.Fatal(err)
	}

	if err = cru.ConnectToRepository(); err != nil {
		log.Fatal(err)
	}
	cru.AssertPathsExists()
	cru.imageRefs = make(ref.ContainerImageReferences, 0)

	if cru.Serve {
		if cru.Url == "" {
			log.Fatalf("cru as a service requires an git url.")
		}
		cru.ListenAndServe()
	}

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

	for _, r := range cru.ImageReference {
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

	if cru.List {
		if err = cru.Walk(List); err != nil {
			log.Fatal(err)
		}
	} else if cru.Update {
		if err = cru.Walk(Update); err != nil {
			log.Fatal(err)
		}
		if len(cru.updatedFiles) > 0 {
			log.Printf("INFO: updated a total of %d files", len(cru.updatedFiles))
			if cru.CommitMsg != "" {
				if _, err = cru.Commit(); err != nil {
					log.Fatal(err)
				}
				if !IsLocalEndpoint(cru.Url) {
					if err = cru.Push(); err != nil {
						log.Fatal(err)
					}
				}
			}
		} else {
			log.Println("INFO: no files were updated by cru")
		}

	}
}
