package kafka

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
)

type Config struct {
	ClusterConfig                         *sarama.Config
	BrokerAddresses                       []string
	Topic                                 string
	GroupID                               string
	MaxInFlight                           int64
	CertPEMBlock, KeyPEMBlock, CaPEMCerts []byte
}

//nolint
func (c *Config) Connect() (sarama.ConsumerGroup, error) {
	// check if keys have been set to use tls.config for secure connection, otherwise likely means running locally
	if len(c.CertPEMBlock) > 0 || len(c.CaPEMCerts) > 0 || len(c.KeyPEMBlock) > 0 {
		tlsCfg, err := c.tlsConfig()
		if err != nil {
			return nil, errors.Wrap(err, "error setting up TLS")
		}

		c.ClusterConfig.Net.TLS.Enable = true
		c.ClusterConfig.Net.TLS.Config = tlsCfg
	}
	group, err := sarama.NewConsumerGroup(c.BrokerAddresses, c.GroupID, c.ClusterConfig)
	if err != nil {
		return nil, errors.Wrap(err, "error starting kafka consumer group")
	}

	return group, nil
}

//nolint
func (c *Config) tlsConfig() (*tls.Config, error) {
	cert, err := tls.X509KeyPair(c.CertPEMBlock, c.KeyPEMBlock)
	if err != nil {
		return &tls.Config{}, errors.Wrap(err, "error loading X509 TLS key pair")
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(c.CaPEMCerts)

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
	}, nil
}
