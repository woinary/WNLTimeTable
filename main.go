package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/slack-go/slack"
	"gopkg.in/yaml.v2"
)

const WNLtimetable_URL = "http://smtgvs.weathernews.jp/a/solive_timetable/timetable.json"
const WNLtitle = "ウェザーニュースLiVE"
const Slack_Token_Filename = "slack_token.yml"

// WNL番組表JSON構造体
type WNLtimetable []struct {
	Hour   string `json:"hour"`
	Title  string `json:"title"`
	Caster string `json:"caster"`
}

// Slacjトークン構造体
type Slack struct {
	Token   string `yaml:"slackToken"`
	Channel string `yaml:"slackChannel"`
}

// キャスターリスト
var CasterList = map[string]string{
	"ailin":      "山岸愛梨",
	"hiyama2018": "檜山沙耶",
	"komaki2018": "駒木結衣",
	"ohshima":    "大島璃音",
	"sayane":     "江川清音",
	"shirai":     "白井ゆかり",
	"takayama":   "高山奈々",
	"tokita":     "戸北美月",
	"yuki":       "内田侑希",
}

func main() {
	// Slackトークンの取得
	token_file, err := os.Open(Slack_Token_Filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	defer token_file.Close()

	var s Slack
	if err := yaml.NewDecoder(token_file).Decode(&s); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	// fmt.Fprintln(os.Stderr, ">>>Slack Token:"+s.Token) // DEBUG

	// 番組表の取得
	resp, err := http.Get(WNLtimetable_URL)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
	defer resp.Body.Close()
	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(3)
	}

	jsonBytes := ([]byte)(byteArray)
	data := new(WNLtimetable)

	if err := json.Unmarshal(jsonBytes, data); err != nil {
		fmt.Println("JSON Unmarsharl error: ", err.Error())
		os.Exit(4)
	}
	// fmt.Fprintf(os.Stderr, "get %d timetables.\n", len(*data)) //DEBUG

	day_offset := 0
	output := ""
	for i := 0; i < len(*data); i++ {
		// 日付の補正
		if (*data)[i].Hour == "00:00" {
			day_offset = 1
		}

		// タイトルが空の行はスキップする
		if (*data)[i].Title == WNLtitle {
			continue
		}

		// 今日の日付を取得
		today := time.Now()

		// 日付を付ける
		date_time := fmt.Sprintf("%04d/%02d/%02d %s", today.Year(), today.Month(), today.Day()+day_offset, (*data)[i].Hour)

		// キャスター名変換
		caster_name, ok := CasterList[(*data)[i].Caster]
		if !ok {
			caster_name = "-"
		}

		// メッセージを作成
		output += fmt.Sprintf("%s %s %s(%s)\n", date_time, (*data)[i].Title, caster_name, (*data)[i].Caster)
	}
	// fmt.Print(output)
	// Slack投稿
	ch := slack.New(s.Token)
	if _, _, err := ch.PostMessage(s.Channel, slack.MsgOptionText(output, true)); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}
