package market

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

const (
	binanceWSBaseURL     = "wss://stream.binance.com:9443/ws"
	binanceTestnetWSBase = "wss://testnet.binance.vision/ws" // for dev, if available
)

type Binance struct{}

type binanceKlineMsg struct {
	K struct {
		StartTime    int64  `json:"t"`
		CloseTime    int64  `json:"T"`
		Open         string `json:"o"`
		High         string `json:"h"`
		Low          string `json:"l"`
		Close        string `json:"c"`
		Volume       string `json:"v"`
		LastTradeID  int64  `json:"L"`
		FirstTradeID int64  `json:"f"`
		TradeCount   int    `json:"n"`
		IsClosed     bool   `json:"x"`
	} `json:"k"`
}

func (b *Binance) StreamLiveCandles(symbol, timeframe string, out chan<- Candle) error {
	url := fmt.Sprintf("%s/%s@kline_%s", binanceWSBaseURL, symbol, timeframe)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		var raw binanceKlineMsg
		if err := json.Unmarshal(msg, &raw); err != nil {
			log.Println("parse error:", err, "raw:", string(msg))
			continue
		}
		log.Println("parsed candle open time:", raw.K.StartTime) // temporary debug

		c := Candle{Symbol: symbol, Timeframe: timeframe, OpenTime: raw.K.StartTime}
		json.Unmarshal([]byte(raw.K.Open), &c.Open)
		json.Unmarshal([]byte(raw.K.High), &c.High)
		json.Unmarshal([]byte(raw.K.Low), &c.Low)
		json.Unmarshal([]byte(raw.K.Close), &c.Close)
		json.Unmarshal([]byte(raw.K.Volume), &c.Volume)

		out <- c
	}
}
