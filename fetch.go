package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
)

// Get phrases from internet.
// Examples:
// - CSV: https://gist.githubusercontent.com/JakubPetriska/060958fd744ca34f099e947cd080b540/raw/963b5a9355f04741239407320ac973a6096cd7b6/quotes.csv
// - JSON: https://raw.githubusercontent.com/AtaGowani/daily-motivation/refs/heads/master/src/data/quotes.json
//
// Supported file types: csv and json.
// This is more slowly because golang will read the files in realtime.

const BodyMaxLength = 200000

func fetchCSV(url string) ([]byte, error) {
	// TODO: Move this to main function.
	db := database{}
	db.New()
	db.ConnectAndTest()
	db.CreateTable()

	resp, err := http.Get(url)
	die(err)

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("Error fetching %s: %d", url, resp.StatusCode)
		os.Exit(1)
	}

	// Read 100 bytes for each loop to prevent buffer overflow.
	// Max 200000 bytes == 200 KB. It's enough to process 2k lines with csv format.
	buf := make([]byte, 100)
	var (
		body   []byte
		n      int
		length int
	)

	for err == nil {
		n, err = resp.Body.Read(buf)
		body = append(body, buf[:n]...)
		length += n

		if length > BodyMaxLength {
			die(fmt.Errorf("the body (%v) exceeded the limit (%v)", length, BodyMaxLength))
		}
	}

	log.Printf("Size is %d\n", length)
	hash := generateHash(string(body))
	log.Println(hash)

	// TODO: Check if this hash already in the database.

	reader := csv.NewReader(bytes.NewReader(body))
	csvLines, err := reader.ReadAll()
	die(err)

	log.Println("OK, body content is a valid CSV format.")
	var dp []databasePhrase

	// Here we jump the first line to not process the headers.
	// Maybe a flag or describe in help message to warn the final user?
	for _, line := range csvLines[1:] {
		// In the CSV content, each line needs to have length of two.
		// First is author and second is phrase.
		if len(line) != 2 {
			continue
		}

		// Don't input in database if author or phrase is empty.
		if line[0] == "" || line[1] == "" {
			log.Printf("Author or phrase is empty: author=\"%s\" phrase=\"%s\"", line[0], line[1])
			continue
		}

		phrase := databasePhrase{
			URL:     url,
			URLHash: hash,
			Author:  line[0],
			Phrase:  line[1],
		}
		dp = append(dp, phrase)
	}

	// TODO: move out of here.
	log.Println("Inserting in database...")
	db.InsertPhrases(dp)

	return []byte{}, nil
}

func generateHash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
