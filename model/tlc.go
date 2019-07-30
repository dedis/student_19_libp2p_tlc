package model

import "fmt"

// Advance will change the step of node to a new one.
func (node *Node) Advance(step int) {
	node.TimeStep = step
	node.Acks = 0

	fmt.Printf("node %d , Broadcast in timeStep %d,%v\n", node.Id, node.TimeStep, node.History)
	node.Comm.Broadcast(node.makeMsg(node.Id, step, node.History))

}

// makeMsg makes a new message to be sent to other nodes.
func (node *Node) makeMsg(source int, step int, history []*Message) *Message {
	var msg Message
	msg.Source = source
	msg.MsgType = Raw
	msg.Step = step
	msg.History = history
	return &msg
}

// waitForMsg endlessly(?) (till a stop point actually) waits for upcoming messages and then decides the next action with respect to msg's contents.
func (node *Node) WaitForMsg(stop int) {
	for node.TimeStep <= stop {
		// For now we assume that the underlying receive function is blocking
		// TODO Implement receive in a blocking way or introduce 2 kinds of receives
		msg := node.Comm.Receive()
		fmt.Printf("node %d,%d\n", node.Id, msg.Step)

		if node.TimeStep == stop {
			fmt.Println("Break reached")

			break
		}

		if msg.Step > node.TimeStep { // Node needs to catch up with the message
			// Update nodes local history. Append history from message to local history
			node.History = append(node.History, msg.History[node.TimeStep:]...)
			// Advance to the virally
			node.Advance(msg.Step)
			node.Acks += 1
		} else if msg.Step == node.TimeStep { // Count message toward the threshold
			node.Acks += 1
			if node.Acks >= node.Threshold {
				// Log the message in history
				node.History = append(node.History, msg)
				// Advance to next time step
				node.Advance(node.TimeStep + 1)
			}
		}

	}
}
