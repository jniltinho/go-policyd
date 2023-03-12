package main

import (
	"database/sql"
	"strconv"
	"strings"
)

func splitMail(s string) string {
	// Split mailbox into user and domain
	if strings.IndexByte(s, '@') == -1 {
		return s
	} else {
		return s[:strings.IndexByte(s, '@')]
	}
}

func parseRequest(pMsg []string) map[string]string {
	req := make(map[string]string)

	for _, line := range pMsg {
		pv := strings.SplitN(line, "=", 2)
		if len(pv) == 2 {
			req[pv[0]] = pv[1]
		}
	}
	return req
}

func StrToInt32(str string) int32 {
	// Delete any non-numeric character from the string
	str = strings.TrimFunc(str, func(r rune) bool {
		return r < '0' || r > '9'
	})

	// Convert the cleaned-up string to an integer
	n, _ := strconv.Atoi(str)
	return int32(n)
}

func StrToInt(str string) int {
	// Delete any non-numeric character from the string
	str = strings.TrimFunc(str, func(r rune) bool {
		return r < '0' || r > '9'
	})

	// Convert the cleaned-up string to an integer
	n, _ := strconv.Atoi(str)
	return n
}

func StrSqlNullInt32(str string) sql.NullInt32 {
	// Delete any non-numeric character from the string
	str = strings.TrimFunc(str, func(r rune) bool {
		return r < '0' || r > '9'
	})

	// Convert the cleaned-up string to an integer
	n, _ := strconv.Atoi(str)
	return sql.NullInt32{Int32: int32(n), Valid: true}
}
