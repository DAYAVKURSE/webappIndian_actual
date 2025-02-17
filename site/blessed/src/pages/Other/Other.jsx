import styles from "./Other.module.scss";
import { Accordion } from "@/components"

export const Other = () => {
    
    const copyLink = () => {
        navigator.clipboard.writeText(window.Telegram.WebApp.initData);
    }

    return (
        <div className={styles.other}>
            <h2 className={styles.other_title}>Resources</h2>
            <p className={styles.other_description}>Get tips, support, and stay connected</p>
            <div className={styles.other__social}>
                <a href="https://t.me/bitrave"><img src="/40=telegram.svg" alt="telegram" /></a>
                <a href="https://discord.com/invite/asdjwqweq"><img src="/40=discord.svg" alt="discord" /></a>
                <a href="https://instagram.com/profile/bitrave"><img src="/40=instagram.svg" alt="instagram" /></a>
            </div>
            <Accordion 
                to="how-to-play"
                icon="/24=book_5.svg" 
                title="How to play?"
                filter="brightness(0) saturate(100%) invert(91%) sepia(42%) saturate(477%) hue-rotate(16deg) brightness(103%) contrast(106%)" 
            />
            <Accordion 
                to="faq"
                icon="/24=help.svg" 
                title="FAQ" 
                filter="brightness(0) saturate(100%) invert(100%) sepia(51%) saturate(2592%) hue-rotate(310deg) brightness(103%) contrast(104%)"
            />
            <Accordion 
                to="https://t.me/BiTRavesupport"
                icon="/24=headset_mic.svg" 
                title="Support" 
                filter="brightness(0) saturate(100%) invert(83%) sepia(20%) saturate(1042%) hue-rotate(321deg) brightness(101%) contrast(101%)"
            />
            <div className={styles.other_dev} onClick={() => copyLink()} />
        </div>
    );
};
