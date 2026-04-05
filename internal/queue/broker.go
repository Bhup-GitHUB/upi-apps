package queue

import "fmt"

type Broker struct {
	ch chan Event
}

func NewBroker() *Broker {
	return &Broker{
		ch: make(chan Event, 100),
	}
}

func (b *Broker) Publish(e Event) {
	fmt.Println("queue: publishing", string(e.Type), "txn", e.TxnID, "trace", e.TraceID)
	b.ch <- e
}

func (b *Broker) Subscribe() <-chan Event {
	return b.ch
}
