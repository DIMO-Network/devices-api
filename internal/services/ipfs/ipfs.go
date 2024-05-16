package ipfs

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
)

const (
	imagePrefix       = "data:image/png;base64,"
	contentTypeHeader = "image/png"
)

type IPFS struct {
	url    string
	client *http.Client
}

type ipfsResponse struct {
	Success bool   `json:"success"`
	CID     string `json:"cid"`
}

func NewGateway(settings *config.Settings) *IPFS {
	return &IPFS{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		url: settings.IPFSURL,
	}
}

func (i *IPFS) UploadImage(ctx context.Context, img string) (string, error) {
	imageData := strings.TrimPrefix(img, imagePrefix)
	image, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %v", err)
	}

	if len(image) == 0 {
		return "", errors.New("empty image field")
	}

	reader := bytes.NewReader(image)
	req, err := http.NewRequest(http.MethodPost, i.url, reader)
	if err != nil {
		return "", fmt.Errorf("failed to create image upload req: %v", err)
	}

	req.Header.Set("Content-Type", contentTypeHeader)
	resp, err := i.client.Do(req.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("IPFS post request failed: %v", err)
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != http.StatusOK {
		return "", fmt.Errorf("status code %d", code)
	}

	var respb ipfsResponse
	if err := json.NewDecoder(resp.Body).Decode(&respb); err != nil {
		return "", fmt.Errorf("failed to decode IPFS response: %v", err)
	}

	if !respb.Success {
		return "", errors.New("failed to upload image to IPFS")
	}

	return respb.CID, nil
}

func (i *IPFS) FetchImage(ctx context.Context, cid string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, i.url+"/"+cid, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create fetch image req: %v", err)
	}

	resp, err := i.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("IPFS get request failed: %v", err)
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != http.StatusOK {
		return nil, fmt.Errorf("status code %d", code)
	}

	bdy, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read IPFS response: %v", err)
	}

	return bdy, nil
}
