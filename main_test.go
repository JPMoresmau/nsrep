package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	item "github.com/JPMoresmau/nsrep/item"
	"github.com/stretchr/testify/require"
)

func DoTestItem(t *testing.T, id item.ID) {
	require := require.New(t)

	s := `{"type":"Table","name":"Table1","contents":{}}`
	data := fmt.Sprintf(`{"id":%s,"type":"Table","name":"Table1","contents":{}}`, "[\""+strings.Join(id, "\",\"")+"\"]")
	url := fmt.Sprintf("http://localhost:9999/items/%s", item.IDToString(id))
	resp, err := http.Get(url)
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(404, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(err)
	require.Equal(`{"id":null,"type":"","name":"","contents":null}`, string(body))

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

	DoTestDelete(t, url)

	resp, err = http.Get("http://localhost:9999/items/123")
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(404, resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	require.Nil(err)
	require.Equal(`{"id":null,"type":"","name":"","contents":null}`, string(body))

	resp, err = http.Get("http://localhost:9999/items/Model")
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(200, resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	require.Nil(err)
	require.Equal(`{"id":["Model"],"type":"Model","name":"Model","contents":{"typeAttributes":{},"typeChildren":{"":["Table"]}}}`, string(body))
}

func DoTestDelete(t *testing.T, url string) {
	require := require.New(t)
	req, err := http.NewRequest("DELETE", url, nil)
	require.Nil(err)
	require.NotNil(req)
	resp, err := http.DefaultClient.Do(req)
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(204, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(err)
	require.Equal(``, string(body))

	resp, err = http.Get(url)
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(404, resp.StatusCode)
}

func TestItemInvalidID(t *testing.T) {
	require := require.New(t)
	store := item.NewLocalStore()
	srv, err := startServer(9999, store, nil)
	require.NoError(err)
	defer stopServer(srv)
	id := "123"
	s := `{"type":"Table","name":"Table1","contents":{}}`
	url := fmt.Sprintf("http://localhost:9999/items/%s", id)
	resp, err := http.Get(url)
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(404, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(err)
	require.Equal(`{"id":null,"type":"","name":"","contents":null}`, string(body))

	resp, err = http.Post(url, "application/json", strings.NewReader(s))
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(400, resp.StatusCode)
}

func TestItemsSlashID(t *testing.T) {
	store := item.NewLocalStore()
	srv, err := startServer(9999, store, nil)
	require.NoError(t, err)
	defer stopServer(srv)
	DoTestItem(t, []string{"Table", "Table1"})

}
