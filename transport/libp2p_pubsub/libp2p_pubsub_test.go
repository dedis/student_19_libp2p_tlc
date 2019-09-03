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
	LeaveRejoin
	ThreeGroups
)

const FailureDelay = 3
const RejoinDelay = 15
const LeaveDelay = 10

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
		host := comm.CreatePeer(i, initialPort+i)
		hosts[i] = host

		// creating pubsubs
		comm.InitializePubSub(*host)

		// Simulating rejoining failures, where a node leaves the delayed set and joins other progressing nodes
		nVictim := 0
		switch failureModel {
		case RejoiningMinorityFailure:
			nVictim = (n - 1) / 2
		case RejoiningMajorityFailure:
			nVictim = (n + 1) / 2
		case LeaveRejoin:
			nVictim = (n - 1) / 2
		case ThreeGroups:
			nVictim = n
		}
		if failureModel == ThreeGroups {
			if i < 3 {
				comm.JoinGroup([]int{0, 1, 2})
			} else if i < 6 {
				comm.JoinGroup([]int{3, 4, 5})
			} else {
				comm.JoinGroup([]int{})
			}
		}

		if i < nVictim {
			comm.InitializeVictim(true)

			go func() {
				time.Sleep(2 * FailureDelay * time.Second)
				comm.AttackVictim()
			}()

			nRejoin := 2
			if failureModel == ThreeGroups {
				nRejoin = 6
			}
			if i < nRejoin {
				go func() {
					// Delay for the node to get out of delayed(victim) group
					time.Sleep((RejoinDelay + time.Duration(FailureDelay*i)) * time.Second)

					comm.ReleaseVictim()
				}()
			}
		} else {
			if failureModel == LeaveRejoin {
				if i == (n - 1) {
					go func() {
						// Delay for the node to leave the progressing group
						time.Sleep(LeaveDelay * time.Second)
						comm.Disconnect()
					}()
				}
			}

			comm.InitializeVictim(false)
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

func minorityFailure(nodes []*model.Node, n int) int {
	nFail := (n - 1) / 2
	//nFail := 4
	go func(nodes []*model.Node, nFail int) {
		time.Sleep(FailureDelay * time.Second)
		failures(nodes, nFail)
	}(nodes, nFail)

	return nFail
}

func majorityFailure(nodes []*model.Node, n int) int {
	nFail := n/2 + 1
	go func(nodes []*model.Node, nFail int) {
		time.Sleep(FailureDelay * time.Second)
		failures(nodes, nFail)
	}(nodes, nFail)
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
	case LeaveRejoin:
		nFail = (n-1)/2 - 1
	case ThreeGroups:
		nFail = (n - 1) / 2
	}

	// PubSub is ready and we can start our algorithm
	test_utils.StartTest(nodes, stop, nFail)
	test_utils.LogOutput(t, nodes)
}

// Testing TLC with majority thresholds with no node failures
func TestWithNoFailure(t *testing.T) {
	// Create hosts in libp2p
	logFile, _ := os.OpenFile("log51.log", os.O_RDWR|os.O_CREATE, 0666)
	model.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	Delayed = false
	simpleTest(t, 11, 9900, 10, NoFailure)
}

// Testing TLC with minor nodes failing
func TestWithMinorFailure(t *testing.T) {
	// Create hosts in libp2p
	logFile, _ := os.OpenFile("log3.log", os.O_RDWR|os.O_CREATE, 0666)
	model.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	simpleTest(t, 11, 9900, 5, MinorFailure)
}

// Testing TLC with majority of nodes failing
func TestWithMajorFailure(t *testing.T) {
	// Create hosts in libp2p
	logFile, _ := os.OpenFile("log4.log", os.O_RDWR|os.O_CREATE, 0666)
	model.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	simpleTest(t, 10, 9900, 5, MajorFailure)
}

// Testing TLC with majority of nodes working correctly and a set of delayed nodes. a node will leave the victim set
// after some seconds and rejoin to the progressing nodes.
func TestWithRejoiningMinorityFailure(t *testing.T) {
	// Create hosts in libp2p
	logFile, _ := os.OpenFile("../../logs/RejoiningMinority.log", os.O_RDWR|os.O_CREATE, 0666)
	model.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	simpleTest(t, 11, 9900, 10, RejoiningMinorityFailure)
}

// Testing TLC with majority of nodes being delayed. a node will leave the victim set after some seconds and rejoin to
// the other connected nodes. This will make progress possible.
func TestWithRejoiningMajorityFailure(t *testing.T) {
	// Create hosts in libp2p
	logFile, _ := os.OpenFile("../../logs/RejoiningMajority.log", os.O_RDWR|os.O_CREATE, 0666)
	model.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	simpleTest(t, 11, 9900, 10, RejoiningMajorityFailure)
}

// Testing TLC with majority of nodes working correctly and a set of delayed nodes. a node will lose connection to
// progressing nodes and will stop the progress. After some seconds another node will join to the set, making progress
// possible.
func TestWithLeaveRejoin(t *testing.T) {
	// Create hosts in libp2p
	logFile, _ := os.OpenFile("log8.log", os.O_RDWR|os.O_CREATE, 0666)
	model.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	simpleTest(t, 11, 9900, 8, LeaveRejoin)
}

func TestWithThreeGroups(t *testing.T) {
	// Create hosts in libp2p
	logFile, _ := os.OpenFile("log9.log", os.O_RDWR|os.O_CREATE, 0666)
	model.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	simpleTest(t, 11, 9900, 8, ThreeGroups)
}
