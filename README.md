# CRU
Container Reference Updater - updates container image references 

Update kubernetes manifests, Terraform configuration files, or any other infrastructural 
with the latest container image reference. 

## Usage
```
cru list [--no-filename] [PATH] ...
cru update [--dry-run] [--resolve-digest] [--all | [--image-reference=REFERENCE] ...] [PATH] ...
cru -h | --help

```
## Options
```
 --no-filename        do not print the filename.
 --resolve-digest     change the image reference tag to a reference of the digest of the image.
 --image-reference=REFERENCE to update.
 --dry-run            pretend to run the update, make no changes.
 --all                replace all container image reference tags with `latest`
```


## Examples
Search for image references in the current directory
```
$: cru list .
README.md: eu.gcr.io/binxio/paas-monitor:v0.3.1
README.md: eu.gcr.io/binxio/paas-monitor:v0.3.2
README.md: eu.gcr.io/binxio/paas-monitor:v1.0.0
README.md: gcr.io/my-project/my-app:v1.2
README.md: gcr.io/my-project/my-app@sha256:70d23423bdb3e4e63255cf62747b5cbfce53210778ca2fc3a2544595a0fce3c6
README.md: gcr.io/my-project/my-app@sha256:9550c0b587e1e07fda5a7bd210a44d868f038944a86fe77673ea613d57d62ef9
README.md: gcr.io/my-project/my-worker@sha256:2a2df1d263e73f6a2cc16a9e4aefe8b44563b74d2f1dca067ba167da1198216c
list_test.go: gcr.io/binxio/paas-monitor:v0.3.1
update_test.go: gcr.io/binxio/paas-monitor:v0.3.1
update_test.go: gcr.io/binxio/paas-monitor:v0.3.2
update_test.go: gcr.io/binxio/paas-monitor:v1.0.0
``` 

Update all image references of eu.gcr.io/binxio/paas-monitor:v3.2.4-5-g49d6871 in the file update\_test.go:
```
$ cru update --image-reference eu.gcr.io/binxio/paas-monitor:v3.2.4-5-g49d6871 update_test.go
2020/06/23 13:19:11 INFO: updating update_test.go
``` 

Update a specific image reference with the exact digest, add `--resolve-digest`: 
```
$ cru update --resolve-digest --image-reference gcr.io/binx-io-public/paas-monitor:v3.2.4-5-g49d6871 update_test.go
2020/06/23 13:29:53 resolving repository gcr.io/binx-io-public/paas-monitor tag v3.2.4-5-g49d6871 to digest sha256:39af528cdc113845360cefa8ac84c653e7d512e3e2fd2f1506fa5e02ee88bda0
2020/06/23 13:29:53 INFO: updating update_test.go
``` 
If you want the image tag back:
```
$ cru update --image-reference gcr.io/binx-io-public/paas-monitor:v3.2.4-5-g49d6871 update_test.go
2020/06/23 13:31:53 INFO: updating update_test.go
``` 

If you want to snapshot all image references to the latest version, use --all --resolve-digest:
```
$ cru update --all --resolve-digest cmd/resolve_test.go
2020/06/23 19:49:17 resolving repository mvanholsteijn/paas-monitor tag latest to digest sha256:673545ac2fc55ff8dd8ef734b928aa34e5f498ef5aed2ec4cdfc028efe7585f3
2020/06/23 19:49:17 INFO: updating cmd/resolve_test.go 
``` 

## Caveats
- cru is not context-aware: anything that looks like a container image references is updated.
- cru will ignore any references to unqualified official images, like docker:latest or nginx:3. To update the official docker image references, prefix them with docker.io/ or docker.io/library/.
