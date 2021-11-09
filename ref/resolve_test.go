package ref

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"testing"
)

func TestImageResolve(t *testing.T) {
	r := MustNewContainerImageReference("gcr.io/binx-io-public/paas-monitor:latest")
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

	if digest.String() != rr.Digest {
		t.Fatalf("expected %s, got %s", digest, rr.Digest)
	}
	rr, err = MustNewContainerImageReference("mvanholsteijn/paas-monitor:3.1.0").ResolveDigest()
	if err != nil {
		t.Fatal(err)
	}
	if digest.String() == rr.Digest {
		t.Fatalf("expected different Digest than %s, got %s", digest, rr.Digest)
	}

}

func TestImageResolves(t *testing.T) {
	references := ContainerImageReferences{
		*MustNewContainerImageReference(`mvanholsteijn/paas-monitor:3.1.0`),
		*MustNewContainerImageReference(`mvanholsteijn/paas-monitor:3.1.0`),
		*MustNewContainerImageReference(`mvanholsteijn/paas-monitor:3.1.0@sha256:c0717cab955aff0a3d2f6bb975808ba9708d8385bcf01a18e23ff436f07c1fb3`),
	}

	resolved, err := references.ResolveDigest()
	if err != nil {
		t.Fatal(err)
	}
	if references.Len() != resolved.Len() {
		t.Fatal("expected length of arrays to be the same")
	}

	for i, r := range resolved {
		if r.Digest == "" {
			t.Fatal("expected Digest to be set")
		}
		if references[i].Digest != "" && r.Digest != references[i].Digest {
			t.Fatal("expected digest to have remained")
		}

		if r.Tag != references[i].Tag {
			t.Fatalf("expected Tag to be equal to original tag, found %s", r.Tag)
		}
		if r.Name != references[i].Name {
			t.Fatalf("expected Name to be %s, got %s", references[i].Name, r.Name)
		}
	}
}

func TestFindAlternateTags(t *testing.T) {
	latest := MustNewContainerImageReference("gcr.io/binx-io-public/paas-monitor:latest")
	tags, err := latest.FindAlternateTags()
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) == 0 {
		t.Fatalf("expected at least one alternate Tag, found 0")
	}
	t.Logf("found alternate tags for %s, %v", latest, tags)
}
