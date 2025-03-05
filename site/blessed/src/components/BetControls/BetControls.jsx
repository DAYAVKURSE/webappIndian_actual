import React from 'react';
import styles from './BetControls.module.scss';

const BetControls = ({ 
  betAmount, 
  onBetAmountChange, 
  onMultiplyAmount, 
  onBet, 
  loading, 
  disabled 
}) => {
  const handleDecrement = () => {
    onBetAmountChange(betAmount - 100);
  };

  const handleIncrement = () => {
    onBetAmountChange(betAmount + 100);
  };

  const handleHalve = () => {
    onMultiplyAmount(0.5);
  };

  const handleDouble = () => {
    onMultiplyAmount(2);
  };

  return (
    <div className={styles.betControls}>
      <div className={styles.betAmountContainer}>
        <div className={styles.betAmount}>
          <span>{betAmount} ₹</span>
          <div className={styles.amountButtons}>
            <button 
              className={styles.amountButton} 
              onClick={handleDecrement}
              aria-label="Decrease amount"
            >
              -
            </button>
            <button 
              className={styles.amountButton} 
              onClick={handleIncrement}
              aria-label="Increase amount"
            >
              +
            </button>
          </div>
        </div>
      </div>

      <div className={styles.actionsContainer}>
        <div className={styles.multiplierButtons}>
          <button 
            className={styles.multiplierButton} 
            onClick={handleHalve}
            aria-label="Halve amount"
          >
            /2
          </button>
          <button 
            className={styles.multiplierButton} 
            onClick={handleDouble}
            aria-label="Double amount"
          >
            ×2
          </button>
        </div>

        <button 
          className={styles.betButton} 
          onClick={onBet}
          disabled={disabled || loading}
        >
          {loading ? 'Loading...' : 'Bet'}
        </button>
      </div>
    </div>
  );
};

export default BetControls; 