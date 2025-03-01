import { GameCard } from "@/components";
import styles from "./Games.module.scss";

export const Games = () => {

    return (
        <div className={styles.games}>
            <h3 className={styles.games_title}>Games</h3>
            <p className={styles.games_text}>Test your luck and win big!</p>

            <div className={styles.games_cards}>
                <GameCard to="/games/dice" src="/game_dice.png" label="Roll" color="#038168" />
                <GameCard to="/games/roulette" src="/game_roulette.png" label="Roulette" color="#FEA205" />
                <GameCard to="/games/crash" src="/game_crash.png" label="Star Crash" color="#124AC6" />
                <GameCard to="/games/nvuti" src="/game_case.png" label="Vault" color="#E659FD" />
                
                <p className={styles.coming_soon_text}>Coming soon</p>
                
                <GameCard to="/games/coinflip" src="/game_coinflip.png" label="Coin" color="#FFFFFF" soon />
                <GameCard to="/games/tickets" src="/game_tickets.png" label="LuckyTicket" color="#FFFFFF" soon />
                <GameCard to="/games/plinko" src="/game_plinko.png" label="Plinko" color="#FFFFFF" soon />
                <GameCard to="/games/cards" src="/game_cards.png" label="RoyalCard" color="#FFFFFF" soon />
                <GameCard to="/games/mines" src="/game_mines.png" label="Mines" color="#FFFFFF" soon />
                <GameCard to="/games/case" src="/game_case.png" label="Cases" color="#FFFFFF" soon />
            </div>
        </div>
    );
};
