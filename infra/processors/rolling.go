package processors

import (
	"bnb/infra"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"
)

type node struct {
	infra.PayloadTrade

	nextNode *node
}

type RollingList struct {
	head *node

	streamData                infra.StreamData
	stop                      chan struct{}
	spoolTo                   []io.Writer
	locationOffsetMiliseconds int64
	retentionSeconds          int
}

var _ infra.IProcessor = &RollingList{}

func NewRolling(symbol string, retentionSeconds int, spoolTo ...io.Writer) *RollingList {
	return &RollingList{
		head: &node{},

		streamData: infra.StreamData{
			Stream: symbol,
			Feed:   make(chan infra.PayloadTrade),
		},

		stop:             make(chan struct{}),
		retentionSeconds: retentionSeconds,
		spoolTo:          spoolTo,
	}
}

// Listen Method listens for events coming with different offsets.
// For GMT: 3 * 3600 * 1000 = 10800000
func (l *RollingList) Listen(locationOffsetMiliseconds int64) {
	l.locationOffsetMiliseconds = locationOffsetMiliseconds

loop:
	for {
		select {
		case <-l.stop:
			{
				log.Println("stopping processor time list")
				break loop
			}

		case payload := <-l.streamData.Feed:
			{
				l.prepend(
					&node{
						PayloadTrade: payload,
					},
				)
			}
		}
	}
}

func (l *RollingList) Payload() infra.StreamData {
	return l.streamData
}

func (l *RollingList) SendBufferTo(w io.Writer) {
	currentNode := l.head

	var length int

	w.Write([]byte("Printing Buffer\n"))

	for currentNode.nextNode != nil {
		w.Write([]byte(fmt.Sprintf("%v", currentNode.PayloadTrade)))

		length++
		currentNode = currentNode.nextNode
	}

	w.Write([]byte("\n"))
	w.Write([]byte("Length: " + strconv.Itoa(length)))
	w.Write([]byte("\n"))
}

func (l *RollingList) Terminate() {
	defer l.cleanUp()

	l.stop <- struct{}{}
}

func (l *RollingList) prepend(n *node) {
	n.nextNode = l.head
	l.head = n

	dropAfterTimeMiliseconds := time.Now().Unix()*1000 - l.locationOffsetMiliseconds - int64(l.retentionSeconds*1000)

	l.walkList(dropAfterTimeMiliseconds) // synchronous to preserve data
}

// walkList Internal method drops data past specified point in time.
func (l *RollingList) walkList(dropPast int64) {
	if len(l.spoolTo) == 0 {
		return
	}

	currentNode := l.head

	var length float64
	var sum float64

	for currentNode.nextNode != nil {
		if dropPast > currentNode.UNIXTimeMiliseconds {
			currentNode.nextNode = nil

			break
		}

		sum = sum + currentNode.PayloadTrade.Quantity
		length++

		currentNode = currentNode.nextNode
	}

	for _, writer := range l.spoolTo {
		go writer.Write([]byte(
			fmt.Sprintf(
				"%s: %.3f --- %.f \n",
				l.streamData.Stream,
				sum/length,
				length),
		),
		)
	}
}

func (l *RollingList) cleanUp() {
	close(l.stop)
	close(l.streamData.Feed)
}
