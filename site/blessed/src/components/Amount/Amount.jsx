import { useEffect, useState } from "react";
import styles from "./Amount.module.scss";
import useStore from "@/store";

export const Amount = ({ bet, setBet }) => {
    const [inputValue, setInputValue] = useState(bet);
    const { BalanceRupee } = useStore();

    useEffect(() => {
        setInputValue(bet);
    }, [bet]);

    const increaseBet = (value) => {
        setBet((prevBet) => prevBet + value);
    };

    const decreaseBet = (value) => {
        setBet((prevBet) => Math.max(prevBet - value, 10));
    };

    const handleInputChange = (e) => {
        const value = e.target.value.replace(/\D/g, '');
        setInputValue(value ? parseInt(value, 10) : '');
    };

    const handleInputBlur = () => {
        const value = inputValue ? Math.max(parseInt(inputValue, 10), 10) : 10;
        setBet(value);
        setInputValue(value);
    };

    const handleKeyDown = (e) => {
        if (e.key === 'Enter') {
            e.target.blur();
        }
    };

    return (
        <div className={styles.amount}>
            <input 
                className={styles.amount_input} 
                type="number"
                inputMode="numeric"
                pattern="[0-9]*"
                enterKeyHint="done"
                value={inputValue ? inputValue.toFixed(0) : ''} 
                onChange={handleInputChange} 
                onBlur={handleInputBlur}
                onKeyDown={handleKeyDown}
            />
            <div className={styles.amount_button_container}>
                <button className={`${styles.amount_button} ${styles.amount_button_icon}`} onClick={() => setBet(10)}>
                    <img src="/24=delete.svg" alt="delete" style={{maxWidth: "fit-content"}} />
                </button>
                <button className={styles.amount_button} onClick={() => decreaseBet(10)}>-10</button>
                <button className={styles.amount_button} onClick={() => decreaseBet(50)}>-50</button>
                <button className={styles.amount_button} onClick={() => increaseBet(10)}>+10</button>
                <button className={styles.amount_button} onClick={() => increaseBet(50)}>+50</button>
                <button className={styles.amount_button} onClick={() => setBet(BalanceRupee)}>all</button>
            </div>
        </div>
    );
};