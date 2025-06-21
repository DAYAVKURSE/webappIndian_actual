import { useState } from 'react';
import { FaLock, FaUser, FaEye, FaEyeSlash } from 'react-icons/fa';
import styles from './Login.module.scss';

export const AuthForm = ({ mode, form, onChange, onSubmit, error, loading }) => {
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);

  return (
    <form onSubmit={onSubmit} className={styles.form}>
      <div className={styles.inputWrapper}>
        <FaUser className={styles.icon} />
        <input type="text" name="username" placeholder="Username" value={form.username} onChange={onChange} required />
      </div>

      <div className={styles.inputWrapper}>
        <FaLock className={styles.icon} />
        <input
          type={showPassword ? 'text' : 'password'}
          name="password"
          placeholder="Password"
          value={form.password}
          onChange={onChange}
          required
        />
        <span className={styles.eye} onClick={() => setShowPassword(!showPassword)}>
          {showPassword ? <FaEyeSlash /> : <FaEye />}
        </span>
      </div>

      {mode === 'register' && (
        <div className={styles.inputWrapper}>
          <FaLock className={styles.icon} />
          <input
            type={showConfirm ? 'text' : 'password'}
            name="confirmPassword"
            placeholder="Confirm Password"
            value={form.confirmPassword}
            onChange={onChange}
            required
          />
          <span className={styles.eye} onClick={() => setShowConfirm(!showConfirm)}>
            {showConfirm ? <FaEyeSlash /> : <FaEye />}
          </span>
        </div>
      )}

      {error && <p className={styles.error}>{error}</p>}

      <button type="submit" disabled={loading} className={styles.btn}>
        {loading ? 'Loading...' : 'Continue'}
        <svg width="10" height="16" viewBox="0 0 10 16" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path
            d="M1 1L8.67819 7.88228C8.84854 8.03497 8.8453 8.30279 8.67131 8.45132L1 15"
            stroke="white"
            strokeWidth="1.5122"
            strokeLinecap="round"
          />
        </svg>
      </button>
    </form>
  );
};
