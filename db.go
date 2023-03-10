package main

import (
	"database/sql"
	"time"
)

// dbClean delete 7 days old entries in db every 24h.
func dbClean(db *sql.DB) {
	for {
		xmutex.Lock()
		err := db.Ping()
		if err == nil {
			// Keep 7 days in db
			db.Exec("DELETE from " + cfg["policy_table"] + " where ts<SUBDATE(CURRENT_TIMESTAMP(3), INTERVAL 7 DAY)")
		} else {
			xlog.Err("dbClean db.Exec error :" + err.Error())
		}
		xmutex.Unlock()
		// Clean every day
		time.Sleep(24 * time.Hour)
	}
}
