`nsrep` is a toy project of building a little database with a REST and GraphQL API in Go.

It stores `items`, which are objects with a name and a type, arbitrary contents and an ID that is made of string components (like a path). Like a file path, items whose id start with the id of another item are said to be in the second item's namespace, and will be deleted with the parent item is deleted. IDs have to be generated and provided by the client.

The base implementation uses Cassandra and ElasticSearch as the underlying data stores:
- Cassandra stores all versions of each item, including deletions, and provide history, since Cassandra writes are cheap
- ElasticSearch provides quick search capabilities on the current version of items

There is a base REST API to do CRUD on items, view their history and do a simple search. There is also a GraphQL API to do searches in the namespace structure.