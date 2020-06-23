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
  cru update [--dry-run] [--resolve-digest] [--all | [--image-reference=REFERENCE] ...] [PATH] ...
  cru -h | --help

Options:
--no-filename	    do not print the filename.
--resolve-digest 	change the image reference tag to a reference of the digest of the image.
--image-reference=REFERENCE to update.
--dry-run			pretend to run the update, make no changes.
--all               replace all container image reference tags with "latest"
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

	if args["--all"].(bool) {
		refMap, err := cmd.SearchAllReferences(paths)
		if err != nil {
			log.Fatalf("%s\n", err)
		}
		references = refMap.Merge()
		for i, _ := range references {
			references[i].SetTag("latest")
		}
		references = references.RemoveDuplicates()
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
