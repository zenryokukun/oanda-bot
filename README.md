# Oanda Trading Bot

Oanda APIを使った取引bot。APIのドキュメントはこちら：  
https://developer.oanda.com/rest-live-v20/introduction/


## 必要なファイル
- <u>key.json</u>  
  Oanda APIの口座IDとトークン

  ```json
  {
    "live":{
        "id":"live-account-id",
        "token":"live-api-key"
    },
    "demo":{
        "id":"demo-account-id",
        "token":"demo-api-key"
    }
  }
  ```

- <u>param.json</u>  

  通貨ペアや損切ライン等のパラメタ。変更の可能性あり。
  ```json
  {
    "Inst":"USD_JOY",
    "Gran":"M5",
    "Seconds":300,
    "Span":12,
    "Thresh":0.0025,
    "ProfRate":0.005,
    "LossRate":-0.005,
    "Spread":0.016,
    "Units":10000
  }
  ```

- <u>twitter.json</u>  
  twitterのAPI。ツイート用。
  ```json
  {
    "API_KEY":"api-key",
    "API_SECRET":"api-secret",
    "BEARER":"bearer",
    "ACCESS_TOKEN":"access-token",
    "ACCESS_SECRET":"access-secret"
  }
  ```
  
  ## 作成されるファイル

  - <u>trade.json</u>  
    取引の履歴が出力される
  ```json
  {
    "X":[unixTimestamp,...],
    "Y":[prices,...],
    "Action":["OPEN","CLOSE",...],
    "Side":["BUY","SELL",...]
  }
  ```

  - <u>balance.json</u>  
    総利益の推移が出力される
  ```json
  {
    "X":[unixTimestamp,...],
    "Y":[prices,...],
    "TotalPL":[totalProfit,...]
  }
  ```

## 起動方法

- 「必要なファイル」をプロジェクトファイルの直下に配置
- ./bring.shを実行。コンパイルし実行ファイルをプロジェクトファイルの直下にmvしてくれる。
- pm2で起動
```bash
# 初回起動
pm2 start oanda-bot
pm2 save
```
```bash
# 再起動
pm2 restart oanda-bot
```

## PM2備忘
```bash
pm2 show "your file name"
pm2 status
pm2 stop "your file name"
```