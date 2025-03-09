import { useEffect, useState, useRef } from 'react';
import { crashPlace, crashCashout, crashGetHistory } from '@/requests';
import styles from "./Crash.module.scss";
import { API_BASE_URL, WS_PROTOCOL, API_PROTOCOL } from '@/config';
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
    const [starPosition, setStarPosition] = useState(0);
    const [starExploding, setStarExploding] = useState(false);
    const [sparkParticles, setSparkParticles] = useState([]);
    const [starAnimation, setStarAnimation] = useState('');
    const [fallingParticles, setFallingParticles] = useState([]);
    const [showWinResult, setShowWinResult] = useState(false);
    
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

    // Независимая функция для генерации падающих сверху вниз частиц
    const generateFallingParticles = () => {
        // Не генерируем новые частицы, если звезда уже взорвалась
        if (isCrashed) {
            return;
        }
        
        const newParticles = [];
        const count = Math.floor(Math.random() * 3) + 1; // 1-3 частицы за раз
        
        for (let i = 0; i < count; i++) {
            // Случайное горизонтальное смещение (-80px до +80px)
            const xOffset = Math.random() * 160 - 80;
            
            // Скорость падения (от 300 до 800 мс)
            const speed = Math.random() * 500 + 300;
            
            // Тип частицы
            const type = Math.random() < 0.5 ? 'brightFalling' : 'goldFalling';
            
            // Время создания частицы и ее ожидаемое время жизни в миллисекундах
            const creationTime = Date.now();
            const lifespan = speed; // Время жизни равно скорости анимации
            
            newParticles.push({
                id: `${creationTime}_${Math.random()}`,
                xOffset,
                speed,
                type,
                size: Math.random() * 0.7 + 0.3, // Размер от 0.3 до 1.0
                creationTime,
                lifespan
            });
        }
        
        setFallingParticles(prev => {
            // Удаляем частицы, которые должны уже закончить свою анимацию
            const currentTime = Date.now();
            const filteredPrev = prev.filter(p => {
                // Получаем время создания из ID или поля creationTime
                const creationTime = p.creationTime || parseInt(p.id.split('_')[0]);
                const lifespan = p.lifespan || p.speed;
                
                // Проверяем, прошло ли время, равное продолжительности анимации + запас 50мс
                return currentTime - creationTime < lifespan + 50;
            });
            
            return [...newParticles, ...filteredPrev].slice(0, 25);
        });
    };

    // Function to simulate multiplier growth on frontend
    const simulateMultiplierGrowth = (startTime, initialMultiplier = 1.0) => {
        if (multiplierTimerRef.current) {
            clearInterval(multiplierTimerRef.current);
        }

        // Сохраняем последнее отображаемое значение чтобы сравнивать с новым
        let lastDisplayedValue = initialMultiplier.toFixed(2);
        setXValue(lastDisplayedValue);
        
        // Сбрасываем позицию звезды и анимацию взрыва
        setStarPosition(0);
        setStarExploding(false);
        setSparkParticles([]);
        // Не трогаем fallingParticles, они генерируются независимо
        
        // Сначала звезда дрожит при старте
        setStarAnimation('rocketStart');
        
        // Через 1 секунду меняем анимацию на полет
        setTimeout(() => {
            setStarAnimation('flying');
        }, 1000);

        // Используем переменную для отслеживания времени последнего обновления
        let lastUpdateTime = Date.now();
        let lastSparkTime = Date.now();
        let sparkIntensity = 1; // Начальная интенсивность искр
        let lastPositionValue = 0; // Для сглаживания движения

        multiplierTimerRef.current = setInterval(() => {
            const now = Date.now();
            // Обеспечиваем минимальный интервал между обновлениями UI для предотвращения мерцания
            const minUpdateInterval = 100; // мс
            
            const elapsedSeconds = (now - startTime) / 1000;
            
            // Using a simplified growth model: multiplier = e^(0.1 * time)
            const currentMultiplier = initialMultiplier * Math.pow(Math.E, 0.1 * elapsedSeconds);
            
            // Format to 2 decimal places
            const formattedMultiplier = parseFloat(currentMultiplier).toFixed(2);
            
            // Увеличиваем интенсивность искр с ростом коэффициента
            sparkIntensity = Math.min(3, 1 + (parseFloat(formattedMultiplier) - 1) / 2);
            
            // Обновляем значение только если изменились сотые доли И прошло достаточно времени с последнего обновления
            if (formattedMultiplier !== lastDisplayedValue && (now - lastUpdateTime) >= minUpdateInterval) {
                lastDisplayedValue = formattedMultiplier;
                lastUpdateTime = now;
                setXValue(formattedMultiplier);
                
                // Рассчитываем позицию звезды на основе мультипликатора
                // Начинаем с 0 (внизу) и поднимаем звезду вверх с ростом коэффициента
                // Максимальное значение 100 (верх контейнера)
                const newRawPosition = Math.min(100, (parseFloat(formattedMultiplier) - 1) * 50);
                
                // Сглаживаем движение, применяя интерполяцию
                const smoothedPosition = lastPositionValue * 0.3 + newRawPosition * 0.7;
                lastPositionValue = smoothedPosition;
                
                setStarPosition(smoothedPosition);
            }
            
            // Генерируем искры каждые 100мс
            if (now - lastSparkTime >= 100) {
                lastSparkTime = now;
                
                // Создаем новые искры, если звезда движется
                if (parseFloat(formattedMultiplier) > 1.05) {
                    const newSparks = [];
                    // Количество искр зависит от интенсивности (скорости роста)
                    const sparkCount = Math.floor(Math.random() * 4 + sparkIntensity * 3);
                    
                    for (let i = 0; i < sparkCount; i++) {
                        // Разные типы искр для разнообразия
                        const sparkType = Math.random() < 0.33 ? 'gold' : (Math.random() < 0.66 ? 'orange' : 'bright');
                        
                        // Чем выше скорость, тем дальше разлетаются искры
                        const xSpread = Math.random() * 40 * sparkIntensity - 20 * sparkIntensity;
                        const ySpread = Math.random() * 30 * sparkIntensity + 20;
                        
                        newSparks.push({
                            id: Date.now() + i,
                            x: xSpread, // Случайное смещение по X, шире с ростом интенсивности
                            y: ySpread, // Смещение вниз, больше с ростом интенсивности
                            type: sparkType,
                            duration: Math.random() * 400 + 600 / sparkIntensity, // Быстрее с ростом интенсивности
                            opacity: Math.random() * 0.3 + 0.7,
                            size: Math.random() * 0.5 + 0.5 * sparkIntensity // Размер искр увеличивается с ростом интенсивности
                        });
                    }
                    
                    // Добавляем новые искры и удаляем старые (более 40)
                    setSparkParticles(prev => [...newSparks, ...prev].slice(0, 40));
                }
            }
        }, 16); // Update ~60 times per second for calculation accuracy
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
        const ws = new WebSocket(`${WS_PROTOCOL}://${API_BASE_URL}/ws/crashgame/live?init_data=${encoded_init_data}`);
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
                    // Immediately update game state
                    setIsBettingClosed(true);
                    setIsCrashed(false);
                    setGameActive(true);
                    setCollapsed(false);
                    
                    // If this is the first multiplier update, start simulation
                    if (!startMultiplierTime) {
                        // Сбрасываем позицию звезды в начале игры
                        setStarPosition(0);
                        setStarExploding(false);
                        setStarAnimation('rocketStart'); // Начинаем с дрожания
                        
                        setStartMultiplierTime(Date.now());
                        simulateMultiplierGrowth(Date.now(), parseFloat(data.multiplier));
                    }
                    
                    // Automatic cashout when reaching the specified multiplier
                    if (isAutoEnabled && bet > 0 && parseFloat(data.multiplier) >= autoOutputCoefficient && autoOutputCoefficient > 0) {
                        handleCashout();
                        toast.success(`Auto cashout at ${data.multiplier}x`);
                    }
                }

                if (data.type === "game_crash" && !isCrashed) {
                    // Immediately stop multiplier growth simulation
                    if (multiplierTimerRef.current) {
                        clearInterval(multiplierTimerRef.current);
                        multiplierTimerRef.current = null;
                    }
                    setStartMultiplierTime(null);
                    
                    // Анимируем взрыв звезды
                    setStarExploding(true);
                    setStarAnimation(''); // Удаляем предыдущие анимации
                    
                    // Очищаем все падающие частицы при взрыве
                    setFallingParticles([]);
                    
                    // Генерируем много частиц для взрыва из центра
                    const explosionParticles = [];
                    for (let i = 0; i < 30; i++) {
                        const angle = Math.random() * Math.PI * 2; // Случайный угол в радианах
                        const distance = Math.random() * 100 + 50; // Расстояние от 50 до 150px
                        
                        // Рассчитываем координаты на основе угла и расстояния
                        const xEnd = Math.cos(angle) * distance;
                        const yEnd = Math.sin(angle) * distance;
                        
                        const particleType = Math.random() < 0.33 ? 'gold' : (Math.random() < 0.66 ? 'orange' : 'bright');
                        
                        explosionParticles.push({
                            id: Date.now() + 1000 + i,
                            x: xEnd,
                            y: yEnd,
                            type: particleType,
                            duration: Math.random() * 1000 + 500,
                            opacity: Math.random() * 0.5 + 0.5,
                            size: Math.random() * 2 + 1
                        });
                    }
                    
                    setSparkParticles(explosionParticles);
                    
                    // Мгновенно обновляем все значения
                    setIsCrashed(true);
                    setGameActive(false);
                    const crashPoint = parseFloat(data.crash_point).toFixed(2);
                    setCollapsed(true);
                    setXValue(crashPoint);
                    
                    // Мгновенно очищаем состояние
                    if (bet > 0) {
                        toast.error(`Game crashed at ${crashPoint}x! You lost ₹${bet}.`);
                        setBet(0);
                    }
                }

                if (data.type === "cashout_result") {
                    toast.success(`You won ₹${data.win_amount.toFixed(0)}! (${parseFloat(data.cashout_multiplier).toFixed(2)}x)`);
                    setBet(0);
                    increaseBalanceRupee(data.win_amount);
                    // Показываем результат выигрыша
                    setShowWinResult(true);
                    // Через 3 секунды скрываем результат
                    setTimeout(() => {
                        setShowWinResult(false);
                    }, 3000);
                }

                if (data.type === "other_cashout") {
                    toast.success(`${data.username} won ₹${data.win_amount.toFixed(0)} at ${parseFloat(data.cashout_multiplier).toFixed(2)}x!`);
                }

                if (data.type === "new_bet") {
                    toast.success(`${data.username} bet ₹${data.amount.toFixed(0)}`);
                }
                
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

                if (data.type === "timer_tick") {
                    setIsBettingClosed(data.remaining_time > 10);
                    setIsCrashed(false);
                    setGameActive(false);
                    setCollapsed(true);
                    
                    if (data.remaining_time <= 10) {
                        setOverlayText(`Game starts in ${data.remaining_time} seconds`);
                    }
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

    useEffect(() => {
        // Начинаем генерировать падающие частицы сразу при загрузке компонента
        // и они будут генерироваться постоянно, пока не произойдет взрыв звезды
        const particleInterval = setInterval(() => {
            if (!isCrashed) {
                generateFallingParticles();
            }
        }, 200);
        
        return () => clearInterval(particleInterval);
    }, [isCrashed]); // Перезапускаем эффект при изменении isCrashed

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
                    {/* Показываем блок с выигрышем, если showWinResult активен */}
                    {showWinResult ? (
                        <div className={styles.winResult}>
                            <div className={styles.winResult_title}>Поздравляем!</div>
                            <div className={styles.winResult_amount}>₹{Math.floor(winAmount)}</div>
                            <div className={styles.winResult_multiplier}>x{xValue}</div>
                        </div>
                    ) : (
                        <>
                            <div className={`${styles.explodedStar} ${starExploding ? styles.explode : ''}`}>
                                <img 
                                    src="/star.svg" 
                                    alt="Star" 
                                    className={styles.starImage}
                                />
                                {/* Не показываем никакой коэффициент внутри звезды */}
                            </div>
                            
                            {/* Показываем коэффициент под звездой только если звезда взорвалась */}
                            {starExploding && 
                                <span className={styles.crashValueBelow}>{xValue}x</span>
                            }
                        </>
                    )}
                </div>
                
                {/* Центральная часть с игровыми элементами */}
                <div className={styles.gameCenter}>
                    {/* Падающие частицы, пролетающие мимо звезды */}
                    {fallingParticles.map(particle => (
                        <div
                            key={particle.id}
                            className={`${styles.fallingParticle} ${styles[particle.type]}`}
                            style={{
                                left: `calc(50% + ${particle.xOffset}px)`,
                                animationDuration: `${particle.speed}ms`,
                                width: `${3 * particle.size}px`,
                                height: `${10 * particle.size}px`,
                                // Добавляем свойство, которое гарантирует исчезновение анимации
                                willChange: 'top, opacity',
                                // Явно указываем режим анимации для уверенности
                                animationFillMode: 'forwards',
                                // Гарантированное удаление с экрана по окончании анимации
                                animationTimingFunction: 'linear'
                            }}
                        />
                    ))}
                    
                    {/* Эффект свечения вокруг звезды */}
                    {parseFloat(xValue) > 1.05 && !isCrashed && (
                        <div 
                            className={`${styles.glowEffect} ${styles.active}`} 
                            style={{ left: '50%', top: '50%' }}
                        />
                    )}
                    
                    {/* Искры вокруг звезды в центре */}
                    {sparkParticles.map(spark => (
                        <div
                            key={spark.id}
                            className={`${styles.sparkParticle} ${styles[spark.type]} ${styles.active}`}
                            style={{
                                '--x': `${spark.x}px`,
                                '--y': `${spark.y}px`,
                                left: isCrashed ? 'calc(50% - 2px)' : '50%',
                                top: isCrashed ? 'calc(50% - 2px)' : '50%',
                                animationDuration: `${spark.duration}ms`,
                                opacity: spark.opacity,
                                width: `${4 * (spark.size || 1)}px`,
                                height: `${4 * (spark.size || 1)}px`
                            }}
                        />
                    ))}
                    
                    {/* Звезда по центру */}
                    <div 
                        className={`${styles.starContainer} ${isCrashed ? styles.hidden : ''}`}
                    >
                        <img 
                            src="/star.svg" 
                            alt="Star" 
                            className={`${styles.star} ${starAnimation ? styles[starAnimation] : ''}`}
                        />
                    </div>
                </div>
                
                {/* Multiplier display - теперь внизу, показываем только во время активной игры */}
                {gameActive && !isCrashed && (
                    <div className={styles.multiplierBottom}>
                        {xValue} x
                    </div>
                )}
                
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