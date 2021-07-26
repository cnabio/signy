module github.com/cnabio/signy

go 1.13

require (
	github.com/cnabio/cnab-go v0.8.2-beta1
	github.com/containerd/cgroups v1.0.1 // indirect
	github.com/containerd/containerd v1.4.8
	github.com/containerd/continuity v0.0.0-20190827140505-75bee3e2ccb6 // indirect
	github.com/containerd/go-runc v1.0.0 // indirect
	github.com/cyberphone/json-canonicalization v0.0.0-20210303052042-6bc126869bf4
	github.com/docker/cli v0.0.0-20191017083524-a8ff7f821017
	github.com/docker/cnab-to-oci v0.3.0-beta4
	github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible
	github.com/docker/docker v1.4.2-0.20191021213818-bebd8206285b
	github.com/docker/go v1.5.1-1
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/in-toto/in-toto-golang v0.0.0-20191106170227-857cd1cfa826
	github.com/oklog/ulid v1.3.1
	github.com/opencontainers/selinux v1.8.2 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.6.1
	github.com/theupdateframework/notary v0.6.1
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210514084401-e8d321eab015 // indirect
	google.golang.org/grpc v1.39.0 // indirect
	gotest.tools/v3 v3.0.3 // indirect
)

replace github.com/in-toto/in-toto-golang => github.com/radu-matei/in-toto-golang v0.0.0-20210426203218-225046ac7465

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
