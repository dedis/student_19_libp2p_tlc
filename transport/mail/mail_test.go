package mail

import (
	"fmt"
	"github.com/dedis/student_19_libp2p_tlc/model"
	"github.com/dedis/student_19_libp2p_tlc/transport/libp2p_pubsub"
	"github.com/dedis/student_19_libp2p_tlc/transport/test_utils"
	"github.com/golang/protobuf/proto"
	"testing"
	"time"
)

func TestSendMail(t *testing.T) {
	//usernames := []string{"a@localhost.localdomain", "b@localhost.localdomain", "c@localhost.localdomain", "d@localhost.localdomain", "e@localhost.localdomain"}
	usernames := []string{"a@localhost.localdomain", "e@localhost.localdomain"}
	SendMail(usernames[0], []string{usernames[1]}, "444", []byte("Testing from mail package"), "apassword")
}

func TestGetMail(t *testing.T) {
	GetMail("e@localhost.localdomain", ""+"epassword", 5)
}

// setupHosts is responsible for creating tlc nodes
func setup(n int) []*model.Node {
	// nodes used in tlc model
	nodes := make([]*model.Node, n)
	// Accounts used for emailing
	usernames := []string{"a@localhost.localdomain", "b@localhost.localdomain", "c@localhost.localdomain", "d@localhost.localdomain", "e@localhost.localdomain"}
	passwords := []string{"apassword", "bpassword", "cpassword", "dpassword", "epassword"}

	for i := range nodes {

		//var comm model.CommunicationInterface
		var comm *mail
		comm = new(mail)

		comm.username = usernames[i]
		comm.password = passwords[i]
		comm.recentIndex = 1
		comm.addressBook = usernames

		nodes[i] = &model.Node{
			Id:           i,
			TimeStep:     0,
			ThresholdWit: n/2 + 1,
			ThresholdAck: n/2 + 1,
			Acks:         0,
			Comm:         comm,
			History:      make([]model.Message, 0)}
	}
	return nodes
}

func TestMail(t *testing.T) {
	nodes := setup(3)
	test_utils.StartTest(nodes, 1)
	test_utils.LogOutput(t, nodes)
}

func TestMailProto(t *testing.T) {
	msg := model.Message{Source: 2, Step: 4, MsgType: model.Raw, History: []model.Message{}}
	msgBytes, err := proto.Marshal(libp2p_pubsub.ConvertModelMessage(msg))
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return
	}
	usernames := []string{"a@localhost.localdomain", "b@localhost.localdomain", "c@localhost.localdomain", "d@localhost.localdomain", "e@localhost.localdomain"}
	passwords := []string{"apassword", "bpassword", "cpassword", "dpassword", "epassword"}
	fmt.Println("BYTES", msgBytes)
	SendMail(usernames[0], usernames, "", msgBytes, passwords[0])
	time.Sleep(time.Second)
	data := GetMail(usernames[2], passwords[2], 1)
	fmt.Println("BYTES", data)
}
