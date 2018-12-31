package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	item "github.com/JPMoresmau/nsrep/item"
	"github.com/stretchr/testify/require"
)

var config = item.Cassandra{Port: 0, Keyspace: "NSRepTest", Endpoints: []string{"localhost"}, Replication: 1}
var elastic = item.Elastic{URL: "http://55.0.0.2:9200", Shards: 1, Replicas: 0, Index: "items_http_test"}

func TestCql(t *testing.T) {
	require := require.New(t)
	store, err := item.NewCqlStore(config)
	require.Nil(err)
	require.NotNil(store)
	srv, err := startServer(9999, store, nil)
	require.NoError(err)
	defer stopServer(srv)

	DoTestItem(t, []string{"Team", "team1"})
	DoTestHistory(t, []string{"Team", "team1"})
}

func TestCqlEs(t *testing.T) {
	require := require.New(t)
	store, err := item.NewCqlStore(config)
	require.NoError(err)
	require.NotNil(store)
	es, err := item.NewElasticStore(elastic)
	require.NoError(err)
	require.NotNil(es)
	srv, err := startServer(9999, store, es)
	require.NoError(err)
	defer stopServer(srv)

	DoTestItem(t, []string{"Team", "team1"})
	DoTestHistory(t, []string{"Team", "team1"})
	DoTestSearch(t)
	DoTestDeleteTree(t)
	DoTestGraphQL(t)
}

func DoTestHistory(t *testing.T, id item.ID) {
	require := require.New(t)

	url := fmt.Sprintf("http://localhost:9999/history/%s?length=10", item.IDToString(id))
	resp, err := http.Get(url)
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(200, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(err)
	// t.Log(string(body))
	var its = []item.Status{}
	require.Empty(its)
	json.Unmarshal(body, &its)
	require.NotEmpty(its)
}

func DoTestSearch(t *testing.T) {
	require := require.New(t)
	id1 := "Organization/1"
	s1 := `{"type":"Organization","name":"Organization1","contents":{"field1":"value1","field2":"value2"}}`
	url1 := fmt.Sprintf("http://localhost:9999/items/%s", id1)
	id2 := "Organization/2"
	s2 := `{"type":"Organization","name":"Organization2","contents":{"field1":"value1","field2":"value3"}}`
	url2 := fmt.Sprintf("http://localhost:9999/items/%s", id2)

	resp, err := http.Post(url1, "application/json", strings.NewReader(s1))
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(200, resp.StatusCode)

	defer DoTestDelete(t, url1)

	resp, err = http.Post(url2, "application/json", strings.NewReader(s2))
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(200, resp.StatusCode)

	defer DoTestDelete(t, url2)

	time.Sleep(time.Second)

	url := fmt.Sprintf("http://localhost:9999/search?query=%s", "value1")
	resp, err = http.Get(url)
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(200, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(err)
	// t.Log(string(body))
	var its = []item.Score{}
	require.Empty(its)
	json.Unmarshal(body, &its)
	require.Equal(2, len(its))

}

func DoTestDeleteTree(t *testing.T) {
	require := require.New(t)
	id1 := "Organization/1"
	s1 := `{"type":"Organization","name":"Organization1","contents":{"field1":"value1","field2":"value2"}}`
	url1 := fmt.Sprintf("http://localhost:9999/items/%s", id1)
	id2 := "Organization/1/Team/1"
	s2 := `{"type":"Team","name":"Team1","contents":{"field1":"value1","field2":"value3"}}`
	url2 := fmt.Sprintf("http://localhost:9999/items/%s", id2)

	resp, err := http.Post(url1, "application/json", strings.NewReader(s1))
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(200, resp.StatusCode)

	resp, err = http.Post(url2, "application/json", strings.NewReader(s2))
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(200, resp.StatusCode)

	time.Sleep(time.Second)

	DoTestDelete(t, url1)

	resp, err = http.Get(url2)
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(404, resp.StatusCode)

	// delete should be idempotent
	DoTestDelete(t, url1)

	resp, err = http.Get(url2)
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(404, resp.StatusCode)
}

func DoTestGraphQL(t *testing.T) {
	require := require.New(t)

	id1 := []string{"Organization", "1"}
	s1 := `{"type":"Organization","name":"Organization1","contents":{"field1":"value1","field2":"value2"}}`
	url1 := fmt.Sprintf("http://localhost:9999/items/%s", item.IDToString(id1))
	id2 := []string{"Organization", "1", "Team", "1"}
	s2 := `{"type":"Team","name":"Team1","contents":{"field1":"value1","field2":"value3"}}`
	url2 := fmt.Sprintf("http://localhost:9999/items/%s", item.IDToString(id2))

	resp, err := http.Post(url1, "application/json", strings.NewReader(s1))
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(200, resp.StatusCode)

	resp, err = http.Post(url2, "application/json", strings.NewReader(s2))
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(200, resp.StatusCode)

	defer DoTestDelete(t, url1)

	time.Sleep(time.Second)
	//graphql := `{
	//	__schema {
	//	  types {
	//		name
	//	  }
	//	}
	// }`
	testGraphQL(require,
		"{Organization(name:\"Organization1\"){field1}}",
		`{"data":{"Organization":[{"field1":"value1"}]}}`)

	testGraphQL(require,
		"{Organization(name:\"Organization1\"){Team{field2}}}",
		`{"data":{"Organization":[{"Team":[{"field2":"value3"}]}]}}`)
}

func testGraphQL(require *require.Assertions, graphql string, expected string) {
	resp, err := http.Post("http://localhost:9999/graphql", "application/json", strings.NewReader(graphql))
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(200, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(err)
	//log.Printf("graphql: %s", body)
	require.Equal(expected, string(body))
}
