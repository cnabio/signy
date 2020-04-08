// Most of the helper functions are adapted from github.com/theupdateframework/notary
//
// Figure out the proper way of making sure we are respecting the licensing from Notary
// While we are also vendoring Notary directly (see LICENSE in vendor/github.com/theupdateframework/notary/LICENSE),
// copying unexported functions could fall under different licensing, so we need to make sure.

package tuf

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/docker/cli/cli/config"
	configtypes "github.com/docker/cli/cli/config/types"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/auth/challenge"
	"github.com/docker/distribution/registry/client/transport"
	log "github.com/sirupsen/logrus"
)

const (
	// DockerNotaryServer is the default Notary server associated with Docker Hub
	DockerNotaryServer = "https://notary.docker.io"

	defaultIndexServer = "https://index.docker.io/v1/"
)

func makeTransport(server, gun, tlsCaCert, timeout string) (http.RoundTripper, error) {
	modifiers := []transport.RequestModifier{
		transport.NewHeaderRequestModifier(http.Header{
			"User-Agent": []string{"signy"},
		}),
	}

	base := http.DefaultTransport
	if tlsCaCert != "" {
		caCert, err := ioutil.ReadFile(tlsCaCert)
		if err != nil {
			return nil, fmt.Errorf("cannot read cert file: %v", err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		base = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		}
	}

	t, err := time.ParseDuration(timeout)
	if err != nil {
		return nil, err
	}

	authTransport := transport.NewTransport(base, modifiers...)
	pingClient := &http.Client{
		Transport: authTransport,
		Timeout:   t * time.Second,
	}
	req, err := http.NewRequest("GET", server+"/v2/", nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create HTTP request: %v", err)
	}

	challengeManager := challenge.NewSimpleManager()
	resp, err := pingClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot get response from ping client: %v", err)
	}
	defer resp.Body.Close()
	if err := challengeManager.AddResponse(resp); err != nil {
		return nil, fmt.Errorf("cannot add response to challenge manager: %v", err)
	}

	defaultAuth, err := getAuth(server)
	if err != nil {
		log.Debug(fmt.Errorf("cannot get default credentials: %v", err))
	} else {
		creds := simpleCredentialStore{auth: defaultAuth}
		tokenHandler := auth.NewTokenHandler(base, creds, gun, "push", "pull")
		modifiers = append(modifiers, auth.NewAuthorizer(challengeManager, tokenHandler))
	}

	return transport.NewTransport(base, modifiers...), nil
}

func getAuth(server string) (configtypes.AuthConfig, error) {
	s, err := url.Parse(server)
	if err != nil {
		return configtypes.AuthConfig{}, fmt.Errorf("cannot parse trust server URL: %v", err)
	}

	cfg, err := config.Load(DefaultDockerCfgDir())
	if err != nil {
		return configtypes.AuthConfig{}, err
	}

	auth, ok := cfg.AuthConfigs[s.Hostname()]
	if !ok {
		if s.Hostname() == DockerNotaryServer {
			return cfg.AuthConfigs[defaultIndexServer], nil
		}
		return configtypes.AuthConfig{}, fmt.Errorf("authentication not found for trust server %v", server)
	}

	return auth, nil
}

type simpleCredentialStore struct {
	auth configtypes.AuthConfig
}

func (scs simpleCredentialStore) Basic(u *url.URL) (string, string) {
	return scs.auth.Username, scs.auth.Password
}

func (scs simpleCredentialStore) RefreshToken(u *url.URL, service string) string {
	return scs.auth.IdentityToken
}

func (scs simpleCredentialStore) SetRefreshToken(*url.URL, string, string) {
}
