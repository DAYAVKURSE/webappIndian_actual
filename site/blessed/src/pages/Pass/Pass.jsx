import { useEffect, useState, useRef } from "react";
import styles from "./Pass.module.scss";
import { travepass, getTravepassProgress, getTravepassRequirements } from "@/requests";
import Carousel from 'react-multi-carousel';
import 'react-multi-carousel/lib/styles.css';

const benefitData = {
    "benefit_credit": { image: "/onboarding_render.png", color: "#EDFF8C" },
    "benefit_fortune_wheel": { image: "/fortune_wheel.png", color: "#ACB5FF" },
    "benefit_clicker": { image: "/BCoin.png", color: "#FFD689" },
    "benefit_binary_option": { image: "/benefit_binary_option.png", color: "#FF8888" },
    "benefit_mini_game": { image: "/benefit_mini_game.png", color: "#88FF88" },
    "benefit_item": { image: "/benefit_item.png", color: "#FFBB88" },
    "benefit_replenishment": { image: "/benefit_replenishment.png", color: "#88BBFF" },
};

const itemImages = {
    "Chevrolet Camaro 2024, valued at 2,700,000 INR": "/camaro.png",
    "Unlock the chance to win a MacBook Air 13 M2 256GB": "/macbook.png",
    "iPhone 15 256gb": "/iphone.png",
};


const requirementIcons = {
    "requirement_replenishment": "/24=rupee.svg",
    "requirement_binary_option": "/24=PassGO.svg",
    "requirement_turnover": "/24=refresh.svg",
    "requirement_clicker": "/24=Clicker.svg",
    "requirement_mini_game": "/24=Fortune.svg",
    "requirement_exchange": "/24=arrow_circle_up.svg",
};

const gameIDs = {
    0: "any game",
    1: "Nvuti",
    2: "Dice",
    3: "Roulette",
};

export const Pass = () => {
    const [levels, setLevels] = useState([]);
    const [progress, setProgress] = useState([]);
    const [currentLevelIndex, setCurrentLevelIndex] = useState(1);
    const [currentLevel, setCurrentLevel] = useState(null);
    const [loading, setLoading] = useState(true);
    const isProgrammaticChange = useRef(false);
    const sliderRef = useRef(null);
    const carouselRef = useRef(null);
    const div = (60) / 100;

    useEffect(() => {
        const fetchLevels = async () => {
            try {
                const response = await travepass();
                setLevels(response);
            } catch (error) {
                console.error("Error fetching TravePass data:", error);
            } finally {
                setLoading(false);
            }
        };
        fetchLevels();

        const fetchProgress = async () => {
            try {
                const response = await getTravepassProgress();
                setProgress(response);
            } catch (error) {
                console.error("Error fetching TravePass progress:", error);
            }
        };
        fetchProgress();

        const fetchTravepassRequirements = async () => {
            try {
                const response = await getTravepassRequirements();
                setCurrentLevelIndex(response[0]?.TravePassLevelID);
                setCurrentLevel(response[0]?.TravePassLevelID);

                setTimeout(() => {
                    const slider = sliderRef.current;
                    if (slider) {
                        slider.style.setProperty('--value', `${response[0]?.TravePassLevelID / div}%`);
                    }
                    carouselRef.current.goToSlide(response[0]?.TravePassLevelID - 1);
                }, 500);
            } catch (error) {
                console.error("Error fetching TravePass progress:", error);
            }
        };
        fetchTravepassRequirements();
    }, [div]);


    useEffect(() => {
        const slider = sliderRef.current;
        if (slider) {
            slider.style.setProperty('--value', `${slider.value / div}%`);

            const handleInput = () => {
                slider.style.setProperty('--value', `${slider.value / div}%`);
            };

            slider.addEventListener('input', handleInput);

            return () => {
                slider.removeEventListener('input', handleInput);
            };
        }
    }, [div, currentLevelIndex]);


    const getRequirementProgress = (requirement, isCurrentLevel) => {
        let progressData;
        if (isCurrentLevel) {
            progressData = progress.find(
                (prog) => prog.RequirementID === requirement.ID
            );
        } else {
            progressData = null;
        }

        let description = "";
        const icon = requirementIcons[requirement.PolymorphicRequirementType] || "/icons/default.png";

        if (progressData) {
            const reqType = requirement.PolymorphicRequirementType;
            const reqProgress = progressData.PolymorphicRequirementProgress;

            switch (reqType) {
                case "requirement_replenishment":
                    description = `Replenishment Progress: ${reqProgress.CurrentReplenishmentRupee}/${requirement.PolymorphicRequirement.AmountRupee}`;
                    return {
                        text: description,
                        icon: icon,
                    };

                case "requirement_binary_option": {
                    const { BetsAmount, WinsAmount, MinBetRupee } = requirement.PolymorphicRequirement;
                    const betText = BetsAmount > 0 ? `Binary options bets: ${reqProgress.BetsAmount || 0}/${BetsAmount}` : "";
                    const winText = WinsAmount > 0 ? `, Wins: ${reqProgress.WinsAmount || 0}/${WinsAmount}` : "";
                    const minBetText = MinBetRupee > 0 ? ` (Min bet: ${MinBetRupee} Rupees)` : "";

                    return {
                        text: betText + winText + minBetText,
                        icon: icon,
                    };
                }

                case "requirement_turnover": {
                    const { AmountRupee, TimeDuration } = requirement.PolymorphicRequirement;
                    const turnoverText = AmountRupee > 0 ? `Turnover: ${reqProgress.CurrentTurnoverRupee || 0}/${AmountRupee}` : "";
                    const timeText = TimeDuration > 0 ? ` within ${TimeDuration / 3600}h` : "";

                    return {
                        text: turnoverText + timeText,
                        icon: icon,
                    };
                }

                case "requirement_clicker":
                    description = requirement.PolymorphicRequirement.ClicksAmount > 0
                        ? `Clicks: ${reqProgress.CurrentClicksAmount || 0}/${requirement.PolymorphicRequirement.ClicksAmount}`
                        : "Hit daily click limit";

                    return {
                        text: description,
                        icon: icon,
                    };

                case "requirement_mini_game": {
                    const { BetsAmount, WinsAmount } = requirement.PolymorphicRequirement;
                    const betText = BetsAmount > 0 ? `Bets: ${reqProgress.CurrentBetsAmount || 0}/${BetsAmount}` : "";
                    const winText = WinsAmount > 0 ? `, Wins: ${reqProgress.CurrentWinsAmount || 0}/${WinsAmount}` : "";

                    return {
                        text: betText + winText,
                        icon: icon,
                    };
                }

                case "requirement_exchange":
                    description = requirement.PolymorphicRequirement.BCoinsAmount > 0
                        ? `BCoins Exchanged: ${reqProgress.CurrentBCoinsAmount || 0}/${requirement.PolymorphicRequirement.BCoinsAmount}`
                        : "";

                    return {
                        text: description,
                        icon: icon,
                    };

                default:
                    return {
                        text: "unknown",
                        icon: "unknown",
                    }
            }
        }

        const reqType = requirement.PolymorphicRequirementType;
        switch (reqType) {
            case "requirement_replenishment":
                description = requirement.PolymorphicRequirement.AmountRupee > 0
                    ? `Replenishment required: ${requirement.PolymorphicRequirement.AmountRupee} Rupees`
                    : "";

                return {
                    text: description,
                    icon: icon,
                };

            case "requirement_binary_option": {
                const { BetsAmount, WinsAmount, MinBetRupee, TotalWinningsRupee } = requirement.PolymorphicRequirement;
                const betText = BetsAmount > 0 ? `Place ${BetsAmount} bets` : "";
                const winText = WinsAmount > 0 ? `Win ${WinsAmount} times` : "";
                const total = TotalWinningsRupee > 0 ? `Earn ${TotalWinningsRupee} Ruppes` : "";
                const minBetText = MinBetRupee > 0 ? ` (min bet of ${MinBetRupee} Rupees)` : "";
                const otherText = " in Binary Options";

                return {
                    text: betText + winText + minBetText + total + otherText,
                    icon: icon,
                };
            }

            case "requirement_turnover": {
                const { AmountRupee, TimeDuration } = requirement.PolymorphicRequirement;
                const turnoverText = AmountRupee > 0 ? `Turnover required: ${AmountRupee} Rupees` : "";
                const timeText = TimeDuration > 0 ? ` within ${TimeDuration / 3600}h` : "";

                return {
                    text: turnoverText + timeText,
                    icon: icon,
                };
            }

            case "requirement_clicker":
                description = requirement.PolymorphicRequirement.ClicksAmount > 0
                    ? `Clicks required: ${requirement.PolymorphicRequirement.ClicksAmount}`
                    : "Hit daily click limit";

                return {
                    text: description,
                    icon: icon,
                };

            case "requirement_mini_game": {
                const { BetsAmount, WinsAmount } = requirement.PolymorphicRequirement;
                const betText = BetsAmount > 0 ? `Place ${BetsAmount} bets in ${gameIDs[requirement.PolymorphicRequirement.GameID]}` : "";
                const winText = WinsAmount > 0 ? `, win ${WinsAmount} times` : "";

                return {
                    text: betText + winText,
                    icon: icon,
                };
            }

            case "requirement_exchange":
                description = requirement.PolymorphicRequirement.BCoinsAmount > 0
                    ? `Exchange ${requirement.PolymorphicRequirement.BCoinsAmount} BCoins for Rupees`
                    : "";

                return {
                    text: description,
                    icon: icon,
                };

            default:
                return {
                    text: "unknown",
                    icon: "unknown",
                }
        }
    }

    const getBenefitDescription = (benefit) => {
        const benefitType = benefit.PolymorphicBenefitType;
        const data = benefit.PolymorphicBenefit;

        switch (benefitType) {
            case "benefit_binary_option": {
                const { FreeBetsAmount, FreeBetDepositRupee } = data;
                const freeBetsText = FreeBetsAmount > 0 ? `Free Bets: ${FreeBetsAmount}` : "";
                const depositText = FreeBetDepositRupee > 0 ? `Deposit per Bet: ${FreeBetDepositRupee} Rupees` : "";
                return `${freeBetsText}${freeBetsText && depositText ? " | " : ""}${depositText}`;
            }

            case "benefit_clicker": {
                const { TimeDuration, BonusMultiplier, Reset } = data;
                const timeText = TimeDuration > 0 ? `Duration: ${TimeDuration / 3600}h` : "";
                const bonusText = BonusMultiplier > 0 ? `Multiplier: ${BonusMultiplier}x` : "";
                const resetText = Reset ? "Resets daily clicks limit" : "";
                return `
                    Clicker:
                    ${bonusText}
                    ${bonusText && timeText ? " | " : ""}
                    ${timeText}
                    ${(timeText || bonusText) && resetText ? " | " : ""}
                    ${resetText}
                `
            }

            case "benefit_credit": {
                const { BCoinsAmount, RupeeAmount } = data;
                const bcoinsText = BCoinsAmount > 0 ? `BCoins: ${BCoinsAmount}` : "";
                const rupeeText = RupeeAmount > 0 ? `Rupees: ${RupeeAmount}` : "";
                return `Reward: ${bcoinsText}${bcoinsText && rupeeText ? " | " : ""}${rupeeText}`;
            }

            case "benefit_mini_game": {
                const { GameID, FreeBetsAmount, FreeDepositRupee } = data;
                const gameText = GameID > 0 ? gameIDs[GameID] : "";
                const freeBetsText = FreeBetsAmount > 0 ? `Free Bets: ${FreeBetsAmount}` : "";
                const depositText = FreeDepositRupee > 0 ? `Deposit per Bet: ${FreeDepositRupee} Rupees` : "";
                return `${gameText}${gameText && freeBetsText ? " | " : ""}${freeBetsText}${(gameText || freeBetsText) && depositText ? " | " : ""}${depositText}`;
            }

            case "benefit_fortune_wheel": {
                const { FreeSpinsAmount } = data;
                const freeSpinsText = FreeSpinsAmount > 0 ? `Fortune Wheel free spins: ${FreeSpinsAmount}` : "";
                return freeSpinsText;
            }

            case "benefit_item": {
                const { ItemName } = data;
                return ItemName ? `${ItemName}` : "";
            }

            case "benefit_replenishment": {
                const { BonusMultiplier, TimeDuration } = data;
                const bonusText = BonusMultiplier > 0 ? `Bonus: ${BonusMultiplier * 100}%` : "";
                const timeText = TimeDuration > 0 ? `Duration: ${TimeDuration / 3600}h` : "";
                return `${bonusText}${bonusText && timeText ? " | " : ""}${timeText}`;
            }

            default:
                return "Unknown benefit type";
        }
    };



    const responsive = {
        all: {
            breakpoint: { max: 4000, min: 0 },
            items: 3,
            partialVisibilityGutter: 0,
            slidesToSlide: 1,
        }
    };

    const carouselItems = levels.map((levelData) => {
        const benefits = levelData?.Benefits?.[0]?.Benefit || {};
        const name = levelData?.Benefits?.[0]?.Benefit?.PolymorphicBenefit?.ItemName
        const isCurrent = levelData.ID === currentLevelIndex;

        return {
            levelData,
            name,
            benefits,
            isCurrent
        };
    });

    const currentSlideIndex = carouselItems.findIndex(item => item.levelData.ID === currentLevelIndex);

    const handleAfterChange = (previousSlide, state) => {
        if (isProgrammaticChange.current) {
            isProgrammaticChange.current = false;
            return;
        }
        const currentSlide = state.currentSlide + 1;
        const newCurrentLevelIndex = carouselItems[currentSlide]?.levelData?.ID;
        if (newCurrentLevelIndex && newCurrentLevelIndex !== currentLevelIndex) {
            setCurrentLevelIndex(newCurrentLevelIndex);
        }
    };


    const handleItemClick = (index) => {
        carouselRef.current.goToSlide(index - 1);
    };

    const handleSliderChange = (event) => {
        const newValue = parseInt(event.target.value, 10);
        setCurrentLevelIndex(newValue);
    };

    const handleSliderRelease = () => {
        if (carouselRef.current) {
            isProgrammaticChange.current = true;
            carouselRef.current.goToSlide(currentLevelIndex - 1);
        }
    };

    const getItemImage = (itemName, defaultImage = "/default_image.png") => {
        return itemImages[itemName] || defaultImage;
    };

    const currentLevelData = levels.find(level => level.ID === currentLevelIndex);
    const currentBenefits = currentLevelData?.Benefits?.[0]?.Benefit || {};

    return (
        <div className={styles.pass}>
            <h2 className={styles.pass_title}>Trave Pass</h2>
            <p className={styles.pass_text}>Complete tasks, earn rewards!</p>

            <div className={styles.pass_level_container}>
                <div className={styles.pass_level}>
                    <p>{`Level ${currentLevelIndex}`}</p>
                </div>
            </div>

            <Carousel
                ref={carouselRef}
                responsive={responsive}
                swipeable={false}
                draggable={false}
                showDots={false}
                arrows={false}
                ssr={true}
                infinite={false}
                autoPlay={false}
                keyBoardControl={false}
                containerClass={styles.carouselContainer}
                itemClass={styles.carouselItem}
                afterChange={handleAfterChange}
                customTransition="transform 300ms ease-in-out"
                transitionDuration={500}
                initialSlide={currentSlideIndex - 1}
                slidesToSlide={1}
                centerMode={false}
            >
                {carouselItems.map((item, index) => {
                    const isCenter = item.levelData.ID === currentLevelIndex;

                    return (
                        <div
                            key={index}
                            className={`${styles.carouselItem} ${isCenter ? styles.carouselItemActive : ''}`}
                            onClick={() => handleItemClick(index)}
                        >
                            {item.benefits.PolymorphicBenefitType === "benefit_item" ? (
                                <img
                                    src={getItemImage(item.name)}
                                    alt={item.benefits.PolymorphicBenefitType}
                                />
                            ) : item.benefits.PolymorphicBenefitType === "benefit_credit" ? (
                                item.benefits.PolymorphicBenefit.BCoinsAmount > 0 ? (
                                    <img
                                        src="/BCoin.png"
                                        alt="BCoins"
                                    />
                                ) : item.benefits.PolymorphicBenefit.RupeeAmount > 0 ? (
                                    <img
                                        src="/Rupee.png"
                                        alt="Rupees"
                                    />
                                ) : (
                                    <img
                                        src="/BCoin_old.png"
                                        alt="BCoin_old"
                                    />
                                )
                            ) : item.benefits.PolymorphicBenefitType === "benefit_clicker" ? (
                                <img
                                    src="/BCoin_old.png"
                                    alt="BCoin_old"
                                />
                            ) : (
                                benefitData[item.benefits.PolymorphicBenefitType]?.image && (
                                    <img
                                        src={benefitData[item.benefits.PolymorphicBenefitType]?.image || "/default_image.png"}
                                        alt={item.benefits.PolymorphicBenefitType}
                                    />
                                )
                            )}
                        </div>
                    );
                })}
            </Carousel>



            <div className={styles.pass_container}>
                <input
                    id="slider"
                    type="range"
                    min="1"
                    max={levels.length - 1}
                    value={currentLevelIndex}
                    onChange={handleSliderChange}
                    onMouseUp={handleSliderRelease}
                    onTouchEnd={handleSliderRelease}
                    className={styles.slider}
                    step={1}
                    ref={sliderRef}
                />
            </div>


            <div className={styles.pass_container}>
                <div className={styles.pass_benefit}>
                    <p>{getBenefitDescription(currentBenefits)}</p>
                </div>

                <div className={styles.pass_requirements}>
                    <h3>Complete all the requirements to advance to the next level.</h3>
                    {currentLevelData?.Requirements.map((requirementWrapper, index) => {
                        const { Requirement } = requirementWrapper;
                        const requirementData = getRequirementProgress(Requirement, true);

                        return (
                            <div key={index} className={styles.pass_requirement}>
                                <img src={requirementData.icon} alt={Requirement.PolymorphicRequirementType} className={styles.requirementIcon} />
                                <p>{requirementData.text}</p>
                            </div>
                        );
                    })}
                </div>
            </div>
        </div>
    );
};
