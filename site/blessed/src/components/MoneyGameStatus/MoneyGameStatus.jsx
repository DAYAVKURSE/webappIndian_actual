import styles from './MoneyGameStatus.module.scss';
import useStore from '@/store';
import { useNavigate } from 'react-router-dom';
export const MoneyGameStatus = () => {
  const navigate = useNavigate();
  const { BalanceRupee } = useStore();

  return (
    <div className={styles.status_money}>
      <button
        className={styles.back}
        onClick={() => {
          navigate(-1);
        }}
      >
        <img src="/mail-reply.svg" alt="mail-reply" />
        Back
      </button>
      <div className={styles.balance}>
        <img src="/coin.png" alt="coin" />
        <span className={styles.count}>{Math.trunc(BalanceRupee) || '0'} ₹</span>
      </div>
    </div>
  );
};
