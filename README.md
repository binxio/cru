# gcru
Google Container Registry Updater - you give it a image reference, it will resolve it to an image digest. It can then find references to the container image in source code and update it in-place with the new digest.

## Image Digests
You should be using image digests to refer to container images. 

Instead of `gcr.io/my-project/my-app:v1.2`, use `gcr.io/my-project/my-app@sha256:9550c0b587e1e07fda5a7bd210a44d868f038944a86fe77673ea613d57d62ef9`

This is why:

 - An image digest is a content hash. When you pull a container using an image digest you will automatically verify the image.
 - Content hashes always point to the same image. This improves cacheability and helps to avoid broken builds. Tags are mutable. This is particularly true for the tag `latest`.

## Usage
Search for image references in the current directory
```
$: gcru list .
IMAGE                        NEEDS UPDATE  LATEST DIGEST
gcr.io/my-project/my-app     yes           sha256:70d23423bdb3e4e63255cf62747b5cbfce53210778ca2fc3a2544595a0fce3c6
gcr.io/my-project/my-worker  no            sha256:2a2df1d263e73f6a2cc16a9e4aefe8b44563b74d2f1dca067ba167da1198216c
``` 

Update all image references in the current directory to latest
```
$: gcru update .
Resolved digest gcr.io/my-project/my-app@sha256:70d23423bdb3e4e63255cf62747b5cbfce53210778ca2fc3a2544595a0fce3c6
Updated 1 file(s)
Resolved digest gcr.io/my-project/my-worker@sha256:2a2df1d263e73f6a2cc16a9e4aefe8b44563b74d2f1dca067ba167da1198216c
Updated 2 file(s)
``` 

Update a specific image reference, resolving the tag. This tool will *always* overwrite the current tag or digest in an image reference. 
```
$: gcru update -i gcr.io/my-project/my-app:v1.2 .
Resolved digest gcr.io/my-project/my-app@sha256:9550c0b587e1e07fda5a7bd210a44d868f038944a86fe77673ea613d57d62ef9
Updated 1 file(s)
``` 

Update all image references in files in the current directory ending with .tf
```
gcru update *.tf
```

Update files all image references in *.tf files recursively
```
find . -iname "*.tf" | xargs gcru update
```

## Use Cases
Keeping Kubernetes manifests, Terraform configuration files, or Dockerfiles up to date. This tool can also be used to practice a GitOps-style workflow.
