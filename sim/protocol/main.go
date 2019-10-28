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
	"sync"
	"time"

	"github.com/dedis/student_19_libp2p_tlc/protobuf/messagepb"
	Libp2p "github.com/dedis/student_19_libp2p_tlc/transport/libp2p_pubsub"

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
		Id:           id,
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
func setupNetworkTopology(n int, id int, host *core.Host, r *onet.Roster) {
	var addr string

	for i := 1; i <= 4; i++ {
		next := (id + i) % n
		fmt.Println("next ", next)
		fmt.Println(string(r.List[next].Address), next)
		address := strings.Split(string(r.List[next].Address)[6:], ":")

		r := mrand.New(mrand.NewSource(int64(next)))
		prvKey, _ := ecdsa.GenerateKey(btcec.S256(), r)
		sk := (*crypto.Secp256k1PrivateKey)(prvKey)

		id, _ := peer.IDFromPrivateKey(sk)
		addr = fmt.Sprintf("/ip4/%s/tcp/%s/p2p/%s", address[0], "2020", id.Pretty())
		connectHostToPeer(*host, addr)
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

func simpleTest(n int, id int, ip string, port string, stop int, failureModel FailureModel, r *onet.Roster) {
	node, host := setupHost(n, id, ip, port, failureModel)

	defer func() {
		_ = (*host).Close()
	}()

	setupNetworkTopology(n, id, host, r)

	// PubSub is ready and we can start our algorithm
	StartTest(node, stop)
	//test_utils.LogOutput(t, nodes)
}

// StartTest is used for starting tlc nodes
func StartTest(node *model.Node, stop int) {
	wg := &sync.WaitGroup{}
	node.Advance(0)
	wg.Add(1)
	go func(node *model.Node, stop int, wg *sync.WaitGroup) {
		defer wg.Done()
		node.WaitForMsg(stop)
	}(node, stop, wg)

	wg.Wait()
	fmt.Println("The END")
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

	// Create hosts in libp2p
	model.Logger1 = log.New(os.Stdout, "", log.Ltime|log.Lmicroseconds)
	Libp2p.Delayed = false
	id, _ := strconv.Atoi(os.Args[1])
	address := strings.Split(string(r.List[id].Address)[6:], ":")
	fmt.Println(address, len(r.List))
	ip := address[0]
	//port := address[1]
	port := "2020"
	//ip := "127.0.0.1"
	//port := strconv.Itoa(9000 + id)
	simpleTest(11, id, ip, port, 5, NoFailure, r)
}
