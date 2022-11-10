package overlays

import (
	"fmt"
	"testing"
)

func TestOverlay(t *testing.T) {
	o := &Overlay{}

	//https://gist.githubusercontent.com/ponelat/daac5912ede1871629b6028bbe715d3a/raw/2871f9f27fb93d1c01567d198fb60cd1271e7dcf/overlay.yml
	fmt.Println(o.Parse("https://gist.githubusercontent.com/ponelat/daac5912ede1871629b6028bbe715d3a/raw/2871f9f27fb93d1c01567d198fb60cd1271e7dcf/overlay.yml"))

	fmt.Println(applyOverlay("/root/go/src/github.com/kubeshop/kusk-gateway/cmd/kusk/internal/overlays/overlay.yaml", ""))

}
