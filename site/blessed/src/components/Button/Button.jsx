import { NavLink } from 'react-router-dom';
import styles from "./Button.module.scss";
import toast from 'react-hot-toast';
import colors from "@/scss/variables.module.scss";

export const Button = ({
    id,
    label,
    icon,
    onClick,
    type = "button",
    disabled = false,
    to,
    focus,
    fill,
    toastType,
    toastText,
    style,
    color,
    filter,
    circle
}) => {
    const Component = to ? NavLink : 'button';

    const handleClick = (e) => {
        if (toastText) {
            const toastMap = {
                success: toast.success,
                error: toast.error,
                default: toast,
            };
            const showToast = toastMap[toastType] || toastMap.default;
            showToast(toastText);
        }

        if (onClick) {
            onClick(e);
        }
    };

    const buttonStyle = {
        ...style,
        backgroundColor: colors[color] || color,
        width: circle ? "48px" : ""
    };

    const combinedClassName = `${styles.button} ${focus ? styles.button__focus : ""} ${fill ? styles.button__fill : ""}`.trim();

    const hasIconAndLabel = icon && label;
    const labelStyle = hasIconAndLabel ? { marginLeft: '6px' } : {};

    return (
        <Component
            id={id}
            value={label}
            className={to ? ({ isActive }) => 
                `${combinedClassName} ${isActive ? styles.active : ""}`.trim() : combinedClassName}
            onClick={handleClick}
            type={to ? undefined : type}
            disabled={disabled ? disabled : undefined }
            to={to}
            style={buttonStyle}
        >
            {icon && (
                <img
                    className={styles.button__icon}
                    src={icon}
                    alt="icon"
                    style={{ filter: filter || "none" }}
                />
            )}
            {label && (
                <span
                    className={styles.button__label}
                    style={{ ...labelStyle, filter: filter || "none" }}
                >
                    {label}
                </span>
            )}
        </Component>
    );
};
