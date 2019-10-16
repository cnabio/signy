# Signy

[![Build Status](https://dev.azure.com/engineerd-dev/signy/_apis/build/status/engineerd.signy?branchName=master)](https://dev.azure.com/engineerd-dev/signy/_build/latest?definitionId=5&branchName=master)

Signy is a tool for exercising the TUF specification in order to sign various cloud-native artifacts. It uses the Notary client libraries, and communicates with a Notary server.
It is an educational project with the purpose of implementing [the entire TUF workflow for signing content](https://github.com/theupdateframework/specification/blob/master/tuf-spec.md#5-detailed-workflows), and validate its correctness for multiple cloud-native artifact types (Helm charts, CNAB bundles, and others).

Currently, it implements signing and verifying for plain text and CNAB bundles.

## Using Signy

Operations:

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

Or, if your trust server is running in a different location, you can pass its URL and TLS CA:

```
$ ./bin/signy --tlscacert=<TLS CA> --server <URL of trust server> list hellosigny
0.1.0	dcd5b548984cfddee4dd9b467bd9c70606cd1d5ebbdd0dba0290ff147db24ea3
```

On the first push to a repository, it also generates the signing keys.
To avoid introducing the passphrases every time, set the following environment variables with the corresponding passphrases:

```
export SIGNY_ROOT_PASSPHRASE
export SIGNY_TARGETS_PASSPHRASE
export SIGNY_SNAPSHOT_PASSPHRASE
export SIGNY_DELEGATION_PASSPHRASE
```

- computing the SHA256 digest of a canonical CNAB bundle, pushing it to the trust server, then pushing the bundle using `cnab-to-oci`:

```
$ signy sign bundle.json docker.io/<user>/<repo>:<tag>
Root key found, using: d701ba005e6d217c7eb6cb56dbc6cf0bd81f41347927acbca1318131cc693fc9

Pushed trust data for docker.io/<user>/<repo>:<tag>: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
Starting to copy image cnab/helloworld:0.1.1...
Completed image cnab/helloworld:0.1.1 copy

Generated relocation map: bundle.ImageRelocationMap{"cnab/helloworld:0.1.1":"docker.io/radumatei/signed-cnab-bundle@sha256:a59a4e74d9cc89e4e75dfb2cc7ea5c108e4236ba6231b53081a9e2506d1197b6"}
Pushed successfully, with digest "sha256:086ef83113475d4582a7431b4b9bc98634d4f71ad1289cca45e661153fc9a46e"
```

- verifying the metadata in the trusted collection for a CNAB bundle against the bundle pushed to an OCI registry

```
$ signy verify docker.io/<user>/<repo>:<tag>
Pulled trust data for docker.io/<user>/<repo>:<tag>, with role targets - SHA256: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
Pulling bundle from registry: docker.io/<user>/<repo>:<tag>
Relocation map map[cnab/helloworld:0.1.1:radumatei/signed-cnab@sha256:a59a4e74d9cc89e4e75dfb2cc7ea5c108e4236ba6231b53081a9e2506d1197b6]

Computed SHA: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
The SHA sums are equal: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
```

- computing the SHA256 digest of a thick bundle, then pushing it to a trust sever

```
$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 sign helloworld-0.1.1.tgz --thick  localhost:5000/thick-bundle-signature:v1

Pushed trust data for localhost:5000/thick-bundle-signature:v1: cd205919129bff138a3402b4de5abbbc1d310ec982e83a780ffee1879adda678
```

- verifying the metadata for a local thick bundle

```
$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 verify --thick --local helloworld-0.1.1.tgz localhost:5000/thick-bundle-signature:v1

Pulled trust data for localhost:5000/thick-bundle-signature:v1, with role targets - SHA256: cd205919129bff138a3402b4de5abbbc1d310ec982e83a780ffee1879adda678
Computed SHA: cd205919129bff138a3402b4de5abbbc1d310ec982e83a780ffee1879adda678
The SHA sums are equal: cd205919129bff138a3402b4de5abbbc1d310ec982e83a780ffee1879adda678
```

### Using In-Toto

Notes:

- it assumes the In-Toto metadata is already generated, and expects the root layout, signing key for the layout, and the links directory as parameters to the push operation
- currently, only bundle attestation metadata is pushed to the trust server

```
$ signy intoto-sign bundle.json radumatei/tuf-intoto-metadata:v1
    --layout testdata/intoto/demo.layout.template
    --layout-key testdata/intoto/alice.pub
    --links testdata/intoto

Adding In-Toto layout and links metadata to TUF
Root key found, using: <root key>
Pushed trust data for <bundle-repo>: <SHA> to server https://notary.docker.io

Starting to copy image cnab/helloworld:0.1.1...
adding entry in relocation map: cnab/helloworld:0.1.1: docker.io/radumatei/tuf-intoto-metadata@sha256:a59a4e74d9cc89e4e75dfb2cc7ea5c108e4236ba6231b53081a9e2506d1197b6Completed image cnab/helloworld:0.1.1 copy

Generated relocation map: bundle.ImageRelocationMap{"cnab/helloworld:0.1.1":"docker.io/radumatei/tuf-intoto-metadata@sha256:a59a4e74d9cc89e4e75dfb2cc7ea5c108e4236ba6231b53081a9e2506d1197b6"}
Pushed successfully, with digest "sha256:086ef83113475d4582a7431b4b9bc98634d4f71ad1289cca45e661153fc9a46e"

$ signy intoto-verify radumatei/tuf-intoto-metadata:v1
    --image radumatei/in-toto-container:v1
	--target testdata/intoto/foo.tar.gz

Pulled trust data for radumatei/tuf-intoto-metadata:v1, with role targets - SHA256: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
Pulling bundle from registry: radumatei/tuf-intoto-metadata:v1
Relocation map map[cnab/helloworld:0.1.1:radumatei/tuf-intoto-metadata@sha256:a59a4e74d9cc89e4e75dfb2cc7ea5c108e4236ba6231b53081a9e2506d1197b6]

Computed SHA: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
The SHA sums are equal: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475

Writing In-Toto metadata files into /tmp/intoto-verification667685729
copying file /in-toto/layout.template in container for verification...
copying file /in-toto/key.pub in container for verification...
copying file in-toto/package.2f89b927.link in container for verification...
copying file in-toto/write-code.776a00e2.link in container for verification...
copying file in-toto/foo.tar.gz in container for verification...
Loading layout...
Loading layout key(s)...
Verifying layout signatures...
Verifying layout expiration...
Reading link metadata files...
Verifying link metadata signatures...
Verifying sublayouts...
Verifying alignment of reported commands...
Verifying command alignment for 'write-code.776a00e2.link'...
Verifying command alignment for 'package.2f89b927.link'...
Verifying threshold constraints...
Skipping threshold verification for step 'write-code' with threshold '1'...
Skipping threshold verification for step 'package' with threshold '1'...
Verifying Step rules...
Verifying material rules for 'write-code'...
Verifying product rules for 'write-code'...
Verifying 'ALLOW foo.py'...
Verifying material rules for 'package'...
Verifying 'MATCH foo.py WITH PRODUCTS FROM write-code'...
Verifying 'DISALLOW *'...
Verifying product rules for 'package'...
Verifying 'ALLOW foo.tar.gz'...
Verifying 'ALLOW foo.py'...
Executing Inspection commands...
Executing command for inspection 'untar'...
Running 'untar'...
Recording materials '.'...
Running command 'tar xfz foo.tar.gz'...
Recording products '.'...
Creating link metadata...
Verifying Inspection rules...
Verifying material rules for 'untar'...
Verifying 'MATCH foo.tar.gz WITH PRODUCTS FROM package'...
Verifying 'DISALLOW foo.tar.gz'...
Verifying product rules for 'untar'...
Verifying 'MATCH foo.py WITH PRODUCTS FROM write-code'...
Verifying 'DISALLOW foo.py'...
The software product passed all verification.
```

## Building from source

```
$ make bootstrap build
```
