package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/markbates/pkger"
	"github.com/mitchellh/go-homedir"

	"gopkg.in/ini.v1"
)

var (
	homeDir string

	cfgDir     string
	cfgFile    string
	cfgDataDir string

	quotesBR []phrase
	quotesUS []phrase
)

type phrase struct {
	Quote  string `json:"quote"`
	Author string `json:"author"`
}

func init() {
	homeDir, err := homedir.Dir()
	check(err)

	cfgDir = path.Join(homeDir, ".motivar")
	cfgFile = path.Join(cfgDir, "motivar.ini")
	cfgDataDir = path.Join(cfgDir, "data")

	err = setup()
	check(err)

	err = readPhrases("/data/br/")
	check(err)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	v := rand.Intn(len(quotesBR)-1) + 1

	q := quotesBR[v]
	fmt.Printf("%+v %+v\n", q.Quote, q.Author)
}

func check(e error) {
	if e != nil {
		fmt.Printf("error: %+v", e)
		os.Exit(1)
	}
}

func setup() error {
	// check and create conf dir
	if _, err := os.Stat(cfgDir); os.IsNotExist(err) {
		err := os.Mkdir(cfgDir, 0764)
		if err != nil {
			return err
		}
	}

	// check and create conf file
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		_, err := os.Create(cfgFile)
		if err != nil {
			return err
		}

		err = makeCfg()
		if err != nil {
			return err
		}
	}

	// create data dir
	if _, err := os.Stat(cfgDataDir); os.IsNotExist(err) {
		err := os.Mkdir(cfgDataDir, 0764)
		if err != nil {
			return err
		}
	}
	return nil
}

func makeCfg() error {
	cfg, err := ini.Load(cfgFile)
	if err != nil {
		return err
	}

	cfg.Section("").Key("languages").SetValue("br,us")
	err = cfg.SaveTo(cfgFile)
	if err != nil {
		return err
	}

	return nil
}

func readPhrases(path string) error {
	pkger.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			f, err := pkger.Open(path)
			defer f.Close()

			content, err := ioutil.ReadAll(f)
			if err != nil {
				return err
			}

			err = json.Unmarshal(content, &quotesBR)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return nil
}
