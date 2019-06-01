package trust

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/auth/challenge"
	"github.com/docker/distribution/registry/client/transport"
)

func makeTransport(server, reference, tlsCaCert string) (http.RoundTripper, error) {
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

	authTransport := transport.NewTransport(base, modifiers...)
	pingClient := &http.Client{
		Transport: authTransport,
		Timeout:   5 * time.Second,
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
	tokenHandler := auth.NewTokenHandler(base, nil, reference, "pull")
	modifiers = append(modifiers, auth.NewAuthorizer(challengeManager, tokenHandler, auth.NewBasicHandler(nil)))

	return transport.NewTransport(base, modifiers...), nil
}

func ensureTrustDir(trustDir string) error {
	return os.MkdirAll(trustDir, 0700)
}
