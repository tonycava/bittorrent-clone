package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func listenForMessages(conn net.Conn, torrent Torrent) {
	WaitFor(conn, MsgBitfield)

	interestedMessage := []byte{0x00, 0x00, 0x00, 0x01, 0x02}
	_, err := conn.Write(interestedMessage)
	if err != nil {
		fmt.Println("Error sending interested message:", err)
		return
	}

	WaitFor(conn, MsgUnchoke)

	pieceIndex := uint32(0)      // Zero-based piece index
	begin := uint32(0)           // Zero-based byte offset within the piece
	blockLength := uint32(16384) // Length of the block in bytes (e.g., 16 * 1024)

	// Construct and send the request message
	requestMessage := constructRequestMessage(pieceIndex, begin, blockLength)
	_, err = conn.Write(requestMessage)
	if err != nil {
		fmt.Println("Error sending request message:", err)
		return
	}

	data := WaitFor(conn, MsgPiece)
	fmt.Println(string(data))
	//count := 0
	//for byteOffset := 0; byteOffset < int(torrent.Info.Length); byteOffset += BLOCK_SIZE {
	//	payload := make([]byte, 12)
	//	binary.BigEndian.PutUint32(payload[0:4], uint32(pieceId))
	//	binary.BigEndian.PutUint32(payload[4:8], uint32(byteOffset))
	//	binary.BigEndian.PutUint32(payload[8:], BLOCK_SIZE)
	//
	//	_, err = conn.Write(createPeerMessage(MsgRequest, payload))
	//	if err != nil {
	//		return
	//	}
	//	count += 1
	//}

	//
	//combinedBlockToPiece := make([]byte, torrent.Info.Length)
	//fmt.Println(count, "count")
	//for i := 0; i < count; i++ {
	//	data := WaitFor(conn, MsgPiece)
	//	index := binary.BigEndian.Uint32(data[0:4])
	//	fmt.Println(index, "lalalalallaa")
	//	if index != uint32(pieceId) {
	//		panic(fmt.Sprintf("something went wrong [expected: %d -- actual: %d]", pieceId, index))
	//	}
	//	begin := binary.BigEndian.Uint32(data[4:8])
	//	block := data[8:]
	//	copy(combinedBlockToPiece[begin:], block)
	//}

	fmt.Print("Writing to file...")
	err = os.WriteFile(os.Args[3], data, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Piece %d downloaded to %s.\n", pieceIndex, os.Args[3])
}

func constructRequestMessage(pieceIndex, begin, length uint32) []byte {
	// Message format: <length prefix (4 bytes)> <message id (1 byte)> <piece index (4 bytes)> <begin (4 bytes)> <length (4 bytes)>
	messageLength := uint32(17) // 4 bytes (length prefix) + 1 byte (message id) + 4 bytes (piece index) + 4 bytes (
	// begin) + 4 bytes (length)
	message := make([]byte, messageLength)

	// Set the message length prefix (excluding itself)
	binary.BigEndian.PutUint32(message[:4], messageLength-4)

	// Set the message id (6 for request)
	message[4] = 6

	// Set the piece index, begin, and length fields
	binary.BigEndian.PutUint32(message[5:9], pieceIndex)
	binary.BigEndian.PutUint32(message[9:13], begin)
	binary.BigEndian.PutUint32(message[13:17], length)

	return message
}
func WaitFor(connection net.Conn, expectedMessageId uint8) []byte {
	log.Printf("[+] Connected: %s\n", connection.RemoteAddr())
	log.Printf("[!] Waiting for %d\n", expectedMessageId)
	msgLength := make([]byte, 4)
	_, err := connection.Read(msgLength)
	if err != nil {
		fmt.Println("ERROR READING")
		log.Fatal(err)
		return nil
	}
	log.Printf("[+] Received: %x\n", msgLength)
	messageLength := binary.BigEndian.Uint32(msgLength)
	log.Printf("[+] messageLength %v\n", messageLength)
	messageIdByte := make([]byte, 1)
	_, err = connection.Read(messageIdByte)
	if err != nil {
		fmt.Println("ERROR READING2")
		log.Fatal(err)
		return nil
	}
	var messageId uint8
	binary.Read(bytes.NewReader(messageIdByte), binary.BigEndian, &messageId)
	log.Printf("[!] Received: %x - Expected: %x\n", messageId, expectedMessageId)
	if messageId != expectedMessageId {
		fmt.Println("ERROR READING3")
		return nil
	}
	// we already consumed 1 byte for message id
	payload := make([]byte, messageLength-1)
	size, err := io.ReadFull(connection, payload)
	if err != nil {
		fmt.Println("ERROR READING4")
		log.Fatal(err)
		return nil
	}

	log.Printf("Payload: %d, size = %d\n", len(payload), size)
	log.Printf("Message ID: %d\n", messageId)
	log.Printf("Return for MessageId: %d\n", messageId)
	return payload
}
