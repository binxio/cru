package main

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"log"
	"regexp"
	"sort"
	"strings"
)

var (
	imageReferencePattern = regexp.MustCompile(`(?P<name>[a-zA-Z0-9\-\.]+(:[0-9]+)?/[\w\-/]+)((:(?P<tag>[\w][\w\.-]+))|(@(?P<digest>[a-zA-Z][a-zA-Z0-9]*:[0-9a-fA-F+\.-_]{32,})))`)
	submatchNames         = imageReferencePattern.SubexpNames()
)

type ContainerImageReference struct {
	tag    string
	name   string
	digest string
}

func NewContainerImageReference(original string) (*ContainerImageReference, error) {
	subMatches := imageReferencePattern.FindStringSubmatch(original)
	if len(subMatches) == 0 {
		return nil, fmt.Errorf("'%s' is not a docker image reference", original)
	}

	var result ContainerImageReference
	for i, name := range submatchNames {
		switch name {
		case "name":
			result.name = subMatches[i]
		case "tag":
			result.tag = subMatches[i]
		case "digest":
			result.digest = subMatches[i]
		default:
			// ignore
		}
	}
	if result.tag == "" && result.digest == "" {
		return nil, fmt.Errorf("docker image reference without a tag or digest")
	}

	return &result, nil
}
func MustNewContainerImageReference(original string) *ContainerImageReference {
	r, err := NewContainerImageReference(original)
	if err != nil {
		log.Fatal(err)
	}
	return r
}

func (r ContainerImageReference) String() string {
	builder := strings.Builder{}
	builder.WriteString(r.name)
	if r.tag != "" {
		builder.WriteString(":")
		builder.WriteString(r.tag)
	}
	if r.digest != "" {
		builder.WriteString("@")
		builder.WriteString(r.digest)
	}
	return builder.String()
}

func FindAllContainerImageReference(content []byte) []ContainerImageReference {
	var result = make([]ContainerImageReference, 0)
	allMatches := imageReferencePattern.FindAllIndex(content, -1)
	for _, match := range allMatches {
		s := string(content[match[0]:match[1]])
		r, err := NewContainerImageReference(s)
		if err == nil {
			result = append(result, *r)
		}
	}
	sort.Sort(ContainerImageReferences(result))
	return Unique(result)
}

func (r ContainerImageReference) SameRepository(o ContainerImageReference) bool {
	return r.name == o.name
}

func (a ContainerImageReference) Compare(b ContainerImageReference) int {
	return strings.Compare(a.String(), b.String())
}

func (r ContainerImageReference) FindLatest() (ContainerImageReference, error) {
	tags, err := crane.ListTags(r.name)
	if err == nil {
		fmt.Println(tags)
	}
	return r, nil
}

func Unique(refs []ContainerImageReference) []ContainerImageReference {
	keys := make(map[string]bool)
	result := []ContainerImageReference{}
	for _, ref := range refs {
		if _, value := keys[ref.String()]; !value {
			keys[ref.String()] = true
			result = append(result, ref)
		}
	}
	return result
}

func (r ContainerImageReference) ResolveToDigest() (*ContainerImageReference, error) {
	ref, err := name.ParseReference(r.String())
	if err != nil {
		return nil, err
	}

	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return nil, err
	}
	digest, err := img.Digest()
	if err != nil {
		return nil, fmt.Errorf("failed to get digest for %s, %s", r, err)
	}

	return &ContainerImageReference{name: r.name, tag: "", digest: digest.String()}, nil
}

func (a ContainerImageReferences) ResolveToDigest() (ContainerImageReferences, error) {
	result := make([]ContainerImageReference, 0, a.Len())
	for _, r := range a {
		rr, err := r.ResolveToDigest()
		if err != nil {
			return nil, err
		}
		result = append(result, *rr)
	}
	return result, nil
}

type ContainerImageReferences []ContainerImageReference

func (a ContainerImageReferences) Len() int           { return len(a) }
func (a ContainerImageReferences) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ContainerImageReferences) Less(i, j int) bool { return (a[i]).Compare(a[j]) < 0 }
