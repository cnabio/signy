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

- signing a new file and publishing to a trust server:

```
$ signy sign --tlscacert=$NOTARY_CA --server https://localhost:4443 --type plaintext Makefile signy-collection
You are about to create a new root signing key passphrase. This passphrase
will be used to protect the most sensitive key in your signing system. Please
choose a long, complex passphrase and be careful to keep the password and the
key file itself secure and backed up. It is highly recommended that you use a
password manager to generate the passphrase and keep it safe. There will be no
way to recover this key. You can find the key in your config directory.
Enter passphrase for new root key with ID 01b4b77: 
Repeat passphrase for new root key with ID 01b4b77: 
Enter passphrase for new targets key with ID 0dd3549: 
Repeat passphrase for new targets key with ID 0dd3549: 
Enter passphrase for new snapshot key with ID 6e4ee53: 
Repeat passphrase for new snapshot key with ID 6e4ee53: 
```

At this point, you can investigate the contents of the `~/.signy` directory:

```
signy$ tree ~/.signy/
/home/radu/.signy/
├── private
│   ├── 01b4b77b1609ae565b3f664283767e5e2cd277a9b43b348a16ec8399a0e54c4c.key
│   ├── 0dd35497653e22d8ba56e0fd9cbf9643da40248c1feb1b07d1dab3e7948b041d.key
│   └── 6e4ee537c17458b2db198db2ae0b637a71b309d79bdd14742921ea2c8b7bbddf.key
└── tuf
    └── signy-collection
        ├── changelist
        └── metadata
            ├── root.json
            ├── snapshot.json
            └── targets.json

5 directories, 6 files

signy$ cat ~/.signy/tuf/signy-collection/metadata/targets.json | jq
{
  "signed": {
    "_type": "Targets",
    "delegations": {
      "keys": {},
      "roles": []
    },
    "expires": "2022-06-09T18:28:42.990074455Z",
    "targets": {
      "signy-collection": {
        "hashes": {
          "sha256": "H8zTaGz9P67HlRSVina2xUyVJDGx2DPB23UIQYrxxhA=",
          "sha512": "nvoMSaNxJv5//prUrEBAuCo+KP+Zr6bJCc6uK8lZmGSMQP13ag3fg6qmhIiszKqmQAyWIJStB7QAEUeqiVL54A=="
        },
        "length": 1069
      }
    },
    "version": 2
  },
  "signatures": [
    {
      "keyid": "0dd35497653e22d8ba56e0fd9cbf9643da40248c1feb1b07d1dab3e7948b041d",
      "method": "ecdsa",
      "sig": "MYnMTOjc+/WXRWfsEHbbELrX88NFWFM0Id8AbBxrkC1Z0IFUGpvgal4lDSs1IxqTo2D0bzbO5vApA/rMdVmljQ=="
    }
  ]
}

$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 list signy-collection
signy-collection        1fccd3686cfd3faec79514958a76b6c54c952431b1d833c1db7508418af1c610

signy$ sha256sum Makefile 
1fccd3686cfd3faec79514958a76b6c54c952431b1d833c1db7508418af1c610  Makefile
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
$ signy sign --type cnab bundle.json docker.io/<user>/<repo>:<tag>
Root key found, using: d701ba005e6d217c7eb6cb56dbc6cf0bd81f41347927acbca1318131cc693fc9

Pushed trust data for docker.io/<user>/<repo>:<tag>: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
Starting to copy image cnab/helloworld:0.1.1...
Completed image cnab/helloworld:0.1.1 copy

Generated relocation map: bundle.ImageRelocationMap{"cnab/helloworld:0.1.1":"docker.io/radumatei/signed-cnab-bundle@sha256:a59a4e74d9cc89e4e75dfb2cc7ea5c108e4236ba6231b53081a9e2506d1197b6"}
Pushed successfully, with digest "sha256:086ef83113475d4582a7431b4b9bc98634d4f71ad1289cca45e661153fc9a46e"
```

- verifying the metadata in the trusted collection for a CNAB bundle against the bundle pushed to an OCI registry

```
$ signy verify --type cnab docker.io/<user>/<repo>:<tag>
Pulled trust data for docker.io/<user>/<repo>:<tag>, with role targets - SHA256: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
Pulling bundle from registry: docker.io/<user>/<repo>:<tag>
Relocation map map[cnab/helloworld:0.1.1:radumatei/signed-cnab@sha256:a59a4e74d9cc89e4e75dfb2cc7ea5c108e4236ba6231b53081a9e2506d1197b6]

Computed SHA: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
The SHA sums are equal: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
```

- verifying the metadata in the trusted collection for a CNAB bundle against the bundle pushed to an OCI registry and against a local file

```
$ signy verify --type cnab --local bundle.json docker.io/<user>/<repo>:<tag>
Pulled trust data for docker.io/<user>/<repo>:<tag>, with role targets - SHA256: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
Pulling bundle from registry: docker.io/<user>/<repo>:<tag>
Relocation map map[cnab/helloworld:0.1.1:radumatei/signed-cnab@sha256:a59a4e74d9cc89e4e75dfb2cc7ea5c108e4236ba6231b53081a9e2506d1197b6]

Computed SHA: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
Computed SHA: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
The SHA sums are equal: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
```

## Building from source

```
$ make bootstrap build
```
