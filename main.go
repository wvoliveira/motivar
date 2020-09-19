package main

import (
	"encoding/json"
	"flag"
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
	name    string = "motivar"
	version string = "v0.1"

	homeDir string

	cfgDir     string
	cfgFile    string
	cfgDataDir string

	flagLanguage string

	banner string = fmt.Sprintf(`
              ._ o o
              \_´-)|_
           ,""       \
         ,"  ## |   ಠ ಠ. 
       ," ##   ,-\__    ´.
     ,"       /     ´--._;)
   ,"     ## / Motivar %v
 ,"   ##    /

 `, version)
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

	flag.StringVar(&flagLanguage, "language", "br", "Choose a language to show quotes [br,us]")
	flag.StringVar(&flagLanguage, "l", "br", "Choose a language to show quotes [br,us]")

	flag.Usage = func() {
		var CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		fmt.Fprintf(CommandLine.Output(), banner)
		fmt.Fprintf(CommandLine.Output(), "Usage of %s:\n", name)
		flag.PrintDefaults()
	}

	flag.Parse()

	readEnv()
}

func main() {
	pathLang := fmt.Sprintf("/data/%v/", flagLanguage)

	var quotes []phrase
	err := readPhrases(pathLang, &quotes)
	check(err)

	rand.Seed(time.Now().UnixNano())
	v := rand.Intn(len(quotes)-1) + 1

	q := quotes[v]
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

	cfg.Section("").Key("language").SetValue("br")
	err = cfg.SaveTo(cfgFile)
	if err != nil {
		return err
	}

	return nil
}

func readPhrases(path string, p *[]phrase) error {
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

			err = json.Unmarshal(content, &p)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return nil
}

func readEnv() {
	language := os.Getenv("MOTIVAR_LANGUAGE")
	if language != "" {
		if language == "br" || language == "us" {
			flagLanguage = language
		}
	}
}
