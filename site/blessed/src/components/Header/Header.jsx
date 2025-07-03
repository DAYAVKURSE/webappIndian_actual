import { NavLink, useLocation } from 'react-router-dom';
import styles from './Header.module.scss';
import useStore from '@/store';

export const Header = () => {
  const location = useLocation();

  const disabledHeader = (pathname) => {
    switch (pathname) {
      case '/trading':
        return true;
      case '/profile':
        return true;
      case '/wallet/topup':
        return true;
      case '/wallet/withdrawal':
        return true;
      case '/games':
        return true;
      case '/games/nvuti':
        return true;
      case '/games/crash':
        return true;
      case '/other':
        return true;
      case '/games/dice':
        return true;
      case '/games/roulette':
        return true;
      default:
        return false;
    }
  };

  const { BalanceRupee } = useStore();

  if (disabledHeader(location.pathname)) {
    return <div className={styles.header__spacer} />;
  }

  return (
    <>
      <div className={styles.header__spacer} />
      <header className={styles.header}>
        <div className={styles.headerContainer}>
          <NavLink to={'clicker'} className={styles.logo}>
            <img src={'/logo.png'}></img>
          </NavLink>

          <div className={styles.rightContainer}>
            <div className={styles.balance}>
              <span className={styles.title}>Balance</span>
              <span>
                <span className={styles.coin}>₹</span>
                <span className={styles.count}>{Math.trunc(BalanceRupee) || '0'}</span>
              </span>
            </div>
            <NavLink to={'/wallet/topup'} className={styles.deposit}>
              <span>Deposit</span>
            </NavLink>
          </div>
        </div>
      </header>
    </>
  );
};
