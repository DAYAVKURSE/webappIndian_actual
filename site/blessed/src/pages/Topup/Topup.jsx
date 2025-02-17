import styles from "./Topup.module.scss";
import { Slider, Button } from "@/components";
import { useState } from "react";
import useStore from "@/store";
import { createPaymentPage } from "@/requests";
import { toast } from "react-hot-toast";

export const Topup = () => {
    const [rupee, setRupee] = useState(50);
    const [amount, setAmount] = useState(500);
    const { BalanceRupee } = useStore();

    const handleAmountChange = (e) => {
        const value = e.target.value;
        if (/^\d*$/.test(value)) {
            setAmount(Number(value));
        }
    };

    const handleTopup = async () => {
        if (amount < 500) {
            toast.error('Minimum topup amount is 500');
            return;
        }
        try {
            const data = await createPaymentPage(amount);
            if (data && data.url) {
                window.location.href = data.url;
            } else {
                console.error('No URL returned from createPaymentPage');
            }
        } catch (error) {
            console.error('Error during topup:', error);
        }
    };

    return (
        <div className={styles.topup}>
            <div className={styles.topup__topup}>
                <h1 className={styles.topup__topup_title}>Top up</h1>
                <div className={styles.topup__topup__balance}>
                    <p className={styles.topup__topup__balance_text}>Your balance</p>
                    <p className={styles.topup__topup__balance_value}>₹ {BalanceRupee.toFixed()}</p>
                </div>
                <div className={styles.topup__deposit}>
                    <div className={styles.topup__deposit_details_container}>
                        <h3 className={styles.topup__deposit_title}>Deposit</h3>
                        <div className={styles.topup__deposit__earnings__detail_container}>
                            I can earn
                            <div className={`${styles.topup__deposit__earnings__detail_value} ${styles.topup__deposit__earnings__detail_red}`}>
                                <img src="/24=BCoin-flat.svg" alt="" />
                                {(rupee / 10).toFixed(1)} per click
                            </div>
                        </div>
                        <div className={styles.topup__deposit__earnings__detail_container}>
                            If I invest
                            <div className={`${styles.topup__deposit__earnings__detail_value} ${styles.topup__deposit__earnings__detail_green}`}>
                                <img src="/24=rupee.svg" alt="" />
                                {rupee}
                            </div>
                        </div>
                    </div>
                    <Slider percent={rupee} setPercent={setRupee} step="25" min="0" max="10000" />
                </div>
                <div className={styles.topup__amount}>
                    <h3 className={styles.topup__deposit_title}>Enter the amount</h3>
                    <div className={styles.topup__amount__input}>
                        <div className={styles.topup__amount__input_container}>
                            <img src="/24=rupee.svg" alt="" />
                            <input
                                type="text"
                                placeholder="500"
                                value={amount}
                                onChange={handleAmountChange}
                            />
                        </div>
                        <div className={styles.topup__amount__input_container}>
                            <button onClick={() => setAmount(amount + 100)}>+100</button>
                            <button onClick={() => setAmount(amount + 500)}>+500</button>
                        </div>
                    </div>
                </div>
                {/* <div className={styles.topup__amount}>
                    <h3 className={styles.topup__deposit_title}>Payment method</h3>
                    <div className={styles.topup__payment__options}>
                        <div className={styles.topup__payment__options_option} onClick={handleTopup}>
                            <img src="/24=card.svg" alt="" />
                            <p>VISA/MASTERCARD</p>
                        </div>
                        <div className={styles.topup__payment_separator} />
                        <div className={styles.topup__payment__options_option}>
                            <img src="/24=BCoin-flat.svg" alt="" />
                            <p>Crypto #2</p>
                        </div>
                    </div>
                </div> */}

                <Button
                    label={"Topup"}
                    onClick={handleTopup}
                    color="lightYellow"
                />
            </div>
        </div>
    );
};
