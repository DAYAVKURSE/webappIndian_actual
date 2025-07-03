import { useEffect, useState, useRef, useCallback } from 'react';
import { createChart } from 'lightweight-charts';
import styles from './Trading.module.scss';
import { placeBet, getOutcome, getBPC, getMe } from '@/requests';
import useStore from '@/store';
import toast from 'react-hot-toast';
import { getAccessToken } from '../../utils/token-storage';
import { MoneyGameStatus } from '@/components';

export const Trading = () => {
  const { BalanceRupee, setBalanceRupee } = useStore();
  const [bet, setBet] = useState(100);
  const [time, setTime] = useState(10);
  const [outcome, setOutcome] = useState({ latestBets: [] });
  const [chartReady, setChartReady] = useState(false);
  const [currentPrice, setCurrentPrice] = useState(null);
  const [priceChange, setPriceChange] = useState(0);

  const initData = encodeURIComponent(getAccessToken() || '');

  const candleSeriesRef = useRef(null);
  const chartRef = useRef(null);
  const [markers, setMarkers] = useState([]);
  const wsLatestRef = useRef(null);
  const lastPriceRef = useRef(null);

  const fetchOutcome = async () => {
    try {
      const outcomeData = await getOutcome();
      if (outcomeData) {
        setOutcome({
          ...outcomeData,
          latestBets: outcomeData.latestBets.slice(0, 5),
        });
        if (outcomeData.userBalance !== undefined) {
          setBalanceRupee(outcomeData.userBalance);
        }
      }
    } catch (err) {
      console.error('Failed to fetch outcome:', err);
    }
  };

  useEffect(() => {
    (async () => {
      try {
        await getMe();
        await getBPC();
        await fetchOutcome();
      } catch (err) {
        console.error('Initial data error:', err);
        toast.error("Couldn't load initial data");
      }
    })();
  }, []);

  const setupWebSockets = useCallback(() => {
    const ws_initial = new WebSocket(`wss://testfakeserver.com/api/ws/kline?init_data=${initData}`);

    ws_initial.onmessage = (e) => {
      try {
        const data = JSON.parse(e.data);
        if (Array.isArray(data) && data.length > 0) {
          const formatted = data.slice(-100).map((item) => ({
            time: item.openTime / 1000,
            open: item.open,
            high: item.high,
            low: item.low,
            close: item.close,
          }));
          candleSeriesRef.current.setData(formatted);
          setChartReady(true);
        }
      } catch (err) {
        console.error('Init WS message error:', err);
      } finally {
        ws_initial.close();
      }
    };

    ws_initial.onerror = (err) => {
      console.error('WebSocket error for source data:', err);
      ws_initial.close();
    };

    const connectLiveUpdates = () => {
      const ws = new WebSocket(`wss://testfakeserver.com/api/ws/kline?init_data=${initData}`);
      wsLatestRef.current = ws;

      ws.onmessage = (e) => {
        try {
          const data = JSON.parse(e.data);
          if (!data || typeof data !== 'object') return;

          const time = data.openTime / 1000;
          if (candleSeriesRef.current && chartReady) {
            candleSeriesRef.current.update({
              time,
              open: data.open,
              high: data.high,
              low: data.low,
              close: data.close,
            });

            setCurrentPrice(data.close);
            if (lastPriceRef.current !== null) {
              const change = ((data.close - lastPriceRef.current) / lastPriceRef.current) * 100;
              setPriceChange(change);
            }
            lastPriceRef.current = data.close;

            setMarkers((prev) =>
              prev.map((marker) =>
                marker.time === time && marker.value === null ? { ...marker, value: data.close } : marker
              )
            );
          }
        } catch (err) {
          console.error('Live WS update error:', err);
        }
      };

      let reconnectTimeout;

      ws.onerror = (e) => {
        console.error('WebSocket error:', e);
        if (ws.readyState !== WebSocket.CLOSED && ws.readyState !== WebSocket.CLOSING) {
          ws.close();
        }
      };

      ws.onclose = () => {
        if (reconnectTimeout) clearTimeout(reconnectTimeout);
        reconnectTimeout = setTimeout(() => connectLiveUpdates(), 2000);
      };
    };

    connectLiveUpdates();

    return () => {
      if (wsLatestRef.current) wsLatestRef.current.close();
    };
  }, [chartReady, initData]);

  useEffect(() => {
    const chart = createChart('chart', {
      layout: {
        background: { color: '#1a1b20' },
        textColor: '#C5C8D1',
        fontSize: 12,
        fontFamily: 'Inter, sans-serif',
      },
      grid: {
        vertLines: { color: 'rgba(255, 255, 255, 0.05)' },
        horzLines: { color: 'rgba(255, 255, 255, 0.05)' },
      },
      rightPriceScale: {
        borderColor: '#1F1F1F',
        scaleMargins: { top: 0.15, bottom: 0.15 },
      },
      timeScale: {
        borderColor: '#1F1F1F',
        timeVisible: true,
        secondsVisible: true,
        tickMarkFormatter: (time) => {
          const date = new Date(time * 1000);
          return `${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`;
        },
      },
      crosshair: {
        mode: 1,
        vertLine: { color: '#4E5A6B', width: 1, visible: true, labelVisible: true },
        horzLine: { color: '#4E5A6B', width: 1, visible: true, labelVisible: true },
      },
      localization: {
        timeFormatter: (time) => {
          const date = new Date(time * 1000);
          return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}:${date
            .getSeconds()
            .toString()
            .padStart(2, '0')}`;
        },
      },
    });

    const candleSeries = chart.addCandlestickSeries({
      upColor: '#0d99ff',
      downColor: '#ef5350',
      borderUpColor: '#0d99ff',
      borderDownColor: '#ef5350',
      wickUpColor: '#0d99ff',
      wickDownColor: '#ef5350',
      borderVisible: true,
      priceFormat: {
        type: 'price',
        precision: 2,
        minMove: 0.01,
      },
    });

    // Скрываем водяной знак
    const watermark_Ebaniy = document.getElementById('tv-attr-logo');
    if (watermark_Ebaniy) {
      watermark_Ebaniy.style.display = 'none';
    }

    candleSeriesRef.current = candleSeries;
    chartRef.current = chart;

    return () => {
      chart.remove();
      if (wsLatestRef.current) wsLatestRef.current.close();
    };
  }, []);

  useEffect(() => {
    const cleanup = setupWebSockets();
    return cleanup;
  }, [setupWebSockets]);

  useEffect(() => {
    const interval = setInterval(fetchOutcome, 30000);
    return () => clearInterval(interval);
  }, []);

  const handleBet = async (direction) => {
    try {
      const response = await placeBet(bet, time, direction);
      if (!response || response.status !== 200) {
        const errorData = await response?.json().catch(() => ({}));
        toast.error(errorData?.error || "Couldn't place a bid");
        return;
      }

      toast.success('The bid was successfully placed');
      setBalanceRupee(Math.max(0, BalanceRupee - bet));

      const now = Math.floor(Date.now() / 1000);
      const shape = direction === 'up' ? 'arrowUp' : 'arrowDown';

      const newMarkers = [
        { time: now, position: 'aboveBar', color: '#ffd689', shape, text: `₹ ${bet}`, value: null },
        { time: now + time, position: 'aboveBar', color: '#ffd689', shape, text: `₹ ${bet}`, value: null },
      ];

      const allMarkers = [...markers, ...newMarkers];
      candleSeriesRef.current?.setMarkers(allMarkers);
      setMarkers(allMarkers);

      setOutcome((prev) => ({
        ...prev,
        latestBets: [{ outcome: '', amount: bet, direction }, ...prev.latestBets].slice(0, 5),
      }));

      setTimeout(fetchOutcome, time * 1000 + 1000);
    } catch (err) {
      console.error('Bet placement error:', err);
      toast.error("Couldn't place a bid");
    }
  };

  return (
    <div className={styles.trading}>
      <MoneyGameStatus />
      <h1 className={styles.title}>Trading</h1>

      <div className={styles.chartWrap}>
        <div id="chart" className={styles.chart} />
      </div>

      <div className={styles.trading__bet}>
        <button className={styles.trading__bet_button} onClick={() => handleBet('down')}>
          <img src="/trading_arrow.svg" alt="Down" />
        </button>
        <button className={styles.trading__bet_button} onClick={() => handleBet('up')}>
          <img src="/trading_arrow.svg" alt="Up" />
        </button>
      </div>

      <div className={styles.trading__timer}>
        <button className={styles.minusBtn} onClick={() => setTime((prevTime) => Math.max(prevTime - 10, 10))}>
          <img src="/trading_min.svg" alt="trading_min" />
        </button>
        <div className={styles.trading__timer_text}>00:10</div>
        <button className={styles.plusBtn} onClick={() => setTime((prevTime) => Math.min(prevTime + 10, 3540))}>
          <img src="/trading_plus.svg" alt="trading_plus" />
        </button>
      </div>

      <div className={styles.trading_button_container}>
        <button className={styles.trading_button} onClick={() => setTime(() => 10)}>
          10 sec
        </button>
        <button className={styles.trading_button} onClick={() => setTime(() => 30)}>
          30 sec
        </button>
        <button className={styles.trading_button} onClick={() => setTime(() => 60)}>
          1 min
        </button>
        <button className={styles.trading_button} onClick={() => setTime(() => 300)}>
          5 min
        </button>
      </div>

      <div className={styles.bet}>
        <div className={styles.betAmount}>
          <div className={styles.amountDisplay}>
            <input
              className={styles.amount__input}
              value={bet}
              onChange={(e) => {
                const value = e.target.value.replace(/\D/g, '');
                setBet(value);
              }}
            />
            <div className={styles.amountControls}>
              <button onClick={() => setBet((prev) => Math.max(prev - 10, 10))}>
                <img src="/trading_min.svg" alt="trading_min" />
              </button>
              <button onClick={() => setBet((prev) => prev + 10)}>
                <img src="/trading_plus.svg" alt="trading_plus" />
              </button>
            </div>
          </div>
          <div className={styles.amountButtons}>
            <button onClick={() => setBet((prev) => Math.floor(prev / 2))}>/ 2</button>
            <button onClick={() => setBet((prev) => prev * 2)}>× 2</button>
          </div>
        </div>
        <button className={styles.betButton}>
          Bet{' '}
          <svg width="24" height="24" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path
              d="M13.0942 10L8.08507 4.99167L6.90674 6.17L10.7401 10.0033L6.90674 13.8308L8.08507 15.0092L13.0942 10Z"
              fill="#FFFFFF"
            ></path>
          </svg>
        </button>
      </div>
    </div>
  );
};
