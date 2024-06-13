package impl

import "context"

/*
Here are the data structures that are used for storage and API calls. A speed run through them:

- ThreadRoot: This is a special node that contains the thread_id and is the root of the tree.
- Message: This is a node that contains the message_id and some attributes like is it the latest message.
- Thread: Thread is a list of messages
- Triple: This is a relation between two nodes, it is a directed edge from startId to endId with a relation.
- ThreadTree: This is the entire tree, it contains the thread_id, messages and relations.
*/

type ThreadRoot struct {
	ThreadId string `json:"thread_id"`
}

type Message struct {
	MessageId     string `json:"id"`
	ThreadId      string `json:"thread_id"`
	LatestMessage bool   `json:"latest_message"`
}

func MessageFromDict(dict map[string]interface{}) Message {
	m := Message{}
	if id := dict["id"]; id != nil {
		m.MessageId = id.(string)
	}
	if threadId := dict["thread_id"]; threadId != nil {
		m.ThreadId = threadId.(string)
	}
	if latestMessage := dict["latest_message"]; latestMessage != nil {
		m.LatestMessage = latestMessage.(bool)
	}
	return m
}

type Thread struct {
	Messages []Message `json:"messages"`
}

type Triple struct {
	StartId  string `json:"start_id"`
	Relation string `json:"relation"`
	EndId    string `json:"end_id"`
}

type ThreadTree struct {
	Root      ThreadRoot `json:"root"`
	Messages  []Message  `json:"messages"`
	Relations []Triple   `json:"relations"`
}

/*
The idea behind the TreeEngine is that it provides a generic interface for day to day things that are required from a
tree data structure. This engine will be implmented by several backends and the user can choose the backend that suits
their needs the best.
*/

type TreeEngine interface {
	// Connect is used to connect to the database
	Connect(driver interface{}, ctx context.Context) error

	// AddMessageToParent adds a message to the parent message
	// If `b` is empty, engine adds the message to the root
	AddMessage(threadId string, a, b Message, ctx context.Context) (Thread, error)

	// Add an entire tree in the database
	AddTree(threadId string, tree ThreadTree, ctx context.Context) error

	// The number of leaves
	Breadth(threadId string, ctx context.Context) (int32, error)

	// Delete a node and all children / relations from it
	Delete(threadId string, message Message, ctx context.Context) error

	// Delete a node and all children / relations from it
	DeleteTree(threadId string, ctx context.Context) error

	// Get is returns the entire tree
	Get(threadId string, ctx context.Context) (ThreadTree, error)

	// GetChildren is returns the children of a particular node
	GetChildren(threadId string, message Message, ctx context.Context) (ThreadTree, error)

	// LatestMessage is the latest added message to the tree
	GetLatestMessage(threadId string, ctx context.Context) (Message, error)

	// Pick returns a thread from one point to another
	// If `a` is empty, engine picks from the root
	// If `a` and `b` are empty, engine picks the latest message
	Pick(threadId string, a, b Message, ctx context.Context) (Thread, error)

	// Sets a particular message as the latest message and returns the node with updated values
	SetLatestMessage(threadId string, latestMessage Message, ctx context.Context) (Message, error)

	// Number of nodes in the tree
	Size(threadId string, ctx context.Context) (int32, error)
}
