import { useState } from "react";
import styles from "./Login.module.scss";
import { Tabs } from "./Tabs";
import { AuthForm } from "./AuthForm";
import { useAuthForm } from "./useAuthForm";
import logo from "../../../public/logoLogin.png";
export const Login = () => {
  const [mode, setMode] = useState("login");
  const { form, handleChange, handleSubmit, error, loading } =
    useAuthForm(mode);

  return (
    <div className={styles.authContainer}>
      <div className={styles.headerForm}>
        <img src={logo} alt="" />
      </div>
      <h3 className={styles.authTitle}>
        {mode === "login" ? "Authorization" : "Registration"}
      </h3>
      <AuthForm
        mode={mode}
        form={form}
        onChange={handleChange}
        onSubmit={handleSubmit}
        error={error}
        loading={loading}
      />
      <Tabs mode={mode} setMode={setMode} />
    </div>
  );
};
