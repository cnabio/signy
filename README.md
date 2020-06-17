# Signy

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/cnabio/signy)

Signy is an experimental tool that implements the CNAB Security specification. It implements signing and verifying for CNAB bundles in [the canonical formats (thin and thick bundles)](https://github.com/deislabs/cnab-spec/blob/master/104-bundle-formats.md).

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
$ ./scripts/signy-sign.sh testdata/intoto/minimal/bundle.json localhost:5000/minimal:latest --in-toto --layout testdata/intoto/minimal/root.layout --links testdata/intoto/minimal/d374df2f6946233546bb4ca97dcee3a01fe07aaef11be1fb09abd37ceb4ecfb7/ --layout-key testdata/intoto/minimal/root.layout.pub
```

- verifying the signature of a thin bundle and running the in-toto verifications in a container:

```
$ ./scripts/signy-verify.sh localhost:5000/minimal:latest --in-toto
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
