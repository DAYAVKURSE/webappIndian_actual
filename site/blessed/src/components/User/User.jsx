import styles from "./User.module.scss"
import useStore from '@/store';

export const User = () => {
    const { userName, avatarId } = useStore();
    return (
        <div className={styles.user}>
            <img className={styles.user__avatar} src={"/avatars/" + avatarId + ".png"} alt="User Avatar" />
            <p className={styles.user__name}>{userName}</p>
        </div>
    )
}
