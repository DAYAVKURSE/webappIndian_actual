import { useNavigate } from 'react-router-dom';
import styles from './BackButton.module.scss';

export const BackButton = ({ className = '' }) => {
  const navigate = useNavigate();

  return (
    <button className={`${styles.backButton} ${className}`} onClick={() => navigate(-1)}>
      Back{' '}
      <svg width="24" height="24" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path
          d="M13.0942 10L8.08507 4.99167L6.90674 6.17L10.7401 10.0033L6.90674 13.8308L8.08507 15.0092L13.0942 10Z"
          fill="#D0D0D0"
        ></path>
      </svg>
    </button>
  );
};
