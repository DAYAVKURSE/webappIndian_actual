import { FooterButton } from "@/components"
import styles from "./Footer.module.scss"

export const Footer = () => {
    return (
        <>
            <div className={styles.footer__spacer} />
            <footer className={styles.footer}>
                <FooterButton icon="/32=clicker.svg" label="Clicker" to="clicker" />
                <FooterButton icon="/32=bid_landscape.svg" label="Trading" to="trading" />
                <FooterButton icon="/32=playing_cards.svg" label="Games" to="games" />
                {/* <FooterButton icon="/32=leaderboard.svg" label="Leaderboard" to="other" /> */}
                <FooterButton icon="/32=more_horiz.svg" label="Other" to="other" />
            </footer>
        </>
    )
}
