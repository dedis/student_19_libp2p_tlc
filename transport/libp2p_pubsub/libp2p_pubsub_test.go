package libp2p_pubsub

import (
	"fmt"
	"github.com/dedis/student_19_libp2p_tlc/transport/test_utils"
	"log"
	"os"
	"testing"
	"time"

	"github.com/dedis/student_19_libp2p_tlc/model"
	core "github.com/libp2p/go-libp2p-core"
)

type FailureModel int

const (
	NoFailure = iota
	MinorFailure
	MajorFailure
	RejoiningMinorityFailure
	RejoiningMajorityFailure
)

// setupHosts is responsible for creating tlc nodes and also libp2p hosts.
func setupHosts(n int, initialPort int, failureModel FailureModel) ([]*model.Node, []*core.Host) {
	// nodes used in tlc model
	nodes := make([]*model.Node, n)
	// hosts used in libp2p communications
	hosts := make([]*core.Host, n)

	for i := range nodes {

		//var comm model.CommunicationInterface
		var comm *libp2pPubSub
		comm = new(libp2pPubSub)
		comm.topic = "TLC"

		// creating libp2p hosts
		host := comm.createPeer(i, initialPort+i)
		hosts[i] = host
		// creating pubsubs
		comm.initializePubSub(*host)
		/*
			if i == (len(nodes) - 1) {
				comm.Cancel(2050, 2060)

			}
		*/
		// Simulating rejoiningFailure with 1 node getting out of delayed set after some seconds
		nVictim := 0
		switch failureModel {
		case RejoiningMinorityFailure:
			nVictim = (n - 1) / 2
		case RejoiningMajorityFailure:
			nVictim = (n + 1) / 2
		}

		if i < nVictim {
			comm.victim = true
			comm.buffer = make(chan model.Message, BufferLen)
			if i == 0 {
				go func() {
					// Delay for the node to get out of delayed(victim) group
					time.Sleep(30 * time.Second)
					comm.Disconnect()
					comm.victim = false
					comm.Reconnect("")
					fmt.Println("REJOINING FROM DELAYED SET")
				}()
			}
		} else {
			comm.victim = false
			comm.buffer = make(chan model.Message, 0)
		}

		nodes[i] = &model.Node{
			Id:           i,
			TimeStep:     0,
			ThresholdWit: n/2 + 1,
			ThresholdAck: n/2 + 1,
			Acks:         0,
			Comm:         comm,
			History:      make([]model.Message, 0)}
	}
	return nodes, hosts
}

// setupNetworkTopology sets up a simple network topology for test.
func setupNetworkTopology(hosts []*core.Host) {

	// Connect hosts to each other in a topology
	// host0 ---- host1 ---- host2 ----- host3 ----- host4
	//	 			|		   				|    	   |
	//	            ------------------------------------
	n := len(hosts)
	/*
		for i := 0; i< n; i++ {
			for j,nxtHost := range hosts {
				if j == i{
					continue
				}
				connectHostToPeer(*hosts[i], getLocalhostAddress(*nxtHost))
			}
		}
	*/
	for i := 0; i < n; i++ {
		connectHostToPeer(*hosts[i], getLocalhostAddress(*hosts[(i+1)%n]))
		connectHostToPeer(*hosts[i], getLocalhostAddress(*hosts[(i+2)%n]))
		connectHostToPeer(*hosts[i], getLocalhostAddress(*hosts[(i+3)%n]))
		connectHostToPeer(*hosts[i], getLocalhostAddress(*hosts[(i+4)%n]))
	}
	// Wait so that subscriptions on topic will be done and all peers will "know" of all other peers
	time.Sleep(time.Second * 2)

}

func simulateFailure(nodes []*model.Node, n int) {
	for i, node := range nodes {
		if i >= n/2 {
			node.Comm.Disconnect()
			if i == n-3 {
				go func(node *model.Node) {
					time.Sleep(5 * time.Second)
					fmt.Println(node.Id)
					node.Comm.Reconnect("")
					node.Advance(node.TimeStep)
				}(node)
			}
		}
	}
}

func minorityFailure(nodes []*model.Node, n int) int {
	nFail := (n - 1) / 2
	//nFail := 4
	failures(nodes, nFail)
	return nFail
}

func majorityFailure(nodes []*model.Node, n int) int {
	nFail := n/2 + 1
	failures(nodes, nFail)
	return nFail
}

func failures(nodes []*model.Node, nFail int) {
	for i, node := range nodes {
		if i < nFail {
			node.Comm.Disconnect()
		}
	}
}

func simpleTest(t *testing.T, n int, initialPort int, stop int, failureModel FailureModel) {
	var nFail int
	nodes, hosts := setupHosts(n, initialPort, failureModel)

	defer func() {
		fmt.Println("Closing hosts")
		for _, h := range hosts {
			_ = (*h).Close()
		}
	}()

	setupNetworkTopology(hosts)

	// Put failures here
	switch failureModel {
	case MinorFailure:
		nFail = minorityFailure(nodes, n)
	case MajorFailure:
		nFail = majorityFailure(nodes, n)
	case RejoiningMinorityFailure:
		nFail = (n-1)/2 - 1
	case RejoiningMajorityFailure:
		nFail = (n+1)/2 - 1
	}

	// PubSub is ready and we can start our algorithm
	test_utils.StartTest(nodes, stop, nFail)
	test_utils.LogOutput(t, nodes)
}

func TestWithNoFailure(t *testing.T) {
	// Create hosts in libp2p
	logFile, _ := os.OpenFile("../../logs/NoFailure_NoDelay.log", os.O_RDWR|os.O_CREATE, 0666)
	model.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	simpleTest(t, 10, 9900, 10, NoFailure)
}

func TestWithMinorFailure(t *testing.T) {
	// Create hosts in libp2p
	logFile, _ := os.OpenFile("../../logs/MinorFailure.log", os.O_RDWR|os.O_CREATE, 0666)
	model.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	simpleTest(t, 10, 9900, 10, MinorFailure)
}

func TestWithMajorFailure(t *testing.T) {
	// Create hosts in libp2p
	logFile, _ := os.OpenFile("../../logs/MajorFailure.log", os.O_RDWR|os.O_CREATE, 0666)
	model.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	simpleTest(t, 10, 9900, 10, MajorFailure)
}

func TestWithRejoiningMinorityFailure(t *testing.T) {
	// Create hosts in libp2p
	logFile, _ := os.OpenFile("log.log", os.O_RDWR|os.O_CREATE, 0666)
	model.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	simpleTest(t, 10, 9900, 10, RejoiningMinorityFailure)
}

func TestWithRejoiningMajorityFailure(t *testing.T) {
	// Create hosts in libp2p
	logFile, _ := os.OpenFile("log2.log", os.O_RDWR|os.O_CREATE, 0666)
	model.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	simpleTest(t, 10, 9900, 10, RejoiningMajorityFailure)
}
