# BÁO CÁO KIỂM TOÁN CHỨC NĂNG MVP (FUNCTIONAL COMPLIANCE AUDIT)

**Dự án:** Crypto Strategy Lab  
**Hạng mục:** NHÓM B. CHỨC NĂNG YÊU CẦU - MVP (Tối đa: 40 điểm)  

---

## 1. BẢNG ĐÁNH GIÁ ĐỐI SOÁT CHI TIẾT 10 TIÊU CHÍ NHÓM B

| STT | Tiêu chí đánh giá | Điểm tối đa | Điểm tự chấm | Bằng chứng mã nguồn (File, Struct, Component, Endpoint, Dòng code) |
|:---:|---|:---:|:---:|---|
| **1** | **Market Data & Candlestick** | **6** | **6 / 6** | - Backfill REST: [cmd/backfill/main.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/cmd/backfill/main.go#L35-L80) lưu vào bảng `candles`<br>- Timeframes: `5m, 15m, 1h, 4h, 1d`; Pairs: `BTCUSDT, ETHUSDT, SOLUSDT...`<br>- Realtime WSS: [binance.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/market/binance.go#L243-L286)<br>- Frontend Chart: [TradingChart.tsx](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/frontend/src/features/market/components/TradingChart.tsx) & [ChartCard.tsx](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/frontend/src/features/market/components/ChartCard.tsx#L40-L100) render nến OHLC + Volume |
| **2** | **Strategy Engine ($\ge$ 4 strategies)** | **5** | **5 / 5** | - Triển khai **7 chiến lược** (vượt yêu cầu $\ge 4$):<br>  1. [MA](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/ma.go#L10-L40), 2. [RSI](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/rsi.go#L10-L50), 3. [Bollinger Bands](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/bollinger.go#L10-L45), 4. [Support/Resistance](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/sr.go#L10-L35), 5. [SMC](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/smc.go#L10-L45), 6. [MACD](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/macd.go#L10-L45), 7. [Sentiment](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/sentiment.go#L47-L80)<br>- Chuẩn hóa đầu ra: `Signal` enum (`BUY`, `SELL`, `HOLD`) trong [strategy.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/strategy.go#L10-L16) |
| **3** | **Composite Strategy** | **5** | **5 / 5** | - Thuật toán gộp tín hiệu: [combination.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/combination.go#L9-L56)<br>  + `MajorityPolicy`: Bầu chọn đa số phiếu giữa các tín hiệu<br>  + `WeightedPolicy`: Tính tổng điểm có trọng số $\sum (S_i \times W_i)$<br>- Giao diện dựng chiến lược: [CompositeStrategyBuilder.tsx](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/frontend/src/features/strategy/components/CompositeStrategyBuilder.tsx#L24-L80) hỗ trợ kéo slider trọng số, chọn policy và gửi mảng `instances[]` |
| **4** | **Backtesting Engine** | **6** | **6 / 6** | - Giả lập khớp lệnh: [backtester.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/experiment/backtester.go#L24-L79)<br>  + Triệt tiêu lookahead bias: `candles[i-Window : i]` (dòng 52)<br>  + Xử lý trượt giá (`SlippageBps`), phí sàn (`FeePct`), Stop Loss & Take Profit (`checkStopTarget`)<br>- Tách biệt hoàn toàn khỏi logic chiến lược<br>- Chi tiết khớp lệnh: Bảng `experiment_trades` (Migration `0010`) & Endpoint `GET /experiments/{id}/trades` |
| **5** | **Strategy Evaluation** | **4** | **4 / 4** | - Đánh giá chỉ số: [evaluator.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/experiment/evaluator.go#L15-L46)<br>  + Lợi nhuận ròng: `Return %`<br>  + Tỷ lệ thắng: `WinRate %`, `Wins`, `Losses`<br>  + Sụt giảm lớn nhất: `MDD %` (Max Drawdown)<br>  + Số lượng giao dịch: `TradeCount`<br>  + Tổng lợi nhuận danh nghĩa: `TotalProfit ($)` |
| **6** | **Strategy Search & Stop Condition** | **5** | **5 / 5** | - Random Search & Enqueue: [random_generator.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/random_generator.go#L27-L65) & [loop.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/experiment/loop.go#L49-L83)<br>- Deduplication: Cơ chế hash key `candidateKey(candidate)` tránh trùng lặp (dòng 107-114)<br>- 3 điều kiện dừng (OR semantics, dòng 85-104):<br>  1. `maxCandidates`, 2. `maxDuration`, 3. `noImprovementLimit` |
| **7** | **Leaderboard Top-K** | **3** | **3 / 3** | - Bảng xếp hạng: [ExperimentLeaderboard.tsx](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/frontend/src/features/experiment/components/ExperimentLeaderboard.tsx) & `GET /experiments`<br>- Cập nhật Realtime: Tự động render lại khi nhận event `LEADERBOARD_UPDATE` qua WebSocket Hub<br>- Tính năng: Sắp xếp theo Overall Score, Return %, Win Rate % và lọc Top-K |
| **8** | **Visualization Strategy & Trade** | **3** | **3 / 3** | - Vẽ Marker vào/thoát lệnh: [BacktestChart.tsx](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/frontend/src/features/experiment/components/BacktestChart.tsx#L38-L45) (▲ BUY / ▼ SELL)<br>- Đồng bộ tương tác: [TradeHistoryTable.tsx](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/frontend/src/features/experiment/components/TradeHistoryTable.tsx#L45-L65) click từng dòng trade để highlight và focus đúng tọa độ nến Entry/Exit trên biểu đồ |
| **9** | **News Collector** | **2** | **2 / 2** | - Thu thập RSS tự động: [news/rss.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/news/rss.go#L48-L100) & [cmd/news-ingest/main.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/cmd/news-ingest/main.go#L22-L80)<br>- Chuẩn hóa cấu trúc: Bóc tách `title`, `content`, `source`, `published_at`, `related_coins`, `url` lưu vào bảng `news_items` (Migration `0011`) |
| **10** | **Sentiment Analysis** | **1** | **1 / 1** | - AI Service: Python FastAPI (`sentiment-service/app/main.py` & [model.py](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/sentiment-service/app/model.py#L11-L42)) phân loại `POSITIVE`, `NEGATIVE`, `NEUTRAL` kèm score tin cậy<br>- Tích hợp Strategy: [SentimentStrategy](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/sentiment.go#L99-L133) nhận điểm cảm xúc làm bộ lọc xu hướng thị trường |
| **TỔNG** | **TỔNG ĐIỂM NHÓM B** | **40** | **40 / 40** | **ĐẠT ĐIỂM TUYỆT ĐỐI (100%)** |

---

## 2. CÂU TRẢ LỜI GIẢI TRÌNH ĐỐI SOÁT CHỨC NĂNG (FUNCTIONAL DEFENSE PROOFS)

### Tiêu chí 1: Thu thập & Hiển thị Dữ liệu Thị trường (Market Data & Candlestick)
> **Giải trình:**  
> "Hệ thống xây dựng hoàn chỉnh cả 2 kênh dữ liệu:
> 1. **Kênh Lịch sử (REST Batch)**: Sử dụng entrypoint [cmd/backfill/main.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/cmd/backfill/main.go#L35-L80) kết nối Binance REST API `/api/v3/klines`, hỗ trợ đa cặp tiền (`BTCUSDT`, `ETHUSDT`, `SOLUSDT`...) và đa khung thời gian (`5m, 15m, 1h, 4h, 1d`), nạp vào bảng PostgreSQL `candles` với khóa chính `(symbol, timeframe, open_time)` đảm bảo tính toàn vẹn và không trùng lặp.
> 2. **Kênh Thời gian thực (WebSocket Live Stream)**: Module [binance.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/market/binance.go#L243-L286) duy trì kết nối WebSocket gộp, đẩy nến trực tiếp về Go Hub. Phía Frontend, [TradingChart.tsx](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/frontend/src/features/market/components/TradingChart.tsx) và [ChartCard.tsx](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/frontend/src/features/market/components/ChartCard.tsx#L40-L100) render đầy đủ nến Candlestick OHLC kèm biểu đồ khối lượng Volume, cập nhật mượt mà theo từng tick giá."

---

### Tiêu chí 2: Động cơ Chiến lược Đơn vị (Strategy Engine $\ge$ 4 Strategies)
> **Giải trình:**  
> "Hệ thống đã phát triển **7 chiến lược giao dịch độc lập** (vượt yêu cầu tối thiểu $\ge 4$):
> 1. **Moving Average Crossover** ([ma.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/ma.go)): Giao cắt đường MA nhanh và MA chậm.
> 2. **RSI** ([rsi.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/rsi.go)): Xác định vùng Quá mua (Overbought) / Quá bán (Oversold).
> 3. **Bollinger Bands** ([bollinger.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/bollinger.go)): Bắt biến động dải Upper/Lower Band.
> 4. **Support & Resistance** ([sr.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/sr.go)): Giao dịch tại các mức Hỗ trợ / Kháng cự kỹ thuật.
> 5. **Smart Money Concepts - SMC** ([smc.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/smc.go)): Phân tích cấu trúc thị trường Order Block.
> 6. **MACD** ([macd.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/macd.go)): Giao cắt đường tín hiệu MACD & Histogram.
> 7. **Sentiment Strategy** ([sentiment.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/sentiment.go)): Tín hiệu dựa trên điểm tin tức tâm lý thị trường.  
> 
> Toàn bộ 7 chiến lược đều chuẩn hóa đầu ra qua enum [Signal](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/strategy.go#L10-L16) (`BUY`, `SELL`, `HOLD`), đảm bảo tính nhất quán tuyệt đối khi tích hợp."

---

### Tiêu chí 3: Chiến lược Kết hợp Tổ hợp (Composite Strategy)
> **Giải trình:**  
> "Hệ thống hỗ trợ cơ chế tổ hợp linh hoạt tại [combination.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/combination.go#L9-L56) với 2 giải thuật phối hợp:
> 1. **Majority Policy**: Tổng hợp biểu quyết đa số; nếu số phiếu `BUY` > `SELL` phát lệnh Mua, ngược lại phát lệnh Bán, hòa phiếu phát `HOLD`.
> 2. **Weighted Policy**: Gán trọng số $W_i$ riêng cho từng chiến lược thành phần, tính tổng điểm đại số $\sum (S_i \times W_i)$.
> 3. **Giao diện trực quan**: Component [CompositeStrategyBuilder.tsx](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/frontend/src/features/strategy/components/CompositeStrategyBuilder.tsx#L24-L80) cho phép người dùng chọn các indicator, kéo thanh trượt điều chỉnh trọng số độc lập và đóng gói thành mảng `instances[]` (theo Schema Evolution Migration 0008) gửi về Backend chạy thử nghiệm."

---

### Tiêu chí 4: Động cơ Kiểm thử Quá khứ (Backtesting Engine & Order Simulation)
> **Giải trình:**  
> "Engine [backtester.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/experiment/backtester.go#L24-L79) được thiết kế đáp ứng các tiêu chuẩn tài chính định lượng nghiêm ngặt:
> 1. **Chống Lookahead Bias**: Sử dụng kỹ thuật cắt cửa sổ `candles[i-Window : i]` (dòng 52), chỉ truyền dữ liệu quá khứ cho chiến lược phân tích, không để lộ nến hiện tại $i$.
> 2. **Mô phỏng thực tế**: Áp dụng mô hình trượt giá (`SlippageBps`), khấu trừ phí giao dịch (`FeePct`), và tự động kích hoạt cắt lỗ / chốt lời (`checkStopTarget` qua `StopLossPct` & `TakeProfitPct`).
> 3. **Lưu vết chi tiết (Trade Log)**: Mọi lệnh khớp được lưu chi tiết vào bảng `experiment_trades` (Migration `0010`) và phục vụ qua endpoint `GET /experiments/{id}/trades`, ghi nhận đầy đủ Entry Price, Exit Price, Slippage, Fee và Net Profit."

---

### Tiêu chí 5: Đánh giá Hiệu suất Định lượng (Strategy Evaluation)
> **Giải trình:**  
> "Module [evaluator.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/experiment/evaluator.go#L15-L46) tính toán chính xác toàn bộ các chỉ số đo lường tài chính bắt buộc:
> - **Lợi nhuận ròng (Return %)**: Tỷ lệ tăng trưởng vốn so với `StartingCapital`.
> - **Tỷ lệ thắng (Win Rate %)**: Tỷ lệ phần trăm giữa số lệnh có lãi (`Wins`) trên tổng số lệnh (`TradeCount`).
> - **Sụt giảm tài sản tối đa (Max Drawdown - MDD %)**: Đo lường mức sụt giảm sâu nhất từ đỉnh vốn cao nhất (`Peak Equity`) trong suốt kỳ backtest.
> - **Tổng lợi nhuận danh nghĩa (Total Profit $)**: Tổng số tiền lãi/lỗ ròng thực tế."

---

### Tiêu chí 6: Tìm kiếm Chiến lược Tự động & Điều kiện Dừng (Search & Stop Conditions)
> **Giải trình:**  
> "Hệ thống cài đặt quy trình Random Search tự động và mạnh mẽ:
> 1. **Phát sinh & Khử trùng lặp**: [random_generator.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/random_generator.go#L27-L65) sinh ngẫu nhiên tổ hợp tham số và trọng số; [loop.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/experiment/loop.go#L107-L114) dùng hàm băm `candidateKey` để deduplicate loại bỏ ứng viên trùng lặp trước khi đẩy vào Queue.
> 2. **3 Điều kiện dừng (OR Semantics)** (dòng 85-104 trong `loop.go`):
>    - `maxCandidates`: Dừng khi đã sinh đủ số lượng ứng viên yêu cầu.
>    - `maxDuration`: Dừng khi vượt quá giới hạn thời gian (Timeout).
>    - `noImprovementLimit`: Dừng sớm nếu qua $N$ ứng viên liên tiếp mà không cải thiện được kỷ lục điểm số Top-1 (Early Stopping tối ưu tài nguyên)."

---

### Tiêu chí 7: Bảng Xếp hạng Thời gian thực (Leaderboard Top-K)
> **Giải trình:**  
> "Bảng xếp hạng được tổ chức chuyên nghiệp:
> 1. **Giao diện & API**: [ExperimentLeaderboard.tsx](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/frontend/src/features/experiment/components/ExperimentLeaderboard.tsx) kết nối endpoint `GET /experiments`.
> 2. **Đồng bộ Thời gian thực**: Khi Worker Pool hoàn thành 1 backtest, hệ thống tự động broadcast event `LEADERBOARD_UPDATE` qua WebSocket Hub; Bảng xếp hạng trên UI lập tức tự refresh và tái xếp hạng mà người dùng không cần F5 trang.
> 3. **Sắp xếp & Lọc**: Hỗ trợ hiển thị Top-K và sắp xếp đa tiêu chí theo Điểm tổ hợp (Overall Composite Score), Lợi nhuận (Return %) hoặc Tỷ lệ thắng (Win Rate %)."

---

### Tiêu chí 8: Trực quan hóa Giao dịch trên Biểu đồ (Trade Visualization)
> **Giải trình:**  
> "Giao diện cung cấp trải nghiệm trực quan hóa sâu sắc:
> 1. **Vẽ Marker Khớp lệnh**: Component [BacktestChart.tsx](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/frontend/src/features/experiment/components/BacktestChart.tsx#L38-L45) chuyển đổi lịch sử trade thành các marker trực quan (Mũi tên Xanh ▲ vào vị thế BUY, Mũi tên Đỏ ▼ vào vị thế SELL).
> 2. **Tương tác Đồng bộ Bảng - Biểu đồ**: Tại [TradeHistoryTable.tsx](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/frontend/src/features/experiment/components/TradeHistoryTable.tsx#L45-L65), khi người dùng click vào bất kỳ dòng lệnh nào trong bảng lịch sử, biểu đồ nến sẽ tự động highlight và tập trung vào đúng tọa độ nến vào lệnh (Entry) và thoát lệnh (Exit) của giao dịch đó."

---

### Tiêu chí 9: Thu thập Tin tức Thị trường (News Collector)
> **Giải trình:**  
> "Đường ống thu thập tin tức vận hành tự động và chuẩn mực:
> 1. **RSS Ingestion Engine**: [news/rss.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/news/rss.go#L48-L100) và CLI tool [cmd/news-ingest/main.go](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/cmd/news-ingest/main.go#L22-L80) tự động cào tin từ các nguồn RSS uy tín (CoinDesk, CoinTelegraph...).
> 2. **Chuẩn hóa Dữ liệu**: Pipeline bóc tách đầy đủ các trường: `title`, `content` (nội dung/tóm tắt), `source`, `published_at` (epoch timestamp), `related_coins` (`JSONB`, vd: `["BTC", "ETH"]`), và `url`, lưu trữ bền vững vào bảng PostgreSQL `news_items` (Migration `0011`)."

---

### Tiêu chí 10: Phân tích Cảm xúc AI & Tích hợp Chiến lược (Sentiment Analysis)
> **Giải trình:**  
> "Kiến trúc phân tích cảm xúc được tích hợp trơn tru:
> 1. **Microservice AI**: Service Python FastAPI (`sentiment-service/app/main.py` & [model.py](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/sentiment-service/app/model.py#L11-L42)) phân tích văn bản tin tức và trả về nhãn `POSITIVE`, `NEGATIVE`, `NEUTRAL` kèm confidence score ($0.0 \dots 1.0$).
> 2. **Tích hợp Vào Lệnh**: Kết quả phân tích được nạp vào [SentimentStrategy](file:///d:/projects/Crypto%20Strategy%20Lab/Crypto-Strategy-Lab/backend/internal/strategy/sentiment.go#L99-L133). Chiến lược sử dụng điểm số này làm bộ lọc: chỉ mở vị thế Mua khi điểm cảm xúc vượt ngưỡng tích cực ($> \text{threshold}$) và Bán khi tâm lý thị trường tiêu cực ($< 1 - \text{threshold}$), kết hợp hoàn hảo giữa chỉ báo kỹ thuật và phân tích cơ bản."
