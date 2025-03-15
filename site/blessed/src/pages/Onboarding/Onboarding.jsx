import { useState } from "react";
import styles from "./Onboarding.module.scss";
import { Input } from "@/components";
import { signUp } from "@/requests";
import { useNavigate } from "react-router-dom";

export const Onboarding = () => {
    const [tempName, setTempName] = useState();
    const navigate = useNavigate();

    const onSubmit = () => {
        const avatarId = Math.floor(Math.random() * 100);

        signUp({ Nickname: tempName, avatarId })
            .then(({ status }) => {
                if (status === 200) {
                    navigate("/clicker");
                } 
            })
            .catch((error) => {
                console.error('Error during sign-up:', error);
            });
    }


    return (
        <>
          
            <div className={styles.onboarding}>
            <div>
                <img src="gif.gif"></img>
            </div>
                <div className={styles.onboarding__content}>
                    <h2 className={styles.title}>Choose a username</h2>
                    <p>Pick a unique name to complete your profile</p>

                        <div className={styles.onboarding__buttons_input}>
                            <Input onChange={(e) => setTempName(e.target.value)} label="Username" placeholder="type your name here" />
                            {
                                tempName && tempName.length > 3
                                    ? <button onClick={onSubmit} className={styles.button}>Continue</button>
                                    : <button className={styles.disableButton} disabled>Continue</button>
                            }
                        </div>
                   
                </div>
            </div>
        </>
    );
};
