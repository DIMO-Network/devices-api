package ipfs

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const (
	img = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAoAAAAKCAYAAACNMs+9AAAAFUlEQVR42mP8z8BQz0AEYBxVSF+FABJADveWkH6oAAAAAElFTkSuQmCC"
)

func TestIPFSUpload_Success(t *testing.T) {
	ctx := context.Background()
	ipfs := NewLoader(&config.Settings{
		IPFSURL: "https://assets.dev.dimo.xyz/ipfs",
	})

	pdb, _ := test.StartContainerDatabase(context.Background(), t, "../../../migrations")

	cid, err := ipfs.UploadImage(ctx, img)
	assert.NoError(t, err)
	assert.Equal(t, "Qme23PqtDXmeyETzG3W3sy3ZWTjF2ZQGJWrCG5svtFq8aB", cid)

	resp, err := http.Get(ipfs.url + "/" + cid)
	assert.NoError(t, err)
	defer resp.Body.Close()

	bdy, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var base64Encoding string
	base64Encoding += "data:image/png;base64,"
	base64Encoding += base64.StdEncoding.EncodeToString(bdy)

	assert.Equal(t, base64Encoding, img)

	ud := models.UserDevice{
		ID:                 ksuid.New().String(),
		UserID:             ksuid.New().String(),
		DeviceDefinitionID: ksuid.New().String(),
		IpfsImageCid:       null.StringFrom(cid),
	}

	if err := ud.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		t.Fatal(err)
	}

}
