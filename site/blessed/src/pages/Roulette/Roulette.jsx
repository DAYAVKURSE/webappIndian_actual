import { useEffect, useState } from 'react';
import styles from './Roulette.module.scss';
import { roulettePlaceBet, rouletteGetHistory, getMe } from '@/requests';
import toast from 'react-hot-toast';
import useStore from '@/store';
import { MoneyGameStatus } from '@/components';

const sectors = 15; // Количество секторов

export const Roulette = () => {
  const { increaseBalanceRupee, decreaseBalanceRupee } = useStore();
  const [bet, setBet] = useState(100);
  const [wins, setWins] = useState([]);
  const [isBettingOpen, setIsBettingOpen] = useState(true);
  const [rotation, setRotation] = useState(0);
  const [mustSpin, setMustSpin] = useState(false);

  useEffect(() => {
    const fetchHistory = async () => {
      try {
        const data = await rouletteGetHistory();
        const lastWins = data.map((result) => result.WinningColor);
        setWins(lastWins);
      } catch (error) {
        console.error('Error fetching game history:', error);
      }
    };

    fetchHistory();
    getMe();
  }, []);

  const handleBet = async (color) => {
    if (!isBettingOpen) {
      toast.error('Betting is closed');
      return;
    }
    try {
      const response = await roulettePlaceBet(bet, color);
      if (response.status === 200) {
        const data = await response.json();

        if (color === 'black') color = 'yellow';
        toast.success(`Bet placed: ${bet} on ${color}`);
        decreaseBalanceRupee(bet);

        // Вычисляем угол вращения
        const newPrizeSectorId = data.winning_number;
        const sectorAngle = 360 / sectors;
        const randomExtraTurns = Math.floor(Math.random() * 5) + 3; // Добавим случайные обороты
        const newRotation = randomExtraTurns * 360 + (sectors - newPrizeSectorId) * sectorAngle;

        setIsBettingOpen(false);
        setMustSpin(true);
        setRotation(newRotation);

        setTimeout(() => {
          setWins((prevWins) => [data.winning_color, ...prevWins].slice(0, 10));
          setIsBettingOpen(true);
          setMustSpin(false);

          if (data.outcome === 'win') {
            toast.success(`You won! +${data.payout}`);
            increaseBalanceRupee(data.payout);
          } else {
            toast.error(`You lost!`);
          }
        }, 6000);
      } else {
        const data = await response.json();
        toast.error(data.error);
      }
    } catch (err) {
      console.error(err.message);
    }
  };

  const handleIncreaseBet = () => {
    setBet((prev) => prev + 1);
  };

  const handleDecreaseBet = () => {
    setBet((prev) => (prev > 1 ? prev - 1 : 1));
  };

  const handleDivideBet = () => {
    setBet((prev) => Math.max(Math.floor(prev / 2), 1));
  };

  const handleMultiplyBet = () => {
    setBet((prev) => prev * 2);
  };

  return (
    <div className={styles.roulette}>
      <MoneyGameStatus />
      <div className={styles.roulette__wheel}>
        <img
          src="/wheel.svg"
          alt="Roulette Wheel"
          className={styles.wheel}
          style={{ transform: `rotate(${rotation}deg)`, transition: mustSpin ? '6s ease-out' : 'none' }}
        />
        <img src="/roulette_arrow.svg" alt="Arrow" className={styles.arrow} />
      </div>

      <div className={styles.roulette__bets}>
        <p className={styles.roulette__bets_title}>Last wins</p>
        <div className={styles.roulette__wins_container}>
          {wins.length > 0 ? (
            wins.map((win, index) => {
              const displayWinColor = win === 'black' ? 'yellow' : win;
              return (
                <div
                  key={index}
                  className={`${styles.roulette__bets__button} ${
                    styles[`roulette__bets__button_${displayWinColor.toLowerCase()}`]
                  }`}
                >
                  <img src="/24=poker_chip.svg" alt="" />
                  {displayWinColor.charAt(0).toUpperCase() + displayWinColor.slice(1)}
                </div>
              );
            })
          ) : (
            <p className={styles.nowins} style={{ width: '100%', textAlign: 'center', color: '#DADADA' }}>
              No wins
            </p>
          )}
        </div>
      </div>

      <div className={styles.roulette__bets_wrapper}>
        {/* <Amount bet={bet} setBet={setBet} /> */}
        <div className={styles.roulette__input}>
          <span className={styles.leftContainer}>
            <input
              className={styles.amount__input}
              value={bet}
              onChange={(e) => {
                const value = e.target.value.replace(/\D/g, ''); // Удаляем все нецифровые символы
                setBet(value);
              }}
            ></input>
          </span>
          <span className={styles.rightContainer}>
            <div onClick={handleDecreaseBet}>
              <img src="/trading_min.svg" alt="trading_min" />
            </div>
            <div onClick={handleIncreaseBet}>
              <img src="/trading_plus.svg" alt="trading_plus" />
            </div>
          </span>
        </div>
        <div className={styles.roulette__downbuttons}>
          <button onClick={handleDivideBet} className={styles.decrimentBet}>
            /2
          </button>
          <button onClick={handleMultiplyBet} className={styles.incrimentBet}>
            x2
          </button>
        </div>
      </div>

      <div className={`${styles.roulette__bets_container} ${styles.roulette__bets_container_buttons}`}>
        <div
          className={`${styles.roulette__bets__button} ${styles.roulette__bets__button_red}`}
          onClick={() => handleBet('red')}
        >
          <svg width="16" height="16" viewBox="0 0 16 16" fill="#FFFFFF" xmlns="http://www.w3.org/2000/svg">
            <g clipPath="url(#clip0_149_468)">
              <path
                d="M13 1C13.5304 1 14.0391 1.21071 14.4142 1.58579C14.7893 1.96086 15 2.46957 15 3V13C15 13.5304 14.7893 14.0391 14.4142 14.4142C14.0391 14.7893 13.5304 15 13 15H3C2.46957 15 1.96086 14.7893 1.58579 14.4142C1.21071 14.0391 1 13.5304 1 13V3C1 2.46957 1.21071 1.96086 1.58579 1.58579C1.96086 1.21071 2.46957 1 3 1H13ZM3 0C2.20435 0 1.44129 0.31607 0.87868 0.87868C0.31607 1.44129 0 2.20435 0 3L0 13C0 13.7956 0.31607 14.5587 0.87868 15.1213C1.44129 15.6839 2.20435 16 3 16H13C13.7956 16 14.5587 15.6839 15.1213 15.1213C15.6839 14.5587 16 13.7956 16 13V3C16 2.20435 15.6839 1.44129 15.1213 0.87868C14.5587 0.31607 13.7956 0 13 0L3 0Z"
                fill="white"
              />
              <path
                d="M5.5 4C5.5 4.39782 5.34196 4.77936 5.06066 5.06066C4.77936 5.34196 4.39782 5.5 4 5.5C3.60218 5.5 3.22064 5.34196 2.93934 5.06066C2.65804 4.77936 2.5 4.39782 2.5 4C2.5 3.60218 2.65804 3.22064 2.93934 2.93934C3.22064 2.65804 3.60218 2.5 4 2.5C4.39782 2.5 4.77936 2.65804 5.06066 2.93934C5.34196 3.22064 5.5 3.60218 5.5 4ZM13.5 4C13.5 4.39782 13.342 4.77936 13.0607 5.06066C12.7794 5.34196 12.3978 5.5 12 5.5C11.6022 5.5 11.2206 5.34196 10.9393 5.06066C10.658 4.77936 10.5 4.39782 10.5 4C10.5 3.60218 10.658 3.22064 10.9393 2.93934C11.2206 2.65804 11.6022 2.5 12 2.5C12.3978 2.5 12.7794 2.65804 13.0607 2.93934C13.342 3.22064 13.5 3.60218 13.5 4ZM13.5 12C13.5 12.3978 13.342 12.7794 13.0607 13.0607C12.7794 13.342 12.3978 13.5 12 13.5C11.6022 13.5 11.2206 13.342 10.9393 13.0607C10.658 12.7794 10.5 12.3978 10.5 12C10.5 11.6022 10.658 11.2206 10.9393 10.9393C11.2206 10.658 11.6022 10.5 12 10.5C12.3978 10.5 12.7794 10.658 13.0607 10.9393C13.342 11.2206 13.5 11.6022 13.5 12ZM5.5 12C5.5 12.3978 5.34196 12.7794 5.06066 13.0607C4.77936 13.342 4.39782 13.5 4 13.5C3.60218 13.5 3.22064 13.342 2.93934 13.0607C2.65804 12.7794 2.5 12.3978 2.5 12C2.5 11.6022 2.65804 11.2206 2.93934 10.9393C3.22064 10.658 3.60218 10.5 4 10.5C4.39782 10.5 4.77936 10.658 5.06066 10.9393C5.34196 11.2206 5.5 11.6022 5.5 12ZM9.5 8C9.5 8.39782 9.34196 8.77936 9.06066 9.06066C8.77936 9.34196 8.39782 9.5 8 9.5C7.60218 9.5 7.22064 9.34196 6.93934 9.06066C6.65804 8.77936 6.5 8.39782 6.5 8C6.5 7.60218 6.65804 7.22064 6.93934 6.93934C7.22064 6.65804 7.60218 6.5 8 6.5C8.39782 6.5 8.77936 6.65804 9.06066 6.93934C9.34196 7.22064 9.5 7.60218 9.5 8Z"
                fill="white"
              />
            </g>
            <defs>
              <clipPath id="clip0_149_468">
                <rect width="16" height="16" fill="#FFFFFF" />
              </clipPath>
            </defs>
          </svg>
          Red
        </div>
        <div
          className={`${styles.roulette__bets__button} ${styles.roulette__bets__button_green}`}
          onClick={() => handleBet('green')}
        >
          <svg width="16" height="16" viewBox="0 0 16 16" fill="#FFFFFF" xmlns="http://www.w3.org/2000/svg">
            <g clipPath="url(#clip0_149_468)">
              <path
                d="M13 1C13.5304 1 14.0391 1.21071 14.4142 1.58579C14.7893 1.96086 15 2.46957 15 3V13C15 13.5304 14.7893 14.0391 14.4142 14.4142C14.0391 14.7893 13.5304 15 13 15H3C2.46957 15 1.96086 14.7893 1.58579 14.4142C1.21071 14.0391 1 13.5304 1 13V3C1 2.46957 1.21071 1.96086 1.58579 1.58579C1.96086 1.21071 2.46957 1 3 1H13ZM3 0C2.20435 0 1.44129 0.31607 0.87868 0.87868C0.31607 1.44129 0 2.20435 0 3L0 13C0 13.7956 0.31607 14.5587 0.87868 15.1213C1.44129 15.6839 2.20435 16 3 16H13C13.7956 16 14.5587 15.6839 15.1213 15.1213C15.6839 14.5587 16 13.7956 16 13V3C16 2.20435 15.6839 1.44129 15.1213 0.87868C14.5587 0.31607 13.7956 0 13 0L3 0Z"
                fill="white"
              />
              <path
                d="M5.5 4C5.5 4.39782 5.34196 4.77936 5.06066 5.06066C4.77936 5.34196 4.39782 5.5 4 5.5C3.60218 5.5 3.22064 5.34196 2.93934 5.06066C2.65804 4.77936 2.5 4.39782 2.5 4C2.5 3.60218 2.65804 3.22064 2.93934 2.93934C3.22064 2.65804 3.60218 2.5 4 2.5C4.39782 2.5 4.77936 2.65804 5.06066 2.93934C5.34196 3.22064 5.5 3.60218 5.5 4ZM13.5 4C13.5 4.39782 13.342 4.77936 13.0607 5.06066C12.7794 5.34196 12.3978 5.5 12 5.5C11.6022 5.5 11.2206 5.34196 10.9393 5.06066C10.658 4.77936 10.5 4.39782 10.5 4C10.5 3.60218 10.658 3.22064 10.9393 2.93934C11.2206 2.65804 11.6022 2.5 12 2.5C12.3978 2.5 12.7794 2.65804 13.0607 2.93934C13.342 3.22064 13.5 3.60218 13.5 4ZM13.5 12C13.5 12.3978 13.342 12.7794 13.0607 13.0607C12.7794 13.342 12.3978 13.5 12 13.5C11.6022 13.5 11.2206 13.342 10.9393 13.0607C10.658 12.7794 10.5 12.3978 10.5 12C10.5 11.6022 10.658 11.2206 10.9393 10.9393C11.2206 10.658 11.6022 10.5 12 10.5C12.3978 10.5 12.7794 10.658 13.0607 10.9393C13.342 11.2206 13.5 11.6022 13.5 12ZM5.5 12C5.5 12.3978 5.34196 12.7794 5.06066 13.0607C4.77936 13.342 4.39782 13.5 4 13.5C3.60218 13.5 3.22064 13.342 2.93934 13.0607C2.65804 12.7794 2.5 12.3978 2.5 12C2.5 11.6022 2.65804 11.2206 2.93934 10.9393C3.22064 10.658 3.60218 10.5 4 10.5C4.39782 10.5 4.77936 10.658 5.06066 10.9393C5.34196 11.2206 5.5 11.6022 5.5 12ZM9.5 8C9.5 8.39782 9.34196 8.77936 9.06066 9.06066C8.77936 9.34196 8.39782 9.5 8 9.5C7.60218 9.5 7.22064 9.34196 6.93934 9.06066C6.65804 8.77936 6.5 8.39782 6.5 8C6.5 7.60218 6.65804 7.22064 6.93934 6.93934C7.22064 6.65804 7.60218 6.5 8 6.5C8.39782 6.5 8.77936 6.65804 9.06066 6.93934C9.34196 7.22064 9.5 7.60218 9.5 8Z"
                fill="white"
              />
            </g>
            <defs>
              <clipPath id="clip0_149_468">
                <rect width="16" height="16" fill="#FFFFFF" />
              </clipPath>
            </defs>
          </svg>
          Green
        </div>
        <div
          className={`${styles.roulette__bets__button} ${styles.roulette__bets__button_yellow}`}
          onClick={() => handleBet('black')}
        >
          <svg width="16" height="16" viewBox="0 0 16 16" fill="#FFFFFF" xmlns="http://www.w3.org/2000/svg">
            <g clipPath="url(#clip0_149_468)">
              <path
                d="M13 1C13.5304 1 14.0391 1.21071 14.4142 1.58579C14.7893 1.96086 15 2.46957 15 3V13C15 13.5304 14.7893 14.0391 14.4142 14.4142C14.0391 14.7893 13.5304 15 13 15H3C2.46957 15 1.96086 14.7893 1.58579 14.4142C1.21071 14.0391 1 13.5304 1 13V3C1 2.46957 1.21071 1.96086 1.58579 1.58579C1.96086 1.21071 2.46957 1 3 1H13ZM3 0C2.20435 0 1.44129 0.31607 0.87868 0.87868C0.31607 1.44129 0 2.20435 0 3L0 13C0 13.7956 0.31607 14.5587 0.87868 15.1213C1.44129 15.6839 2.20435 16 3 16H13C13.7956 16 14.5587 15.6839 15.1213 15.1213C15.6839 14.5587 16 13.7956 16 13V3C16 2.20435 15.6839 1.44129 15.1213 0.87868C14.5587 0.31607 13.7956 0 13 0L3 0Z"
                fill="white"
              />
              <path
                d="M5.5 4C5.5 4.39782 5.34196 4.77936 5.06066 5.06066C4.77936 5.34196 4.39782 5.5 4 5.5C3.60218 5.5 3.22064 5.34196 2.93934 5.06066C2.65804 4.77936 2.5 4.39782 2.5 4C2.5 3.60218 2.65804 3.22064 2.93934 2.93934C3.22064 2.65804 3.60218 2.5 4 2.5C4.39782 2.5 4.77936 2.65804 5.06066 2.93934C5.34196 3.22064 5.5 3.60218 5.5 4ZM13.5 4C13.5 4.39782 13.342 4.77936 13.0607 5.06066C12.7794 5.34196 12.3978 5.5 12 5.5C11.6022 5.5 11.2206 5.34196 10.9393 5.06066C10.658 4.77936 10.5 4.39782 10.5 4C10.5 3.60218 10.658 3.22064 10.9393 2.93934C11.2206 2.65804 11.6022 2.5 12 2.5C12.3978 2.5 12.7794 2.65804 13.0607 2.93934C13.342 3.22064 13.5 3.60218 13.5 4ZM13.5 12C13.5 12.3978 13.342 12.7794 13.0607 13.0607C12.7794 13.342 12.3978 13.5 12 13.5C11.6022 13.5 11.2206 13.342 10.9393 13.0607C10.658 12.7794 10.5 12.3978 10.5 12C10.5 11.6022 10.658 11.2206 10.9393 10.9393C11.2206 10.658 11.6022 10.5 12 10.5C12.3978 10.5 12.7794 10.658 13.0607 10.9393C13.342 11.2206 13.5 11.6022 13.5 12ZM5.5 12C5.5 12.3978 5.34196 12.7794 5.06066 13.0607C4.77936 13.342 4.39782 13.5 4 13.5C3.60218 13.5 3.22064 13.342 2.93934 13.0607C2.65804 12.7794 2.5 12.3978 2.5 12C2.5 11.6022 2.65804 11.2206 2.93934 10.9393C3.22064 10.658 3.60218 10.5 4 10.5C4.39782 10.5 4.77936 10.658 5.06066 10.9393C5.34196 11.2206 5.5 11.6022 5.5 12ZM9.5 8C9.5 8.39782 9.34196 8.77936 9.06066 9.06066C8.77936 9.34196 8.39782 9.5 8 9.5C7.60218 9.5 7.22064 9.34196 6.93934 9.06066C6.65804 8.77936 6.5 8.39782 6.5 8C6.5 7.60218 6.65804 7.22064 6.93934 6.93934C7.22064 6.65804 7.60218 6.5 8 6.5C8.39782 6.5 8.77936 6.65804 9.06066 6.93934C9.34196 7.22064 9.5 7.60218 9.5 8Z"
                fill="white"
              />
            </g>
            <defs>
              <clipPath id="clip0_149_468">
                <rect width="16" height="16" fill="#FFFFFF" />
              </clipPath>
            </defs>
          </svg>
          Yellow
        </div>
      </div>
    </div>
  );
};
