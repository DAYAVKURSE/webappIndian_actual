import styles from "./GameCard.module.scss";
import { Link } from "react-router-dom";

export const GameCard = ({ src, label, to, desc, soon, color }) => {
	const Wrapper = soon ? 'div' : Link;

	return (
		<Wrapper
			to={!soon ? to : undefined}
			className={`${styles.gamecard}`}
		>
			<div className={styles.gamecard_img_wrapper}>
				<img className={styles.gamecard_img} src={src} alt="" style={{ filter: soon ? "grayscale(100%)" : "none" }}/>
				{soon && <img src="/coming_soon.png" alt="" style={{position: "absolute", zIndex: "2", width: "79px", height: "58px"}} />}
			</div>
			<h3 className={styles.gamecard_label}>
				<p style={{ color: color }}>{label}</p>
				{desc && <p className={styles.gamecard_desc}>{desc}</p>}
			</h3>
		</Wrapper>
	);
};
