package main

import (
	"flag"
	"fmt"
	"math/rand"
	"motivar/motivar"
	"os"
	"path"
	"time"

	"github.com/mitchellh/go-homedir"
)

var cfg motivar.Conf

func init() {
	homeDir, err := homedir.Dir()
	motivar.Check(err)

	cfg.Dir = path.Join(homeDir, ".motivar")
	cfg.File = path.Join(cfg.Dir, "motivar.ini")
	cfg.DataDir = path.Join(cfg.Dir, "data")

	err = cfg.Setup()
	motivar.Check(err)

	flag.StringVar(&cfg.Language, "language", "br", "Choose a language to show quotes [br,us]")
	flag.StringVar(&cfg.Language, "l", "br", "Choose a language to show quotes [br,us]")

	flag.Usage = func() {
		var CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		fmt.Fprintf(CommandLine.Output(), motivar.Banner)
		fmt.Fprintf(CommandLine.Output(), "Usage of %s:\n", motivar.Name)
		flag.PrintDefaults()
	}
	flag.Parse()

	cfg.ReadEnv()

	err = motivar.CheckLanguages(cfg.Language)
	motivar.Check(err)
}

func main() {
	pathLang := fmt.Sprintf("/data/%v/", cfg.Language)

	var quotes []motivar.Phrase
	err := motivar.ReadPhrases(pathLang, &quotes)
	motivar.Check(err)

	rand.Seed(time.Now().UnixNano())
	v := rand.Intn(len(quotes)-1) + 1

	q := quotes[v]
	fmt.Printf("%+v %+v\n", q.Quote, q.Author)
}
