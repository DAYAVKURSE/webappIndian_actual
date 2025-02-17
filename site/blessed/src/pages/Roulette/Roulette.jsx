import { useEffect, useState, useMemo, useRef } from "react";
import styles from "./Roulette.module.scss";
import { Wheel } from "react-custom-roulette";
import { roulettePlaceBet, rouletteGetHistory, getMe } from "@/requests";
import { Amount } from "@/components";
import toast from "react-hot-toast";
import useStore from "@/store";

const apiData = [
	{ color: "red", sector_id: 1 },
	{ color: "black", sector_id: 8 },
	{ color: "red", sector_id: 2 },
	{ color: "black", sector_id: 9 },
	{ color: "red", sector_id: 3 },
	{ color: "black", sector_id: 10 },
	{ color: "red", sector_id: 4 },
	{ color: "black", sector_id: 11 },
	{ color: "red", sector_id: 5 },
	{ color: "black", sector_id: 12 },
	{ color: "red", sector_id: 6 },
	{ color: "black", sector_id: 13 },
	{ color: "red", sector_id: 7 },
	{ color: "black", sector_id: 14 },
	{ color: "green", sector_id: 0 },
];

const getColorByApiColor = (color) => {
	switch (color) {
		case "red":
			return "#f24156";
		case "black":
			return "#fcc51f";
		case "green":
			return "#91d233";
		default:
			return "#ffffff";
	}
};

export const Roulette = () => {
	const { increaseBalanceRupee, decreaseBalanceRupee } = useStore();
	const [bet, setBet] = useState(100);
	const [wins, setWins] = useState([]);
	const [isBettingOpen, setIsBettingOpen] = useState(true);
	const [prizeNumber, setPrizeNumber] = useState(0);
	const [mustSpin, setMustSpin] = useState(false);

	const previousPrizeNumber = useRef(prizeNumber);

	const wheelData = useMemo(() => {
		return apiData.map((item) => ({
			style: { backgroundColor: getColorByApiColor(item.color) },
		}));
	}, []);

	useEffect(() => {
		const fetchHistory = async () => {
			try {
				const data = await rouletteGetHistory();
				const lastWins = data.map(result => result.WinningColor);
				setWins(lastWins);
			} catch (error) {
				console.error('Error fetching game history:', error);
			}
		};

		fetchHistory();
		getMe();
	}, []);

	useEffect(() => {
		previousPrizeNumber.current = prizeNumber;
	}, [prizeNumber]);

	const handleBet = async (color) => {
		if (!isBettingOpen) {
			toast.error("Betting is closed");
			return;
		}
		try {
			const response = await roulettePlaceBet(bet, color);
			
			if (response.status === 200) {
				const data = await response.json();

				if (color === "black") {
					color = "yellow";
				}

				toast.success(`Bet placed: ${bet} on ${color}`);
				decreaseBalanceRupee(bet)

				const newPrizeSectorId = apiData.findIndex(item => item.sector_id === data.winning_number);
				const newWin = data.winning_color;
				
				setIsBettingOpen(false);
				setMustSpin(true);
				setPrizeNumber(newPrizeSectorId);

				setTimeout(() => {
					setWins((prevWins) => [newWin, ...prevWins].slice(0, 10));
					setIsBettingOpen(true);
					setMustSpin(false);
					
					if (data.outcome === "win") {
						toast.success(`You won! +${data.payout}`);
						increaseBalanceRupee(data.payout);
					} else if (data.outcome === "lose") {
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
	}

	const memoizedWheel = useMemo(() => {
		return (
			<div className={styles.roulette__wheel_rotate}>
				<Wheel
					data={wheelData}
					mustStartSpinning={mustSpin}
					prizeNumber={prizeNumber}
					startingOptionIndex={14}
					outerBorderColor="#000"
					outerBorderWidth={5}
					innerRadius={90}
					innerBorderColor="#000"
					innerBorderWidth={5}
					radiusLineColor="#000"
					radiusLineWidth={5}
					fontSize={20}
					spinDuration={0.5}
					pointerProps={{
						style: {
							display: "none",
						}
					}}
				/>
			</div>
		);
	}, [wheelData, mustSpin, prizeNumber]);

	return (
		<div className={styles.roulette}>
			<div className={styles.roulette__wheel}>
				{memoizedWheel}
				<img src="/roulette_arrow.svg" alt="" width={24} />
			</div>
			<div className={styles.roulette__bets}>
				<p className={styles.roulette__bets_title}>Last wins</p>
				<div className={styles.roulette__wins_container}>
					{wins.length > 0 ? wins.map((win, index) => {
						const displayWinColor = win === 'black' ? 'yellow' : win;

						return (
							<div key={index} className={`${styles.roulette__bets__button} ${styles[`roulette__bets__button_${displayWinColor.toLowerCase()}`]}`}>
								<img src="/24=poker_chip.svg" alt="" />
								{displayWinColor.charAt(0).toUpperCase() + displayWinColor.slice(1)}
							</div>
						);
					}) : <p style={{ width: "100%", textAlign: "center", color: "#DADADA" }}>No wins</p>}

				</div>
			</div>

			<div className={styles.roulette__bets_wrapper}>
				<Amount bet={bet} setBet={setBet} />
			</div>

			<div className={`${styles.roulette__bets_container} ${styles.roulette__bets_container_buttons}`}>
				<div className={`${styles.roulette__bets__button} ${styles.roulette__bets__button_red}`} onClick={() => handleBet("red")}>
					<img src="/24=poker_chip.svg" alt="" />
					Red
				</div>
				<div className={`${styles.roulette__bets__button} ${styles.roulette__bets__button_green}`} onClick={() => handleBet("green")}>
					<img src="/24=poker_chip.svg" alt="" />
					Green
				</div>
				<div className={`${styles.roulette__bets__button} ${styles.roulette__bets__button_yellow}`} onClick={() => handleBet("black")}>
					<img src="/24=poker_chip.svg" alt="" />
					Yellow
				</div>
			</div>
		</div>
	);
};
