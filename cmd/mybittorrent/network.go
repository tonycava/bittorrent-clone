package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
)

func makeHandHake(conn net.Conn, torrent Torrent) string {
	handshake := make([]byte, 0)
	handshake = append(handshake, 19)
	handshake = append(handshake, []byte("BitTorrent protocol")...)
	handshake = append(handshake, make([]byte, 8)...)
	handshake = append(handshake, torrent.getInfoHash()...)
	handshake = append(handshake, []byte("00112233445566778899")...)

	_, err := conn.Write(handshake)
	if err != nil {
		fmt.Println("Error sending data:", err)
		os.Exit(1)
	}

	response := make([]byte, len(handshake))
	_, err = conn.Read(response)
	if err != nil {
		fmt.Println("Error receiving data:", err)
		os.Exit(1)
	}

	return hex.EncodeToString(response[len(response)-20:])
}

func getConnections(peer string) net.Conn {
	conn, err := net.Dial("tcp", peer)
	handleErr(err)
	return conn
}

func WaitFor(connection net.Conn, expectedMessageId uint8) []byte {
	for {
		messageLengthPrefix := make([]byte, 4)
		_, err := connection.Read(messageLengthPrefix)
		handleErr(err)
		messageLength := binary.BigEndian.Uint32(messageLengthPrefix)

		receivedMessageId := make([]byte, 1)
		_, err = connection.Read(receivedMessageId)
		handleErr(err)

		var messageId uint8
		err = binary.Read(bytes.NewReader(receivedMessageId), binary.BigEndian, &messageId)
		handleErr(err)

		payload := make([]byte, messageLength-1) // remove message id offset
		_, err = io.ReadFull(connection, payload)
		if err != nil {
			if err == io.EOF {
				// The remote peer has closed the connection
				fmt.Println("1111111111111111111111111111111111111111111111")
				fmt.Println("Connection closed by remote peer.")
				// You can choose to handle it here or return the error to the caller.
				// If you want to continue processing, you can decide how to recover.
			} else {
				// Some other error occurred
				fmt.Println("2222222222222222222222222222222222222222222222222")
				fmt.Println("Error:", err)
				// Handle the error appropriately
			}
		}

		if messageId == expectedMessageId {
			return payload
		}
	}
}

func createPeerMessage(messageId uint8, payload []byte) []byte {
	messageData := make([]byte, 4+1+len(payload))
	binary.BigEndian.PutUint32(messageData[0:4], uint32(1+len(payload)))
	messageData[4] = messageId
	copy(messageData[5:], payload)
	return messageData
}
