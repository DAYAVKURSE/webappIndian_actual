import { useEffect, useState, useRef } from 'react';
import { crashPlace, crashCashout, crashGetHistory } from '@/requests';
import styles from "./Crash.module.scss";
import { API_BASE_URL } from '@/config';
const initData = window.Telegram.WebApp.initData;
import toast from 'react-hot-toast';
import useStore from '@/store';

export const Crash = () => {
    const { BalanceRupee, increaseBalanceRupee, decreaseBalanceRupee } = useStore();
    const [betAmount, setBetAmount] = useState(100);
    const [bet, setBet] = useState(0);
    const [isBettingClosed, setIsBettingClosed] = useState(false);
    const [autoOutputCoefficient, setAutoOutputCoefficient] = useState(0);
    const [xValue, setXValue] = useState(1.2);
    const [collapsed, setCollapsed] = useState(false);
    const [overlayText, setOverlayText] = useState('Game starts soon');
    const [dimensions, setDimensions] = useState({ width: 0, height: 0 });
    const crashRef = useRef(null);

    // Получение истории игр при загрузке компонента
    useEffect(() => {
        const fetchHistory = async () => {
            try {
                const data = await crashGetHistory();
                if (data && data.results) {
                    const lastResult = data.results[0];
                    if (lastResult) {
                        setXValue(parseFloat(lastResult.CrashPointMultiplier.toFixed(2)));
                    }
                }
            } catch (error) {
                console.error('Error fetching game history:', error);
            }
        };

        fetchHistory();
    }, []);

    // Настройка измерений и WebSocket соединения
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
                setCollapsed(true);
                setBet(0);
                setTimeout(() => {
                    setXValue(1.2);
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
                toast.success(`Выиграли ₹${data.win_amount.toFixed(0)}!`);
            }
        };

        return () => {
            window.removeEventListener('resize', updateDimensions);
            ws.close();
        };
    }, [increaseBalanceRupee]);

    // Обработка ставки
    const handleBet = async () => {
        try {
            const response = await crashPlace(betAmount, autoOutputCoefficient);
            const data = await response.json();

            if (response.status === 200) {
                setBet(betAmount);
                decreaseBalanceRupee(betAmount);
                toast.success('Ставка размещена');
            } else {
                toast.error(data.error || 'Ошибка размещения ставки');
            }
        } catch (err) {
            console.error(err.message);
            toast.error('Не удалось разместить ставку');
        }
    }

    // Обработка вывода средств
    const handleCashout = async () => {
        try {
            const response = await crashCashout();
            const data = await response.json();

            if (response.status === 200) {
                toast.success(`Вы выиграли ₹${(bet * data.multiplier).toFixed(0)}`);
            } else {
                toast.error(data.error || 'Ошибка при выводе средств');
            }
        } catch (err) {
            console.error(err.message);
            toast.error('Не удалось вывести средства');
        }
    }

    // Изменение суммы ставки
    const handleAmountChange = (delta) => {
        setBetAmount(prevAmount => {
            const newAmount = prevAmount + delta;
            return newAmount > 0 ? newAmount : prevAmount;
        });
    };

    // Удвоение или деление суммы ставки пополам
    const handleMultiplyAmount = (factor) => {
        setBetAmount(prevAmount => {
            const newAmount = Math.round(prevAmount * factor);
            return newAmount > 0 ? newAmount : prevAmount;
        });
    };

    return (
        <div className={styles.crash}>
            {/* Баланс пользователя */}
            <div className={styles.balance}>
                <div className={styles.balanceIcon}>₹</div>
                <div className={styles.balanceValue}>{Math.floor(BalanceRupee || 0)}</div>
            </div>

            {/* Основной экран игры */}
            <div className={styles.crash_wrapper} ref={crashRef}>
                <div className={`${styles.crash__collapsed} ${collapsed ? styles.fadeIn : styles.fadeOut}`}>
                    <p>{overlayText}</p>
                </div>
                
                {/* Анимация звезды */}
                <div className={styles.starContainer}>
                    <img src="/star.svg" alt="Star" className={styles.star} />
                </div>
                
                {/* Отображение коэффициента */}
                <div className={styles.multiplier}>
                    {xValue} x
                </div>
            </div>

            {/* Раздел управления ставкой */}
            <div className={styles.betSection}>
                <div className={styles.coefficientLabel}>Coefficient</div>

                <div className={styles.betControls}>
                    <div className={styles.betAmount}>
                        <span>{betAmount} ₹</span>
                        <div className={styles.betAmountButtons}>
                            <button className={styles.betButton} onClick={() => handleAmountChange(-100)}>-</button>
                            <button className={styles.betButton} onClick={() => handleAmountChange(100)}>+</button>
                        </div>
                    </div>

                    <div className={styles.quickButtons}>
                        <button className={styles.quickButton} onClick={() => handleMultiplyAmount(0.5)}>/2</button>
                        <button className={styles.quickButton} onClick={() => handleMultiplyAmount(2)}>x2</button>
                    </div>

                    {bet > 0 ? (
                        <button 
                            className={styles.mainButton} 
                            onClick={handleCashout} 
                            disabled={!isBettingClosed}
                        >
                            Cash Out
                        </button>
                    ) : (
                        <button 
                            className={styles.mainButton} 
                            onClick={handleBet} 
                            disabled={isBettingClosed}
                        >
                            Bet
                        </button>
                    )}
                </div>
            </div>
        </div>
    );
};
