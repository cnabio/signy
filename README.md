# Signy

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/cnabio/signy)

Signy is an experimental tool that implements the CNAB Security specification. It implements signing and verifying for CNAB bundles in [the canonical formats (thin and thick bundles)](https://github.com/deislabs/cnab-spec/blob/master/104-bundle-formats.md). As an added feature, it also supports pushing (signing) and pulling (verifying) of container images to a registry alongside it's in-toto metadata, which `docker pull` and `docker push` is unable to do.

## Notes

- the CNAB security specification uses TUF as a protocol for distributing trust metadata about bundles. This implementation uses Notary, a Go implementation of the TUF specification.
- this project has been tested using the open source Notary and Docker distribution.
- currently, the in-toto signing key for the root layout is passed in the TUF `custom` object. This invalidates the security model, and the priority is to move the distribution of that key out of bound (possibly using a TUF signing key - targets, for example).
- if pushing in-toto metadata, this tool assumes the in-toto metadata has already been generated using a different workflow.
- authentication currently has some transient issues. For now, it is best to use a local registry and trust server (see instructions below).

## Building Signy

```bash
$ cd $GOPATH/src/github.com
$ mkdir cnabio && cd cnabio && git clone https://github.com/cnabio/signy && cd signy
$ make bootstrap build
$ mv bin/signy $GOPATH/bin
```

## Using Signy

### Setting up

- Run local Docker Distribution and Notary services:

```bash
# Setup Docker Distribution and Notary.
$ ./scripts/bootstrap.sh
# Start Docker Distribution and Notary.
$ ./scripts/signy-start.sh
```

- Before running Signy, test pushing and pulling from local registry and Notary server:

```bash
# Push a signed hello-world image.
$ ./scripts/docker-push.sh
# Pull the signed hello-world image.
$ ./scripts/docker-pull.sh
```

At this point, Signy can be used by passing the Notary CA and URL to the trust server:

```
$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443
```

### Common operations

- Computing the SHA256 digest of a canonical CNAB bundle, pushing it to the trust server, then pushing the bundle using `cnab-to-oci`:

```bash
$ ./scripts/signy-sign.sh testdata/cnab/bundle.json localhost:5000/cnab/thin-bundle:v1
INFO[0000] Starting to copy image cnab/helloworld:0.1.1
INFO[0000] Completed image cnab/helloworld:0.1.1 copy
INFO[0000] Generated relocation map: relocation.ImageRelocationMap{"cnab/helloworld:0.1.1":"localhost:5000/cnab/thin-bundle@sha256:a59a4e74d9cc89e4e75dfb2cc7ea5c108e4236ba6231b53081a9e2506d1197b6"}
INFO[0000] Pushed successfully, with digest "sha256:b4936e42304c184bafc9b06dde9ea1f979129e09a021a8f40abc07f736de9268"
INFO[0000] Pushed trust data for localhost:5000/cnab/thin-bundle:v1: c7e92bd51f059d60b15ad456edf194648997d739f60799b37e08edafd88a81b5
```

- Verifying the metadata in the trusted collection for a CNAB bundle against the bundle pushed to an OCI registry

```
$ ./scripts/signy-verify.sh localhost:5000/cnab/thin-bundle:v1
INFO[0000] Pulled trust data for localhost:5000/thin-bundle:v1, with role targets - SHA256: c7e92bd51f059d60b15ad456edf194648997d739f60799b37e08edafd88a81b5
INFO[0000] Pulling bundle from registry: localhost:5000/thin-bundle:v1
INFO[0000] Computed SHA: c7e92bd51f059d60b15ad456edf194648997d739f60799b37e08edafd88a81b5
INFO[0000] The SHA sums are equal: c7e92bd51f059d60b15ad456edf194648997d739f60799b37e08edafd88a81b5
```

- Listing the targets for a trusted collection:

```bash
$ ./scripts/signy-list.sh
0.1.1	d9dfd104723ea5b037000931a876e98e5e0bf492d665436d123d0dfc7c40c8e8
```


- Computing the SHA256 digest of a thick bundle, then pushing it to a trust sever

```
$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 sign --thick testdata/cnab/helloworld-0.1.1.tgz localhost:5000/thick-bundle:v1
INFO[0000] Pushed trust data for localhost:5000/thick-bundle:v1: 540cc4dc213548ebbdffb2ab0ef58729e089d1887edbcde6eeca851de624da70
```

- Verifying the metadata for a local thick bundle

```
$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 verify --thick --local testdata/cnab/helloworld-0.1.1.tgz localhost:5000/thick-bundle:v1
INFO[0000] Pulled trust data for localhost:5000/thick-bundle:v1, with role targets - SHA256: 540cc4dc213548ebbdffb2ab0ef58729e089d1887edbcde6eeca851de624da70
INFO[0000] Computed SHA: 540cc4dc213548ebbdffb2ab0ef58729e089d1887edbcde6eeca851de624da70
INFO[0000] The SHA sums are equal: 540cc4dc213548ebbdffb2ab0ef58729e089d1887edbcde6eeca851de624da70
```

### Using in-toto

- Add in-toto metadata when signing a thin bundle:

```
$ ./scripts/signy-sign.sh testdata/cnab/bundle.json localhost:5000/thin-intoto:v2 --in-toto --layout testdata/intoto/root.layout --links testdata/intoto --layout-key testdata/intoto/alice.pub
INFO[0000] Adding In-Toto layout and links metadata to TUF
INFO[0000] Pushed trust data for localhost:5000/thin-intoto:v2: c7e92bd51f059d60b15ad456edf194648997d739f60799b37e08edafd88a81b5
INFO[0000] Starting to copy image cnab/helloworld:0.1.1
INFO[0001] Completed image cnab/helloworld:0.1.1 copy
INFO[0001] Generated relocation map: relocation.ImageRelocationMap{"cnab/helloworld:0.1.1":"localhost:5000/thin-intoto@sha256:a59a4e74d9cc89e4e75dfb2cc7ea5c108e4236ba6231b53081a9e2506d1197b6"}
INFO[0001] Pushed successfully, with digest "sha256:b4936e42304c184bafc9b06dde9ea1f979129e09a021a8f40abc07f736de9268"
```

- verifying the signature of a thin bundle and running the in-toto verifications in a container:

```
$ ./scripts/signy-verify.sh localhost:5000/thin-intoto:v2 --in-toto
INFO[0000] Pulled trust data for localhost:5000/thin-intoto:v2, with role targets - SHA256: c7e92bd51f059d60b15ad456edf194648997d739f60799b37e08edafd88a81b5
INFO[0000] Pulling bundle from registry: localhost:5000/thin-intoto:v2
INFO[0000] Computed SHA: c7e92bd51f059d60b15ad456edf194648997d739f60799b37e08edafd88a81b5
INFO[0000] The SHA sums are equal: c7e92bd51f059d60b15ad456edf194648997d739f60799b37e08edafd88a81b5
INFO[0000] Writing In-Toto metadata files into /tmp/intoto-verification169227773
INFO[0000] copying file /in-toto/layout.template in container for verification...
INFO[0000] copying file /in-toto/key.pub in container for verification...
INFO[0000] copying file in-toto/package.2f89b927.link in container for verification...
INFO[0000] copying file in-toto/write-code.776a00e2.link in container for verification...
INFO[0000] copying file in-toto/foo.tar.gz in container for verification...
INFO[0000] Loading layout...
INFO[0000] Loading layout key(s)...
INFO[0000] Verifying layout signatures...
INFO[0001] Verifying layout expiration...
INFO[0001] Reading link metadata files...
INFO[0001] Verifying link metadata signatures...
INFO[0001] Verifying sublayouts...
INFO[0001] Verifying alignment of reported commands...
INFO[0001] Verifying command alignment for 'write-code.776a00e2.link'...
INFO[0001] Verifying command alignment for 'package.2f89b927.link'...
INFO[0001] Verifying threshold constraints...
INFO[0001] Skipping threshold verification for step 'write-code' with threshold '1'...
INFO[0001] Skipping threshold verification for step 'package' with threshold '1'...
INFO[0001] Verifying Step rules...
INFO[0001] Verifying material rules for 'write-code'...
INFO[0001] Verifying product rules for 'write-code'...
INFO[0001] Verifying 'ALLOW foo.py'...
INFO[0001] Verifying material rules for 'package'...
INFO[0001] Verifying 'MATCH foo.py WITH PRODUCTS FROM write-code'...
INFO[0001] Verifying 'DISALLOW *'...
INFO[0001] Verifying product rules for 'package'...
INFO[0001] Verifying 'ALLOW foo.tar.gz'...
INFO[0001] Verifying 'ALLOW foo.py'...
INFO[0001] Executing Inspection commands...
INFO[0001] Executing command for inspection 'untar'...
INFO[0001] Running 'untar'...
INFO[0001] Recording materials '.'...
INFO[0001] Running command 'tar xfz foo.tar.gz'...
INFO[0001] Recording products '.'...
INFO[0001] Creating link metadata...
INFO[0001] Verifying Inspection rules...
INFO[0001] Verifying material rules for 'untar'...
INFO[0001] Verifying 'MATCH foo.tar.gz WITH PRODUCTS FROM package'...
INFO[0001] Verifying 'DISALLOW foo.tar.gz'...
INFO[0001] Verifying product rules for 'untar'...
INFO[0001] Verifying 'MATCH foo.py WITH PRODUCTS FROM write-code'...
INFO[0001] Verifying 'DISALLOW foo.py'...
INFO[0001] The software product passed all verification.
```

- similarly for a thick bundle:

```
$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 sign testdata/cnab/helloworld-0.1.1.tgz --thick  localhost:5000/thick-bundle-signature:v2 --in-toto --layout testdata/intoto/root.layout --links testdata/intoto --layout-key testdata/intoto/alice.pub
INFO[0000] Adding In-Toto layout and links metadata to TUF
INFO[0000] Pushed trust data for localhost:5000/thick-bundle-signature:v2: 540cc4dc213548ebbdffb2ab0ef58729e089d1887edbcde6eeca851de624da70

$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 verify localhost:5000/thick-bundle-signature:v2 --thick --local testdata/cnab/helloworld-0.1.1.tgz --in-toto
```

Notes:

- see current limitations about the in-toto signing key of the root layout

### To sign container images and put the info in TUF alongside it's in-toto metadata

`signy --tlscacert root-ca.crt push -i [image]`

This command is nearly identical to the docker CLI command `docker push` when the environment variable `DOCKER_CONTENT_TRUST=1` and `DOCKER_CONTENT_TRUST_SERVER=[server:4443]` are set. In addition to signing the digest, we additionally push the in-toto metadata to the trust server just like `signy sign` does.

To pull and image and verify it's digest SHA:

`signy --tlscacert root-ca.crt pull -i [image]`

This will pull the image from the registry, verify it's digest against what is stored in notary, and verify it's in-toto metadata that was pulled down from TUF.

### Tearing down

- Stop all services:

```bash
./scripts/stop.sh
```

### Tips

On the first push to a repository, Signy generates the signing keys (using Notary). To avoid introducing the passphrases every time, set the following environment variables with the corresponding passphrases:

```
$ export SIGNY_ROOT_PASSPHRASE=PassPhrase#123
$ export SIGNY_TARGETS_PASSPHRASE=PassPhrase#123
$ export SIGNY_RELEASES_PASSPHRASE=PassPhrase#123
```

## Contributing

This project welcomes all contributions. See the issue queue for existing issues, and make sure to also check the CNAB Security specification.
