package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	item "github.com/JPMoresmau/metarep/item"
	"github.com/stretchr/testify/require"
)

func DoTestItem(t *testing.T, id string) {
	require := require.New(t)

	s := `{"type":"Table","name":"Table1","contents":{}}`
	data := fmt.Sprintf(`{"id":"%s","type":"Table","name":"Table1","contents":{}}`, id)
	url := fmt.Sprintf("http://localhost:9999/items/%s", id)
	resp, err := http.Get(url)
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(404, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(err)
	require.Equal(`{"id":"","type":"","name":"","contents":null}`, string(body))

	resp, err = http.Post(url, "application/json", strings.NewReader(s))
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(200, resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	require.Nil(err)
	require.Equal(data, string(body))

	resp, err = http.Get(url)
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(200, resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	require.Nil(err)
	require.Equal(data, string(body))

	req, err := http.NewRequest("DELETE", url, nil)
	require.Nil(err)
	require.NotNil(req)
	resp, err = http.DefaultClient.Do(req)
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(204, resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	require.Nil(err)
	require.Equal(``, string(body))

	resp, err = http.Get("http://localhost:9999/items/123")
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(404, resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	require.Nil(err)
	require.Equal(`{"id":"","type":"","name":"","contents":null}`, string(body))
}

func TestItems(t *testing.T) {
	store := item.NewLocalStore()
	srv := startServer(9999, store, nil)
	defer stopServer(srv)
	DoTestItem(t, "123")
}

func TestItemsSlashID(t *testing.T) {
	store := item.NewLocalStore()
	srv := startServer(9999, store, nil)
	defer stopServer(srv)
	DoTestItem(t, "Table/Table1")

}
