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

    const [starPosition, setStarPosition] = useState({ x: 50, y: -40 });

    
    const wsRef = useRef(null);
    const multiplierTimerRef = useRef(null);
    const [startMultiplierTime, setStartMultiplierTime] = useState(null);

    const valXValut = useRef(1);

    // Добавляем новое состояние для отслеживания ставки в очереди
    const [queuedBet, setQueuedBet] = useState(0);

    const lastUpdateRef = useRef(null);

    const [connectionStatus, setConnectionStatus] = useState('connecting');
    const [isReconnecting, setIsReconnecting] = useState(false);
    const reconnectAttempts = useRef(0);
    const MAX_RECONNECT_ATTEMPTS = 5;
    const RECONNECT_DELAY = 1000;

    // Добавляем константы для управления переподключением
    const NETWORK_ERROR_CODES = {
        NORMAL_CLOSURE: 1000,
        GOING_AWAY: 1001,
        PROTOCOL_ERROR: 1002,
        UNSUPPORTED_DATA: 1003,
        NO_STATUS_RECEIVED: 1005,
        ABNORMAL_CLOSURE: 1006,
        INVALID_FRAME_PAYLOAD_DATA: 1007,
        POLICY_VIOLATION: 1008,
        MESSAGE_TOO_BIG: 1009,
        MISSING_EXTENSION: 1010,
        INTERNAL_ERROR: 1011,
        SERVICE_RESTART: 1012,
        TRY_AGAIN_LATER: 1013,
        BAD_GATEWAY: 1014,
        TLS_HANDSHAKE: 1015
    };

    useEffect(() => {
        const queuedBetFromStorage = localStorage.getItem('queuedBet');
        if (queuedBetFromStorage) {
            setQueuedBet(parseInt(queuedBetFromStorage));
        }
    }, []);

    useEffect(() => {
        const interval = setInterval(() => {
            setXValue(valXValut.current);
        }, 80);
    
        return () => clearInterval(interval);
    }, []);

    const placeBetQueue = async (queueBetFromStorage) => {
        try {
            // Проверяем, что ставка все еще в очереди
            const currentQueuedBet = localStorage.getItem('queuedBet');
            if (!currentQueuedBet || Number(currentQueuedBet) !== Number(queueBetFromStorage)) {
                console.log('Queued bet was changed or removed');
                return;
            }

            // Добавляем небольшую задержку перед размещением ставки
            await new Promise(resolve => setTimeout(resolve, 1000));
            
            // Проверяем состояние игры перед размещением ставки
            if (gameActive || isCrashed) {
                console.log('Game is not ready for placing bet');
                return;
            }

            const response = await crashPlace(Number(queueBetFromStorage), autoOutputCoefficient);

            if (response.ok) {
                setBet(parseInt(queueBetFromStorage));
                localStorage.removeItem('queuedBet');
                setQueuedBet(0);
                
                // Сбрасываем множитель и перезапускаем симуляцию
                valXValut.current = 1.0;
                setXValue(1.0);
                setStartMultiplierTime(Date.now());
                simulateMultiplierGrowth(Date.now(), 1.0);
                
                // Обновляем позицию звезды
                setStarPosition({ x: 50, y: -40 });
                
                toast.success('Queued bet placed successfully!');
            } else {
                // Если не удалось поставить, пробуем еще раз через 1 секунду
                console.log('Failed to place queued bet, retrying...');
                setTimeout(() => placeBetQueue(queueBetFromStorage), 1000);
            }
        } catch (error) {
            console.error('Error placing queued bet:', error);
            // В случае ошибки пробуем еще раз через 1 секунду
            setTimeout(() => placeBetQueue(queueBetFromStorage), 1000);
        }
    }

    useEffect(() => {
        if (!isBettingClosed && !gameActive && !isCrashed) {
            const queueBetFromStorage = localStorage.getItem('queuedBet');
            if (queueBetFromStorage) {
                placeBetQueue(queueBetFromStorage);
            }
        }
    }, [isBettingClosed, gameActive, isCrashed]);
    

    console.log(dimensions)
    // Getting game history on component load
    useEffect(() => {
        const fetchHistory = async () => {
            try {
                const data = await crashGetHistory();
                if (data && data.results) {
                    const lastResult = data.results[0];
                    if (lastResult) {
                        valXValut.current = parseFloat(lastResult.CrashPointMultiplier.toFixed(2));
                    }
                }
            } catch (error) {
                console.error('Error fetching game history:', error);
            }
        };

        fetchHistory();
    }, []);

    // Добавляем функцию для проверки состояния игры
    const checkGameState = () => {
        const now = Date.now();
        if (gameActive && lastUpdateRef.current && now - lastUpdateRef.current > 3000) {
            console.log('Game appears to be stalled, attempting recovery...');
            setGameActive(false);
            setIsBettingClosed(true);
            if (multiplierTimerRef.current) {
                clearInterval(multiplierTimerRef.current);
            }
            if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
                wsRef.current.close();
            }
            return true;
        }
        return false;
    };

    // Добавляем интервал для проверки состояния игры
    useEffect(() => {
        const gameStateCheckInterval = setInterval(() => {
            if (gameActive) {
                checkGameState();
            }
        }, 1000);

        return () => {
            clearInterval(gameStateCheckInterval);
        };
    }, [gameActive]);

    // Модифицируем функцию simulateMultiplierGrowth
    const simulateMultiplierGrowth = (startTime, initialMultiplier = 1.0) => {
        if (multiplierTimerRef.current) {
            clearInterval(multiplierTimerRef.current);
        }

        valXValut.current = initialMultiplier;
        const updateInterval = 100;
        const growthFactor = 0.03;
        const maxMultiplier = 100;
        
        let lastValue = initialMultiplier;
        let lastUpdateTime = Date.now();
        let stallCount = 0;
        
        multiplierTimerRef.current = setInterval(() => {
            const now = Date.now();
            
            // Проверка на зависание
            if (now - lastUpdateTime > 1000) {
                stallCount++;
                if (stallCount >= 3) {
                    console.log('Game multiplier stalled multiple times, resetting...');
                    clearInterval(multiplierTimerRef.current);
                    setGameActive(false);
                    setIsBettingClosed(true);
                    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
                        wsRef.current.close();
                    }
                    return;
                }
            } else {
                stallCount = 0;
            }
            
            const elapsedSeconds = (now - startTime) / 1000;
            const newMultiplier = Math.min(Math.exp(elapsedSeconds * growthFactor), maxMultiplier);
            
            // Сглаживание
            const smoothedMultiplier = (lastValue * 0.8 + newMultiplier * 0.2).toFixed(2);
            lastValue = parseFloat(smoothedMultiplier);
            lastUpdateTime = now;
            
            if (lastValue >= maxMultiplier) {
                clearInterval(multiplierTimerRef.current);
                setGameActive(false);
                setIsBettingClosed(true);
                return;
            }
            
            valXValut.current = lastValue;
            setXValue(lastValue);
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

        // Инициализируем WebSocket соединение
        connectWebSocket();

        return () => {
            window.removeEventListener('resize', updateDimensions);
            if (wsRef.current) {
                wsRef.current.close();
            }
            if (multiplierTimerRef.current) {
                clearInterval(multiplierTimerRef.current);
            }
        };
    }, [initData, bet, autoOutputCoefficient, isAutoEnabled]);

    // Добавляем функцию для логирования
    const logError = (error, context) => {
        console.error(`[${context}] Error:`, error);
        toast.error(`Error in ${context}. Please try again.`);
    };

    // Добавляем функцию для обработки сообщений WebSocket
    const handleWebSocketMessage = async (event) => {
        try {
            const data = JSON.parse(event.data);
            console.log('WebSocket data received:', data);
            
            switch (data.type) {
                case "multiplier_update":
                    handleMultiplierUpdate(data);
                    break;
                case "game_crash":
                    handleGameCrash(data);
                    break;
                case "timer_tick":
                    handleTimerTick(data);
                    break;
                case "game_started":
                    handleGameStarted();
                    break;
                case "cashout_result":
                    handleCashoutResult(data);
                    break;
                case "other_cashout":
                    handleOtherCashout(data);
                    break;
                case "new_bet":
                    handleNewBet(data);
                    break;
                default:
                    console.warn('Unknown message type:', data.type);
            }
        } catch (error) {
            logError(error, 'processing WebSocket message');
        }
    };

    // Добавляем обработчики для каждого типа сообщения
    const handleMultiplierUpdate = (data) => {
        const currentTime = Date.now();
        
        // Проверка на зависание
        if (lastUpdateRef.current && currentTime - lastUpdateRef.current > 3000) {
            console.log('Game stalled, resetting...');
            setGameActive(false);
            setIsBettingClosed(true);
            if (multiplierTimerRef.current) {
                clearInterval(multiplierTimerRef.current);
            }
            if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
                wsRef.current.close();
            }
            return;
        }
        
        lastUpdateRef.current = currentTime;
        
        setIsBettingClosed(true);
        setIsCrashed(false);
        setGameActive(true);
        setCollapsed(false);

        setStarPosition({
            x: Math.min(200, 50 + data.multiplier * 40 - 40),
            y: Math.max(-200, -data.multiplier * 40),
        });
        
        if (!startMultiplierTime) {
            setStartMultiplierTime(Date.now());
            simulateMultiplierGrowth(Date.now(), parseFloat(data.multiplier));
        }
        
        if (isAutoEnabled && bet > 0 && parseFloat(data.multiplier) >= autoOutputCoefficient && autoOutputCoefficient > 0) {
            handleCashout();
            toast.success(`Auto cashout at ${data.multiplier}x`);
        }
    };

    const handleGameCrash = (data) => {
        if (multiplierTimerRef.current) {
            clearInterval(multiplierTimerRef.current);
            multiplierTimerRef.current = null;
        }
        setStartMultiplierTime(null);
        
        setIsCrashed(true);
        setGameActive(false);
        setIsBettingClosed(true);
        setOverlayText(`Crashed at ${data.crash_point.toFixed(2)}x`);
        setCollapsed(true);
        valXValut.current = parseFloat(data.crash_point).toFixed(2);
        setXValue(parseFloat(data.crash_point).toFixed(2));

        setStarPosition({ x: 50, y: -40 });
        
        // Обработка ставки в очереди
        const queueBetFromStorage = localStorage.getItem('queuedBet');
        if (queueBetFromStorage) {
            setTimeout(async () => {
                try {
                    const response = await crashPlace(Number(queueBetFromStorage), autoOutputCoefficient);
                    if (response.ok) {
                        setBet(parseInt(queueBetFromStorage));
                        localStorage.removeItem('queuedBet');
                        setQueuedBet(0);
                        toast.success('Queued bet placed successfully!');
                    } else {
                        setTimeout(() => placeBetQueue(queueBetFromStorage), 1000);
                    }
                } catch (error) {
                    logError(error, 'placing queued bet');
                    setTimeout(() => placeBetQueue(queueBetFromStorage), 1000);
                }
            }, 1000);
        }
        
        setTimeout(() => {
            if (bet > 0) {
                toast.error(`Game crashed at ${data.crash_point.toFixed(2)}x! You lost ₹${bet}.`);
                setBet(0);
            }
            valXValut.current = 1.2;
            setXValue(1.2);
            setStarPosition({ x: 50, y: -40 });
        }, 3000);
    };

    const handleTimerTick = (data) => {
        setCollapsed(true);
        console.log('Timer tick received:', data.remaining_time);
        
        if (data.remaining_time > 5) {
            setIsBettingClosed(true);
            setGameActive(false);
            setOverlayText('Game starts soon');
        } else if (data.remaining_time > 0) {
            setIsBettingClosed(false);
            setIsCrashed(false);
            setGameActive(false);
            setOverlayText(`Game starts in ${data.remaining_time} seconds`);
        } else {
            setIsBettingClosed(false);
            setGameActive(true);
            setOverlayText('Game started!');
        }
    };

    const handleGameStarted = () => {
        try {
            toast.success('Game started!');
            setIsBettingClosed(false);
            setIsCrashed(false);
            setGameActive(true);
            setCollapsed(false);
            
            setStartMultiplierTime(Date.now());
            simulateMultiplierGrowth(Date.now(), 1.0);
            setXValue(1.0);

            const queueBetFromStorage = localStorage.getItem('queuedBet');
            if (queueBetFromStorage) {
                setTimeout(async () => {
                    try {
                        const response = await crashPlace(Number(queueBetFromStorage), autoOutputCoefficient);
                        if (response.ok) {
                            setBet(parseInt(queueBetFromStorage));
                            localStorage.removeItem('queuedBet');
                            setQueuedBet(0);
                            toast.success('Queued bet placed successfully!');
                        } else {
                            const errorData = await response.json().catch(() => ({ error: 'Failed to place queued bet' }));
                            logError(errorData.error, 'placing queued bet');
                            setTimeout(() => placeBetQueue(queueBetFromStorage), 1000);
                        }
                    } catch (error) {
                        logError(error, 'placing queued bet');
                        setTimeout(() => placeBetQueue(queueBetFromStorage), 1000);
                    }
                }, 1000);
            }
        } catch (error) {
            logError(error, 'starting game');
            setGameActive(false);
            setIsBettingClosed(true);
        }
    };

    const handleCashoutResult = (data) => {
        toast.success(`You won ₹${data.win_amount.toFixed(0)}! (${data.cashout_multiplier}x)`);
        increaseBalanceRupee(data.win_amount);
    };

    const handleOtherCashout = (data) => {
        toast.success(`${data.username} won ₹${data.win_amount.toFixed(0)} at ${data.cashout_multiplier}x!`);
    };

    const handleNewBet = (data) => {
        toast.success(`${data.username} bet ₹${data.amount.toFixed(0)}`);
    };

    // Модифицируем функцию handleBet
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
            
            if (gameActive || queuedBet > 0) {
                setQueuedBet(betAmount);
                decreaseBalanceRupee(betAmount);
                localStorage.setItem('queuedBet', betAmount);
                toast.success('Bet will be placed in the next game!');
                setLoading(false);
                return;
            }

            const response = await crashPlace(betAmount, autoOutputCoefficient);
            
            if (response.ok) {
                setBet(betAmount);
                decreaseBalanceRupee(betAmount);
                toast.success('Bet accepted! Waiting for game to start');
                setCollapsed(true);
                setOverlayText('Your bet is accepted! Waiting for game...');
                setTimeout(() => {
                    setCollapsed(false);
                }, 2000);
            } else {
                const errorData = await response.json().catch(() => ({ error: 'Failed to place bet' }));
                logError(errorData.error, 'placing bet');
                setQueuedBet(betAmount);
                decreaseBalanceRupee(betAmount);
                localStorage.setItem('queuedBet', betAmount);
                toast.success('Bet will be placed in the next game!');
            }
        } catch (err) {
            logError(err, 'placing bet');
            setQueuedBet(betAmount);
            decreaseBalanceRupee(betAmount);
            localStorage.setItem('queuedBet', betAmount);
            toast.success('Bet will be placed in the next game!');
        } finally {
            setLoading(false);
        }
    };

    // Модифицируем функцию handleCashout
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
                setBet(0);
                setGameActive(false);
                setIsCrashed(false);
                toast.success(`Cashout request sent at multiplier ${xValue}x`);
            } else {
                const errorData = await response.json().catch(() => ({ error: 'Failed to cash out' }));
                logError(errorData.error, 'cashing out');
            }
        } catch (err) {
            logError(err, 'cashing out');
        } finally {
            setLoading(false);
        }
    };

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

    // Добавляем компонент для отображения статуса соединения
    const ConnectionStatus = () => {
        const getStatusColor = () => {
            switch (connectionStatus) {
                case 'connected':
                    return 'green';
                case 'connecting':
                    return 'orange';
                case 'error':
                case 'disconnected':
                    return 'red';
                default:
                    return 'gray';
            }
        };

        return (
            <div style={{
                position: 'absolute',
                top: '10px',
                right: '10px',
                padding: '5px 10px',
                borderRadius: '5px',
                backgroundColor: getStatusColor(),
                color: 'white',
                fontSize: '12px',
                zIndex: 1000
            }}>
                {connectionStatus.toUpperCase()}
            </div>
        );
    };

    // Модифицируем функцию connectWebSocket
    const connectWebSocket = () => {
        if (!initData) {
            toast.error('Authorization error. Please restart the application.');
            return;
        }

        const encoded_init_data = encodeURIComponent(initData);
        const ws = new WebSocket(`wss://${API_BASE_URL}/ws/crashgame/live?init_data=${encoded_init_data}`);
        wsRef.current = ws;

        // Пинг для поддержания соединения
        const pingInterval = setInterval(() => {
            if (ws.readyState === WebSocket.OPEN) {
                try {
                    ws.send('ping');
                } catch (error) {
                    console.error('Error sending ping:', error);
                    ws.close();
                }
            }
        }, 30000);

        ws.onopen = () => {
            console.log('WebSocket connection established');
            setConnectionStatus('connected');
            setIsReconnecting(false);
            reconnectAttempts.current = 0;
        };

        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            setConnectionStatus('error');
            toast.error('Connection error. Please reload the page.');
        };

        ws.onmessage = handleWebSocketMessage;

        ws.onclose = (event) => {
            console.log('WebSocket connection closed:', event.code, event.reason);
            clearInterval(pingInterval);
            
            if (multiplierTimerRef.current) {
                clearInterval(multiplierTimerRef.current);
            }
            setGameActive(false);
            setIsBettingClosed(true);
            
            // Определяем, нужно ли пытаться переподключиться
            const shouldReconnect = ![
                NETWORK_ERROR_CODES.NORMAL_CLOSURE,
                NETWORK_ERROR_CODES.GOING_AWAY,
                NETWORK_ERROR_CODES.PROTOCOL_ERROR,
                NETWORK_ERROR_CODES.UNSUPPORTED_DATA
            ].includes(event.code);

            if (shouldReconnect && !isReconnecting && reconnectAttempts.current < MAX_RECONNECT_ATTEMPTS) {
                setIsReconnecting(true);
                reconnectAttempts.current += 1;
                const delay = RECONNECT_DELAY * Math.min(reconnectAttempts.current, 5);
                console.log(`Attempting to reconnect in ${delay}ms (attempt ${reconnectAttempts.current})`);
                
                setTimeout(() => {
                    connectWebSocket();
                }, delay);
            } else if (reconnectAttempts.current >= MAX_RECONNECT_ATTEMPTS) {
                setConnectionStatus('disconnected');
                toast.error('Failed to reconnect. Please refresh the page.');
            }
        };

        return () => {
            clearInterval(pingInterval);
            if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
                ws.close();
            }
        };
    };

    return (
        <div className={styles.crash}>
            <ConnectionStatus />
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
                    className={`${styles.star} ${isCrashed ? styles.falling : ''}`} 
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
                {queuedBet > 0 && <div className={styles.queuedBet}>
                    Queued bet: ₹{queuedBet}
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
                        <input className={styles.amount__input} value={betAmount}
                                onChange={(e) => {
                                    const value = e.target.value.replace(/\D/g, ""); // Remove non-digit characters
                                    setBetAmount(value);
                                }}
                            ></input>
                        <div className={styles.betAmountButtons}>
                            <button className={styles.betButton} onClick={() => handleAmountChange(-100)}>-</button>
                            <button className={styles.betButton} onClick={() => handleAmountChange(100)}>+</button>
                        </div>
                    </div>

                    <div className={styles.quickButtons}>
                        <button className={styles.quickButton} onClick={() => handleMultiplyAmount(0.5)}>/2</button>
                        <button className={styles.quickButton} onClick={() => handleMultiplyAmount(2)}>x2</button>
                    </div>

                    {bet > 0 && !queuedBet ? (
                        <button 
                            className={`${styles.mainButton} ${(gameActive && !isCrashed) ? styles.activeButton : ''}`} 
                            onClick={handleCashout} 
                            disabled={!gameActive || loading || isCrashed || !bet}
                        >
                            {loading ? 'Loading...' : 
                             !gameActive ? 'Waiting for game...' :
                             isCrashed ? 'Game finished' :
                             'Cashout'}
                        </button>
                    ) : (
                        <button 
                            className={`${styles.mainButton} ${isBettingClosed ? styles.queuedButton : ''}`}
                            onClick={handleBet} 
                            disabled={loading || queuedBet > 0 || (gameActive && bet > 0)}
                        >
                            {loading ? 'Loading...' : 
                             queuedBet > 0 ? `Queued: ₹${queuedBet}` :
                             gameActive ? 'Queue Bet' : 'Place Bet'}
                        </button>
                    )} 
                </div>
            </div>
        </div>
    );
};