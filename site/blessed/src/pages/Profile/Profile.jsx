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
        const link = document.getElementById('link');
        navigator.clipboard.writeText(link.placeholder);
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

    const formatBalance = (balance) => {
        let balanceStr = Math.trunc(balance).toString();

        const balanceParts = balanceStr.replace(/\B(?=(\d{3})+(?!\d))/g, ".").split(".");

        const main = balanceParts.shift();
        const fraction = balanceParts.join('.');

        return { main, fraction };
    };

    const { main, fraction } = formatBalance(BalanceRupee || 0);

    return (
        <div className={styles.profile}>
            <p className={styles.profile_username}>{userName}</p>
            <div className={styles.profile__balance_container}>
                <p className={styles.profile__balance_desc}>Your balance</p>
                <p className={styles.trading_balance}>
                    <span className={styles.profile__balance}>₹ {main}</span>
                    {fraction && (
                        <span className={styles.profile__balance} style={{fontSize: "26px"}}>
                            .{fraction}
                        </span>
                    )}
                </p>
            </div>
            <div className={styles.profile_container}>
                <div className={styles.profile_container}>
                    <Button label="Refer a Friend and Earn 20%" fill color="green" style={{ justifyContent: "left" }} />
                    <p className={styles.profile_block}>Share your link with friends, family, or on social media. When someone signs up and makes a deposit using your link, you&apos;ll get 20% of their deposit.</p>
                    <div className={styles.profile__buttons}>
                        <Input id="link" placeholder={`https://t.me/BiTRave_bot?start=${userId}`} disabled />
                        <Button
                            icon="24=content_copy.svg"
                            circle
                            filter
                            color="beige"
                            toastText="Copied!"
                            toastType="success"
                            onClick={() => copyLink()}
                        />
                    </div>
                </div>
                <div className={styles.profile__referrals}>
                    <h3 className={styles.profile__referrals_title}>Your referrals</h3>
                    <div className={styles.profile_block}>
                        <p className={styles.profile__referrals__details_title}>Total earned:</p>
                        <div className={styles.profile__referrals__details}>
                            <p className={`${styles.profile__referrals__details_detail} ${styles.profile__referrals__details_green}`}>
                                <img src="/24=rupee.svg" alt="" />
                                <p className={styles.profile__referrals__details_text}>{totalEarned}</p>
                            </p>
                        </div>
                    </div>
                    <div className={styles.profile__list}>
                        {referrals.length > 0 ? (
                            <ol className={styles.profile__users}>
                                {referrals.map((user, index) => (
                                    <li key={index} className={styles.profile__user}>{user.ReferredNickname}</li>
                                ))}
                            </ol>
                        ) : (
                            <p className={styles.profile__empty}>No referrals yet</p>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};