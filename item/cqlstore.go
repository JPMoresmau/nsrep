package item

import (
	"fmt"

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
	cluster.Port = config.Port
	cluster.Keyspace = "system"
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
