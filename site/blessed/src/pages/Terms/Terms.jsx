import styles from "./Terms.module.scss";

export const Terms = () => {
    const BiTraveMail = "bitrave@bitrave.com"
    return (
        <>
            <div className={styles.terms}>
                <header className={`${styles.terms__header} ${styles.terms__header__phone}`}>
                    <a href="/landing" className={styles.logo}>
                        <img src="/logo.svg" alt="" width={42} height={26}/>
                        BiTrave
                    </a>
                    <div className={styles.nav__action}>
                        <a href="https://t.me/BiTRave_bot" className={styles.nav__action__link}>Try Now</a>
                    </div>
                </header>
                <header className={styles.terms__header}>
                    <a href="/landing" className={styles.logo}>
                        <img src="/logo.svg" alt="" width={42} height={26}/>
                        BiTrave
                    </a>
                    <nav className={styles.nav}>
                        <button className={styles.nav__button}><a href="/landing">Home</a></button>
                    </nav>
                    <div className={styles.nav__action}>
                        <a href="https://t.me/BiTRave_bot" className={styles.nav__action__link}>Try Now</a>
                    </div>
                </header>
            </div>
            <div className={styles.terms}>
                <h1>Terms of use (BiTRave)</h1>

                <h2>User Agreement</h2>
                <p>PLEASE READ THESE TERMS CAREFULLY BEFORE ACCEPTING THEM.</p>
                <p>By using our Services, you acknowledge:</p>
                <ul>
                    <li>You enter into a binding agreement with BiTRave and accept these Terms and Conditions;</li>
                    <li>You agree to all applicable terms, policies, and guidelines published on our Website.</li>
                </ul>

                <p>When engaging with BiTRave’s betting services, you also confirm that:</p>
                <ul>
                    <li>You meet the minimum legal age to use our services in your country of residence;</li>
                    <li>You are not located in restricted areas;</li>
                    <li>You are 18 years or older;</li>
                    <li>You are playing solely for yourself, using your own funds;</li>
                    <li>This is your first and only registration on BiTRave;</li>
                    <li>You have not self-excluded from any gambling platform in the past 12 months;</li>
                    <li>You understand and accept the terms and conditions fully;</li>
                    <li>All personal information provided by you is accurate, verifiable, and truthful. Providing incorrect information will result in the voiding of any winnings;</li>
                    <li>BiTRave sets specific limits on the maximum amount that can be won and withdrawn daily/weekly/monthly.</li>
                </ul>

                <h3>General Terms and Conditions</h3>
                <ol>
                    <li><strong>Service Usage:</strong> By accessing BiTRave, the user acknowledges and agrees to follow the platform's rules and policies. Users must be 18 years or older, reside in India, and meet the legal age for online gaming.</li>
                    <li><strong>Account Creation:</strong> Users can register using valid personal information (email, phone number). Each user is allowed only one account. Any duplicate accounts will be closed.</li>
                    <li><strong>User Responsibility:</strong> Users are responsible for the accuracy of their personal data. Misrepresentation may lead to account suspension.</li>
                    <li><strong>Fraud and Misuse:</strong> BiTRave actively monitors for any suspicious activity, such as multi-accounting, using bots, or other fraudulent practices. Accounts involved in such activities will be suspended.</li>
                    <li><strong>Bonuses and Rewards:</strong> Users can access special bonuses as part of the loyalty program, including deposit bonuses and rewards earned through gameplay. Withdrawal of bonus funds is possible only after fulfilling all wagering requirements.</li>
                    <li><strong>Data Privacy:</strong> BiTRave adheres to strict data privacy policies, ensuring that personal information is protected and used only for service provision.</li>
                    <li><strong>Force Majeure:</strong> BiTRave shall not be held liable for service interruptions caused by external factors beyond control, such as technical malfunctions or natural disasters.</li>
                    <li><strong>Disputes:</strong> Any disputes or disagreements related to account activity or services offered must be submitted to BiTRave's support within 10 days for resolution.</li>
                </ol>

                <h3>BiTRave Privacy Policy</h3>
                <p>This Privacy Policy applies to BiTRave’s services, including our online and mobile platforms. It explains how we collect, use, and protect your personal data.</p>

                <h4>1. Introduction</h4>
                <p>By ‘Personal Data,’ we refer to identifiable information such as your name, email, phone number, payment details, user activity, and support requests. Aggregated or anonymized data not linked to an individual does not fall under this policy.</p>
                <p>We may update this policy periodically, and any significant changes will be communicated via notification on our platform or by email.</p>

                <h4>2. Data Collection</h4>
                <p>We collect personal data through the following methods:</p>
                <h5>a. Information You Provide</h5>
                <p>When registering or using our services, we collect details like your contact information, payment details, and identity verification documents. This also includes any data provided when contacting support.</p>
                <h5>b. Automatic Data Collection</h5>
                <p>When you access BiTRave, we automatically gather information about your device, IP address, browser settings, operating system, and user activity. This helps us improve your experience on our platform. Cookies and third-party analytics tools such as Google Analytics assist in collecting this information.</p>
                <h5>c. Information from Third Parties</h5>
                <p>We may gather personal data from third-party sources like payment providers, helping us verify the information and improve your experience.</p>

                <h4>3. Legal Basis for Processing Personal Data</h4>
                <p>We process your data based on the following legal grounds:</p>
                <ul>
                    <li><strong>Performance of a Contract:</strong> To deliver the services you’ve registered for.</li>
                    <li><strong>Legal Obligations:</strong> To comply with legal requirements like anti-money laundering laws.</li>
                    <li><strong>Legitimate Interests:</strong> For operational, administrative, or marketing purposes, including fraud prevention.</li>
                    <li><strong>Consent:</strong> For specific cases like sending you direct marketing communications.</li>
                </ul>

                <h4>4. How We Use Your Data</h4>
                <p>We use your personal data for:</p>
                <ul>
                    <li><strong>Service Delivery:</strong> Ensuring our platform functions properly and delivers the services you need.</li>
                    <li><strong>Account Setup and Verification:</strong> Confirming your identity and eligibility to use our platform.</li>
                    <li><strong>Regulatory Compliance:</strong> Fulfilling legal obligations such as fraud prevention and responsible gaming regulations.</li>
                    <li><strong>Customer Support:</strong> Resolving any issues you encounter.</li>
                    <li><strong>Improvement of Services:</strong> Enhancing your user experience by analyzing and optimizing platform features.</li>
                    <li><strong>Security:</strong> Detecting and preventing any fraudulent or malicious activities.</li>
                    <li><strong>Data Aggregation and Analytics:</strong> Creating reports to improve our services.</li>
                </ul>

                <h4>5. Data Sharing</h4>
                <p>We may share your personal data with:</p>
                <ul>
                    <li><strong>Affiliates and Service Providers:</strong> To deliver and improve our services.</li>
                    <li><strong>Legal Authorities:</strong> If required by law or to protect our legal interests.</li>
                    <li><strong>Trusted Third-Party Partners:</strong> For operational and promotional purposes.</li>
                </ul>

                <h4>6. International Data Transfers</h4>
                <p>Your personal data may be transferred outside your country, including to countries with different data protection laws. We ensure such transfers comply with applicable safeguards, including Standard Contractual Clauses for users from the European Economic Area (EEA).</p>

                <h4>7. Security Measures</h4>
                <p>We employ robust security measures to protect your personal data, including:</p>
                <ul>
                    <li><strong>Data Encryption:</strong> TLS encryption to secure personal and financial information.</li>
                    <li><strong>Restricted Access:</strong> Only authorized personnel have access to your data.</li>
                    <li><strong>Secure Data Centers:</strong> Physical and digital security protocols in our hosting facilities.</li>
                    <li><strong>Continuous Monitoring:</strong> Our systems are monitored for security threats at all times.</li>
                </ul>

                <h4>8. Data Retention</h4>
                <p>We retain your data as needed for the following:</p>
                <ul>
                    <li><strong>Service Continuity:</strong> Data remains in your account until you choose to delete it or for a period of five years after account closure to meet regulatory obligations.</li>
                    <li><strong>Legal Compliance:</strong> To comply with financial regulations and prevent fraud.</li>
                    <li><strong>Support and Issue Resolution:</strong> To ensure that any disputes or issues can be addressed even after account closure.</li>
                </ul>

                <h4>9. Your Rights</h4>
                <p>You have the following rights concerning your personal data:</p>
                <ul>
                    <li><strong>Access:</strong> Request a copy of the personal data we hold about you.</li>
                    <li><strong>Correction:</strong> Request corrections to inaccurate personal information.</li>
                    <li><strong>Erasure:</strong> Ask us to delete your data under certain conditions.</li>
                    <li><strong>Objection:</strong> Object to how we process your data if based on legitimate interests.</li>
                    <li><strong>Withdrawal of Consent:</strong> Withdraw consent where it was the basis for processing.</li>
                    <li><strong>Data Portability:</strong> Request your data in a machine-readable format.</li>
                    <li><strong>Marketing Opt-Out:</strong> Unsubscribe from marketing communications at any time.</li>
                </ul>
                <p>You can exercise your rights by adjusting your account settings or contacting us at support button.</p>

                <h4>10. Google Analytics and Cookies</h4>
                <p>We use Google Analytics to understand how you use BiTRave. Google may place cookies on your browser to track your visits. You can disable cookies in your browser settings to prevent tracking.</p>

                <h4>12. Account Blocking Due to Violations or Suspicious Activity</h4>
                <p>BiTRave reserves the right to block or suspend any user account without prior notification if we detect hacks, unauthorized access attempts, or violations of our terms of service. This includes, but is not limited to, suspicious activities, exploitation of platform vulnerabilities, or any action that compromises the security and fairness of the platform. We may not notify the account holder before taking such action if it’s deemed necessary for the protection of our services and users.</p>

                <h2>Responsible Gaming</h2>
                <p>At BiTRave, we prioritize responsible gaming to ensure a safe and enjoyable environment for all users. Gambling can be a fun activity, but for some, it can become problematic. Our platform encourages players to always maintain control over their gaming behavior.</p>

                <h3>Key Guidelines:</h3>
                <ul>
                    <li>Play responsibly: Remember, gaming should be a leisure activity, not a way to earn money.</li>
                    <li>Set limits: Always determine how much time and money you're willing to spend before playing.</li>
                    <li>Seek help: If you feel your gaming behavior is becoming harmful, resources are available to help.</li>
                </ul>

                <h3>Signs of Gaming Addiction:</h3>
                <ul>
                    <li>You spend more money than you intended.</li>
                    <li>Gaming becomes a way to escape from problems.</li>
                    <li>You feel anxious or irritable when not playing.</li>
                </ul>

                <p>If you or someone you know is facing issues with gambling, self-exclusion tools are available through our platform. Contact our Customer Support at {BiTraveMail} for assistance, including temporary or permanent account closure.</p>
                <p>For additional resources and support:</p>
                <ul>
                    <li><a href="https://www.gamblingtherapy.org/">https://www.gamblingtherapy.org/</a></li>
                    <li><a href="https://www.gamcare.org.uk/">https://www.gamcare.org.uk/</a></li>
                    <li><a href="https://www.gamblersanonymous.org.uk/">https://www.gamblersanonymous.org.uk/</a></li>
                </ul>

                <h2>Refund Policy for BiTRave</h2>
                <h4>1. Non-refundable Transactions:</h4>
                <p>Once funds (including any bonuses) have been used within the application for gameplay or other services, refunds cannot be issued.</p>

                <h4>2. Refund Eligibility:</h4>
                <p>Refund requests will only be considered if submitted within 24 hours of the transaction, or within 30 days if the user claims unauthorized access (e.g., by another individual or a minor).</p>

                <h4>3. Verification Requirement:</h4>
                <p>BiTRave reserves the right to withhold any refund until the user’s identity is verified. Users may be required to provide certified identification or notarized documents. If the requested identification is not provided within 30 days, the refund will not be processed, and the account may be closed with all funds forfeited.</p>

                <h4>4. Fair Play Requirement:</h4>
                <p>Users are required to play games fairly. The use of external aids (such as computer tools, algorithms, or betting systems) to manipulate game outcomes is strictly prohibited. Violations will result in immediate account suspension.</p>

                <h2>Risk Disclosure</h2>
                <p>By participating in the BiTRave platform’s games, you acknowledge the potential risk of losing funds deposited into your account.</p>
                <p>The legality of online gaming varies by jurisdiction, and BiTRave cannot offer legal advice or guarantee the compliance of its services with your local laws. It is your responsibility to determine if using BiTRave is lawful in your location.</p>
                <p>You access and use BiTRave’s services at your own risk, and all games and services are provided without any express or implied guarantees.</p>

                <h2>KYC & AML Policy</h2>
                <p>At BiTRave, we strictly adhere to Anti-Money Laundering (AML) laws and ensure compliance with all relevant regulations to combat illegal activities and terrorism financing. If suspicious activities are detected, we are required to report them to authorities and may freeze the related funds.</p>

                <h3>What We Do:</h3>
                <ul>
                    <li>Monitor for suspicious transactions.</li>
                    <li>Conduct initial and ongoing identity checks.</li>
                    <li>Store verification documents securely.</li>
                </ul>

                <h3>User Obligations:</h3>
                <ul>
                    <li>Comply with all AML laws.</li>
                    <li>Ensure that funds deposited are from legal sources.</li>
                    <li>Provide requested identification documents when necessary.</li>
                </ul>

                <h2>Cancellation Policy</h2>
                <p>Once a transaction is confirmed, it is considered final and cannot be altered or canceled. You can place another transaction to offset losses, but the original action remains intact. All outcomes are calculated based on the rate active at the time of confirmation. Any future changes to rates will not affect the confirmed actions.</p>
                <p>To avoid mistakes, we highly recommend double-checking all transactions before submitting and reviewing confirmations thoroughly.</p>

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
                            <a href="https://t.me/rupexsupport">Support</a>
                        </div>
                    </div>
                </div>
            </footer>
        </>
    );
};
