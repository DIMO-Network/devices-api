package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DIMO-Network/devices-api/internal/constants"
)

func TestGetSliceDiff(t *testing.T) {
	enabled := []string{constants.DoorsLock, constants.DoorsUnlock}
	capable := []string{constants.TelemetrySubscribe, constants.DoorsLock, constants.DoorsUnlock}
	actual := GetSliceDiff(enabled, capable)
	assert.Equal(t, []string{constants.TelemetrySubscribe}, actual)
}
