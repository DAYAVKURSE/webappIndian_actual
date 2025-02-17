import { useRef, useState, useEffect } from "react";
import styles from "./Dice.module.scss";
import { ActionButtons, Amount } from "@/components";
import Lottie from "lottie-react";
import animationData from "./dice.json";
import { rollDice } from "@/requests";
import { toast } from "react-hot-toast";
import useStore from "@/store";

export const Dice = () => {
	const { increaseBalanceRupee, decreaseBalanceRupee } = useStore();
    const [bet, setBet] = useState(100);
    const [percent, setPercent] = useState(50);
    const [isNumberVisible, setIsNumberVisible] = useState(false); 
    const [number, setNumber] = useState(null); 
    const [isAnimating, setIsAnimating] = useState(false);
    const lottieRef = useRef();

    const calculateRange = () => {
        const maxValue = 999999;
        const winRange = Math.round((percent / 100) * (maxValue + 1));
        const lossRange = Math.round(maxValue - winRange + 1);
    
        return {
            less: `0 - ${(winRange).toLocaleString()}`,
            more: `${lossRange.toLocaleString()} - ${maxValue.toLocaleString()}`,
        };
    };
    
    const handlePercentChange = (value) => {
        let newValue = parseInt(value, 10);
    
        if (isNaN(newValue)) {
            setPercent(5);
        } else if (newValue < 5) {
            setPercent(5);
        } else if (newValue > 95) {
            setPercent(95);
        } else {
            setPercent(newValue);
        }
    };

    const validateNumber = (num) => {
        const minValue = 0;
        const maxValue = 999999;
        if (num < minValue || num > maxValue) {
            toast.error(`Number out of range! It should be between ${minValue} and ${maxValue}.`);
            return false;
        }
        return true;
    };

    useEffect(() => {
        let timer;
        if (isNumberVisible) {
            timer = setTimeout(() => {
                setIsAnimating(true);
            }, 2000);
        }

        return () => clearTimeout(timer);
    }, [isNumberVisible]);

    const handleBet = async (direction) => {
        setIsAnimating(false); 
        setIsNumberVisible(false);

        setTimeout(async () => {
            setIsNumberVisible(true);

            try {
                const response = await rollDice(bet, percent, direction);

                if (response.status === 200) {

                    if (lottieRef.current) {
                        lottieRef.current.stop();
                        lottieRef.current.play();
                    }
                    
                    const data = await response.json();
                    const rolledNumber = data.result.number;
                    decreaseBalanceRupee(bet);

                    if (validateNumber(rolledNumber)) {
                        setNumber(rolledNumber);
                    }

                    setTimeout(() => {
                        if (data.result.won === true) {
                            toast.success(`You won! (+${Math.trunc(data.payout)})`);
                            increaseBalanceRupee(data.payout);
                        } else {
                            toast.error(`You lost! (-${bet})`);
                        }
                    }, 2000);
                } else if (response.status === 400) {
                    toast.error("Invalid request body");
                } else if (response.status === 402) {
                    toast.error("Insufficient balance");
                } else {
                    toast.error("An error occurred");
                }
            } catch (err) {
                toast.error(err.message);
            }
        }, 500);
    };

    const { less, more } = calculateRange();

    useEffect(() => {
        const slider = document.querySelector('#slider');

        slider.addEventListener('input', function() {
            this.style.setProperty('--value', `${this.value}%`);
        });
    }, []);

    return (
        <div className={styles.dice}>
            <div className={styles.dice_bet_container}>
                <p className={`${styles.dice_number} ${isAnimating ? styles.fadeIn : styles.fadeOut}`}>
                    {number !== null ? number : ""}
                </p>
                <div className={styles.dice_lottie}>
                    <Lottie
                        lottieRef={lottieRef}
                        animationData={animationData}
                        loop={false}
                    />
                </div>
                <ActionButtons
                    onclick1={() => handleBet("less")}
                    src1="/24=arrow_circle_down.svg"
                    label1="LESS"
                    color1="#FFC397"
                    onclick2={() => handleBet("more")}
                    src2="/24=arrow_circle_up.svg"
                    label2="MORE"
                    color2="#EDFF8C"
                />
                <div className={styles.dice_tip}>
                    <p>{less}</p>
                    <p>{more}</p>
                </div>
                <h3>Bet</h3>
                <Amount bet={bet} setBet={setBet} />
                <h3>Percent</h3>
                <p className={styles.slider_value}>{percent}%</p>
                <input
                    type="range"
                    min="5"
                    max="95"
                    value={percent}
                    onChange={(e) => handlePercentChange(e.target.value)}
                    className={styles.slider}
                    id="slider"
                />
                <br />
            </div>
        </div>
    );
};
