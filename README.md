# Signy

[![Build Status](https://dev.azure.com/engineerd-dev/signy/_apis/build/status/engineerd.signy?branchName=master)](https://dev.azure.com/engineerd-dev/signy/_build/latest?definitionId=5&branchName=master)

Signy is a tool for exercising the TUF specification in order to sign various cloud-native artifacts. It uses the Notary client libraries, and communicates with a Notary server.
It an educational project with the purpose of implementing [the entire TUF workflow for signing content](https://github.com/theupdateframework/specification/blob/master/tuf-spec.md#5-detailed-workflows), and validate its correctness for multiple cloud-native artifact types (Helm charts, CNAB bundles, and others).

## Using Signy

Operations:

- listing the targets for a trusted collection:

```
$ signy list docker.io/library/alpine
2.6	9ace551613070689a12857d62c30ef0daa9a376107ec0fff0e34786cedb3399b
3.9.3	28ef97b8686a0b5399129e9b763d5b7e5ff03576aa5580d6f4182a49c5fe1913
latest	769fddc7cc2f0a1c35abb2f91432e8beecf83916c421420e6a6da9f8975464b6
2.7	9f08005dff552038f0ad2f46b8e65ff3d25641747d3912e3ea8da6785046561a
3.9.4	769fddc7cc2f0a1c35abb2f91432e8beecf83916c421420e6a6da9f8975464b6
3.1	4dfc68bc95af5c1beb5e307133ce91546874dcd0d880736b25ddbe6f483c65b4
3.5	66952b313e51c3bd1987d7c4ddf5dba9bc0fb6e524eed2448fa660246b3e76ec
integ-test-base	3952dc48dcc4136ccdde37fbef7e250346538a55a0366e3fccc683336377e372
3.4	b733d4a32c4da6a00a84df2ca32791bb03df95400243648d8c539e7b4cce329c
3.9	769fddc7cc2f0a1c35abb2f91432e8beecf83916c421420e6a6da9f8975464b6
3.3	6bff6f65597a69246f79233ef37ff0dc50d97eaecbabbe4f8a885bf358be1ecf
3.8	ea47a59a33f41270c02c8c7764e581787cf5b734ab10d27e876e62369a864459
3.6	66790a2b79e1ea3e1dabac43990c54aca5d1ddf268d9a5a0285e4167c8b24475
3.7	02c076fdbe7d116860d9fb10f856ed6753a50deecb04c65443e2c6388d97ee35
3.7.3	02c076fdbe7d116860d9fb10f856ed6753a50deecb04c65443e2c6388d97ee35
20190508	db9c935c5445f75cace46d0418fac19d0b70b1723193e3b47d0d06bcddd05272
20190408	8b6b8c0f71e83cdbf888169bdd9b89f028cba03abff05c50246a191fec31b35a
3.2	e9a2035f9d0d7cee1cdd445f5bfa0c5c646455ee26f14565dce23cf2d2de7570
3.6.5	66790a2b79e1ea3e1dabac43990c54aca5d1ddf268d9a5a0285e4167c8b24475
3.8.4	ea47a59a33f41270c02c8c7764e581787cf5b734ab10d27e876e62369a864459
20190228	6199d795f07e4520fa0169efd5779dcf399cbfd33c73e15b482fcd21c42e1750
3.9.2	644fcb1a676b5165371437feaa922943aaf7afcfa8bfee4472f6860aad1ef2a0
edge	db9c935c5445f75cace46d0418fac19d0b70b1723193e3b47d0d06bcddd05272
```

Or, if your trust server is running in a different location, you can pass its URL and TLS CA:

```
$ ./bin/signy --tlscacert=<TLS CA> --server <URL of trust server> list hellosigny
0.1.0	dcd5b548984cfddee4dd9b467bd9c70606cd1d5ebbdd0dba0290ff147db24ea3
```

- signing a new file and publishing to a trust server:

```
$ ./bin/signy --tlscacert=$NOTARY_CA --server https://localhost:4443 sign Makefile signy-collection
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
radu:signy$ ./bin/signy --tlscacert=$NOTARY_CA --server https://localhost:4443 list signy-collection
signy-collection        1fccd3686cfd3faec79514958a76b6c54c952431b1d833c1db7508418af1c610

signy$ sha256sum Makefile 
1fccd3686cfd3faec79514958a76b6c54c952431b1d833c1db7508418af1c610  Makefile

```

## Building from source

```
$ make bootstrap build
```
