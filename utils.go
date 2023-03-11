package main

import "strings"

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
