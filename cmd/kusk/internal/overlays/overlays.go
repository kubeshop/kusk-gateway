package overlays

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/ghodss/yaml"
)

const imageName = "docker.io/jasmingacic/overlay-cli"

type Overlay struct {
	Overlays string   `json:"overlays,omitempty" yaml:"overlays,omitempty"`
	Extends  string   `json:"extends,omitempty" yaml:"extends,omitempty"`
	Actions  []Action `json:"actions,omitempty" yaml:"actions,omitempty"`
}

type Action struct {
	Target string `json:"target,omitempty" yaml:"target,omitempty"`
	Remove bool   `json:"remove,omitempty" yaml:"remove,omitempty"`
	Update Update `json:"update,omitempty" yaml:"update,omitempty"`
}

type Update struct {
	Title       string `json:"title,omitempty" yaml:"title,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

func (o *Overlay) Parse(path string) (*Overlay, string, error) {

	// isPath := false
	// _, err := url.ParseRequestURI(path)
	// isPath = (err != nil)

	if IsUrl(path) {
		dat, err := os.ReadFile(path)
		if err != nil {
			return nil, "", err
		}
		if err := yaml.Unmarshal(dat, o); err != nil {
			return nil, "", err
		}

		f, err := os.CreateTemp("kusk-cli", "overlay")
		if err != nil {
			return nil, "", err
		} else {
			if _, err := f.Write(dat); err != nil {
				return nil, "", err
			}
		}
		applyOverlay(f.Name(), "")

		return o, f.Name(), nil
	} else {
		overlay, err := getFile(path)
		// if
		if !IsUrl(overlay.Extends) {
			applyOverlay(path, o.Extends)
		} else {
			applyOverlay(path, "")
		}
		return overlay, path, err
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

	if err := yaml.Unmarshal(o, overlay); err != nil {
		return nil, err
	}

	return overlay, nil
}

func applyOverlay(path string, extends string) string {

	// docker run  --rm -ti  -v ${PWD}/samples/only-user-tag.yml:/overlay.yaml overlay-cli
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	defer cli.Close()
	reader, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	defer reader.Close()
	io.Copy(io.Discard, reader)
	volumes := fmt.Sprintf("-v=%s:/overlay.yaml", path)
	if len(extends) > 0 {
		volumes = fmt.Sprintf(volumes, "-v=%s:/%s", extends, extends)

	}
	cmd := exec.Command("docker", "run", "--rm", volumes, imageName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(out)
}

func IsUrl(str string) bool {
	url, err := url.ParseRequestURI(str)
	if err != nil {
		return false
	}

	address := net.ParseIP(url.Host)

	if address == nil {
		return strings.Contains(url.Host, ".")
	}

	return true
}
