import styles from "./Topup.module.scss";
import { useState } from "react";
import { createPaymentPage } from "@/requests";
import { toast } from "react-hot-toast";
import { isMobile } from 'react-device-detect'

export const Topup = () => {
    const [amount, setAmount] = useState(500);
    const [loading, setLoading] = useState(false);

    const handleAmountChange = (e) => {
        const value = e.target.value;
        if (/^\d*$/.test(value)) {
            setAmount(Number(value));
        }
    };

    const handleTopup = async () => {
        if (amount < 500) {
            toast.error('Minimum deposit amount is 500 rupees');
            return;
        }
        setLoading(true);
        try {
            const data = await createPaymentPage(amount);
            if (data && data.url) {
                window.location.href = data.url;
            } else {
                toast.error('Error creating payment page');
                console.error('No URL returned from createPaymentPage');
            }
        } catch (error) {
            console.error('Error during topup:', error);
            toast.error('Error processing deposit');
        } finally {
            setLoading(false);
        }
    };

    return (
        <>
        {
            isMobile && <div style={{minHeight:'50px', background: 'transparent'}}></div>
          }
        <div className={styles.topup}>
            <div className={styles.topup__topup}>
                <h1 className={styles.topup__topup_title}>Deposit</h1>
                <div className={styles.topup__topup__balance}>
                    <p className={styles.topup__topup__balance_text}>Enter amount</p>
                 
                </div>
              
                <div className={styles.topup__amount}>
                    <div className={styles.topup__amount__input}>
                        <div className={styles.topup__amount__input_container}>
                     
                            <input
                                type="text"
                                placeholder="500"
                                value={`${amount}`}
                                onChange={handleAmountChange}
                            />
                        </div>
                       
                    </div>
                </div>
                <div className={styles.buttons}>
                    <button onClick={() => setAmount(500)} className={styles.depositButton}>500₹</button>
                    <button onClick={() => setAmount(1000)} className={styles.depositButton}>1 000₹</button>
                    <button onClick={() => setAmount(2000)} className={styles.depositButton}>2 000₹</button>
                    <button onClick={() => setAmount(5000)} className={styles.depositButton}>5 000₹</button>
                    <button onClick={() => setAmount(10000)} className={styles.depositButton}>10 000₹</button>
                    <button onClick={() => setAmount(20000)} className={styles.depositButton}>2 0000₹</button>
                </div>
                {/* <div className={styles.topup__amount__input_container}>
                            <button onClick={() => setAmount(amount + 100)}>+100</button>
                            <button onClick={() => setAmount(amount + 500)}>+500</button>
                        </div> */}

               <button
                    className={styles.nextButton}
                    onClick={handleTopup}
                    disabled={loading}
               >
                    {loading ? 'Loading...' : 'Next'}
               </button>
            </div>
        </div>
        </>
    );
};
