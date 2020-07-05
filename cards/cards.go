package cards

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

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

type Response struct {
	ResultCardHeaderProps    `json:"resultCardHeaderProps"`
	SdDictionaryResultsProps `json:"sdDictionaryResultsProps"`
}

type ResultCardHeaderProps struct {
	HeadwordAndQuickdefsProps struct {
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
}

type SdDictionaryResultsProps struct {
	Entry struct {
		Chambers string `json:"chambers"`
		Collins  string `json:"collins"`
		Neodict  `json:"neodict"`
	} `json:"entry"`
	HegemoneAssetHost string `json:"hegemoneAssetHost"`
}

type Neodict []struct {
	Subheadword string `json:"subheadword"`
	PosGroups   []struct {
		Pos struct {
			AbbrEn string `json:"abbrEn"`
			AbbrEs string `json:"abbrEs"`
			NameEn string `json:"nameEn"`
			NameEs string `json:"nameEs"`
		} `json:"pos"`
		EntryLang  string      `json:"entryLang"`
		Gender     interface{} `json:"gender"`
		Senses     []Sense     `json:"senses"`
		PosDisplay struct {
			Name    string `json:"name"`
			Tooltip struct {
				Def  string `json:"def"`
				Href string `json:"href"`
			} `json:"tooltip"`
		} `json:"posDisplay"`
	} `json:"posGroups"`
}

type Sense struct {
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

func GetData(word string) (Response, error) {
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

func MakeCards(response Response, ankiCollectionsDir string, defnNumber int) ([]Card, error) {
	if len(response.SdDictionaryResultsProps.Entry.Neodict) == 0 {
		logger.Error("no dict entry")
		return []Card{}, errors.New("no dict entry")
	}

	audioTag, err := getAudioFile(response, ankiCollectionsDir)
	if err != nil {
		logger.Error("getting audio file")
		return []Card{}, err
	}
	cloudFront := fmt.Sprintf("https://%s", response.HegemoneAssetHost)
	senses := flattenDefns(response, defnNumber)
	var cards []Card
	for _, sense := range senses {
		if len(sense.Translations) == 0 || len(sense.Translations[0].Examples) == 0 {
			logger.Infof("no examples for defnNumber %d", defnNumber)
			continue
		}
		card, err := makeCard(sense, ankiCollectionsDir, cloudFront)
		if err != nil {
			logger.Error("error creating card", err)
		} else {
			card.Audio = audioTag
			cards = append(cards, card)
		}
	}
	return cards, nil
}

func getAudioFile(response Response, ankiCollectionsDir string) (string, error) {
	audioUrl := response.HeadwordAndQuickdefsProps.Headword.AudioURL
	var audioName string
	if audioUrl != "" {
		logger.Infof(audioUrl)
		name := strings.Replace(response.HeadwordAndQuickdefsProps.Headword.DisplayText, " ", "_", -1)
		audioName = fmt.Sprintf("%s%d.mp3", name, time.Now().UnixNano())
		audioPath := fmt.Sprintf("%s/%s", ankiCollectionsDir, audioName)
		err := downloadFile(audioPath, audioUrl)
		if err != nil {
			return "", err
		}
		logger.Infof("saving file audio file to: %s", audioPath)
	}
	audioTag := fmt.Sprintf("[sound:%s]", audioName)
	return audioTag, nil
}

func MakeCsv(cards []Card, outputFile string) error {
	file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.New("can't csv create file")
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range cards {
		if value.Infinitive == "" {
			continue
		}
		err := writer.Write([]string{
			value.Sentence,
			value.Infinitive,
			value.Picture,
			value.Audio,
			value.Definition,
			value.Conjugation,
		})
		if err != nil {
			return errors.New("writing cards to csv")
		}
	}
	return writer.Error()
}

func downloadFile(path string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	extension := filepath.Ext(path)
	if extension == "jpg" {
		img, _, err := image.Decode(resp.Body)
		if err != nil {
			return err
		}
		return jpeg.Encode(out, img, nil)
	}
	_, err = io.Copy(out, resp.Body)
	return err
}

func flattenDefns(response Response, defnNumber int) []Sense {
	var senses []Sense
	posGroups := response.SdDictionaryResultsProps.Entry.Neodict[0].PosGroups
	for i := 0; i < len(posGroups); i++ {
		posGroup := posGroups[i]
		for j := 0; j < len(posGroup.Senses); j++ {
			if len(posGroup.Senses) > j {
				senses = append(senses, posGroup.Senses[j])
				if len(senses) >= defnNumber {
					return senses
				}
			}
		}
	}
	return senses
}

func makeCard(sense Sense, ankiCollectionsDir string, cloudfront string) (Card, error) {
	card := Card{}
	card.Sentence = sense.Translations[0].Examples[0].TextEs
	file := sense.Translations[0].ImagePath
	if file != "" {
		// shouldn't be doing this but appears image name only in cloudfront urls
		// have been double escaped so spaces %20 -> were escaped again to %2520
		file = url.PathEscape(filepath.Base(sense.Translations[0].ImagePath))
		dir := filepath.Dir(sense.Translations[0].ImagePath)
		img := fmt.Sprintf("%s/%s", dir, file)
		imgUrl := fmt.Sprintf("%s%s", cloudfront, strings.Replace(img, "/original/", "/300/", 1))
		logger.Infof(imgUrl)
		name := strings.Replace(sense.Subheadword, " ", "_", -1)
		imgName := fmt.Sprintf("%s%d.jpg", name, time.Now().UnixNano())
		imgPath := fmt.Sprintf("%s/%s", ankiCollectionsDir, imgName)
		err := downloadFile(imgPath, imgUrl)
		if err != nil {
			return Card{}, err
		}
		logger.Infof("saving image to: %v", imgPath)
		card.Picture = fmt.Sprintf("<img src=\"%s\">", imgName)
	}
	card.Infinitive = sense.Subheadword
	card.Definition = sense.Translations[0].Translation
	return card, nil
}
