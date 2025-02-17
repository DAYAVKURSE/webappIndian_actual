import { useState } from "react";
import { SwitchTransition, CSSTransition } from "react-transition-group";
import styles from "./Onboarding.module.scss";
import { Input, Button } from "@/components";
import { signUp } from "@/requests";
import { useNavigate } from "react-router-dom";
import toast, { Toaster } from "react-hot-toast";
import toastStyles from "@/scss/toast.module.scss";
import video from '/render-h265-vbr.mp4'

const slides = [
    {
        title: "Start Earning",
        description: "Easily exchange the coins you've earned in our Telegram Mini App for real currency.",
        button: "Get started"
    },
    {
        title: "Track your earnings",
        description: "Monitor your progress and see how much you’ve earned over time.",
        button: "Continue"
    },
    {
        title: "Get Paid",
        description: "Transfer your earnings to your bank account quickly and securely.",
        button: "Continue"
    },
    {
        title: "Choose a username",
        description: "Pick a unique name to complete your profile.",
        button: "Continue"
    },
];

export const Onboarding = () => {
    const [tempName, setTempName] = useState();
    const [currentSlide, setCurrentSlide] = useState(0);
    const navigate = useNavigate();

    const nextSlide = () => {
        if (currentSlide < slides.length - 1) {
            setCurrentSlide(currentSlide + 1);
        }
    };

    const skipToEnd = () => {
        setCurrentSlide(slides.length - 1);
    };

    const onSubmit = () => {
        const avatarId = Math.floor(Math.random() * 100);

        signUp({ Nickname: tempName, avatarId })
            .then(({ status, data }) => {
                if (status === 200) {
                    navigate("/clicker");
                } else if (status === 400 || status === 409) {
                    toast.error(data.error);
                } else {
                    toast.error("An unexpected error occurred");
                }
            })
            .catch((error) => {
                toast.error("Failed to register user");
                console.error('Error during sign-up:', error);
            });
    }


    return (
        <>
            <div>
                <Toaster
                    toastOptions={{
                        position: 'top-center',
                        success: {
                            className: toastStyles.toastSuccess,
                            iconTheme: {
                              primary: '#EDFF8C',
                              secondary: '#0B0B0B',
                            },
                        },
                        error: {
                            className: toastStyles.toastError,
                            iconTheme: {
                              primary: '#FFC397',
                              secondary: '#0B0B0B',
                            },
                        },
                        default: {
                            className: toastStyles.toast,
                        },
                    }}
                />
            </div>
            <div className={styles.onboarding}>
                <SwitchTransition>
                    <CSSTransition
                        key={currentSlide}
                        timeout={250}
                        classNames={{
                            enter: styles.slideEnter,
                            enterActive: styles.slideEnterActive,
                            exit: styles.slideExit,
                            exitActive: styles.slideExitActive,
                        }}
                    >
                        <>
                            <video 
                                src={video}
                                type="video/mp4"
                                autoPlay
                                loop
                                muted
                                playsInline
                                height={400}
                            />
                        </>
                    </CSSTransition>
                </SwitchTransition>
                <div className={styles.onboarding__content}>
                    <h1 className={styles.onboarding__content_title}>{slides[currentSlide].title}</h1>
                    <p className={styles.onboarding__content_desc}>{slides[currentSlide].description}</p>
                    <div className={styles.onboarding__dots}>
                        {slides.map((_, index) => (
                            <div
                                key={index}
                                className={`${styles.onboarding__dots_dot} ${currentSlide === index ? styles.onboarding__dots_dot_active : ""}`}
                            />
                        ))}
                    </div>
                    {currentSlide < slides.length - 1 && (
                        <div className={styles.onboarding__buttons}>
                            <Button onClick={nextSlide} label="Next" fill color="lightYellow"/>
                            <Button onClick={skipToEnd} circle icon="24=arrow_forward.svg" color="white" filter />
                        </div>
                    )}
                    {currentSlide === slides.length - 1 && (
                        <div className={styles.onboarding__buttons_input}>
                            <Input onChange={(e) => setTempName(e.target.value)} label="Username" placeholder="type your name here" />
                            {
                                tempName && tempName.length > 3
                                    ? <Button onClick={onSubmit} label="Continue" fill color="lightYellow" />
                                    : <Button label="Continue" fill color="lightYellow" disabled />
                            }
                        </div>
                    )}
                </div>
            </div>
        </>
    );
};
