package motivar

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/markbates/pkger"
	"gopkg.in/ini.v1"
)

const (
	// Name cli
	Name = "motivar"

	version = "v0.1"
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

// Phrase phrases struct
type Phrase struct {
	Quote  string `json:"quote"`
	Author string `json:"author"`
}

// Check if errors appers
func Check(e error) {
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
	return errors.New("Language not supported. Use 'br' or 'us'")
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

// ReadPhrases func
func ReadPhrases(path string, p *[]Phrase) error {
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
