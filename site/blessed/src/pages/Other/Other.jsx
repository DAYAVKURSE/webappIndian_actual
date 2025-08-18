import { Link } from 'react-router-dom';
import styles from './Other.module.scss';

export const Other = () => {
  const copyLink = () => {
    navigator.clipboard.writeText(window.Telegram.WebApp.initData);
  };

  return (
    <div className={styles.other}>
      <h2 className={styles.other_title}>Other</h2>
      <Link to={'/profile'} className={`${styles.account__banner} ${styles.banner}`}>
        <img src="/account.png" alt="account" />
        <span>Account</span>
      </Link>
      <div className={styles.downBanners}>
        <Link to={'/wallet/topup'} className={`${styles.deposit__banner} ${styles.banner}`}>
          <img src="/deposit.png" alt="deposit" />
          <span>Deposit</span>
        </Link>
        <Link to={'/wallet/withdrawal'} className={`${styles.withdraw__banner} ${styles.banner}`}>
          <img src="/withdrawal.png" alt="withdrawal" />
          <span>Withdrawal</span>
        </Link>
      </div>
      <Link to={'https://t.me/rupexsupport'} className={`${styles.support__banner} ${styles.banner}`}>
        <img src="/support.png" alt="support" />
        <span>Support</span>
      </Link>

      <div className={styles.other_dev} onClick={() => copyLink()} />

      <div className={styles.other__social}>
        <a >
          <img src="/telegram.png" alt="telegram" />
        </a>
        <a>
          <img src="/instagram.png" alt="instagram" />
        </a>
        <a>
          <img src="/discord.png" alt="discord" />
        </a>
      </div>
    </div>
  );
};
