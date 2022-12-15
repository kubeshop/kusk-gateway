package overlays

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"sigs.k8s.io/yaml"
)

const imageName = "kubeshop/overlay-cli"

type Overlay struct {
	Overlays string   `json:"overlays,omitempty" yaml:"overlays,omitempty"`
	Extends  string   `json:"extends,omitempty" yaml:"extends,omitempty"`
	Actions  []Action `json:"actions,omitempty" yaml:"actions,omitempty"`
	path     string
	url      string
}

type Action struct {
	Target string      `json:"target,omitempty" yaml:"target,omitempty"`
	Remove bool        `json:"remove,omitempty" yaml:"remove,omitempty"`
	Update interface{} `json:"update,omitempty" yaml:"update,omitempty"`
	Where  interface{} `json:"where,omitempty" yaml:"where,omitempty"`
}

func NewOverlay(path string) (o *Overlay, err error) {
	o = &Overlay{}
	if !IsUrl(path) {
		dat, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		if err := yaml.UnmarshalStrict(dat, o); err != nil {
			return nil, err
		}
		overlay, err := os.CreateTemp("", "overlay")
		if err != nil {
			return nil, err
		} else {
			if _, err := overlay.Write(dat); err != nil {
				return nil, err
			}
		}
		o.path = overlay.Name()

		return o, nil
	}
	return getFile(path)
}

func (o *Overlay) Apply() (string, error) {
	var err error
	var overlayed string
	if !IsUrl(o.Extends) {
		overlayed, err = applyOverlay(o.path, o.Extends)
		if err != nil {
			return "", err
		}
	} else {
		overlayed, err = applyOverlay(o.path, "")
		if err != nil {
			return "", err
		}
	}
	if f, err := os.CreateTemp("", "overlay"); err != nil {
		return "", err
	} else {
		if _, err := f.Write([]byte(overlayed)); err != nil {
			return "", err
		}
		return f.Name(), err
	}
}

func getFile(url string) (overlay *Overlay, err error) {
	// Get the data
	overlay = &Overlay{}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	o, err := io.ReadAll(resp.Body)

	if err != nil {
		return overlay, err
	}

	if err := yaml.UnmarshalStrict(o, overlay); err != nil {
		return nil, err
	}

	if ov, err := os.CreateTemp("", "overlay"); err != nil {
		return nil, err
	} else {
		if _, err := ov.Write(o); err != nil {
			return nil, err
		}

		overlay.url = url
		overlay.path = ov.Name()

		return overlay, nil
	}
}

func applyOverlay(path string, extends string) (string, error) {
	abs, _ := filepath.Abs(path)
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", err
	}
	defer cli.Close()
	reader, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return "", err
	}

	defer reader.Close()
	if _, err := io.Copy(io.Discard, reader); err != nil {
		return "", fmt.Errorf("download of %v image did not complete: %w", imageName, err)
	}

	volumes := fmt.Sprintf("-v=%s:/overlay.yaml", abs)
	var extendVolume string
	if len(extends) > 0 {
		if FileExists(extends) {
			base := filepath.Base(extends)
			extendsAbs, _ := filepath.Abs(extends)
			extendVolume = fmt.Sprintf("-v=%s:/%s", extendsAbs, base)
		} else {
			return "", fmt.Errorf(fmt.Sprintf("%s file does not exist", extends))
		}
	}

	var cmd *exec.Cmd
	if len(extendVolume) > 0 {
		cmd = exec.Command("docker", "run", "--rm", volumes, extendVolume, imageName)
	} else {
		cmd = exec.Command("docker", "run", "--rm", volumes, imageName)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func IsUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
