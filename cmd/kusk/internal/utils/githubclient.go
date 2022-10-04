// MIT License
//
// Copyright (c) 2022 Kubeshop
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	baseURL = "https://api.github.com/repos/kubeshop/" //kusk-gateway/"
)

func (c *GithubClient) GetTags(repo string) (*Latest, *ErrorResponse, error) {
	i := &Latest{}

	path := fmt.Sprintf("%s/releases/latest", repo)
	if resp, err := c.DoRequest("GET", path, nil, i); err == nil {
		return i, resp, err
	} else {
		return nil, resp, err
	}
}

func (c *GithubClient) GetLatest(repo string) (string, error) {
	i, _, err := c.GetTags(repo)
	if err != nil {
		return "", err
	}

	return i.TagName, nil
}

func NewGithubClient(apiKey string, httpClient *http.Client) (*GithubClient, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	url, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	client := &GithubClient{
		apiKey:  apiKey,
		baseURL: url,
		client:  httpClient,
	}

	return client, nil
}

type GithubClient struct {
	client  *http.Client
	apiKey  string
	baseURL *url.URL
}

type ErrorResponse struct {
	StatusCode int
	Errors     string
}

func (c *GithubClient) DoRequest(method, path string, body, v interface{}) (*ErrorResponse, error) {
	req, err := c.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}

	return c.Do(req, v)
}

func (c *GithubClient) NewRequest(method, path string, body interface{}) (*http.Request, error) {
	// relative path to append to the endpoint url, no leading slash please
	if path[0] == '/' {
		path = path[1:]
	}
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	u := c.baseURL.ResolveReference(rel)
	var req *http.Request
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		req, _ = http.NewRequest(method, u.String(), bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest(method, u.String(), nil)
	}
	if err != nil {
		return nil, err
	}

	req.Close = true

	req.Header.Add("MC-Api-Key", c.apiKey)
	return req, nil
}

func (c *GithubClient) Do(req *http.Request, v interface{}) (*ErrorResponse, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		o, _ := io.ReadAll(resp.Body)
		errResp := &ErrorResponse{
			StatusCode: resp.StatusCode,
			Errors:     string(o),
		}

		return errResp, fmt.Errorf("%s returned %d", req.URL, resp.StatusCode)
	}

	o, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(o, v)
	if err != nil {
		return nil, err
	}

	return nil, err
}

// https://api.github.com/repos/kubeshop/kusk-gateway/releases/latest
type Tag struct {
	Ref    string `json:"ref,omitempty"`
	NodeID string `json:"node_id,omitempty"`
	URL    string `json:"url,omitempty"`
	Object Obj    `json:"object,omitempty"`
}

type Obj struct {
	SHA   string `json:"sha,omitempty"`
	TType string `json:"type,omitempty"`
	URL   string `json:"url,omitempty"`
}

type Latest struct {
	URL       string `json:"url"`
	AssetsURL string `json:"assets_url"`
	UploadURL string `json:"upload_url"`
	HTMLURL   string `json:"html_url"`
	ID        int    `json:"id"`
	Author    struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"author"`
	NodeID          string    `json:"node_id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at"`
	Assets          []struct {
		URL      string `json:"url"`
		ID       int    `json:"id"`
		NodeID   string `json:"node_id"`
		Name     string `json:"name"`
		Label    string `json:"label"`
		Uploader struct {
			Login             string `json:"login"`
			ID                int    `json:"id"`
			NodeID            string `json:"node_id"`
			AvatarURL         string `json:"avatar_url"`
			GravatarID        string `json:"gravatar_id"`
			URL               string `json:"url"`
			HTMLURL           string `json:"html_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			OrganizationsURL  string `json:"organizations_url"`
			ReposURL          string `json:"repos_url"`
			EventsURL         string `json:"events_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"uploader"`
		ContentType        string    `json:"content_type"`
		State              string    `json:"state"`
		Size               int       `json:"size"`
		DownloadCount      int       `json:"download_count"`
		CreatedAt          time.Time `json:"created_at"`
		UpdatedAt          time.Time `json:"updated_at"`
		BrowserDownloadURL string    `json:"browser_download_url"`
	} `json:"assets"`
	TarballURL    string `json:"tarball_url"`
	ZipballURL    string `json:"zipball_url"`
	Body          string `json:"body"`
	DiscussionURL string `json:"discussion_url"`
}
