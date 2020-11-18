package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/utils/env"
	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/pkg/errors"
)

func NewServiceClient(service string, opts *clientconfig.ClientOpts) (*gophercloud.ServiceClient, error) {
	var cert string

	cloud, err := getCloud(opts)
	if err != nil {
		return nil, err
	}
	// Get the ca-cert-bundle key if there is a value for cacert in clouds.yaml
	if caPath := cloud.CACertFile; caPath != "" {
		caFile, err := ioutil.ReadFile(caPath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read clouds.yaml ca-cert from disk")
		}
		cert = string(bytes.TrimSpace(caFile))
	}

	if cert != "" && opts.HTTPClient == nil {
		certPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, errors.Wrap(err, "create system cert pool failed")
		}
		certPool.AppendCertsFromPEM([]byte(cert))
		client := http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: certPool,
				},
			},
		}
		opts.HTTPClient = &client
	}

	// Create a standard client for the given service
	return clientconfig.NewServiceClient(service, opts)
}

func getCloud(opts *clientconfig.ClientOpts) (*clientconfig.Cloud, error) {
	cloud := new(clientconfig.Cloud)

	// If no opts were passed in, create an empty ClientOpts.
	if opts == nil {
		opts = new(clientconfig.ClientOpts)
	}

	// Determine if a clouds.yaml entry should be retrieved.
	// Start by figuring out the cloud name.
	// First check if one was explicitly specified in opts.
	var cloudName string
	if opts.Cloud != "" {
		cloudName = opts.Cloud
	}

	// Next see if a cloud name was specified as an environment variable.
	envPrefix := "OS_"
	if opts.EnvPrefix != "" {
		envPrefix = opts.EnvPrefix
	}

	if v := env.Getenv(envPrefix + "CLOUD"); v != "" {
		cloudName = v
	}

	// If a cloud name was determined, try to look it up in clouds.yaml.
	if cloudName != "" {
		// Get the requested cloud.
		var err error
		cloud, err = clientconfig.GetCloudFromYAML(opts)
		if err != nil {
			return nil, err
		}
	}

	return cloud, nil
}
