import { useRef, useEffect } from "react";
import styles from "./Wallet.module.scss";
import { Accordion } from "@/components";
export const Wallet = () => {
    const moneyInputRef = useRef(null);

    useEffect(() => {
        const moneyInput = moneyInputRef.current;

        const resizeInput = () => {
            moneyInput.style.width = moneyInput.value.length + "ch";
        };

        if (moneyInput) {
            moneyInput.addEventListener('input', resizeInput);
            resizeInput();

            return () => {
                moneyInput.removeEventListener('input', resizeInput);
            };
        }
    }, []);

    return (
        <div className={styles.wallet}>
            <h2 className={styles.wallet_title}>Wallet</h2>
            <p className={styles.wallet_description}>Deposit, exchange or make withdrawals</p>
            <Accordion
                to="topup"
                icon="/24=book_5.svg"
                title="Top up"
                filter="brightness(0) saturate(100%) invert(91%) sepia(42%) saturate(477%) hue-rotate(16deg) brightness(103%) contrast(106%)"
            />
            <Accordion
                to="withdrawal"
                icon="/24=refresh.svg"
                title="Withdrawal"
                filter="brightness(0) saturate(100%) invert(72%) sepia(15%) saturate(4631%) hue-rotate(197deg) brightness(107%) contrast(107%)"
            />
            <Accordion
                to="exchange"
                icon="/24=exchange.svg"
                title="Exchange"
                filter="brightness(0) saturate(100%) invert(100%) sepia(51%) saturate(2592%) hue-rotate(310deg) brightness(103%) contrast(104%)"
            />
        </div>
    );
};
