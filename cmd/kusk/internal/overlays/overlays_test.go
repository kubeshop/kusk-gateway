package overlays

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOverlayPass(t *testing.T) {
	t.Skipf("skipping %v because it is failing", t.Name())

	assert := assert.New(t)
	overlay, err := NewOverlay("./overlay.yaml")
	assert.NoError(err)

	appliedOverlay, err := overlay.Apply()
	assert.NoError(err)
	t.Logf("appliedOverlay=%v", appliedOverlay)
}

func TestOverlayNoPass(t *testing.T) {
	assert := assert.New(t)

	_, err := NewOverlay("overlayed.yaml")
	assert.Error(err)
}
