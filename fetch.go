package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

// Get phrases from internet.
// Examples:
// - CSV: https://gist.githubusercontent.com/JakubPetriska/060958fd744ca34f099e947cd080b540/raw/963b5a9355f04741239407320ac973a6096cd7b6/quotes.csv
// - JSON: https://raw.githubusercontent.com/AtaGowani/daily-motivation/refs/heads/master/src/data/quotes.json
//
// Supported file types: csv and json.
// Use files in the samples folder to develop some fetch feature.

const BodyMaxLength = 200000

func fetchAndSave(db *database, kind string, url string) (err error) {
	if kind == "" || url == "" {
		return errors.New("kind or url is empty")
	}
	if kind != "csv" && kind != "json" {
		return errors.New("kind must be csv or json")
	}

	if kind == "csv" {
		logg.Info(fmt.Sprintf("Fetching %s", url))
		content, contentHash, err := csvFetch(url)
		if err != nil {
			return err
		}
		logg.Debug(fmt.Sprintf("Hash of content: %s", contentHash))

		logg.Info("Trying to convert bytes to CSV format...")
		csvLines, err := convertToCSV(content)
		if err != nil {
			return err
		}

		logg.Info("Checking if hash content exists in database.")
		exists, err := db.contentHashExists(contentHash)
		if err != nil {
			return err
		}

		if exists {
			logg.Warn("This content already exists in the database. Exiting...")
			os.Exit(1)
		}

		logg.Info("Parse CSV content to language object.")
		phrases, err := csvParse(csvLines, url, contentHash)
		if err != nil {
			return err
		}

		logg.Info("Inserting in the database...")
		err = db.InsertPhrases(phrases, contentHash)
		if err != nil {
			return err
		}

		logg.Info("OK, phrases into database.")
	}
	return nil
}

func csvFetch(url string) ([]byte, string, error) {
	resp, err := http.Get(url)
	die(err)

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		slog.Error("Fetching %s: %d", url, resp.StatusCode)
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
			slog.Error("the body (%v) exceeded the limit (%v)", length, BodyMaxLength)
			os.Exit(1)
		}
	}
	contentHash := generateHash(string(body))

	// TODO:
	// - Check if this hash already in the database.
	// - Check if phrase hash exists in the database.

	return body, contentHash, nil
}

func convertToCSV(body []byte) (csvLines [][]string, err error) {
	reader := csv.NewReader(bytes.NewReader(body))
	csvLines, err = reader.ReadAll()
	if err != nil {
		return [][]string{}, err
	}
	return csvLines, nil
}

func csvParse(csvLines [][]string, url, hash string) (dp []databasePhrase, err error) {
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
			slog.Debug(fmt.Sprintf("Author or phrase is empty: author=\"%s\" phrase=\"%s\"", line[0], line[1]))
			continue
		}

		phraseHash := generateHash(line[1])

		phrase := databasePhrase{
			URL:         url,
			ContentHash: hash,
			Author:      line[0],
			Phrase:      line[1],
			PhraseHash:  phraseHash,
		}
		dp = append(dp, phrase)
	}
	return dp, nil
}

func generateHash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
