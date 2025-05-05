//go:generate go run data/data_generate.go

package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/wvoliveira/motivar/data"
	"gopkg.in/ini.v1"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/mitchellh/go-homedir"
)

const (
	Name    = "motivar"
	version = "v0.0.2"
)

var (
	// Banner to show when run flags
	Banner = fmt.Sprintf(`
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

// Conf director/file struct
type Conf struct {
	Dir      string
	File     string
	DataDir  string
	Language string
}

var cfg Conf

// Init use uppercase to not run automatically in the tests.
func Init() {
	homeDir, err := homedir.Dir()
	die(err)

	cfg.Dir = path.Join(homeDir, ".motivar")
	cfg.File = path.Join(cfg.Dir, "motivar.ini")
	cfg.DataDir = path.Join(cfg.Dir, "data")

	err = cfg.Setup()
	die(err)

	flag.StringVar(&cfg.Language, "language", "br", "Choose a language to show quotes [br,us]")
	flag.StringVar(&cfg.Language, "l", "br", "Choose a language to show quotes [br,us]")

	flag.Usage = func() {
		var CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		_, _ = fmt.Fprint(CommandLine.Output(), Banner)
		_, _ = fmt.Fprintf(CommandLine.Output(), "Usage of %s:\n", Name)
		flag.PrintDefaults()
	}
	flag.Parse()

	cfg.ReadEnv()

	err = CheckLanguages(cfg.Language)
	die(err)
}

func die(e error) {
	if e != nil {
		fmt.Printf("error: %+v\n", e)
		os.Exit(1)
	}
}

// CheckLanguages check languages supported
func CheckLanguages(lang string) error {
	langs := []string{"br", "us"}

	for _, l := range langs {
		if lang == l {
			return nil
		}
	}
	return errors.New("language not supported. Use 'br' or 'us'")
}

// Setup func
func (c Conf) Setup() error {
	// check and create conf dir
	if _, err := os.Stat(c.Dir); os.IsNotExist(err) {
		err := os.Mkdir(c.Dir, 0764)
		if err != nil {
			return err
		}
	}

	// check and create conf file
	if _, err := os.Stat(c.File); os.IsNotExist(err) {
		_, err := os.Create(c.File)
		if err != nil {
			return err
		}

		err = c.MakeConf()
		if err != nil {
			return err
		}
	}

	// create data dir
	if _, err := os.Stat(c.DataDir); os.IsNotExist(err) {
		err := os.Mkdir(c.DataDir, 0764)
		if err != nil {
			return err
		}
	}
	return nil
}

// MakeConf func
func (c Conf) MakeConf() error {
	cfg, err := ini.Load(c.File)
	if err != nil {
		return err
	}

	cfg.Section("").Key("language").SetValue("br")
	err = cfg.SaveTo(c.File)
	if err != nil {
		return err
	}

	return nil
}

// ReadEnv read environment variables
func (c Conf) ReadEnv() {
	lang := os.Getenv("MOTIVAR_LANGUAGE")
	if lang != "" {
		err := CheckLanguages(lang)
		if err == nil {
			c.Language = lang
		}
	}
}

func getRandomPhrase(phrases []data.Phrase) (phrase data.Phrase) {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	v := rand.Intn(len(phrases)-1) + 1
	return phrases[v]
}

func printPhrase(p data.Phrase) {
	fmt.Printf("%+v %+v\n", p.Quote, p.Author)
}

func main() {
	//Init()
	//
	//if cfg.Language == "br" {
	//	printPhrase(getRandomPhrase(data.PhrasesBR))
	//	return
	//}
	//
	//printPhrase(getRandomPhrase(data.PhrasesUS))

	fetchCSV("http://localhost:8000/quotes.csv")
}
