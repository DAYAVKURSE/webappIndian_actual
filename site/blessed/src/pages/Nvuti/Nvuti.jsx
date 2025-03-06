import { useEffect, useState } from "react";
import styles from "./Nvuti.module.scss";
import { ActionButtons, Amount } from "@/components";
import { rollDice } from "@/requests";
import { toast } from "react-hot-toast";
import useStore from "@/store";

export const Nvuti = () => {
	const { increaseBalanceRupee, decreaseBalanceRupee } = useStore();
    const [bet, setBet] = useState(100);
    const [percent, setPercent] = useState(50);
    const [number, setNumber] = useState(null);

    const calculateRange = () => {
        const maxValue = 999999;
        const boundary = Math.round((percent / 100) * maxValue);
        
        return {
            less: `0 - ${boundary.toLocaleString()}`,
            more: `${boundary.toLocaleString()} - ${maxValue.toLocaleString()}`,
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

    const handleBet = async (direction) => {
        setTimeout(async () => {
            try {
                const response = await rollDice(bet, percent, direction);

                if (response.status === 200) {
                    const data = await response.json();
                    const rolledNumber = data.result.number;
                    decreaseBalanceRupee(bet);

                    if (validateNumber(rolledNumber)) {
                        setNumber(rolledNumber);
                    }

                    if (data.result.won === true) {
                        toast.success(`You won! (+${Math.trunc(data.payout)})`);
                        increaseBalanceRupee(data.payout);
                    } else {
                        toast.error(`You lost! (-${bet})`);
                    }
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
        <div className={styles.nvuti}>
            <p className={styles.nvuti_number}>
                {number !== null ? number : "0"}
            </p>
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
            <div className={styles.nvuti_tip}>
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
    );
};
