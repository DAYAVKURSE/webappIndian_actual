import styles from "./FooterButton.module.scss";
import { NavLink } from 'react-router-dom';

export const FooterButton = ({ label, icon, to }) => {
    return (
        <NavLink 
            to={to} 
            className={({ isActive }) => 
                isActive ? `${styles.button} ${styles.active}` : styles.button
            }
        >
            {icon && (<img className={styles.button__icon} src={icon} alt="icon" />)}
            {label}
        </NavLink>
    );
}
