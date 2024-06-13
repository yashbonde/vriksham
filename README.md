# Vriksham

Tree based conversation storage engine interface in go, currectly implmented in `neo4j` backend, see `main.go` for more
information.


```cypher
// find all nodes without any node label
MATCH (n) WHERE size(labels(n)) = 0 RETURN n

// delete a node from it's ID
MATCH (n) WHERE ID(n) in [47223] DETACH DELETE n
MATCH (n) WHERE ID(n) in [47223, 47224] DETACH DELETE n

// All nodes that are not ThreadRoot and do not have any inputs
MATCH (a:!ThreadRoot) WHERE NOT (a)<-[]-() RETURN a

// All threads that do not have any children
MATCH (a:ThreadRoot) WHERE NOT (a)-[]->() RETURN a
```