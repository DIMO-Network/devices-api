package ipfs

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	imgBase64   = "iVBORw0KGgoAAAANSUhEUgAAAAoAAAAKCAYAAACNMs+9AAAAFUlEQVR42mP8z8BQz0AEYBxVSF+FABJADveWkH6oAAAAAElFTkSuQmCC"
	fullImgData = "data:image/png;base64," + imgBase64
	testCID     = "Qme23PqtDXmeyETzG3W3sy3ZWTjF2ZQGJWrCG5svtFq8aB"
)

func TestIPFSUpload_Success(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	imgBytes, err := base64.StdEncoding.DecodeString(imgBase64)
	require.NoError(err)

	serv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		assert.Equal("image/png", r.Header.Get("Content-Type"))
		b, err := io.ReadAll(r.Body)
		require.NoError(err)

		assert.Equal(imgBytes, b)

		outB, err := json.Marshal(struct {
			Success bool
			CID     string
		}{Success: true, CID: testCID})
		require.NoError(err)

		w.WriteHeader(200)
		_, err = w.Write(outB)
		require.NoError(err)
	}))
	defer serv.Close()

	ctx := context.Background()
	ipfs, err := NewGateway(&config.Settings{
		IPFSURL: serv.URL,
	})
	require.NoError(err)

	cid, err := ipfs.UploadImage(ctx, fullImgData)
	require.NoError(err)
	assert.Equal(testCID, cid)
}
