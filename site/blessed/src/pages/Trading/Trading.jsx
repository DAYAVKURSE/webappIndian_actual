import { useEffect, useState, useRef } from 'react';
import { createChart } from 'lightweight-charts';
import styles from './Trading.module.scss';
import { placeBet, getOutcome, getBPC, getMe } from '@/requests';
import useStore from "@/store";
import toast from "react-hot-toast";
import { Amount, ActionButtons } from '@/components';
import { API_BASE_URL } from '@/config';

const initData = window.Telegram.WebApp.initData;

export const Trading = () => {
    const { BalanceRupee, setBalanceRupee } = useStore();
    const [bet, setBet] = useState(100);
    const [time, setTime] = useState(10);
    const [outcome, setOutcome] = useState({ latestBets: [] });

    const lineSeriesRef = useRef(null);
    const [markers, setMarkers] = useState([]);

    useEffect(() => {
        getMe();
        getBPC();
        getOutcome().then((data) => {
            setOutcome({
                ...data,
                latestBets: data.latestBets.slice(0, 5)
            });
        });
    }, []);

    const formatTime = (seconds) => {
        const minutes = String(Math.floor((seconds % 3600) / 60)).padStart(2, "0");
        const secs = String(seconds % 60).padStart(2, "0");
        return `${minutes}:${secs}`;
    };

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

        const areaSeries = chart.addAreaSeries({
            lastValueVisible: false,
            crosshairMarkerVisible: false,
            lineColor: 'transparent',
            topColor: 'rgba(63, 127, 251, 1)',
            bottomColor: 'rgba(63, 127, 251, 0.1)',
        });

        const lineSeries = chart.addLineSeries();
        lineSeriesRef.current = lineSeries;

        const encoded_init_data = encodeURIComponent(initData);
        const ws_url_initial = `wss://${API_BASE_URL}/ws/kline?init_data=${encoded_init_data}`;
        const ws_url_latest = `wss://${API_BASE_URL}/ws/kline/latest?init_data=${encoded_init_data}`;

        const ws_initial = new WebSocket(ws_url_initial);
        ws_initial.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                const last50Data = data.slice(-100).map(item => ({
                    time: item.openTime / 1000,
                    value: item.close,
                }));
                lineSeries.setData(last50Data);
                areaSeries.setData(last50Data);
            } catch (error) {
                console.error('Error parsing initial WebSocket message:', error);
            } finally {
                ws_initial.close();
            }
        };

        const ws_latest = new WebSocket(ws_url_latest);
        ws_latest.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                lineSeries.update({
                    time: data.openTime / 1000,
                    value: data.close,
                });
                areaSeries.update({
                    time: data.openTime / 1000,
                    value: data.close,
                });

                setMarkers((prevMarkers) =>
                    prevMarkers.map((marker) =>
                        marker.time === data.openTime / 1000 && marker.value === null
                            ? { ...marker, value: data.close }
                            : marker
                    )
                );
            } catch (error) {
                console.error('Error parsing live update WebSocket message:', error);
            }
        };

        const watermark_Ebaniy = document.getElementById('tv-attr-logo');
        if (watermark_Ebaniy) {
            watermark_Ebaniy.style.display = 'none';
        }

        return () => {
            chart.remove();
            ws_initial.close();
            ws_latest.close();
        };
    }, []);

    const handleBet = async (direction) => {
        try {
            const response = await placeBet(bet, time, direction);
            const data = await response.clone().json();

            if (response.status === 200) {
                toast.success('Bet placed successfully');
                setBalanceRupee(BalanceRupee - bet);

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

                lineSeriesRef.current.setMarkers(newMarkers);
                setMarkers(newMarkers);

                setOutcome((prevOutcome) => ({
                    ...prevOutcome,
                    latestBets: [{ outcome: '' }, ...prevOutcome.latestBets].slice(0, 5),
                }));

                const interval = setInterval(() => {
                    getOutcome().then((data) => {
                        setOutcome({
                            ...data,
                            latestBets: data.latestBets.slice(0, 5),
                        });
                        setBalanceRupee(data.userBalance);
                        if (data.latestBets[0].outcome === "win") {
                            toast.success(`You won! (+₹ ${data.latestBets[0].payout})`);
                        } else {
                            toast.error(`You lost! (-₹ ${bet})`);
                        }
                    });
                    clearInterval(interval);
                }, time * 1000);
            } else {
                toast.error(data.error || 'Failed to place bet');
            }
        } catch (error) {
            toast.error('Error placing bet');
            console.error('Error placing bet:', error);
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
