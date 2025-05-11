package main

import (
	"crypto/rand"
	"crypto/sha1"
	"database/sql"
	"embed"
	"encoding/binary"
	"errors"
	"github.com/mitchellh/go-homedir"
	"github.com/wvoliveira/motivar/data"
	"log/slog"
	"math"
	_ "modernc.org/sqlite"
	"os"
	"path"
	"time"
)

//go:embed all:migrations
var embedContent embed.FS

type databasePhrase struct {
	URL         string
	ContentHash string
	Author      string
	Phrase      string
	PhraseHash  string
	CreateAt    time.Time
	UpdateAt    time.Time
}

type database struct {
	conn *sql.DB
}

func (d *database) New() {
	homeDir, err := homedir.Dir()
	die(err)

	dbFile := path.Join(homeDir, ".motivar", "data", "database.db")

	d.conn, err = sql.Open("sqlite", dbFile)
	die(err)
	return
}

func (d *database) ConnectAndTest() {
	_, err := d.conn.Exec("SELECT 1;")
	die(err)
}

func (d *database) RunMigrations() {
	dir, err := embedContent.ReadDir("migrations")
	if err != nil {
		slog.Error("Error reading migrations folder: %v", err)
		os.Exit(1)
	}
	for _, file := range dir {
		sqlContent, err := embedContent.ReadFile("migrations/" + file.Name())
		die(err)

		_, err = d.conn.Exec(string(sqlContent))
		die(err)
	}
}

func (d *database) InsertPhrases(phrases []databasePhrase, contentHash string) (err error) {
	tx, err := d.conn.Begin()
	if err != nil {
		return
	}

	defer tx.Rollback()
	stHash, err := tx.Prepare("INSERT INTO hashes (id, url, content_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return
	}

	stPhrase, err := tx.Prepare("INSERT INTO phrases (id, author, phrase, phrase_hash, created_at, updated_at, hash_id) VALUES (?, ?, ?, ?, ?, ?, ?) ON CONFLICT(phrase_hash) DO NOTHING")
	if err != nil {
		return
	}

	defer stHash.Close()
	defer stPhrase.Close()

	now := time.Now()
	hashID := generateHashTimestamp()
	_, err = stHash.Exec(hashID, phrases[0].URL, contentHash, now, now)
	if err != nil {
		return
	}

	for _, item := range phrases {
		phraseID := generateHashTimestamp()

		_, err = stPhrase.Exec(phraseID, item.Author, item.Phrase, item.PhraseHash, now, now, hashID)
		if err != nil {
			return
		}
	}

	err = tx.Commit()
	return
}

func (d *database) GetRandomPhrase() (data.Phrase, error) {
	row := d.conn.QueryRow("SELECT phrase, author FROM phrases ORDER BY RANDOM() LIMIT 1")

	var phrase data.Phrase
	err := row.Scan(&phrase.Phrase, &phrase.Author)
	if err != nil {
		return phrase, err
	}
	return phrase, nil
}

func (d *database) contentHashExists(hash string) (bool, error) {
	row := d.conn.QueryRow("SELECT 1 FROM hashes WHERE content_hash = ? LIMIT 1", hash)

	var temp int
	err := row.Scan(&temp)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func generateHashTimestamp() int64 {
	timestamp := time.Now().UnixNano()

	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}

	input := make([]byte, 16)
	binary.BigEndian.PutUint64(input[0:8], uint64(timestamp))
	copy(input[8:], randomBytes)

	hash := sha1.Sum(input)
	val := binary.BigEndian.Uint64(hash[0:8])
	val = val & uint64(math.MaxInt64)

	return int64(val)
}
