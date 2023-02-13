package elasticsearch

import (
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/rs/zerolog"
)

type ElasticSearch struct {
	esClient *elasticsearch.Client
	logger   *zerolog.Logger
}

func NewElasticSearch(s config.Settings, logger *zerolog.Logger) (ElasticSearch, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{s.ElasticSearchEnrichStatusHost},
		Username:  s.ElasticSearchEnrichStatusUsername,
		Password:  s.ElasticSearchEnrichStatusPassword,
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return ElasticSearch{}, nil
	}
	return ElasticSearch{
		esClient: es,
		logger:   logger,
	}, nil
}
