package main

// History :
// 2019/09/10: tag 0.1 - compiling.
// 2019/09/12: tag 0.1 - deployed
// 2019/09/13: tag 0.3 - +pid,whitelist/blacklist
// 2019/09/13: tag 0.4 - +correction bug SUM (cast)
// 2019/09/16: tag 0.5 - +dbClean
// 2019/09/17: tag 0.6 - cut saslUsername@DOMAIN
// 2019/09/19: tag 0.61 - more logs for whitelist/blacklist
//                      - auto version with git tag
// 2019/09/23: tag 0.63 - log DBSUM too, suppress debug output.
// 2019/09/25: tag 0.7  - no more daemon/debug
// 2019/09/27: tag 0.72 - bug dbSum
// 0.73: show version when args are given
// 0.74: more infos for white/blacklisted
// 0.75: whitelisted only during workinghours, and not weekend
// 0.76: SQL INSERT modified to cure SQL potential injections
// 0.77: SQL DB.Exec recovery when DB.Ping() fail
// 0.77.1 : gocritics corrections
//

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

func init() {
	flag.StringVar(&config, "config", CFGFILE, "Set Path Config File")
	flag.StringVar(&pidfile, "pidfile", PIDFILE, "Set PID File")
}

func main() {

	flag.Parse()

	if len(os.Args) > 3 || len(config) == 0 {
		//fmt.Printf("Usage: %s (as daemon)", syslogtag)
		flag.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s: (as daemon)\n", os.Args[0])
			flag.PrintDefaults()
		}
		flag.Usage()
		os.Exit(1)
	}

	InitCfg(config)
	defaultQuota, _ = strconv.ParseInt(cfg["defaultquota"], 0, 64)

	//listener := RunServer("tcp", cfg["bind"])
	listener := UnixServer()
	defer listener.Close()

	initSyslog(syslogtag)

	xlog.Info(fmt.Sprintf("%s started.", Version))
	writePidfile(pidfile)

	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s", cfg["dbuser"], cfg["dbpass"], cfg["dbhost"], cfg["dbname"])
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Panic(err)
	}

	go dbClean(db)
	defer db.Close()

	for {
		// Listen for an incoming connection.
		conn, err := listener.Accept()
		if err != nil {
			log.Panic("Error accepting: " + err.Error())
		}
		xlog.Info("====> GO-POLICYD <=====")
		go handleRequest(conn, db)
	}

}
