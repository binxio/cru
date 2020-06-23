package main

import (
	"github.com/binxio/cru/cmd"
	"github.com/docopt/docopt-go"
	"log"
	"os"
)

func main() {
	usage := `cru - container image reference updater

Usage:
  cru list [--no-filename] [PATH] ...
  cru update [--dry-run] [--resolve-digest] [--image-reference=REFERENCE] ... [PATH] ...
  cru -h | --help

Options:
  list 				- image references in the specified files and directories 
  update 			- update image references in the specified files and directories
  --dry-run 		- pretent to update the references
  --resolve-digest 	- change the image reference tag to a reference of the digest of the image
  --image-reference - image reference to update
  -h --help    		- show this screen.
`

	args, _ := docopt.ParseDoc(usage)
	paths := args["PATH"].([]string)
	references := make(cmd.ContainerImageReferences, 0)
	if len(paths) == 0 {
		paths = append(paths, ".")
	}
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Fatalf("ERROR: %s is not a file or directory\n", path)
		}
	}

	for _, ref := range args["--image-reference"].([]string) {
		r, err := cmd.NewContainerImageReference(ref)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}
		references = append(references, *r)
	}

	resolveLatest := args["--resolve-digest"].(bool)
	if resolveLatest {
		var err error
		references, err = references.ResolveDigest()
		if err != nil {
			log.Fatal(err)
		}
	}

	if args["list"].(bool) {
		cmd.List(paths, !args["--no-filename"].(bool))
	} else if args["update"].(bool) {
		cmd.Update(paths, references, args["--dry-run"].(bool))
	}
}
