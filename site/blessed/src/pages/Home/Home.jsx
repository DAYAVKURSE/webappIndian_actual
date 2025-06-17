import styles from "./Home.module.scss";

export const Home = () => {
    return (
        <>
            <div className={styles.home__container}>
                <header className={`${styles.home__header} ${styles.home__header__phone}`}>
                    <a href="/landing" className={styles.logo}>
                        <img src="/logo.svg" alt="" width={42} height={26}/>
                        BiTrave
                    </a>
                    <div className={styles.nav__action}>
                        <a href="https://t.me/BiTRave_bot" className={styles.nav__action__link}>Try Now</a>
                    </div>
                </header>
                <header className={styles.home__header}>
                    <a href="/landing" className={styles.logo}>
                        <img src="/logo.svg" alt="" width={42} height={26}/>
                        BiTrave
                    </a>
                    <nav className={styles.nav}>
                        <button className={styles.nav__button}><a href="/landing">Home</a></button>
                        <button className={styles.nav__button}><a href="#howitworks">How it works</a></button>
                    </nav>
                    <div className={styles.nav__action}>
                        <a href="https://t.me/BiTRave_bot" className={styles.nav__action__link}>Try Now</a>
                    </div>
                </header>
                <div className={styles.home__main}>
                    <div className={styles.content}>
                        <div className={styles.content__toncoin}>
                            <img src="/logo.svg" alt="" />
                        </div>

                        <div className={styles.data}>
                            <span className={styles.data__title}>Welcome to the new era of Gaming. Earn, Win, Withdraw.</span>
                            <p className={styles.data__desc}>Experience the first clicker that offers a wide variety of games and earning methods, all of which can be instantly withdrawn.</p>
                        </div>
                    </div>
                    <div className={styles.home__footer}>
                        {/* <a className={`${styles.home__footer__button} ${styles.home__footer__button_orange}`} href="#">Explore Web App</a> */}
                        <a href="https://t.me/BiTRave_bot" className={styles.home__footer__button}>Open in Telegram</a>
                    </div>
                </div>
                <section id="howitworks" className={styles.home__hiw}>
                    <h2 className={styles.home__hiw__title}>
                        {/* <div className={`${styles.pulseCircle} ${styles.pulseCircle1}`}></div> */}
                        How it works
                    </h2>

                    <div className={styles.home__hiw__card}>
                        <div className={`${styles.pulseCircle} ${styles.pulseCircle2}`}></div>
                        <div className={`${styles.pulseCircle} ${styles.pulseCircle3}`}></div>
                        <img src="/1.png" alt="Deposit" className={styles.home__hiw__card_image} />
                        <div className={styles.home__hiw__card_content}>
                            <h3 className={styles.home__hiw__card_title}>Earn Real Currency with Your Clicks</h3>
                            <p className={styles.home__hiw__card_desc}>
                                Easily exchange the coins you&apos;ve earned in our Telegram Mini App for real currency. Your in-app activity directly translates into tangible rewards.
                            </p>
                            <a href="https://t.me/BiTRave_bot" className={styles.home__hiw__card_button}>Start Earning</a>
                        </div>
                    </div>

                    <div className={styles.home__hiw__card}>
                        <div className={`${styles.pulseCircle} ${styles.pulseCircle4}`}></div>
                        <div className={`${styles.pulseCircle} ${styles.pulseCircle5}`}></div>
                        <img src="/2.png" alt="Earn" className={styles.home__hiw__card_image} />
                        <div className={styles.home__hiw__card_content}>
                            <h3 className={styles.home__hiw__card_title}>Play Mini Games and Earn</h3>
                            <p className={styles.home__hiw__card_desc}>
                                Enjoy a variety of mini-games and earn additional rewards. Entertainment and profit go hand in hand here.
                            </p>
                            <a href="https://t.me/BiTRave_bot" className={styles.home__hiw__card_button}>Play</a>
                        </div>
                    </div>

                    <div className={styles.home__hiw__card}>
                        <div className={`${styles.pulseCircle} ${styles.pulseCircle6}`}></div>
                        <div className={`${styles.pulseCircle} ${styles.pulseCircle7}`}></div>
                        <img src="/3.png" alt="Exchange" className={styles.home__hiw__card_image} />
                        <div className={styles.home__hiw__card_content}>
                            <h3 className={styles.home__hiw__card_title}>Trade Binary Options</h3>
                            <p className={styles.home__hiw__card_desc}>
                                Engage in binary options trading right within the app. No need to switch platforms—everything is conveniently at your fingertips.
                            </p>
                            <a href="https://t.me/BiTRave_bot" className={styles.home__hiw__card_button}>Start Trading</a>
                        </div>
                    </div>

                    <div className={styles.home__hiw__card}>
                        <div className={`${styles.pulseCircle} ${styles.pulseCircle8}`}></div>
                        <div className={`${styles.pulseCircle} ${styles.pulseCircle9}`}></div>
                        <img src="/4.png" alt="Exchange" className={styles.home__hiw__card_image} />
                        <div className={styles.home__hiw__card_content}>
                            <h3 className={styles.home__hiw__card_title}>Telegram Native Experience</h3>
                            <p className={styles.home__hiw__card_desc}>
                                Our app is fully integrated within Telegram, offering a seamless and user-friendly experience. All your favorite features are right where you need them, with no need to leave Telegram.
                            </p>
                            <a href="https://t.me/BiTRave_bot" className={styles.home__hiw__card_button}>Open in Telegram</a>
                        </div>
                    </div>

                    <div className={styles.home__hiw__card}>
                        <div className={`${styles.pulseCircle} ${styles.pulseCircle10}`}></div>
                        <img src="/referral.png" alt="Exchange" className={styles.home__hiw__card_image} />
                        <div className={styles.home__hiw__card_content}>
                            <h3 className={styles.home__hiw__card_title}>Referral System — Boost your revenue easily</h3>
                            <p className={styles.home__hiw__card_desc}>
                                Increase your revenue by inviting friends. Earn bonus 20% for every successful referral, and speed up your progress in the Trave Pass.
                            </p>
                            <a href="https://t.me/BiTRave_bot" className={styles.home__hiw__card_button}>Try it out</a>
                        </div>
                    </div>

                    <div className={styles.home__hiw__card}>
                        <img src="/camaro.png" alt="Exchange" className={styles.home__hiw__card_image} />
                        <div className={styles.home__hiw__card_content}>
                            <h3 className={styles.home__hiw__card_title}>Unlock the Ultimate Reward — Chevrolet Camaro 2024</h3>
                            <p className={styles.home__hiw__card_desc}>
                                Progress through the Trave Pass and reach the final level to unlock the grand prize: a brand-new Chevrolet Camaro 2024, valued at 2,700,000 INR. This isn&apos;t just a digital reward—it&apos;s a real-world achievement waiting for you to claim!
                            </p>
                            <a href="https://t.me/BiTRave_bot" className={styles.home__hiw__card_button}>Open in Telegram</a>
                        </div>
                    </div>
                </section>
            </div>
            <footer className={styles.footer__container}>
                <div className={styles.footer}>
                    <div className={styles.footer__wrapper}>
                        <div className={styles.footer__logo}>BiTrave</div>
                        <p className={styles.footer__copyright}>© 2024 BiTRave. </p>
                        <p className={styles.footer__copyright}>© 2024 FirstGen Development (the Republic of Seychelles). All rights reserved. The content on this site is the exclusive property of FirstGen Development Labs Ltd. Unauthorized reproduction, modification, distribution, publication, transmission, or any form of copying is strictly prohibited.</p>
                        <div className={styles.footer_socials}>
                            <a href="/terms">Terms of use</a>
                            <a href="https://t.me/BiTRaveofficial">Telegram</a>

                            <a href="https://t.me/BiTRaveofficial">Support</a>
                        </div>
                    </div>
                </div>
            </footer>
        </>
    );
};
