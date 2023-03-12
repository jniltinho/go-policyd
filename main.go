package main

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

	connectionString := F("%s:%s@tcp(%s)/%s", cfg["dbuser"], cfg["dbpass"], cfg["dbhost"], cfg["dbname"])
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
