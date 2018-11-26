package item

import (
	"encoding/json"

	"github.com/go-errors/errors"
	"github.com/gocql/gocql"
)

// Cassandra connection information
type Cassandra struct {
	Port        int
	Keyspace    string
	Endpoints   []string
	Replication int
}

// CqlStore is a store using Cassandra
type CqlStore struct {
	session *gocql.Session
}

// NewCqlStore creates a new Cassandra Store
func NewCqlStore(config Cassandra) (*CqlStore, error) {
	cluster := gocql.NewCluster(config.Endpoints...)
	cluster.ProtoVersion = 4
	if config.Port > 0 {
		cluster.Port = config.Port
	}
	/*cluster.Keyspace = "system"
	controlSession, err := cluster.CreateSession()
	if err != nil {
		return nil, NewStoreCreationError(err)
	}
	defer controlSession.Close()
	repl := 3
	if config.Replication > 0 {
		repl = config.Replication
	}
	err = controlSession.Query(fmt.Sprintf(
		`create keyspace if not exists "%s" with replication = {'class':'SimpleStrategy','replication_factor':%d}`,
		config.Keyspace, repl)).Exec()
	if err != nil {
		return nil, NewStoreCreationError(err)
	}
	*/

	cluster.Keyspace = config.Keyspace
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, NewStoreCreationError(err)
	}
	err = session.Query("create table if not exists items ( id text, updated uuid, status text, type text, name text, contents text,primary key (id, updated)) WITH CLUSTERING ORDER BY (updated DESC)").
		Exec()
	if err != nil {
		return nil, NewStoreCreationError(err)
	}
	return &CqlStore{session}, nil
}

//Close the store
func (s *CqlStore) Close() error {
	if s.session != nil {
		s.session.Close()
		s.session = nil
	}
	return nil
}

// Write stores an item in the store
func (s *CqlStore) Write(item Item) error {
	if item.IsEmpty() {
		return NewEmptyItemError()
	}
	if s.session == nil {
		return NewStoreClosedError()
	}
	b, err := json.Marshal(item.Contents)
	if err != nil {
		return NewItemMarshallError(err)
	}
	err = s.session.Query("insert into items (id, updated, status, type, name, contents) values(?,now(),?,?,?,?)",
		item.ID, "ALIVE", item.Type, item.Name, string(b)).Exec()
	return errors.Wrap(err, 1)
}

// Read reads the latest version of an item
func (s *CqlStore) Read(id string) (Item, error) {
	var item = Item{}
	if s.session == nil {
		return item, NewStoreClosedError()
	}
	iter := s.session.Query("select status, type, name, contents from items where id=? order by updated desc limit 1", id).Iter()
	var status, ttype, name, contents string
	if iter.Scan(&status, &ttype, &name, &contents) {
		if status == "ALIVE" {
			var cnts interface{}
			err := json.Unmarshal([]byte(contents), &cnts)
			if err != nil {
				return item, NewItemUnmarshallError(err)
			}
			item = Item{id, ttype, name, cnts.(map[string]interface{})}
		}
	}
	return item, nil
}

// Delete marks an item as deleted
func (s *CqlStore) Delete(id string) error {
	if s.session == nil {
		return NewStoreClosedError()
	}
	err := s.session.Query("insert into items (id, updated, status) values(?,now(),?)",
		id, "DELETED").Exec()
	return errors.Wrap(err, 1)
}
