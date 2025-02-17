import { API_BASE_URL } from '@/config';
import styles from "./FortuneWheel.module.scss";
import { useEffect, useState } from "react";
import { getWheelWins, spinWheel, getWheelSpins, getMe } from "@/requests";
import useStore from "@/store";
const initData = window.Telegram.WebApp.initData;
import toast from "react-hot-toast";
import { WheelOfFortune } from '@/components';

export const FortuneWheel = () => {
    const [mustSpin, setMustSpin] = useState(false);
    const [prizeNumber, setPrizeNumber] = useState(0);
    const { BalanceBi } = useStore();
    const [wins, setWins] = useState([]);
    const [spinning, setSpinning] = useState(false);
    const [spins, setSpins] = useState(0);

    const getBenefitText = (benefit) => {
        switch (benefit.PolymorphicBenefitType) {
            case "benefit_credit":
                return benefit.PolymorphicBenefit.BCoinsAmount > 0
                    ? `${benefit.PolymorphicBenefit.BCoinsAmount} BCoins`
                    : `${benefit.PolymorphicBenefit.RupeeAmount} Rupees`;
            case "benefit_fortune_wheel":
                return `${benefit.PolymorphicBenefit.FreeSpinsAmount} Free Spins`;
            case "benefit_replenishment":
                return `${benefit.PolymorphicBenefit.BonusMultiplier * 100}% Bonus`;
            case "benefit_clicker":
                return benefit.PolymorphicBenefit.Reset
                    ? "Reset Clicks"
                    : `${benefit.PolymorphicBenefit.BonusMultiplier}x Click Bonus`;
            case "benefit_binary_option":
                return `${benefit.PolymorphicBenefit.FreeBetsAmount} Free Bets`;
            case "benefit_item":
                return `${benefit.PolymorphicBenefit.ItemName.substring(0, 13)}`;
            default:
                return "Unknown";
        }
    };

    useEffect(() => {
        let firstMessageIgnored = false;


        getWheelSpins().then((data) => {
            setSpins(data.TotalSpins);
        });

        getWheelWins()
            .then((response) => {
                if (response.status === 200) {
                    response.json().then((data) => {
                        const formattedWins = data.map((win) => ({
                            nickname: win.Nickname,
                            wheelSector: {
                                id: win.FortuneWheelSector.ID,
                                color: win.FortuneWheelSector.ColorHex,
                                benefit: {
                                    id: win.FortuneWheelSector.Benefit.ID,
                                    type: win.FortuneWheelSector.Benefit.PolymorphicBenefitType,
                                    details: win.FortuneWheelSector.Benefit.PolymorphicBenefit
                                }
                            },
                            amount: getBenefitText(win.FortuneWheelSector.Benefit),
                            timestamp: new Date(win.Timestamp).toLocaleString()
                        }));

                        setWins(formattedWins);
                    }).catch((error) => {
                        console.error("Error parsing JSON from getWheelWins:", error);
                    });
                } else {
                    firstMessageIgnored = true;
                    console.error("Error fetching wins:", response.status);
                }
            })
            .catch((error) => {
                console.error("Request error:", error);
            });

        const encoded_init_data = encodeURIComponent(initData);
        const ws = new WebSocket(`wss://${API_BASE_URL}/ws/fortunewheel/live?init_data=${encoded_init_data}`);

        ws.onmessage = (event) => {
            try {
                if (!firstMessageIgnored) {
                    firstMessageIgnored = true;
                    return;
                }
        
                const data = JSON.parse(event.data);
                const benefitText = getBenefitText(data.FortuneWheelSector.Benefit);
                const newWin = {
                    nickname: data.Nickname,
                    amount: benefitText,
                    timestamp: new Date(data.Timestamp).toLocaleString(),
                };
        
                setWins((prevWins) => {
                    if (!prevWins.some((win) =>
                        win.nickname === newWin.nickname &&
                        win.amount === newWin.amount &&
                        win.timestamp === newWin.timestamp
                    )) {
                        return [...prevWins, newWin];
                    }
                    return prevWins;
                });
            } catch (error) {
                console.error("Error parsing live update WebSocket message:", error);
            }
        };

        return () => {
            ws.close();
        };
    }, []);


    const handleSpinClick = async () => {
        try {
            if (spinning) return;
            setSpinning(true);
            const spinResponse = await spinWheel();
            const spinResult = await spinResponse.json();
            if (spinResponse.status !== 200) {
                toast.error(spinResult.error);
                setSpinning(false);
                return;
            }
            const prizeIndex = spinResult.ID;
            setTimeout(() => {
                setSpinning(false);
                getWheelSpins().then((data) => {
                    setSpins(data.TotalSpins);
                });
                getMe();
            }, 6000);
            setMustSpin(true);
            setPrizeNumber(prizeIndex);
        } catch (error) {
            console.error("Spin error:", error);
            toast.error("Spin error");
        }
    };

    const WinsList = ({ wins }) => (
        <div className={styles.fortunewheel__wins_list}>
            {wins.length > 0 ? wins.map((win, index) => (
                <div key={index} className={styles.fortunewheel__wins__item}>
                    <p className={styles.fortunewheel__wins__item_username}>{win.nickname}</p>
                    <p className={styles.fortunewheel__wins__item_amount}>
                        <img src="/24=BCoin-flat.svg" alt="bcoin" />
                        {win.amount}
                    </p>
                </div>
            ))
                : <p className={styles.fortunewheel__wins__empty}>No wins</p>}
        </div>
    );
    
    const formatBalance = (balance) => {
        let balanceStr = Math.trunc(balance).toString();

        const balanceParts = balanceStr.replace(/\B(?=(\d{3})+(?!\d))/g, ".").split(".");

        const main = balanceParts.shift();
        const fraction = balanceParts.join('.');

        return { main, fraction };
    };

    const { main, fraction } = formatBalance(BalanceBi || 0);

    return (
        <div className={styles.fortunewheel}>
            <div className={styles.fortunewheel__balance}>
                <p className={styles.fortunewheel__balance__title}>Balance</p>
                {/* <div className={styles.fortunewheel__block}>
                    <p className={styles.fortunewheel__balance__value}>{BalanceBi.toFixed()}</p>
                    <img src="/24=BCoin-text-flat-crop.svg" alt="" />
                </div> */}
                
                <div className={styles.fortunewheel__block}>
                    <p className={styles.fortunewheel__balance__value}>
                        <span className={styles.main}>{main}</span>
                        {fraction && (
                            <span className={styles.fortunewheel__balance__value_fraction}>
                                .{fraction}
                            </span>
                        )}
                    </p>
                    <img
                        className={`${styles.fortunewheel__balance__value_img} ${fraction ? styles.fortunewheel__balance__value_img_fraction : ''}`}
                        src="/24=BCoin-text-flat-crop.svg"
                        alt=""
                    />
                </div>
                <div className={`${styles.fortunewheel__block} ${styles.fortunewheel__block_spins}`}>
                    <img src="/24=Fortune.svg" alt="" />
                    <p>{spins} spins left</p>
                </div>
            </div>
            <div className={styles.fortunewheel__wheel} onClick={handleSpinClick}>
                <WheelOfFortune
                    mustStartSpinning={mustSpin}
                    winningSector={prizeNumber - 1}
                    numberOfSectors={12}
                    onClick={handleSpinClick}
                    onStopSpinning={() => setMustSpin(false)}
                />
                <img className={styles.fortunewheel_img} src="wheel_of_fortune_fixed.png" alt="" />
            </div>
            <div className={styles.fortunewheel__wins}>
                <p className={styles.fortunewheel__wins_title}>Last wins</p>
                <WinsList wins={wins.slice(-5)} />
            </div>
        </div>
    );
};
