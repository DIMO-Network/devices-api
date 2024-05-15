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
	img = "data:image/png;base64,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"
)

func TestIPFSUpload_Success(t *testing.T) {
	ctx := context.Background()
	ipfs := NewIPFSLoader(&config.Settings{
		IPFSURL: "https://assets.dev.dimo.xyz/ipfs",
	})

	pdb, _ := test.StartContainerDatabase(context.Background(), t, "../../../migrations")

	cid, err := ipfs.UploadImage(ctx, img)
	assert.NoError(t, err)
	assert.Equal(t, "QmT9TF5mWHmcCJqdGxQ5DDB3jdwFmJ2DdPQY9JyC9Cow6U", cid)

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
