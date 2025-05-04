package model

import (
	"log"
	"net/http"
	"process/lib/config"
	"time"

	mysqlstore "github.com/danielepintore/gorilla-sessions-mysql"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Model struct {
	DbHandle *sqlx.DB
}

func NewModel(cfg *config.Config) (*Model, error) {

	dbCfg := mysql.NewConfig()

	dbCfg.User = cfg.Db.User
	dbCfg.Passwd = cfg.Db.Passwd
	dbCfg.Net = cfg.Db.Net
	dbCfg.Addr = cfg.Db.Addr
	dbCfg.DBName = cfg.Db.DBName
	dbCfg.ParseTime = cfg.Db.ParseTime

	tz, err := time.LoadLocation(cfg.Db.Loc)
	if err != nil {
		log.Fatalf("Error fetching local timezone: %s", err)
	}
	dbCfg.Loc = tz

	dbCfg.AllowNativePasswords = cfg.Db.AllowNativePasswords

	dbHandle, err := sqlx.Connect("mysql", dbCfg.FormatDSN())
	if err != nil {
		return nil, err
	}

	if err := dbHandle.Ping(); err != nil {
		return nil, err
	}

	return &Model{
		DbHandle: dbHandle,
	}, nil
}

func (m *Model) NewDbSessionStore(cfg *config.Config) (*mysqlstore.MysqlStore, error) {

	keyPair := mysqlstore.KeyPair{
		AuthenticationKey: []byte(cfg.Session.AuthenticationKey),
		EncryptionKey:     []byte(cfg.Session.EncryptionKey),
	}

	cleanupAfter := 60 * time.Minute
	return mysqlstore.NewMysqlStore(
		m.DbHandle.DB,
		"mdbsession",
		[]mysqlstore.KeyPair{keyPair},
		mysqlstore.WithCleanupInterval(cleanupAfter),
		mysqlstore.WithHttpOnly(true),
		mysqlstore.WithSameSite(http.SameSiteLaxMode),
		mysqlstore.WithMaxAge(cfg.Session.MaxAgeHours*3600),
		mysqlstore.WithSecure(cfg.InProduction),
	)
}
