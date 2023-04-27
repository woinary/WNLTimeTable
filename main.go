package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"gopkg.in/yaml.v2"
)

const WNLtimetable_URL = "http://smtgvs.weathernews.jp/a/solive_timetable/timetable.json"
const WNLtitle = "ウェザーニュースLiVE"
const Slack_Token_Filename = "slack_token.yml"
const SEPALATE_CHAR = "-"
const SEPALATE_LINE_WIDTH = 80

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
	"kawabata":   "川畑玲",
	"komaki2018": "駒木結衣",
	"ohshima":    "大島璃音",
	"sayane":     "江川清音",
	"shirai":     "白井ゆかり",
	"takayama":   "高山奈々",
	"tokita":     "戸北美月",
	"yuki":       "内田侑希",
	"kobayashi":  "小林李衣奈",
	"ogawa":      "小川千奈",
	"uozumi":     "魚住茉由",
}

// Slack情報の取得
func check_slack_info() (Slack, error) {
	var s Slack
	// Slack情報ファイルの有無確認
	_, err := os.Stat(Slack_Token_Filename)
	if !os.IsNotExist(err) {
		token_file, err := os.Open(Slack_Token_Filename)
		if err != nil {
			return s, errors.New("Slack token file Open error: " + err.Error())
		}
		defer token_file.Close()

		if err := yaml.NewDecoder(token_file).Decode(&s); err != nil {
			return s, errors.New("Slack token file read error: " + err.Error())
		}
		// fmt.Fprintln(os.Stderr, ">>>Slack Token(file):"+s.Token) // DEBUG
	} else {
		s.Token = os.Getenv("SLACK_TOKEN")
		s.Channel = os.Getenv("SLACK_CHANNEL")
	}

	if s.Token == "" || s.Channel == "" {
		return s, errors.New("Cannot get Slack token")
		// fmt.Fprintln(os.Stderr, ">>>Slack Token(env):"+s.Token) // DEBUG
	}
	return s, nil
}

// 番組表の取得
func get_time_table() (*WNLtimetable, error) {
	resp, err := http.Get(WNLtimetable_URL)
	if err != nil {
		return nil, errors.New("Cannot get timetable: " + err.Error())
	}
	defer resp.Body.Close()
	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("Cannot get timetable: " + err.Error())
	}

	jsonBytes := ([]byte)(byteArray)
	data := new(WNLtimetable)

	if err := json.Unmarshal(jsonBytes, data); err != nil {
		return nil, errors.New("Cannot parse JSON data: " + err.Error())
	}
	// fmt.Fprintf(os.Stderr, "get %d timetables.\n", len(*data)) //DEBUG

	return data, nil
}

// 番組表の作成
func make_time_table(data *WNLtimetable) string {
	day_offset := 0
	output := ""
	for i := 0; i < len(*data); i++ {
		// 日付の補正
		if (*data)[i].Hour == "00:00" {
			day_offset = 1
			output += strings.Repeat(SEPALATE_CHAR, SEPALATE_LINE_WIDTH) + "\n"
		}

		// タイトルが空の行はスキップする
		if (*data)[i].Title == WNLtitle {
			continue
		}

		// 今日の日付を取得
		tz_JST := time.FixedZone("JST", +9*60*60)

		today := time.Now().UTC().In(tz_JST)
		if day_offset > 0 {
			today = today.Add(time.Hour * 24 * time.Duration(day_offset))
		}

		// 日付を付ける
		date_time := fmt.Sprintf("%04d/%02d/%02d %s", today.Year(), today.Month(), today.Day(), (*data)[i].Hour)

		// キャスター名変換
		caster_name, ok := CasterList[(*data)[i].Caster]
		if !ok {
			caster_name = "-"
		}

		// メッセージを作成
		output += fmt.Sprintf("%s %s %s(%s)\n", date_time, (*data)[i].Title, caster_name, (*data)[i].Caster)
	}

	return output
}

// Slack投稿
func put_slack(s Slack, output string) error {
	ch := slack.New(s.Token)
	if _, _, err := ch.PostMessage(s.Channel, slack.MsgOptionText(output, true)); err != nil {
		return errors.New("Cannot put to Slack: " + err.Error())
	}
	return nil
}

func main() {
	// デバッグかどうか
	debug := false
	env_debug := os.Getenv("DEBUG")
	if strings.ToUpper(env_debug) == "TRUE" {
		debug = true
	}

	// Slack情報の取得
	s, err := check_slack_info()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// 番組表の取得
	data, err := get_time_table()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}

	// 番組表の整形
	output := make_time_table(data)

	// 番組表の出力
	if debug {
		fmt.Print(output)
	} else {
		// Slack投稿
		if err := put_slack(s, output); err != nil {
			fmt.Println(err.Error())
			os.Exit(3)
		}
	}
	os.Exit(0)
}
