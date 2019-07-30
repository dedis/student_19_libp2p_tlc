package channel

import (
	"testing"

	"github.com/dedis/student_19_libp2p_tlc/model"
	"github.com/dedis/student_19_libp2p_tlc/transport/test_utils"
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
			History:   make([]*model.Message, 0)}
	}
	return nodes
}

func TestChannel(t *testing.T) {
	nodes := setup(10)
	test_utils.StartTest(nodes, 10)
	test_utils.LogOutput(t, nodes)

}
