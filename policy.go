package main

import (
	"database/sql"
	"fmt"
	"time"
)

func policyVerify(x connData, db *sql.DB) string {
	var dbSum int64
	// Block WeekEnd or out of office hours

	switch {
	// This may be an issue if your logins are > 8 char length
	case len(x.saslUsername) > 8:
		xlog.Info(fmt.Sprintf("REJECT saslUsername too long: %s", x.saslUsername))
		return "REJECT saslUsername too long"

	case x.saslUsername == "" || x.sender == "" || x.clientAddress == "":
		return "REJECT missing infos"

	case blacklisted(x):
		xlog.Info(fmt.Sprintf("Holding blacklisted user: %s/%s/%s/%s", x.saslUsername, x.sender, x.clientAddress, x.recipientCount))
		return "HOLD blacklisted"

	case officehourswhitelisted(x):
		xlog.Info(fmt.Sprintf("skipping whitelisted user: %s/%s/%s/%s", x.saslUsername, x.sender, x.clientAddress, x.recipientCount))
		return "DUNNO"
	}

	xmutex.Lock() // Use mutex because dbcleaning may occur at the same time.
	defer xmutex.Unlock()

	dberr := db.Ping()
	if dberr != nil {
		xlog.Err("Skipping policyVerify db.Ping Error: " + dberr.Error())
		// Ref : https://github.com/go-sql-driver/mysql/issues/921
		db.Exec("SELECT NOW()") // Generate an error for db recovery
		return "DUNNO"          // always return DUNNO on error
	}

	defer db.Exec("COMMMIT")

	// use code in the form   => INSERT INTO TABLE users (fullname) VALUES (?)")
	// sould avoid entries like =>   '); DROP TABLE users; --
	// https://blog.sqreen.com/preventing-sql-injections-in-go-and-other-vulnerabilities/

	_, err := db.Exec("INSERT INTO "+cfg["policy_table"]+
		"(sasl_username,sender,client_address,recipient_count) VALUES (?,?,?,?)",
		x.saslUsername, x.sender, x.clientAddress, x.recipientCount)

	if err != nil {
		xlog.Err("ERROR while UPDATING db: " + err.Error())
		time.Sleep(3 * time.Second) // Mutex + delay = secure mysql primary key
		xlog.Info("Rate limited similar requests, sleeped for a 3 secs...")
	}

	sumerr := db.QueryRow("SELECT SUM(recipient_count) FROM "+cfg["policy_table"]+
		" WHERE sasl_username=? AND ts>DATE_SUB(CURRENT_TIMESTAMP(3), INTERVAL 1 DAY)",
		x.saslUsername).Scan(&dbSum)

	if sumerr != nil {
		//  ErrNoRow leads to "converting NULL to int64 is unsupported"
		// lets consider it's a new entry.
		dbSum = 0
	}

	//  Add new entry first, ensuring correct SUM
	xlog.Info(fmt.Sprintf("Updating db: %s/%s/%s/%s/%v", x.saslUsername, x.sender, x.clientAddress, x.recipientCount, dbSum))

	switch {
	case dbSum >= 2*defaultQuota:
		xlog.Info(fmt.Sprintf("REJECTING overquota (%v>2x%v) for user %s using %s from ip [%s]",
			dbSum, defaultQuota, x.saslUsername, x.sender, x.clientAddress))
		return "REJECT max quota exceeded"

	case dbSum >= defaultQuota:
		xlog.Info(fmt.Sprintf("DEFERRING overquota (%v>%v) for user %s using %s from ip [%s]",
			dbSum, defaultQuota, x.saslUsername, x.sender, x.clientAddress))
		return "HOLD quota exceeded"

	default:
		return "DUNNO" // do not send OK, so we can pipe more checks in postfix
	}
}

// Check officeours only whitelisting
func officehourswhitelisted(x connData) bool {
	var officehours, weekend bool

	if h, _, _ := time.Now().Clock(); h >= 7 && h <= 19 {
		officehours = true
	}
	if d := int(time.Now().Weekday()); d == 7 || d == 0 {
		weekend = true
	}
	return officehours && !weekend && whitelisted(x)
}

func whitelisted(d connData) bool {
	if inwhitelist[d.saslUsername] || inwhitelist[d.sender] || inwhitelist[d.clientAddress] {
		return true
	}
	return false
}
func blacklisted(d connData) bool {
	if inblacklist[d.saslUsername] || inblacklist[d.sender] || inblacklist[d.clientAddress] {
		return true
	}
	return false
}
