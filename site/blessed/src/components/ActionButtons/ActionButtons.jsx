import styles from "./ActionButtons.module.scss"

export const ActionButtons = ({ onclick1, src1, label1, color1, onclick2, src2, label2, color2 }) => {
    return (
        <div className={styles.actionbuttons}>
            <button 
                className={styles.actionbuttons_button} 
                onClick={onclick1}
            >
                <img src={src1} alt={label1} />
            </button>
            <button 
                className={styles.actionbuttons_button} 
                onClick={onclick2}
            >
                <img src={src2} alt={label2} />
            </button>
        </div>
    )
}