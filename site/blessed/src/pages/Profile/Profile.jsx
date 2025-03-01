import styles from "./Profile.module.scss";
import { Button, Input } from "@/components";
import useStore from "@/store";
import { useEffect, useState } from "react";
import { getReferrals, getMe } from "@/requests";

export const Profile = () => {
    const { userName, BalanceRupee } = useStore();
    const [referrals, setReferrals] = useState([]);
    const [totalEarned, setTotalEarned] = useState(0);
    const userId = window.Telegram.WebApp.initDataUnsafe.user.id;

    const copyLink = () => {
        const link = `https://t.me/BiTRave_bot?start=${userId}`;
        navigator.clipboard.writeText(link);
    }

    useEffect(() => {
        getMe();
        getReferrals().then((data) => {
            if (Array.isArray(data)) {
                setReferrals(data);
                const earned = data.reduce((total, referral) => total + (referral.EarnedAmount || 0), 0);
                setTotalEarned(earned);
            } else {
                console.error('Unexpected data format:', data);
            }
        }).catch((error) => {
            console.error('Error fetching referrals:', error);
        });
    }, []);

    return (
        <div className={styles.profile}>
            <h1 className={styles.title}>Account</h1>
            <p className={styles.username}>Name {userName}</p>
            
            <div className={styles.referralSection}>
                <h2 className={styles.referHeader}>Refer a friend and earn 20%</h2>
                
                <p className={styles.referDescription}>
                    Share your link with friends, family, or on social media. 
                    When someone signs up and makes a deposit using your link, 
                    you'll get 20% of their deposit
                </p>
                
                <div className={styles.linkContainer}>
                    <input 
                        type="text" 
                        className={styles.linkInput} 
                        value={`https://t.me/BiTRave_bot?start=${userId}`} 
                        readOnly 
                    />
                    <button className={styles.copyButton} onClick={copyLink}>
                        <img src="/24=content_copy.svg" alt="Copy" />
                    </button>
                </div>
            </div>
            
            <div className={styles.referralsContainer}>
                <h2 className={styles.referralsTitle}>Your referrals</h2>
                <p className={styles.totalEarnedLabel}>Total earned</p>
                
                <div className={styles.totalEarned}>
                    <div className={styles.rupeeIcon}>₹</div>
                    <span className={styles.earnedAmount}>{totalEarned}</span>
                </div>
                
                {referrals.length > 0 ? (
                    <ol className={styles.referralsList}>
                        {referrals.map((user, index) => (
                            <li key={index} className={styles.referralItem}>{user.ReferredNickname}</li>
                        ))}
                    </ol>
                ) : (
                    <p className={styles.noReferrals}>No referrals yet</p>
                )}
            </div>
        </div>
    );
};