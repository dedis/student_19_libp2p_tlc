package transport

import "../model"

type Channel struct {
	outgoingChannels map[int]chan *model.Message
	incomingChannel  chan *model.Message
}

// Send is used for sending a message i
func (c *Channel) Broadcast(msg *model.Message) {
	go func() {
		for _, peerChannel := range c.outgoingChannels {
			peerChannel <- msg
		}
	}()
}

func (c *Channel) Send(msg *model.Message, n *model.Node) {
	go func() {
		c.outgoingChannels[n.Id] <- msg
	}()
}

func (c *Channel) Receive() *model.Message {
	return <-c.incomingChannel
}
