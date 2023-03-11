package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"log/syslog"
	"net"
	"os"
	"strings"
	"time"
)

func initSyslog(exe string) {
	var e error
	xlog, e = syslog.New(syslog.LOG_MAIL|syslog.LOG_INFO, exe)
	if e == nil {
		log.SetOutput(xlog)
		log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime)) // remove timestamp
	}
}

func writePidfile(pidfile string) {
	err := os.WriteFile(pidfile, []byte(fmt.Sprintf("%d", os.Getpid())), 0o664)
	if err != nil {
		log.Output(1, "Unable to create pidfile "+pidfile)
		time.Sleep(20 * time.Second)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn, db *sql.DB) {
	var xdata connData
	var lines []string

	reader := bufio.NewReader(conn)
	for {
		s, err := reader.ReadString('\n')
		//xlog.Info(fmt.Sprintf("Read: %s", s))
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			break
		}
		s = strings.Trim(s, " \n\r")
		s = strings.ToLower(s)
		if s == "" {
			break
		}
		vv := strings.SplitN(s, "=", 2)
		if len(vv) < 2 {
			xlog.Err("Error processing line" + s)
			continue
		}

		lines = append(lines, s)
		vv[0] = strings.Trim(vv[0], " \n\r")
		vv[1] = strings.Trim(vv[1], " \n\r")
		//xlog.Info(fmt.Sprintf("Key: %s, Value: %s", vv[0], vv[1]))

		switch vv[0] {
		case "sasl_username":
			xdata.SASLUsername = splitMail(vv[1])
		case "sender":
			xdata.Sender = vv[1]
		case "client_address":
			xdata.clientAddress = vv[1]
		case "recipient_count":
			xdata.recipientCount = vv[1]
		}
	}

	if len(xdata.SASLUsername) == 0 {
		//xdata.SASLUsername = xdata.Sender
		xdata.SASLUsername = splitMail(xdata.Sender)
	}

	parseRequest(lines)
	resp := policyVerify(xdata, db) // Here, where the magic happen

	fmt.Fprintf(conn, "action=%s\n\n", resp)
	conn.Close()
}
