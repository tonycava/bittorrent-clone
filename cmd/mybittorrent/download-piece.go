package main

import (
	"crypto/sha1"
)

//func listenForMessages(conn net.Conn, torrent Torrent) {
//	WaitFor(conn, MsgBitfield)
//
//	_, err := conn.Write(createPeerMessage(MsgInterested, []byte{}))
//	if err != nil {
//		fmt.Println("Error sending request message:", err)
//		return
//	}
//	log.Printf("Sent INTERESTED message\n")
//
//	WaitFor(conn, MsgUnchoke)
//
//	pieceId, err := strconv.Atoi(os.Args[5])
//	pieces := getPieces(torrent)
//	handleErr(err)
//
//	data := sendPieceRequest(torrent, pieceId, conn)
//	fmt.Printf("For Piece : [%d] Sent Requests for Blocks\n", pieceId)
//
//	//dataFile := getDataFile(count, pieceId, conn, torrent)
//
//	ok := verifyPiece(data, pieces, pieceId)
//
//	if !ok {
//		panic("unequal pieces")
//	}
//
//	err = os.WriteFile(os.Args[3], data, os.ModePerm)
//	handleErr(err)
//	fmt.Printf("Piece %d downloaded to %s\n", pieceId, os.Args[3])
//}

//func getDataFile(count int, pieceId int, conn net.Conn, torrent Torrent) []byte {
//	combinedBlockToPiece := make([]byte, torrent.Info.PiecesLen)
//	for i := 0; i < count; i++ {
//		data := WaitFor(conn, MsgPiece)
//
//		index := binary.BigEndian.Uint32(data[0:4])
//		if index != uint32(pieceId) {
//			panic(fmt.Sprintf("something went wrong [expected: %d -- actual: %d]", pieceId, index))
//		}
//		begin := binary.BigEndian.Uint32(data[4:8])
//		block := data[8:]
//		copy(combinedBlockToPiece[begin:], block)
//	}
//	return combinedBlockToPiece
//}

//func sendPieceRequest(torrent Torrent, pieceId int, conn net.Conn) []byte {
//	if pieceId > len(torrent.Info.Pieces) {
//		fmt.Println("Invalid piece id")
//	}
//
//	pieceLength := torrent.Info.PiecesLen
//
//	if pieceId == len(torrent.Info.Pieces)-1 {
//		pieceLength = torrent.Info.Length - (int64(pieceId) * torrent.Info.PiecesLen)
//	}
//
//	lastBlockSize := pieceLength % int64(BLOCK_SIZE)
//	numBlocks := (pieceLength - lastBlockSize) / BLOCK_SIZE
//
//	if lastBlockSize > 0 {
//		log.Printf("piece %d has an unaligned block of size %d\n", pieceId, lastBlockSize)
//		numBlocks += 1
//	} else {
//		log.Printf("piece %d has size of %d and is aligned with blocksize of %d\n", pieceId, torrent.Info.PiecesLen, BLOCK_SIZE)
//	}
//
//	combinedBlockToPiece := make([]byte, torrent.Info.PiecesLen)
//
//	for i := 0; i < int(numBlocks); i++ {
//		begin := i * BLOCK_SIZE
//		length := BLOCK_SIZE
//
//		payload := make([]byte, 12)
//
//		if lastBlockSize > 0 && i == int(numBlocks)-1 {
//			log.Printf("reached last block, changing size to %d\n", lastBlockSize)
//			length = int(lastBlockSize)
//		}
//
//		log.Printf("requesting block %d of %d (offset=%d, size=%d)\n", i, numBlocks-1, begin, length)
//
//		binary.BigEndian.PutUint32(payload[0:4], uint32(pieceId))
//		binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
//		binary.BigEndian.PutUint32(payload[8:12], uint32(length))
//
//		_, err := conn.Write(createPeerMessage(MsgRequest, payload))
//		handleErr(err)
//
//		data, err := WaitFor(conn, MsgPiece)
//
//		handleErr(err)
//
//		index := binary.BigEndian.Uint32(data[0:4])
//		if index != uint32(pieceId) {
//			panic(fmt.Sprintf("something went wrong [expected: %d -- actual: %d]", pieceId, index))
//		}
//
//		beginResponse := binary.BigEndian.Uint32(data[4:8])
//		block := data[8:]
//		copy(combinedBlockToPiece[beginResponse:], block)
//
//		log.Printf("received block %d of %d\n", i, numBlocks-1)
//		time.Sleep(1 * time.Second)
//	}
//	return combinedBlockToPiece
//}

func verifyPiece(combinedBlockToPiece []byte, pieces []string, pieceId int) bool {
	sum := sha1.Sum(combinedBlockToPiece)
	return string(sum[:]) == pieces[pieceId]
}
