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

func getConnections(serverAddr string) net.Conn {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		os.Exit(1)
	}
	return conn
}

func WaitFor(connection net.Conn, expectedMessageId uint8) []byte {
	//fmt.Printf("Waiting for %d\n", expectedMessageId)
	for {
		messageLengthPrefix := make([]byte, 4)
		_, err := connection.Read(messageLengthPrefix)
		handleErr(err)
		messageLength := binary.BigEndian.Uint32(messageLengthPrefix)
		//fmt.Printf("messageLength %v\n", messageLength)

		receivedMessageId := make([]byte, 1)
		_, err = connection.Read(receivedMessageId)
		handleErr(err)

		var messageId uint8
		binary.Read(bytes.NewReader(receivedMessageId), binary.BigEndian, &messageId)
		//fmt.Printf("MessageId: %d\n", messageId)

		payload := make([]byte, messageLength-1) // remove message id offset
		_, err = io.ReadFull(connection, payload)
		handleErr(err)
		//fmt.Printf("Payload: %d, size = %d\n", len(payload), size)

		if messageId == expectedMessageId {
			//fmt.Printf("Return for MessageId: %d\n", messageId)
			return payload
		}
	}
}

func createPeerMessage(messageId uint8, payload []byte) []byte {
	// Peer messages consist of a message length prefix (4 bytes), message id (1 byte) and a payload (variable size).
	messageData := make([]byte, 4+1+len(payload))
	binary.BigEndian.PutUint32(messageData[0:4], uint32(1+len(payload)))
	messageData[4] = messageId
	copy(messageData[5:], payload)
	return messageData
}
