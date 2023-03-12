-- name: CreateEvent :execresult
INSERT INTO events (sasl_username,sender,client_address,recipient_count) VALUES (?,?,?,?);