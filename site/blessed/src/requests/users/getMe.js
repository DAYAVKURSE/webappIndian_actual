import useStore from '../../store';
import { apiClient } from '../../apiClient';

export async function getMe() {
  try {
    const response = await apiClient('/users', {
      method: 'GET',
    });

    const data = await response.json();

    const store = useStore.getState();
    store.setUserName(data.Nickname);
    store.setBalanceRupee(data.BalanceRupee);
    store.setBalanceBi(data.BalanceBi);
    store.setDailyClicks(data.DailyClicks);

    useStore.setState({ avatarId: data.AvatarID });

    return data;
  } catch (error) {
    if (error?.response?.status === 400) {
      window.location.href = '/login';
    } else {
      console.error('Error fetching user data:', error);
    }
    throw error;
  }
}
