import { Accordion } from "@/components";
import styles from "./Howtoplay.module.scss";

export const Howtoplay = () => {
    return (
        <div className={styles.howtoplay}>
            <h2 className={styles.howtoplay_title}>How to play?</h2>
            <p className={styles.howtoplay_description}>Learn the basics, earn coins, and master mini-games</p>

            <Accordion
                title="1. Welcome to BiTRave!"
                description={
                    <>
                        <p>This guide will help you navigate through our exciting features and maximize your experience on the platform. Let’s dive in!</p>
                    </>
                }
            />

            <Accordion
                title="2. Clicker"
                description={
                    <>
                        <p>When you top up your balance with any amount, you can make up to <b>10,000</b> clicks in the Clicker over a span of <b>20 days.</b></p>
                        <p>Each day, you earn <b>10% of your initial deposit</b> through clicks. The coins you accumulate can be exchanged on the <b>Exchange page.</b> However, to withdraw these coins as rupees, you must wager them with a <b>10x wager.</b></p>
                    </>
                }
            />

            <Accordion
                title="3. Trave Pass"
                description={
                    <>
                        <p><b>Trave Pass</b> is our unique leveling system that rewards players as they progress. Complete tasks and earn exclusive rewards along the way!</p>
                    </>
                }
            />

            <Accordion
                title="4. Mini-Games"
                description={
                    <>
                        <p><b>Dice:</b> Test your luck by rolling the dice and aiming for your desired outcome. Simple and fun!</p>
                        <p><b>Roulette x14:</b> This game features 15 fields:</p>
                        <ul>
                            <li><b>1 field</b> pays <b>14x</b> your bet.</li>
                            <li><b>7 fields</b> are <b>yellow</b> and <b>7 fields</b> are red, paying 2x your bet.</li>
                            <li>You can place only <b>one bet</b> at a time in Roulette.</li>
                        </ul>
                        <p><b>Crash:</b></p>
                        <ul>
                            <li>Objective: Bet on a multiplier that increases until it “crashes.”</li>
                            <li>Cash Out: If you cash out before the crash, you win your bet multiplied by the current multiplier. If not, you lose your bet.</li>
                        </ul>
                        <p><b>Nvuti:</b></p>
                        <ul>
                            <li>Objective: Choose your winning probability from 5% to 95%.</li>
                            <li>Winning: A random number is drawn; if it matches one of your chosen numbers, you win!</li>
                        </ul>
                    </>
                }
            />

            <Accordion
                title="5. Referral System"
                description={
                    <>
                        <ul>
                            <li>Invite your friends and earn <b>20%</b> of their deposits!</li>
                            <li>Share your unique link with friends, family, or on social media. When someone signs up and deposits using your link, you’ll earn a percentage of their deposit as a reward!</li>
                        </ul>
                    </>
                }
            />

            <Accordion
                title="6. Withdrawal and Exchange"
                description={
                    <>
                        <ul>
                            <li>To withdraw your funds, navigate to the <b>Wallet</b> section. You can top up, withdraw, or exchange your earnings easily.</li>
                            <li>Follow the on-screen instructions for seamless transactions.</li>
                        </ul>
                    </>
                }
            />

            <Accordion
                title="7. Wheel of Fortune"
                description={
                    <>
                        <ul>
                            <li>You can spin the <b>Wheel of Fortune</b> once per session:</li>
                            <li>Earn spins by winning in the <b>Trave Pass</b> or by topping up your balance with more than <b>5,000 rupees.</b></li>
                        </ul>
                    </>
                }
            />
        </div>
    );
};
