package impl

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Backend_Neo4j struct {
	DbUrl    string `json:"db_url"`
	AuthUser string `json:"auth_user"`
	AuthPass string `json:"auth_pass"`
	driver   neo4j.DriverWithContext
}

func (backend *Backend_Neo4j) Connect(ctx context.Context) error {
	backend.DbUrl = "neo4j://localhost" // scheme://host(:port) (default port is 7687)
	driver, err := neo4j.NewDriverWithContext(
		backend.DbUrl,
		neo4j.BasicAuth(backend.AuthUser, backend.AuthPass, ""),
	)
	if err != nil {
		return err
	}
	backend.driver = driver
	return nil
}

// implement interface

func (db Backend_Neo4j) AddMessage(threadId string, a, b *Message, ctx context.Context) error {
	if a == nil {
		return fmt.Errorf("message to be inserted cannot be empty")
	}
	addToRoot := b == nil
	query := ""
	parentId := ""
	if addToRoot {
		parentId = threadId
		query += "MATCH (parent:ThreadRoot {thread_id: $parentId})\n"
	} else {
		parentId = b.MessageId
		query += "MATCH (parent:Message {id: $parentId})\n"
	}
	query += "MERGE (child:Message {id: $childId})\n"
	query += "MERGE (parent)-[:CHILD]->(child)\n"
	fullData := map[string]any{
		"parentId": parentId,
		"childId":  a.MessageId,
	}
	// fmt.Println(query)
	// fmt.Println(fullData)

	// execute query and get results
	result, err := neo4j.ExecuteQuery(
		ctx,
		db.driver,
		query,
		fullData,
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return err
	}

	fmt.Printf("Created %v nodes in %+v.\n",
		result.Summary.Counters().NodesCreated(),
		result.Summary.ResultAvailableAfter(),
	)
	if result.Summary.Counters().NodesCreated() == 0 {
		return fmt.Errorf("no nodes created, does the parent exist?")
	}

	return nil
}

func (db Backend_Neo4j) AddTree(threadId string, tree ThreadTree, ctx context.Context) error {
	// validations
	if tree.Root.ThreadId != threadId {
		return fmt.Errorf("threadId mismatch")
	} else if len(tree.Messages) == 0 {
		return fmt.Errorf("no messages in the tree")
	} else if len(tree.Relations) == 0 {
		return fmt.Errorf("no relations in the tree")
	}

	query := ""
	fullData := map[string]any{"threadId": tree.Root.ThreadId}
	messageIdToQueryId := map[string]string{}
	query += "MERGE (root:ThreadRoot {thread_id: $threadId})\n"
	for i, m := range tree.Messages {
		fullData[fmt.Sprintf("m%d_id", i)] = m.MessageId
		messageIdToQueryId[m.MessageId] = fmt.Sprintf("m%d", i)
		if !m.Latest {
			query += fmt.Sprintf("MERGE (m%d:Message {id: $m%d_id})\n", i, i)
		} else {
			query += fmt.Sprintf("MERGE (m%d:Message {id: $m%d_id, latest: true})\n", i, i)
		}
	}

	for _, r := range tree.Relations {
		startQueryId := messageIdToQueryId[r.StartId]
		endQueryId := messageIdToQueryId[r.EndId]
		if r.StartId == "" {
			query += fmt.Sprintf("MERGE (root)-[:CHILD]->(%s)\n", endQueryId)
		} else {
			query += fmt.Sprintf("MERGE (%s)-[:CHILD]->(%s)\n", startQueryId, endQueryId)
		}
	}

	// fmt.Println(query)
	// fmt.Println(fullData)

	result, err := neo4j.ExecuteQuery(
		ctx,
		db.driver,
		query,
		fullData,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created %v nodes in %+v.\n",
		result.Summary.Counters().NodesCreated(),
		result.Summary.ResultAvailableAfter())

	return nil
}

func (db Backend_Neo4j) Breadth(threadId string, ctx context.Context) (int, error) {
	output := 0
	result, err := neo4j.ExecuteQuery(
		ctx,
		db.driver,
		`
		MATCH (t:ThreadRoot {thread_id: $threadId})-[:CHILD*0..]->(c:Message)
		WHERE NOT (c)-[:CHILD]->()
		RETURN COUNT(c) as count
		`,
		map[string]any{
			"threadId": threadId,
		},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return output, err
	}
	for _, record := range result.Records {
		count, _ := record.Get("count")
		output = int(count.(int64))
	}
	return output, nil
}

func (db Backend_Neo4j) Degree(threadId string, message *Message, ctx context.Context) (int, error) {
	fullData := map[string]any{}
	var query string
	if message == nil {
		fullData["startId"] = threadId
		query = "MATCH (t:ThreadRoot {thread_id: $startId})-[:CHILD]->(c:Message) RETURN COUNT(c) as count"
	} else {
		fullData["startId"] = message.MessageId
		query = "MATCH (m:Message {id: $startId})-[:CHILD]->(c:Message) RETURN COUNT(c) as count"
	}

	output := 0
	result, err := neo4j.ExecuteQuery(
		ctx,
		db.driver,
		query,
		fullData,
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return output, err
	}
	for _, record := range result.Records {
		count, _ := record.Get("count")
		output = int(count.(int64))
	}
	return output, nil
}

func (db Backend_Neo4j) Delete(threadId string, message *Message, ctx context.Context) error {
	fromRoot := message == nil
	query := ""
	startId := ""
	if fromRoot {
		query += "MATCH (t:ThreadRoot {thread_id: $startId})"
		startId = threadId
	} else {
		query += "MATCH (m:Message {id: $startId})"
		startId = message.MessageId
	}
	query += "-[*0..]->(n:Message) DETACH DELETE n"
	if fromRoot {
		query += ", t"
	}

	result, err := neo4j.ExecuteQuery(
		ctx,
		db.driver,
		query,
		map[string]any{
			"startId": startId,
		},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Deleted %v nodes in %+v.\n",
		result.Summary.Counters().NodesDeleted(),
		result.Summary.ResultAvailableAfter())

	return nil
}

func (db Backend_Neo4j) Depth(threadId string, ctx context.Context) (int, error) {
	output := 0
	result, err := neo4j.ExecuteQuery(
		ctx,
		db.driver,
		`
		MATCH p=(t:ThreadRoot {thread_id: $threadId})-[:CHILD*0..]->(c:Message)
		WHERE NOT (c)-[:CHILD]->()
		RETURN  LENGTH(p) as depth
		ORDER BY LENGTH(p) DESC
		LIMIT 1;
		`,
		map[string]any{
			"threadId": threadId,
		},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return output, nil
	}
	for _, record := range result.Records {
		depth, _ := record.Get("depth")
		output = int(depth.(int64))
	}
	return output, nil
}

func (db Backend_Neo4j) Get(threadId string, ctx context.Context) (ThreadTree, error) {
	output := ThreadTree{}
	result, err := neo4j.ExecuteQuery(
		ctx,
		db.driver,
		`
			MATCH r=(ThreadRoot {thread_id: $threadId})-[:CHILD*0..100]->(c:Message)
			WITH apoc.agg.graph(r) AS g
			RETURN g.nodes AS nodes, g.relationships AS edges;
		`,
		map[string]any{
			"threadId": threadId,
		},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return output, err
	}
	elementMessages := map[string]Message{}
	for _, record := range result.Records {
		nodes, _ := record.Get("nodes")
		relations, _ := record.Get("edges")
		for _, n := range nodes.([]interface{}) {
			node := n.(neo4j.Node)
			m := MessageFromDict(node.GetProperties())
			if m.MessageId != "" {
				output.Messages = append(output.Messages, m)
			}
			elementMessages[node.GetElementId()] = m
		}

		for _, r := range relations.([]interface{}) {
			relation := r.(neo4j.Relationship)
			startMessage := elementMessages[relation.StartElementId]
			endMessage := elementMessages[relation.EndElementId]
			output.Relations = append(output.Relations, Triple{
				StartId:  startMessage.MessageId,
				Relation: relation.Type,
				EndId:    endMessage.MessageId,
			})
		}
	}
	if len(output.Messages) > 0 && len(output.Relations) > 0 {
		output.Root = ThreadRoot{ThreadId: threadId}
	} else {
		return output, fmt.Errorf("no root found, does this thread exist?")
	}
	return output, nil
}

func (db Backend_Neo4j) GetChildren(threadId string, message *Message, depth int, ctx context.Context) (ThreadTree, error) {
	output := ThreadTree{}
	if depth <= 0 {
		return output, fmt.Errorf("depth cannot be less than 1")
	} else if depth > 10 {
		return output, fmt.Errorf("depth cannot be more than 10")
	} else if depth == 1 {
		depth = 2
	}
	query := ""
	startId := ""
	if message == nil {
		query += "MATCH r= (t:ThreadRoot {thread_id: $startId})"
		startId = threadId
	} else {
		query += "MATCH r= (m:Message {id: $startId})"
		startId = message.MessageId
	}
	query += fmt.Sprintf("-[:CHILD*0..%d]->(c:Message)\n", depth-1)
	query += "WITH apoc.agg.graph(r) AS g RETURN g.nodes AS nodes, g.relationships AS edges;"
	fmt.Println(query)
	result, err := neo4j.ExecuteQuery(
		ctx,
		db.driver,
		query,
		map[string]any{"startId": startId},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return output, err
	}
	elementMessages := map[string]Message{}
	for _, record := range result.Records {
		nodes, _ := record.Get("nodes")
		relations, _ := record.Get("edges")
		for _, n := range nodes.([]interface{}) {
			node := n.(neo4j.Node)
			m := MessageFromDict(node.GetProperties())
			if m.MessageId != "" {
				output.Messages = append(output.Messages, m)
			}
			elementMessages[node.GetElementId()] = m
		}

		for _, r := range relations.([]interface{}) {
			relation := r.(neo4j.Relationship)
			startMessage := elementMessages[relation.StartElementId]
			endMessage := elementMessages[relation.EndElementId]
			output.Relations = append(output.Relations, Triple{
				StartId:  startMessage.MessageId,
				Relation: relation.Type,
				EndId:    endMessage.MessageId,
			})
		}
	}
	if len(output.Messages) > 0 && len(output.Relations) > 0 {
		output.Root = ThreadRoot{ThreadId: threadId}
	} else {
		return output, fmt.Errorf("no root found, does this thread exist?")
	}
	return output, nil
}

func (db Backend_Neo4j) GetLatestMessage(threadId string, ctx context.Context) (Message, error) {
	output := Message{}
	result, err := neo4j.ExecuteQuery(
		ctx,
		db.driver,
		"MATCH r=(t:ThreadRoot {thread_id: $threadId})-[:CHILD*..100]->(c:Message {latest: true}) RETURN c",
		map[string]any{"threadId": threadId},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return output, err
	}
	for _, record := range result.Records {
		node, _ := record.Get("c")
		output = MessageFromDict(node.(neo4j.Node).GetProperties())
	}
	if output.MessageId == "" {
		return output, fmt.Errorf("no latest message found")
	}
	return output, nil
}

func (db Backend_Neo4j) Pick(threadId string, a *Message, b *Message, ctx context.Context) (Thread, error) {
	output := Thread{}
	fromRoot := a == nil
	uptoLatest := b == nil

	startId := ""
	toMessageId := ""
	query := "MATCH p = shortestPath("
	if fromRoot {
		query += "(t: ThreadRoot {thread_id: $startId})"
		startId = threadId
	} else {
		query += "(m0: Message {id: $startId})"
		startId = a.MessageId
	}
	query += "-[CHILD*..40]->"
	if uptoLatest {
		query += "(m1: Message {latest: true}))"
		toMessageId = ""
	} else {
		query += "(m1: Message {id: $toMessageId}))"
		toMessageId = b.MessageId
	}
	query += "RETURN nodes(p) as nodes, relationships(p) as edges"

	// execute query and get results
	result, err := neo4j.ExecuteQuery(
		ctx,
		db.driver,
		query,
		map[string]any{
			"startId":     startId,
			"toMessageId": toMessageId,
		},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return output, err
	}
	elementMessages := map[string]Message{}
	for _, record := range result.Records {
		nodes, _ := record.Get("nodes")
		relations, _ := record.Get("edges")
		for _, n := range nodes.([]interface{}) {
			node := n.(neo4j.Node)
			elementMessages[node.GetElementId()] = MessageFromDict(node.GetProperties())
		}
		for rid, r := range relations.([]interface{}) {
			relation := r.(neo4j.Relationship)
			startMessage := elementMessages[relation.StartElementId]
			if !(fromRoot && rid == 0) {
				output.Messages = append(output.Messages, startMessage)
			}
			if rid == len(relations.([]interface{}))-1 {
				endMessage := elementMessages[relation.EndElementId]
				output.Messages = append(output.Messages, endMessage)
			}
		}
	}
	return output, nil
}

func (db Backend_Neo4j) SetLatestMessage(threadId string, latestMessage *Message, ctx context.Context) (Message, error) {
	output := Message{}
	if latestMessage == nil {
		return output, fmt.Errorf("latest message cannot be empty")
	}
	result, err := neo4j.ExecuteQuery(
		ctx,
		db.driver,
		`
		MATCH (t:ThreadRoot {thread_id: $threadId})-[:CHILD*0..]->(c:Message)
		SET c.latest = false
		WITH c
		WHERE c.id = $latestMessageId
		SET c.latest = true
		RETURN c
		`,
		map[string]any{
			"threadId":        threadId,
			"latestMessageId": latestMessage.MessageId,
		},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return output, err
	}
	for _, record := range result.Records {
		node, _ := record.Get("c")
		output = MessageFromDict(node.(neo4j.Node).GetProperties())
	}
	if output.MessageId == "" {
		return output, fmt.Errorf("no latest message found")
	}
	return output, nil
}

func (db Backend_Neo4j) Size(threadId string, ctx context.Context) (int, error) {
	output := 0
	result, err := neo4j.ExecuteQuery(
		ctx,
		db.driver,
		`
		MATCH r=(ThreadRoot {thread_id: $threadId})-[:CHILD*0..]->(c:Message)
		RETURN COUNT(nodes(r)) as count
		`,
		map[string]any{
			"threadId": threadId,
		},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return output, err
	}
	for _, record := range result.Records {
		count, _ := record.Get("count")
		output = int(count.(int64))
	}
	return output, nil
}
