# Signy

Signy is a tool for exercising the TUF specification in order to sign various cloud-native artifacts. It uses the Notary client libraries, and communicates with a Notary server.
It an educational project with the purpose of implementing [the entire TUF workflow for signing content](https://github.com/theupdateframework/specification/blob/master/tuf-spec.md#5-detailed-workflows), and validate its correctness for multiple cloud-native artifact types (Helm charts, CNAB bundles, and others).

# Using Signy

For now, you can only list all targets for a remote trusted collection:

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

# Building from source and using

```
$ make bootstrap build
```
