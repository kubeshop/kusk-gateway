package utils

import (
	"fmt"
	"testing"
)

func TestGetTags(t *testing.T) {
	c, err := NewGithubClient("", nil)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	tags, err := c.GetLatest("kusk-gateway")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	fmt.Println(tags)

	tags, err = c.GetLatest("kusk-gateway-dashboard")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	fmt.Println(tags)

	tags, err = c.GetLatest("kuskgateway-api-server")
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	fmt.Println(tags)
}
