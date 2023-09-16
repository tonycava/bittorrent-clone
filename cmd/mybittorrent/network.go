//package main
//
//import (
//	"encoding/binary"
//	"encoding/hex"
//	"fmt"
//	"io"
//	"log"
//	"net"
//	"os"
//)
//
//func makeHandHake(conn net.Conn, torrent Torrent) string {
//	handshake := make([]byte, 0)
//	handshake = append(handshake, 19)
//	handshake = append(handshake, []byte("BitTorrent protocol")...)
//	handshake = append(handshake, make([]byte, 8)...)
//	handshake = append(handshake, torrent.getInfoHash()...)
//	handshake = append(handshake, []byte("00112233445566778899")...)
//
//	_, err := conn.Write(handshake)
//	if err != nil {
//		fmt.Println("Error sending data:", err)
//		os.Exit(1)
//	}
//
//	response := make([]byte, len(handshake))
//	_, err = conn.Read(response)
//	if err != nil {
//		fmt.Println("Error receiving data:", err)
//		os.Exit(1)
//	}
//
//	return hex.EncodeToString(response[len(response)-20:])
//}
//
//func getConnections(peer string) net.Conn {
//	conn, err := net.Dial("tcp", peer)
//	handleErr(err)
//	return conn
//}
//
//func WaitFor(connection net.Conn, expectedMessageId uint8) ([]byte, error) {
//	log.Printf("waiting for message %v\n", expectedMessageId)
//	var messageLength uint32
//	var messageID byte
//
//	fmt.Println("Waiting for message length")
//	if err := binary.Read(connection, binary.BigEndian, &messageLength); err != nil {
//		return nil, err
//	}
//	fmt.Println("Message Length: ", messageLength)
//	if err := binary.Read(connection, binary.BigEndian, &messageID); err != nil {
//		return nil, err
//	}
//	if messageID != expectedMessageId {
//		return nil, fmt.Errorf("unexpected message ID: (actual=%d, expected=%d)", messageID, expectedMessageId)
//	}
//	log.Printf("received message %d\n", messageID)
//
//	if messageLength > 1 {
//		log.Printf("message %d has attached payload of size %d\n", expectedMessageId, messageLength-1)
//		payload := make([]byte, messageLength-1)
//		if _, err := io.ReadAtLeast(connection, payload, len(payload)); err != nil {
//			return nil, fmt.Errorf("error while reading payload: %s", err.Error())
//		}
//		return payload, nil
//	}
//	return nil, nil
//}
//
//func createPeerMessage(messageId uint8, payload []byte) []byte {
//	// Peer messages consist of a message length prefix (4 bytes), message id (1 byte) and a payload (variable size).
//	messageData := make([]byte, 4+1+len(payload))
//	binary.BigEndian.PutUint32(messageData[0:4], uint32(1+len(payload)))
//	messageData[4] = messageId
//	copy(messageData[5:], payload)
//	return messageData
//}

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

const INTERESTED = 2
const BITFIELD = 5
const UNCHOKE = 1
const REQUEST = 6
const PIECE = 7
const BlockSize = 16 * 1024

func downloadPiece(pieceId, pieceLength int, conn net.Conn, pieces []string) []byte {
	//fmt.Printf("PieceHash for id: %d --> %x\n", pieceId, pieces[pieceId])
	// say 256 KB
	// for each block
	sendRequestForPiece(pieceId, pieceLength, conn)

	fmt.Printf("For Piece : [%d] of possible Size :[%d] Sent Requests for Blocks of size %d\n", pieceId, pieceLength, BlockSize)

	combinedBlockToPiece := downloadRequestedPiece(pieceId, pieceLength, conn)

	ok := verifyPiece(combinedBlockToPiece, pieces, pieceId)

	if !ok {
		panic("unequal pieces")
	}

	return combinedBlockToPiece
}

func sendRequestForPiece(pieceId, pieceLength int, conn net.Conn) {
	count := calculateBlockCount(pieceLength)
	requests := make([]RequestPayload, count)

	for i := range requests {
		begin := uint32(i * BlockSize)
		blockSize := uint32(BlockSize)
		if uint32(pieceLength)-begin < BlockSize {
			blockSize = uint32(pieceLength) - begin
		}
		requests[i] = RequestPayload{
			Index:     uint32(pieceId),
			Begin:     begin,
			BlockSize: blockSize,
		}
	}

	for _, request := range requests {
		payload := make([]byte, 12)
		binary.BigEndian.PutUint32(payload[0:4], request.Index)    // index
		binary.BigEndian.PutUint32(payload[4:8], request.Begin)    // begin
		binary.BigEndian.PutUint32(payload[8:], request.BlockSize) // block size
		_, err := conn.Write(createPeerMessage(REQUEST, payload))
		handleErr(err)
	}
}

func calculateBlockCount(pieceLength int) int {
	var carry int
	if pieceLength%BlockSize > 0 {
		carry = 1
	}
	count := pieceLength/BlockSize + carry
	return count
}

func downloadRequestedPiece(pieceId, pieceLength int, conn net.Conn) []byte {
	blockCount := calculateBlockCount(pieceLength)
	combinedBlockToPiece := make([]byte, pieceLength)
	for i := 0; i < blockCount; i++ {
		data := waitFor(conn, PIECE)

		index := binary.BigEndian.Uint32(data[0:4])
		if index != uint32(pieceId) {
			panic(fmt.Sprintf("something went wrong [expected: %d -- actual: %d]", pieceId, index))
		}
		begin := binary.BigEndian.Uint32(data[4:8])
		block := data[8:]
		copy(combinedBlockToPiece[begin:], block)
	}
	return combinedBlockToPiece
}

func preDownload(metaInfo MetaInfo, conn net.Conn) {
	handshake(metaInfo, conn)

	waitFor(conn, BITFIELD)

	_, err := conn.Write(createPeerMessage(INTERESTED, []byte{}))
	handleErr(err)
	//fmt.Printf("Sent INTERESTED message\n")

	waitFor(conn, UNCHOKE)
}

func waitFor(connection net.Conn, expectedMessageId uint8) []byte {
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

func makeGetRequest(metaInfo MetaInfo) ([]byte, error) {
	baseUrl := metaInfo.Announce
	params := url.Values{}
	infoHash := createInfoHash(metaInfo)
	// took help from code examples for - string(infoHash[:])
	params.Add("info_hash", string(infoHash[:]))
	params.Add("peer_id", "00112233445566778899")
	params.Add("port", "6881")
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("left", strconv.Itoa(int(metaInfo.Info.Length)))
	params.Add("compact", "1")

	// Escape the params
	escapedParams := params.Encode()

	// Construct full URL
	URI := fmt.Sprintf("%s?%s", baseUrl, escapedParams)
	fmt.Printf("URI %v\n", URI)

	resp, err := http.DefaultClient.Get(URI)

	//fmt.Printf("StatusCode = %v\n", resp.Status)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return body, nil
}

func createConnection(peer string) net.Conn {
	// Connect to a TCP server
	conn, err := net.Dial("tcp", peer)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	return conn
}

func handshake(metaInfo MetaInfo, conn net.Conn) {
	infoHash := createInfoHash(metaInfo)
	//messageHolder := make([]byte, 1+19+8+20+20)
	//messageHolder[0] = 19
	//copy(messageHolder[1:1+19], "BitTorrent protocol")
	//copy(messageHolder[20:20+8], make([]byte, 8))
	//copy(messageHolder[28:28+20], infoHash[:])
	//copy(messageHolder[48:48+20], "00112233445566778899")

	handshakeMessage :=
		"" +
			string(byte(19)) + // length of string `BitTorrent protocol`
			"BitTorrent protocol" + // fixed header
			"00000000" + // reserved bytes
			string(infoHash[:]) +
			"00112233445566778899" // peerId

	// issue here is that 19 is encoded as 2 characters instead of 1
	//handshakeMessage := "19" + "BitTorrent protocol" + "00000000" + string(infoHash[:]) + "00112233445566778899"
	_, err := conn.Write([]byte(handshakeMessage))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	fmt.Println("Handshake Message Sent, waiting for handshake message myself")

	// Receive response
	buf := make([]byte, 1+19+8+20+20)
	_, err = conn.Read(buf)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	fmt.Printf("Peer ID: %x\n", string(buf[48:]))
}

func closeAllConnections(connections map[string]net.Conn) {
	for _, conn := range connections {
		conn.Close()
	}
}

type RequestPayload struct {
	Index     uint32
	Begin     uint32
	BlockSize uint32
}
