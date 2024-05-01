package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DIMO-Network/devices-api/internal/constants"
)

func TestGetSliceDiff(t *testing.T) {
	capable := []string{constants.TelemetrySubscribe, constants.DoorsLock, constants.DoorsUnlock}
	enabled := []string{constants.DoorsLock, constants.DoorsUnlock}
	actual := SliceDiff(capable, enabled)
	assert.ElementsMatch(t, []string{constants.TelemetrySubscribe}, actual)
}
