package channel

import (
	"github.com/dedis/student_19_libp2p_tlc/model"
)

type Channel struct {
	outgoingChannels *map[int]*chan model.Message
	incomingChannel  *chan model.Message
}

// Send is used for sending a message i
func (c *Channel) Broadcast(msg model.Message) {
	go func() {
		for _, peerChannel := range *c.outgoingChannels {
			// TODO handle when channel is full
			*peerChannel <- msg
			//select {
			//case *peerChannel <- msg:
			//default:
			//	fmt.Printf("channel full, msg : %v\n", msg.Source)
			//}
		}
	}()
}

func (c *Channel) Send(msg model.Message, id int) {
	go func() {
		*((*c.outgoingChannels)[id]) <- msg
	}()
}

func (c *Channel) Receive() *model.Message {
	msg := <-*(c.incomingChannel)
	return &msg
	//return <-*c.incomingChannel
}
