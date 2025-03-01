import { useEffect, useState, useRef, useCallback } from 'react';
import { createChart } from 'lightweight-charts';
import styles from './Trading.module.scss';
import { placeBet, getOutcome, getBPC, getMe } from '@/requests';
import useStore from "@/store";
import toast from "react-hot-toast";
import { Amount, ActionButtons } from '@/components';
import { API_BASE_URL } from '@/config';

export const Trading = () => {
    const { BalanceRupee, setBalanceRupee } = useStore();
    const [bet, setBet] = useState(100);
    const [time, setTime] = useState(10);
    const [outcome, setOutcome] = useState({ latestBets: [] });
    const [chartReady, setChartReady] = useState(false);

    const lineSeriesRef = useRef(null);
    const chartRef = useRef(null);
    const [markers, setMarkers] = useState([]);
    const wsLatestRef = useRef(null);

    const initData = useRef(window.Telegram.WebApp.initData);

    // Загрузка начальных данных
    useEffect(() => {
        const loadInitialData = async () => {
            try {
                await getMe();
                await getBPC();
                const outcomeData = await getOutcome();
                if (outcomeData) {
                    setOutcome({
                        ...outcomeData,
                        latestBets: outcomeData.latestBets.slice(0, 5)
                    });
                }
            } catch (error) {
                console.error('Ошибка при загрузке начальных данных:', error);
                toast.error('Не удалось загрузить данные');
            }
        };

        loadInitialData();
    }, []);

    const formatTime = (seconds) => {
        const minutes = String(Math.floor((seconds % 3600) / 60)).padStart(2, "0");
        const secs = String(seconds % 60).padStart(2, "0");
        return `${minutes}:${secs}`;
    };

    // Установка и обработка WebSocket соединений
    const setupWebSockets = useCallback(() => {
        if (!chartRef.current || !lineSeriesRef.current) return;

        const areaSeries = chartRef.current.addAreaSeries({
            lastValueVisible: false,
            crosshairMarkerVisible: false,
            lineColor: 'transparent',
            topColor: 'rgba(63, 127, 251, 1)',
            bottomColor: 'rgba(63, 127, 251, 0.1)',
        });

        const encoded_init_data = encodeURIComponent(initData.current);
        const ws_url_initial = `wss://${API_BASE_URL}/ws/kline?init_data=${encoded_init_data}`;
        const ws_url_latest = `wss://${API_BASE_URL}/ws/kline/latest?init_data=${encoded_init_data}`;

        // Получение исторических данных
        const ws_initial = new WebSocket(ws_url_initial);
        ws_initial.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                if (Array.isArray(data) && data.length > 0) {
                    const last100Data = data.slice(-100).map(item => ({
                        time: item.openTime / 1000,
                        value: item.close,
                    }));
                    lineSeriesRef.current.setData(last100Data);
                    areaSeries.setData(last100Data);
                    setChartReady(true);
                }
            } catch (error) {
                console.error('Ошибка при обработке исходных данных WebSocket:', error);
            } finally {
                ws_initial.close();
            }
        };

        ws_initial.onerror = (error) => {
            console.error('Ошибка WebSocket для исходных данных:', error);
            toast.error('Ошибка соединения с сервером');
            ws_initial.close();
        };

        // Обработка обновлений в реальном времени
        const connectLatestWs = () => {
            const ws_latest = new WebSocket(ws_url_latest);
            wsLatestRef.current = ws_latest;
            
            ws_latest.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    if (data && typeof data === 'object') {
                        const time = data.openTime / 1000;
                        const value = data.close;
                        
                        if (time && value && lineSeriesRef.current && chartReady) {
                            lineSeriesRef.current.update({
                                time,
                                value,
                            });
                            areaSeries.update({
                                time,
                                value,
                            });

                            setMarkers((prevMarkers) =>
                                prevMarkers.map((marker) =>
                                    marker.time === time && marker.value === null
                                        ? { ...marker, value }
                                        : marker
                                )
                            );
                        }
                    }
                } catch (error) {
                    console.error('Ошибка при обработке обновления WebSocket:', error);
                }
            };
            
            ws_latest.onerror = (error) => {
                console.error('Ошибка WebSocket для обновлений:', error);
                ws_latest.close();
                setTimeout(connectLatestWs, 2000); // Повторное подключение через 2 секунды
            };
            
            ws_latest.onclose = () => {
                setTimeout(connectLatestWs, 2000); // Повторное подключение через 2 секунды
            };
        };
        
        connectLatestWs();
        
        return () => {
            if (wsLatestRef.current) {
                wsLatestRef.current.close();
            }
        };
    }, [chartReady]);

    // Инициализация графика
    useEffect(() => {
        const chart = createChart('chart', {
            layout: {
                background: { color: '#000000' },
                textColor: '#3F7FFB',
            },
            grid: {
                vertLines: { color: 'rgba(63, 127, 251, 0.1)' },
                horzLines: { color: 'rgba(63, 127, 251, 0.1)' },
            },
            rightPriceScale: {
                borderColor: 'rgba(63, 127, 251, 0.2)',
            },
            timeScale: {
                borderColor: 'rgba(63, 127, 251, 0.2)',
                visible: false,
            },
        });

        const lineSeries = chart.addLineSeries();
        lineSeriesRef.current = lineSeries;
        chartRef.current = chart;

        const watermark_Ebaniy = document.getElementById('tv-attr-logo');
        if (watermark_Ebaniy) {
            watermark_Ebaniy.style.display = 'none';
        }

        return () => {
            chart.remove();
            if (wsLatestRef.current) {
                wsLatestRef.current.close();
            }
        };
    }, []);

    // Установка WebSocket соединений после инициализации графика
    useEffect(() => {
        const cleanup = setupWebSockets();
        return cleanup;
    }, [setupWebSockets]);

    // Периодическое обновление баланса и исходов
    useEffect(() => {
        const updateInterval = setInterval(async () => {
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
            } catch (error) {
                console.error('Ошибка при обновлении данных:', error);
            }
        }, 30000); // Обновлять каждые 30 секунд
        
        return () => clearInterval(updateInterval);
    }, [setBalanceRupee]);

    const handleBet = async (direction) => {
        try {
            const response = await placeBet(bet, time, direction);
            
            if (!response || response.status !== 200) {
                const errorData = await response?.json().catch(() => ({}));
                toast.error(errorData?.error || 'Не удалось разместить ставку');
                return;
            }
            
            const data = await response.json().catch(() => ({}));
            
            if (data) {
                toast.success('Ставка успешно размещена');
                setBalanceRupee(Math.max(0, BalanceRupee - bet));

                const betTime = Math.floor(Date.now() / 1000);
                const endTime = betTime + time;
                const shape = direction === 'up' ? 'arrowUp' : 'arrowDown';

                const newMarkers = [
                    {
                        time: betTime,
                        position: 'aboveBar',
                        color: '#ffd689',
                        shape: shape,
                        text: `₹ ${bet}`,
                        value: null,
                    },
                    {
                        time: endTime,
                        position: 'aboveBar',
                        color: '#ffd689',
                        shape: shape,
                        text: `₹ ${bet}`,
                        value: null,
                    }
                ];

                // Обновить маркеры на графике
                if (lineSeriesRef.current) {
                    const allMarkers = [...markers, ...newMarkers];
                    lineSeriesRef.current.setMarkers(allMarkers);
                    setMarkers(allMarkers);
                }

                // Обновить последние ставки
                setOutcome((prevOutcome) => ({
                    ...prevOutcome,
                    latestBets: [{ outcome: '', amount: bet, direction }, ...prevOutcome.latestBets].slice(0, 5),
                }));

                // Получить результат через установленное время
                const checkInterval = setInterval(async () => {
                    try {
                        const outcomeData = await getOutcome();
                        if (outcomeData) {
                            setOutcome({
                                ...outcomeData,
                                latestBets: outcomeData.latestBets.slice(0, 5),
                            });
                            
                            // Обновить баланс
                            if (outcomeData.userBalance !== undefined) {
                                setBalanceRupee(outcomeData.userBalance);
                            }
                            
                            // Показать уведомление о результате ставки
                            const latestBet = outcomeData.latestBets[0];
                            if (latestBet && latestBet.outcome) {
                                if (latestBet.outcome === "win") {
                                    toast.success(`Вы выиграли! (+₹ ${latestBet.payout})`);
                                } else if (latestBet.outcome === "lose") {
                                    toast.error(`Вы проиграли! (-₹ ${bet})`);
                                }
                            }
                        }
                    } catch (error) {
                        console.error('Ошибка при получении результата:', error);
                    } finally {
                        clearInterval(checkInterval);
                    }
                }, (time * 1000) + 1000); // Добавляем 1 секунду для завершения обработки на сервере
            }
        } catch (error) {
            toast.error('Ошибка при размещении ставки');
            console.error('Ошибка при размещении ставки:', error);
        }
    };


    const formatBalance = (balance) => {
        let balanceStr = Math.trunc(balance).toString();

        const balanceParts = balanceStr.replace(/\B(?=(\d{3})+(?!\d))/g, ".").split(".");

        const main = balanceParts.shift();
        const fraction = balanceParts.join('.');

        return { main, fraction };
    };

    const { main, fraction } = formatBalance(BalanceRupee || 0);

    return (
        <div className={styles.trading}>
            {/* <h3 className={styles.trading_balance}>₹ {Math.trunc(BalanceRupee)}</h3> */}
            <h3 className={styles.trading_balance}>
                <span className={styles.main}>₹ {main}</span>
                {fraction && (
                    <span className={styles.clicker__balance__value_fraction} style={{ fontSize: "26px" }}>
                        .{fraction}
                    </span>
                )}
            </h3>
            <p className={styles.trading_text}>Your balance</p>
            <div id="chart" className={styles.trading__chart} />
            <ActionButtons
                onclick1={() => handleBet("down")} src1="/24=arrow_circle_down.svg" label1="Down" color1="#D22C32"
                onclick2={() => handleBet("up")} src2="/24=arrow_circle_up.svg" label2="Up" color2="#26CC57"
            />
            <div className={styles.trading__timer}>
                <div className={styles.trading__timer_button_container} onClick={() => setTime((prevTime) => Math.max(prevTime - 60, 10))}>
                    <img src="/24=remove.svg" alt="" className={styles.trading__timer_button} />
                </div>
                <p className={styles.trading__timer_text}>{formatTime(time)}</p>
                <div className={styles.trading__timer_button_container} onClick={() => setTime((prevTime) => Math.min(prevTime + 60, 3540))}>
                    <img src="/24=add.svg" alt="" className={styles.trading__timer_button} />
                </div>
            </div>
            <div className={styles.trading_button_container}>
                <button className={styles.trading_button} onClick={() => setTime(() => 10)}>10 <p>sec</p></button>
                <button className={styles.trading_button} onClick={() => setTime(() => 30)}>30 <p>sec</p></button>
                <button className={styles.trading_button} onClick={() => setTime(() => 60)}>1 <p>min</p></button>
                <button className={styles.trading_button} onClick={() => setTime(() => 300)}>5 <p>min</p></button>
            </div>
            <Amount bet={bet} setBet={setBet} />
        </div>
    );
};
