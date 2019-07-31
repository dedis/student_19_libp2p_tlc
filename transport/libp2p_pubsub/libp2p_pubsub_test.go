package libp2p_pubsub

import (
	"fmt"
	"github.com/dedis/student_19_libp2p_tlc/transport/test_utils"
	"testing"
	"time"

	"github.com/dedis/student_19_libp2p_tlc/model"
	core "github.com/libp2p/go-libp2p-core"
)

// setupHosts is responsible for creating tlc nodes and also libp2p hosts.
func setupHosts(n int, initialPort int) ([]*model.Node, []*core.Host) {
	// nodes used in tlc model
	nodes := make([]*model.Node, n)
	// hosts used in libp2p communications
	hosts := make([]*core.Host, n)

	for i := range nodes {

		//var comm model.CommunicationInterface
		var comm *libp2pPubSub
		comm = new(libp2pPubSub)

		// creating libp2p hosts
		host := comm.createPeer(i, initialPort+i)
		hosts[i] = host
		// creating pubsubs
		comm.initializePubSub(*host)

		nodes[i] = &model.Node{
			Id:        i,
			TimeStep:  0,
			Threshold: n/2 + 1,
			Acks:      0,
			Comm:      comm,
			History:   make([]*model.Message, 0)}
	}
	return nodes, hosts
}

// setupNetworkTopology sets up a simple network topology for test.
func setupNetworkTopology(hosts []*core.Host) {

	// Connect hosts to each other in a topology
	// host0 ---- host1 ---- host2 ----- host3 ----- host4
	//	 			|		   				|    	   |
	//	            ------------------------------------
	connectHostToPeer(*hosts[1], getLocalhostAddress(*hosts[0]))
	connectHostToPeer(*hosts[2], getLocalhostAddress(*hosts[1]))
	connectHostToPeer(*hosts[3], getLocalhostAddress(*hosts[2]))
	connectHostToPeer(*hosts[4], getLocalhostAddress(*hosts[3]))
	connectHostToPeer(*hosts[4], getLocalhostAddress(*hosts[1]))
	connectHostToPeer(*hosts[3], getLocalhostAddress(*hosts[1]))
	connectHostToPeer(*hosts[4], getLocalhostAddress(*hosts[1]))

	// Wait so that subscriptions on topic will be done and all peers will "know" of all other peers
	time.Sleep(time.Second * 2)

}

func firstTest(t *testing.T, n int, initialPort int) {
	nodes, hosts := setupHosts(n, initialPort)

	defer func() {
		fmt.Println("Closing hosts")
		for _, h := range hosts {
			_ = (*h).Close()
		}
	}()

	setupNetworkTopology(hosts)

	// PubSub is ready and we can start our algorithm
	test_utils.StartTest(nodes, 10)
	test_utils.LogOutput(t, nodes)
}

func TestPubSub(t *testing.T) {
	// Create hosts in libp2p
	firstTest(t, 5, 9000)

}
