import { useState } from "react";
import styles from "./Exchange.module.scss";
import { Slider, Button } from "@/components";
import useStore from "@/store";
import { exchange, getMe } from "@/requests";
import { toast } from "react-hot-toast";

export const Exchange = () => {
    const [percent, setPercent] = useState(0);
    const [amount, setAmount] = useState(100);
    const { BalanceRupee, BalanceBi } = useStore();

    const handleAmountChange = (e) => {
        const value = e.target.value;
        if (/^\d*$/.test(value)) {
            setAmount(Number(value));
        }
    };

    const handleExchange = async () => {
        try {
            if (percent === 0) {
                toast.error("Please select the amount of BCoins to exchange");
                getMe();
                return;
            }
			const response = await exchange(percent);
			if (response.status === 200) {
				toast.success(`Successfully exchanged ${percent} BCoins`);
                getMe();
                setPercent(0);
                setAmount(0);
			} else {
				const data = await response.json();
				toast.error(data.error);
                getMe();
			}
		} catch (err) {
			console.error(err.message);
		}
    }

    return (
        <div className={styles.exchange}>
            <div className={styles.exchange__withdrawal}>
                <h1 className={styles.exchange__withdrawal_title}>Exchange</h1>
                <div className={styles.exchange__withdrawal__balance}>
                    <p className={styles.exchange__withdrawal__balance_text}>Your balance</p>
                    <p className={styles.exchange__withdrawal__balance_value}>₹ {Math.trunc(BalanceRupee)}</p>
                </div>
                <div className={styles.exchange__deposit}>
                    <div className={styles.exchange__deposit_details_container}>
                        <div className={styles.exchange__deposit__earnings__detail_container}>
                            If I exchange
                            <div className={`${styles.exchange__deposit__earnings__detail_value} ${styles.exchange__deposit__earnings__detail_red}`}>
                                <img src="/24=BCoin-flat.svg" alt="" />
                                {percent}
                            </div>  
                        </div>
                        <div className={styles.exchange__deposit__earnings__detail_container}>
                            I can earn
                            <div className={`${styles.exchange__deposit__earnings__detail_value} ${styles.exchange__deposit__earnings__detail_green}`}>
                                <img src="/24=rupee.svg" alt="" />
                                {(percent / 10000).toFixed(2)}
                            </div>
                        </div> 
                    </div>
                    <Slider percent={percent} setPercent={setPercent} step="25" min="0" max={BalanceBi} />
                    <Button 
                        filter
                        color="green"
                        fill
                        label="Exchange"
                        onClick={handleExchange}
                    />
                </div>
            </div>
        </div>
    );
};
