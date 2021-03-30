package main

import (
	"log"
	"time"

	"github.com/dustin/go-broadcast"
)

// Example of a simple broadcaster sending numbers to two workers.

func main() {
	// create the broadcaster
	broadcaster := broadcast.NewBroadcaster(100)

	// start workers
	go worker(1, broadcaster)
	go worker(2, broadcaster)

	// sending 5 messages to the broadcaster
	for i := 0; i < 5; i++ {
		log.Printf("Sending %+v to broadcaster", i)
		broadcaster.Submit(i)
		time.Sleep(time.Second)
	}

	// closing the broadcaster
	err := broadcaster.Close()
	if err != nil {
		log.Printf("Closing broadcaster failed: %s", err.Error())
	}
}

func worker(id int, broadcaster broadcast.Broadcaster) {
	broadcastChan := make(chan interface{})
	broadcaster.Register(broadcastChan)

	defer broadcaster.Unregister(broadcastChan)
	defer log.Printf("worker %d is done \n", id)

	// dump out each message sent to the broadcaster.
	for {
		select {
		case value := <-broadcastChan:
			log.Printf("worker %d read %+v \n", id, value)
		}
	}
}
