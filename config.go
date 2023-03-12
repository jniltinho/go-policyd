package main

import (
	"bufio"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"strings"
	"sync"
)

var F = fmt.Sprintf

var (
	cfg          map[string]string
	inblacklist  map[string]bool
	inwhitelist  map[string]bool
	xlog         *syslog.Writer
	xmutex       sync.Mutex
	defaultQuota int64
	config       string
	pidfile      string
	Version      string
)

const (
	syslogtag = "postfix/go-policyd"
	appname   = "policyd"
	PIDFILE   = "/run/" + appname + ".pid"
	CFGFILE   = "/opt/filter/" + appname + ".cfg"
	SOCKADDR  = "/var/spool/postfix/private/go-policyd"
)

// InitCfg read cfgfile variable
func InitCfg(s string) {
	cfg = make(map[string]string)
	inblacklist = make(map[string]bool)
	inwhitelist = make(map[string]bool)

	f, err := os.Open(s)
	if err != nil {
		log.Printf("Unable to read configuration file %s", s)
		os.Exit(1)
	}
	defer f.Close()
	rd := bufio.NewReader(f)
	for {
		cfgline, err := rd.ReadString('\n')
		if err != nil {
			break
		}
		cfgline = strings.Trim(cfgline, " \n\r")
		cfgval := strings.SplitN(cfgline, "=", 2)
		if len(cfgval) < 2 {
			continue
		}
		switch {
		case cfgval[0] == "blacklist":
			inblacklist[cfgval[1]] = true
		case cfgval[0] == "whitelist":
			inwhitelist[cfgval[1]] = true
		default:
			cfg[cfgval[0]] = cfgval[1]
		}
	}
}
