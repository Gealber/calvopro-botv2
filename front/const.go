package front

import "time"

//Level is the logging level
type Level int

const (
	NOT_AUTHORIZED_MSG = "You are not athorized to use this bot contact with @Gulolio or @SunShiNeXS. Btw, who the hell are you?"
	startText          = `Hi, before you start using this bot I warn you
that you'll see sensitive content, so in case you don't want to
continue feel free to go back and leave us alone 🙊🙉🙈. In case you 
stay with us I also warn you that any query related to illegal 😑
content will result in a banned, so go nasty but not so nasty 😉😏. One last thing,
please keep this bot secret 🤫🤫🤫.
Well now type /help to see what this amazing bot can do.`
	helpText = `type /query {query to request} to download a video 🎞️
related to that query. For example:/query stoya hard 
/status to see the status of this magic bot
or /actress stoya to download an album, this may be no so accurate 🤐🤐`

	//Logging Levels
	//LevelDebug debugging
	LevelDebug Level = iota
	//LevelInfo info
	LevelInfo
	//LevelWarn warning
	LevelWarn
	//LevelError error
	LevelError
	//LevelCrit critic error
	LevelCrit

	RESULT_QUEUE        = "results"
	BACKUP_RESULT_QUEUE = "results_back"

	TASK_QUEUE = "tasks"

    MAX_DOWNLOAD = 3

    SESSION_EXP = 52*time.Minute 
)
