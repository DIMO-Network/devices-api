package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type jsonObj map[string]any

type EsDeviceStatusDTO struct {
	Region    string
	MakeSlug  string
	ModelSlug string
}

func (e ElasticSearch) UpdateAutopiDevicesByQuery(d services.DeviceDefinitionDTO, elasticDeviceStatusIndex string) error {
	query := jsonObj{
		"query": jsonObj{
			"bool": jsonObj{
				"filter": []jsonObj{
					{
						"term": jsonObj{
							"subject": d.UserDeviceID,
						},
					},
				},
			},
		},
		"script": jsonObj{
			"source": `ctx._source.data.deviceDefinitionId = params['deviceDefinitionId']; 
ctx._source.data.make = params['make']; 
ctx._source.data.model = params['model']; 
ctx._source.data.year = params['year'];`,
			"lang": "painless",
			"params": jsonObj{
				"deviceDefinitionId": d.DeviceDefinitionID,
				"make":               d.Make,
				"model":              d.Model,
				"year":               d.Year,
			},
		},
	}

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return fmt.Errorf("error encoding es query: %w", err)
	}

	req := esapi.UpdateByQueryRequest{
		Index:     []string{elasticDeviceStatusIndex},
		Body:      &buf,
		Conflicts: "proceed",
	}

	res, err := req.Do(context.Background(), e.esClient)
	if err != nil {
		return fmt.Errorf("error updating es by query: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var eResp jsonObj
		if err := json.NewDecoder(res.Body).Decode(&eResp); err != nil {
			return fmt.Errorf("error parsing the response body: %w", err)
		}
		return fmt.Errorf("es error during update, status:%s, error: %v", res.Status(), eResp)
	}

	var r UpdateByQueryResponse
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		e.logger.Err(err).Msgf("Error parsing response.")
	} else {
		e.logger.Info().Str("userDeviceId", d.UserDeviceID).Int64("took", r.Took).Msg("Updated device statuses.")
	}

	return nil
}
