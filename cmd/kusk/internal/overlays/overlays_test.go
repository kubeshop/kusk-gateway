package overlays

import (
	"testing"
)

func TestOverlayPass(t *testing.T) {
	overlay, err := NewOverlay("overlay.yaml")
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	if _, err := overlay.Apply(); err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestOverlayNoPass(t *testing.T) {
	if _, err := NewOverlay("overlayed.yaml"); err == nil {
		t.Fail()
	}
}
