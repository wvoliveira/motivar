//go:generate go run data/data_generate.go

package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/wvoliveira/motivar/data"
	"gopkg.in/ini.v1"
	"log/slog"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/mitchellh/go-homedir"
)

const (
	Name    = "motivar"
	version = "v0.1.0"
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

// Conf directory/file struct
type Conf struct {
	Dir     string
	File    string
	DataDir string
}

type Flags struct {
	Language string
	Debug    bool
}

type FlagsAdd struct {
	Format   string
	URL      string
	Language string
}

var (
	cfg           Conf
	flags         Flags
	flagsAdd      FlagsAdd
	logg          *slog.Logger
	cmdMain       *flag.FlagSet
	cmdAddPhrases *flag.FlagSet
)

func main() {
	logg = NewLogger()

	homeDir, err := homedir.Dir()
	die(err)

	cfg.Dir = path.Join(homeDir, ".motivar")
	cfg.File = path.Join(cfg.Dir, "motivar.ini")
	cfg.DataDir = path.Join(cfg.Dir, "data")

	err = cfg.Setup()
	die(err)

	cmdMain = flag.NewFlagSet("", flag.ExitOnError)
	cmdMain.BoolVar(&flags.Debug, "debug", false, "Enable debug mode")
	cmdMain.StringVar(&flags.Language, "l", "br", "Choose a language to show quotes [br,us]")

	cmdAddPhrases = flag.NewFlagSet("add-phrases", flag.ExitOnError)
	cmdAddPhrases.StringVar(&flagsAdd.Format, "fmt", "csv", "Specify format phrases content [csv,json]")
	cmdAddPhrases.StringVar(&flagsAdd.URL, "url", "", "Specify URL to download from")
	cmdAddPhrases.StringVar(&flagsAdd.Language, "language", "", "The language of phrases [br,us]")

	cmdMain.Usage = func() {
		var cmd = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		_, _ = fmt.Fprint(cmd.Output(), Banner)
		_, _ = fmt.Fprintf(cmd.Output(), "Usage:\n")
		cmdMain.PrintDefaults()

		_, _ = fmt.Fprintf(cmd.Output(), "Subcommand %s:\n", cmdAddPhrases.Name())
		cmdAddPhrases.PrintDefaults()
	}
	cmdAddPhrases.Usage = cmdMain.Usage

	db := initDatabase()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "add-phrases":
			cmdAddPhrases.Parse(os.Args[2:])
			if flagsAdd.Format == "" || flagsAdd.URL == "" || flagsAdd.Language == "" {
				cmdAddPhrases.Usage()
				return
			}
			err = CheckLanguages(flagsAdd.Language)
			die(err)

			err = CheckFormat(flagsAdd.Format)
			die(err)

			err := fetchAndSave(&db, flagsAdd.Format, flagsAdd.URL, flagsAdd.Language)
			if err != nil {
				logg.Error(err.Error())
			}
			return
		}
	}

	cmdMain.Parse(os.Args[1:])
	flags.ReadEnv()

	if flags.Debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	err = CheckLanguages(flags.Language)
	die(err)

	var phrasesData []data.Phrase
	switch flags.Language {
	case "br":
		phrasesData = data.PhrasesBR
	case "us":
		phrasesData = data.PhrasesUS
	default:
		fmt.Println("Unsupported language")
		return
	}

	phrase, err := getRandomPhrase(flags.Language, phrasesData, &db)
	if err != nil {
		fmt.Println(err)
	}

	printPhrase(phrase)
}

func initDatabase() database {
	db := database{}
	db.New()
	db.ConnectAndTest()
	db.RunMigrations()
	return db
}

func die(e error) {
	if e != nil {
		fmt.Sprintln(e.Error())
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

// CheckFormat check format supported
func CheckFormat(format string) error {
	langs := []string{"csv", "json"}

	for _, l := range langs {
		if format == l {
			return nil
		}
	}
	return errors.New("format not supported. Use \"csv\" or \"json\"")
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
func (f Flags) ReadEnv() {
	lang := os.Getenv("MOTIVAR_LANGUAGE")
	if lang != "" {
		err := CheckLanguages(lang)
		if err == nil {
			f.Language = lang
		}
	}
}

func getRandomPhrase(language string, phrases []data.Phrase, db *database) (phrase data.Phrase, err error) {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	randomNumber := rand.Intn(2)
	if randomNumber == 1 {
		phrase, err = db.GetRandomPhrase(language)
		if phrase.Phrase != "" {
			return
		}
	}

	v := rand.Intn(len(phrases)-1) + 1
	return phrases[v], nil
}

func printPhrase(p data.Phrase) {
	fmt.Printf("%+v %+v\n", p.Phrase, p.Author)
}
