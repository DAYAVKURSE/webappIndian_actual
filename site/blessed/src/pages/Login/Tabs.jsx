import styles from "./Login.module.scss";

export const Tabs = ({ mode, setMode }) => (
  <div className={styles.tabs} data-mode={mode}>
    <span
      className={`${styles.btnTabs} ${mode === "login" ? styles.active : ""}`}
      onClick={() => setMode("login")}
    >
      Login
    </span>
    <span
      className={`${styles.btnTabs} ${mode === "register" ? styles.active : ""}`}
      onClick={() => setMode("register")}
    >
      Sign up
    </span>
  </div>
);
