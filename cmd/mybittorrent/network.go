package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
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

func WaitFor(connection net.Conn, expectedMessageId uint8) ([]byte, error) {
	log.Printf("waiting for message %v\n", expectedMessageId)
	var messageLength uint32
	var messageID byte

	if err := binary.Read(connection, binary.BigEndian, &messageLength); err != nil {
		return nil, err
	}
	if err := binary.Read(connection, binary.BigEndian, &messageID); err != nil {
		return nil, err
	}
	if messageID != expectedMessageId {
		return nil, fmt.Errorf("unexpected message ID: (actual=%d, expected=%d)", messageID, expectedMessageId)
	}
	log.Printf("received message %d\n", messageID)

	if messageLength > 1 {
		log.Printf("message %d has attached payload of size %d\n", expectedMessageId, messageLength-1)
		payload := make([]byte, messageLength-1)
		if _, err := io.ReadAtLeast(connection, payload, len(payload)); err != nil {
			return nil, fmt.Errorf("error while reading payload: %s", err.Error())
		}
		return payload, nil
	}
	return nil, nil
}

func createPeerMessage(messageId uint8, payload []byte) []byte {
	// Peer messages consist of a message length prefix (4 bytes), message id (1 byte) and a payload (variable size).
	messageData := make([]byte, 4+1+len(payload))
	binary.BigEndian.PutUint32(messageData[0:4], uint32(1+len(payload)))
	messageData[4] = messageId
	copy(messageData[5:], payload)
	return messageData
}
