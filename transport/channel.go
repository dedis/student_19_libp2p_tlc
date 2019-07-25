package transport

import (
	"../model"
	"fmt"
)

type Channel struct {
	outgoingChannels *map[int]*chan *model.Message
	incomingChannel  *chan *model.Message
}

// Send is used for sending a message i
func (c *Channel) Broadcast(msg *model.Message) {
	go func() {
		for _, peerChannel := range *c.outgoingChannels {
			select {
			case *peerChannel <- msg:
			default:
				fmt.Printf("channel full, msg : %v\n", msg.Source)
			}
		}
	}()
}

func (c *Channel) Send(msg *model.Message, n *model.Node) {
	go func() {
		*(*c.outgoingChannels)[n.Id] <- msg
	}()
}

func (c *Channel) Receive() *model.Message {
	msg := <-*(c.incomingChannel)
	return msg
	//return <-*c.incomingChannel
}
