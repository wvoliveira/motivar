package main

import (
	"crypto/rand"
	"crypto/sha1"
	"database/sql"
	"encoding/binary"
	"github.com/mitchellh/go-homedir"
	"log"
	"math"
	_ "modernc.org/sqlite"
	"path"
	"time"
)

type databasePhrase struct {
	URL      string
	URLHash  string
	Author   string
	Phrase   string
	CreateAt time.Time
}

type database struct {
	conn *sql.DB
}

func (d *database) New() {
	homeDir, err := homedir.Dir()
	die(err)

	dbFile := path.Join(homeDir, ".motivar", "data", "database.db")
	log.Println(dbFile)

	d.conn, err = sql.Open("sqlite", dbFile)
	die(err)
	return
}

func (d *database) ConnectAndTest() {
	_, err := d.conn.Exec("SELECT 1;")
	die(err)
}

func (d *database) CreateTable() {
	_, err := d.conn.Exec(`
		CREATE TABLE IF NOT EXISTS phrases (
			id INTEGER PRIMARY KEY,
			url TEXT NOT NULL,
			url_hash TEXT NOT NULL,
			author TEXT NOT NULL,
			phrase TEXT NOT NULL,
			create_at DATETIME
		)
	`)
	die(err)
}

func (d *database) InsertPhrases(phrases []databasePhrase) {
	tx, err := d.conn.Begin()
	die(err)

	defer tx.Rollback()
	//_, err := d.conn.Exec("INSERT INTO phrases (id, url, url_hash, author, phrase) VALUES (?, ?, ?, ?, ?)", generateHashTimestamp(), p.URL, p.URLHash, p.Author, p.Phrase)
	stmt, err := tx.Prepare("INSERT INTO phrases (id, url, url_hash, author, phrase, create_at) VALUES (?, ?, ?, ?, ?, ?)")
	die(err)

	defer stmt.Close()

	now := time.Now()
	for _, item := range phrases {
		hash := generateHashTimestamp()
		_, err = stmt.Exec(hash, item.URL, item.URLHash, item.Author, item.Phrase, now)
		die(err)
	}

	err = tx.Commit()
	die(err)

	log.Println("Batch insert successful")
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
