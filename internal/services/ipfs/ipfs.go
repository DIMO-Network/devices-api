package ipfs

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
)

const (
	imagePrefix          = "data:image/png;base64,"
	contentTypeHeaderKey = "Content-Type"
	pngContentType       = "image/png"
)

type IPFS struct {
	url    *url.URL
	client *http.Client
}

type ipfsResponse struct {
	Success bool   `json:"success"`
	CID     string `json:"cid"`
}

func NewGateway(settings *config.Settings) (*IPFS, error) {
	url, err := url.ParseRequestURI(settings.IPFSURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing IPFS URL %q: %w", settings.IPFSURL, err)
	}

	return &IPFS{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		url: url,
	}, nil
}

func (i *IPFS) UploadImage(ctx context.Context, img string) (string, error) {
	imageData := strings.TrimPrefix(img, imagePrefix)
	image, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	if len(image) == 0 {
		return "", errors.New("empty image field")
	}

	reader := bytes.NewReader(image)
	req, err := http.NewRequest(http.MethodPost, i.url.String(), reader)
	if err != nil {
		return "", fmt.Errorf("failed to create image upload req: %w", err)
	}

	req.Header.Set(contentTypeHeaderKey, pngContentType)
	resp, err := i.client.Do(req.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("IPFS post request failed: %w", err)
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != http.StatusOK {
		return "", fmt.Errorf("status code %d", code)
	}

	var respb ipfsResponse
	if err := json.NewDecoder(resp.Body).Decode(&respb); err != nil {
		return "", fmt.Errorf("failed to decode IPFS response: %w", err)
	}

	if !respb.Success {
		return "", errors.New("failed to upload image to IPFS")
	}

	return respb.CID, nil
}
