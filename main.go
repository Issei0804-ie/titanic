package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"sync"

	"github.com/google/go-cmp/cmp"

	"github.com/PuerkitoBio/goquery"
	"github.com/int128/slack"
)

type Mattermost struct {
	webhook string
}

func NewMattermost(webhook string) *Mattermost {
	return &Mattermost{webhook: webhook}
}

func (m *Mattermost) SendMessage(message string) {
	if err := slack.Send(m.webhook, &slack.Message{
		Username: "Titanic",
		Text:     message,
	}); err != nil {
		log.Fatalf("Could not send the message to Slack: %s", err)
	}
}

func AcceseToHomePage(code string, title string, wg *sync.WaitGroup) {
	// Request the HTML page.
	res, err := http.Get("https://tiglon.jim.u-ryukyu.ac.jp/Portal/Public/Syllabus/SyllabusSearchStart.aspx?lct_year=2021&lct_cd=" + code + "&je_cd=1")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	var tmp []string
	// Find the review items
	doc.Find(".ItemName").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		if i%2 == 0 {
			text := s.Find("span").Text()
			tmp = append(tmp, text)
		}
	})
	htmlBody := HTMLBody{
		URGCC:                tmp[0],
		ActiveLearning:       tmp[1],
		AcademicCredit:       tmp[2],
		ClassForm:            tmp[3],
		ClassContent:         tmp[4],
		ClassPlan:            tmp[5],
		EvaluationCriteria:   tmp[6],
		AchievementGoal:      tmp[7],
		PreLearning:          tmp[8],
		PostLearning:         tmp[9],
		TextbookInformation:  tmp[10],
		TextbookRemarks:      tmp[11],
		TextbookInformation2: tmp[12],
		TextbookRemarks2:     tmp[13],
		Language:             tmp[14],
		Message:              tmp[15],
		OfficeHour:           tmp[16],
		Mail:                 tmp[17],
		URL:                  tmp[18],
	}

	bodyJson, err := json.Marshal(htmlBody)
	if err != nil {
		fmt.Println(err.Error())
	}
	filename := "syllabus/" + code + ".json"
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		oldHTMLBody := HTMLBody{}
		oldBodyJson, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Println(err.Error())
			os.Exit(1)
		}
		err = json.Unmarshal(oldBodyJson, &oldHTMLBody)
		if err != nil {
			log.Println(err.Error())
			os.Exit(1)
		}
		if !reflect.DeepEqual(&oldHTMLBody, &htmlBody) {
			log.Println(title + ":違うよ")
			mattermost := NewMattermost("https://mattermost.ie.u-ryukyu.ac.jp/hooks/sx5bz6zmmif15r335hrebj5xro")
			mattermost.SendMessage(title + ":違うよ")
			mattermost.SendMessage(fmt.Sprintf("%s", cmp.Diff(&oldHTMLBody, &htmlBody)))
			_ = ioutil.WriteFile(filename, bodyJson, os.ModePerm)

		}
	}
	_ = ioutil.WriteFile(filename, bodyJson, os.ModePerm)
	log.Println("title:" + title + ", code:" + code + ",status:done")

	wg.Done()
}

func main() {
	fh, err := os.Open("./course.json")
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
	body, _ := ioutil.ReadAll(fh)
	courseInformation := CourseInformation{}
	err = json.Unmarshal(body, &courseInformation)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
	var wg sync.WaitGroup
	wg.Add(len(courseInformation.Course))
	for _, course := range courseInformation.Course {
		go AcceseToHomePage(course.Code, course.Title, &wg)
	}
	wg.Wait()
	mattermost := NewMattermost("https://mattermost.ie.u-ryukyu.ac.jp/hooks/sx5bz6zmmif15r335hrebj5xro")
	mattermost.SendMessage("finish")
}

type HTMLBody struct {
	URGCC                string `json:"URGCC学習教育目標"`
	ActiveLearning       string `json:"アクティブラーニング"`
	AcademicCredit       string `json:"履修条件"`
	ClassForm            string `json:"授業の形態"`
	ClassContent         string `json:"授業内容と方法"`
	ClassPlan            string `json:"授業計画"`
	EvaluationCriteria   string `json:"評価基準と評価方法"`
	AchievementGoal      string `json:"達成目標"`
	PreLearning          string `json:"事前学習"`
	PostLearning         string `json:"事後学習"`
	TextbookInformation  string `json:"教科書にかかわる情報"`
	TextbookRemarks      string `json:"教科書全体備考"`
	TextbookInformation2 string `json:"教科書にかかわる情報2"`
	TextbookRemarks2     string `json:"教科書全体備考2"`
	Language             string `json:"使用言語"`
	Message              string `json:"メッセージ"`
	OfficeHour           string `json:"オフィスアワー"`
	Mail                 string `json:"メールアドレス"`
	URL                  string `json:"URL"`
}

type CourseInformation struct {
	Course []struct {
		Code  string `json:"code"`
		Title string `json:"title"`
	} `json:"course"`
}
