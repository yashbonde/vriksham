package main

import (
	"context"
	"fmt"

	Impl "github.com/yashbonde/vriksham/impl"
)

/*
Here's a simple thread we are going to use to load the tree

// <ThreadTree: tree_0000>
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
*/
func GetDemoTree() Impl.ThreadTree {
	messages := []Impl.Message{}
	for i := 0; i < 28; i++ {
		messages = append(messages, Impl.Message{MessageId: fmt.Sprintf("msg_%02d", i)})
	}
	relations := []Impl.Triple{
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
	messages[len(messages)-1].LatestMessage = true
	return Impl.ThreadTree{
		Root:      Impl.ThreadRoot{ThreadId: "tree_0000"},
		Messages:  messages,
		Relations: relations,
	}
}

func run(backend Impl.TreeEngine) {
	ctx := context.Context(context.Background())
	backend.Connect(ctx)

	// Writing
	demoTree := GetDemoTree()
	// err := backend.AddTree(demoTree.Root.ThreadId, demoTree, ctx)
	// err := backend.AddMessage(demoTree.Root.ThreadId, Impl.Message{MessageId: "new_00"}, nil, ctx)
	// err := backend.AddMessage(demoTree.Root.ThreadId, Impl.Message{MessageId: "new_01"}, &Impl.Message{MessageId: "new_00"}, ctx)

	// Querying
	//
	out, err := backend.Get(demoTree.Root.ThreadId, ctx)
	// out, err := backend.Breadth(demoTree.Root.ThreadId, ctx)
	// out, err := backend.Size(demoTree.Root.ThreadId, ctx)
	// out, err := backend.Pick(demoTree.Root.ThreadId, nil, &Impl.Message{MessageId: "msg_27"}, ctx)
	// out, err := backend.Pick(demoTree.Root.ThreadId, &Impl.Message{MessageId: "msg_06"}, &Impl.Message{MessageId: "msg_27"}, ctx)

	// Deleting
	//
	// err := backend.Delete(demoTree.Root.ThreadId, nil, ctx)
	// err := backend.Delete(demoTree.Root.ThreadId, &Impl.Message{MessageId: "new_00"}, ctx)

	if err != nil {
		panic(err)
	} else {
		fmt.Println("Success")
	}
	fmt.Println(out)
}

func main() {
	run(Impl.Backend_Neo4j{
		DbUrl:    "bolt://localhost:7687/neo4j",
		AuthUser: "neo4j",
		AuthPass: "password",
	})
}
