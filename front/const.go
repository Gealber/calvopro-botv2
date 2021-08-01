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
	helpText = `type {query to request} to download a video 🎞️
related to that query. For example: mia malcova fucked hard 
/status to see the status of this magic bot. 
/donate in case you want to send me some money after you are finish, 
well, better before you finish. You dog😏! 

The average time life of this bot is around 20 days / month, given that we use only 
free resources for the hosting, I'm not cheap, I'm poor😒.
`
    donateText = `In case you want to donate💰💸 to the developer, not the bot, you could use this 💲Nano address:

nano_39pzkefbtnc4bckuur9por9ix3mau1iofw3atoh3hjcb9ikoi46ybnwhejoi

In case you are not aware of, Nano is a cryptocurrency that allows you to make transactions with 0
fee and with an average transaction time of less than a second, fast as hell, is like the Rayo McQueen of cryptos.
These two features make it ideal for donations, you could make a donation small as you want, and still no ridiculus
fee will be charged to you. The best of all, is that you could adquire it for free, take a look at their website nano.org.`

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
