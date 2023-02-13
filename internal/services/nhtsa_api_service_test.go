package services

import (
	_ "embed"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

//go:embed test_nhtsa_decoded_vin.json
var testNhtsaDecodedVin string

func TestNHTSAService_DecodeVIN(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	ns := NewNHTSAService()
	url := "https://vpic.nhtsa.dot.gov/api//vehicles/decodevinextended/VIN123?format=json"
	httpmock.RegisterResponder(http.MethodGet, url, httpmock.NewStringResponder(200, testNhtsaDecodedVin))

	got, err := ns.DecodeVIN("VIN123")

	assert.NoError(t, err, "no error expected")
	assert.Len(t, got.Results, 138)
	assert.Equal(t, "TESLA", got.LookupValue("Make"))
	assert.Equal(t, "Model Y", got.LookupValue("Model"))
	assert.Equal(t, "2020", got.LookupValue("Model Year"))
}
