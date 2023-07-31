package ref

import (
	"bytes"
	"testing"
)

func TestRefParseWithDigest(t *testing.T) {
	ref := MustNewContainerImageReference(`gcr.io/binx-io-public/paas-monitor:v0.3.1@sha256:3b27ac95762ce1340a4824dc3cab2dc9d63f194f0899a0a9887402c1b1463f41`)
	if ref.Digest != "sha256:3b27ac95762ce1340a4824dc3cab2dc9d63f194f0899a0a9887402c1b1463f41" {
		t.Errorf("incorrect digest, %s", ref.Digest)
	}
}

func TestUpdateReferenceSimple(t *testing.T) {
	simple := []byte(`gcr.io/binx-io-public/paas-monitor:v0.3.1`)
	expect := []byte(`gcr.io/binx-io-public/paas-monitor:v0.3.2`)
	ref := MustNewContainerImageReference(`gcr.io/binx-io-public/paas-monitor:v0.3.2`)
	result, updated := UpdateReference(simple, *ref, "myfile.txt", false, true)
	if !updated {
		t.Errorf("expected the reference to be updated\n")
	}
	if bytes.Compare(expect, result) != 0 {
		t.Errorf("expected %s, got %s\n", string(expect), string(result))
	}
}

func TestUpdateDigestToTag(t *testing.T) {
	simple := []byte(`gcr.io/binx-io-public/paas-monitor@sha256:4c9eeab8adf54d893450f6199f52cf7bb39264750ee2a11018dd41acfe6aeaba`)
	expect := []byte(`gcr.io/binx-io-public/paas-monitor:latest`)
	ref := MustNewContainerImageReference(`gcr.io/binx-io-public/paas-monitor:latest`)
	result, updated := UpdateReference(simple, *ref, "myfile.txt", false, true)
	if !updated {
		t.Errorf("expected the reference to be updated\n")
	}
	if bytes.Compare(expect, result) != 0 {
		t.Errorf("expected %s, got %s\n", string(expect), string(result))
	}
}

func TestUpdateReferenceMultiple(t *testing.T) {
	input := []byte(`this is one gcr.io/binx-io-public/paas-monitor:v0.3.1 reference
and this is another mvanholsteijn/paas-monitor:3.1.0
and this is just a directory Name mvanholsteijn/paas-monitor, which should
not be changed.
And how about a digest like gcr.io/binx-io-public/paas-monitor@sha256:4c9eeab8adf54d893450f6199f52cf7bb39264750ee2a11018dd41acfe6aeaba?
does that work?
`)
	expect := []byte(`this is one gcr.io/binx-io-public/paas-monitor:v1.0.0 reference
and this is another mvanholsteijn/paas-monitor:3.1.0
and this is just a directory Name mvanholsteijn/paas-monitor, which should
not be changed.
And how about a digest like gcr.io/binx-io-public/paas-monitor:v1.0.0?
does that work?
`)
	references := []ContainerImageReference{*MustNewContainerImageReference(`gcr.io/binx-io-public/paas-monitor:v1.0.0`),
		*MustNewContainerImageReference(`mvanholsteijn/paas-monitor:3.1.0`)}

	result, updated := UpdateReferences(input, references, "myfile.txt", false, true)
	if !updated {
		t.Errorf("expected the references to be updated\n")
	}
	if bytes.Compare(expect, result) != 0 {
		t.Errorf("expected %s, got %s\n", string(expect), string(result))
	}
}

func TestUpdateReferenceRealLifeExample(t *testing.T) {
	input := []byte(`
resource "google_cloud_run_service" "app" {
  Name     = "app"
  location = var.region

  template {
    spec {
      service_account_name = google_service_account.packer-reaper.email
      containers {
        image = "gcr.io/binx-io-public/paas-monitor:v0.3.1"
      }
    }
  }
  timeouts {
    create = "10m"
  }
  depends_on = [google_project_service.run]
  project    = data.google_project.current.project_id
}
`)
	expect := []byte(`
resource "google_cloud_run_service" "app" {
  Name     = "app"
  location = var.region

  template {
    spec {
      service_account_name = google_service_account.packer-reaper.email
      containers {
        image = "gcr.io/binx-io-public/paas-monitor:v0.3.2"
      }
    }
  }
  timeouts {
    create = "10m"
  }
  depends_on = [google_project_service.run]
  project    = data.google_project.current.project_id
}
`)
	ref, _ := NewContainerImageReference(`gcr.io/binx-io-public/paas-monitor:v0.3.2`)
	result, updated := UpdateReference(input, *ref, "myfile.txt", false, true)
	if !updated {
		t.Errorf("expected the reference to be updated\n")
	}
	if bytes.Compare(expect, result) != 0 {
		t.Errorf("expected %s, got %s\n", string(expect), string(result))
	}
}

func TestUpdateMatchingTag(t *testing.T) {
	simple := []byte(`
		gcr.io/binx-io-public/paas-monitor:3.4.0
		gcr.io/binx-io-public/paas-monitor:3.5.0
	`)
	expect := []byte(`
		gcr.io/binx-io-public/paas-monitor:3.4.0
		gcr.io/binx-io-public/paas-monitor:3.5.0@sha256:6af765830476b70cdb41d0f05e34a267f0868f56811b31a9a4d8e13c40188063
	`)
	ref := MustNewContainerImageReference(`gcr.io/binx-io-public/paas-monitor:3.5.0@sha256:6af765830476b70cdb41d0f05e34a267f0868f56811b31a9a4d8e13c40188063`)
	result, updated := UpdateReference(simple, *ref, "myfile.txt", true, true)
	if !updated {
		t.Errorf("expected the reference to be updated\n")
	}
	if bytes.Compare(expect, result) != 0 {
		t.Errorf("expected %s, got %s\n", string(expect), string(result))
	}
}
