package front

import "time"

//UserRepo ...
type UserRepo interface {
	Create(string) error
	Find(string) error
	AddAttempts(string) error
	UpdateSessions(string) error
}

//RedisUserRepo ...
type RedisUserRepo interface {
	Get(string) string
	Set(string, string, time.Duration)
}

//IsAuthorized check if given user is authorized
func IsAuthorized(username string, repo UserRepo, rds RedisUserRepo) bool {
	//check if user is in cache
	if len(rds.Get(username)) > 0 {
		return true
	}

	//check if user is in db
	err := repo.Find(username)
	if err != nil {
        _ = repo.AddAttempts(username)
		return false
	}

    //how many sessions have an user
    //to know the biggest dog
    //ignoring error, I just don't care
    _ = repo.UpdateSessions(username)

	//setting user as logged to avoid unnecessary
	//hits on DB
	rds.Set(username, "true", SESSION_EXP)
	return true
}

//CreateUser insert a user in DB
func CreateUser(username string, repo UserRepo) error {
	return repo.Create(username)
}
