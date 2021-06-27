package front

import (
	"fmt"

	"github.com/zeromq/goczmq"
	log "gopkg.in/inconshreveable/log15.v2"
)

//initSink is the listener from the worker responses
func initSink(pongChan chan string, fileChan chan string) {
	endpoint := "tcp://*:5557"
	sinkSock, err := goczmq.NewPull(endpoint)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to create PULL socket in: %s", endpoint), "err", err)
	}
	log.Info(fmt.Sprintf("Sink running at: %s...", endpoint), "info", "initSink")

	//processing incomming messages
	for {
		//ignoring the flag for now
		msg, _, err := sinkSock.RecvFrame()
		if err != nil {
			log.Warn("Unable to recv frame on sink", "err", err)
		}
		if string(msg) == "Pong" {
			pongChan <- string(msg)
		} else {
			fileChan <- string(msg)
		}
	}
}
