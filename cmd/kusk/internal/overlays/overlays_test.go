package overlays

import (
	"testing"
)

func TestOverlay(t *testing.T) {
	if _, err := applyOverlay("overlay.yaml", ""); err != nil {
		t.Log(err)
		t.Fail()
	}
}
