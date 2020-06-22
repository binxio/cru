package main

import (
	"github.com/docopt/docopt-go"
	"log"
	"os"
)

func main() {
	usage := `cru - container image reference updater

Usage:
  cru list [--no-filename] [PATH] ...
  cru update [--dry-run] [--resolve-latest] [--image-reference=REFERENCE] ... [PATH] ...
  cru -h | --help

Options:
  list - image references in the specified files and directories
  -h --help     Show this screen.
`

	args, _ := docopt.ParseDoc(usage)
	paths := args["PATH"].([]string)
	references := make([]ContainerImageReference, 0)
	if len(paths) == 0 {
		paths = append(paths, ".")
	}
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Fatalf("ERROR: %s is not a file or directory\n", path)
		}
	}

	for _, ref := range args["--image-reference"].([]string) {
		r, err := NewContainerImageReference(ref)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}
		references = append(references, *r)
	}

	if args["list"].(bool) {
		list(paths, !args["--no-filename"].(bool))
	} else if args["update"].(bool) {
		resolveLatest := args["--resolve-latest"].(bool)
		if resolveLatest {
			log.Fatal("--resolve-latest not yet supported")
		}
		update(paths, references, !args["--resolve-latest"].(bool), args["--dry-run"].(bool))
	}
}
