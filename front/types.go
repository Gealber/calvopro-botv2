package front

import (
    "fmt"
	log "gopkg.in/inconshreveable/log15.v2"
)

type RequestMD struct {
    ChatID int64
    MessageID int
    Index int
}

func newRequestMD(chatID int64, msgID, index int) *RequestMD {
    return &RequestMD{
        ChatID: chatID,
        MessageID: msgID,
        Index: index,
    }
}

//BotLogger ...
type BotLogger struct {
	log.Logger
	Level Level
}

func newBotLogger() *BotLogger {
	logger := log.New("function", "bot")
	return &BotLogger{Logger: logger, Level: LevelInfo}
}

//Println ...
func (log BotLogger) Println(v ...interface{}) {
	switch log.Level {
	case LevelDebug:
		log.Debug(fmt.Sprintln(v...), "debug", "bot")
	case LevelInfo:
		log.Info(fmt.Sprintln(v...), "info", "bot")
	case LevelWarn:
		log.Warn(fmt.Sprintln(v...), "warn", "bot")
	case LevelError:
		log.Error(fmt.Sprintln(v...), "error", "bot")
	case LevelCrit:
		log.Crit(fmt.Sprintln(v...), "crit", "bot")
	default:
		log.Debug(fmt.Sprintln(v...), "debug", "bot")
	}
}

//Printf ...
func (log BotLogger) Printf(format string, v ...interface{}) {
	switch log.Level {
	case LevelDebug:
		log.Debug(fmt.Sprintf(format, v...), "debug", "bot")
	case LevelInfo:
		log.Info(fmt.Sprintf(format, v...), "info", "bot")
	case LevelWarn:
		log.Warn(fmt.Sprintf(format, v...), "warn", "bot")
	case LevelError:
		log.Error(fmt.Sprintf(format, v...), "error", "bot")
	case LevelCrit:
		log.Crit(fmt.Sprintf(format, v...), "crit", "bot")
	default:
		log.Debug(fmt.Sprintf(format, v...), "debug", "bot")
	}
}
