# gcru
Google Container Registry Updater - it finds references to gcr.io containers in files and can update the reference in-place to set the digest for you.

You should be using image digests to refer to container images. This is why:

 - An image digest is a content hash. When you pull a container using an image digest you will automatically verify the image.
 - Content hashes always point to the same image, unlike tags. This improves cacheability and helps to avoid broken builds. 
 - You want to use `latest` but your Infrastructure as Code provisioner will never know when to update.
 - You feel that tagging (and/or versioning) internal container images is way too much work.
 
Instead of `gcr.io/my-project/my-app:v1.2`, use `gcr.io/my-project/my-app@sha256:9550c0b587e1e07fda5a7bd210a44d868f038944a86fe77673ea613d57d62ef9`

## Usage
Search for image references in *.tf files in the current directory
```
$: gcru *.tf
gcr.io/my-project/my-app
gcr.io/my-project/my-worker
``` 

Update a specific image reference in the same set of files, resolving the tag.
```
$: gcru -u gcr.io/my-project/my-app:v1.2 *.tf
Resolved digest gcr.io/my-project/my-app@sha256:9550c0b587e1e07fda5a7bd210a44d868f038944a86fe77673ea613d57d62ef9
Updated 1 file(s)
``` 

Update all image references to latest
```
$: gcru -u *.tf 
Resolved digest gcr.io/my-project/my-app@sha256:70d23423bdb3e4e63255cf62747b5cbfce53210778ca2fc3a2544595a0fce3c6
Updated 1 file(s)
Resolved digest gcr.io/my-project/my-worker@sha256:2a2df1d263e73f6a2cc16a9e4aefe8b44563b74d2f1dca067ba167da1198216c
Updated 2 file(s)
``` 

Update files all image references in *.tf files recursively
```
find . -iname "*.tf" | xargs gcru -u
```

Update all files in the current directory
```
gcru -u
```

## Use Cases
Keeping Kubernetes manifests, Terraform configuration files, or Dockerfiles up to date. Can also be used to practice a GitOps-style workflow.

## Where is the Code? 
Sorry, this is still in development. 
