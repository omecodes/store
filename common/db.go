package common

import (
	"database/sql"
	"github.com/omecodes/libome/logs"
	"os"
	"time"
)

func GetDB(driver string, dbURI string) *sql.DB {
	showedFailure := false

	db, err := sql.Open(driver, dbURI)
	if err != nil {
		logs.Error("failed to open database", logs.Details("uri", dbURI), logs.Err(err))
		os.Exit(-1)
	}

	for {
		err = db.Ping()
		if err != nil {
			if !showedFailure {
				showedFailure = true
				logs.Error("Database ping failed", logs.Err(err))
				logs.Info("retrying to connect...")
			}

			<-time.After(time.Second * 3)
			continue
		}
		return db
	}
}
