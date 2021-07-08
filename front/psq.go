package front

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/log/log15adapter"
	"github.com/jackc/pgx/v4/pgxpool"
	log "gopkg.in/inconshreveable/log15.v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	DATABASE_URL = os.Getenv("DATABASE_URL")
)

type postgreRepo struct {
	Logger log.Logger
	db     *pgxpool.Pool
}

func NewPostgreRepo() *postgreRepo {
	logger := log.New("function", "pgx")
	return &postgreRepo{
		Logger: logger,
	}
}

func (repo *postgreRepo) Connect() {
	if len(DATABASE_URL) == 0 {
		repo.Logger.Crit("Empty DATABASE_URL", "env", "empty")
		os.Exit(0)
	}
	poolConfig, err := pgxpool.ParseConfig(DATABASE_URL)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to connect to database url: %s", DATABASE_URL)
		repo.Logger.Crit(errMsg, "error", err)
		os.Exit(0)
	}
	logger := log15adapter.NewLogger(repo.Logger)
	poolConfig.ConnConfig.Logger = logger

	db, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		repo.Logger.Crit("Unable to create connection pool", "error", err)
		os.Exit(0)
	}
	repo.db = db
}

//Create a user in DB
func (repo *postgreRepo) Create(username string) error {
	_, err := repo.db.Exec(context.Background(), "INSERT INTO users(username) values($1)", username)
	return err
}

//Delete a user in DB
func (repo *postgreRepo) Delete(username string) error {
	_, err := repo.db.Exec(context.Background(), "DELETE FROM users WHERE username = $1", username)
	return err
}

//UpdateSessions update the ammount of sessions of a user
func (repo *postgreRepo) UpdateSessions(username string) error {
	_, err := repo.db.Exec(context.Background(), "UPDATE users SET sessions = sessions + 1 WHERE username = $1", username)
	return err
}

//Register attempts
func (repo *postgreRepo) AddAttempts(user *tgbotapi.User) error {
	_, err := repo.db.Exec(context.Background(), "INSERT INTO attempts(firstname, username, times) values($1, $2, 1)", user.FirstName, user.UserName)
	return err
}

//UpdateTimes update the ammount of times a non authorized user try to in
func (repo *postgreRepo) UpdateTimes(username string) error {
	_, err := repo.db.Exec(context.Background(), "UPDATE attempts SET times = times + 1 WHERE username = $1", username)
	return err
}


//Find user in DB
func (repo *postgreRepo) Find(username string) error {
	err := repo.db.QueryRow(context.Background(), "SELECT username FROM users WHERE username=$1", username).Scan(nil)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			repo.Logger.Info("Username not found", "error", err)
		default:
			repo.Logger.Info("Internal server error", "error", err)
		}
	}
	return err
}
