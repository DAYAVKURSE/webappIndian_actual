import { API_BASE_URL, WS_PROTOCOL, API_PROTOCOL } from '@/config';
import useStore from '@/store';

const initData = window.Telegram.WebApp.initData;

export async function getMe() {
  try {
    const response = await fetch(`${API_PROTOCOL}://${API_BASE_URL}/users`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'X-Telegram-Init-Data': initData,
      },
    });

    if (response.status === 401) {
      window.location.href = '/onboarding';
    }

    const data = await response.json();

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
