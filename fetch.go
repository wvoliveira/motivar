package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wvoliveira/motivar/data"
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
//
// TODO:
// - Create content hash after unmarshal in lang object.
// - Format quotes with capitalize first letter and add period at the end.

const BodyMaxLength = 200000

type req struct {
	Body           []byte
	BodyHash       string
	CSVContent     [][]string
	Phrases        []data.Phrase
	DatabaseObject []databasePhrase
}

func fetchAndSave(db *database, kind, url, language string) (err error) {
	if kind == "" || url == "" || language == "" {
		return errors.New("kind, url or language is empty")
	}

	var r req

	switch kind {
	case "csv":
		logg.Info(fmt.Sprintf("Fetching %s", url))
		content, contentHash, err := fetch(url)
		if err != nil {
			return err
		}
		logg.Debug(fmt.Sprintf("Hash of content: %s", contentHash))
		r.Body = content
		r.BodyHash = contentHash

		logg.Info("Validating content format...")
		if !r.IsCSV() {
			return errors.New("invalid CSV format")
		}

		logg.Info("Trying to convert bytes to CSV format...")
		csvLines, err := r.ConvertToCSV()
		if err != nil {
			return err
		}

		r.CSVContent = csvLines

		logg.Info("Checking if hash content exists in database.")
		exists, err := db.contentHashExists(r.BodyHash)
		if err != nil {
			return err
		}

		if exists {
			logg.Warn("This content already exists in the database. Exiting...")
			os.Exit(1)
		}

		logg.Info("Parse CSV content to database object.")
		databasePhrases, err := r.CSVToDatabaseObject(language)
		if err != nil {
			return err
		}
		r.DatabaseObject = databasePhrases

		logg.Info("Inserting in the database...")
		err = db.InsertPhrases(r.DatabaseObject, url, r.BodyHash)
		if err != nil {
			return err
		}

		logg.Info("OK, phrases into database.")
	case "json":
		logg.Info(fmt.Sprintf("Fetching %s", url))
		content, contentHash, err := fetch(url)
		if err != nil {
			return err
		}
		logg.Debug(fmt.Sprintf("Hash of content: %s", contentHash))

		r.Body = content
		r.BodyHash = contentHash

		logg.Info("Validating content format...")
		if !r.IsJSON() {
			return errors.New("invalid JSON format")
		}

		logg.Info("Checking if hash content exists in database.")
		exists, err := db.contentHashExists(r.BodyHash)
		if err != nil {
			return err
		}

		if exists {
			logg.Warn("This content already exists in the database. Bye.")
			os.Exit(1)
		}

		logg.Info("Trying to convert to database Object.")
		databasePhrases, err := r.JSONToDatabaseObject(language)
		if err != nil {
			return err
		}
		r.DatabaseObject = databasePhrases

		logg.Info("Inserting in the database...")
		err = db.InsertPhrases(r.DatabaseObject, url, r.BodyHash)
		if err != nil {
			return err
		}

		logg.Info("OK, phrases into database.")
	default:
		return errors.New("unknown kind")
	}
	return nil
}

func fetch(url string) ([]byte, string, error) {
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
			slog.Error(fmt.Sprintf("the body (%v) exceeded the limit (%v)", length, BodyMaxLength))
			os.Exit(1)
		}
	}
	contentHash := generateHash(string(body))

	// TODO:
	// - Check if this hash already in the database.
	// - Check if phrase hash exists in the database.

	return body, contentHash, nil
}

func (r req) IsCSV() bool {
	reader := csv.NewReader(bytes.NewReader(r.Body))
	_, err := reader.ReadAll()
	return err == nil
}

func (r req) IsJSON() bool {
	var temp []map[string]interface{}
	return json.Unmarshal(r.Body, &temp) == nil
}

func (r req) ConvertToCSV() (csvLines [][]string, err error) {
	reader := csv.NewReader(bytes.NewReader(r.Body))
	csvLines, err = reader.ReadAll()
	if err != nil {
		return [][]string{}, err
	}
	return csvLines, nil
}

func convertFromJSON(body []byte, object *[]data.Phrase) (err error) {
	return json.Unmarshal(body, &object)
}

func (r req) CSVToDatabaseObject(language string) (dbPhrases []databasePhrase, err error) {
	if language == "" {
		return dbPhrases, errors.New("url or language is empty")
	}

	// Here we jump the first line to not process the headers.
	// Maybe a flag or describe in help message to warn the final user?
	for _, line := range r.CSVContent[1:] {
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
			ContentHash: r.BodyHash,
			Author:      line[0],
			Phrase:      line[1],
			PhraseHash:  phraseHash,
			Language:    language,
		}
		dbPhrases = append(dbPhrases, phrase)
	}
	return dbPhrases, nil
}

func (r req) JSONToDatabaseObject(language string) (dbPhrases []databasePhrase, err error) {
	if language == "" {
		return dbPhrases, errors.New("url or language is empty")
	}

	err = json.Unmarshal(r.Body, &dbPhrases)
	if err != nil {
		return []databasePhrase{}, err
	}

	// Here we jump the first line to not process the headers.
	// Maybe a flag or describe in help message to warn the final user?
	for _, item := range dbPhrases {
		// Don't input in database if author or phrase is empty.
		if item.Phrase == "" || item.Author == "" {
			slog.Info(fmt.Sprintf("Author or phrase is empty: author=\"%s\" phrase=\"%s\"", item.Author, item.Phrase))
			continue
		}

		phraseHash := generateHash(item.Phrase)
		phrase := databasePhrase{
			ContentHash: r.BodyHash,
			Author:      item.Author,
			Phrase:      item.Phrase,
			PhraseHash:  phraseHash,
			Language:    language,
		}
		dbPhrases = append(dbPhrases, phrase)
	}

	return dbPhrases, nil
}

func generateHash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
