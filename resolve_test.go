package main

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"testing"
)

func TestImageResolve(t *testing.T) {
	r := MustNewContainerImageReference("mvanholsteijn/paas-monitor:3.0.1")
	rr, err := r.ResolveDigest()
	if err != nil {
		t.Fatal(err)
	}

	ref, err := name.ParseReference(r.String())
	if err != nil {
		t.Fatal(err)
	}

	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		t.Fatal(err)
	}
	digest, err := img.Digest()
	if err != nil {
		t.Fatal(err)
	}

	if digest.String() != rr.digest {
		t.Fatalf("expected %s, got %s", digest, rr.digest)
	}
	rr, err = MustNewContainerImageReference("mvanholsteijn/paas-monitor:3.0.2").ResolveDigest()
	if err != nil {
		t.Fatal(err)
	}
	if digest.String() == rr.digest {
		t.Fatalf("expected different digest than %s, got %s", digest, rr.digest)
	}

}

func TestImageResolves(t *testing.T) {
	references := ContainerImageReferences{*MustNewContainerImageReference(`mvanholsteijn/paas-monitor:3.0.2`),
		*MustNewContainerImageReference(`mvanholsteijn/paas-monitor:3.0.1`)}

	resolved, err := references.ResolveDigest()
	if err != nil {
		t.Fatal(err)
	}
	if references.Len() != resolved.Len() {
		t.Fatal("expected length of arrays to be the same")
	}

	for i, r := range resolved {
		if r.digest == "" {
			t.Fatal("expected digest to be set")
		}
		if r.tag != "" {
			t.Fatalf("expected tag to be cleared, found %s", r.tag)
		}
		if r.name != references[i].name {
			t.Fatalf("expected name to be %s, got %s", references[i].name, r.name)
		}
	}
}
