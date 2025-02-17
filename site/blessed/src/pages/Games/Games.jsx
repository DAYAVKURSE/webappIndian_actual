import { GameCard } from "@/components";
import styles from "./Games.module.scss";

export const Games = () => {

    return (
        <div className={styles.games}>
            <h3 className={styles.games_title}>Games</h3>
            <p className={styles.games_text}>Test your luck and win big!</p>

            <div className={styles.games_cards}>
                <GameCard to="/games/dice" src="/game_dice.png" label="Dice" color="#FFD689" />
                <GameCard to="/games/nvuti" src="/game_nvuti.png" label="Nvuti" color="#FFD689" />
                <GameCard to="/games/roulette" src="/game_roulette.png" label="Roulette" color="#FFD689" />
                <GameCard to="/games/crash" src="/game_crash.png" label="Crash" color="#FFD689" />
                <div className={styles.games_separator} />
                <GameCard to="/games/plinko" src="/game_plinko.png" label="Plinko" color="#FFD689" soon />
                <GameCard to="/games/tickets" src="/game_tickets.png" label="Tickets" color="#FFC397" soon />
                <GameCard to="/games/case" src="/game_case.png" label="Case" color="#CFA3F2" soon />
                <GameCard to="/games/cards" src="/game_cards.png" label="Cards" color="#FFC397" soon />
                <GameCard to="/games/mines" src="/game_mines.png" label="Mines" color="#CFA3F2" soon />
                <GameCard to="/games/ladder" src="/game_ladder.png" label="Ladder" color="#FFC397" soon />
                <GameCard to="/games/coinflip" src="/game_coinflip.png" label="Coinflip" color="#CFA3F2" soon />
            </div>
        </div>
    );
};
