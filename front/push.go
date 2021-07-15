package front

import (
	"bufio"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"strconv"
	"time"

	"github.com/zeromq/goczmq"
	log "gopkg.in/inconshreveable/log15.v2"
)

const PING_TIMEOUT = 10 * time.Second

var (
	poolSocks []*goczmq.Sock
	sinkSock  *goczmq.Sock
)

//getHash good all stackoverflow
func getHash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

type TaskWorker struct {
	URL string
    ImageURL string
	ChatID  int64
	Title   string
	HashKey string
}

//DataTransfer is the data passed between
//ZeroMQ sockets
type DataTransfer struct {
	URL     string `json:"url"`
    ImageURL string `json:imageurl"`
	ChatID  string `json:"chatid"`
	Path    string `json:"path"`
	HashKey string `json:"hashkey"`
}

func newDataTransfer(task *TaskWorker) *DataTransfer {
	chatID := strconv.FormatInt(task.ChatID, 10)
	return &DataTransfer{
		URL:     task.URL,
        ImageURL: task.ImageURL,
		ChatID:  chatID,
		Path:    fmt.Sprintf("%s/%s.mp4", chatID, task.Title),
		HashKey: task.HashKey,
	}
}

func newTaskWorker(url, imageURL, title string, chatID int64) *TaskWorker {
	hashKey := fmt.Sprintf("%d", getHash(url))
	return &TaskWorker{
		URL: url,
        ImageURL: imageURL,
		ChatID:  chatID,
		Title:   title,
		HashKey: hashKey,
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
			log.Error(fmt.Sprintf("Unable to create PUSH socket in: %s", line), "err", err)
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

//pushRedisQueue push task information into redis queue
func pushRedisQueue(rdsRepo *redisRepo, task *TaskWorker) {
	data, err := json.Marshal(newDataTransfer(task))
	if err != nil {
		log.Error(fmt.Sprintf("Unable to Marshall: %+v", task), "err", err)
		return
	}
	_, err = rdsRepo.LPush(string(data))
	if err != nil {
		log.Error("Unable to push task into tasks queue")
		return
	}
	log.Info(fmt.Sprintf("Pushed task into Redis %s QUEUE", TASK_QUEUE), "info", "pushRedisQueue")
    chatID := strconv.FormatInt(task.ChatID, 10)
    _, err = rdsRepo.Incr(chatID)
	if err != nil {
		log.Error("Unable to increase counter task")
	}
}
