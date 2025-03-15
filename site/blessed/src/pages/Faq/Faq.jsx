import { Accordion } from "@/components"
import styles from "./Faq.module.scss"

export const Faq = () => {
    return (
        <div className={styles.faq}>
            <h2 className={styles.faq_title}>FAQ</h2>
            <Accordion 
                title="1. What is this?" 
                description="Welcome to your new favorite spot! It’s a casino-based platform where the thrill of winning real-world money is just a click away. Are you ready to get lucky?"
            />
           
            <Accordion 
                title="2. What’s Vault?" 
                description="Nvuti is all about control and risk. Bet on whether the next number is higher or lower—your choice! The riskier the bet, the bigger the reward. Think you can handle the pressure?"
            />
            <Accordion 
                title="3. What’s the Deal with Dice?" 
                description="Dice is a game of choice and chance. Bet on whether the next number will be higher or lower than your pick. The higher the risk, the greater the reward—play smart and enjoy the thrill of controlling your destiny!"
            />
            <Accordion 
                title="4. Roulette: What’s that?" 
                description="It’s your classic roulette with a twist! Three sectors: Red, Yellow, and Green. Bet on any sector or all of them if you're feeling lucky. Will you play it safe or go all-in? Your strategy, your win!"
            />
            <Accordion 
                title="5. Binary Options" 
                description="Get in on the action with Binary Options! Predict the BTC/USD price movement and earn if you nail it. Guess right, and you could cash in big! Ready to put your instincts to the test?"
            />
        </div>
    )
}