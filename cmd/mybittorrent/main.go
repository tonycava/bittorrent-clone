//package main
//
//import (
//	"encoding/json"
//	"fmt"
//	"os"
//)
//
//func main() {
//	command := os.Args[1]
//
//	switch command {
//	case "decode":
//		bencodedValue := os.Args[2]
//		decoded, _ := decodeBencode(bencodedValue, new(int))
//		toDisplay, _ := json.Marshal(decoded)
//		fmt.Println(string(toDisplay))
//		break
//	case "info":
//		torrentFilePath := os.Args[2]
//		torrent := getTorrentFileInfo(torrentFilePath)
//		printInfo(torrent)
//	case "peers":
//		torrentFilePath := os.Args[2]
//		torrent := getTorrentFileInfo(torrentFilePath)
//		peers := getPeers(torrent)
//		for _, peer := range peers {
//			fmt.Printf("%v:%v\n", peer.IP, peer.Port)
//		}
//	case "handshake":
//		serverAddr := os.Args[3]
//		torrentFilePath := os.Args[2]
//		torrent := getTorrentFileInfo(torrentFilePath)
//		connection := getConnections(serverAddr)
//		peerId := makeHandHake(connection, torrent)
//		fmt.Printf("Peer ID: %s\n", peerId)
//	case "download_piece":
//		torrentFilePath := os.Args[4]
//		torrent := getTorrentFileInfo(torrentFilePath)
//		peer := getPeers(torrent)[2]
//
//		serverAddr := peer.IP + ":" + peer.Port
//		fmt.Print(serverAddr)
//
//		conn := getConnections(serverAddr)
//		defer conn.Close()
//
//		fmt.Println("Connected to server")
//
//		makeHandHake(conn, torrent)
//		listenForMessages(conn, torrent)
//	default:
//		fmt.Println("Unknown command: " + command)
//		os.Exit(1)
//	}
//
//}

package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/jackpal/bencode-go"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	//fmt.Println("Logs from your program will appear here!")

	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]
		decoded, err := bencode.Decode(bytes.NewReader([]byte(bencodedValue)))

		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else if command == "info" {
		// read the file
		fileNameOrPath := os.Args[2]
		metaInfo, err := getMetaInfo(fileNameOrPath)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Tracker URL: %v\n", metaInfo.Announce)
		fmt.Printf("Length: %v\n", metaInfo.Info.Length)

		sum := createInfoHash(metaInfo)
		// %x for hex formatting
		fmt.Printf("Info Hash: %x\n", sum)

		//Piece Length: 262144
		fmt.Printf("Piece Length: %v\n", metaInfo.Info.PiecesLen)
		//Piece Hashes:
		// split metaInfo.Info.Pieces for each 20 bytes
		// each 20 bytes is a SHA1 hash

		//fmt.Printf("numberOfPieces %v\n", numberOfPieces)
		pieces := getPieces(Torrent(metaInfo))
		fmt.Printf("Piece Hashes: \n")
		for _, piece := range pieces {
			fmt.Printf("%x\n", piece)
		}

	} else if command == "peers" {
		// read the file
		fileNameOrPath := os.Args[2]
		metaInfo, err := getMetaInfo(fileNameOrPath)
		if err != nil {
			fmt.Println(err)
			return
		}

		printPeers(metaInfo)

	} else if command == "handshake" {
		fileNameOrPath := os.Args[2]
		metaInfo, err := getMetaInfo(fileNameOrPath)
		if err != nil {
			fmt.Println(err)
			return
		}

		peer := os.Args[3]
		connection := createConnection(peer)
		handshake(metaInfo, connection)
		connection.Close()

	} else if command == "download_piece" {
		fileNameOrPath := os.Args[4]
		pieceId, err := strconv.Atoi(os.Args[5])
		handleErr(err)
		metaInfo, err := getMetaInfo(fileNameOrPath)
		handleErr(err)

		peers := getPeers(metaInfo)
		connections := map[string]net.Conn{}
		defer closeAllConnections(connections)
		//for _, peerObj := range peers {
		// since for this problem all peer will have the full file
		peerObj := peers[0]
		peer := fmt.Sprintf("%s:%d", peerObj.IP, peerObj.Port)
		connections[peer] = createConnection(peer)

		preDownload(metaInfo, connections[peer])

		pieces := getPieces(Torrent(metaInfo))

		piece := downloadPiece(pieceId, int(metaInfo.Info.PiecesLen), connections[peer], pieces)
		err = os.WriteFile(os.Args[3], piece, os.ModePerm)
		handleErr(err)

		//}
	} else if command == "download" {
		outPutFileName := os.Args[3]
		fileNameOrPath := os.Args[4]
		metaInfo, err := getMetaInfo(fileNameOrPath)
		handleErr(err)

		peers := getPeers(metaInfo)
		connections := map[string]net.Conn{}
		defer closeAllConnections(connections)

		rand.Shuffle(len(peers), func(i, j int) {
			temp := peers[i]
			peers[i] = peers[j]
			peers[j] = temp
		})

		for _, peerObj := range peers {
			peer := fmt.Sprintf("%s:%d", peerObj.IP, peerObj.Port)

			connections[peer] = createConnection(peer)

			preDownload(metaInfo, connections[peer])

			pieces := getPieces(Torrent(metaInfo))
			fmt.Printf("--------Total Pieces To Download: %d, Total Size: %d--------\n", len(pieces), metaInfo.Info.Length)
			fullFile := make([]byte, metaInfo.Info.Length)
			curr := 0
			for pieceIndex, _ := range pieces {
				fmt.Printf("Start download for piece %d\n", pieceIndex)
				var piece []byte
				// last piece
				if pieceIndex == len(pieces)-1 {
					lastPieceSize := metaInfo.Info.Length - (metaInfo.Info.PiecesLen * int64(pieceIndex))
					fmt.Printf("Last Piece Size [%d - (%d*%d) = %d]\n", metaInfo.Info.Length, metaInfo.Info.PiecesLen, pieceIndex, lastPieceSize)
					piece = downloadPiece(pieceIndex, int(lastPieceSize), connections[peer], pieces)
				} else {
					piece = downloadPiece(pieceIndex, int(metaInfo.Info.PiecesLen), connections[peer], pieces)
				}
				copy(fullFile[curr:], piece)
				curr += len(piece)
			}

			err = os.WriteFile(outPutFileName, fullFile, os.ModePerm)
			handleErr(err)
			return
		}

	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

func closeAllConnections(connections map[string]net.Conn) {
	for _, conn := range connections {
		conn.Close()
	}
}

func getMetaInfo(fileNameOrPath string) (MetaInfo, error) {
	// use std lib to read file's contents as a string
	file, err := os.ReadFile(fileNameOrPath)
	if err != nil {
		return MetaInfo{}, err
	}

	var metaInfo MetaInfo
	if err := bencode.Unmarshal(bytes.NewReader(file), &metaInfo); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return MetaInfo{}, err
	}

	return metaInfo, nil
}

func getPeers(metaInfo MetaInfo) []Peer {
	response, _ := makeGetRequest(metaInfo)

	var trackerResponse TrackerResponse
	bencode.Unmarshal(bytes.NewReader(response), &trackerResponse)
	//fmt.Printf("trackerResponse %v\n", trackerResponse)

	numPeers := len(trackerResponse.Peers) / 6
	peers := make([]Peer, numPeers)
	//fmt.Printf("numPeers %v\n", numPeers)
	for i := 0; i < numPeers; i++ {
		start := i * 6
		end := start + 6
		peer := trackerResponse.Peers[start:end]
		ip := net.IP(peer[0:4])
		port := binary.BigEndian.Uint16([]byte(peer[4:6]))
		peers[i] = Peer{IP: ip, Port: int(port)}
	}
	return peers
}

func printPeers(metaInfo MetaInfo) {
	peers := getPeers(metaInfo)
	for i := 0; i < len(peers); i++ {
		fmt.Printf("%s:%d\n", peers[i].IP, peers[i].Port)
	}
}

func createInfoHash(metaInfo MetaInfo) [20]byte {
	var buffer_ bytes.Buffer
	if err := bencode.Marshal(&buffer_, metaInfo.Info); err != nil {
		fmt.Println("Error marshalling BEncode:", err)
		return [20]byte{}
	}
	sum := sha1.Sum(buffer_.Bytes())
	return sum
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

	myStr :=
		"BitTorrent protocol" + // fixed header
			"00000000" + // reserved bytes
			string(infoHash[:]) +
			"00112233445566778899" // peerId

	// Convert int 19 to byte
	b := make([]byte, 1)
	b[0] = byte(19)

	// Concatenate byte with rest of string
	myBytes := append(b, []byte(myStr)...)

	// issue here is that 19 is encoded as 2 characters instead of 1
	//myStr := "19" + "BitTorrent protocol" + "00000000" + string(infoHash[:]) + "00112233445566778899"
	_, err := conn.Write(myBytes)
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

type MetaInfo struct {
	Announce string `json:"announce" bencode:"announce"`
	Info     Info   `json:"info" bencode:"info"`
}
type Info struct {
	Length    int64  `json:"length" bencode:"length"`
	Name      string `json:"name" bencode:"name"`
	PiecesLen int64  `json:"piece length" bencode:"piece length"`
	Pieces    string `json:"pieces" bencode:"pieces"`
}
type TrackerResponse struct {
	Interval int64  `json:"interval" bencoded:"interval"`
	Peers    string `json:"peers" bencoded:"peers"`
}
type Peer struct {
	IP   net.IP
	Port int
}

func downloadRequestedPiece(pieceId, pieceLength int, conn net.Conn) []byte {
	blockCount := calculateBlockCount(pieceLength)
	combinedBlockToPiece := make([]byte, pieceLength)
	for i := 0; i < blockCount; i++ {
		data := WaitFor(conn, MsgPiece)

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

	WaitFor(conn, MsgBitfield)

	_, err := conn.Write(createPeerMessage(MsgInterested, []byte{}))
	handleErr(err)
	//fmt.Printf("Sent INTERESTED message\n")

	WaitFor(conn, MsgUnchoke)
}

func downloadPiece(pieceId, pieceLength int, conn net.Conn, pieces []string) []byte {
	//fmt.Printf("PieceHash for id: %d --> %x\n", pieceId, pieces[pieceId])
	// say 256 KB
	// for each block
	sendRequestForPiece(pieceId, pieceLength, conn)

	fmt.Printf("For Piece : [%d] of possible Size :[%d] Sent Requests for Blocks of size %d\n", pieceId, pieceLength, BLOCK_SIZE)

	combinedBlockToPiece := downloadRequestedPiece(pieceId, pieceLength, conn)

	ok := verifyPiece(combinedBlockToPiece, pieces, pieceId)

	if !ok {
		panic("unequal pieces")
	}

	return combinedBlockToPiece
}

type RequestPayload struct {
	Index     uint32
	Begin     uint32
	BlockSize uint32
}

func calculateBlockCount(pieceLength int) int {
	var carry int
	if pieceLength%BLOCK_SIZE > 0 {
		carry = 1
	}
	count := pieceLength/BLOCK_SIZE + carry
	return count
}

func sendRequestForPiece(pieceId, pieceLength int, conn net.Conn) {
	count := calculateBlockCount(pieceLength)
	requests := make([]RequestPayload, count)

	for i := range requests {
		begin := uint32(i * BLOCK_SIZE)
		blockSize := uint32(BLOCK_SIZE)
		if uint32(pieceLength)-begin < BLOCK_SIZE {
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
		_, err := conn.Write(createPeerMessage(MsgRequest, payload))
		handleErr(err)
	}
}
