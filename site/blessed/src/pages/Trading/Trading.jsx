блбimport { useEffect, useState, useRef, useCallback } from 'react';
import { createChart } from 'lightweight-charts';
import styles from './Trading.module.scss';
import { placeBet, getOutcome, getBPC, getMe } from '@/requests';
import useStore from "@/store";
import toast from "react-hot-toast";
import { Amount, ActionButtons } from '@/components';
import { API_BASE_URL, WS_PROTOCOL, API_PROTOCOL } from '@/config';

export const Trading = () => {
    const { BalanceRupee, setBalanceRupee } = useStore();
    const [bet, setBet] = useState(100);
    const [time, setTime] = useState(10);
    const [outcome, setOutcome] = useState({ latestBets: [] });
    const [chartReady, setChartReady] = useState(false);
    const [currentPrice, setCurrentPrice] = useState(null);
    const [priceChange, setPriceChange] = useState(0);

    const candleSeriesRef = useRef(null);
    const chartRef = useRef(null);
    const [markers, setMarkers] = useState([]);
    const wsLatestRef = useRef(null);
    const lastPriceRef = useRef(null);

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
        if (!chartRef.current || !candleSeriesRef.current) return;

        console.log('Настраиваем WebSocket соединения');
        const encoded_init_data = encodeURIComponent(initData.current);
        const ws_url_initial = `${WS_PROTOCOL}://${API_BASE_URL}/ws/kline?init_data=${encoded_init_data}`;
        const ws_url_latest = `${WS_PROTOCOL}://${API_BASE_URL}/ws/kline/latest?init_data=${encoded_init_data}`;

        console.log('URL для исторических данных:', ws_url_initial);
        console.log('URL для обновлений:', ws_url_latest);

        // Получение исторических данных
        const ws_initial = new WebSocket(ws_url_initial);
        ws_initial.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                if (Array.isArray(data) && data.length > 0) {
                    const last100Data = data.slice(-100).map(item => ({
                        time: item.openTime / 1000,
                        open: item.open,
                        high: item.high,
                        low: item.low,
                        close: item.close,
                    }));
                    
                    candleSeriesRef.current.setData(last100Data);
                    setChartReady(true);
                }
            } catch (error) {
                console.error('Ошибка при обработке исходных данных WebSocket:', error);
            } finally {
                ws_initial.close();
            }
        };

        const connectInitialWs = () => {
            try {
                const ws_initial = new WebSocket(ws_url_initial);
                
                ws_initial.onopen = () => {
                    console.log('WebSocket соединение для исторических данных установлено');
                    reconnectAttempts = 0; // Сбросить счетчик попыток
                };
                
                ws_initial.onmessage = (event) => {
                    try {
                        const data = JSON.parse(event.data);
                        console.log('Получены исторические данные, количество элементов:', data?.length);
                        
                        if (Array.isArray(data) && data.length > 0) {
                            const last100Data = data.slice(-100).map(item => ({
                                time: item.openTime / 1000,
                                open: item.open,
                                high: item.high,
                                low: item.low,
                                close: item.close
                            }));
                            
                            if (candleSeriesRef.current) {
                                candleSeriesRef.current.setData(last100Data);
                                setChartReady(true);
                                console.log('Данные установлены в график, chartReady =', true);
                            } else {
                                console.error('candleSeriesRef.current отсутствует при попытке установить данные');
                            }
                        } else {
                            console.warn('Получен пустой массив данных или некорректные данные');
                        }
                    } catch (error) {
                        console.error('Ошибка при обработке исходных данных WebSocket:', error);
                    } finally {
                        ws_initial.close();
                    }
                };

                ws_initial.onerror = (error) => {
                    console.error('Ошибка WebSocket для исходных данных:', error);
                    ws_initial.close();
                    
                    // Повторная попытка подключения с увеличением задержки
                    if (reconnectAttempts < maxReconnectAttempts) {
                        reconnectAttempts++;
                        const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000);
                        console.log(`Попытка повторного подключения ${reconnectAttempts} из ${maxReconnectAttempts} через ${delay}мс`);
                        setTimeout(connectInitialWs, delay);
                    } else {
                        toast.error('Не удалось подключиться к серверу. Пожалуйста, обновите страницу.');
                    }
                };
                
                ws_initial.onclose = (event) => {
                    if (!event.wasClean) {
                        console.warn('WebSocket для исходных данных был закрыт некорректно', event);
                    }
                };
                
            } catch (error) {
                console.error('Ошибка при создании WebSocket соединения:', error);
                
                // Повторная попытка подключения с увеличением задержки
                if (reconnectAttempts < maxReconnectAttempts) {
                    reconnectAttempts++;
                    const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000);
                    console.log(`Попытка повторного подключения ${reconnectAttempts} из ${maxReconnectAttempts} через ${delay}мс`);
                    setTimeout(connectInitialWs, delay);
                } else {
                    toast.error('Не удалось подключиться к серверу. Пожалуйста, обновите страницу.');
                }
            }
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
                        
                        if (time && candleSeriesRef.current && chartReady) {
                            // Обновляем свечной график
                            candleSeriesRef.current.update({
                                time,
                                open: data.open,
                                high: data.high,
                                low: data.low,
                                close: data.close
                            });
                            
                            // Обновляем текущую цену и изменение
                            setCurrentPrice(data.close);
                            if (lastPriceRef.current !== null) {
                                const change = (data.close - lastPriceRef.current) / lastPriceRef.current * 100;
                                setPriceChange(change);
                            }
                            lastPriceRef.current = data.close;

                            // Обновляем маркеры ставок
                            setMarkers((prevMarkers) =>
                                prevMarkers.map((marker) =>
                                    marker.time === time && marker.value === null
                                        ? { ...marker, value: data.close }
                                        : marker
                                )
                            );
                        }
                    } 
                }
                catch (error) {
                    console.error('Ошибка при обработке обновления WebSocket:', error);
                }
                
                ws_latest.onerror = (error) => {
                    console.error('Ошибка WebSocket для обновлений:', error);
                    if (wsLatestRef.current === ws_latest) {
                        wsLatestRef.current = null;
                    }
                    ws_latest.close();
                    setTimeout(connectLatestWs, 2000); // Повторное подключение через 2 секунды
                };
                
                ws_latest.onclose = (event) => {
                    console.log('WebSocket для обновлений закрыт:', event.wasClean ? 'корректно' : 'некорректно');
                    if (wsLatestRef.current === ws_latest) {
                        wsLatestRef.current = null;
                    }
                    setTimeout(connectLatestWs, 2000); // Повторное подключение через 2 секунды
                };
            } 
        };
        
        // Сначала подключаемся к исходным данным, затем к обновлениям
        connectInitialWs();
        
        // Задержка перед подключением к потоку обновлений, чтобы убедиться,
        // что исторические данные загружены
        setTimeout(connectLatestWs, 1000);
        
        return () => {
            if (wsLatestRef.current) {
                console.log('Закрываем WebSocket соединение при размонтировании компонента');
                wsLatestRef.current.close();
                wsLatestRef.current = null;
            }
        };
    }, [chartReady]);

    // Инициализация графика
    useEffect(() => {
        const chart = createChart('chart', {
            layout: {
                background: { color: '#000000' },
                textColor: '#3F7FFB',
                fontSize: 12,
            },
            grid: {
                vertLines: { color: 'rgba(63, 127, 251, 0.1)' },
                horzLines: { color: 'rgba(63, 127, 251, 0.1)' },
            },
            rightPriceScale: {
                borderColor: 'rgba(63, 127, 251, 0.2)',
                scaleMargins: {
                    top: 0.1,
                    bottom: 0.1,
                },
                visible: true,
            },
            timeScale: {
                borderColor: 'rgba(63, 127, 251, 0.2)',
                visible: true,
                timeVisible: true,
                secondsVisible: false,
                tickMarkFormatter: (time) => {
                    const date = new Date(time * 1000);
                    const hours = date.getHours().toString().padStart(2, '0');
                    const minutes = date.getMinutes().toString().padStart(2, '0');
                    return `${hours}:${minutes}`;
                },
            },
            crosshair: {
                mode: 1,
                vertLine: {
                    color: 'rgba(63, 127, 251, 0.5)',
                    width: 1,
                    style: 1,
                    visible: true,
                    labelVisible: true,
                },
                horzLine: {
                    color: 'rgba(63, 127, 251, 0.5)',
                    width: 1,
                    style: 1,
                    visible: true,
                    labelVisible: true,
                },
            },
            localization: {
                timeFormatter: (time) => {
                    const date = new Date(time * 1000);
                    const hours = date.getHours().toString().padStart(2, '0');
                    const minutes = date.getMinutes().toString().padStart(2, '0');
                    const seconds = date.getSeconds().toString().padStart(2, '0');
                    return `${hours}:${minutes}:${seconds}`;
                },
            },
        });

        // Создаем свечной график вместо линейного
        const candleSeries = chart.addCandlestickSeries({
            upColor: 'rgba(0, 150, 136, 0.8)',
            downColor: 'rgba(255, 82, 82, 0.8)',
            borderVisible: false,
            wickUpColor: 'rgba(0, 150, 136, 0.8)',
            wickDownColor: 'rgba(255, 82, 82, 0.8)',
            priceFormat: {
                type: 'price',
                precision: 2,
                minMove: 0.01,
            },
        });
        
        candleSeriesRef.current = candleSeries;
        chartRef.current = chart;

        // Скрываем водяной знак
        const watermark_Ebaniy = document.getElementById('tv-attr-logo');
        if (watermark_Ebaniy) {
            watermark_Ebaniy.style.display = 'none';
        }
        
        try {
            const chart = createChart('chart', {
                layout: {
                    background: { color: '#000000' },
                    textColor: '#3F7FFB',
                    fontSize: 12,
                },
                grid: {
                    vertLines: { color: 'rgba(63, 127, 251, 0.1)' },
                    horzLines: { color: 'rgba(63, 127, 251, 0.1)' },
                },
                rightPriceScale: {
                    borderColor: 'rgba(63, 127, 251, 0.2)',
                    scaleMargins: {
                        top: 0.1,
                        bottom: 0.1,
                    },
                    visible: true,
                },
                timeScale: {
                    borderColor: 'rgba(63, 127, 251, 0.2)',
                    visible: true,
                    timeVisible: true,
                    secondsVisible: false,
                    tickMarkFormatter: (time) => {
                        const date = new Date(time * 1000);
                        const hours = date.getHours().toString().padStart(2, '0');
                        const minutes = date.getMinutes().toString().padStart(2, '0');
                        return `${hours}:${minutes}`;
                    },
                },
                crosshair: {
                    mode: 1,
                    vertLine: {
                        color: 'rgba(63, 127, 251, 0.5)',
                        width: 1,
                        style: 1,
                        visible: true,
                        labelVisible: true,
                    },
                    horzLine: {
                        color: 'rgba(63, 127, 251, 0.5)',
                        width: 1,
                        style: 1,
                        visible: true,
                        labelVisible: true,
                    },
                },
                localization: {
                    timeFormatter: (time) => {
                        const date = new Date(time * 1000);
                        const hours = date.getHours().toString().padStart(2, '0');
                        const minutes = date.getMinutes().toString().padStart(2, '0');
                        const seconds = date.getSeconds().toString().padStart(2, '0');
                        return `${hours}:${minutes}:${seconds}`;
                    },
                },
                // Добавляем обработку обновления размера
                handleScale: true,
                handleScroll: true,
            });

            // Создаем свечной график вместо линейного
            const candleSeries = chart.addCandlestickSeries({
                upColor: 'rgba(0, 150, 136, 0.8)',
                downColor: 'rgba(255, 82, 82, 0.8)',
                borderVisible: false,
                wickUpColor: 'rgba(0, 150, 136, 0.8)',
                wickDownColor: 'rgba(255, 82, 82, 0.8)',
                priceFormat: {
                    type: 'price',
                    precision: 2,
                    minMove: 0.01,
                },
            });
            
            candleSeriesRef.current = candleSeries;
            chartRef.current = chart;
            console.log('График инициализирован успешно');

            // Адаптация размера при изменении размеров окна
            const resizeHandler = () => {
                if (chartRef.current) {
                    chartRef.current.resize(
                        chartContainer.clientWidth,
                        chartContainer.clientHeight
                    );
                }
            };
            window.addEventListener('resize', resizeHandler);

            // Скрываем водяной знак
            const watermark_Ebaniy = document.getElementById('tv-attr-logo');
            if (watermark_Ebaniy) {
                watermark_Ebaniy.style.display = 'none';
            }

            return () => {
                window.removeEventListener('resize', resizeHandler);
                chart.remove();
                if (wsLatestRef.current) {
                    wsLatestRef.current.close();
                }
                chartRef.current = null;
                candleSeriesRef.current = null;
            };
        } catch (error) {
            console.error('Ошибка при создании графика:', error);
            toast.error('Не удалось инициализировать график');
            return () => {};
        }
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
                    },
                ];

                // Обновить маркеры на графике
                if (candleSeriesRef.current) {
                    const allMarkers = [...markers, ...newMarkers];
                    candleSeriesRef.current.setMarkers(allMarkers);
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
            console.error('Ошибка при размещении ставки:', error);
            toast.error('Не удалось разместить ставку');
        }
    };

    // Форматирование баланса для отображения
    const formatBalance = (balance) => {
        if (balance === undefined || balance === null) return { main: '0', suppl: '00' };
        
        const balanceStr = balance.toFixed(2);
        const [main, suppl] = balanceStr.split('.');
        
        return { main, suppl: suppl || '00' };
    };

    // Форматирование цены для отображения
    const formatPrice = (price) => {
        if (price === null || price === undefined) return "Loading...";
        return price.toFixed(2);
    };

    // Форматирование изменения цены
    const formatPriceChange = (change) => {
        if (change === null || change === undefined) return "";
        const sign = change >= 0 ? "+" : "";
        return `${sign}${change.toFixed(2)}%`;
    };

    const { main, suppl } = formatBalance(BalanceRupee);
    
    return (
        <div className={styles.trading}>
            <h1 className={styles.title}>Trading</h1>
            <p className={styles.trading_text}>Your balance</p>
            <h3 className={styles.trading_balance}>
                ₹ {BalanceRupee ? BalanceRupee.toFixed(0) : 0}
            </h3>
            
            <div className={styles.chartWrap}>
                <div id="chart" className={styles.chart} />
            </div>
            
            <div className={styles.trading__bet}>
                <button className={styles.trading__bet_button} onClick={() => handleBet("down")}>
                    <img src="/24=arrow_circle_down.svg" alt="Down" />
                </button>
                <button className={styles.trading__bet_button} onClick={() => handleBet("up")}>
                    <img src="/24=arrow_circle_up.svg" alt="Up" />
                </button>
            </div>

            <div className={styles.trading__timer}>
                <button className={styles.minusBtn} onClick={() => setTime((prevTime) => Math.max(prevTime - 10, 10))}>
                    −
                </button>
                <p className={styles.trading__timer_text}>00:10</p>
                <button className={styles.plusBtn} onClick={() => setTime((prevTime) => Math.min(prevTime + 10, 3540))}>
                    +
                </button>
            </div>
            
            <div className={styles.trading_button_container}>
                <button className={styles.trading_button} onClick={() => setTime(() => 10)}>10 <span>sec</span></button>
                <button className={styles.trading_button} onClick={() => setTime(() => 30)}>30 <span>sec</span></button>
                <button className={styles.trading_button} onClick={() => setTime(() => 60)}>1 <span>min</span></button>
                <button className={styles.trading_button} onClick={() => setTime(() => 300)}>5 <span>min</span></button>
            </div>
            
            <div className={styles.betAmount}>
                <div className={styles.amountDisplay}>
                    <span>100 ₹</span>
                    <div className={styles.amountControls}>
                        <button onClick={() => setBet(prev => Math.max(prev - 10, 10))}>−</button>
                        <button onClick={() => setBet(prev => prev + 10)}>+</button>
                    </div>
                </div>
                <div className={styles.amountButtons}>
                    <button onClick={() => setBet(prev => Math.floor(prev / 2))}>/ 2</button>
                    <button onClick={() => setBet(prev => prev * 2)}>× 2</button>
                </div>
            </div>
        </div>
    );
};
