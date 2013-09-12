package memberlist

/*
The broadcast mechanism works by maintain a sorted list of messages to be
sent out. When a message is to be broadcast, the retransmit count
is set to zero and appended to the queue. The retransmit count serves
as the "priority", ensuring that newer messages get sent first. Once
a message hits the retransmit limit, it is removed from the queue.

Additionally, older entries can be invalidated by new messages that
are contradictory. For example, if we send "{suspect M1 inc: 1},
then a following {alive M1 inc: 2} will invalidate that message
*/

import (
	"bytes"
	"sort"
)

type broadcast struct {
	transmits int           // Number of times we've transmitted
	node      string        // Which node this is about
	msg       *bytes.Buffer // Message
}

type broadcasts []*broadcast

// queueBroadcast is used to start dissemination of a message. It will be
// sent up to a configured number of times. The message could potentially
// be invalidated by a future message about the same node
func (m *Memberlist) queueBroadcast(node string, msg *bytes.Buffer) {
	m.broadcastLock.Lock()
	defer m.broadcastLock.Unlock()

	// Check if this message invalidates another
	n := len(m.bcQueue)
	for i := 0; i < n; i++ {
		if m.bcQueue[i].node == node {
			copy(m.bcQueue[i:], m.bcQueue[i+1:])
			m.bcQueue[n-1] = nil
			m.bcQueue = m.bcQueue[:n-1]
			break
		}
	}

	// Append to the queue
	m.bcQueue = append(m.bcQueue, &broadcast{transmits: 0, node: node, msg: msg})
}

func (b broadcasts) Len() int {
	return len(b)
}

func (b broadcasts) Less(i, j int) bool {
	return b[i].transmits < b[j].transmits
}

func (b broadcasts) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

// Sort will order the broadcasts from most transmits to least
func (b broadcasts) Sort() {
	sort.Sort(sort.Reverse(b))
}