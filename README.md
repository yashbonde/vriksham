# Vriksham

Tree based conversation storage engine interface in go, currectly implmented in `neo4j` backend, see `main.go` for more
information.

## Interface

This entire repo is just one interface, this:

```go
type TreeEngine interface {
	// AddMessageToParent adds a message to the parent message
	// If `b` is empty, engine adds the message to the root
	AddMessage(threadId string, a Message, b *Message, ctx context.Context) error

	// Add an entire tree in the database
	AddTree(threadId string, tree ThreadTree, ctx context.Context) error

	// The number of leaves
	Breadth(threadId string, ctx context.Context) (int, error)

	// Delete a node and all children / relations from it
	// if message is empty, engine deletes the entire tree
	Delete(threadId string, message *Message, ctx context.Context) error

	// Get is returns the entire tree
	Get(threadId string, ctx context.Context) (ThreadTree, error)

	// GetChildren is returns the children of a particular node
	GetChildren(threadId string, message *Message, ctx context.Context) (ThreadTree, error)

	// LatestMessage is the latest added message to the tree
	GetLatestMessage(threadId string, ctx context.Context) (Message, error)

	// Pick returns a thread from a to b
	// If `a` is empty, engine picks from the root
	// If `a` and `b` are empty, engine picks the latest message
	Pick(threadId string, a *Message, b *Message, ctx context.Context) (Thread, error)

	// Sets a particular message as the latest message and returns the node with updated values
	SetLatestMessage(threadId string, latestMessage Message, ctx context.Context) (Message, error)

	// Number of nodes in the tree
	Size(threadId string, ctx context.Context) (int, error)
}
```

## Cheatsheet

Setup a few things for the database like:
- constraint for unique thread id
    ```go
    with neo4j.GraphDatabase.driver(URI, auth=AUTH) as driver:
        driver.execute_query(`CREATE CONSTRAINT constraint_thread_id_unique IF NOT EXISTS 
                            FOR (thread:ThreadRoot) 
                            REQUIRE thread.id IS UNIQUE`)
    ```
- install APOC from here: https://neo4j.com/labs/apoc/4.3/installation/

Some simple commands:

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

## Contribution ?

You can just copy [neo4j.go](impl/neo4j.go) and implement all the functions. To test add in [main.go](main.go) and leave
it there if it works.
