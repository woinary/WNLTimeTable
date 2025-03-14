# ウェザーニュースLIVE番組表を取得して、Slackに通知する

ウェザーニュースLIVE番組表を取得し、その内容を読みやすい形に変換してSlackに書き込みます。

## 使い方

slack_token.yml.templateを参考にして、slack_token.ymlを作成しておきます。
slackに投稿するためのトークンの取得方法は以下を参照してください。

[Slack API を使用してメッセージを投稿する](https://zenn.dev/kou_pg_0131/articles/slack-api-post-message)

取得したトークンをslack_token.ymlのslackTokenの所に転記してください。
または、環境変数SLACK_TOKEN,SLACK_CHANNELを設定してください。

書き込み先のチャンネルの設定や、アプリの設定は事前に行なってください。
以下を参考にさせていただきました。

[【Go】slack-go を使用して Slack にメッセージを投稿する](https://zenn.dev/kou_pg_0131/articles/go-slack-go-usage)

実行すると、タイムテーブルを取得して、以下のように整形して、指定したSlackのチャンネルに書き込みます。

>2022/05/06 17:00 ウェザーニュースLiVE・イブニング 白井ゆかり(shirai)
2022/05/06 20:00 ウェザーニュースLiVE・ムーン  山岸愛梨(ailin)
2022/05/07 05:00 ウェザーニュースLiVE・モーニング 内田侑希(yuki)
2022/05/07 08:00 ウェザーニュースLiVE・サンシャイン 高山奈々(takayama)
2022/05/07 11:00 ウェザーニュースLiVE・コーヒータイム 戸北美月(tokita)
2022/05/07 14:00 ウェザーニュースLiVE・アフタヌーン 駒木結衣(komaki2018)
2022/05/07 17:00 ウェザーニュースLiVE・イブニング 檜山沙耶(hiyama2018)
2022/05/07 20:00 ウェザーニュースLiVE・ムーン  大島璃音(ohshima)

### デバッグ出力

環境変数DEBUGにTRUEを設定すると、Slackに投稿せずに、標準出力に出します。大文字小文字は区別しません。

## 仕組みについて

タイムテーブルはウェザーニュースからJSONファイルを取得しています。元のデータには時刻、番組名、キャスター名（英数字）が入っているので、以下整形をしています。

* 時刻には日付を付与
* キャスター名は日本語（漢字）に置き換え。名前の後ろに`()`で元の英数字の表記を残してあります。
* 元データには23時と0時がありますが、担当キャスターが空なのでスキップしています。

キャスター名の変換はソース埋め込みになっているので、キャスターの増減はソースの修正が必要になります。

ウェザーニュース非公認のため、このプログラムについてウェザーニューズ株式会社への問い合わせはおやめください。また、サーバの負荷になるようなアクセスはお控えください。
このプログラムはJSONを取得してSlackに書き込むサンプルとして作成しています。

## 履歴

* 2025/03/03
  * GitHub Copilot(free)を利用したプログラムの見直し
    * loadSlickInfo()のエラーハンドリングを修正し、何らかのエラーが発生した場合の処理を追加
    * 退職済みのキャスターを削除
* 2025/02/26(2)
  * Gemini Code Assistを利用したプログラムの見直し
    * エラーハンドリングの変更、エラーメッセージの詳細化
    * Go言語の慣習に合わせた関数名、変数名、定数名への変更
    * 番組表取得部分と解析部分を分離
    * マジックナンバーの定数化
    * その他、スペルミス等の修正
* 2025/02/26(1)
  * 福吉キャスターに対応
  * "error strings should not be capitalized"対応
* 2024/01/04
  * 青原、岡本両キャスターの名前が入れ替わっていた不具合の修正
* 2023/10/22
  * 青原桃香、岡本結子リサキャスターに対応
* 2023/05/03
  * 松雪彩花キャスターに対応
* 2023/04/27
  * 小川千奈、魚住茉由キャスターに対応
* 2022/09/01
  * 小林李衣奈キャスターに対応
  * 時間帯で日付がおかしくなる不具合対応
* 2022/05/26
  * 川畑玲キャスターに対応
* 2022/05/31 (1.1)
  * 日付計算の不具合修正
  * Slack情報を環境変数からも取得できるように変更
  * Slack情報ファイルの存在チェック
  * Heroku用Procfileを追加
* 2022/06/01
  * 日付の変わり目に区切り線を表示
  * DEBUG出力オプションの追加
