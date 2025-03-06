import { useEffect, useState, useRef } from 'react';
import { crashPlace, crashCashout, crashGetHistory } from '@/requests';
import styles from "./Crash.module.scss";
import { API_BASE_URL } from '@/config';
const initData = window.Telegram?.WebApp?.initData || '';
import toast from 'react-hot-toast';
import useStore from '@/store';
import BetControls from '@/components/BetControls';

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
    
    // Refs для анимации звезды
    const starRef = useRef(null);
    const starContainerRef = useRef(null);
    const particlesRef = useRef([]);
    const [isStarFlying, setIsStarFlying] = useState(false);
    const [isStarExploding, setIsStarExploding] = useState(false);
    const trailTimerRef = useRef(null);
    const smallParticleTimerRef = useRef(null);
    const sparksTimerRef = useRef(null);
    
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

        // Set initial value
        setXValue(initialMultiplier);
        
        // Запускаем анимацию звезды при начале роста коэффициента
        setIsStarFlying(true);
        setIsStarExploding(false);
        
        // Сначала создаем эффект мощного запуска
        if (starContainerRef.current && starRef.current) {
            // Дополнительный класс для начальной анимации взлета
            starRef.current.classList.add(styles.rocketStart);
            
            // Создаем начальный "взрыв" искр при запуске
            for (let i = 0; i < 3; i++) {
                setTimeout(() => {
                    createSparksEffect();
                }, i * 100);
            }
            
            // Через секунду удаляем класс начальной анимации
            setTimeout(() => {
                if (starRef.current) {
                    starRef.current.classList.remove(styles.rocketStart);
                }
            }, 1000);
        }
        
        // Запускаем создание следа звезды 
        if (trailTimerRef.current) {
            clearInterval(trailTimerRef.current);
        }
        
        trailTimerRef.current = setInterval(() => {
            createStarTrail();
        }, 40); // Увеличиваем частоту следа для большей плотности
        
        // Запускаем создание маленьких частиц
        if (smallParticleTimerRef.current) {
            clearInterval(smallParticleTimerRef.current);
        }
        
        smallParticleTimerRef.current = setInterval(() => {
            createSmallParticles();
        }, 180);
        
        // Запускаем создание искр, летящих снизу
        if (sparksTimerRef.current) {
            clearInterval(sparksTimerRef.current);
        }
        
        sparksTimerRef.current = setInterval(() => {
            createSparksEffect();
        }, 120); // Увеличиваем частоту появления искр
        
        // Добавляем начальное свечение
        setTimeout(() => {
            if (starContainerRef.current && starRef.current) {
                const starRect = starRef.current.getBoundingClientRect();
                const containerRect = starContainerRef.current.getBoundingClientRect();
                
                if (starRect && containerRect) {
                    const starCenterX = starRect.left + starRect.width / 2 - containerRect.left;
                    const starCenterY = starRect.top + starRect.height / 2 - containerRect.top;
                    
                    createGlowEffect(starCenterX, starCenterY);
                }
            }
        }, 100);
        
        const updateInterval = 100; // ms
        const growthFactor = 0.03; // how fast the multiplier grows
        
        multiplierTimerRef.current = setInterval(() => {
            const elapsedSeconds = (Date.now() - startTime) / 1000;
            // Formula for calculating multiplier: e^(elapsedSeconds * growthFactor)
            const newMultiplier = Math.exp(elapsedSeconds * growthFactor);
            
            // Создаем дополнительные частицы при высоких значениях коэффициента
            if (newMultiplier > 2 && Math.random() > 0.7) {
                createSmallParticles();
            }
            
            setXValue(parseFloat(newMultiplier.toFixed(2)));
        }, updateInterval);
    };

    // Создаем след за звездой
    const createStarTrail = () => {
        if (!starContainerRef.current || !starRef.current || !isStarFlying) return;
        
        const starRect = starRef.current.getBoundingClientRect();
        const containerRect = starContainerRef.current.getBoundingClientRect();
        
        // Создаем след
        const trail = document.createElement('div');
        
        // Случайно выбираем тип следа
        const trailTypes = ['goldTrail', 'redTrail', 'whiteTrail'];
        const randomTrailType = trailTypes[Math.floor(Math.random() * trailTypes.length)];
        
        trail.className = `${styles.starTrail} ${styles[randomTrailType]} ${styles.active}`;
        
        // Позиционируем след по центру звезды с небольшим смещением для красоты
        const starCenterX = starRect.left + starRect.width / 2 - containerRect.left;
        const starBottomY = starRect.bottom - containerRect.top;
        
        // Добавляем случайное смещение для разнообразия
        const xOffset = (Math.random() - 0.5) * 10;
        
        trail.style.left = `${starCenterX + xOffset}px`;
        trail.style.top = `${starBottomY}px`;
        
        starContainerRef.current.appendChild(trail);
        
        // Удаляем след после окончания анимации
        setTimeout(() => {
            if (starContainerRef.current && trail.parentNode === starContainerRef.current) {
                starContainerRef.current.removeChild(trail);
            }
        }, 1000);
        
        // С некоторой вероятностью создаем эффект свечения вокруг звезды
        if (Math.random() > 0.7) {
            createGlowEffect(starCenterX, starRect.top + starRect.height / 2 - containerRect.top);
        }
    };
    
    // Создаем эффект свечения вокруг звезды
    const createGlowEffect = (x, y) => {
        if (!starContainerRef.current) return;
        
        const glow = document.createElement('div');
        glow.className = `${styles.glowEffect} ${styles.active}`;
        
        glow.style.left = `${x}px`;
        glow.style.top = `${y}px`;
        
        starContainerRef.current.appendChild(glow);
        
        // Удаляем эффект свечения после окончания анимации
        setTimeout(() => {
            if (starContainerRef.current && glow.parentNode === starContainerRef.current) {
                starContainerRef.current.removeChild(glow);
            }
        }, 1500);
    };
    
    // Создаем маленькие частицы при полете звезды
    const createSmallParticles = () => {
        if (!starContainerRef.current || !starRef.current || !isStarFlying) return;
        
        const starRect = starRef.current.getBoundingClientRect();
        const containerRect = starContainerRef.current.getBoundingClientRect();
        
        // Создаем 3-5 маленьких частиц
        const particleCount = 3 + Math.floor(Math.random() * 3);
        const particleTypes = ['goldParticle', 'redParticle', 'whiteParticle'];
        
        for (let i = 0; i < particleCount; i++) {
            const particle = document.createElement('div');
            
            // Выбираем случайный тип частицы
            const typeIndex = Math.floor(Math.random() * particleTypes.length);
            const particleType = particleTypes[typeIndex];
            
            particle.className = `${styles.smallParticle} ${styles[particleType]} ${styles.active}`;
            
            // Случайное направление для каждой частицы
            const angle = Math.random() * Math.PI * 2;
            const distance = 15 + Math.random() * 30;
            const x = Math.cos(angle) * distance;
            const y = Math.sin(angle) * distance;
            
            particle.style.setProperty('--x', `${x}px`);
            particle.style.setProperty('--y', `${y}px`);
            
            // Позиционируем частицу на звезде
            const starCenterX = starRect.left + starRect.width / 2 - containerRect.left;
            const starCenterY = starRect.top + starRect.height / 2 - containerRect.top;
            
            // Добавляем небольшое смещение для разнообразия
            const offsetX = (Math.random() - 0.5) * 6;
            const offsetY = (Math.random() - 0.5) * 6;
            
            particle.style.left = `${starCenterX + offsetX}px`;
            particle.style.top = `${starCenterY + offsetY}px`;
            
            starContainerRef.current.appendChild(particle);
            
            // Удаляем частицу после окончания анимации
            setTimeout(() => {
                if (starContainerRef.current && particle.parentNode === starContainerRef.current) {
                    starContainerRef.current.removeChild(particle);
                }
            }, 1500);
        }
    };

    // Функция для создания частиц при взрыве звезды
    const createExplosionParticles = () => {
        if (!starContainerRef.current || !starRef.current) return;
        
        const starRect = starRef.current.getBoundingClientRect();
        const containerRect = starContainerRef.current.getBoundingClientRect();
        
        // Очищаем предыдущие частицы
        while (starContainerRef.current.querySelector(`.${styles.starParticle}`)) {
            starContainerRef.current.removeChild(
                starContainerRef.current.querySelector(`.${styles.starParticle}`)
            );
        }
        
        // Останавливаем таймеры создания следа и маленьких частиц
        if (trailTimerRef.current) {
            clearInterval(trailTimerRef.current);
            trailTimerRef.current = null;
        }
        
        if (smallParticleTimerRef.current) {
            clearInterval(smallParticleTimerRef.current);
            smallParticleTimerRef.current = null;
        }
        
        // Создаем больше частиц для более эффектного взрыва
        const particleCount = 50; // Увеличиваем количество частиц
        const particleTypes = ['type1', 'type2', 'type3', 'gold', 'bright', 'orange']; // Добавляем типы частиц
        
        // Получаем позицию звезды
        const starCenterX = starRect.left + starRect.width / 2 - containerRect.left;
        const starCenterY = starRect.top + starRect.height / 2 - containerRect.top;
        
        // Создаем вспышку в центре взрыва
        const flash = document.createElement('div');
        flash.className = styles.explosionFlash;
        flash.style.left = `${starCenterX}px`;
        flash.style.top = `${starCenterY}px`;
        starContainerRef.current.appendChild(flash);
        
        // Удаляем вспышку через некоторое время
        setTimeout(() => {
            if (starContainerRef.current && flash.parentNode === starContainerRef.current) {
                starContainerRef.current.removeChild(flash);
            }
        }, 400);
        
        // Создаем частицы в разных направлениях
        for (let i = 0; i < particleCount; i++) {
            const particle = document.createElement('div');
            
            // Выбираем случайный тип частицы
            const typeIndex = Math.floor(Math.random() * particleTypes.length);
            const particleType = particleTypes[typeIndex];
            
            // Добавляем разные типы частиц для более богатого эффекта
            const isFastParticle = Math.random() > 0.5; // 50% частиц будут быстрыми
            
            if (isFastParticle) {
                particle.className = `${styles.starParticle} ${styles[particleType]} ${styles.activeFast}`;
            } else {
                particle.className = `${styles.starParticle} ${styles[particleType]} ${styles.active}`;
            }
            
            // Случайное направление для каждой частицы
            const angle = Math.random() * Math.PI * 2;
            const distance = 80 + Math.random() * 200; // Увеличиваем дистанцию разлета
            
            // Разная скорость для частиц
            const speed = 0.5 + Math.random() * 1.5;
            particle.style.setProperty('--speed', speed);
            
            const x = Math.cos(angle) * distance;
            const y = Math.sin(angle) * distance;
            
            particle.style.setProperty('--x', `${x}px`);
            particle.style.setProperty('--y', `${y}px`);
            
            // Позиционируем частицу на звезде
            particle.style.left = `${starCenterX}px`;
            particle.style.top = `${starCenterY}px`;
            
            // Добавляем вращение некоторым частицам
            if (Math.random() > 0.5) {
                particle.style.animation = `${styles.rotate} ${1 + Math.random() * 2}s linear infinite`;
            }
            
            starContainerRef.current.appendChild(particle);
            
            // Удаляем частицы через разное время
            setTimeout(() => {
                if (starContainerRef.current && particle.parentNode === starContainerRef.current) {
                    starContainerRef.current.removeChild(particle);
                }
            }, 1000 + Math.random() * 1000);
        }
    };

    // Создаем искры, летящие снизу звезды
    const createSparksEffect = () => {
        if (!starContainerRef.current || !isStarFlying) return;
        
        // Создаем больше искр для эффекта ракеты
        const sparkCount = 12 + Math.floor(Math.random() * 8); // 12-20 искр за раз
        const sparkTypes = ['gold', 'bright', 'orange'];
        
        for (let i = 0; i < sparkCount; i++) {
            const spark = document.createElement('div');
            
            // Выбираем случайный тип искры
            const typeIndex = Math.floor(Math.random() * sparkTypes.length);
            const sparkType = sparkTypes[typeIndex];
            
            spark.className = `${styles.sparkParticle} ${styles[sparkType]} ${styles.active}`;
            
            // Получаем позицию звезды
            const starRect = starRef.current.getBoundingClientRect();
            const containerRect = starContainerRef.current.getBoundingClientRect();
            const starCenterX = starRect.left + starRect.width / 2 - containerRect.left;
            const starBottomY = starRect.bottom - containerRect.top;
            
            // Позиционируем искру внизу звезды
            const startX = starCenterX - 15 + Math.random() * 30; // разброс вокруг центра звезды
            
            // Генерируем случайный угол в нижнем конусе (направление вниз с разбросом)
            const angle = Math.PI / 2 + (Math.random() - 0.5) * Math.PI / 2; // Угол вниз с разбросом
            const distance = 50 + Math.random() * 100;
            const x = Math.cos(angle) * distance;
            const y = Math.sin(angle) * distance;
            
            spark.style.setProperty('--x', `${x}px`);
            spark.style.setProperty('--y', `${y}px`);
            
            spark.style.left = `${startX}px`;
            spark.style.top = `${starBottomY - 10}px`; // Начало чуть ниже звезды
            
            starContainerRef.current.appendChild(spark);
            
            // Удаляем искру после окончания анимации
            setTimeout(() => {
                if (starContainerRef.current && spark.parentNode === starContainerRef.current) {
                    starContainerRef.current.removeChild(spark);
                }
            }, 1000);
        }
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
                    
                    // Активируем контейнер звезды
                    if (starContainerRef.current) {
                        starContainerRef.current.classList.add(styles.active);
                    }
                    
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
                    
                    // Останавливаем таймеры создания следа и маленьких частиц
                    if (trailTimerRef.current) {
                        clearInterval(trailTimerRef.current);
                        trailTimerRef.current = null;
                    }
                    
                    if (smallParticleTimerRef.current) {
                        clearInterval(smallParticleTimerRef.current);
                        smallParticleTimerRef.current = null;
                    }
                    
                    if (sparksTimerRef.current) {
                        clearInterval(sparksTimerRef.current);
                        sparksTimerRef.current = null;
                    }
                    
                    // Взрываем звезду при окончании игры
                    setIsStarFlying(false);
                    setIsStarExploding(true);
                    
                    // Создаем эффект вспышки на экране
                    const flash = document.createElement('div');
                    flash.style.position = 'absolute';
                    flash.style.top = '0';
                    flash.style.left = '0';
                    flash.style.width = '100%';
                    flash.style.height = '100%';
                    flash.style.backgroundColor = 'rgba(255, 215, 0, 0.3)';
                    flash.style.zIndex = '4';
                    flash.style.pointerEvents = 'none';
                    flash.style.transition = 'opacity 0.5s ease-out';
                    
                    if (crashRef.current) {
                        crashRef.current.appendChild(flash);
                        
                        // Удаляем вспышку через некоторое время
                        setTimeout(() => {
                            flash.style.opacity = '0';
                            setTimeout(() => {
                                if (crashRef.current && flash.parentNode === crashRef.current) {
                                    crashRef.current.removeChild(flash);
                                }
                            }, 500);
                        }, 100);
                    }
                    
                    // Запускаем взрыв с разными группами частиц для более богатого эффекта
                    createExplosionParticles();
                    
                    // Через небольшую задержку создаем вторую волну частиц
                    setTimeout(() => {
                        createExplosionParticles();
                    }, 150);
                    
                    // Через некоторое время скрываем контейнер звезды
                    setTimeout(() => {
                        if (starContainerRef.current) {
                            starContainerRef.current.classList.remove(styles.active);
                        }
                        setIsStarExploding(false);
                    }, 800);
                    
                    setIsCrashed(true);
                    setGameActive(false);
                    setOverlayText(`Crashed at ${data.crash_point.toFixed(2)}x`);
                    setCollapsed(true);
                    setXValue(parseFloat(data.crash_point).toFixed(2));
                    
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
                    
                    // Активируем контейнер звезды
                    if (starContainerRef.current) {
                        starContainerRef.current.classList.add(styles.active);
                        
                        // Очищаем предыдущие эффекты
                        const oldElements = starContainerRef.current.querySelectorAll(
                            `.${styles.starTrail}, .${styles.smallParticle}, .${styles.glowEffect}, .${styles.sparkParticle}`
                        );
                        oldElements.forEach(el => {
                            if (el.parentNode === starContainerRef.current) {
                                starContainerRef.current.removeChild(el);
                            }
                        });
                    }
                    
                    // Start multiplier growth simulation with initial value of 1.0
                    setStartMultiplierTime(Date.now());
                    simulateMultiplierGrowth(Date.now(), 1.0);
                    
                    // Добавляем мгновенный эффект частиц для начала игры
                    setTimeout(() => {
                        createSmallParticles();
                        if (starRef.current && starContainerRef.current) {
                            const starRect = starRef.current.getBoundingClientRect();
                            const containerRect = starContainerRef.current.getBoundingClientRect();
                            if (starRect && containerRect) {
                                const starCenterX = starRect.left + starRect.width / 2 - containerRect.left;
                                const starCenterY = starRect.top + starRect.height / 2 - containerRect.top;
                                createGlowEffect(starCenterX, starCenterY);
                            }
                        }
                    }, 50);
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
            if (trailTimerRef.current) {
                clearInterval(trailTimerRef.current);
            }
            if (smallParticleTimerRef.current) {
                clearInterval(smallParticleTimerRef.current);
            }
            if (sparksTimerRef.current) {
                clearInterval(sparksTimerRef.current);
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
    const handleBetAmountChange = (newAmount) => {
        if (newAmount > 0) {
            setBetAmount(newAmount);
        }
    };

    // Multiplying bet amount
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
                <div className={styles.starContainer} ref={starContainerRef}>
                    <div className={styles.sparkTrail}></div>
                    <img 
                        src="/star.svg" 
                        alt="Star" 
                        className={`${styles.star} ${isStarFlying ? styles.flying : ''} ${isStarExploding ? styles.exploding : ''}`} 
                        ref={starRef}
                    />
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

                {bet > 0 ? (
                    <button 
                        className={`${styles.mainButton} ${(gameActive && !isCrashed) ? styles.activeButton : ''}`} 
                        onClick={handleCashout} 
                        disabled={!gameActive || loading || isCrashed}
                    >
                        {loading ? 'Loading...' : 'Cash Out'}
                    </button>
                ) : (
                    <BetControls 
                        betAmount={betAmount}
                        onBetAmountChange={handleBetAmountChange}
                        onMultiplyAmount={handleMultiplyAmount}
                        onBet={handleBet}
                        loading={loading}
                        disabled={isBettingClosed}
                    />
                )}
            </div>
        </div>
    );
};
