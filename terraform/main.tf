resource "google_cloud_run_service" "container_reference_updater" {
  name     = "container-reference-updater"
  location = "us-central1"
  project  = data.google_client_config.current.project

  template {
    spec {
      container_concurrency = 1
      service_account_name  = google_service_account.container_reference_updater.email
      containers {
        image = "gcr.io/binx-io-public/cru:0.9.0"
        args = [
          "serve",
          "--repository",
          "https://source.developers.google.com/p/speeltuin-mvanholsteijn/r/scratch",
          "--branch",
          "main",
        ]
      }
    }
  }
  timeouts {
    create = "10m"
  }
}

resource "google_cloud_run_service_iam_binding" "container_reference_updater_invokers" {
  service = google_cloud_run_service.container_reference_updater.name
  location = google_cloud_run_service.container_reference_updater.location
  project = google_cloud_run_service.container_reference_updater.project
  role = "roles/run.invoker"
  members = [
    "allUsers"
  ]
}

resource "google_service_account" "container_reference_updater" {
  display_name = "Container image reference updater"
  account_id   = "container-reference-updater"
}

resource "google_project_iam_member" "cru_source_code_repository_writer" {
  member = format("serviceAccount:%s", google_service_account.container_reference_updater.email)
  role   = "roles/source.writer"
  project = data.google_client_config.current.project
}

data google_client_config current {}
