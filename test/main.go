package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"strconv"
	"time"

	"github.com/goombaio/namegenerator"
	"github.com/zeromq/goczmq"
	log "gopkg.in/inconshreveable/log15.v2"
)

const PING_TIMEOUT = 10 * time.Second

var (
	poolSocks []*goczmq.Sock
	sinkSock  *goczmq.Sock
)

func main() {
	loadWorkerConf()
	defer destroySocks()

	tasks := make(chan *TaskWorker)
	pongChan := make(chan string)
	fileChan := make(chan string)

	go initSink(pongChan, fileChan)

	go func() {
		for task := range tasks {
			loadBalancer(task, pongChan)
		}
	}()

	//processing data comming from channel
	go func() {
		for data := range fileChan {
			d := &DataTransfer{}
			_ = json.Unmarshal([]byte(data), d)
			fmt.Printf("PATH: %s\n", d.Path)
			os.Remove(d.Path)
		}
	}()

	ticker := time.NewTicker(2 * time.Second)

	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)

	for {
		select {
		case <-ticker.C:
			log.Info("Sending New task...", "info", "tasker")
			name := nameGenerator.Generate()
			task := newTaskWorker("http://localhost:4000/static/videos/1.mp4", name, 123)
			tasks <- task
		case <-time.After(5 * time.Second):
			break
		}
	}
}

func loadWorkerConf() {
	//load worker.conf
	log.Info("Setting up front workers from worker.conf", "infor", "loadWorkerConf")
	file, err := os.Open("worker.conf")
	if err != nil {
		log.Crit("Unable to open worker.conf", "err", err)
	}

	//process file lines
	input := bufio.NewScanner(file)
	for input.Scan() {
		line := input.Text()
		//make some validation in case someone
		//mess the config file

		//append validated line
		pushSock, err := goczmq.NewPush(line)
		if err != nil {
			log.Error(fmt.Sprintf("Unable to create Dealer socket in: %s", line), "err", err)
		} else {
			log.Info(fmt.Sprintf("Adding socket to pool, address: %s", line), "infor", "loadWorkerConf")
			poolSocks = append(poolSocks, pushSock)
		}
	}
}

func destroySocks() {
	for i, sock := range poolSocks {
		log.Info(fmt.Sprintf("Destroying socket: %d", i), "info", "destroySocks")
		sock.Destroy()
	}
}

//getHash good all stackoverflow
func getHash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

type TaskWorker struct {
	URL      string
	WorkerID int
	ChatID   int64
	Title    string
}

//DataTransfer is the data passed between
//ZeroMQ sockets
type DataTransfer struct {
	URL    string `json:"url"`
	ChatID string `json:"chatid"`
	Path   string `json:"path"`
}

func newDataTransfer(task *TaskWorker) *DataTransfer {
	chatID := strconv.FormatInt(task.ChatID, 10)
	return &DataTransfer{
		URL:    task.URL,
		ChatID: chatID,
		Path:   fmt.Sprintf("%s/%s.mp4", chatID, task.Title),
	}
}

func newTaskWorker(url, title string, chatID int64) *TaskWorker {
	index := int(getHash(url)) % len(poolSocks)
	return &TaskWorker{
		URL:      url,
		WorkerID: index,
		ChatID:   chatID,
		Title:    title,
	}
}

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
			fmt.Printf("Passing %s", msg)
			fileChan <- string(msg)
		}
	}
}

//pingSock ping to the other part until we get a response
//with a fixed timeout PING_TIMEOUT
func pingSock(sock *goczmq.Sock, pongChan chan string) bool {
	ticker := time.NewTicker(2 * time.Second)
	msg := ""

	for {
		select {
		case <-ticker.C:
			log.Info("Sending Ping...", "info", "pingSock")
			err := sock.SendFrame([]byte("Ping"), goczmq.FlagNone)
			if err != nil {
				log.Error("Unable to send Ping", "err", err)
			}
		case msg = <-pongChan:
			if msg == "Pong" {
				return true
			}
		case <-time.After(PING_TIMEOUT):
			break
		}
	}

	return false
}

//loadBalancer not so balanced, just random dispatcher
func loadBalancer(task *TaskWorker, pongChan chan string) {
	sock := poolSocks[task.WorkerID]
	if pingSock(sock, pongChan) {
		data, err := json.Marshal(newDataTransfer(task))
		if err != nil {
			log.Error(fmt.Sprintf("Unable to Marshall: %+v", task), "err", err)
			return
		}
		err = sock.SendFrame(data, goczmq.FlagNone)
		if err != nil {
			log.Error(fmt.Sprintf("Unable to send message with URL: %s", task.URL), "err", err)
		}
		return
	}
	log.Warn(fmt.Sprintf("No Ping response from socket: %s", task.URL), "info", "loadBalancer")
}
