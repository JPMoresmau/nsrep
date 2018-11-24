package main

import (
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
	res := m.Run()
	stopServer(srv)
	os.Exit(res)
}

func TestItems(t *testing.T) {
	require := require.New(t)

	s := `{"type":"Table","name":"Table1","contents":{}}`

	resp, err := http.Get("http://localhost:9999/items/123")
	require.Nil(err)
	require.NotNil(resp)
	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(err)
	require.Equal(`{"id":"","type":"","name":"","contents":null}`, string(body))

	resp, err = http.Post("http://localhost:9999/items/123", "application/json", strings.NewReader(s))
	require.Nil(err)
	require.NotNil(resp)
	body, err = ioutil.ReadAll(resp.Body)
	require.Nil(err)
	require.Equal(`{"id":"123","type":"Table","name":"Table1","contents":{}}`, string(body))

	resp, err = http.Get("http://localhost:9999/items/123")
	require.Nil(err)
	require.NotNil(resp)
	body, err = ioutil.ReadAll(resp.Body)
	require.Nil(err)
	require.Equal(`{"id":"123","type":"Table","name":"Table1","contents":{}}`, string(body))

	req, err := http.NewRequest("DELETE", "http://localhost:9999/items/123", nil)
	require.Nil(err)
	require.NotNil(req)
	resp, err = http.DefaultClient.Do(req)
	require.Nil(err)
	require.NotNil(resp)
	body, err = ioutil.ReadAll(resp.Body)
	require.Nil(err)
	require.Equal(`{"id":"123","type":"Table","name":"Table1","contents":{}}`, string(body))

	resp, err = http.Get("http://localhost:9999/items/123")
	require.Nil(err)
	require.NotNil(resp)
	body, err = ioutil.ReadAll(resp.Body)
	require.Nil(err)
	require.Equal(`{"id":"","type":"","name":"","contents":null}`, string(body))

}
