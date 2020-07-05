package main

import (
    "encoding/csv"
    "encoding/json"
    "errors"
    "flag"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "regexp"
    "time"
)

type Response struct {
    ResultCardHeaderProps    `json:"resultCardHeaderProps"`
    SdDictionaryResultsProps `json:"sdDictionaryResultsProps"`
}

type ResultCardHeaderProps struct {
    HeadwordAndQuickdefsProps struct {
        UILang   string `json:"uiLang"`
        IsMobile bool   `json:"isMobile"`
        Headword struct {
            DisplayText     string `json:"displayText"`
            TextToPronounce string `json:"textToPronounce"`
            AudioURL        string `json:"audioUrl"`
            Pronunciations  []struct {
                ID        int    `json:"id"`
                Ipa       string `json:"ipa"`
                Abc       string `json:"abc"`
                Spa       string `json:"spa"`
                Region    string `json:"region"`
                HasVideo  int    `json:"hasVideo"`
                SpeakerID int    `json:"speakerId"`
                Version   int    `json:"version"`
            } `json:"pronunciations"`
            WordLang string `json:"wordLang"`
            Type     string `json:"type"`
        } `json:"headword"`
        Quickdef1 struct {
            DisplayText     string `json:"displayText"`
            TextToPronounce string `json:"textToPronounce"`
            AudioURL        string `json:"audioUrl"`
            Pronunciations  []struct {
                ID        int    `json:"id"`
                Ipa       string `json:"ipa"`
                Abc       string `json:"abc"`
                Spa       string `json:"spa"`
                Region    string `json:"region"`
                HasVideo  int    `json:"hasVideo"`
                SpeakerID int    `json:"speakerId"`
                Version   int    `json:"version"`
                Source    string `json:"source"`
            } `json:"pronunciations"`
            WordLang string `json:"wordLang"`
            Type     string `json:"type"`
        } `json:"quickdef1"`
        Quickdef2 struct {
            DisplayText     string `json:"displayText"`
            TextToPronounce string `json:"textToPronounce"`
            AudioURL        string `json:"audioUrl"`
            Pronunciations  []struct {
                ID        int    `json:"id"`
                Ipa       string `json:"ipa"`
                Abc       string `json:"abc"`
                Spa       string `json:"spa"`
                Region    string `json:"region"`
                HasVideo  int    `json:"hasVideo"`
                SpeakerID int    `json:"speakerId"`
                Version   int    `json:"version"`
                Source    string `json:"source"`
            } `json:"pronunciations"`
            WordLang string `json:"wordLang"`
            Type     string `json:"type"`
        } `json:"quickdef2"`
    } `json:"headwordAndQuickdefsProps"`
    AddToListProps struct {
        UILang   string `json:"uiLang"`
        IsMobile bool   `json:"isMobile"`
        Word     struct {
            ID     int    `json:"id"`
            Source string `json:"source"`
            Lang   string `json:"lang"`
        } `json:"word"`
        UserCount        int    `json:"userCount"`
        PlaygroundHost   string `json:"playgroundHost"`
        ShouldShow       bool   `json:"shouldShow"`
        ShouldBeDisabled bool   `json:"shouldBeDisabled"`
    } `json:"addToListProps"`
}

type SdDictionaryResultsProps struct {
    Entry struct {
        Chambers string `json:"chambers"`
        Collins  string `json:"collins"`
        Neodict  []struct {
            Subheadword string `json:"subheadword"`
            PosGroups   []struct {
                Pos struct {
                    AbbrEn string `json:"abbrEn"`
                    AbbrEs string `json:"abbrEs"`
                    NameEn string `json:"nameEn"`
                    NameEs string `json:"nameEs"`
                } `json:"pos"`
                EntryLang string      `json:"entryLang"`
                Gender    interface{} `json:"gender"`
                Senses    []struct {
                    ContextEn    string      `json:"contextEn"`
                    ContextEs    string      `json:"contextEs"`
                    Gender       interface{} `json:"gender"`
                    ID           int         `json:"id"`
                    PartOfSpeech struct {
                        AbbrEn string `json:"abbrEn"`
                        AbbrEs string `json:"abbrEs"`
                        NameEn string `json:"nameEn"`
                        NameEs string `json:"nameEs"`
                    } `json:"partOfSpeech"`
                    Regions        []interface{} `json:"regions"`
                    RegisterLabels []interface{} `json:"registerLabels"`
                    Translations   []struct {
                        ContextEn string `json:"contextEn"`
                        ContextEs string `json:"contextEs"`
                        Examples  []struct {
                            TextEn string `json:"textEn"`
                            TextEs string `json:"textEs"`
                        } `json:"examples"`
                        Gender                     interface{}   `json:"gender"`
                        ID                         int           `json:"id"`
                        ImagePath                  string        `json:"imagePath"`
                        IsOppositeLanguageHeadword bool          `json:"isOppositeLanguageHeadword"`
                        IsQuickTranslation         bool          `json:"isQuickTranslation"`
                        Regions                    []interface{} `json:"regions"`
                        RegisterLabels             []interface{} `json:"registerLabels"`
                        Translation                string        `json:"translation"`
                    } `json:"translations"`
                    Subheadword           string        `json:"subheadword"`
                    Idx                   int           `json:"idx"`
                    Context               string        `json:"context"`
                    RegionsDisplay        []interface{} `json:"regionsDisplay"`
                    RegisterLabelsDisplay []interface{} `json:"registerLabelsDisplay"`
                    TranslationsDisplay   []struct {
                        Translation         string      `json:"translation"`
                        Gender              interface{} `json:"gender"`
                        IsQuickTranslation  bool        `json:"isQuickTranslation"`
                        ImagePath           string      `json:"imagePath"`
                        TranslationsDisplay struct {
                            Texts    []string      `json:"texts"`
                            Tooltips []interface{} `json:"tooltips"`
                        } `json:"translationsDisplay"`
                        Letters               string        `json:"letters"`
                        Context               string        `json:"context"`
                        RegionsDisplay        []interface{} `json:"regionsDisplay"`
                        RegisterLabelsDisplay []interface{} `json:"registerLabelsDisplay"`
                        ExamplesDisplay       [][]string    `json:"examplesDisplay"`
                    } `json:"translationsDisplay"`
                } `json:"senses"`
                PosDisplay struct {
                    Name    string `json:"name"`
                    Tooltip struct {
                        Def  string `json:"def"`
                        Href string `json:"href"`
                    } `json:"tooltip"`
                } `json:"posDisplay"`
            } `json:"posGroups"`
        } `json:"neodict"`
    } `json:"entry"`
}

type Card struct {
    Sentence    string
    Picture     string
    Audio       string
    Infinitive  string
    Definition  string
    Conjugation string
    Tag         string
}

func downloadFile(filepath string, url string) error {
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    out, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer out.Close()

    _, err = io.Copy(out, resp.Body)
    return err
}

func makeCard(response Response, ankiCollectionsDir string) (Card, error) {

    card := Card{}
    //TODO add null checks for various indices
    card.Sentence = response.SdDictionaryResultsProps.Entry.Neodict[0].PosGroups[0].Senses[0].Translations[0].Examples[0].TextEs

    img := response.Entry.Neodict[0].PosGroups[0].Senses[0].Translations[0].ImagePath
    if img != "" {
        log.Println(img)
        imgUrl := fmt.Sprintf("https://d25rq8gxcq0p71.cloudfront.net%+v", img)
        imgName := fmt.Sprintf("%s%d.jpg", response.Entry.Neodict[0].PosGroups[0].Senses[0].Subheadword, time.Now().Unix())
        imgPath := fmt.Sprintf("%s/%s", ankiCollectionsDir, imgName)
        err := downloadFile(imgPath, imgUrl)
        if err != nil {
            return Card{}, err
        }
        log.Printf("saving image to: %s", imgPath)
        card.Picture = fmt.Sprintf("<img src=\"%s\">", imgName)
    }

    audioUrl := response.HeadwordAndQuickdefsProps.Headword.AudioURL
    audioName := fmt.Sprintf("%s%d.mpg3", response.HeadwordAndQuickdefsProps.Headword.DisplayText, time.Now().Unix())
    audioPath := fmt.Sprintf("%s/%s", ankiCollectionsDir, audioName)
    err := downloadFile(audioPath, audioUrl)
    if err != nil {
        return Card{}, err
    }
    log.Printf("saving file audio file to: %s", audioPath)
    card.Audio = fmt.Sprintf("[sound:%s]", audioName)

    card.Infinitive = response.Entry.Neodict[0].PosGroups[0].Senses[0].Subheadword
    card.Definition = response.Entry.Neodict[0].PosGroups[0].Senses[0].Translations[0].Translation
    card.Tag = "to_add"
    return card, nil
}

func getData(word string) (Response, error) {
    url := "https://www.spanishdict.com/translate/"
    resp, err := http.Get(fmt.Sprintf("%s/%s", url, word))
    if err != nil {
        return Response{}, err
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return Response{}, err
    }

    re := regexp.MustCompile(`SD_COMPONENT_DATA(?:\s).*=(?:\s)(.*);`)
    m := re.FindStringSubmatch(string(body))
    if len(m) < 2 {
        return Response{}, errors.New("SD_COMPONENT_DATA not found")
    }

    r := Response{}
    err = json.Unmarshal([]byte(m[1]), &r)
    if err != nil {
        return Response{}, err
    }
    return r, nil
}

func makeCsv(cards []Card, outputFile string) error {
    file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return errors.New("can't csv create file")
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    for _, value := range cards {
        err := writer.Write([]string{
            value.Sentence,
            value.Picture,
            value.Audio,
            value.Infinitive,
            value.Definition,
            value.Conjugation,
            value.Tag,
        })
        if err != nil {
            return errors.New("writing cards to csv")
        }
    }
    return nil
}

func main() {
    wordPtr := flag.String("word", "", "word to create card for")
    collectionsDirPt := flag.String("collectionsDir", "", "location of anki collections dir. (e.g. Users/username/Library/Application Support/Anki2/collection.media)")
    outputFilePt := flag.String("outputFile", "", "csv output file name")
    flag.Parse()
    word := *wordPtr
    if word == "" {
        log.Fatal("word is required")
    }
    collectionsDir := *collectionsDirPt
    if collectionsDir == "" {
        log.Fatal("collectionsDir is required")
    }
    if _, err := os.Stat(collectionsDir); os.IsExist(err) {
        log.Fatalf("directory does not exist: %s", collectionsDir)
    }
    outputFile := *outputFilePt
    if outputFile == "" {
        log.Fatal("outputFile is required")
    }
    log.Printf("args: \n\tword: %s\n\toutPutFile: %s\n\tcollectionsDir: %s\n", word, outputFile, collectionsDir)

    r, err := getData(word)
    if err != nil {
        log.Fatal(err)
    }

    //TODO turn this into make cards to allow making for multiple definitions
    //TODO use pipe as delmiter so not collisions
    card, err := makeCard(r, collectionsDir)
    if err != nil {
        log.Fatal(err)
    }

    err = makeCsv([]Card{card}, outputFile)
    if err != nil {
        log.Fatal(err)
    }
}
