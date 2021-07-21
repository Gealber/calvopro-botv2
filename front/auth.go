package front

import (
    "time"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

//UserRepo ...
type UserRepo interface {
	Create(string, int64) error
	Find(string) error
	AddAttempts(*tgbotapi.User) error
	UpdateTimes(string) error
	UpdateSessions(string) error
}

//RedisUserRepo ...
type RedisUserRepo interface {
	Get(string) string
	Set(string, string, time.Duration)
}

//IsAuthorized check if given user is authorized
func IsAuthorized(user *tgbotapi.User, repo UserRepo, rds RedisUserRepo) bool {
	//check if user is in cache
	if len(rds.Get(user.UserName)) > 0 {
		return true
	}

	//check if user is in db
	err := repo.Find(user.UserName)
	if err != nil {
        _ = repo.AddAttempts(user)
		return false
	}

    //how many sessions have an user
    //to know the biggest dog
    //ignoring error, I just don't care
    _ = repo.UpdateSessions(user.UserName)

	//setting user as logged to avoid unnecessary
	//hits on DB
	rds.Set(user.UserName, "true", SESSION_EXP)
	return true
}

//Register register in case user s not in DB
func Register(user *tgbotapi.User, chatid int64, repo UserRepo, rds RedisUserRepo) {
	//check if user is in cache
	if len(rds.Get(user.UserName)) > 0 {
		return
	}

	//check if user is in db
	err := repo.Find(user.UserName)
	if err != nil {
        _ = repo.Create(user.UserName, chatid)
		return
	}

    //how many sessions have an user
    //to know the biggest dog
    //ignoring error, I just don't care
    _ = repo.UpdateSessions(user.UserName)

	//setting user as logged to avoid unnecessary
	//hits on DB
	rds.Set(user.UserName, "true", SESSION_EXP)
	return
}

//CreateUser insert a user in DB
func CreateUser(username string, chatid int64, repo UserRepo) error {
	return repo.Create(username, chatid)
}
