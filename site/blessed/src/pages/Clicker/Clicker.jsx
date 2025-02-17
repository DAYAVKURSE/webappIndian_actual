import { useState, useEffect, useRef } from "react";
import styles from "./Clicker.module.scss";
import { Link } from "react-router-dom";
import { getBPC, sendClicks, getMe } from "@/requests";
import useStore from "@/store";
import toast from "react-hot-toast";
import { useLocation } from "react-router-dom";

export const Clicker = () => {
    const { dailyClicks, BiPerClick, increaseBalanceBi, BalanceBi, incrementDailyClicks } = useStore();
    const [isAnimating, setIsAnimating] = useState(false);
    const [spawnedTexts, setSpawnedTexts] = useState([]);
    const clicksBuffer = useRef(0);
    const location = useLocation();

    useEffect(() => {
        getBPC();
        getMe();
    }, []);

    useEffect(() => {
        const interval = setInterval(() => {
            if (clicksBuffer.current > 0) {
                sendClicks(clicksBuffer.current, BiPerClick);
                clicksBuffer.current = 0;
            }
        }, 3000);

        return () => {
            clearInterval(interval);
        };
    }, [BiPerClick]);

    useEffect(() => {
        const handleBeforeUnload = (event) => {
            if (clicksBuffer.current > 0) {
                sendClicks(clicksBuffer.current, BiPerClick);
                clicksBuffer.current = 0;
            }
        };

        window.addEventListener("beforeunload", handleBeforeUnload);

        return () => {
            window.removeEventListener("beforeunload", handleBeforeUnload);
        };
    }, [BiPerClick]);

    useEffect(() => {
        return () => {
            if (clicksBuffer.current > 0) {
                sendClicks(clicksBuffer.current, BiPerClick);
                clicksBuffer.current = 0;
            }
        };
    }, [location, BiPerClick]);

    const handleClick = (e) => {
        if (dailyClicks < 10000) {
            increaseBalanceBi();
            incrementDailyClicks();
            clicksBuffer.current += 1;
            setIsAnimating(true);

            const x = e.clientX;
            const y = e.clientY - 40;

            setSpawnedTexts((prev) => [
                ...prev,
                { id: Date.now(), text: `+${BiPerClick?.toFixed(1) || '0.0'}`, x, y }
            ]);

            setTimeout(() => setIsAnimating(false), 200);
        } else {
            toast.error('Daily limit exceeded');
        }
    };

    const formatBalance = (balance) => {
        let balanceStr = Math.trunc(balance).toString();

        const balanceParts = balanceStr.replace(/\B(?=(\d{3})+(?!\d))/g, ".").split(".");

        const main = balanceParts.shift();
        const fraction = balanceParts.join('.');

        return { main, fraction };
    };

    const { main, fraction } = formatBalance(BalanceBi || 0);

    useEffect(() => {
        const interval = setInterval(() => {
            setSpawnedTexts((prev) => prev.filter(text => Date.now() - text.id < 1000));
        }, 100);

        return () => clearInterval(interval);
    }, []);

    return (
        <div className={styles.clicker__container}>
            <div className={styles.clicker__balance}>
                <p className={styles.clicker__balance__title}>Balance</p>
                <div className={styles.clicker__block}>
                    <p className={styles.clicker__balance__value}>
                        <span className={styles.main}>{main}</span>
                        {fraction && (
                            <span className={styles.clicker__balance__value_fraction}>
                                .{fraction}
                            </span>
                        )}
                    </p>
                    <img
                        className={`${styles.clicker__balance__value_img} ${fraction ? styles.clicker__balance__value_img_fraction : ''}`}
                        src="/24=BCoin-text-flat-crop.svg"
                        alt=""
                    />
                </div>
                <div className={`${styles.clicker__block} ${styles.clicker_daily}`}>
                    <img src="/24=Clicker.svg" alt="" />
                    <p>{dailyClicks ? dailyClicks : '0'}/10000</p>
                </div>
            </div>
            <div className={`${styles.clicker__coin_container} ${isAnimating ? `${styles.animate} ${styles.clicker__coin_glow}` : ''}`} onClick={handleClick}>
                <img src="/BCoin_base.png" alt=""
                    className={`${styles.clicker__coin} ${dailyClicks >= 10000 ? styles.clicker__coin_disabled : ''}`}
                />
                <img src="/BCoin_logo.png" alt=""
                    className={`${styles.clicker__coin} ${dailyClicks >= 10000 ? styles.clicker__coin_disabled : ''}`}
                    style={{ mixBlendMode: "screen", position: "absolute"}}
                />
            </div>
            <div className={styles.clicker__block} style={{ alignItems: 'center', opacity: 0.5, lineHeight: "18px" }}>
                <img className={`${dailyClicks >= 10000 ? styles.clicker__block_disabled : ''}`} src="/24=BCoin-flat.svg" alt="" />
                <p className={`${dailyClicks >= 10000 ? styles.clicker__block_disabled : ''}`}>{BiPerClick?.toFixed(1) || '0.0'} per click</p>
            </div>

            <div className={styles.clicker__navigation_container}>
                <div className={styles.clicker__navigation_select}>
                    <Link to="/wheel" className={styles.clicker__navigation_button}>
                        <img className={`${styles.clicker__navigation_icon} ${styles.clicker__navigation_icon_green}`} src="/wheel_select.png" alt="" />
                        <p className={styles.clicker__navigation_button_greenText}>Fortune Wheel</p>
                    </Link>
                    <div className={styles.clicker__navigation_separator}></div>
                    <Link to="/pass" className={styles.clicker__navigation_button}>
                        <img className={`${styles.clicker__navigation_icon} ${styles.clicker__navigation_icon_blue}`} src="/pass_case.png" alt="" />
                        <p className={styles.clicker__navigation_button_blueText}>Trave Pass</p>
                    </Link>
                </div>
            </div>

            <div>
                {spawnedTexts.map(({ id, text, x, y }) => (
                    <span
                        key={id}
                        className={styles.spawnedText}
                        style={{ top: y, left: x }}
                    >
                        {text}
                    </span>
                ))}
            </div>
        </div>
    );
};
