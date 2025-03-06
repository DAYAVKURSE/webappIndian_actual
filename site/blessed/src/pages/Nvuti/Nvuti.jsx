import { useEffect, useState } from "react";
import styles from "./Nvuti.module.scss";
import { ActionButtons } from "@/components";
import { rollDice } from "@/requests";
import { toast } from "react-hot-toast";
import useStore from "@/store";

export const Nvuti = () => {
	const { increaseBalanceRupee, decreaseBalanceRupee, BalanceRupee } = useStore();
	const [bet, setBet] = useState(100);
	const [percent, setPercent] = useState(45);
	const [number, setNumber] = useState(null);

	const calculateRange = () => {
		return {
			less: "0 - 500 000",
			more: "500 000 - 999 999",
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
		const slider = document.querySelector('#slider');
		if (slider) {
			slider.style.setProperty('--value', `${percent}%`);
		}
	}, [percent]);

	return (
		<div className={styles.nvuti}>
			<div className={styles.nvuti_header}>
				<button className={styles.nvuti_header_button}>
					<img src="/profile_icon.svg" alt="Profile" />
				</button>
				<div className={styles.nvuti_header_balance}>
					<span className={styles.nvuti_balance_icon}>₹</span>
					<span className={styles.nvuti_balance_value}>50552</span>
				</div>
				<button className={styles.nvuti_header_button}>
					<img src="/menu_icon.svg" alt="Menu" />
				</button>
			</div>
			
			<div className={styles.nvuti_main_balance}>
				<span className={styles.nvuti_main_balance_icon}>₹</span>
				<span className={styles.nvuti_main_balance_value}>50552.333333333334</span>
			</div>
			
			<p className={styles.nvuti_number}>
				{number !== null ? number : "0"}
			</p>
			
			<ActionButtons
				onclick1={() => handleBet("less")}
				src1="/arrow_down.svg"
				label1="Less"
				color1="#D32E26"
				onclick2={() => handleBet("more")}
				src2="/arrow_up.svg"
				label2="More"
				color2="#17B322"
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
			<p className={styles.slider_value}>{percent}%</p>
			
			<div className={styles.slider_container}>
				<input
					type="range"
					min="5"
					max="95"
					value={percent}
					onChange={(e) => handlePercentChange(e.target.value)}
					className={styles.slider}
					id="slider"
				/>
				<div 
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
