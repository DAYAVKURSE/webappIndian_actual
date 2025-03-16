
//    const userId = Number(window.Telegram.WebApp.initDataUnsafe.user.id);
import styles from "./Profile.module.scss";
import { useEffect, useState } from "react";
import { getReferrals, getMe } from "@/requests";

export const Profile = () => {
    const [referrals, setReferrals] = useState([]);
    const [totalEarned, setTotalEarned] = useState(0);
    const userId = Number(window.Telegram.WebApp.initDataUnsafe.user.id);
    const userName = localStorage.getItem('store')
    console.log(userName);
    
    const copyLink = () => {
        const link = `https://t.me/RupeXBot=${userId}`;
        navigator.clipboard.writeText(link);
    }

    useEffect(() => {
        try {
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
        }
        catch(e) {
            console.log(e);
        }
    }, []);

    return (
        <div className={styles.profile}>
            <h1 className={styles.title}>Account</h1>
            <p className={styles.username}>Name {userName.userName || ''}</p>
            
            <div className={styles.referralSection}>
                <h2 className={styles.referHeader}>Refer a friend and earn 20%</h2>
                
                <p className={styles.referDescription}>
                    Share your link with friends, family, or on social media. 
                    When someone signs up and makes a deposit using your link, 
                    youll get 20% of their deposit
                </p>
                
                <div className={styles.linkContainer}>
                    <input 
                        type="text" 
                        className={styles.linkInput} 
                        value={`https://t.me/RupeXBot?start=${userId}`} 
                        readOnly 
                    />
                    <button className={styles.copyButton} onClick={copyLink}>
                    <svg width="25" height="25" viewBox="0 0 25 25" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <g clipPath="url(#clip0_849_9421)">
                            <path d="M20.9821 0C22.0477 0 23.0697 0.423309 23.8232 1.1768C24.5767 1.9303 25 2.95226 25 4.01786V17.4107C25 17.7659 24.8589 18.1066 24.6077 18.3577C24.3566 18.6089 24.0159 18.75 23.6607 18.75C23.3055 18.75 22.9649 18.6089 22.7137 18.3577C22.4625 18.1066 22.3214 17.7659 22.3214 17.4107V4.01786C22.3214 3.66266 22.1803 3.322 21.9292 3.07084C21.678 2.81967 21.3373 2.67857 20.9821 2.67857H7.58929C7.23408 2.67857 6.89343 2.53747 6.64227 2.2863C6.3911 2.03514 6.25 1.69449 6.25 1.33929C6.25 0.984085 6.3911 0.643433 6.64227 0.392268C6.89343 0.141103 7.23408 5.2929e-09 7.58929 0L20.9821 0ZM16.9643 5.35714C17.6747 5.35714 18.356 5.63935 18.8583 6.14168C19.3607 6.64401 19.6429 7.32531 19.6429 8.03571V22.3214C19.6429 23.0318 19.3607 23.7131 18.8583 24.2155C18.356 24.7178 17.6747 25 16.9643 25H2.67857C1.96817 25 1.28687 24.7178 0.784536 24.2155C0.282206 23.7131 0 23.0318 0 22.3214V8.03571C0 7.32531 0.282206 6.64401 0.784536 6.14168C1.28687 5.63935 1.96817 5.35714 2.67857 5.35714H16.9643Z" fill="#6EBEFF" />
                        </g>
                        <defs>
                            <clipPath id="clip0_849_9421">
                            <rect width="25" height="25" fill="white" transform="matrix(-1 0 0 1 25 0)" />
                            </clipPath>
                        </defs>
                    </svg>
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