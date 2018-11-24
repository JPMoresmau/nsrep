package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	item "github.com/JPMoresmau/metarep/item"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	store := item.NewLocalStore()
	srv := startServer(9999, store)
	defer stopServer(srv)
	res := m.Run()
	os.Exit(res)
}

func testItem(t *testing.T, id string) {
	require := require.New(t)

	s := `{"type":"Table","name":"Table1","contents":{}}`
	data := fmt.Sprintf(`{"id":"%s","type":"Table","name":"Table1","contents":{}}`, id)
	url := fmt.Sprintf("http://localhost:9999/items/%s", id)
	resp, err := http.Get(url)
	require.Nil(err)
	require.NotNil(resp)
	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(err)
	require.Equal(`{"id":"","type":"","name":"","contents":null}`, string(body))

	resp, err = http.Post(url, "application/json", strings.NewReader(s))
	require.Nil(err)
	require.NotNil(resp)
	body, err = ioutil.ReadAll(resp.Body)
	require.Nil(err)
	require.Equal(data, string(body))

	resp, err = http.Get(url)
	require.Nil(err)
	require.NotNil(resp)
	body, err = ioutil.ReadAll(resp.Body)
	require.Nil(err)
	require.Equal(data, string(body))

	req, err := http.NewRequest("DELETE", url, nil)
	require.Nil(err)
	require.NotNil(req)
	resp, err = http.DefaultClient.Do(req)
	require.Nil(err)
	require.NotNil(resp)
	body, err = ioutil.ReadAll(resp.Body)
	require.Nil(err)
	require.Equal(data, string(body))

	resp, err = http.Get("http://localhost:9999/items/123")
	require.Nil(err)
	require.NotNil(resp)
	body, err = ioutil.ReadAll(resp.Body)
	require.Nil(err)
	require.Equal(`{"id":"","type":"","name":"","contents":null}`, string(body))
}

func TestItems(t *testing.T) {
	testItem(t, "123")
}

func TestItemsSlashID(t *testing.T) {
	testItem(t, "Table/Table1")

}
