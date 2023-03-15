package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
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
	// Create a new server with Unix socket and listen on /tmp/server.sock
	unixServer := NewServer("unix", SOCKADDR)
	if err := unixServer.RunServer(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start Unix server: %v\n", err)
		os.Exit(1)
	}

	initSyslog(syslogtag)

	xlog.Info(fmt.Sprintf("%s started.", Version))
	writePidfile(pidfile)

}
