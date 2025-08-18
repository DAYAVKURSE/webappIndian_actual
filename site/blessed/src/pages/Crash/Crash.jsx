import { useCallback, useEffect, useRef, useState } from 'react';
import styles from './Crash.module.scss';

import { crashPlace, crashCashout } from '@/requests';
import useStore from '@/store';

import { MoneyGameStatus } from '@/components';
import toast from 'react-hot-toast';

import { getAccessToken } from '@/utils/token-storage';
import { useWebSocketWithReconnect } from '@/hooks/useWebSocketWithReconnect.js';
import { GAME_STATES } from './types';

export const Crash = () => {
  const initData = getAccessToken();
  const { BalanceRupee, increaseBalanceRupee, decreaseBalanceRupee } = useStore();

  // Game state
  const [gameState, setGameState] = useState(GAME_STATES.WAITING);
  const [multiplier, setMultiplier] = useState(1.0);
  const [countdown, setCountdown] = useState(0);
  const [crashPoint, setCrashPoint] = useState(0);

  // Bet state
  const [betAmount, setBetAmount] = useState(100);
  const [activeBet, setActiveBet] = useState(null);
  const [queuedBet, setQueuedBet] = useState(null);
  const [autoCashoutMultiplier, setAutoCashoutMultiplier] = useState('');
  const [loading, setLoading] = useState(false);

  // Animation state
  const starAnimationRef = useRef(null);
  const starAnimationIsPlaying = useRef(null);
  const animationPhaseRef = useRef(0);
  const progressRef = useRef(0);
  const [starPosition, setStarPosition] = useState({ x: 5, y: 90 });

  // Refs

  const queuedBetRef = useRef(queuedBet);
  const activeBetRef = useRef(activeBet);
  const autoCashoutRef = useRef(autoCashoutMultiplier);

  useEffect(() => {
    queuedBetRef.current = queuedBet;
  }, [queuedBet]);

  useEffect(() => {
    activeBetRef.current = activeBet;
  }, [activeBet]);

  useEffect(() => {
    autoCashoutRef.current = autoCashoutMultiplier;
  }, [autoCashoutMultiplier]);

  const handleWebSocketMessage = useCallback((message) => {
    const { type, data } = message;

    const queuedBet = queuedBetRef.current;
    const activeBet = activeBetRef.current;

    switch (type) {
      case 'new_round':
        handleNewRound(queuedBet);
        break;
      case 'countdown_tick':
        handleCountdownTick(data);
        break;
      case 'multiplier_update':
        handleMultiplierUpdate(data);
        break;
      case 'crash':
        handleCrash(data, activeBet);
        break;
      case 'cashout':
        handleCashout(data, activeBet);
        break;
      default:
        console.log('Unknown message type:', type);
    }
  }, []);

  const handleNewRound = () => {
    setGameState(GAME_STATES.NEW_ROUND);
    setMultiplier(0);
    setCountdown(10);
    setCrashPoint(0);
    setStarPosition({ x: 5, y: 90 });

    const queuedBet = queuedBetRef.current;
    if (queuedBet) {
      setActiveBet(queuedBet);
      setQueuedBet(null);
      queuedBetRef.current = null;
      toast.success(`Queued bet ₹${queuedBet.amount} is now active!`);
    }
  };

  const handleCountdownTick = (data) => {
    setGameState(GAME_STATES.COUNTDOWN);
    setCountdown(data.seconds_left);

    const queuedBet = queuedBetRef.current;
    if (queuedBet) {
      setActiveBet(queuedBet);
      setQueuedBet(null);
      queuedBetRef.current = null;
      toast.success(`Queued bet ₹${queuedBet.amount} is now active!`);
    }
  };

  const handleMultiplierUpdate = (data) => {
    setGameState(GAME_STATES.PLAYING);
    setMultiplier(parseFloat(data.value.toFixed(2)));
    startStarAnimation();
  };

  const handleCrash = (data, activeBet) => {
    setGameState(GAME_STATES.CRASHED);
    setCrashPoint(data.crash_point);
    setMultiplier(data.crash_point);

    if (activeBet) {
      toast.error(`Crashed at ${data.crash_point.toFixed(2)}x! Lost ₹${activeBet.amount}`);
      setActiveBet(null);
    }

    stopStarAnimation();
    setStarPosition((prev) => ({ ...prev, y: 100 }));
  };

  const handleCashout = (data, activeBet) => {
    setGameState(GAME_STATES.CASHOUT);
    if (activeBet) {
      const winAmount = data.win_amount;
      toast.success(`Won ₹${winAmount.toFixed(0)} at ${data.multiplier.toFixed(2)}x!`);
      increaseBalanceRupee(winAmount);
      setActiveBet(null);
    }
  };

  // Animation functions

  const animateStar = () => {
    setStarPosition((prev) => {
      let { x, y } = prev;

      if (animationPhaseRef.current === 0) {
        const startX = 5;
        const startY = 90;
        const endX = 90;
        const endY = 20;
        const arcHeight = -20;

        progressRef.current += 0.003;
        const t = Math.min(progressRef.current, 1);

        const easeInOutCubic = (t) => t * t;

        const easedT = easeInOutCubic(t);

        x = startX + (endX - startX) * easedT;
        y = startY + (endY - startY) * easedT - arcHeight * Math.sin(Math.PI * easedT);

        if (t >= 1) {
          animationPhaseRef.current = 1;
          progressRef.current = 0;
        }
      } else if (animationPhaseRef.current === 1) {
        const centerX = 90;
        const centerY = 20;
        const radius = 3;
        const speed = 0.05;

        progressRef.current += speed;

        x = centerX + radius * Math.sin(progressRef.current);
        y = centerY + radius * Math.sin(progressRef.current) * 0.5;
      } else if (animationPhaseRef.current === 2) {
        const startX = 80;
        const startY = 20;
        const endX = 75;
        const endY = 15;
        const arcHeight = -15;

        progressRef.current += 0.01;
        const t = Math.min(progressRef.current, 1);

        x = startX + (endX - startX) * t;
        y = startY + (endY - startY) * t - arcHeight * Math.sin(Math.PI * t);

        if (t >= 1) {
          animationPhaseRef.current = 1;
          progressRef.current = 0;
        }
      }

      return { x, y };
    });

    starAnimationRef.current = requestAnimationFrame(animateStar);
  };

  const startStarAnimation = () => {
    if (starAnimationIsPlaying.current) return;
    starAnimationIsPlaying.current = true;

    cancelAnimationFrame(starAnimationRef.current);
    setStarPosition({ x: 5, y: 90 });
    animationPhaseRef.current = 0;
    progressRef.current = 0;
    starAnimationRef.current = requestAnimationFrame(animateStar);
  };

  const stopStarAnimation = () => {
    cancelAnimationFrame(starAnimationRef.current);
    starAnimationIsPlaying.current = false;
  };

  // Bet handling functions
  const handlePlaceBet = async () => {
    if (!initData || loading) return;

    if (betAmount <= 0) {
      toast.error('Bet amount must be greater than 0');
      return;
    }

    if (betAmount > BalanceRupee) {
      toast.error('Insufficient funds');
      return;
    }

    setLoading(true);

    try {
      const betData = {
        Amount: betAmount,
        CashOutMultiplier: autoCashoutMultiplier ? parseFloat(autoCashoutMultiplier) : 0,
      };

      const response = await crashPlace(betData.Amount, betData.CashOutMultiplier);

      if (!response.ok) return toast.error('Failed to place bet');

      const data = await response.json();

      if (data.queued) {
        setQueuedBet({ amount: betAmount, cashoutMultiplier: betData.CashOutMultiplier });
        decreaseBalanceRupee(betAmount);
        toast.success('Bet queued for next round!');
      } else {
        setActiveBet({ amount: betAmount, cashoutMultiplier: betData.CashOutMultiplier });
        decreaseBalanceRupee(betAmount);
        toast.success('Bet placed for current round!');
      }
    } catch (error) {
      console.error('Error placing bet:', error);
      toast.error('Failed to place bet');
    } finally {
      setLoading(false);
    }
  };

  const performCashout = async () => {
    if (!activeBet || loading) return;

    setLoading(true);

    try {
      setActiveBet(null);
      const response = await crashCashout();

      if (response.ok) {
        toast.success(`Cashout requested at ${multiplier}x`);
      } else {
        const error = await response.json();
        toast.error(error.error || 'Failed to cashout');
      }
    } catch (error) {
      console.error('Error cashing out:', error);
      toast.error('Failed to cashout');
    } finally {
      setLoading(false);
    }
  };

  // Utility functions
  const adjustBetAmount = (delta) => {
    setBetAmount((prev) => Math.max(10, prev + delta));
  };

  const multiplyBetAmount = (factor) => {
    setBetAmount((prev) => Math.max(10, Math.round(prev * factor)));
  };

  const canPlaceBet = () => {
    return !loading && !activeBet && betAmount > 0 && betAmount <= BalanceRupee;
  };

  const canCashout = () => {
    return !loading && activeBet && gameState === GAME_STATES.PLAYING;
  };

  const getStatusText = useCallback(() => {
    switch (gameState) {
      case GAME_STATES.NEW_ROUND:
        return 'New round';
      case GAME_STATES.WAITING:
        return 'Waiting for next round';
      case GAME_STATES.PLAYING:
        return 'Game in progress';
      case GAME_STATES.CASHOUT:
        return 'Cashout';
      case GAME_STATES.CRASHED:
        return `Crashed at ${crashPoint.toFixed(2)}x`;
      case GAME_STATES.COUNTDOWN:
        return `Game starts in ${countdown}s`;
      default:
        return 'Waiting for next round';
    }
  }, [gameState, countdown, crashPoint]);

  const getStatusTextButton = useCallback(() => {
    if (loading) return 'Processing...';
    if (queuedBet) return 'Bet Queued';
    if (gameState === GAME_STATES.COUNTDOWN) return 'Place Bet';
    return 'Queue Bet';
  }, [loading, queuedBet, gameState]);

  useWebSocketWithReconnect(
    `wss://rupex.io/api/ws/crashgame/live?init_data=${initData}`,
    handleWebSocketMessage
  );

  return (
    <div className={styles['crash-game']}>
      <MoneyGameStatus />

      {/* Game Area */}
      <div className={styles['game-container']}>
        {/* Grid Background */}
        <div className={styles['grid-background']} />

        {/* Status Overlay */}
        <div className={styles['status-overlay']}>
          <div className={styles['status-text']}>{getStatusText()}</div>
        </div>

        {/* Star */}
        <div
          className={`${styles.star} ${gameState === GAME_STATES.CRASHED ? styles.crashed : ''}`}
          style={{
            left: `${starPosition.x}%`,
            top: `${starPosition.y}%`,
          }}
        >
          <img src="/star.svg" alt="Star" />
        </div>

        {/* Multiplier Display */}
        <div className={styles['multiplier-display']}>{multiplier.toFixed(2)}x</div>

        {/* Active Bet Indicator */}
        {activeBet && (
          <div className={styles['active-bet-indicator']}>
            Active: ₹{activeBet.amount}
            {activeBet.cashoutMultiplier > 0 && <span> @ {activeBet.cashoutMultiplier}x</span>}
          </div>
        )}

        {/* Queued Bet Indicator */}
        {queuedBet && <div className={styles['queued-bet-indicator']}>Queued: ₹{queuedBet.amount}</div>}
      </div>

      {/* Controls */}
      <div className={styles['controls-section']}>
        {/* Auto Cashout */}
        <div className={styles['auto-cashout-section']}>
          <label>Auto Cashout Multiplier</label>
          <div className={styles['coefficient-input-container']}>
            <input
              type="number"
              min="1.01"
              step="0.01"
              value={autoCashoutMultiplier}
              onChange={(e) => setAutoCashoutMultiplier(e.target.value)}
              placeholder="e.g. 2.00"
              className={styles['auto-cashout-input']}
            />
            <div
              className={`${styles['bet-adjust-buttons']} ${styles['bet-input-container']} ${styles['quick-coefficient-buttons']}`}
            >
              <button onClick={() => setAutoCashoutMultiplier('1.5')}>1.5x</button>
              <button onClick={() => setAutoCashoutMultiplier('2.0')}>2.0x</button>
              <button onClick={() => setAutoCashoutMultiplier('5.0')}>5.0x</button>
              <button onClick={() => setAutoCashoutMultiplier('10.0')}>10x</button>
            </div>
          </div>
        </div>

        {/* Bet Amount Controls */}
        <div className={styles['bet-controls']}>
          <div className={styles['bet-amount-section']}>
            <label>Bet Amount</label>
            <div className={`${styles['bet-adjust-buttons']} ${styles['bet-input-container']}`}>
              <button
                style={{
                  fontSize: '20px',
                  fontWeight: 500,
                }}
                onClick={() => adjustBetAmount(-100)}
              >
                -
              </button>
              <input
                type="number"
                min="10"
                value={betAmount}
                style={{
                  textAlign: 'center',
                  height: '60px',
                }}
                onChange={(e) => setBetAmount(Math.max(10, parseInt(e.target.value) || 10))}
                className={styles['bet-amount-input']}
              />
              <button
                style={{
                  fontSize: '20px',
                  fontWeight: 500,
                }}
                onClick={() => adjustBetAmount(100)}
              >
                +
              </button>
            </div>

            <div className={styles['bet-multiplier-buttons']}>
              <button onClick={() => multiplyBetAmount(0.5)}>÷2</button>
              <button onClick={() => multiplyBetAmount(2)}>×2</button>
            </div>
          </div>

          {/* Main Action Button */}
          <div className={styles['main-button-container']}>
            {canCashout() ? (
              <button
                className={`${styles['main-button']} ${styles['cashout-button']}`}
                onClick={performCashout}
                disabled={loading}
              >
                {loading ? 'Processing...' : `Cashout ${multiplier.toFixed(2)}x`}
              </button>
            ) : (
              <button
                className={`${styles['main-button']} ${
                  gameState === GAME_STATES.COUNTDOWN ? styles['bet-button'] : styles['queue-button']
                }`}
                onClick={handlePlaceBet}
                disabled={!canPlaceBet() || !!queuedBet}
              >
                {getStatusTextButton()}
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
