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
package crunch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"moul.io/http2curl"
)

const baseURL = "https://platform.42crunch.com/api/v1"

const debug = "CRUNCH_DEBUG"

func (c *Client) GetCollection(id string) (*Item, *ErrorResponse, error) {

	collection := &Item{}

	path := fmt.Sprintf("%s/collections/%s", c.baseURL, id)
	resp, err := c.DoRequest("GET", path, nil, collection)
	if err != nil {
		return nil, resp, err

	}
	return collection, resp, err
}

func (c *Client) ListCollections() (*Collections, *ErrorResponse, error) {
	collections := &Collections{}

	path := fmt.Sprintf("%s/collections", c.baseURL)
	resp, err := c.DoRequest("GET", path, nil, collections)
	if err != nil {
		return nil, resp, err

	}
	return collections, resp, err
}

func (c *Client) CreateCollection(collection *Collection) (*Item, *ErrorResponse, error) {

	toReturn := &Item{}
	path := fmt.Sprintf("%s/collections", c.baseURL)
	resp, err := c.DoRequest("POST", path, collection, toReturn)
	if err != nil {
		return nil, resp, err

	}
	return toReturn, resp, err
}

func (c *Client) CreateAPI(api *API) (*Item, *ErrorResponse, error) {
	i := &Item{}
	path := "https://platform.42crunch.com/api/v2/apis"

	resp, err := c.DoRequest("POST", path, api, i)
	if err != nil {
		return nil, resp, err

	}
	return i, resp, err
}

type Client struct {
	client  *http.Client
	apiKey  string
	baseURL *url.URL
}

func NewClient(apiKey string, httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	url, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	client := &Client{
		apiKey:  apiKey,
		baseURL: url,
		client:  httpClient,
	}

	return client, nil
}

func (c *Client) DoRequest(method, path string, body, v interface{}) (*ErrorResponse, error) {
	req, err := c.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}

	return c.Do(req, v)
}

func (c *Client) NewRequest(method, path string, body interface{}) (*http.Request, error) {
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
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest(method, u.String(), bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, u.String(), nil)
		if err != nil {
			return nil, err
		}
	}

	if err != nil {
		return nil, err
	}

	req.Close = true

	if _, ok := os.LookupEnv(debug); ok {
		command, _ := http2curl.GetCurlCommand(req)
		fmt.Println(command)
	}
	req.Header.Add("X-API-KEY", c.apiKey)
	return req, nil
}

func (c *Client) Do(req *http.Request, v interface{}) (*ErrorResponse, error) {
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
	if _, ok := os.LookupEnv(debug); ok {
		fmt.Println("Response body:", string(o))
	}
	err = json.Unmarshal(o, v)
	if err != nil {
		return nil, err
	}

	return nil, err
}

type Collection struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	IsShared      bool   `json:"isShared"`
	IsSharedWrite bool   `json:"isSharedWrite"`
}

type API struct {
	CollectionID string `json:"cid"`
	Name         string `json:"name"`
	OAS          string `json:"specfile"`
	IsYaml       bool   `json:"yaml"`
}

type Collections struct {
	List []Item `json:"list"`
}

type Item struct {
	Desc CollectionDescription `json:"desc"`
}
type CollectionDescription struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	TechnicalName string `json:"technicalName"`
	Source        string `json:"source"`
	IsShared      bool   `json:"isShared"`
	IsSharedWrite bool   `json:"isSharedWrite"`
}

type ErrorResponse struct {
	StatusCode int
	Errors     string
}
