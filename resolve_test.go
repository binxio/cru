package main

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"testing"
)

func TestImageResolve(t *testing.T) {
	r := MustNewContainerImageReference("mvanholsteijn/paas-monitor:3.0.1")
	rr, err := r.Resolve()
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
	rr, err = MustNewContainerImageReference("mvanholsteijn/paas-monitor:3.0.2").Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if digest.String() == rr.digest {
		t.Fatalf("expected different digest than %s, got %s", digest, rr.digest)
	}

}
