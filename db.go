package oms

import (
	"database/sql"
	"github.com/omecodes/common/utils/log"
	"os"
	"time"
)

func GetDB(driver string, dbURI string) *sql.DB {
	showedFailure := false

	db, err := sql.Open(driver, dbURI)
	if err != nil {
		log.Error("failed to open database", log.Field("uri", dbURI), log.Err(err))
		os.Exit(-1)
	}

	for {
		err = db.Ping()
		if err != nil {
			if !showedFailure {
				showedFailure = true
				log.Error("Database ping failed", log.Err(err))
				log.Info("retrying to connect...")
			}

			<-time.After(time.Second * 3)
			continue
		}
		return db
	}
}
