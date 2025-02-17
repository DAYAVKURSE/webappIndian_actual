import { useLocation } from "react-router-dom";
import styles from "./Header.module.scss";
import { Button } from "@/components";
import useStore from "@/store";

export const Header = () => {
    const { BalanceRupee } = useStore();
    const location = useLocation();
    const isGamesPath = location.pathname.includes("games/");

    return (
        <>
            <div className={styles.header__spacer} />
            <header className={styles.header}>
                <div className={styles.header__buttons} style={{width: "100%", justifyContent: "space-between"}}>
                    <Button to="profile" icon="/32=account_circle.svg" color="lowBlack" filter="none" circle />
                    {isGamesPath && <Button icon="/24=rupee.svg" label={Math.trunc(BalanceRupee) || '0'} color="lowBlack" filter="brightness(0) saturate(100%) invert(88%) sepia(64%) saturate(333%) hue-rotate(16deg) brightness(105%) contrast(102%)" />}
                    <Button to="wallet" icon="/32=account_balance_wallet.svg" color="lowBlack" filter="none" circle />
                </div>
                {/* <div className={styles.header__buttons}>
                    <Button to="other" icon="/32=more_horiz.svg" color="lowBlack" filter="none" circle />
                </div> */}
            </header>
        </>
    );
};
