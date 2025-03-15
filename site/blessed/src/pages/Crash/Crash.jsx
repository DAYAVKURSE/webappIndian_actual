import { useEffect, useState, useRef } from 'react';
import { crashPlace, crashCashout, crashGetHistory } from '@/requests';
import styles from "./Crash.module.scss";
import { API_BASE_URL } from '@/config';
const initData = window.Telegram?.WebApp?.initData || '';
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
    const [loading, setLoading] = useState(false);
    const [isCrashed, setIsCrashed] = useState(false);
    const [isAutoEnabled, setIsAutoEnabled] = useState(false);
    const [gameActive, setGameActive] = useState(false);

    const [starPosition, setStarPosition] = useState({ x: 0, y: 0 });
    const [isFalling, setIsFalling] = useState(false);

    
    const wsRef = useRef(null);
    const multiplierTimerRef = useRef(null);
    const [startMultiplierTime, setStartMultiplierTime] = useState(null);

    console.log(dimensions)
    // Getting game history on component load
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

    // Function to simulate multiplier growth on frontend
    const simulateMultiplierGrowth = (startTime, initialMultiplier = 1.0) => {
        if (multiplierTimerRef.current) {
            clearInterval(multiplierTimerRef.current);
        }
    
        setXValue(initialMultiplier);
        
        const updateInterval = 100; 
        const growthFactor = 0.03; 
    
        let lastValue = initialMultiplier;
        
        multiplierTimerRef.current = setInterval(() => {
            const elapsedSeconds = (Date.now() - startTime) / 1000;
            const newMultiplier = Math.exp(elapsedSeconds * growthFactor);
    
            // 📌 Экспоненциальное усреднение
            const smoothedMultiplier = (lastValue * 0.8 + newMultiplier * 0.2).toFixed(2);
            lastValue = smoothedMultiplier;
            
            setXValue(parseFloat(smoothedMultiplier));
        }, updateInterval);
    };
    

    // Setting up dimensions and WebSocket connection
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

        // Checking for initData before creating WebSocket connection
        if (!initData) {
            toast.error('Authorization error. Please restart the application.');
            return;
        }

        const encoded_init_data = encodeURIComponent(initData);
        const ws = new WebSocket(`wss://${API_BASE_URL}/ws/crashgame/live?init_data=${encoded_init_data}`);
        wsRef.current = ws;

        ws.onopen = () => {
            console.log('WebSocket connection established');
        };

        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            toast.error('Connection error. Please reload the page.');
        };

        ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                console.log('WebSocket data received:', data);
                
                if (data.type === "multiplier_update") {
                    // Updating game state
                    setIsBettingClosed(true);
                    setIsCrashed(false);
                    setGameActive(true);
                    setCollapsed(false);

                    setStarPosition({
                        x: data.multiplier * 50,  // Чем больше множитель, тем дальше вправо
                        y: -data.multiplier * 40, // Чем больше множитель, тем выше
                    });
                    
                    // If this is the first multiplier update, start simulation
                    if (!startMultiplierTime) {
                        setStartMultiplierTime(Date.now());
                        simulateMultiplierGrowth(Date.now(), parseFloat(data.multiplier));
                    }
                    
                    // Automatic cashout when reaching the specified multiplier
                    if (isAutoEnabled && bet > 0 && parseFloat(data.multiplier) >= autoOutputCoefficient && autoOutputCoefficient > 0) {
                        handleCashout();
                        toast.success(`Auto cashout at ${data.multiplier}x`);
                    }
                }

                if (data.type === "game_crash") {
                    // Stop multiplier growth simulation
                    if (multiplierTimerRef.current) {
                        clearInterval(multiplierTimerRef.current);
                        multiplierTimerRef.current = null;
                    }
                    setStartMultiplierTime(null);
                    
                    setIsCrashed(true);
                    setGameActive(false);
                    setOverlayText(`Crashed at ${data.crash_point.toFixed(2)}x`);
                    setCollapsed(true);
                    setXValue(parseFloat(data.crash_point).toFixed(2));

                    setIsFalling(true);
                    setStarPosition(prev => ({ x: prev.x - 100, y: prev.y + 200 })); // Опускаем звезду вниз
                
                    
                    setTimeout(() => {
                        if (bet > 0) {
                            // If the player had an active bet, show a loss message
                            toast.error(`Game crashed at ${data.crash_point.toFixed(2)}x! You lost ₹${bet}.`);
                            setBet(0);
                        }
                        setXValue(1.2);
                    }, 3000);
                }

                if (data.type === "timer_tick") {
                    setCollapsed(true);
                    if (data.remaining_time > 10) {
                        setIsBettingClosed(true);
                        setGameActive(false);
                    } else {
                        setIsBettingClosed(false);
                        setIsCrashed(false);
                        setGameActive(false);
                    }

                    if (data.remaining_time <= 10) {
                        setOverlayText(`Game starts in ${data.remaining_time} seconds`);
                    }
                }

                if (data.type === "cashout_result") {
                    // Don't reset bet here to show the player they won
                    toast.success(`You won ₹${data.win_amount.toFixed(0)}! (${data.cashout_multiplier}x)`);
                    
                    // Delay resetting the bet to give the user time to see the result
                    setTimeout(() => {
                        setBet(0);
                        increaseBalanceRupee(data.win_amount);
                    }, 2000);
                }

                // Processing another player's cashout message
                if (data.type === "other_cashout") {
                    toast.success(`${data.username} won ₹${data.win_amount.toFixed(0)} at ${data.cashout_multiplier}x!`);
                }

                // Processing another player's bet message
                if (data.type === "new_bet") {
                    toast.success(`${data.username} bet ₹${data.amount.toFixed(0)}`);
                }
                
                // Displaying active game start
                if (data.type === "game_started") {
                    toast.success('Game started!');
                    setIsBettingClosed(true);
                    setIsCrashed(false);
                    setGameActive(true);
                    setCollapsed(false);
                    
                    // Start multiplier growth simulation with initial value of 1.0
                    setStartMultiplierTime(Date.now());
                    simulateMultiplierGrowth(Date.now(), 1.0);
                }
            } catch (error) {
                console.error('Error processing WebSocket message:', error);
            }
        };

        return () => {
            window.removeEventListener('resize', updateDimensions);
            if (multiplierTimerRef.current) {
                clearInterval(multiplierTimerRef.current);
            }
            if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
                ws.close();
            }
        };
    }, [increaseBalanceRupee, bet, autoOutputCoefficient, isAutoEnabled]);

    // Handling bet
    const handleBet = async () => {
        if (!initData) {
            toast.error('Authorization error. Please restart the application.');
            return;
        }

        if (betAmount <= 0) {
            toast.error('Bet amount must be greater than 0');
            return;
        }

        if (betAmount > BalanceRupee) {
            toast.error('Insufficient funds');
            return;
        }

        try {
            setLoading(true);
            const response = await crashPlace(betAmount, autoOutputCoefficient);
            
            if (response.ok) {
                const data = await response.json();
                console.log('Server response to bet:', data);
                setBet(betAmount);
                decreaseBalanceRupee(betAmount);
                toast.success('Bet placed! Waiting for game to start');
                
                // Animation for feedback
                setCollapsed(true);
                setOverlayText('Your bet has been accepted! Waiting for game...');
                setTimeout(() => {
                    setCollapsed(false);
                }, 2000);
            } else {
                const errorData = await response.json().catch(() => ({ error: 'An error occurred' }));
                console.error('Bet error:', errorData);
                toast.error(errorData.error || 'Error placing bet');
            }
        } catch (err) {
            console.error('Exception during bet:', err.message);
            toast.error('Failed to place bet');
        } finally {
            setLoading(false);
        }
    }

    // Handling cashout
    const handleCashout = async () => {
        if (!initData) {
            toast.error('Authorization error. Please restart the application.');
            return;
        }

        if (bet <= 0) {
            toast.error('No active bet');
            return;
        }

        if (isCrashed) {
            toast.error('Game already finished');
            return;
        }

        try {
            setLoading(true);
            const response = await crashCashout();
            
            if (response.ok) {
                const data = await response.json();
                console.log('Server response to cashout:', data);
                // Don't reset bet here as it will happen when cashout_result is received via WebSocket
                toast.success(`Cashout request sent at multiplier ${xValue}x`);
            } else {
                const errorData = await response.json().catch(() => ({ error: 'An error occurred' }));
                console.error('Cashout error:', errorData);
                toast.error(errorData.error || 'Error cashing out');
            }
        } catch (err) {
            console.error('Exception during cashout:', err.message);
            toast.error('Failed to cash out');
        } finally {
            setLoading(false);
        }
    }

    // Toggling auto-cashout
    const toggleAutoCashout = () => {
        setIsAutoEnabled(!isAutoEnabled);
        if (!isAutoEnabled) {
            toast.success(`Auto-cashout enabled at ${autoOutputCoefficient}x`);
        } else {
            toast.success("Auto-cashout disabled");
        }
    };

    // Changing auto-cashout coefficient
    const handleCoefficientChange = (e) => {
        const value = parseFloat(e.target.value);
        if (!isNaN(value) && value >= 1) {
            setAutoOutputCoefficient(value);
        }
    };

    // Changing bet amount
    const handleAmountChange = (delta) => {
        setBetAmount(prevAmount => {
            const newAmount = prevAmount + delta;
            return newAmount > 0 ? newAmount : prevAmount;
        });
    };

    // Doubling or halving bet amount
    const handleMultiplyAmount = (factor) => {
        setBetAmount(prevAmount => {
            const newAmount = Math.round(prevAmount * factor);
            return newAmount > 0 ? newAmount : prevAmount;
        });
    };

    return (
        <div className={styles.crash}>
            {/* User balance */}
            <div className={styles.balance}>
                <div className={styles.balanceIcon}>₹</div>
                <div className={styles.balanceValue}>{Math.floor(BalanceRupee || 0)}</div>
            </div>

            {/* Main game screen */}
            <div className={styles.crash_wrapper} ref={crashRef}>
                <div className={`${styles.crash__collapsed} ${collapsed ? styles.fadeIn : styles.fadeOut}`}>
                    <p>{overlayText}</p>
                </div>
                
                {/* Star animation */}
                <div 
                className={`${styles.star} ${isFalling ? styles.falling : ''}`} 
                style={{
                    transform: `translate(${starPosition.x}px, ${starPosition.y}px)`,
                }}
            >
                <img src="/star.svg" alt="Star" />
            </div>
                
                {/* Multiplier display */}
                <div className={styles.multiplier}>
                    {`${xValue}`.slice(0, 3)} x
                </div>
                
                {bet > 0 && !isCrashed && <div className={styles.activeBet}>
                    Your bet: ₹{bet}
                </div>}
            </div>

            {/* Bet control section */}
            <div className={styles.betSection}>
                <div className={styles.coefficientContainer}>
                    <div className={styles.coefficientLabel}>
                        Coefficient
                        <button 
                            className={`${styles.autoCashoutBtn} ${isAutoEnabled ? styles.active : ''}`} 
                            onClick={toggleAutoCashout}
                        >
                            Auto {isAutoEnabled ? 'ON' : 'OFF'}
                        </button>
                    </div>
                    
                    <div className={styles.coefficientInput}>
                        <input 
                            type="number" 
                            min="1.0" 
                            step="0.1"
                            value={autoOutputCoefficient} 
                            onChange={handleCoefficientChange}
                            className={styles.autoInput}
                            placeholder="Example: 2.0"
                        />
                        <span className={styles.inputLabel}>x</span>
                    </div>
                </div>

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
                            className={`${styles.mainButton} ${(gameActive && !isCrashed) ? styles.activeButton : ''}`} 
                            onClick={handleCashout} 
                            disabled={!gameActive || loading || isCrashed}
                        >
                            {loading ? 'Loading...' : 'Cash Out'}
                        </button>
                    ) : (
                        <button 
                            className={styles.mainButton} 
                            onClick={handleBet} 
                            disabled={isBettingClosed || loading}
                        >
                            {loading ? 'Loading...' : 'Bet'}
                        </button>
                    )}
                </div>
            </div>
        </div>
    );
};