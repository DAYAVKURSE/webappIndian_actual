import styles from './GameCard.module.scss';
import { Link } from 'react-router-dom';

export const GameCard = ({ src, label, to, desc, soon, color }) => {
  const Wrapper = soon ? 'div' : Link;

  return (
    <Wrapper
      to={!soon ? to : undefined}
      className={`${styles.gamecard}`}
      style={{
        borderColor: soon ? 'rgba(255, 255, 255, 0.15)' : color,
      }}
    >
      <div className={styles.gamecard_img_wrapper}>
        <img src={src} alt="" style={{ filter: soon ? 'grayscale(100%)' : 'none' }} />
      </div>
      <h3 className={styles.gamecard_label}>
        <p className={styles.gamecard_title}>{label}</p>
        <p className={styles.gamecard_play}>
          Play{' '}
          <svg width="20" height="20" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path
              d="M13.0942 10L8.08507 4.99167L6.90674 6.17L10.7401 10.0033L6.90674 13.8308L8.08507 15.0092L13.0942 10Z"
              fill="#D0D0D0"
            />
          </svg>
        </p>
        {desc && <p className={styles.gamecard_desc}>{desc}</p>}
      </h3>
    </Wrapper>
  );
};
