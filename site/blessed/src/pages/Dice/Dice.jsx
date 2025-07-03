import { useRef, useState, useEffect } from 'react';
import { isMobile } from 'react-device-detect';
import styles from './Dice.module.scss';
import Lottie from 'lottie-react';
import animationData from './dice.json';
import { rollDice } from '@/requests';
import { toast } from 'react-hot-toast';
import useStore from '@/store';
import { MoneyGameStatus } from '@/components';

export const Dice = () => {
  const { increaseBalanceRupee, decreaseBalanceRupee } = useStore();
  const [bet, setBet] = useState(100);
  const [percent, setPercent] = useState(50);
  const [isNumberVisible, setIsNumberVisible] = useState(false);
  const [number, setNumber] = useState(null);
  const [isAnimating, setIsAnimating] = useState(false);
  const lottieRef = useRef();
  const [activeButton, setActiveButton] = useState('less');
  const [isLoading, setIsLoading] = useState(false);

  const calculateRange = () => {
    const maxValue = 999999;
    const winRange = Math.round((percent / 100) * (maxValue + 1));
    const lossRange = Math.round(maxValue - winRange + 1);

    return {
      less: `0 - ${winRange.toLocaleString()}`,
      more: `${lossRange.toLocaleString()} - ${maxValue.toLocaleString()}`,
    };
  };

  const handlePercentChange = (value) => {
    let newValue = parseInt(value, 10);

    if (isNaN(newValue)) {
      setPercent(5);
    } else if (newValue < 5) {
      setPercent(5);
    } else if (newValue > 95) {
      setPercent(95);
    } else {
      setPercent(newValue);
    }
  };

  const validateNumber = (num) => {
    const minValue = 0;
    const maxValue = 999999;
    if (num < minValue || num > maxValue) {
      toast.error(`Number out of range! It should be between ${minValue} and ${maxValue}.`);
      return false;
    }
    return true;
  };

  useEffect(() => {
    let timer;
    setIsLoading(false);
    if (isNumberVisible) {
      timer = setTimeout(() => {
        setIsAnimating(true);
      }, 2000);
    }

    return () => clearTimeout(timer);
  }, [isNumberVisible]);

  const handleBet = async (direction) => {
    setIsAnimating(false);
    setIsNumberVisible(false);
    setActiveButton(direction);

    setTimeout(async () => {
      setIsNumberVisible(true);

      try {
        const response = await rollDice(bet, percent, direction);

        if (response.status === 200) {
          if (lottieRef.current) {
            lottieRef.current.stop();
            lottieRef.current.play();
          }

          const data = await response.json();
          const rolledNumber = data.result.number;
          decreaseBalanceRupee(bet);

          if (validateNumber(rolledNumber)) {
            setNumber(rolledNumber);
          }

          setTimeout(() => {
            if (data.result.won === true) {
              toast.success(`You won! (+${Math.trunc(data.payout)})`);
              increaseBalanceRupee(data.payout);
            } else {
              toast.error(`You lost! (-${bet})`);
            }
          }, 2000);
        } else if (response.status === 400) {
          toast.error('Invalid request body');
        } else if (response.status === 402) {
          toast.error('Insufficient balance');
        } else {
          toast.error('An error occurred');
        }
      } catch (err) {
        toast.error(err.message);
      }
    }, 500);
  };

  const { less, more } = calculateRange();

  useEffect(() => {
    const slider = document.querySelector('#slider');

    slider.addEventListener('input', function () {
      this.style.setProperty('--value', `${this.value}%`);
    });
  }, []);

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
    <div className={styles.dice}>
      <MoneyGameStatus />
      <div className={styles.dice_bet_container}>
        <p className={`${styles.dice_number} ${isAnimating ? styles.fadeIn : styles.fadeOut}`}>
          {number !== null ? number : ''}
        </p>
        <div className={styles.dice_lottie}>
          <Lottie lottieRef={lottieRef} animationData={animationData} loop={false} />
        </div>

        <div className={styles.range_buttons}>
          <button
            className={`${styles.range_button_less} ${activeButton === 'less' ? styles.active : ''}`}
            onClick={() => handleBet('less')}
          >
            <img src="/trading_arrow.svg" alt="Down" />
          </button>
          <button
            className={`${styles.range_button_more} ${activeButton === 'more' ? styles.active : ''}`}
            onClick={() => handleBet('more')}
          >
            <img src="/trading_arrow.svg" alt="Up" />
          </button>
        </div>

        <div className={styles.range_values}>
          <p>{less}</p>
          <p>{more}</p>
        </div>

        {/* <Amount bet={bet} setBet={setBet} /> */}
        <div className={styles.bet_control_group}>
          <div className={styles.bet_control}>
            <div className={styles.bet_amount}>
              {/* <span>{bet}</span> */}
              <input
                className={styles.amount__input}
                value={bet}
                onChange={(e) => {
                  const value = e.target.value.replace(/\D/g, ''); // Удаляем все нецифровые символы
                  setBet(value);
                }}
              ></input>
              <div className={styles.amount_controls}>
                <button className={styles.minus_button} onClick={handleDecreaseBet} disabled={isLoading}>
                  <img src="/trading_min.svg" alt="trading_min" />
                </button>
                <button className={styles.plus_button} onClick={handleIncreaseBet} disabled={isLoading}>
                  <img src="/trading_plus.svg" alt="trading_plus" />
                </button>
              </div>
            </div>

            <div className={styles.bet_multipliers}>
              <button className={styles.divide_button} onClick={handleDivideBet} disabled={isLoading}>
                /2
              </button>
              <button className={styles.multiply_button} onClick={handleMultiplyBet} disabled={isLoading}>
                ×2
              </button>
            </div>
          </div>

          <button className={styles.bet_button} onClick={() => handleBet(activeButton)} disabled={isLoading}>
            {isLoading ? (
              '...'
            ) : (
              <span>
                Bet{' '}
                <svg width="24" height="24" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path
                    d="M13.0942 10L8.08507 4.99167L6.90674 6.17L10.7401 10.0033L6.90674 13.8308L8.08507 15.0092L13.0942 10Z"
                    fill="#FFFFFF"
                  ></path>
                </svg>
              </span>
            )}
          </button>
        </div>
        <div className={styles.percent_container}>
          <div className={styles.percent_label}>Percent</div>
          <div className={styles.percent_value}>{percent}%</div>

          <input
            type="range"
            min="5"
            max="95"
            value={percent}
            onChange={(e) => handlePercentChange(e.target.value)}
            className={styles.slider}
            id="slider"
            disabled={isLoading}
          />
        </div>
      </div>
    </div>
  );
};
