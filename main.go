package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"gopkg.in/yaml.v2"
)

// 設定情報
const (
	WNL_TIMETABLE_URL    = "http://smtgvs.weathernews.jp/a/solive_timetable/timetable.json"
	WNL_TITLE            = "ウェザーニュースLiVE"
	SLACK_TOKEN_FILENAME = "slack_token.yml"
	LINE_SEPARATOR_CHAR  = "-"
	LINE_SEPARATOR_WIDTH = 80
)

// JSTの補正
const (
	JST_OFFSET_SECONDS = 9 * 60 * 60
	HOURS_OF_DAY       = 24
)

// 終了コード
const (
	ExitCodeOK = iota // 0
	ExitCodeErrorLoadSlackInfo
	ExitCodeErrorFetchTimeTable
	ExitCodeErrorParseTimeTable
	ExitCodeErrorPostToSlack
)

// WNL番組表JSON構造体
type TimeTableEntry []struct {
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
var casterList = map[string]string{
	"ailin":       "山岸愛梨",
	"kawabata":    "川畑玲",
	"komaki2018":  "駒木結衣",
	"sayane":      "江川清音",
	"shirai":      "白井ゆかり",
	"takayama":    "高山奈々",
	"tokita":      "戸北美月",
	"kobayashi":   "小林李衣奈",
	"ogawa":       "小川千奈",
	"uozumi":      "魚住茉由",
	"matsu":       "松雪彩花",
	"okamoto2023": "岡本結子リサ",
	"aohara2023":  "青原桃香",
	"fukuyoshi":   "福吉貴文",
}

// Slack情報の取得
func loadSlackInfo() (Slack, error) {
	var s Slack
	// Slack情報ファイルの有無確認
	_, err := os.Stat(SLACK_TOKEN_FILENAME)
	if !os.IsNotExist(err) {
		// Slack情報ファイルが存在した場合
		tokenFile, err := os.Open(SLACK_TOKEN_FILENAME)
		if err != nil {
			return s, fmt.Errorf("failed to open file(%s): %w", SLACK_TOKEN_FILENAME, err)
		}
		defer tokenFile.Close()

		if err := yaml.NewDecoder(tokenFile).Decode(&s); err != nil {
			return s, fmt.Errorf("failed to read Slack token file(%s): %w", SLACK_TOKEN_FILENAME, err)
		}
		// fmt.Fprintln(os.Stderr, ">>>Slack Token(file):"+s.Token) // DEBUG
	} else if os.IsNotExist(err) {
		// Slack情報ファイルが存在しない場合
		s.Token = os.Getenv("SLACK_TOKEN")
		s.Channel = os.Getenv("SLACK_CHANNEL")
	} else {
		// その他のエラー
		return s, fmt.Errorf("failed to get Slack token file(%s): %w", SLACK_TOKEN_FILENAME, err)
	}

	if s.Token == "" || s.Channel == "" {
		return s, fmt.Errorf("failed to get Slack token: %w", err)
		// fmt.Fprintln(os.Stderr, ">>>Slack Token(env):"+s.Token) // DEBUG
	}
	return s, nil
}

// 番組表の取得
func fetchTimeTable() ([]byte, error) {
	resp, err := http.Get(WNL_TIMETABLE_URL)
	if err != nil {
		return nil, fmt.Errorf("failed to get timetable(%s): %w", WNL_TIMETABLE_URL, err)
	}
	defer resp.Body.Close()
	timeTableJSON, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read timetable: %w", err)
	}

	return ([]byte)(timeTableJSON), nil
}

// 番組表の解析
func parseTimeTable(timeTableJSON []byte) (*TimeTableEntry, error) {
	timeTable := new(TimeTableEntry)

	if err := json.Unmarshal(timeTableJSON, timeTable); err != nil {
		return nil, fmt.Errorf("failed to parse timetable JSON: %w", err)
	}
	// fmt.Fprintf(os.Stderr, "get %d timetables.\n", len(*data)) //DEBUG

	return timeTable, nil
}

// 番組表文字列の作成
func buildTimeTableMessage(timeTable *TimeTableEntry) string {
	dayOffset := 0
	output := ""
	for _, v := range *timeTable {
		// 日付の補正
		if v.Hour == "00:00" {
			dayOffset = 1
			output += strings.Repeat(LINE_SEPARATOR_CHAR, LINE_SEPARATOR_WIDTH) + "\n"
		}

		// タイトルが空の行はスキップする
		if v.Title == WNL_TITLE {
			continue
		}

		// 今日の日付を取得
		tzJST := time.FixedZone("JST", JST_OFFSET_SECONDS)

		today := time.Now().UTC().In(tzJST)
		if dayOffset > 0 {
			today = today.Add(time.Hour * HOURS_OF_DAY * time.Duration(dayOffset))
		}

		// 日付を付ける
		dateTime := fmt.Sprintf("%04d/%02d/%02d %s", today.Year(), today.Month(), today.Day(), v.Hour)

		// キャスター名変換
		casterName, ok := casterList[v.Caster]
		if !ok {
			casterName = "-"
		}

		// メッセージを作成
		output += fmt.Sprintf("%s %s %s(%s)\n", dateTime, v.Title, casterName, v.Caster)
	}

	return output
}

// Slack投稿
func postMessageToSlack(s Slack, output string) error {
	ch := slack.New(s.Token)
	if _, _, err := ch.PostMessage(s.Channel, slack.MsgOptionText(output, true)); err != nil {
		return fmt.Errorf("failed to post message to Slack(%s): %w", s.Channel, err)
	}
	return nil
}

func main() {
	// デバッグかどうか
	debug := false
	envDebug := os.Getenv("DEBUG")
	if strings.ToUpper(envDebug) == "TRUE" {
		debug = true
	}

	// Slack情報の取得
	s, err := loadSlackInfo()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(ExitCodeErrorLoadSlackInfo)
	}

	// 番組表の取得
	timeTableJSON, err := fetchTimeTable()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(ExitCodeErrorFetchTimeTable)
	}

	// 番組表の解析
	timeTable, err := parseTimeTable(timeTableJSON)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(ExitCodeErrorParseTimeTable)
	}

	// 番組表の整形
	output := buildTimeTableMessage(timeTable)

	// 番組表の出力
	if debug {
		fmt.Print(output)
	} else {
		// Slack投稿
		if err := postMessageToSlack(s, output); err != nil {
			fmt.Println(err.Error())
			os.Exit(ExitCodeErrorPostToSlack)
		}
	}
	os.Exit(ExitCodeOK)
}
