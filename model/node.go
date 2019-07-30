package model

type Node struct {
	Id        int                    // Id of the node
	TimeStep  int                    // Node's local time step
	Threshold int                    // Threshold on number of messages
	Acks      int                    // Number of acknowledges
	Comm      CommunicationInterface // interface for communicating with other nodes
	History   []*Message             // History of received messages by a node

}

type CommunicationInterface interface {
	Send(msg *Message, n *Node) // Send a message to a specific node
	Broadcast(msg *Message)     // Broadcast messages to other nodes
	Receive() *Message          // Blocking(?) receive
}
