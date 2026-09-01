package market

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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

const (
	liveReconnectInitialBackoff = time.Second
	liveReconnectMaxBackoff     = 30 * time.Second
)

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
		backoff := liveReconnectInitialBackoff
		for ctx.Err() == nil {
			endpoint := strings.TrimRight(b.WSBaseURL, "/") + "/" + strings.ToLower(symbol) + "@kline_" + timeframe
			conn, _, err := websocket.Dial(ctx, endpoint, nil)
			if err == nil {
				backoff = liveReconnectInitialBackoff
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
			backoff = nextLiveReconnectBackoff(backoff)
		}
	}()
	return out
}

func (b *Binance) StreamMarketEvents(ctx context.Context, symbols, timeframes []string) <-chan LiveEvent {
	out := make(chan LiveEvent)
	go func() {
		defer close(out)
		candleStreams := make([]string, 0, len(symbols)*len(timeframes))
		tradeStreams := make([]string, 0, len(symbols))
		seen := make(map[string]struct{})
		for _, rawSymbol := range symbols {
			symbol := strings.ToUpper(strings.TrimSpace(rawSymbol))
			if !IsSupportedSymbol(symbol) {
				continue
			}
			for _, rawFrame := range timeframes {
				frame := strings.TrimSpace(rawFrame)
				if !validTimeframe(frame) {
					continue
				}
				stream := strings.ToLower(symbol) + "@kline_" + frame
				if _, ok := seen[stream]; !ok {
					seen[stream] = struct{}{}
					candleStreams = append(candleStreams, stream)
				}
			}
			tradeStream := strings.ToLower(symbol) + "@aggTrade"
			if _, ok := seen[tradeStream]; !ok {
				seen[tradeStream] = struct{}{}
				tradeStreams = append(tradeStreams, tradeStream)
			}
		}
		if len(candleStreams) == 0 && len(tradeStreams) == 0 {
			return
		}
		candles := b.streamCombined(ctx, candleStreams)
		trades := b.streamCombined(ctx, tradeStreams)
		for candles != nil || trades != nil {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-candles:
				if !ok {
					candles = nil
					continue
				}
				select {
				case out <- event:
				case <-ctx.Done():
					return
				}
			case event, ok := <-trades:
				if !ok {
					trades = nil
					continue
				}
				select {
				case out <- event:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

func (b *Binance) streamCombined(ctx context.Context, streams []string) <-chan LiveEvent {
	out := make(chan LiveEvent, 256)
	go func() {
		defer close(out)
		if len(streams) == 0 {
			return
		}
		backoff := liveReconnectInitialBackoff
		for ctx.Err() == nil {
			endpoint := combinedStreamEndpoint(b.WSBaseURL, streams)
			conn, _, err := websocket.Dial(ctx, endpoint, nil)
			if err == nil {
				backoff = liveReconnectInitialBackoff
				for {
					var wrapped binanceCombinedEvent
					if err = wsjson.Read(ctx, conn, &wrapped); err != nil {
						break
					}
					event, parseErr := wrapped.liveEvent()
					if parseErr != nil {
						log.Printf("skip Binance stream event %q: %v", wrapped.Stream, parseErr)
						continue
					}
					select {
					case out <- event:
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
			backoff = nextLiveReconnectBackoff(backoff)
		}
	}()
	return out
}

func combinedStreamEndpoint(base string, streams []string) string {
	base = strings.TrimRight(base, "/")
	base = strings.TrimSuffix(base, "/ws")
	return base + "/stream?streams=" + strings.Join(streams, "/")
}

func nextLiveReconnectBackoff(current time.Duration) time.Duration {
	if current <= 0 {
		return liveReconnectInitialBackoff
	}
	if current >= liveReconnectMaxBackoff {
		return liveReconnectMaxBackoff
	}
	next := current * 2
	if next > liveReconnectMaxBackoff {
		return liveReconnectMaxBackoff
	}
	return next
}

type binanceKlineEvent struct {
	K struct {
		Start    int64  `json:"t"`
		LastID   int64  `json:"L"`
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

type binanceCombinedEvent struct {
	Stream string          `json:"stream"`
	Data   json.RawMessage `json:"data"`
}

type binanceAggTradeEvent struct {
	Symbol    string `json:"s"`
	TradeID   int64  `json:"a"`
	Price     string `json:"p"`
	Quantity  string `json:"q"`
	TradeTime int64  `json:"T"`
	BuyerMade bool   `json:"m"`
}

func (e binanceCombinedEvent) liveEvent() (LiveEvent, error) {
	switch {
	case strings.Contains(e.Stream, "@kline_"):
		var payload binanceKlineEvent
		if err := json.Unmarshal(e.Data, &payload); err != nil {
			return LiveEvent{}, fmt.Errorf("decode combined kline: %w", err)
		}
		candle, err := payload.candle()
		if err != nil {
			return LiveEvent{}, err
		}
		return LiveEvent{Type: "CANDLE_UPDATE", Candle: candle}, nil
	case strings.HasSuffix(strings.ToLower(e.Stream), "@aggtrade"):
		var payload binanceAggTradeEvent
		if err := json.Unmarshal(e.Data, &payload); err != nil {
			return LiveEvent{}, fmt.Errorf("decode aggregate trade: %w", err)
		}
		trade, err := payload.trade()
		if err != nil {
			return LiveEvent{}, err
		}
		return LiveEvent{Type: "TRADE_TICK", Trade: trade}, nil
	default:
		return LiveEvent{}, errors.New("unsupported combined stream event")
	}
}

func (e binanceAggTradeEvent) trade() (TradeTick, error) {
	price, err := strconv.ParseFloat(e.Price, 64)
	if err != nil {
		return TradeTick{}, fmt.Errorf("invalid aggregate trade price: %w", err)
	}
	quantity, err := strconv.ParseFloat(e.Quantity, 64)
	if err != nil {
		return TradeTick{}, fmt.Errorf("invalid aggregate trade quantity: %w", err)
	}
	if e.Symbol == "" || e.TradeID <= 0 || e.TradeTime <= 0 || price <= 0 || quantity <= 0 {
		return TradeTick{}, errors.New("invalid aggregate trade identity")
	}
	side := "BUY"
	if e.BuyerMade {
		side = "SELL"
	}
	return TradeTick{Symbol: e.Symbol, TradeID: e.TradeID, TradeTime: e.TradeTime, Price: price, Quantity: quantity, Side: side}, nil
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
