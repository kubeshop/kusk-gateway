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
	val, ok := os.LookupEnv("CRUNCH42")
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
		t.Skip("skip test") // remove to run test
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
		t.Skip("skip test") // remove to run test
	}

	col, _, err := c.ListCollections()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	fmt.Println(col.List[0].Desc.Name)
}

func TestGetCollection(t *testing.T) {
	fmt.Println("collID", collID)
	if shouldSkip() {
		t.Skip("skip test") // remove to run test
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
