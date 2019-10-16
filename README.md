# Signy

[![Build Status](https://dev.azure.com/engineerd-dev/signy/_apis/build/status/engineerd.signy?branchName=master)](https://dev.azure.com/engineerd-dev/signy/_build/latest?definitionId=5&branchName=master)

Signy is a tool for exercising the TUF and in-toto specifications in order to sign various cloud-native artifacts. It uses the Notary client libraries, and communicates with a Notary server.
It is an educational project with the purpose of implementing [the entire TUF workflow for signing content](https://github.com/theupdateframework/specification/blob/master/tuf-spec.md#5-detailed-workflows), and validate its correctness for [Cloud Native Application Bundles (CNAB)](https://github.com/deislabs/cnab-spec), and it is intended as a WIP reference implementation for its security specification.

It implements signing and verifying for CNAB bundles in [the canonical formats (thin and thick bundles)](https://github.com/deislabs/cnab-spec/blob/master/104-bundle-formats.md).

## Notes

- the CNAB security specification uses TUF as a protocol for distributing trust metadata about bundles. This implementation uses Notary, a Go implementation of the TUF specification.
- this project has been tested using the open source Notary and Docker distribution.
- currently, the in-toto signing key for the root layout is passed in the TUF `custom` object. This invalidates the security model, and the priority is to move the distribution of that key out of bound (possibly using a TUF signing key - targets, for example).
- if pushing in-toto metadata, this tool assumes the in-toto metadata has already been generated using a different workflow.
- authentication currently has some transient issues. For now, it is best to use a local registry and trust server (see instructions below).

## Building Signy

```
$ cd $GOPATH/src/github.com
$ mkdir engineerd && cd engineerd && git clone https://github.com/engineerd/signy && cd signy
$ make bootstrap build
$ mv bin/signy $GOPATH/bin
```

## Using Signy

- Docker Hub (https://index.docker.io) and Docker Notary (https://notary.docker.io) can be used to push bundles and trust metadata, but current recommended way to test Signy is to run a registry and trust server locally.

- running Docker Distribution:

```
$ docker run -it -d -p 5000:5000 registry
```

- running Notary:

```
$ cd $GOPATH/src/github.com && mkdir theupdateframework && cd theupdateframework && git clone https://github.com/theupdateframework/notary && cd notary && docker-compose up -d
$ export NOTARY_CA=$GOPATH/src/github.com/theupdateframework/notary/cmd/notary/root-ca.crt
```

On the first push to a repository, Signy generates the signing keys (using Notary).
To avoid introducing the passphrases every time, set the following environment variables with the corresponding passphrases:

```
$ export SIGNY_ROOT_PASSPHRASE=PassPhrase#123
$ export SIGNY_TARGETS_PASSPHRASE=PassPhrase#123
$ export SIGNY_SNAPSHOT_PASSPHRASE=PassPhrase#123
$ export SIGNY_DELEGATION_PASSPHRASE=PassPhrase#123
```

At this point, Signy can be used by passing the Notary CA and URL to the trust server:

```
$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443
```

### Operations:

- listing the targets for a trusted collection:

```
$ signy list docker.io/library/alpine

3.5     66952b313e51c3bd1987d7c4ddf5dba9bc0fb6e524eed2448fa660246b3e76ec
3.8     04696b491e0cc3c58a75bace8941c14c924b9f313b03ce5029ebbc040ed9dcd9
3.2     e9a2035f9d0d7cee1cdd445f5bfa0c5c646455ee26f14565dce23cf2d2de7570
3.6     66790a2b79e1ea3e1dabac43990c54aca5d1ddf268d9a5a0285e4167c8b24475
3.10    6a92cd1fcdc8d8cdec60f33dda4db2cb1fcdcacf3410a8e05b3741f44a9b5998
3.9.4   7746df395af22f04212cd25a92c1d6dbc5a06a0ca9579a229ef43008d4d1302a
```

- computing the SHA256 digest of a canonical CNAB bundle, pushing it to the trust server, then pushing the bundle using `cnab-to-oci`:

```
$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 sign testdata/cnab/bundle.json localhost:5000/thin-bundle:v1
INFO[0000] Pushed trust data for localhost:5000/thin-bundle:v1: c7e92bd51f059d60b15ad456edf194648997d739f60799b37e08edafd88a81b5
INFO[0000] Starting to copy image cnab/helloworld:0.1.1
INFO[0002] Completed image cnab/helloworld:0.1.1 copy
INFO[0002] Generated relocation map: relocation.ImageRelocationMap{"cnab/helloworld:0.1.1":"localhost:5000/thin-bundle@sha256:a59a4e74d9cc89e4e75dfb2cc7ea5c108e4236ba6231b53081a9e2506d1197b6"}
INFO[0002] Pushed successfully, with digest "sha256:b4936e42304c184bafc9b06dde9ea1f979129e09a021a8f40abc07f736de9268"
```

- verifying the metadata in the trusted collection for a CNAB bundle against the bundle pushed to an OCI registry

```
$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 verify localhost:5000/thin-bundle:v1
INFO[0000] Pulled trust data for localhost:5000/thin-bundle:v1, with role targets - SHA256: c7e92bd51f059d60b15ad456edf194648997d739f60799b37e08edafd88a81b5
INFO[0000] Pulling bundle from registry: localhost:5000/thin-bundle:v1
INFO[0000] Computed SHA: c7e92bd51f059d60b15ad456edf194648997d739f60799b37e08edafd88a81b5
INFO[0000] The SHA sums are equal: c7e92bd51f059d60b15ad456edf194648997d739f60799b37e08edafd88a81b5
```

- computing the SHA256 digest of a thick bundle, then pushing it to a trust sever

```
$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 sign --thick testdata/cnab/helloworld-0.1.1.tgz localhost:5000/thick-bundle:v1
INFO[0000] Pushed trust data for localhost:5000/thick-bundle:v1: 540cc4dc213548ebbdffb2ab0ef58729e089d1887edbcde6eeca851de624da70
```

- verifying the metadata for a local thick bundle

```
$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 verify --thick --local testdata/cnab/helloworld-0.1.1.tgz localhost:5000/thick-bundle:v1
INFO[0000] Pulled trust data for localhost:5000/thick-bundle:v1, with role targets - SHA256: 540cc4dc213548ebbdffb2ab0ef58729e089d1887edbcde6eeca851de624da70
INFO[0000] Computed SHA: 540cc4dc213548ebbdffb2ab0ef58729e089d1887edbcde6eeca851de624da70
INFO[0000] The SHA sums are equal: 540cc4dc213548ebbdffb2ab0ef58729e089d1887edbcde6eeca851de624da70
```

### Using In-Toto

- add in-toto metadata when signing a thin bundle:

```
$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 sign testdata/cnab/bundle.json localhost:5000/thin-intoto:v2 --in-toto --layout testdata/intoto/demo.layout.template --links testdata/intoto --layout-key testdata/intoto/alice.pub
INFO[0000] Adding In-Toto layout and links metadata to TUF
INFO[0000] Pushed trust data for localhost:5000/thin-intoto:v2: c7e92bd51f059d60b15ad456edf194648997d739f60799b37e08edafd88a81b5
INFO[0000] Starting to copy image cnab/helloworld:0.1.1
INFO[0001] Completed image cnab/helloworld:0.1.1 copy
INFO[0001] Generated relocation map: relocation.ImageRelocationMap{"cnab/helloworld:0.1.1":"localhost:5000/thin-intoto@sha256:a59a4e74d9cc89e4e75dfb2cc7ea5c108e4236ba6231b53081a9e2506d1197b6"}
INFO[0001] Pushed successfully, with digest "sha256:b4936e42304c184bafc9b06dde9ea1f979129e09a021a8f40abc07f736de9268"
```

- verifying the signature of a thin bundle and running the in-toto verifications in a container:

```
$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 verify localhost:5000/thin-intoto:v2 --in-toto --target testdata/intoto/foo.tar.gz
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
$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 sign testdata/cnab/helloworld-0.1.1.tgz --thick  localhost:5000/thick-bundle-signature:v2 --in-toto --layout testdata/intoto/demo.layout.template --links testdata/intoto --layout-key testdata/intoto/alice.pub
INFO[0000] Adding In-Toto layout and links metadata to TUF
INFO[0000] Pushed trust data for localhost:5000/thick-bundle-signature:v2: 540cc4dc213548ebbdffb2ab0ef58729e089d1887edbcde6eeca851de624da70

$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 verify localhost:5000/thick-bundle-signature:v2 --thick --local testdata/cnab/helloworld-0.1.1.tgz --in-toto --target testdata/intoto/foo.tar.gz
```

Notes:

- see current limitations about the in-toto signing key of the root layout
- the `--target` currently passed is because the in-toto verification used as example needs to validate that file. In a real scenario, the verification would perform operations on the CNAB bundle. (Help needed to create a real-world in-toto layout)

## Contributing

This project welcomes all contributions. See the issue queue for existing issues, and make sure to also check the CNAB Security specification.
