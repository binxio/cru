# CRU
Container Reference Updater - updates container image references

Update kubernetes manifests, Terraform configuration files, or any other file with
the required container image reference.

If can work on a remote git repository, a local git workspace or a directory.

cru can also run as a service, to allow clients to update the image references in a
git repository, without requiring complete write access to the repository.


## Usage
```
  cru list   [--verbose] [--no-filename] [--repository=URL [--branch=BRANCH]] [PATH] ...
  cru update [--verbose] [--dry-run] [(--resolve-digest|--resolve-tag)] [--repository=URL [--branch=BRANCH] [--commit=MESSAGE]] (--all | --image-reference=REFERENCE ...) [PATH] ...
  cru serve  [--verbose] [--dry-run] [--port=PORT] --repository=URL --branch=BRANCH [PATH] ...
```

## Options
```
--dry-run           pretend to run the update, make no changes.
--verbose           show more output.

--no-filename       do not print the filename.

--all               replace all container image reference tags with "latest"
--image-reference=REFERENCE to update.
--resolve-digest    change the image reference tag to a reference of the digest of the image.
--resolve-tag       change the image reference tag to the first alternate tag of the reference.

--commit=MESSAGE    commit the changes with the specified message.
--repository=URL    to read and/or update.
--branch=BRANCH     to update.

--port=PORT         to listen on, defaults to 8080 or PORT environment variable.
```

## Examples
In the following paragraphs you will find a number of examples of the use of cru.

### listing container image references
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

### updating specific container image references
Update all image references of eu.gcr.io/binxio/paas-monitor:v3.2.4-5-g49d6871 in the file ref/update\_test.go:
```
$ cru update --image-reference eu.gcr.io/binxio/paas-monitor:v3.2.4-5-g49d6871 ref/update_test.go
2020/06/23 13:19:11 INFO: updating ref/update_test.go
```
### updating specific container image reference with digest
Update a specific image reference with the exact digest, add `--resolve-digest`:
```
$ cru update --resolve-digest --image-reference gcr.io/binx-io-public/paas-monitor:v3.2.4-5-g49d6871 update_test.go
2020/06/23 13:29:53 resolving repository gcr.io/binx-io-public/paas-monitor tag v3.2.4-5-g49d6871 to digest sha256:39af528cdc113845360cefa8ac84c653e7d512e3e2fd2f1506fa5e02ee88bda0
2020/06/23 13:29:53 INFO: updating update_test.go
```
If you want the image tag back:
```
$ cru update --image-reference gcr.io/binx-io-public/paas-monitor:v3.2.4-5-g49d6871 ref/update_test.go
2020/06/23 13:31:53 INFO: updating ref/update_test.go
```

### update all to latest and pin to a specific image
If you want to snapshot all image references to the latest version, use --all --resolve-digest:
```
$ cru update --all --resolve-digest ref/resolve_test.go
2020/06/23 19:49:17 resolving repository mvanholsteijn/paas-monitor tag latest to digest sha256:673545ac2fc55ff8dd8ef734b928aa34e5f498ef5aed2ec4cdfc028efe7585f3
2020/06/23 19:49:17 INFO: updating ref/resolve_test.go
```

### update latest with a specific tag
If you want to update an image reference to the latest version, without knowing the alternate tag:
```
$ cru update --image-reference gcr.io/binx-io-public/paas-monitor:latest --resolve-tag ref/resolve_test.go
2020/07/18 22:55:32 resolving repository gcr.io/binx-io-public/paas-monitor tag 'latest' to 'v3.2.4-5-g49d6871'
2020/07/18 22:55:32 INFO: updating ref/resolve_test.go
```

### cru-as-a-service
To run cru as a service, type:

```shell-terminal
$ cru serve   \
  --repository git@github.com:mvanholsteijn/scratch.git \
  --branch master
2021/11/09 20:30:27 Listening on port 8080
```

To update references, type:
```shell-terminal
$ curl -sS -X POST -H 'content-type: application/json' http://localhost:8080  -d '
{
  "image-references": [ "mvanholsteijn/paas-monitor:3.0.1" ],
  "commit-message": "you can do this."
}'
```
The response will look like this:

```json
{
  "git-url": "git@github.com:mvanholsteijn/scratch.git",
  "files": [
    "/README.md"
  ],
  "commit-sha": "43025b999e3ccc30557543c0cd336696f9073d49"
}
```


## installation
To install you have a number of different options:
- download the binary from https://github.com/binxio/cru/releases
- install using `go install github.com/binx/cru@0.9.0
- or use the docker image: gcr.io/binx-io-public/cru:0.9.0

## Caveats
- cru is not context-aware: anything that looks like a container image references is updated.
- cru will ignore any references to unqualified official images, like docker:latest or nginx:3. To update the official docker image references, prefix them with docker.io/ or docker.io/library/.
- the time to find the alternate tag is proportional to the number of tags associated with the image.
