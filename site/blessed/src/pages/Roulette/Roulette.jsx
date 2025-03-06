import { useEffect, useState } from "react";
import styles from "./Roulette.module.scss";
import { roulettePlaceBet, rouletteGetHistory, getMe } from "@/requests";
import toast from "react-hot-toast";
import useStore from "@/store";

const sectors = 15; // Количество секторов

export const Roulette = () => {
	const { increaseBalanceRupee, decreaseBalanceRupee } = useStore();
	const [bet, setBet] = useState(100);
	const [wins, setWins] = useState([]);
	const [isBettingOpen, setIsBettingOpen] = useState(true);
	const [rotation, setRotation] = useState(0);
	const [mustSpin, setMustSpin] = useState(false);

	useEffect(() => {
		const fetchHistory = async () => {
			try {
				const data = await rouletteGetHistory();
				const lastWins = data.map(result => result.WinningColor);
				setWins(lastWins);
			} catch (error) {
				console.error("Error fetching game history:", error);
			}
		};

		fetchHistory();
		getMe();
	}, []);

	const handleBet = async (color) => {
		if (!isBettingOpen) {
			toast.error("Betting is closed");
			return;
		}
		try {
			const response = await roulettePlaceBet(bet, color);
			if (response.status === 200) {
				const data = await response.json();

				if (color === "black") color = "yellow";
				toast.success(`Bet placed: ${bet} on ${color}`);
				decreaseBalanceRupee(bet);

				// Вычисляем угол вращения
				const newPrizeSectorId = data.winning_number; 
				const sectorAngle = 360 / sectors;
				const randomExtraTurns = Math.floor(Math.random() * 5) + 3; // Добавим случайные обороты
				const newRotation = randomExtraTurns * 360 + (sectors - newPrizeSectorId) * sectorAngle;

				setIsBettingOpen(false);
				setMustSpin(true);
				setRotation(newRotation);

				setTimeout(() => {
					setWins((prevWins) => [data.winning_color, ...prevWins].slice(0, 10));
					setIsBettingOpen(true);
					setMustSpin(false);

					if (data.outcome === "win") {
						toast.success(`You won! +${data.payout}`);
						increaseBalanceRupee(data.payout);
					} else {
						toast.error(`You lost!`);
					}
				}, 6000);
			} else {
				const data = await response.json();
				toast.error(data.error);
			}
		} catch (err) {
			console.error(err.message);
		}
	};

	const handleIncreaseBet = () => {
		setBet(prev => prev + 1);
	};
	
	const handleDecreaseBet = () => {
		setBet(prev => prev > 1 ? prev - 1 : 1);
	};
	
	const handleDivideBet = () => {
		setBet(prev => Math.max(Math.floor(prev / 2), 1));
	};
	
	const handleMultiplyBet = () => {
		setBet(prev => prev * 2);
	};

	return (
		<div className={styles.roulette}>
			<div className={styles.roulette__wheel}>
				<img
					src="/wheel.png"
					alt="Roulette Wheel"
					className={styles.wheel}
					style={{ transform: `rotate(${rotation}deg)`, transition: mustSpin ? "6s ease-out" : "none" }}
				/>
				<img src="/roulette_arrow.svg" alt="Arrow" className={styles.arrow} />
			</div>

			<div className={styles.roulette__bets}>
				<p className={styles.roulette__bets_title}>Last wins</p>
				<div className={styles.roulette__wins_container}>
					{wins.length > 0 ? (
						wins.map((win, index) => {
							const displayWinColor = win === "black" ? "yellow" : win;
							return (
								<div key={index} className={`${styles.roulette__bets__button} ${styles[`roulette__bets__button_${displayWinColor.toLowerCase()}`]}`}>
									<img src="/24=poker_chip.svg" alt="" />
									{displayWinColor.charAt(0).toUpperCase() + displayWinColor.slice(1)}
								</div>
							);
						})
					) : (
						<p className={styles.nowins} style={{ width: "100%", textAlign: "center", color: "#DADADA" }}>
							No wins
						</p>
					)}
				</div>
			</div>

			<div className={styles.roulette__bets_wrapper}>
				{/* <Amount bet={bet} setBet={setBet} /> */}
				<div className={styles.roulette__input}>
					<span className={styles.leftContainer}>{bet}₹</span>
					<span className={styles.rightContainer}>
						<div onClick={handleDecreaseBet}>-</div>
						<div onClick={handleIncreaseBet}>+</div>
					</span>
				</div>
				<div className={styles.roulette__downbuttons}>
					<button onClick={handleDivideBet} className={styles.decrimentBet}>/2</button>
					<button onClick={handleMultiplyBet} className={styles.incrimentBet}>x2</button>
				</div>
			</div>

			<div className={`${styles.roulette__bets_container} ${styles.roulette__bets_container_buttons}`}>
				<div className={`${styles.roulette__bets__button} ${styles.roulette__bets__button_red}`} onClick={() => handleBet("red")}>
					Red
				</div>
				<div className={`${styles.roulette__bets__button} ${styles.roulette__bets__button_green}`} onClick={() => handleBet("green")}>
					Green
				</div>
				<div className={`${styles.roulette__bets__button} ${styles.roulette__bets__button_yellow}`} onClick={() => handleBet("black")}>
					Yellow
				</div>
			</div>
		</div>
	);
};
