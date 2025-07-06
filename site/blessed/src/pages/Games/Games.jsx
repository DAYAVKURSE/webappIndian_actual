import { GameCard } from '@/components';
import styles from './Games.module.scss';

export const Games = () => {
  return (
    <div className={styles.games}>
      <h3 className={styles.games_title}>Games</h3>
      <p className={styles.games_text}>Test your luck and win big!</p>

      <div className={styles.games_cards}>
        <GameCard to="/games/dice" src="/game_dice.png" label="Roll" />
        <GameCard to="/games/crash" src="/game_crash.png" label="Star Crash" />
        <GameCard to="/games/roulette" src="/game_roulette.png" label="Roulette" />
        <GameCard to="/games/nvuti" src="/game_vault.png" label="Vault" />

        <p className={styles.coming_separator}></p>

        <p className={styles.coming_soon_text}>Coming soon</p>

        <GameCard to="/games/coinflip" src="/game_coinflip.png" label="Coin" soon />
        <GameCard to="/games/tickets" src="/game_tickets.png" label="LuckyTicket" soon />
        <GameCard to="/games/plinko" src="/game_plinko.png" label="Plinko" soon />
        <GameCard to="/games/cards" src="/game_cards.png" label="RoyalCard" soon />
        <GameCard to="/games/mines" src="/game_mines.png" label="Mines" soon />
        <GameCard to="/games/case" src="/game_case.png" label="Cases" soon />
      </div>
    </div>
  );
};
