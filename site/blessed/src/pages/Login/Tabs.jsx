import styles from "./Login.module.scss";

export const Tabs = ({ mode, setMode }) => (
  <div className={styles.tabs} data-mode={mode}>
    <button
      className={`${styles.btn} ${mode === "login" ? styles.active : ""}`}
      onClick={() => setMode("login")}
    >
      Login
    </button>
    <button
      className={`${styles.btn} ${mode === "register" ? styles.active : ""}`}
      onClick={() => setMode("register")}
    >
      Sign up
    </button>
  </div>
);
