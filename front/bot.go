package front

import (
    "context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"


    "github.com/cloudinary/cloudinary-go"
    "github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/Gealber/calvopro-botv2/scrapper"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "gopkg.in/inconshreveable/log15.v2"
)

var (
	TOKEN_API = os.Getenv("TOKEN_TELGRAM_API")
    CLOUDINARY_URL = os.Getenv("CLOUDINARY_URL")
)

//Repos repos of DBs
type Repos struct {
	psq *postgreRepo
	rds *redisRepo
}

//InitBot initialize bot
func InitBot() {
	if len(TOKEN_API) == 0 {
		log.Crit("Empty TOKEN_TELGRAM_API", "env")
		os.Exit(0)
	}
	bot, err := tgbotapi.NewBotAPI(TOKEN_API)
	if err != nil {
		log.Crit("Unable to initiate bot", "err", err)
		os.Exit(0)
	}

	//seting up the logger
	logger := newBotLogger()
	tgbotapi.SetLogger(logger)

	bot.Debug = false

	//process incoming messages
	botUpdates(bot)
}

//botUpdates process all the incomming messages
//from the users using the bot
func botUpdates(bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	//connect to PostgreSql
	repoPsq := NewPostgreRepo()
	repoPsq.Connect()

	//connect to Redis
	rdsRepo := NewRedisRepo()
	rdsRepo.Connect()

	repos := &Repos{
		psq: repoPsq,
		rds: rdsRepo,
	}

	//This need to be improved
	tasks := make(chan *TaskWorker)

	//firing sink and loadbalancer
	go func() {
		for task := range tasks {
			pushRedisQueue(repos.rds, task)
		}
	}()

	for update := range updates {
		go handleUpdates(bot, update, repos, tasks)
	}
}

//handleUpdates ...
func handleUpdates(bot *tgbotapi.BotAPI, update tgbotapi.Update, repos *Repos, tasks chan *TaskWorker) {
	if update.Message != nil {
		if update.Message.IsCommand() {
			handleCommand(bot, update, repos.psq)
            return
		}
		user := update.Message.From
        chatID := update.Message.Chat.ID
        //no need for authorization
        //anyone can use it
        Register(user, chatID, repos.psq, repos.rds)
		handleQuery(bot, update, repos.rds)
        return
		//if IsAuthorized(user, repos.psq, repos.rds) {
		//	handleQuery(bot, update, repos.rds)
        //    return
		//}
        //chatID := update.Message.Chat.ID
        //msg := tgbotapi.NewMessage(chatID, NOT_AUTHORIZED_MSG)
        //bot.Send(msg)
        //return
	}

	if update.CallbackQuery != nil {
		handleCallbackQuery(bot, update, repos.rds, tasks)
	}
}

//handleQuery incomming queries
func handleQuery(bot *tgbotapi.BotAPI, update tgbotapi.Update, rdsRepo *redisRepo) {
	chatID := update.Message.Chat.ID

	_, err := bot.Request(tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping))
	if err != nil {
		log.Error("Error sending action to chat")
	}

	var infoVideos []*scrapper.Video
	rawQuery := update.Message.Text
	query := scrapper.PrepareQuery(rawQuery)

	//check if query is in cache
	infoString := rdsRepo.Get(query)
	if len(infoString) > 0 {
		infoVideos = scrapper.Deserialize([]byte(infoString))
	} else {
		infoVideos = scrapper.InfoVideos(query)
	}

	msg := tgbotapi.NewMessage(chatID, "🚚 Delivering porn!!!")
	rows := formListMarkup(infoVideos)
	msg.ReplyMarkup = rows

	message, err := bot.Send(msg)
	if err != nil {
		log.Info("Unable to send message, handleQuery", "err", err)
	}

	vidByte := scrapper.Serialize(infoVideos)
	//storing info in redis, for later use
	rdsRepo.Set(strconv.Itoa(message.MessageID), string(vidByte), SESSION_EXP)
	//for caching query result,
	//in case someone make the same query
	//I'm not going to scrape again
	rdsRepo.Set(query, string(vidByte), SESSION_EXP)
}

//handleCommand incomming commands
func handleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update, repo *postgreRepo) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	switch update.Message.Command() {
	case "help":
		msg.Text = helpText
	case "start":
		msg.Text = startText
	case "donate":
		msg.Text = donateText
	case "status":
		msg.Text = "I'm feeling good! Purum purum purum"
	default:
		msg.Text = "Type /help to see how to use this bot."
	}
	_, err := bot.Send(msg)
	if err != nil {
		log.Info("Unable to send message, handleCommand", "err", err)
	}
}

//handleCommand incomming callbacks, when user press a botton mostly
func handleCallbackQuery(bot *tgbotapi.BotAPI, update tgbotapi.Update, rdsRepo *redisRepo, tasks chan *TaskWorker) {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "You dog...!😏")
	_, err := bot.Request(callback)
	if err != nil {
		log.Info("Unable to send callback", "err", err)
	}

	//download callback
	if isDownloadCallback(update) {
		downloadVideo(bot, update, rdsRepo, tasks)
		return
	}

	//handle particular video
	videoInfo(bot, update, rdsRepo)
}

//isDownloadCallback identify if the callback query is from the
//download button
func isDownloadCallback(update tgbotapi.Update) bool {
	data := strings.Split(update.CallbackQuery.Data, "-")
	if len(data) == 2 {
		return true
	}
	return false
}

//videoInfo send the message with info of the particular video
func videoInfo(bot *tgbotapi.BotAPI, update tgbotapi.Update, rdsRepo *redisRepo) {
	chatID := update.CallbackQuery.Message.Chat.ID
	msgID := update.CallbackQuery.Message.MessageID
	//when the callbackData is a number
	index, err := strconv.Atoi(update.CallbackQuery.Data)
	if err != nil {
		log.Error("Callback data is not an integer", "err", err)
		return
	}
	//send information of video
    requestMD := newRequestMD(chatID, msgID, index)
	videoInfoMsg(bot, requestMD, rdsRepo)
}

//videoInfoMsg form the message to be send when requesting
//the information of a particular video
func videoInfoMsg(bot *tgbotapi.BotAPI, requestMD *RequestMD, rdsRepo *redisRepo) {
	//retrieving info from redis
	key := strconv.Itoa(requestMD.MessageID)
	infoString := rdsRepo.Get(key)
	if len(infoString) == 0 {
		text := "Query has expired, make a new one boy"
		sendMsg(bot, requestMD.ChatID, text)
		return
	}
	infoVideos := scrapper.Deserialize([]byte(infoString))
	if len(infoVideos) == 0 {
		text := "Nothing to show, choose a better query"
		sendMsg(bot, requestMD.ChatID, text)
		return
	}
	if requestMD.Index > len(infoVideos) {
		txtMsg := "Query has expired, make a new one boy"
		sendMsg(bot, requestMD.ChatID, txtMsg)
		return
	}
	video := infoVideos[requestMD.Index-1]

	//Sending image with caption and btn
	if len(video.ImageURL) == 0 {
		log.Crit("Empty URL supplied", "err", "videoInfoMsg")
		return
	}
    sendVideoInfo(bot, requestMD, video)
}

//send the message form in videoInfoMsg
func sendVideoInfo( bot *tgbotapi.BotAPI, requestMD *RequestMD, video *scrapper.Video) {
    key := strconv.Itoa(requestMD.MessageID)
	file := tgbotapi.FileURL(video.ImageURL)
	photo := tgbotapi.NewPhoto(requestMD.ChatID, file)

	text := fmt.Sprintf(
		"🎥 Title: %s\n\n🕐 Duration: %s\n\n 💾Size: %dMB",
		video.Title,
		video.Duration,
		video.Size>>20,
	)
	photo.Caption = text
	photo.ReplyMarkup = createBtn("Download", key+"-"+strconv.Itoa(requestMD.Index))
	_, err := bot.Send(photo)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to send message, ImageURL: %s.", video.ImageURL), "err", err)
		log.Info("Trying without photo", "info", "videoInfoMsg")
		msg := tgbotapi.NewMessage(requestMD.ChatID, text)
		msg.ReplyMarkup = createBtn("Download", key+"-"+strconv.Itoa(requestMD.Index))
		_, err := bot.Send(msg)
		if err != nil {
			log.Info("Unable to send message, videoInfoMsg", "err", err)
		}
	}
}

//infoFromUpdate extract info from update
func infoFromUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update, rdsRepo *redisRepo) *TaskWorker {
	chatID := update.CallbackQuery.Message.Chat.ID
    //checking if user can put download another video
    chatIDStr := strconv.FormatInt(chatID, 10)
    countStr := rdsRepo.Get(chatIDStr)
    count, err := strconv.ParseInt(countStr, 10, 64)
    if err != nil {
		log.Crit("counter is not int", "err", "infoFromUpdate")
        count = 0
    }
    if count >= MAX_DOWNLOAD {
		text := "Wait until other downloads finish, we are poor..."
		sendMsg(bot, chatID, text)
		return nil
    }
    if count == 0 {
        //set expiration for the key
        rdsRepo.Set(chatIDStr, "0", SESSION_EXP)
    }

	data := strings.Split(update.CallbackQuery.Data, "-")
	if len(data) != 2 {
		log.Crit("Data in download callback query, doesn't have the correct format", "err", "dowbloadVideo")
		return nil
	}
	originalMsgID, videoIndex := data[0], data[1]
	index, err := strconv.Atoi(videoIndex)
	if err != nil {
		log.Crit("video index in callback data from download is not int", "err", "dowbloadVideo")
		return nil
	}
	//retrieve video information from redis
	infoString := rdsRepo.Get(originalMsgID)
	infoVideos := scrapper.Deserialize([]byte(infoString))
	if len(infoVideos) == 0 {
		text := "Query has expired, make a new one boy"
		sendMsg(bot, chatID, text)
		return nil
	}
	video := infoVideos[index-1]
	url := video.URL
	if len(url) == 0 {
		log.Warn("Empty URL from video info, this shouldn't happened", "err", "dowbloadVideo")
		return nil
	}
	imageURL := uploadToCloud(video.ImageURL)
    return newTaskWorker(url, imageURL, genName(), chatID)
}

//uploadToCloud upload the image to cloudinary for transformation
func uploadToCloud(url string) string {
    var imageURL string
    ctx := context.Background()
    cld, _ := cloudinary.NewFromURL(CLOUDINARY_URL)
    resp, err := cld.Upload.Upload(ctx, url, uploader.UploadParams{
        Eager: "w_300,h_300,c_scale",
    })
    if err != nil {
        log.Warn("Unable to upload image to Cloudinary", "err", err)
        return imageURL
    }

    if len(resp.Eager) > 0 {
        eager := resp.Eager
        imageURL = eager[0].SecureURL
    }
    return imageURL
}

//downloadVideo when the callback is the Download button
func downloadVideo(bot *tgbotapi.BotAPI, update tgbotapi.Update, rdsRepo *redisRepo, tasks chan *TaskWorker) {
	//dispatching task
	task := infoFromUpdate(bot, update, rdsRepo)
    if task == nil {
        return
    }
	hashKey := task.HashKey
	fileID := rdsRepo.Get(hashKey)
	if len(fileID) > 0 {
		file := tgbotapi.FileID(fileID)
		msg := tgbotapi.NewVideo(task.ChatID, file)
        _, err := bot.Send(msg)
		if err != nil {
			log.Info("Unable to send message, downloadVideo", "err", err)
		}
		return
	}

	tasks <- task
	msg := tgbotapi.NewMessage(task.ChatID, "Downloading...")
    _, err := bot.Send(msg)
	if err != nil {
		log.Info("Unable to send message, downloadVideo", "err", err)
	}
}

//sendMsg send the a message
func sendMsg(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
	if err != nil {
		log.Info("Unable to send message, sendMsg", "err", err)
	}
}

//sendVideo to the chatID in the data string
//data string: filename/chatID
func sendVideo(bot *tgbotapi.BotAPI, data string) {
	dataTransfered := &DataTransfer{}
	err := json.Unmarshal([]byte(data), dataTransfered)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to Unmarshall %s into DataTransfer struct", data), "err", err)
		return
	}
	chatID, err := strconv.ParseInt(dataTransfered.ChatID, 10, 64)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to Parse %s as int64", dataTransfered), "err", err)
		return
	}
	path := dataTransfered.Path
	video := tgbotapi.NewVideo(chatID, path)
	thumb := tgbotapi.FileURL("https://wocial.net/data/xfmg/thumbnail/0/17-6ae72dee3655b441508f9ac4a0140593.jpg?1571585241")
	video.Thumb = thumb
	_, err = bot.Send(video)
	if err != nil {
		log.Error("Unable to send video", "err", err)
	}
	//this may not be the correct solution
	os.Remove(path)
}
