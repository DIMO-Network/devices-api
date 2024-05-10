package ipfs

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
)

type IPFS struct {
	url    string
	client *http.Client
}

type ipfsResponse struct {
	Success bool   `json:"success"`
	CID     string `json:"cid"`
}

func NewIPFSLoader(settings *config.Settings) *IPFS {
	return &IPFS{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		url: settings.IPFSURL,
	}
}

func (i *IPFS) UploadImage(img string) (string, error) {
	var respb ipfsResponse
	imageData := strings.TrimPrefix(img, "data:image/png;base64,")
	image, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return respb.CID, err
	}

	if len(image) == 0 {
		return respb.CID, fmt.Errorf("Empty image field.")
	}
	reader := bytes.NewReader(image)
	resp, err := i.client.Post(i.url, "image/png", reader)
	if err != nil {
		return respb.CID, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != http.StatusOK {
		return respb.CID, fmt.Errorf("status code %d", code)
	}

	if err := json.NewDecoder(resp.Body).Decode(&respb); err != nil {
		return respb.CID, err
	}

	if !respb.Success {
		return respb.CID, fmt.Errorf("failed to upload image to IPFS")
	}

	return respb.CID, nil
}

func (i *IPFS) EncodeImage(path string) (string, error) {
	var base64Encoding string
	img, err := os.ReadFile(path)
	if err != nil {
		return base64Encoding, err
	}

	base64Encoding += "data:image/png;base64,"
	base64Encoding += base64.StdEncoding.EncodeToString(img)

	return base64Encoding, nil
}
