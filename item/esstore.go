package item

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/go-errors/errors"
	"github.com/olivere/elastic"
)

// Elastic configuration
type Elastic struct {
	URL      string
	Shards   int
	Replicas int
}

// EsStore is the elastic store handle
type EsStore struct {
	client *elastic.Client
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
	ex, err := client.IndexExists("items").Do(ctx)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	if !ex {
		js := map[string]interface{}{
			"settings": map[string]interface{}{
				"number_of_shards":   conf.Shards,
				"number_of_replicas": conf.Replicas,
			},
		}

		_, err := client.CreateIndex("items").BodyJson(js).Do(ctx)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
	}
	return &EsStore{client}, nil
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
func (es *EsStore) Read(id string) (Item, error) {
	var item = Item{}
	if es.client == nil {
		return item, NewStoreClosedError()
	}
	gr, err := es.client.Get().Index("items").Type("doc").Id(id).Do(context.Background())
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
	_, err := es.client.Index().Index("items").Type("doc").Id(item.ID).BodyJson(body).Refresh("true").
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
	// body["item.id"] = item.ID
	return body
}

func fromES(id string, msg *json.RawMessage) (Item, error) {
	var item Item
	var fields = make(map[string]interface{})
	err := json.Unmarshal(*msg, &fields)
	if err != nil {
		return item, err
	}
	item.ID = id
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
func (es *EsStore) Delete(id string) error {
	if es.client == nil {
		return NewStoreClosedError()
	}
	_, err := es.client.Delete().Index("items").Type("doc").Id(id).Do(context.Background())
	if err != nil {
		return errors.Wrap(err, 0)
	}
	return nil
}
