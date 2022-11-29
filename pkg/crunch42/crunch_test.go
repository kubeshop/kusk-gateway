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
	"fmt"
	"os"
	"testing"
)

var apiKey = ""
var c *Client
var collID = ""

func shouldSkip() bool {
	val, ok := os.LookupEnv("CRUNCH42_TOKEN")
	apiKey = val

	var err error
	c, err = NewClient(apiKey, nil)
	if err != nil {
		fmt.Println(err)
		return true
	}
	return !ok
}

func TestCreateCollection(t *testing.T) {
	if shouldSkip() {
		t.Skip("skip test")
	}

	col, _, err := c.CreateCollection(&Collection{
		Name: "test",
	})
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	collID = col.Desc.ID
	fmt.Println("collID", collID)

}

func TestListCollections(t *testing.T) {
	if shouldSkip() {
		t.Skip("skip test")
	}

	col, _, err := c.ListCollections()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	fmt.Println(col.List[0].Desc.Name)
}

func TestGetCollection(t *testing.T) {
	if shouldSkip() {
		t.Skip("skip test")
	}

	col, _, err := c.GetCollection(collID)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	if col.Desc.ID != collID {
		t.Log("Retrieved ID isn't the same as expected")
		t.Fail()
	}
}

func TestCreateAPI(t *testing.T) {
	if shouldSkip() {
		t.Skip("skip test")
	}

	_, _, err := c.CreateAPI(&API{
		Name:         "test",
		CollectionID: collID,
		IsYaml:       true,
		OAS:          `ewogICAgIm9wZW5hcGkiOiAiMy4wLjAiLAogICAgImluZm8iOiB7CiAgICAgICAgInRpdGxlIjogInNpbXBsZS1hcGkiLAogICAgICAgICJ2ZXJzaW9uIjogIjAuMS4wIgogICAgfSwKICAgICJ4LWt1c2siOiB7CiAgICAgICAgImNvcnMiOiB7CiAgICAgICAgICAgICJvcmlnaW5zIjogWwogICAgICAgICAgICAgICAgIioiCiAgICAgICAgICAgIF0sCiAgICAgICAgICAgICJtZXRob2RzIjogWwogICAgICAgICAgICAgICAgIkdFVCIsCiAgICAgICAgICAgICAgICAiUE9TVCIKICAgICAgICAgICAgXQogICAgICAgIH0sCiAgICAgICAgInVwc3RyZWFtIjogewogICAgICAgICAgICAic2VydmljZSI6IHsKICAgICAgICAgICAgICAgICJuYW1lIjogImhlbGxvLXdvcmxkLXN2YyIsCiAgICAgICAgICAgICAgICAibmFtZXNwYWNlIjogImRlZmF1bHQiLAogICAgICAgICAgICAgICAgInBvcnQiOiA4MDgwCiAgICAgICAgICAgIH0KICAgICAgICB9CiAgICB9LAogICAgInBhdGhzIjogewogICAgICAgICIvaGVsbG8iOiB7CiAgICAgICAgICAgICJnZXQiOiB7CiAgICAgICAgICAgICAgICAicmVzcG9uc2VzIjogewogICAgICAgICAgICAgICAgICAgICIyMDAiOiB7CiAgICAgICAgICAgICAgICAgICAgICAgICJkZXNjcmlwdGlvbiI6ICJBIHNpbXBsZSBoZWxsbyB3b3JsZCEiLAogICAgICAgICAgICAgICAgICAgICAgICAiY29udGVudCI6IHsKICAgICAgICAgICAgICAgICAgICAgICAgICAgICJhcHBsaWNhdGlvbi9qc29uIjogewogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICJzY2hlbWEiOiB7CiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICJ0eXBlIjogIm9iamVjdCIsCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICJwcm9wZXJ0aWVzIjogewogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIm1lc3NhZ2UiOiB7CiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgInR5cGUiOiAic3RyaW5nIgogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgfQogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICB9CiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgfSwKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAiZXhhbXBsZSI6IHsKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIm1lc3NhZ2UiOiAiSGVsbG8gZnJvbSBhIG1vY2tlZCByZXNwb25zZSEiCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgfQogICAgICAgICAgICAgICAgICAgICAgICAgICAgfQogICAgICAgICAgICAgICAgICAgICAgICB9CiAgICAgICAgICAgICAgICAgICAgfQogICAgICAgICAgICAgICAgfQogICAgICAgICAgICB9CiAgICAgICAgfSwKICAgICAgICAiL3ZhbGlkYXRlZCI6IHsKICAgICAgICAgICAgIngta3VzayI6IHsKICAgICAgICAgICAgICAgICJ2YWxpZGF0aW9uIjogewogICAgICAgICAgICAgICAgICAgICJyZXF1ZXN0IjogewogICAgICAgICAgICAgICAgICAgICAgICAiZW5hYmxlZCI6IHRydWUKICAgICAgICAgICAgICAgICAgICB9CiAgICAgICAgICAgICAgICB9CiAgICAgICAgICAgIH0sCiAgICAgICAgICAgICJwb3N0IjogewogICAgICAgICAgICAgICAgInJlcXVlc3RCb2R5IjogewogICAgICAgICAgICAgICAgICAgICJkZXNjcmlwdGlvbiI6ICIiLAogICAgICAgICAgICAgICAgICAgICJjb250ZW50IjogewogICAgICAgICAgICAgICAgICAgICAgICAiYXBwbGljYXRpb24vanNvbiI6IHsKICAgICAgICAgICAgICAgICAgICAgICAgICAgICJzY2hlbWEiOiB7CiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgInR5cGUiOiAib2JqZWN0IiwKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAicHJvcGVydGllcyI6IHsKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIm5hbWUiOiB7CiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAidHlwZSI6ICJzdHJpbmciCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIH0KICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICB9LAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICJyZXF1aXJlZCI6IFsKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIm5hbWUiCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgXQogICAgICAgICAgICAgICAgICAgICAgICAgICAgfQogICAgICAgICAgICAgICAgICAgICAgICB9CiAgICAgICAgICAgICAgICAgICAgfQogICAgICAgICAgICAgICAgfSwKICAgICAgICAgICAgICAgICJyZXNwb25zZXMiOiB7CiAgICAgICAgICAgICAgICAgICAgIjIwMCI6IHsKICAgICAgICAgICAgICAgICAgICAgICAgImRlc2NyaXB0aW9uIjogIiIsCiAgICAgICAgICAgICAgICAgICAgICAgICJjb250ZW50IjogewogICAgICAgICAgICAgICAgICAgICAgICAgICAgInRleHQvcGxhaW4iOiB7CiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgInNjaGVtYSI6IHsKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgInR5cGUiOiAic3RyaW5nIgogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIH0sCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgImV4YW1wbGUiOiAiSGVsbG8gbW9ja2VkIEt1c2shIgogICAgICAgICAgICAgICAgICAgICAgICAgICAgfQogICAgICAgICAgICAgICAgICAgICAgICB9CiAgICAgICAgICAgICAgICAgICAgfQogICAgICAgICAgICAgICAgfQogICAgICAgICAgICB9CiAgICAgICAgfQogICAgfQp9`,
	})

	if err != nil {
		t.Log(err)
		t.Fail()
	}
}
