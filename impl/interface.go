package impl

import (
	"context"
	"fmt"
)

/*
Here are the data structures that are used for storage and API calls. A speed run through them:

- ThreadRoot: This is a special node that contains the thread_id and is the root of the tree.
- Message: This is a node that contains the message_id and some attributes like is it the latest message.
- Thread: Thread is a list of messages
- Triple: This is a relation between two nodes, it is a directed edge from startId to endId with a relation.
- ThreadTree: This is the entire tree, it contains the thread_id, messages and relations.

After the types there is an example of tree that shows an example.
*/

type ThreadRoot struct {
	ThreadId string `json:"thread_id"`
}

type Message struct {
	MessageId string `json:"id"`
	Latest    bool   `json:"latest"`
}

func MessageFromDict(dict map[string]interface{}) Message {
	m := Message{}
	if id := dict["id"]; id != nil {
		m.MessageId = id.(string)
	}
	if latestMessage := dict["latest"]; latestMessage != nil {
		m.Latest = latestMessage.(bool)
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
Here's a simple thread we are going to use to load the tree

// <ThreadRoot: tree_0000>
// ├── [msg_00]
// │   ╰── [msg_06]
// │       ├── [msg_14]
// │       │   ╰── [msg_15]
// │       │       ╰── [msg_22]
// │       │           ╰── [msg_23]
// │       │               ├── [msg_24]
// │       │               │   ╰── [msg_25]
// │       │               ╰── [msg_26]
// │       │                   ╰── [msg_27]
// │       ╰── [msg_16]
// │           ╰── [msg_17]
// │               ├── [msg_18]
// │               │   ╰── [msg_19]
// │               ╰── [msg_20]
// │                   ╰── [msg_21]
// ├── [msg_01]
// │   ╰── [msg_07]
// ├── [msg_02]
// │   ╰── [msg_08]
// ├── [msg_03]
// │   ╰── [msg_09]
// ├── [msg_04]
// │   ╰── [msg_10]
// ╰── [msg_05]
//     ╰── [msg_11]
//         ╰── [msg_12]
//             ╰── [msg_13]

Here the ThreadRoot contains the ThreadId and the root of the tree. The tree has 28 messages.

> For a tree with ( n ) nodes, the total number of edges is ( n - 1 ).

So it has 27 edges.
*/
func GetDemoTree() *ThreadTree {
	messages := []Message{}
	for i := 0; i < 28; i++ {
		messages = append(messages, Message{MessageId: fmt.Sprintf("msg_%02d", i)})
	}
	relations := []Triple{
		{Relation: "CHILD", EndId: "msg_00"},
		{Relation: "CHILD", EndId: "msg_01"},
		{Relation: "CHILD", EndId: "msg_02"},
		{Relation: "CHILD", EndId: "msg_03"},
		{Relation: "CHILD", EndId: "msg_04"},
		{Relation: "CHILD", EndId: "msg_05"},
		{StartId: "msg_00", Relation: "CHILD", EndId: "msg_06"},
		{StartId: "msg_06", Relation: "CHILD", EndId: "msg_14"},
		{StartId: "msg_14", Relation: "CHILD", EndId: "msg_15"},
		{StartId: "msg_15", Relation: "CHILD", EndId: "msg_22"},
		{StartId: "msg_22", Relation: "CHILD", EndId: "msg_23"},
		{StartId: "msg_23", Relation: "CHILD", EndId: "msg_24"},
		{StartId: "msg_24", Relation: "CHILD", EndId: "msg_25"},
		{StartId: "msg_23", Relation: "CHILD", EndId: "msg_26"},
		{StartId: "msg_26", Relation: "CHILD", EndId: "msg_27"},
		{StartId: "msg_06", Relation: "CHILD", EndId: "msg_16"},
		{StartId: "msg_16", Relation: "CHILD", EndId: "msg_17"},
		{StartId: "msg_17", Relation: "CHILD", EndId: "msg_18"},
		{StartId: "msg_18", Relation: "CHILD", EndId: "msg_19"},
		{StartId: "msg_17", Relation: "CHILD", EndId: "msg_20"},
		{StartId: "msg_20", Relation: "CHILD", EndId: "msg_21"},
		{StartId: "msg_01", Relation: "CHILD", EndId: "msg_07"},
		{StartId: "msg_02", Relation: "CHILD", EndId: "msg_08"},
		{StartId: "msg_03", Relation: "CHILD", EndId: "msg_09"},
		{StartId: "msg_04", Relation: "CHILD", EndId: "msg_10"},
		{StartId: "msg_05", Relation: "CHILD", EndId: "msg_11"},
		{StartId: "msg_11", Relation: "CHILD", EndId: "msg_12"},
		{StartId: "msg_12", Relation: "CHILD", EndId: "msg_13"},
	}
	messages[len(messages)-1].Latest = true
	return &ThreadTree{
		Root:      ThreadRoot{ThreadId: "tree_0000"},
		Messages:  messages,
		Relations: relations,
	}
}

/*
The idea behind the TreeEngine is that it provides a generic interface for day to day things that are required from a
tree data structure. This engine will be implmented by several backends and the user can choose the backend that suits
their needs the best.
*/

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
	// if `message` is empty, engine returns the children of the root
	// maximum `depth` is 10
	GetChildren(threadId string, message *Message, depth int, ctx context.Context) (ThreadTree, error)

	// LatestMessage is the latest added message to the tree
	GetLatestMessage(threadId string, ctx context.Context) (Message, error)

	// Pick returns a thread from a to b
	// If `a` is empty, engine picks from the root
	// If `b` is empty, engine picks upto latest message
	Pick(threadId string, a, b *Message, ctx context.Context) (Thread, error)

	// Sets a particular message as the latest message and returns the node with updated values
	SetLatestMessage(threadId string, latestMessage *Message, ctx context.Context) (Message, error)

	// Number of nodes in the tree
	Size(threadId string, ctx context.Context) (int, error)
}
