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
      <Link to={'https://t.me/BiTRavesupport'} className={`${styles.support__banner} ${styles.banner}`}>
        <img src="/support.png" alt="support" />
        <span>Support</span>
      </Link>

      <div className={styles.links}>
        <Link to={'/other/faq'}>
          FAQ{' '}
          <svg width="20" height="20" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path
              d="M13.0942 10L8.08507 4.99167L6.90674 6.17L10.7401 10.0033L6.90674 13.8308L8.08507 15.0092L13.0942 10Z"
              fill="#D0D0D0"
            ></path>
          </svg>
        </Link>
      </div>

      <div className={styles.other_dev} onClick={() => copyLink()} />

      <div className={styles.other__social}>
        <a href="https://t.me/RupeXBot">
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
