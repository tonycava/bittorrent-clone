package main

import (
	"bytes"
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

func getConnections(serverAddr string) net.Conn {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		os.Exit(1)
	}
	return conn
}

func WaitFor(connection net.Conn, expectedMessageId uint8) []byte {
	log.Printf("[+] Connected: %s\n", connection.RemoteAddr())
	log.Printf("[!] Waiting for %d\n", expectedMessageId)
	msgLength := make([]byte, 4)
	_, err := connection.Read(msgLength)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	log.Printf("[+] Received: %x\n", msgLength)
	messageLength := binary.BigEndian.Uint32(msgLength)
	log.Printf("[+] messageLength %v\n", messageLength)
	messageIdByte := make([]byte, 1)
	_, err = connection.Read(messageIdByte)
	if err != nil {
		log.Fatal(err, "fff")
		return nil
	}
	var messageId uint8
	binary.Read(bytes.NewReader(messageIdByte), binary.BigEndian, &messageId)
	log.Printf("[!] Received: %x - Expected: %x\n", messageId, expectedMessageId)
	if messageId != expectedMessageId {
		return nil
	}
	// we already consumed 1 byte for message id
	payload := make([]byte, messageLength-1)
	size, _ := io.ReadFull(connection, payload)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	log.Printf("Payload: %d, size = %d\n", len(payload), size)
	log.Printf("Message ID: %d\n", messageId)
	log.Printf("Return for MessageId: %d\n", messageId)
	return payload
}

func createPeerMessage(messageId uint8, payload []byte) []byte {
	// Peer messages consist of a message length prefix (4 bytes), message id (1 byte) and a payload (variable size).
	messageData := make([]byte, 4+1+len(payload))
	binary.BigEndian.PutUint32(messageData[0:4], uint32(1+len(payload)))
	messageData[4] = messageId
	copy(messageData[5:], payload)
	return messageData
}
