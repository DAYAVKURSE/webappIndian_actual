import { useEffect, useRef, useState } from "react";
import styles from "./Onboarding.module.scss";
import { Input } from "@/components";
import { signUp } from "@/requests";
import { useNavigate } from "react-router-dom";

export const Onboarding = () => {
    const [tempName, setTempName] = useState("");
    const navigate = useNavigate();

    const containerRef = useRef(null);
    const inputRef = useRef(null);
    const outContainerRef = useRef(null);

    const onSubmit = () => {
        const avatarId = Math.floor(Math.random() * 100);

        signUp({ Nickname: tempName, avatarId })
            .then(({ status }) => {
                if (status === 200) {
                    navigate("/clicker");
                }
            })
            .catch((error) => {
                console.error("Error during sign-up:", error);
            });
    };

    useEffect(() => {
        const handleFocus = () => {
            if (containerRef.current) {
                containerRef.current.style.minHeight = "120vh";
                outContainerRef.current.scrollTop = 250;
            }
        };

        const handleClickOutside = () => {
            if (inputRef.current && !inputRef.current.contains(event.target)) {
                containerRef.current.style.minHeight = "100vh";
                outContainerRef.current.scrollTop = 0;
            }
        }

        inputRef.current.addEventListener("click", handleFocus);
        inputRef.current.addEventListener("input", handleFocus);
        document.addEventListener("mousedown", handleClickOutside);

        return () => {
            if (inputRef.current) {
                inputRef.current.removeEventListener("focus", handleFocus);
                inputRef.current.removeEventListener("input", handleFocus);
                inputRef.current.removeEventListener("blur", handleFocus);
            }
        };
    }, []);

    return (
        <div ref={outContainerRef} className={styles.outContainer}>
            <div className={styles.container} ref={containerRef}>
                <div className={styles.onboarding}>
                    <div>
                        <img src="gif.gif" alt="Loading animation" />
                    </div>
                    <div className={styles.onboarding__content}>
                        <h2 className={styles.title}>Choose a username</h2>
                        <p>Pick a unique name to complete your profile</p>

                        <div className={styles.onboarding__buttons_input}>
                            <div ref={inputRef}>
                                <Input
                                    onChange={(e) => setTempName(e.target.value)}
                                    label="Username"
                                    placeholder="Type your name here"
                                />
                            </div>
                            {tempName?.length > 3 ? (
                                <button onClick={onSubmit} className={styles.button}>
                                    Continue
                                </button>
                            ) : (
                                <button className={styles.disableButton} disabled>
                                    Continue
                                </button>
                            )}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};
