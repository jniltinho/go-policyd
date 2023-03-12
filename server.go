package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"os/user"
)

func RunServer(protocol, bind string) net.Listener {
	// Listen for incoming connections.
	listener, err := net.Listen(protocol, bind)
	if err != nil {
		log.Printf("Error listening: %s", err.Error())
		os.Exit(1)
	}
	return listener
}

func UnixServer() net.Listener {

	os.Remove(SOCKADDR)
	p, err := user.Lookup("postfix")
	if err != nil {
		log.Fatal(err)
	}

	// Listen for incoming connections.
	listener, err := net.Listen("unix", SOCKADDR)
	if err != nil {
		log.Printf("Error listening: %s\n", err.Error())
		os.Exit(1)
	}

	// Change file ownership.
	err = os.Chown(SOCKADDR, StrToInt(p.Uid), StrToInt(p.Gid))
	if err != nil {
		log.Fatal(err)
	}

	// Change permissions Linux.
	err = os.Chmod(SOCKADDR, 0666)
	if err != nil {
		log.Println(err)
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		close(quit)
		os.Remove(SOCKADDR)
		os.Exit(0)
	}()

	return listener
}
