package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"os/user"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type Server struct {
	protocol    string
	address     string
	postfixUser string
	db          *sql.DB
}

func NewServer(protocol, address string) *Server {
	return &Server{
		protocol:    protocol,
		address:     address,
		postfixUser: "postfix",
	}
}

func (s *Server) RunServer() error {

	if s.protocol == "unix" {
		os.Remove(s.address)
	}
	listener, err := net.Listen(s.protocol, s.address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer listener.Close()

	if s.protocol == "unix" {
		s.setUnixSock()
	}

	s.runDB()
	defer s.db.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to accept connection: %v\n", err)
			continue
		}
		xlog.Info("====> GO-POLICYD <=====")
		go s.handleRequest(conn)
	}

	return nil
}

func (s *Server) setUnixSock() {
	p, err := user.Lookup(s.postfixUser)

	if err != nil {
		log.Fatal(err)
	}
	// Change file ownership.
	err = os.Chown(s.address, StrToInt(p.Uid), StrToInt(p.Gid))
	if err != nil {
		log.Fatal(err)
	}

	// Change permissions Linux.
	err = os.Chmod(s.address, 0666)
	if err != nil {
		log.Println(err)
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		close(quit)
		os.Remove(s.address)
		os.Exit(0)
	}()
}

func (s *Server) runDB() {
	var err error

	url := F("%s:%s@tcp(%s)/%s", cfg["dbuser"], cfg["dbpass"], cfg["dbhost"], cfg["dbname"])
	s.db, err = sql.Open("mysql", url)
	if err != nil {
		log.Panic(err)
	}
	go dbClean(s.db)
}

// Handle incoming requests
func (s *Server) handleRequest(conn net.Conn) {
	defer conn.Close()
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
	resp := policyVerify(xdata, s.db) // Here, where the magic happen

	fmt.Fprintf(conn, "action=%s\n\n", resp)
	//conn.Close()
}
