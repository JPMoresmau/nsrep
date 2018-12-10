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
	if err != nil {
		return errors.Wrap(err, 0)
	}
	return nil
}

// Read reads the latest version of an item
func (s *CqlStore) Read(id string) (Item, error) {
	var item Item
	sts, err := s.History(id, 1)
	if err != nil {
		return item, err
	}
	if len(sts) == 1 && sts[0].Status == "ALIVE" {
		item = sts[0].Item
	}
	return item, nil
}

// History reads the history of a given item
func (s *CqlStore) History(id string, limit int) ([]Status, error) {
	if s.session == nil {
		return []Status{}, NewStoreClosedError()
	}
	var items []Status
	var errors []string
	iter := s.session.Query("select status, type, name, contents from items where id=? order by updated desc limit ?", id, limit).Iter()
	var status, ttype, name, contents string
	for iter.Scan(&status, &ttype, &name, &contents) {
		var cnts map[string]interface{}
		var err error
		if len(contents) > 0 {
			err = json.Unmarshal([]byte(contents), &cnts)
		}
		if err != nil {
			errors = append(errors, NewItemUnmarshallError(err).Error())
		} else {
			items = append(items, Status{Item{id, ttype, name, cnts}, status})
		}

	}
	if err := iter.Close(); err != nil {
		errors = append(errors, NewStoreInternalError(err).Error())
	}
	return items, NewMultipleItemErrors(errors)
}

// Delete marks an item as deleted
func (s *CqlStore) Delete(id string) error {
	if s.session == nil {
		return NewStoreClosedError()
	}
	err := s.session.Query("insert into items (id, updated, status) values(?,now(),?)",
		id, "DELETED").Exec()
	if err != nil {
		return errors.Wrap(err, 0)
	}
	return nil
}
