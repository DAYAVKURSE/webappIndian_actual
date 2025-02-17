import { Accordion } from "@/components"
import styles from "./Faq.module.scss"

export const Faq = () => {
    return (
        <div className={styles.faq}>
            <h2 className={styles.faq_title}>FAQ</h2>
            <p className={styles.faq_description}>Quick answers to common questions</p>
            <Accordion 
                title="1. What is this?" 
                description="Welcome to your new favorite spot! It’s a casino-based platform where the thrill of winning real-world money is just a click away. Are you ready to get lucky?"
            />
            <Accordion 
                title="2. What’s the Clicker?" 
                description="Think of it as your daily bonus! After each deposit, you can snag up to 20% cashback in the form of Bcoins. Visit the Clicker every day for 20 days, and by the end, you could have earned a whopping 200% of your deposit. The more consistent you are, the more you rake in. Ready to claim your daily treasure??"
            />
            <Accordion 
                title="3. How do Exchange and Withdraw?" 
                description="It’s simple! 10,000 Bcoins = 1 rupee. To withdraw, you’ve got to complete the Voyager challenge (5x turnover). For example, if you exchange 1,000,000 Bcoins, you’ll get 100 rupees. You can bet them on any game you like. Make a 500 rupee turnover, and voilà—withdraw all your hard-earned winnings!?"
            />
            <Accordion 
                title="4. What’s Nvuti?" 
                description="Nvuti is all about control and risk. Bet on whether the next number is higher or lower—your choice! The riskier the bet, the bigger the reward. Think you can handle the pressure?"
            />
            <Accordion 
                title="5. What’s the Deal with Dice?" 
                description="Dice is a game of choice and chance. Bet on whether the next number will be higher or lower than your pick. The higher the risk, the greater the reward—play smart and enjoy the thrill of controlling your destiny!"
            />
            <Accordion 
                title="6. Roulette x14: What’s that?" 
                description="It’s your classic roulette with a twist! Three sectors: Red, Yellow, and Green. Bet on any sector or all of them if you're feeling lucky. Will you play it safe or go all-in? Your strategy, your win!"
            />
            <Accordion 
                title="7. Binary Options" 
                description="Get in on the action with Binary Options! Predict the BTC/USD price movement and earn if you nail it. Guess right, and you could cash in big! Ready to put your instincts to the test?"
            />
        </div>
    )
}