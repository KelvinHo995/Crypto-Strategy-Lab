package market

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Binance struct {
	BaseURL   string
	WSBaseURL string
	Client    *http.Client
}

func NewBinance(client *http.Client) *Binance {
	if client == nil {
		client = http.DefaultClient
	}
	return &Binance{BaseURL: "https://api.binance.com", WSBaseURL: "wss://stream.binance.com:9443/ws", Client: client}
}

func (b *Binance) FetchHistoricalCandles(ctx context.Context, symbol, timeframe string, from, to int64) ([]Candle, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	timeframe = strings.TrimSpace(timeframe)
	if symbol == "" || !validTimeframe(timeframe) || from <= 0 || to <= from {
		return nil, errors.New("invalid historical candle request")
	}
	var out []Candle
	start := from
	for start < to {
		u, err := url.Parse(b.BaseURL + "/api/v3/klines")
		if err != nil {
			return nil, err
		}
		q := u.Query()
		q.Set("symbol", symbol)
		q.Set("interval", timeframe)
		q.Set("startTime", strconv.FormatInt(start, 10))
		q.Set("endTime", strconv.FormatInt(to, 10))
		q.Set("limit", "1000")
		u.RawQuery = q.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}
		resp, err := b.Client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("binance klines: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("binance klines: status %d", resp.StatusCode)
		}
		var rows [][]json.RawMessage
		err = json.NewDecoder(resp.Body).Decode(&rows)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("decode binance klines: %w", err)
		}
		if len(rows) == 0 {
			break
		}
		last := int64(0)
		for _, row := range rows {
			c, err := parseKline(symbol, timeframe, row)
			if err != nil {
				return nil, err
			}
			last = c.OpenTime
			if c.OpenTime >= from && c.OpenTime <= to && c.IsClosed {
				out = append(out, c)
			}
		}
		if last < start {
			return nil, errors.New("binance klines did not advance")
		}
		start = last + 1
		if len(rows) < 1000 {
			break
		}
	}
	return out, nil
}
func parseKline(symbol, timeframe string, row []json.RawMessage) (Candle, error) {
	if len(row) < 7 {
		return Candle{}, errors.New("binance kline has fewer than 7 fields")
	}
	var ts int64
	if json.Unmarshal(row[0], &ts) != nil {
		return Candle{}, errors.New("invalid kline timestamp")
	}
	values := make([]float64, 5)
	for i := range values {
		var s string
		if json.Unmarshal(row[i+1], &s) != nil {
			return Candle{}, errors.New("invalid kline price")
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return Candle{}, fmt.Errorf("invalid kline number: %w", err)
		}
		values[i] = v
	}
	var closeTime int64
	if json.Unmarshal(row[6], &closeTime) != nil {
		return Candle{}, errors.New("invalid kline close timestamp")
	}
	return Candle{Symbol: symbol, Timeframe: timeframe, OpenTime: ts, Open: values[0], High: values[1], Low: values[2], Close: values[3], Volume: values[4], IsClosed: closeTime < time.Now().UnixMilli()}, nil
}
func validTimeframe(v string) bool {
	switch v {
	case "5m", "15m", "1h", "4h":
		return true
	}
	return false
}

func (b *Binance) StreamLiveCandles(ctx context.Context, symbol, timeframe string) <-chan Candle {
	out := make(chan Candle)
	go func() {
		defer close(out)
		symbol = strings.ToUpper(strings.TrimSpace(symbol))
		timeframe = strings.TrimSpace(timeframe)
		if symbol == "" || !validTimeframe(timeframe) {
			return
		}
		backoff := time.Second
		for ctx.Err() == nil {
			endpoint := strings.TrimRight(b.WSBaseURL, "/") + "/" + strings.ToLower(symbol) + "@kline_" + timeframe
			conn, _, err := websocket.Dial(ctx, endpoint, nil)
			if err == nil {
				backoff = time.Second
				for {
					var event binanceKlineEvent
					if err = wsjson.Read(ctx, conn, &event); err != nil {
						break
					}
					candle, parseErr := event.candle()
					if parseErr != nil {
						continue
					}
					select {
					case out <- candle:
					case <-ctx.Done():
						conn.CloseNow()
						return
					}
				}
				conn.CloseNow()
			}
			timer := time.NewTimer(backoff)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
			if backoff < 30*time.Second {
				backoff *= 2
			}
		}
	}()
	return out
}

type binanceKlineEvent struct {
	K struct {
		Start    int64  `json:"t"`
		Symbol   string `json:"s"`
		Interval string `json:"i"`
		Open     string `json:"o"`
		Close    string `json:"c"`
		High     string `json:"h"`
		Low      string `json:"l"`
		Volume   string `json:"v"`
		Closed   bool   `json:"x"`
	} `json:"k"`
}

func (e binanceKlineEvent) candle() (Candle, error) {
	values := make([]float64, 5)
	for i, raw := range []string{e.K.Open, e.K.High, e.K.Low, e.K.Close, e.K.Volume} {
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return Candle{}, fmt.Errorf("invalid live kline number: %w", err)
		}
		values[i] = value
	}
	if e.K.Symbol == "" || !validTimeframe(e.K.Interval) || e.K.Start <= 0 {
		return Candle{}, errors.New("invalid live kline identity")
	}
	return Candle{Symbol: e.K.Symbol, Timeframe: e.K.Interval, OpenTime: e.K.Start, Open: values[0], High: values[1], Low: values[2], Close: values[3], Volume: values[4], IsClosed: e.K.Closed}, nil
}
