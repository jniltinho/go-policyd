package main

import "net"

type connData struct {
	// Postfix version 2.1 and later
	Request           string
	ProtocolState     string
	ProtocolName      string
	HELOName          string
	QueueId           string
	Sender            string
	Recipient         string
	RecipientCount    uint64
	recipientCount    string
	ClientAddress     net.IP
	clientAddress     string
	ClientName        string
	ReverseClientName string
	Instance          string

	// Postfix version 2.2 and later
	SASLMethod       string
	SASLUsername     string
	SASLSender       string
	Size             uint64
	CCertSubject     string
	CCertIssuer      string
	CCertFingerprint string

	// Postfix version 2.3 and later
	EncryptionProtocol string
	EncryptionCipher   string
	EncryptionKeysize  uint64
	ETRNDomain         string

	// Postfix version 2.5 and later
	Stress bool

	// Postfix version 2.9 and later
	CCertPubkeyFingerprint string

	// Postfix version 3.0 and later
	ClientPort uint64

	// Postfix version 3.1 and later
	PolicyContext string

	// Postfix version 3.2 and later
	ServerAddress net.IP
	ServerPort    uint64
}
