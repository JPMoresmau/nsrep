package item

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/go-errors/errors"
	"github.com/olivere/elastic"
)

// Elastic configuration
type Elastic struct {
	URL      string
	Shards   int
	Replicas int
	Index    string
}

// EsStore is the elastic store handle
type EsStore struct {
	client *elastic.Client
	index  string
}

// NewElasticStore creates a new elastic store
func NewElasticStore(conf Elastic) (*EsStore, error) {
	ctx := context.Background()
	args := make([]elastic.ClientOptionFunc, 0)
	if len(conf.URL) > 0 {
		args = append(args, elastic.SetURL(conf.URL))
	}
	client, err := elastic.NewClient(args...)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	// useful for tests, to ensure index is empty
	//client.DeleteIndex(conf.Index).Do(ctx)
	ex, err := client.IndexExists(conf.Index).Do(ctx)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	if !ex {
		js := map[string]interface{}{
			"settings": map[string]interface{}{
				"number_of_shards":   conf.Shards,
				"number_of_replicas": conf.Replicas,
			},
			"mappings": map[string]interface{}{
				"doc": map[string]interface{}{
					"properties": map[string]interface{}{
						"item.id": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
			},
		}

		_, err := client.CreateIndex(conf.Index).BodyJson(js).Do(ctx)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
	}
	return &EsStore{client, conf.Index}, nil
}

// Close closes the store
func (es *EsStore) Close() error {
	if es != nil && es.client != nil {
		es.client.Stop()
		es.client = nil
	}
	return nil
}

// Read reads the latest version of an item
func (es *EsStore) Read(id ID) (Item, error) {
	var item = Item{}
	if es.client == nil {
		return item, NewStoreClosedError()
	}
	gr, err := es.client.Get().Index(es.index).Type("doc").Id(IDToString(id)).Do(context.Background())
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return item, nil
		}
		return item, errors.Wrap(err, 0)
	}
	return fromES(gr.Id, gr.Source)
}

// Write an item into Elastic
func (es *EsStore) Write(item Item) error {
	if item.IsEmpty() {
		return NewEmptyItemError()
	}
	if es.client == nil {
		return NewStoreClosedError()
	}
	body := toES(item)
	_, err := es.client.Index().Index(es.index).Type("doc").Id(IDToString(item.ID)).BodyJson(body).Refresh("true").
		Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func toES(item Item) map[string]interface{} {
	body := make(map[string]interface{})
	for k, v := range item.Contents {
		body[k] = v
	}
	body["item.name"] = item.Name
	body["item.type"] = item.Type
	body["item.id"] = IDToString(item.ID)
	return body
}

func fromES(id string, msg *json.RawMessage) (Item, error) {
	var item Item
	var fields = make(map[string]interface{})
	err := json.Unmarshal(*msg, &fields)
	if err != nil {
		return item, err
	}
	item.ID = StringToID(id)
	item.Name = fields["item.name"].(string)
	item.Type = fields["item.type"].(string)
	item.Contents = make(map[string]interface{})
	for k, v := range fields {
		if !strings.HasPrefix(k, "item.") {
			item.Contents[k] = v
		}
	}
	return item, nil
}

// Delete an item from Elastic
func (es *EsStore) Delete(id ID) error {
	if es.client == nil {
		return NewStoreClosedError()
	}
	_, err := es.client.Delete().Index(es.index).Type("doc").Id(IDToString(id)).Do(context.Background())
	if err != nil && !strings.Contains(err.Error(), "404") {
		return errors.Wrap(err, 0)
	}
	return nil
}

// Search inside Elastic
func (es *EsStore) Search(query Query) ([]Score, error) {
	var items []Score
	if es.client == nil {
		return items, NewStoreClosedError()
	}
	searchResult, err := es.client.Search(es.index).Type("doc").
		Query(elastic.NewQueryStringQuery(escapeQuery(query.QueryString))).
		Pretty(true).From(query.From).Size(query.Length).
		Do(context.Background())
	if err != nil {
		return items, err
	}
	// log.Printf("Found %d hits ", searchResult.TotalHits())
	var errors []string
	for _, hit := range searchResult.Hits.Hits {
		item, err := fromES(hit.Id, hit.Source)
		if err != nil {
			errors = append(errors, NewItemUnmarshallError(err).Error())
		} else {
			items = append(items, Score{item, *hit.Score})
		}
	}

	return items, NewMultipleItemErrors(errors)
}

// Scroll through elasticsearch result
func (es *EsStore) Scroll(query string, scoreChannel chan Score, errorChannel chan error) {
	defer close(scoreChannel)
	if es.client == nil {
		errorChannel <- NewStoreClosedError()
	}
	ctx := context.TODO()
	svc := es.client.Scroll(es.index).Type("doc").
		Query(elastic.NewQueryStringQuery(escapeQuery(query))).
		Pretty(true)
	for {
		res, err := svc.Do(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			errorChannel <- err
			break
		}
		for _, hit := range res.Hits.Hits {
			var sc float64
			if hit.Score != nil {
				sc = *(hit.Score)
			}
			item, err := fromES(hit.Id, hit.Source)
			if err == nil {
				select {
				case scoreChannel <- Score{item, sc}:
				case <-ctx.Done():
					errorChannel <- ctx.Err()
				}
			} else {
				select {
				case errorChannel <- err:
				case <-ctx.Done():
					errorChannel <- ctx.Err()
				}
			}
		}
	}
}

func escapeQuery(queryString string) string {
	s := strings.Replace(queryString, "/", "\\/", -1)
	return s
}
