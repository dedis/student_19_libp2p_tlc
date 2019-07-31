package test_utils

import (
	"fmt"
	"github.com/dedis/student_19_libp2p_tlc/model"
	"sync"
	"testing"
)

func runNode(node *model.Node, stop int, wg *sync.WaitGroup) {
	defer wg.Done()
	node.WaitForMsg(stop)
}

// StartTest is used for starting tlc nodes
func StartTest(nodes []*model.Node, stop int) {
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

func LogOutput(t *testing.T, nodes []*model.Node) {
	for i := range nodes {
		t.Logf("nodes: %d , TimeStep : %d , %v", i, nodes[i].TimeStep, *(nodes[i].History[1]))
	}
}
