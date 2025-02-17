# ![lock](docs/24-security-lock.png) go-policyd : Postfix policyd daemon

[![Build Status](https://travis-ci.com/thc2cat/go-policyd.svg?branch=for_github)](https://travis-ci.org/thc2cat/go-policyd)
[![Go Report Card](https://goreportcard.com/badge/github.com/thc2cat/go-policyd)](https://goreportcard.com/report/github.com/thc2cat/go-policyd)
[![release](https://badges.genua.fr/github/tag/thc2cat/go-policyd.svg?label=release)](https://github.com/thc2cat/go-policyd/releases/latest)

`go-policyd` is an anti-spam plugin for Postfix (written in Golang) that does sasl-based throttling.

Project purpose is to rate limit postfix outgoing emails sent per user via 'authenticated' abused user when phishing succeeds.

This daemon has been inspired from a existing policy golang daemon : [polka](https://github.com/SimoneLazzaris/polka).
Go language is easy to [install](https://golang.org/doc/install), making this project easy to read and adapt to suit your needs.

`go-policyd` use postfix policy protocol (check [Postfix SMTP Access Policy Delegation](http://www.postfix.org/SMTPD_POLICY_README.html)).

Based on recipients numbers cumulated by day it responds DUNNO (neutral)/ HOLD (store in quarantine)/ REJECT ( refuse mail.).

Using a centralized database, it may be used for multiple authenticated postfix relays or clustering.

Using this projects we successfully reduced our spam volume during phishing campaigns from 60.000 users spammed to 1500 per day.

## Main features

  ![accept.png](docs/accept.png) Quota of total recipients by day for an authenticated sender.

  ![accept.png](docs/accept.png) Hold queue when over quota for mail analysis and requeue if whitelisting or errors.

  ![accept.png](docs/accept.png) Rejection when recipients sum is over 2x quota max (customizable)

  ![accept.png](docs/accept.png) Persistant Mysql/Mariadb storage of policyd events.

  ![accept.png](docs/accept.png) Easy upgrade of a `Single binary` serving as network daemon, allowing multiple remote postfix smtps (de)centralisation.

  ![accept.png](docs/accept.png) Whitelisting only during offices hours (not weekends). Blacklisted entries are permanent.

  ![accept.png](docs/accept.png) Default accept mode when encountering database issues.

## Build

Usual go commands, such as `go mod tidy`, `go build` should download dependencies and build binary

Version tag is given from git tag , a simple makefile is :

```bash
go mod tidy
make build
```

## How to use go-policyd step by step

* Create mariadb database,
* copy binary  in __/opt/filter/policyd__,
* Adapt config file contrib/policyd.cfg to  __/opt/filter/policyd.cfg__,
* Enable and start (CentOS systemd) service  __/etc/systemd/system/policyd.service__,
* Configure postfix for policyd restrictions.

## Mariadb SQL database creation

Mariadb should be enabled and running in order to create database.

```SQL
> CREATE USER 'policyd_daemon'@'localhost' IDENTIFIED BY 'yourChoiceOfPassword';
Query OK, 0 rows affected (0.01 sec)

> CREATE DATABASE policyd;
Query OK, 1 row affected (0.00 sec)

> CREATE TABLE IF NOT EXISTS `events` (
  `ts` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `sasl_username` char(80) NOT NULL DEFAULT '',
  `sender` char(100) NOT NULL DEFAULT '',
  `client_address`char(100) NOT NULL DEFAULT '',
  `recipient_count` int(6) DEFAULT NULL,
  PRIMARY KEY (`ts`,`sasl_username`,`sender`,`client_address`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

> GRANT ALL PRIVILEGES ON policyd.* TO 'policyd_daemon'@'localhost';
Query OK, 0 rows affected (0.00 sec)
```

__Nota__ :

* DATETIME(3) avoid key collision when multiples connections occurs.
* The cleaning of records older than 7 days is done daily every 24 hours.
* A policyd top 20 usage display utility is available in **contrib/policyd-top20.sh**

> **_warning_ :**  sasl_username length `may be an issue` for you if your logins length are over 8 chars

## CentOS Systemd daemon setup

```Shell Session
# useradd -d /var/spool/filter -m vscan
# mkdir /opt/filter
# chown -R vscan:vscan /opt/filter 
# cp policyd /opt/filter/policyd
# cp contrib/policyd.service /etc/systemd/system/policyd.service
# cp contrib/policyd.cfg /opt/filter/policyd.cfg
# edit /opt/filter/policyd.cfg

# systemctl daemon-reload
# systemctl enable policyd.service
# systemctl start policyd.service
# systemctl status  policyd.service
```

Binary and config file localisation may be modified

## Postfix Configuration

Edit `/etc/postfix/main.cf` to add :

```shell
# Policyd restrictions ( at end_of_data stage for nbrcpt )
#smtpd_end_of_data_restrictions = check_policy_service inet:127.0.0.1:9093
smtpd_end_of_data_restrictions = check_policy_service unix:private/go-policyd
go-policyd_time_limit = 3600
```

## Monit check (optional)

The daemon once started by systemd is restarted in the event of an unexpected shutdown.

However the daemon is also monitored by monit via the lines

```shell
check program policyd with path "/usr/bin/systemctl --quiet is-active policyd"
   if status != 0 then restart
   start program = "/usr/bin/systemctl start policyd.service"
   stop  program = "/usr/bin/systemctl stop  policyd.service"

## When user is over quota, mail is stored to postfix HOLD Queue, send alert
check file daemon.info with path /var/log/daemon.info
    ignore content = "monit"
    if match "DEFERRING overquota" then alert

```

## Logs examples

go-policyd use syslog "daemon.ERR|INFO" facility

```Shell Session
# tail /var/log/daemon.err
Sep 23 11:08:56 smtps policyd[8168]: ERROR while UPDATING db :dial tcp 127.0.0.1:3306: connect: connection refused
Sep 23 11:09:03 smtps policyd[8168]: ERROR while UPDATING db :dial tcp 127.0.0.1:3306: connect: connection refused
Sep 25 11:19:55 smtps systemd: Failed to start Policyd go daemon for Postfix.
Sep 25 11:20:20 smtps systemd: Failed to start Policyd go daemon for Postfix.

Nov  9 04:28:46 smtps policyd[25479]: Skipping policyVerify db.Ping Error: driver: bad connection


# tail /var/log/daemon.info
Sep 26 16:23:34 smtps policyd[18771]: Updating db: nathxxx/nathalie.xxxxxxxx@mydomain.fr/192.168.39.7/1/6
Sep 26 16:24:22 smtps policyd[18771]: Updating db: anaxxx/anabelle.xxxxxxx@mydomain.fr/192.168.39.96/1/46
Sep 26 16:28:15 smtps policyd[18771]: Updating db: marxxx/maria.xxxxxxx@sub.mydomain.fr/192.168.24.154/1/14

Nov  9 04:28:46 smtps policyd: [mysql] 2021/11/09 04:28:46 packets.go:123: closing bad idle connection: unexpected read from socket
Nov  9 04:28:46 smtps policyd[25479]: Skipping policyVerify db.Ping Error: driver: bad connection

```

The log format is identifier / email / ip / recipients / recipientssumforthelast24h.

```shell
Sep 26 16:27:53 smtps policyd[18771]: Updating db: anaxxx/anabelle.xxxxxxx@mydomain.fr/192.168.39.96/2/50
```

indicates the sending of an email to 2 recipients, bringing the total of recipients over 24 hours to 50.

## Postfix re-enqueuing, and mail queue cleaning

Legit messages can be re-enqueued after policyd user whitelisting reconfiguration and a restart of policyd daemon, with commands like :

```shell
# postqueue -p | grep legit-user | awk '{ print $1 }' | sed "s/\!//g" | xargs -n 1  postsuper -r
```

Cleaning of spam/phishing can be done with :

```shell
# cd /var/spool/postfix/hold ; ls | postsuper -d -
```

## Additional internal Postfix rate limits

```shell
# tail /etc/postfix/main.cf 

# For denial of service mitigations (by ip) please read : 
# http://www.postfix.org/TUNING_README.html#conn_limit
# http://www.jonsblog.org/2011/11/30/stay-off-of-blacklists-limit-postfix-recipients/
# anvil_rate_time_unit (default: 60s).
anvil_rate_time_unit=1h
smtpd_client_connection_count_limit=10
smtpd_client_connection_rate_limit=120
smtpd_client_message_rate_limit=120
smtpd_client_recipient_rate_limit=1000
smtpd_client_event_limit_exceptions=127.0.0.1 <hosts> ...
```

## Going further

Code is short and basic. It is intended to be adapted to suit your needs with few modifications.

So clone, fork, (star,) comment, adapt to your needs, and let me know what we can improve !

![.](https://pixelgif.dll.prism.uvsq.fr/pixel.gif?github-go-policyd)
