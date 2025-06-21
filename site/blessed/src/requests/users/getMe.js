import useStore from '../../store';
import { apiClient } from '../../apiClient';

export async function getMe() {
  try {
    const data = await apiClient('/users', { method: 'GET' });

    const store = useStore.getState();
    store.setUserName(data.Nickname);
    store.setBalanceRupee(data.BalanceRupee);
    store.setBalanceBi(data.BalanceBi);
    store.setDailyClicks(data.DailyClicks);

    useStore.setState({ avatarId: data.AvatarID });

    return data;
  } catch (error) {
    console.error('Error fetching user data:', error);
    throw error;
  }
}
