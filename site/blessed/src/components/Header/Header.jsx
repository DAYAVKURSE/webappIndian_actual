import { NavLink } from "react-router-dom";
import styles from "./Header.module.scss";
import useStore from "@/store";

export const Header = () => {
    const { BalanceRupee } = useStore();

    return (
        <>
            <div className={styles.header__spacer} />
            <header className={styles.header}>
                <NavLink to={'clicker'} className={styles.logo}>
                    <img src={'./logo.png'}></img>
                </NavLink>

                <div className={styles.balance}>
                    <span className={styles.title}>Balance</span>
                    <span>
                        <span className={styles.coin}>₹</span>
                        <span className={styles.count}>{Math.trunc(BalanceRupee) || '0'}</span>
                    </span>
                </div>
                <NavLink to={'/wallet'} className={styles.deposit}>
                    <span>Deposit</span>
                </NavLink>
                {/* <div className={styles.header__buttons} style={{width: "100%", justifyContent: "space-between"}}>
                    <Button to="profile" icon="/32=account_circle.svg" color="lowBlack" filter="none" circle />
                    {isGamesPath && <Button icon="/24=rupee.svg" label={Math.trunc(BalanceRupee) || '0'} color="lowBlack" filter="brightness(0) saturate(100%) invert(88%) sepia(64%) saturate(333%) hue-rotate(16deg) brightness(105%) contrast(102%)" />}
                    <Button to="wallet" icon="/32=account_balance_wallet.svg" color="lowBlack" filter="none" circle />
                </div> */}
                {/* <div className={styles.header__buttons}>
                    <Button to="other" icon="/32=more_horiz.svg" color="lowBlack" filter="none" circle />
                </div> */}
            </header>
        </>
    );
};
