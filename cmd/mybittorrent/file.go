package main

import (
	"bytes"
	"fmt"
	"github.com/jackpal/bencode-go"
	"io/ioutil"
	"log"
	"os"
)

func getTorrentFileInfo(path string) Torrent {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return Torrent{}
	}

	var torrent Torrent
	if err := bencode.Unmarshal(bytes.NewReader(content), &torrent); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return Torrent{}
	}

	return torrent
}

func writeFile(path string, data []byte) {
	err := os.WriteFile(path, data, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}
