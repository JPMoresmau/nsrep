#!/bin/bash
CQL="CREATE KEYSPACE IF NOT EXISTS \"NSRep\" WITH replication = {'class':'SimpleStrategy','replication_factor':'1'};CREATE KEYSPACE IF NOT EXISTS \"NSRepTest\" WITH replication = {'class':'SimpleStrategy','replication_factor':'1'};"
until echo $CQL | cqlsh; do
  echo "cqlsh: Cassandra is unavailable to initialize - will retry later"
  sleep 5
done &
exec /docker-entrypoint.sh "$@"