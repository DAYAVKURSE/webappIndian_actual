import styles from "./Accordion.module.scss";
import { useState, useRef } from "react";
import { Link } from "react-router-dom";

export const Accordion = ({ icon, title, description, to, filter }) => {
    const [isOpen, setIsOpen] = useState(false);
    const contentRef = useRef(null);
    const Component = to ? Link : 'div';

    return (
        <Component className={styles.accordion} to={to}>
            <div
                className={styles.accordion_question}
                onClick={() => setIsOpen(!isOpen)}
            >
                <div className={styles.accordion_question_title}>
                    {icon && <img src={icon} alt="icon" style={{ filter: filter }} />}
                    <h3 style={{ filter: filter }}>{title}</h3>
                </div>
                <img src="/24=arrow_forward.svg" alt="arrow" style={{ rotate: isOpen ? "90deg" : "", filter: filter }} />
            </div>
            {!to && (
                <div
                    className={styles.accordion__answer}
                    style={{
                        maxHeight: isOpen ? `${contentRef.current.scrollHeight}px` : "0px",
                    }}
                    ref={contentRef}
                >
                    <div>{description}</div>
                </div>
            )}
        </Component>
    );
};
