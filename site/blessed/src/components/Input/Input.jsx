import styles from "./Input.module.scss";

export const Input = ({ id, type, onChange, placeholder, disabled, center, min, max, value, name }) => {
    return (
        <div className={styles.input__container}>
            <input
                className={styles.input}
                type={type || "text"}
                onChange={onChange}
                placeholder={placeholder}
                disabled={disabled}
                id={id}
                min={min !== undefined ? min : undefined}
                max={max !== undefined ? max : undefined}
                value={value}
                name={name}
                style={{ textAlign: center ? "center" : "left" }}
            />
        </div>
    );
};
