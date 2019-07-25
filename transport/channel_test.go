package transport

import (
	"../model"
	"fmt"
	"sync"
	"testing"
)

func setup(n int) []*model.Node {
	nodes := make([]*model.Node, n)

	channelsMap := make(map[int]*chan *model.Message)

	for i := range nodes {
		c := make(chan *model.Message, n)
		channelsMap[i] = &c
	}

	for i := range nodes {

		var comm model.CommunicationInterface
		comm = &Channel{&channelsMap, channelsMap[i]}
		nodes[i] = &model.Node{
			Id:        i,
			TimeStep:  0,
			Threshold: n/2 + 1,
			Acks:      0,
			Comm:      comm,
			History:   make([]model.Message, 0)}
	}
	return nodes
}

func runNode(node *model.Node, stop int, wg *sync.WaitGroup) {
	defer wg.Done()
	node.WaitForMsg(stop)
}

func startTest(nodes []*model.Node, stop int) {
	wg := &sync.WaitGroup{}

	for _, node := range nodes {
		node.Advance(0)
	}
	for _, node := range nodes {
		wg.Add(1)
		go runNode(node, stop, wg)
	}

	wg.Wait()
	fmt.Println("The END")
}

func logOutput(t *testing.T, nodes []*model.Node) {
	for i := range nodes {
		t.Logf("nodes: %d , TimeStep : %d , History: %v", i, nodes[i].TimeStep, nodes[i].History)
	}
}

func TestChannel(t *testing.T) {
	nodes := setup(3)
	startTest(nodes, 10)
	logOutput(t, nodes)

}
