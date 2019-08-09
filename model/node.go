package model

type Node struct {
	Id           int                    // Id of the node
	TimeStep     int                    // Node's local time step
	ThresholdAck int                    // Threshold on number of messages
	ThresholdWit int                    // Threshold on number of witnessed messages
	Acks         int                    // Number of acknowledges
	Wits         int                    // Number of witnesses
	Comm         CommunicationInterface // interface for communicating with other nodes
	CurrentMsg   Message
	History      []Message // History of received messages by a node

}

type CommunicationInterface interface {
	Send(msg Message, id int) // Send a message to a specific node
	Broadcast(msg Message)    // Broadcast messages to other nodes
	Receive() *Message        // Blocking(?) receive
}
