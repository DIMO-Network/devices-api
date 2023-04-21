package services

import (
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/shared"
	"time"
)

type DCNService interface {
	GetRecords() any
}

type dcnService struct {
	Settings   *config.Settings
	httpClient shared.HTTPClientWrapper
	//dbs        func() *db.ReaderWriter
}

func NewDcnService(settings *config.Settings) DCNService {
	client, _ := shared.NewHTTPClientWrapper("https://multicall.dimo/", "", 20*time.Second, nil, true)
	return &dcnService{
		Settings:   settings,
		httpClient: client,
	}
}

func (ds *dcnService) GetRecords() any {
	return nil
}
