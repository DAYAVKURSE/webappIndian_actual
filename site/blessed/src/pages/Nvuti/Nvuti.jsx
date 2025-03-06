import { useEffect, useState, useRef } from "react";
import styles from "./Nvuti.module.scss";
import { ActionButtons } from "@/components";
import { rollDice } from "@/requests";
import { toast } from "react-hot-toast";
import useStore from "@/store";

export const Nvuti = () => {
	const { increaseBalanceRupee, decreaseBalanceRupee, BalanceRupee } = useStore();
	const [bet, setBet] = useState(100);
	const [percent, setPercent] = useState(65);
	const [number, setNumber] = useState(null);
	const sliderRef = useRef(null);
	const trackRef = useRef(null);
	const thumbRef = useRef(null);

	const calculateRange = () => {
		const maxValue = 999999;
		const winRange = Math.round((percent / 100) * (maxValue + 1));
		const lossRange = Math.round(maxValue - winRange + 1);

		return {
			less: `0 - ${winRange.toLocaleString('ru-RU').replace(',', ' ')}`,
			more: `${lossRange.toLocaleString('ru-RU').replace(',', ' ')} - ${maxValue.toLocaleString('ru-RU').replace(',', ' ')}`,
		};
	};

	const handlePercentChange = (value) => {
		let newValue = parseInt(value, 10);

		if (isNaN(newValue)) {
			newValue = 5;
		} else if (newValue < 5) {
			newValue = 5;
		} else if (newValue > 95) {
			newValue = 95;
		}
		
		setPercent(newValue);
		updateSliderPosition(newValue);
	};

	const updateSliderPosition = (value) => {
		if (sliderRef.current && trackRef.current && thumbRef.current) {
			trackRef.current.style.setProperty('--value', value);
			thumbRef.current.style.left = `${value}%`;
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

	const handleHalfBet = () => {
		const halved = Math.max(10, Math.floor(bet / 2));
		setBet(halved);
	};

	const handleDoubleBet = () => {
		const doubled = Math.min(BalanceRupee, bet * 2);
		setBet(doubled);
	};

	const handleIncreaseBet = () => {
		setBet(prev => Math.min(BalanceRupee, prev + 10));
	};

	const handleDecreaseBet = () => {
		setBet(prev => Math.max(10, prev - 10));
	};

	const { less, more } = calculateRange();

	useEffect(() => {
		updateSliderPosition(percent);
	}, [percent]);

	return (
		<div className={styles.nvuti}>
			<p className={styles.nvuti_number}>
				{number !== null ? number : "0"}
			</p>
			
			<ActionButtons
				onclick1={() => handleBet("less")}
				src1="/arrow_down.svg"
				label1=""
				color1="#BC1303"
				onclick2={() => handleBet("more")}
				src2="/arrow_up.svg"
				label2=""
				color2="#007E34"
			/>
			
			<div className={styles.nvuti_tip}>
				<p>{less}</p>
				<p>{more}</p>
			</div>
			
			<div className={styles.nvuti_bet_section}>
				<div className={styles.nvuti_bet_input}>
					<div className={styles.nvuti_bet_input_value}>
						{bet} ₹
					</div>
					<div className={styles.nvuti_bet_input_buttons}>
						<button onClick={handleDecreaseBet}>−</button>
						<button onClick={handleIncreaseBet}>+</button>
					</div>
				</div>
				<button className={styles.nvuti_bet_button} onClick={() => handleBet("more")}>
					Bet
				</button>
			</div>
			
			<div className={styles.nvuti_multiplier}>
				<button onClick={handleHalfBet}>/2</button>
				<button onClick={handleDoubleBet}>×2</button>
			</div>
			
			<h3>Percent</h3>
			
			<div className={styles.percent_value_container}>
				<p className={styles.percent_value}>{percent}%</p>
			</div>
			
			<div className={styles.slider_container}>
				<input
					ref={sliderRef}
					type="range"
					min="5"
					max="95"
					value={percent}
					onChange={(e) => handlePercentChange(e.target.value)}
					className={styles.slider}
					id="slider"
				/>
				<div className={styles.slider_track} ref={trackRef}>
					<div className={styles.slider_track_red}></div>
					<div className={styles.slider_track_green}></div>
				</div>
				<div 
					ref={thumbRef}
					className={styles.slider_thumb_icon} 
					style={{ left: `${percent}%` }}
				>
					<span></span>
					<span></span>
					<span></span>
				</div>
			</div>
		</div>
	);
};
