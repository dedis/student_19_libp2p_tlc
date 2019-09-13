package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/dedis/student_19_libp2p_tlc/protobuf/messagepb"
	Libp2p "github.com/dedis/student_19_libp2p_tlc/transport/libp2p_pubsub"
	"github.com/dedis/student_19_libp2p_tlc/transport/test_utils"

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
func setupHost(n int, id int, ip string, port string, failureModel FailureModel) (*model.Node, *core.Host) {
	//var comm model.CommunicationInterface
	var comm *Libp2p.Libp2pPubSub
	comm = new(Libp2p.Libp2pPubSub)
	// creating libp2p hosts
	host := comm.CreatePeerWithIp(id, ip, port)

	// creating pubsubs
	comm.InitializePubSub(*host)

	comm.InitializeVictim(false)

	node := &model.Node{
		Id:           i,
		TimeStep:     0,
		ThresholdWit: n/2 + 1,
		ThresholdAck: n/2 + 1,
		Acks:         0,
		ConvertMsg:   &messagepb.Convert{},
		Comm:         comm,
		History:      make([]model.Message, 0)}

	return node, host

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
		//connectHostToPeer(*hosts[i], getLocalhostAddress(*hosts[(i+2)%n]))
		//connectHostToPeer(*hosts[i], getLocalhostAddress(*hosts[(i+3)%n]))
		//connectHostToPeer(*hosts[i], getLocalhostAddress(*hosts[(i+4)%n]))
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

func simpleTest(n int, id int, ip string, port string, stop int, failureModel FailureModel) {
	var nFail int
	nodes, hosts := setupHost(n, id, ip, port, failureModel)

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
func main() {
	// Create hosts in libp2p
	logFile, _ := os.OpenFile("log51.log", os.O_RDWR|os.O_CREATE, 0666)
	model.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	Libp2p.Delayed = false
	id, _ := strconv.Atoi(os.Args[1])
	simpleTest(11, id, ip, port, 10, NoFailure)
}
