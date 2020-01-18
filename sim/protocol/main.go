package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"go.dedis.ch/onet/v3"
	"go.dedis.ch/protobuf"
	"io/ioutil"
	"log"
	mrand "math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dedis/student_19_libp2p_tlc/protobuf/messagepb"
	Libp2p "github.com/dedis/student_19_libp2p_tlc/transport/libp2p_pubsub"
	"github.com/dedis/student_19_libp2p_tlc/transport/test_utils"

	"github.com/dedis/student_19_libp2p_tlc/model"
	core "github.com/libp2p/go-libp2p-core"
)

const DefaultPort = 2020
const TLCNodesPerServer = 10
const PerServerConnectionsNum = 9
const InterServerConnectionsNum = 18
const ServersNum = 19

// setupServerHosts is responsible for creating tlc nodes and also libp2p hosts on each server.
func setupServerHosts(n int, serverId int, ip string) ([]*model.Node, []*core.Host) {
	// nodes used in tlc model
	nodes := make([]*model.Node, TLCNodesPerServer)

	// hosts used in libp2p communications
	hosts := make([]*core.Host, TLCNodesPerServer)

	for i := range nodes {
		id := serverId*TLCNodesPerServer + i

		var comm *Libp2p.Libp2pPubSub
		comm = new(Libp2p.Libp2pPubSub)

		// creating libp2p hosts
		host := comm.CreatePeerWithIp(id, ip, DefaultPort+i)
		hosts[i] = host

		// creating pubsubs
		comm.InitializePubSub(*host)

		comm.InitializeVictim(false)

		nodes[i] = &model.Node{
			Id:           id,
			TimeStep:     0,
			ThresholdWit: n/2 + 1,
			ThresholdAck: n/2 + 1,
			Acks:         0,
			ConvertMsg:   &messagepb.Convert{},
			Comm:         comm,
			History:      make([]model.Message, 0)}
	}
	return nodes, hosts

}

// setupNetworkTopology sets up a simple network topology for test.
func setupNetworkTopology(serverId int, hosts []*core.Host, r *onet.Roster) {

	for i := 0; i < TLCNodesPerServer; i++ {
		// First each node gets connected to some other nodes in the same server
		for j := 1; j <= PerServerConnectionsNum; j++ {
			next := (i + j) % TLCNodesPerServer
			connectHostToPeer(*hosts[i], Libp2p.GetLocalhostAddress(*hosts[next]))
		}
		// Then it makes a connection to nodes in other servers
		for j := 1; j <= InterServerConnectionsNum; j++ {
			nextServer := (serverId + j) % ServersNum
			//fmt.Println("nextServer ", nextServer)
			//fmt.Println(string(r.List[nextServer].Address), nextServer)
			// Getting address of next server from roster
			address := strings.Split(string(r.List[nextServer].Address)[6:], ":")

			nextNodeId := nextServer*TLCNodesPerServer + i
			r := mrand.New(mrand.NewSource(int64(nextNodeId)))
			prvKey, _ := ecdsa.GenerateKey(btcec.S256(), r)
			sk := (*crypto.Secp256k1PrivateKey)(prvKey)

			id, _ := peer.IDFromPrivateKey(sk)

			addr := fmt.Sprintf("/ip4/%s/tcp/%s/p2p/%s", address[0], strconv.Itoa(DefaultPort+i), id.Pretty())
			connectHostToPeer(*hosts[i], addr)
		}
	}
	time.Sleep(time.Second * 2)

}

// connectHostToPeer is used for connecting a host to another peer
func connectHostToPeer(h core.Host, connectToAddress string) {
	// Creating multi address
	multiAddr, err := multiaddr.NewMultiaddr(connectToAddress)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return
	}

	pInfo, err := peer.AddrInfoFromP2pAddr(multiAddr)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return
	}

	err = h.Connect(context.Background(), *pInfo)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
	}
}

func simpleTest(n int, serverId int, ip string, stop int, r *onet.Roster) {
	// setup TLC nodes on current server
	nodes, hosts := setupServerHosts(n, serverId, ip)

	defer func() {
		fmt.Println("Closing hosts")
		for _, h := range hosts {
			_ = (*h).Close()
		}
	}()

	// Define network topology for connecting to other nodes
	// Network topology consists of links between nodes inside the server and some other servers
	setupNetworkTopology(serverId, hosts, r)

	// PubSub is ready and we can start our algorithm
	test_utils.StartTest(nodes, stop, 0)
}

// Testing TLC with majority thresholds with no node failures
func main() {
	buf, err := ioutil.ReadFile("config")
	if err != nil {
		log.Fatal("could not read config", err)
	}

	var r = new(onet.Roster)
	err = protobuf.Decode(buf, r)
	if err != nil {
		log.Fatal("could not read decode", err)
	}

	model.Logger1 = log.New(os.Stdout, "", log.Ltime|log.Lmicroseconds)
	Libp2p.Delayed = false
	id, _ := strconv.Atoi(os.Args[1])
	address := strings.Split(string(r.List[id].Address)[6:], ":")
	fmt.Println(address, len(r.List))
	ip := address[0]
	//port := address[1]
	simpleTest(ServersNum*TLCNodesPerServer, id, ip, 5, r)
}
