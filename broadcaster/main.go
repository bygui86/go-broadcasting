package main

import (
	"errors"
	"fmt"
	"time"
)

// PLEASE NOTE: it's a pretty complex implementation and maybe a bit too over-engineered

// There is no sync usage, and there's a go-routine Listen() for the Broadcaster that manages the state.
// It listens for new Receivers and it listens for notification events.
// It is single-threaded, and as a consequence, it does not need any locking.

// It's very important to note that all the channels are unbuffered.
// This allows for a clean and reliable communication path, you can't have stale receivers, etc.

func main() {
	// create Broadcaster
	broadcaster := NewBroadcaster()
	defer broadcaster.Close()

	// start Broadcaster
	go broadcaster.Listen()

	// create control channels
	ctrlDone := make(chan bool)
	ctrlQuit := make(chan struct{})

	fmt.Println()
	fmt.Println()
	time.Sleep(3 * time.Second)

	// start Controller
	go startController(2, ctrlDone, ctrlQuit)

	fmt.Println()
	fmt.Println()
	time.Sleep(3 * time.Second)

	// start Listeners
	// go startListener(1, 1, broadcaster, ctrlDone)
	go startListener(2, 4, broadcaster, ctrlDone)

	fmt.Println()
	fmt.Println()
	time.Sleep(3 * time.Second)

	// start Sender
	go startSender(broadcaster, ctrlQuit)

	// Use a dirty hack to avoid everything ending too early
	fmt.Println("waiting 5 sec before ending main go-routine")
	fmt.Println()
	fmt.Println()
	time.Sleep(5 * time.Second)

	fmt.Println()
	fmt.Println()
}

// Listeners

func startListener(id, loops int, broadcaster *Broadcaster, ctrlDone chan bool) {
	fmt.Println(fmt.Sprintf("Listener %d START with %d loops", id, loops))
	defer fmt.Println(fmt.Sprintf("Listener %d STOP", id))

	for i := 1; i <= loops; i++ {
		<-broadcaster.Receive(id)
		fmt.Println(fmt.Sprintf("Listener %d: received from Broadcaster (%d)", id, i))
	}

	time.Sleep(1 * time.Second)
	ctrlDone <- true
}

// Controller

func startController(listeners int, ctrlDone chan bool, ctrlQuit chan struct{}) {
	fmt.Println("Controller START")
	defer fmt.Println("Controller STOP")

	for i := 1; i <= listeners; i++ {
		<-ctrlDone
		fmt.Println("Controller received ctrlDone signal from a listener")
	}

	close(ctrlDone)
	close(ctrlQuit)
}

// Sender

func startSender(broadcaster *Broadcaster, ctrlQuit chan struct{}) {
	fmt.Println("Sender START")
	defer fmt.Println("Sender STOP")

	for {
		select {

		case <-time.After(500 * time.Millisecond):
			fmt.Println("Sender SEND to Broadcaster")
			err := broadcaster.Send()
			if err != nil {
				fmt.Printf(fmt.Sprintf("Sender ERROR sending to Broadcaster: %s", err.Error()))
			}

		case <-ctrlQuit:
			fmt.Println("Sender received ctrlQuit signal")
			return
		}
	}
}

// Broadcaster

// Broadcaster allows to send a signal to all listeners
// Broadcaster implements the Closer interface
type Broadcaster struct {
	poke      chan bool
	receivers chan chan struct{}
	done      chan bool
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		poke:      make(chan bool),
		receivers: make(chan chan struct{}),
		done:      make(chan bool),
	}
}

// TODO better understand
func (b *Broadcaster) Listen() {
	fmt.Println("Broadcaster LISTEN")

	defer close(b.done)

	notify := make(chan struct{})
	for {
		select {

		case poke := <-b.poke:
			if !poke {
				fmt.Println("Broadcaster: missed POKE, stop listening")
				return
			}

			// all current receivers get a closed channel
			fmt.Println("Broadcaster: all current receivers get a closed channel")
			close(notify)

			// set up next batch of receivers
			fmt.Println("Broadcaster: set up next batch of receivers")
			notify = make(chan struct{})

		case b.receivers <- notify:
			fmt.Println("Broadcaster: a Listener has the channel")
		}
	}
}

// Send a signal to all current Receivers
func (b *Broadcaster) Send() error {
	fmt.Println("Broadcaster SEND")

	select {

	case b.poke <- true:
		return nil

	case <-b.done:
		return errors.New("attempt to send on a closed Broadcaster")
	}
}

// Receive a channel on which the next (close) signal will be sent
func (b *Broadcaster) Receive(listenerId int) <-chan struct{} {
	fmt.Println(fmt.Sprintf("Broadcaster RECEIVE on Listener %d", listenerId))

	select {

	case channel := <-b.receivers:
		fmt.Println(fmt.Sprintf("Broadcaster RECEIVE on Listener %d: channel received", listenerId))
		return channel

	case <-b.done:
		// TODO should probably return an error
		return nil
		// return errors.New("attempt to send on a closed Broadcaster")
	}
}

// Close makes Broadcaster implement the Closer interface
func (b *Broadcaster) Close() error {
	fmt.Println("Broadcaster CLOSE")

	select {
	case b.poke <- false:
		// no action
	case <-b.done:
		// no action
	}
	return nil
}
