package main

import (
	"anki/cards"
	"flag"
	"fmt"
	_ "image/jpeg"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func init() {
	logger = logrus.New()
	logger.SetReportCaller(true)
	logger.SetFormatter(&logrus.JSONFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			s := strings.Split(f.Function, ".")
			funcName := s[len(s)-1]
			return funcName, fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
		},
	})
}

func main() {
	wordPtr := flag.String("word", "", "word to create card for")
	collectionsDirPt := flag.String("collectionsDir", "", "location of anki collections dir. (e.g. Users/username/Library/Application Support/Anki2/collection.media)")
	outputFilePt := flag.String("outputFile", "", "csv output file name")
	numberDefns := flag.Int("numberDefns", 1, "number of cards to create (default: 1 creats card for most common defn)")

	flag.Parse()
	word := *wordPtr
	if word == "" {
		logger.Fatal("word is required")
	}
	collectionsDir := *collectionsDirPt
	if collectionsDir == "" {
		logger.Fatal("collectionsDir is required")
	}
	if _, err := os.Stat(collectionsDir); os.IsExist(err) {
		logger.Fatalf("directory does not exist: %s", collectionsDir)
	}
	outputFile := *outputFilePt
	if outputFile == "" {
		logger.Fatal("outputFile is required")
	}
	logger.Printf("args: \n\tword: %s\n\toutPutFile: %s\n\tcollectionsDir: %s\n numberDefns: %d\n", word, outputFile, collectionsDir, *numberDefns)

	r, err := cards.GetData(word)
	if err != nil {
		logger.Fatal(err)
	}

	cs, err := cards.MakeCards(r, collectionsDir, *numberDefns)
	if err != nil {
		logger.Fatal(err)
	}

	err = cards.MakeCsv(cs, outputFile)
	if err != nil {
		logger.Fatal(err)
	}
}
