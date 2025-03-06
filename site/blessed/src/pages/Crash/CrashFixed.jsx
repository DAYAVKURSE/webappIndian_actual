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
    const [startingFlash, setStartingFlash] = useState(false);
    const [crashParticles, setCrashParticles] = useState([]);
    const [isExploding, setIsExploding] = useState(false);
    
    const wsRef = useRef(null);
    const multiplierTimerRef = useRef(null);
    const [startMultiplierTime, setStartMultiplierTime] = useState(null);

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

        multiplierTimerRef.current = setInterval(() => {
            const elapsedSeconds = (Date.now() - startTime) / 1000;
            
            // Using a simplified growth model: multiplier = e^(0.1 * time)
            const currentMultiplier = initialMultiplier * Math.pow(Math.E, 0.1 * elapsedSeconds);
            
            // Format to 2 decimal places
            setXValue(currentMultiplier.toFixed(2));
        }, 100); // Update every 100ms for smooth animation
    };

    // Connecting to WebSocket
    useEffect(() => {
        const connectWebSocket = () => {
            const ws = new WebSocket(`${API_BASE_URL.replace('http', 'ws')}/ws/crash?initData=${encodeURIComponent(initData)}`);
            wsRef.current = ws;

            ws.onopen = () => {
                console.log('WebSocket connection established');
            };

            ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                toast.error('Authorization error. Please restart the application.');
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
                        setIsExploding(false); // Сбрасываем состояние взрыва при новых обновлениях
                        
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

                    if (data.type === "game_crash" && !isExploding) { // Проверяем, не запущена ли уже анимация взрыва
                        // Stop multiplier growth simulation
                        if (multiplierTimerRef.current) {
                            clearInterval(multiplierTimerRef.current);
                            multiplierTimerRef.current = null;
                        }
                        setStartMultiplierTime(null);
                        
                        // Запускаем анимацию взрыва
                        setIsExploding(true);
                        
                        // Генерируем частицы взрыва - умеренное количество, не слишком яркие
                        const explosionParticles = [];
                        const crashPoint = parseFloat(data.crash_point);
                        const particleCount = 15 + Math.floor(crashPoint * 3);
                        
                        for (let i = 0; i < particleCount; i++) {
                            const angle = Math.random() * 360;
                            const distance = 30 + Math.random() * 100;
                            const size = 1.5 + Math.random() * 3;
                            const type = Math.random() > 0.7 ? 'gold' : Math.random() > 0.5 ? 'orange' : 'bright';
                            const delay = Math.random() * 0.2;
                            
                            explosionParticles.push({
                                id: i,
                                angle,
                                distance,
                                size,
                                type,
                                delay
                            });
                        }
                        
                        setCrashParticles(explosionParticles);
                        
                        // Показываем сообщение о крахе с небольшой задержкой
                        setTimeout(() => {
                            setIsCrashed(true);
                            setGameActive(false);
                            setOverlayText(`Crashed at ${crashPoint.toFixed(2)}x`);
                            setCollapsed(true);
                            setXValue(crashPoint.toFixed(2));
                        }, 200);
                        
                        // Очищаем состояние анимации взрыва через некоторое время
                        setTimeout(() => {
                            if (bet > 0) {
                                // If the player had an active bet, show a loss message
                                toast.error(`Game crashed at ${crashPoint.toFixed(2)}x! You lost ₹${bet}.`);
                                setBet(0);
                            }
                            setXValue(1.2);
                            setCrashParticles([]);
                            setIsExploding(false);
                        }, 2000);
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

            ws.onclose = () => {
                console.log('WebSocket connection closed');
                // Reconnect after a short delay
                setTimeout(connectWebSocket, 3000);
            };
        };

        connectWebSocket();

        // Cleanup on component unmount
        return () => {
            if (wsRef.current) {
                wsRef.current.close();
            }
            if (multiplierTimerRef.current) {
                clearInterval(multiplierTimerRef.current);
            }
        };
    }, [increaseBalanceRupee, bet, autoOutputCoefficient, isAutoEnabled]);

    // Обработка начала новой игры
    useEffect(() => {
        if (gameActive && startMultiplierTime) {
            // Добавляем эффект вспышки при старте игры
            setStartingFlash(true);
            setTimeout(() => setStartingFlash(false), 600);
        }
    }, [gameActive, startMultiplierTime]);

    // Handling bet
    const handleBet = async () => {
        if (loading || isBettingClosed) return;
        
        if (betAmount <= 0) {
            toast.error('Bet amount must be greater than 0');
            return;
        }
        
        if (betAmount > BalanceRupee) {
            toast.error('Insufficient balance');
            return;
        }
        
        setLoading(true);
        
        try {
            const response = await crashPlace(betAmount);
            
            if (response && response.id) {
                // Bet placed successfully
                setBet(betAmount);
                decreaseBalanceRupee(betAmount);
                toast.success('Bet placed! Waiting for game to start');
            } else {
                toast.error('Failed to place bet. Please try again.');
            }
        } catch (error) {
            console.error('Error placing bet:', error);
            toast.error('Error placing bet. Please try again.');
        } finally {
            setLoading(false);
        }
    };
    
    // Handling cashout
    const handleCashout = async () => {
        if (loading || !gameActive || isCrashed) return;
        
        setLoading(true);
        
        try {
            const response = await crashCashout();
            
            if (!response.success) {
                toast.error('Failed to cash out. Please try again.');
            }
        } catch (error) {
            console.error('Error cashing out:', error);
            toast.error('Authorization error. Please restart the application.');
        } finally {
            setLoading(false);
        }
    };

    // Handling window resize
    useEffect(() => {
        const updateDimensions = () => {
            if (crashRef.current) {
                const { width, height } = crashRef.current.getBoundingClientRect();
                setDimensions({ width, height });
            }
        };
        
        window.addEventListener('resize', updateDimensions);
        updateDimensions();
        
        return () => window.removeEventListener('resize', updateDimensions);
    }, []);

    // Toggle auto cashout
    const toggleAutoCashout = () => {
        setIsAutoEnabled(!isAutoEnabled);
    };
    
    // Handle coefficient input change
    const handleCoefficientChange = (e) => {
        const value = parseFloat(e.target.value);
        setAutoOutputCoefficient(isNaN(value) ? 0 : value);
    };
    
    // Handling bet amount change
    const handleAmountChange = (delta) => {
        setBetAmount(prevAmount => {
            const newAmount = prevAmount + delta;
            return newAmount > 0 ? newAmount : prevAmount;
        });
    };
    
    // Handling bet amount multiplication
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
                <div className={`${styles.starContainer} ${gameActive || isExploding ? styles.active : ''}`}>
                    {startingFlash && (
                        <div className={styles.explosionFlash} style={{ left: '50%', top: '50%' }} />
                    )}
                    <img 
                        src="/star.svg" 
                        alt="Star" 
                        className={`${styles.star} ${gameActive && !isExploding ? styles.flying : ''} ${isExploding ? styles.exploding : ''} ${startingFlash ? styles.rocketStart : ''}`} 
                        style={gameActive && !isExploding ? {
                            filter: `drop-shadow(0 0 ${Math.min(40, 10 + xValue * 3)}px rgba(255, 215, 0, ${Math.min(1, 0.6 + xValue * 0.05)}))`
                        } : {}}
                    />
                    
                    {/* Огненный след за звездой при активной игре */}
                    {gameActive && !isExploding && (
                        <div className={styles.sparkTrail} />
                    )}
                    
                    {/* Частицы взрыва */}
                    {isExploding && crashParticles.map(particle => {
                        const radians = particle.angle * (Math.PI / 180);
                        const x = Math.cos(radians) * particle.distance;
                        const y = Math.sin(radians) * particle.distance;
                        
                        return (
                            <div
                                key={`crash-${particle.id}`}
                                className={`${styles.smallParticle} ${styles.active} ${styles[particle.type + 'Particle']}`}
                                style={{
                                    left: `calc(50% + ${x}px)`,
                                    top: `calc(50% - ${y}px)`,
                                    width: `${particle.size}px`,
                                    height: `${particle.size}px`,
                                    '--x': `${x * 1.5}px`,
                                    '--y': `${-y * 1.5}px`,
                                    animationDelay: `${particle.delay}s`,
                                    opacity: '0.8' // снижаем яркость
                                }}
                            />
                        );
                    })}
                    
                    {/* Основные частицы */}
                    {gameActive && !isExploding && Array(12).fill().map((_, index) => {
                        const angle = (index * 30) * (Math.PI / 180);
                        const offsetX = Math.cos(angle) * 30;
                        const offsetY = Math.sin(angle) * 30;
                        
                        return (
                            <div 
                                key={index} 
                                className={`${styles.starParticle} ${gameActive ? styles.active : ''} ${index % 3 === 0 ? styles.type1 : index % 3 === 1 ? styles.type2 : styles.type3}`} 
                                style={{ 
                                    left: `calc(50% + ${offsetX}px)`, 
                                    top: `calc(50% - ${offsetY}px)`,
                                    '--x-end': `${offsetX * (2 + Math.min(2, xValue / 2))}px`,
                                    '--y-end': `${offsetY * (2 + Math.min(2, xValue / 2))}px`,
                                    animationDelay: `${index * 0.1}s`
                                }}
                            />
                        );
                    })}
                    
                    {/* Дополнительные искры при высоком мультипликаторе */}
                    {gameActive && !isExploding && xValue > 1.5 && Array(8).fill().map((_, index) => {
                        const angle = ((index * 45) + 20) * (Math.PI / 180);
                        const offsetX = Math.cos(angle) * 20;
                        const offsetY = Math.sin(angle) * 20;
                        
                        return (
                            <div 
                                key={`spark-${index}`} 
                                className={`${styles.starParticle} ${styles.activeFast} ${index % 2 === 0 ? styles.gold : styles.bright}`} 
                                style={{ 
                                    left: `calc(50% + ${offsetX}px)`, 
                                    top: `calc(50% - ${offsetY}px)`,
                                    '--x-end': `${offsetX * (3 + Math.min(3, xValue / 1.5))}px`,
                                    '--y-end': `${offsetY * (3 + Math.min(3, xValue / 1.5))}px`,
                                    animationDelay: `${index * 0.05 + 0.2}s`
                                }}
                            />
                        );
                    })}
                    
                    {/* Более интенсивные эффекты при очень высоком мультипликаторе */}
                    {gameActive && !isExploding && xValue > 3 && Array(6).fill().map((_, index) => {
                        const angle = ((index * 60) + 10) * (Math.PI / 180);
                        const offsetX = Math.cos(angle) * 25;
                        const offsetY = Math.sin(angle) * 25;
                        
                        return (
                            <div 
                                key={`intense-${index}`} 
                                className={`${styles.starParticle} ${styles.activeFast} ${styles.orange}`} 
                                style={{ 
                                    left: `calc(50% + ${offsetX}px)`, 
                                    top: `calc(50% - ${offsetY}px)`,
                                    '--x-end': `${offsetX * (4 + Math.min(5, xValue / 2))}px`,
                                    '--y-end': `${offsetY * (4 + Math.min(5, xValue / 2))}px`,
                                    animationDelay: `${index * 0.03}s`,
                                    transform: `scale(${Math.min(1.5, 1 + (xValue - 3) / 10)})`
                                }}
                            />
                        );
                    })}
                    
                    {gameActive && !isExploding && (
                        <div 
                            className={`${styles.glowEffect} ${gameActive ? styles.active : ''}`} 
                            style={{ 
                                left: '50%', 
                                top: '50%',
                                width: `${60 + Math.min(40, xValue * 10)}px`,
                                height: `${60 + Math.min(40, xValue * 10)}px`,
                                opacity: Math.min(0.7, 0.3 + xValue * 0.05)
                            }} 
                        />
                    )}
                </div>
                
                {/* Multiplier display */}
                <div className={styles.multiplier}>
                    {xValue} x
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