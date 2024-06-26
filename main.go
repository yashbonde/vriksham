package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	Impl "github.com/yashbonde/vriksham/impl"
)

func run(backend Impl.TreeEngine, ctx context.Context) {
	// get tree and all
	demoTree := Impl.GetDemoTree()
	threadId := *threadId
	if threadId == "" {
		threadId = demoTree.Root.ThreadId
	}
	log.Println("Thread ID: ", threadId)

	// Writing
	// err := backend.AddTree(threadId, *demoTree, ctx)
	// err := backend.AddMessage(threadId, Impl.Message{MessageId: "new_00"}, nil, ctx)
	// err := backend.AddMessage(threadId, Impl.Message{MessageId: "new_01"}, &Impl.Message{MessageId: "new_00"}, ctx)

	// Querying
	//
	// out, err := backend.Get(threadId, ctx)
	// out, err := backend.GetLatestMessage(threadId, ctx)
	// out, err := backend.SetLatestMessage(threadId, &demoTree.Messages[1], ctx)
	// out, err := backend.GetChildren(threadId, nil, 1, ctx)
	// out, err := backend.GetChildren(threadId, &Impl.Message{MessageId: messageId}, 1, ctx)
	// out, err := backend.Breadth(threadId, ctx)
	// out, err := backend.Size(threadId, ctx)
	// out, err := backend.Depth(threadId, ctx)
	// out, err := backend.Degree(threadId, &Impl.Message{MessageId: "msg_00"}, ctx)
	// out, err := backend.Degree(threadId, nil, ctx)
	// out, err := backend.Pick(threadId, nil, nil, ctx)
	// out, err := backend.Pick(threadId, nil, &Impl.Message{MessageId: "msg_27"}, ctx)
	// out, err := backend.Pick(threadId, &Impl.Message{MessageId: "msg_06"}, &Impl.Message{MessageId: "msg_27"}, ctx)

	// Deleting
	//
	err := backend.Delete(threadId, nil, ctx)
	// err := backend.Delete(threadId, &Impl.Message{MessageId: "new_00"}, ctx)

	if err != nil {
		panic(err)
	} else {
		fmt.Println("Success")
	}
	// fmt.Println(out)
}

var (
	// fix the command below:
	threadId = flag.String("t", "", "Thread ID")
)

func main() {
	flag.Parse()

	// create a context
	ctx := context.Context(context.Background())

	// Use the Neo4j backend
	backend := Impl.Backend_Neo4j{
		DbUrl:    "bolt://localhost:7687/neo4j",
		AuthUser: "neo4j",
		AuthPass: "password",
	}
	backend.Connect(ctx)

	// Run code
	run(backend, ctx)
}
