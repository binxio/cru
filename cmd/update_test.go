package cmd

import (
	"bytes"
	"testing"
)

func TestUpdateReferenceSimple(t *testing.T) {
	simple := []byte(`gcr.io/binx-io-public/paas-monitor:v0.3.1`)
	expect := []byte(`gcr.io/binx-io-public/paas-monitor:v0.3.2`)
	ref := MustNewContainerImageReference(`gcr.io/binx-io-public/paas-monitor:v0.3.2`)
	result, updated := updateReference(simple, *ref)
	if !updated {
		t.Errorf("expected the reference to be updated\n")
	}
	if bytes.Compare(expect, result) != 0 {
		t.Errorf("expected %s, got %s\n", string(expect), string(result))
	}
}

func TestUpdateReferenceMultiple(t *testing.T) {
	input := []byte(`this is one gcr.io/binx-io-public/paas-monitor:v0.3.1 reference
and this is another mvanholsteijn/paas-monitor:v0.1.0
and this is just a directory name mvanholsteijn/paas-monitor, which should
not be changed`)
	expect := []byte(`this is one gcr.io/binx-io-public/paas-monitor:v1.0.0 reference
and this is another mvanholsteijn/paas-monitor:v0.2.0-beta
and this is just a directory name mvanholsteijn/paas-monitor, which should
not be changed`)
	references := []ContainerImageReference{*MustNewContainerImageReference(`gcr.io/binx-io-public/paas-monitor:v1.0.0`),
		*MustNewContainerImageReference(`mvanholsteijn/paas-monitor:v0.2.0-beta`)}

	result, updated := updateReferences(input, references)
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
  name     = "app"
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
  name     = "app"
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
	result, updated := updateReference(input, *ref)
	if !updated {
		t.Errorf("expected the reference to be updated\n")
	}
	if bytes.Compare(expect, result) != 0 {
		t.Errorf("expected %s, got %s\n", string(expect), string(result))
	}
}
