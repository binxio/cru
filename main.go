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
--no-filename	    do not print the filename.
--resolve-digest 	change the image reference tag to a reference of the digest of the image.
--image-reference=REFERENCE to update.
--dry-run			pretend to run the update, make no changes.
`

	args, err := docopt.ParseDoc(usage)
	if err != nil {
		log.Fatal(err)
	}
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
