import { useEffect, useRef } from 'react';
import {
  createChart,
  CandlestickSeries,
  HistogramSeries,
  LineSeries,
  LineStyle,
  createSeriesMarkers,
  type IChartApi,
  type ISeriesApi,
  type CandlestickData,
  type HistogramData,
  type UTCTimestamp,
  type Time,
  type ISeriesMarkersPluginApi,
  type IPriceLine,
} from 'lightweight-charts';
import type { Candle } from '../../../types/candle';

export interface SRZone {
  price: number;
  type: 'SUPPORT' | 'RESISTANCE';
}

export interface ChartMarker {
  time: number; // Unix timestamp in milliseconds
  position: 'aboveBar' | 'belowBar' | 'inBar';
  color: string;
  shape: 'arrowUp' | 'arrowDown' | 'circle' | 'square';
  text: string;
}

interface TradingChartProps {
  candles: Candle[];
  ma20Line?: number[]; // Precalculated MA(20) values matching candles length
  bbands?: {
    upper: number[];
    basis: number[];
    lower: number[];
  };
  srZones?: SRZone[]; // Support & Resistance price levels
  markers?: ChartMarker[]; // BUY/SELL markers
}

export function TradingChart({
  candles,
  ma20Line,
  bbands,
  srZones,
  markers,
}: TradingChartProps) {
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null);
  const volumeSeriesRef = useRef<ISeriesApi<'Histogram'> | null>(null);
  const maSeriesRef = useRef<ISeriesApi<'Line'> | null>(null);
  const bbUpperSeriesRef = useRef<ISeriesApi<'Line'> | null>(null);
  const bbBasisSeriesRef = useRef<ISeriesApi<'Line'> | null>(null);
  const bbLowerSeriesRef = useRef<ISeriesApi<'Line'> | null>(null);
  const markersPluginRef = useRef<ISeriesMarkersPluginApi<Time> | null>(null);
  const srPriceLinesRef = useRef<IPriceLine[]>([]); // To clean up price lines

  useEffect(() => {
    if (!chartContainerRef.current) return;

    // 1. Initialize Chart
    const chart = createChart(chartContainerRef.current, {
      layout: {
        background: { color: '#ffffff' },
        textColor: '#64748b',
      },
      grid: {
        vertLines: { color: '#f1f5f9' },
        horzLines: { color: '#f1f5f9' },
      },
      rightPriceScale: {
        borderColor: '#0f172a',
        autoScale: true,
      },
      timeScale: {
        borderColor: '#0f172a',
        timeVisible: true,
        secondsVisible: false,
      },
    });

    chartRef.current = chart;

    // 2. Add Candlestick Series
    const candlestickSeries = chart.addSeries(CandlestickSeries, {
      upColor: '#10b981', // Emerald-500
      downColor: '#ef4444', // Red-500
      borderUpColor: '#10b981',
      borderDownColor: '#ef4444',
      wickUpColor: '#10b981',
      wickDownColor: '#ef4444',
    });
    candlestickSeriesRef.current = candlestickSeries;

    // 3. Add Volume Series (render below candlesticks)
    const volumeSeries = chart.addSeries(HistogramSeries, {
      color: '#3b82f6',
      priceFormat: {
        type: 'volume',
      },
      priceScaleId: 'volume', // Render on separate scale or overlay
    });

    // Configure volume overlay scale
    chart.priceScale('volume').applyOptions({
      scaleMargins: {
        top: 0.8, // Take only bottom 20% of the height
        bottom: 0,
      },
    });
    volumeSeriesRef.current = volumeSeries;

    // Initialize Markers Plugin
    const markersPlugin = createSeriesMarkers(candlestickSeries, []);
    markersPluginRef.current = markersPlugin;

    // 4. Listen to container resize
    const handleResize = () => {
      if (chartContainerRef.current && chartRef.current) {
        chartRef.current.resize(
          chartContainerRef.current.clientWidth,
          chartContainerRef.current.clientHeight
        );
      }
    };
    const resizeObserver = new ResizeObserver(handleResize);
    resizeObserver.observe(chartContainerRef.current);

    // Cleanup on unmount
    return () => {
      resizeObserver.disconnect();
      if (chartRef.current) {
        chartRef.current.remove();
        chartRef.current = null;
      }
      candlestickSeriesRef.current = null;
      volumeSeriesRef.current = null;
      maSeriesRef.current = null;
      bbUpperSeriesRef.current = null;
      bbBasisSeriesRef.current = null;
      bbLowerSeriesRef.current = null;
      markersPluginRef.current = null;
      srPriceLinesRef.current = [];
    };
  }, []);

  // Update chart data whenever props change
  useEffect(() => {
    if (
      !chartRef.current ||
      !candlestickSeriesRef.current ||
      !volumeSeriesRef.current ||
      candles.length === 0
    ) {
      return;
    }

    const chart = chartRef.current;
    const candlestickSeries = candlestickSeriesRef.current;
    const volumeSeries = volumeSeriesRef.current;

    // 1. Prepare candle & volume data (Converting time to seconds unix timestamp)
    const candleData: CandlestickData[] = [];
    const volumeData: HistogramData[] = [];

    candles.forEach((c) => {
      const timeSecs = Math.floor(c.openTime / 1000) as UTCTimestamp;
      candleData.push({
        time: timeSecs,
        open: c.open,
        high: c.high,
        low: c.low,
        close: c.close,
      });

      volumeData.push({
        time: timeSecs,
        value: c.volume,
        color: c.close >= c.open ? 'rgba(16, 185, 129, 0.3)' : 'rgba(239, 68, 68, 0.3)',
      });
    });

    candlestickSeries.setData(candleData);
    volumeSeries.setData(volumeData);

    // 2. Draw Moving Average (MA20)
    if (ma20Line && ma20Line.length === candles.length) {
      const maSeries = maSeriesRef.current || chart.addSeries(LineSeries, {
        color: '#3b82f6', // Cyan-blue
        lineWidth: 2,
        title: 'MA(20)',
      });
      maSeriesRef.current = maSeries;

      const maData = candles.map((c, i) => ({
        time: Math.floor(c.openTime / 1000) as UTCTimestamp,
        value: ma20Line[i],
      })).filter(d => d.value !== undefined && !isNaN(d.value));
      maSeries.setData(maData);
    } else if (maSeriesRef.current) {
      chart.removeSeries(maSeriesRef.current);
      maSeriesRef.current = null;
    }

    // 3. Draw Bollinger Bands
    if (
      bbands &&
      bbands.upper.length === candles.length &&
      bbands.basis.length === candles.length &&
      bbands.lower.length === candles.length
    ) {
      const bbUpperSeries = bbUpperSeriesRef.current || chart.addSeries(LineSeries, {
        color: 'rgba(168, 85, 247, 0.5)', // Transparent Purple
        lineWidth: 1,
        lineStyle: LineStyle.Dotted,
        title: 'BB Upper',
      });
      bbUpperSeriesRef.current = bbUpperSeries;

      const bbBasisSeries = bbBasisSeriesRef.current || chart.addSeries(LineSeries, {
        color: 'rgba(168, 85, 247, 0.3)',
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        title: 'BB Basis',
      });
      bbBasisSeriesRef.current = bbBasisSeries;

      const bbLowerSeries = bbLowerSeriesRef.current || chart.addSeries(LineSeries, {
        color: 'rgba(168, 85, 247, 0.5)',
        lineWidth: 1,
        lineStyle: LineStyle.Dotted,
        title: 'BB Lower',
      });
      bbLowerSeriesRef.current = bbLowerSeries;

      const mapBand = (arr: number[]) =>
        candles
          .map((c, i) => ({
            time: Math.floor(c.openTime / 1000) as UTCTimestamp,
            value: arr[i],
          }))
          .filter((d) => d.value !== undefined && !isNaN(d.value));

      bbUpperSeries.setData(mapBand(bbands.upper));
      bbBasisSeries.setData(mapBand(bbands.basis));
      bbLowerSeries.setData(mapBand(bbands.lower));
    } else {
      if (bbUpperSeriesRef.current) {
        chart.removeSeries(bbUpperSeriesRef.current);
        bbUpperSeriesRef.current = null;
      }
      if (bbBasisSeriesRef.current) {
        chart.removeSeries(bbBasisSeriesRef.current);
        bbBasisSeriesRef.current = null;
      }
      if (bbLowerSeriesRef.current) {
        chart.removeSeries(bbLowerSeriesRef.current);
        bbLowerSeriesRef.current = null;
      }
    }

    // 4. Draw Support & Resistance horizontal price lines
    // Clean up old price lines first
    srPriceLinesRef.current.forEach((line) => {
      candlestickSeries.removePriceLine(line);
    });
    srPriceLinesRef.current = [];

    if (srZones) {
      srZones.forEach((sr) => {
        const line = candlestickSeries.createPriceLine({
          price: sr.price,
          color: sr.type === 'SUPPORT' ? 'rgba(16, 185, 129, 0.4)' : 'rgba(239, 68, 68, 0.4)',
          lineWidth: 1,
          lineStyle: LineStyle.Dashed,
          axisLabelVisible: true,
          title: sr.type,
        });
        srPriceLinesRef.current.push(line);
      });
    }

    // 5. Draw Signals & Trades Markers
    if (markersPluginRef.current) {
      if (markers) {
        const formattedMarkers = markers
          .map((m) => {
            // Align marker time with the closest candle openTime
            const timeSecs = Math.floor(m.time / 1000) as UTCTimestamp;
            return {
              time: timeSecs,
              position: m.position,
              color: m.color,
              shape: m.shape,
              text: m.text,
            };
          })
          // Sort by time ascending (lightweight charts requirement)
          .sort((a, b) => a.time - b.time);

        markersPluginRef.current.setMarkers(formattedMarkers);
      } else {
        markersPluginRef.current.setMarkers([]);
      }
    }
  }, [candles, ma20Line, bbands, srZones, markers]);

  return (
    <div style={{ position: 'relative', width: '100%', height: '100%' }}>
      <div
        ref={chartContainerRef}
        style={{ width: '100%', height: '100%', minHeight: '260px' }}
      />
    </div>
  );
}
