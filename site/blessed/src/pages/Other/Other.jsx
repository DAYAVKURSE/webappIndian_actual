import { Link } from "react-router-dom";
import styles from "./Other.module.scss";

export const Other = () => {
    
    const copyLink = () => {
        navigator.clipboard.writeText(window.Telegram.WebApp.initData);
    }

    return (
        <div className={styles.other}>
            <h2 className={styles.other_title}>Other</h2>
            <Link to={'/profile'} className={`${styles.account__banner} ${styles.banner}`}>
                <svg width="25" height="25" viewBox="0 0 25 25" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <circle cx="12.5" cy="12.5" r="12.5" fill="white" />
                    <path d="M12.6639 13.8421C7.75562 13.7383 6.20831 16.9764 6.01788 19.0964C6.00384 19.2527 6.13017 19.382 6.28709 19.382H18.6967C18.8596 19.382 18.988 19.2432 18.9665 19.0816C18.6915 17.0175 17.3454 13.7387 12.6639 13.8421Z" stroke="#10202E" strokeWidth="1.2" />
                    <path d="M15.4497 9.49472C15.4497 10.9599 14.2543 12.1535 12.7724 12.1535C11.2906 12.1535 10.0951 10.9599 10.0951 9.49472C10.0951 8.02954 11.2906 6.83596 12.7724 6.83596C14.2543 6.83596 15.4497 8.02954 15.4497 9.49472Z" stroke="#10202E" strokeWidth="1.2" />
                </svg>
                <span>Account</span>
            </Link>
            <div className={styles.downBanners}>
                <Link to={'/wallet/topup'} className={`${styles.deposit__banner} ${styles.banner}`}>
                <svg width="25" height="25" viewBox="0 0 25 25" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <circle cx="12.5" cy="12.5" r="12.5" fill="#6EBEFF" />
                    <path d="M18.5317 6H6.84004C5.82511 6 4.99951 6.81111 4.99951 7.80822V15.3096C4.99951 16.3065 5.82511 17.1176 6.84004 17.1176H8.45062V15.8456H6.84004C6.53911 15.8456 6.29431 15.6052 6.29431 15.3096V10.0852H19.0774V15.3096C19.0774 15.6051 18.8326 15.8456 18.5317 15.8456H16.9509V17.1176H18.5317C19.5466 17.1176 20.3722 16.3065 20.3722 15.3096V7.80822C20.3722 6.81127 19.5466 6 18.5317 6ZM6.29431 8.81312V7.80822C6.29431 7.51256 6.53911 7.27206 6.84004 7.27206H18.5318C18.8326 7.27206 19.0774 7.51256 19.0774 7.80822V8.81312H6.29431Z" fill="#10202E" />
                    <path d="M6.97705 11.1177H14.978V12.353H6.97705V11.1177Z" fill="#10202E" />
                    <path d="M9.58447 16.9709L10.5023 17.8673L12.0255 16.3795V21H13.3233V16.3567L14.8697 17.8673L15.7875 16.9709L12.6861 13.9412L9.58447 16.9709Z" fill="#10202E" />
                </svg>
                    <span>Deposit</span>
                </Link>
                <Link to={'/wallet/withdrawal'} className={`${styles.withdraw__banner} ${styles.banner}`}>
                    <svg width="25" height="25" viewBox="0 0 25 25" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <circle cx="12.5" cy="12.5" r="12.5" fill="#28D4D7" />
                        <path d="M18.3941 6H6.79013C5.80306 6 5 6.79364 5 7.76932V15.2599C5 16.2356 5.80306 17.0294 6.79013 17.0294H8.38868V15.8315H6.79013C6.47136 15.8315 6.21193 15.575 6.21193 15.2599V10.0069H18.9721V15.2599C18.9721 15.575 18.7128 15.8315 18.3939 15.8315H16.825V17.0294H18.3939C19.381 17.0294 20.184 16.2356 20.184 15.2599V7.76932C20.184 6.79364 19.3811 6 18.3941 6ZM6.21209 8.80894V7.76932C6.21209 7.45423 6.47136 7.19779 6.79013 7.19779H18.3941C18.7128 7.19779 18.9723 7.45423 18.9723 7.76932V8.80894H6.21209Z" fill="#10202E" />
                        <path d="M7.05371 11.1177H14.9137V12.353H7.05371V11.1177Z" fill="#10202E" />
                        <path d="M13.1854 18.7345V14.0293H11.9767V18.7119L10.4104 17.1707L9.55566 18.0116L12.5926 20.9999L15.6293 18.0116L14.7747 17.1707L13.1854 18.7345Z" fill="#10202E" />
                    </svg>
                    <span>Withdrawal</span>
                </Link>
            </div>
            
            <div className={styles.links}>
                <Link to={'/other/faq'}>FAQ</Link>
                <Link to={"https://t.me/rupexsupport"}>Support</Link>
            </div>
            
            <div className={styles.other_dev} onClick={() => copyLink()} />

            <div className={styles.other__social}>
                <a href="https://t.me/RupeXBot"><img src="/telegram.png" alt="telegram" /></a>
                <a ><img src="/discord.png" alt="discord" /></a>
                <a ><img src="/instagram.png" alt="instagram" /></a>
                
            </div>
        </div>
    );
};

  
