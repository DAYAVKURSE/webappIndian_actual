import { useEffect, useState, useRef } from 'react';
import { Amount, Input, Button } from '@/components';
import { crashPlace, crashCashout, crashGetHistory } from '@/requests';
import styles from "./Crash.module.scss";
import { API_BASE_URL } from '@/config';
const initData = window.Telegram.WebApp.initData;
import toast from 'react-hot-toast';
import useStore from '@/store';

export const Crash = () => {
    const { increaseBalanceRupee, decreaseBalanceRupee } = useStore();
    const [betAmount, setBetAmount] = useState(100);
    const [bet, setBet] = useState(0);
    const [bets, setBets] = useState([]);
    const [isBettingClosed, setIsBettingClosed] = useState(false);
    const [autoOutputCoefficient, setAutoOutputCoefficient] = useState(0);
    const [wins, setWins] = useState([]);
    const [xValue, setXValue] = useState(1);
    const [collapsed, setCollapsed] = useState(false);
    const [overlayText, setOverlayText] = useState('Game starts soon');
    const [dimensions, setDimensions] = useState({ width: 0, height: 0 });
    const crashRef = useRef(null);


    useEffect(() => {
        const fetchHistory = async () => {
            try {
                const data = await crashGetHistory();
                const lastWins = data.results.map(result => result.CrashPointMultiplier.toFixed(2));
                setWins(lastWins.reverse());
            } catch (error) {
                console.error('Error fetching game history:', error);
            }
        };

        fetchHistory();
    }, []);

    useEffect(() => {
        const updateDimensions = () => {
            if (crashRef.current) {
                setDimensions({
                    width: crashRef.current.offsetWidth,
                    height: crashRef.current.offsetHeight,
                });
            }
        };

        updateDimensions();
        window.addEventListener('resize', updateDimensions);

        const encoded_init_data = encodeURIComponent(initData);
        const ws = new WebSocket(`wss://${API_BASE_URL}/ws/crashgame/live?init_data=${encoded_init_data}`);

        ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            if (data.type === "multiplier_update") {
                setIsBettingClosed(true);
                const multiplier = data.multiplier.toFixed(2);
                setXValue(multiplier);
                setCollapsed(false);
            }

            if (data.type === "game_crash") {
                setOverlayText(`Crashed at ${data.crash_point.toFixed(2)}x`);
                setWins((prevWins) => [...prevWins, data.crash_point.toFixed(2)]);
                setCollapsed(true);
                setBet(0);
                setBets((prevBets) =>
                    prevBets.map((bet) =>
                        bet.status === 'pending' ? { ...bet, status: 'lost' } : bet
                    )
                );
                setTimeout(() => {
                    setBets([]);
                    setXValue(1);
                }, 5000);
            }

            if (data.type === "timer_tick") {
                setCollapsed(true);
                if (data.remaining_time > 10) {
                    setIsBettingClosed(true);
                } else {
                    setIsBettingClosed(false);
                }

                if (data.remaining_time <= 10) {
                    setOverlayText(`Game starts in ${data.remaining_time} seconds`);
                }
            }

            if (data.type === "cashout_result") {
                setBet(0);
                increaseBalanceRupee(data.win_amount);
            }

            if (data.type === "user_cashout") {
                setBets((prevBets) =>
                    prevBets.map((bet) =>
                        bet.username === data.username ? { ...bet, status: 'won', multiplier: data.multiplier } : bet
                    )
                );
            }

            if (data.type === "new_bet") {
                const newBet = {
                    username: data.username,
                    amount: data.amount,
                    auto_cashout_multiplier: data.auto_cashout_multiplier,
                    is_benefit_bet: data.is_benefit_bet,
                    status: 'pending', // 'pending', 'won', 'lost'
                    multiplier: 0,
                };
                setBets((prevBets) => [...prevBets, newBet]);
            }
        };

        return () => {
            window.removeEventListener('resize', updateDimensions);
            ws.close();
        };
    }, [increaseBalanceRupee]);

    const { width, height } = dimensions;

    if (width === 0 || height === 0) {
        return <div className={styles.crash_wrapper} ref={crashRef}></div>;
    }

    const progress = Math.min(xValue / 5.0, 2);

    const arrowTipX = progress * (width / 2);
    const arrowTipY = height - progress * (height / 2);

    const handleBet = async () => {
        try {
            const response = await crashPlace(betAmount, autoOutputCoefficient);
            const data = await response.json();

            if (response.status === 200) {
                setBet(betAmount);
                decreaseBalanceRupee(betAmount);
                toast.success('Bet placed');
            } else {
                toast.error(data.error);
            }
        } catch (err) {
            console.error(err.message);
        }
    }

    const handleCashout = async () => {
        try {
            const response = await crashCashout();
            const data = await response.json();

            if (response.status === 200) {
                toast.success(`You won ₹${(bet * data.multiplier).toFixed()}`);
            } else {
                toast.error(data.error);
            }
        } catch (err) {
            console.error(err.message);
        }
    }

    const getBackgroundColor = (win) => {
        if (win > 5) return '#CFA3F2';
        if (win > 2.5) return '#FFC397';
        if (win > 1.5) return '#FFD689';
        return '#EDFF8C';
    };

    const Crash = () => {
        return (
            <div className={styles.crash_wrapper} ref={crashRef}>
                <svg width={width} height={height} className={styles.svg}>
                    <line
                        x1="0"
                        y1={height}
                        x2={arrowTipX}
                        y2={arrowTipY}
                        stroke="#CFA3F2"
                        strokeWidth="2"
                    />
                </svg>
                <div
                    className={styles.circle}
                    style={{
                        left: `${arrowTipX}px`,
                        top: `${arrowTipY}px`,
                    }}
                />
                <div
                    className={styles.tip}
                    style={{
                        left: `${arrowTipX + 45}px`,
                        top: `${arrowTipY - 35}px`,
                    }}
                >
                    <img className={styles.tip_arrow} src="/crash_arrow.svg" alt="" />
                    <div className={styles.tip_xvalue}>
                        {xValue}x
                    </div>
                    <div className={styles.tip_bet}>
                        {(bet * xValue).toFixed()} ₹
                    </div>
                </div>
            </div>
        )
    }

    return (
        <div className={styles.crash}>

            <div className={styles.crash_wrapper}>
                <div className={`${styles.crash__collapsed} ${collapsed ? styles.fadeIn : styles.fadeOut}`}>
                    <p>{overlayText}</p>
                </div>
                <Crash />
            </div>
            {wins && wins.length > 0 && (
                <div className={styles.crash_wins}>
                    <h3 className={styles.crash_title}>Last games</h3>
                    <div className={styles.crash_wins_list}>
                        {wins.slice().reverse().map((win, index) => (
                            <div
                                key={index}
                                className={styles.crash_wins_item}
                                style={{ backgroundColor: getBackgroundColor(win) }}
                            >
                                <p className={styles.crash_wins_item_amount}>
                                    {win}x
                                </p>
                            </div>
                        ))}
                    </div>
                </div>
            )}
            <h3 className={styles.crash_title}>Bet</h3>
            <Amount bet={betAmount} setBet={setBetAmount} />
            <div className={styles.crash_container}>
                <h3 className={styles.crash_title}>Auto output</h3>
                <Input onChange={(e) => setAutoOutputCoefficient(e.target.value)} center placeholder="Coefficient" type="number" min="0" />
                {bet > 0 ?
                    <Button label={`Withdraw ₹ ${(bet * xValue).toFixed()}`} color="#FFD689" fill onClick={handleCashout} disabled={!isBettingClosed} />
                    : <Button label="Place a bet" color="#FFD689" fill onClick={handleBet} disabled={isBettingClosed} />
                }
            </div>
            <div className={styles.crash_container}>
                <div className={styles.crash_bets}>
                    <h3 className={styles.crash_title}>Room rates</h3>
                    <div className={styles.crash_bets_list}>
                        {bets.length > 0 ? bets.map((bet, index) => (
                            <div
                                key={index}
                                className={styles.crash_bets_item}
                            >
                                <p className={styles.crash_bets_item_username}>
                                    {bet.username}</p>
                                <div
                                    className={`${styles.crash_bets_item_wrapper} ${bet.status === 'won'
                                            ? styles.won
                                            : bet.status === 'lost'
                                                ? styles.lost
                                                : ''
                                        }`}>
                                    <p className={`${styles.crash_bets_item_amount} ${styles.noFilter}`}>
                                        <img className={styles.noFilter} src="/24=rupee.svg" alt="" />
                                        {bet.amount}
                                    </p>
                                    {bet.status !== 'pending' && (
                                        <p className={styles.crash_bets_item_amount}>
                                            {bet.status !== 'pending' && (<img src="/24=rupee.svg" alt="" />)}
                                            {bet.status === 'won' ? `${(bet.amount * bet.multiplier).toFixed()}` : bet.status === 'lost' ? `${bet.amount}` : ``}
                                        </p>
                                    )}
                                    {bet.status === 'won' && (
                                        <p
                                            className={`${styles.crash_bets_item_multiplier} ${styles.noFilter}`}
                                            style={{ color: getBackgroundColor(bet.multiplier) }}
                                        >{bet.multiplier.toFixed(2)}x</p>
                                    )}
                                </div>
                            </div>
                        )) : <p>No bets yet</p>}
                    </div>
                </div>

            </div>
        </div>
    );
};
