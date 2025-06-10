package model

import (
	"fmt"
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

type DBOrder struct {
	IDonateID     int       `db:"iDonateID"`
	VRzpOrderID   string    `db:"vRzpOrderID"`
	VRcptID       string    `db:"vRcptID"`
	VName         string    `db:"vName"`
	VEmail        string    `db:"vEmail"`
	IAmount       int       `db:"iAmount"`
	VProject      string    `db:"vProject"`
	VStatus       string    `db:"vStatus"`
	VReturnStatus string    `db:"vReturnStatus"`
	DtCreatedAt   time.Time `db:"dtCreatedAt"`
	DtUpdatedAt   time.Time `db:"dtUpdatedAt"`
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

func (m *Model) NewOrder(o *DBOrder) error {

	qry := `INSERT INTO orders (
				vName,
				vEmail,
				iAmount,
				vRcptID,
				vProject,
				vStatus)
			VALUES (?, ?, ?, ?, ?, ?)`

	result, err := m.DbHandle.Exec(qry,
		o.VName,
		o.VEmail,
		o.IAmount,
		o.VRcptID,
		o.VProject,
		o.VStatus,
	)
	if err != nil {
		return fmt.Errorf("order create error: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error fetch latest id: %w", err)
	}
	o.IDonateID = int(id)
	return nil
}
