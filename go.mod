module github.com/cnabio/signy

go 1.13

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78
	github.com/Masterminds/semver v1.5.0
	github.com/Microsoft/go-winio v0.4.14
	github.com/Microsoft/hcsshim v0.8.6
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412
	github.com/beorn7/perks v1.0.1
	github.com/cnabio/cnab-go v0.8.2-beta1
	github.com/containerd/containerd v1.3.0
	github.com/containerd/continuity v0.0.0-20190827140505-75bee3e2ccb6
	github.com/containerd/fifo v0.0.0-20190816180239-bda0ff6ed73c
	github.com/containerd/ttrpc v0.0.0-20191028202541-4f1b8fe65a5c
	github.com/containerd/typeurl v0.0.0-20190911142611-5eb25027c9fd
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/cli v0.0.0-20191017083524-a8ff7f821017
	github.com/docker/cnab-to-oci v0.3.0-beta4
	github.com/docker/compose-on-kubernetes v0.4.24 // indirect
	github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible
	github.com/docker/docker v1.4.2-0.20191021213818-bebd8206285b
	github.com/docker/docker-credential-helpers v0.6.3
	github.com/docker/go v1.5.1-1
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c
	github.com/docker/go-metrics v0.0.1
	github.com/docker/go-units v0.4.0
	github.com/engineerd/in-toto-container v0.0.0-20200225021239-288a901d4234 // indirect
	github.com/gogo/googleapis v1.3.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.2
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/go-version v1.2.0
	github.com/in-toto/in-toto-golang v0.0.0-20191106170227-857cd1cfa826
	github.com/inconshreveable/mousetrap v1.0.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/miekg/pkcs11 v1.0.3
	github.com/morikuni/aec v1.0.0
	github.com/oklog/ulid v1.3.1
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/opencontainers/image-spec v1.0.1
	github.com/opencontainers/runc v0.1.1
	github.com/opencontainers/runtime-spec v1.0.1
	github.com/pkg/errors v0.8.1
	github.com/pmezard/go-difflib v1.0.0
	github.com/prometheus/client_golang v1.1.0
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4
	github.com/prometheus/common v0.7.0
	github.com/prometheus/procfs v0.0.5
	github.com/qri-io/jsonpointer v0.1.0
	github.com/qri-io/jsonschema v0.1.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.4.0
	github.com/syndtr/gocapability v0.0.0-20180916011248-d98352740cb2
	github.com/theupdateframework/notary v0.6.1
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/net v0.0.0-20201021035429-f5854403a974
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	golang.org/x/sys v0.0.0-20210426230700-d19ff857e887
	golang.org/x/text v0.3.3
	golang.org/x/tools v0.1.0 // indirect
	google.golang.org/genproto v0.0.0-20191009194640-548a555dbc03
	google.golang.org/grpc v1.24.0
	gopkg.in/yaml.v2 v2.2.4
	vbom.ml/util v0.0.0-20180919145318-efcd4e0f9787
)

replace github.com/in-toto/in-toto-golang => github.com/radu-matei/in-toto-golang v0.0.0-20210426203218-225046ac7465

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
