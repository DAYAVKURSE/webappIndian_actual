import { API_BASE_URL, WS_PROTOCOL, API_PROTOCOL } from '@/config';
import useStore from '@/store';

const initData = window.Telegram.WebApp.initData;

export async function signUp({ Nickname, avatarId }) {
  const referredBy = useStore.getState().referredBy;
  
  const bodyData = {
    Nickname,
    avatarId,
  };

  try {
    const response = await fetch(`${API_PROTOCOL}://${API_BASE_URL}/users/auth/signup${referredBy ? `?referral=${referredBy}` : ''}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Telegram-Init-Data': initData,
      },
      body: JSON.stringify(bodyData),
    });

    if (response.status === 200) {
      return {
        status: response.status,
      };
    } else {
      return {
        status: response.status,
        data: await response.json(),
      };
    }
  } catch (error) {
    console.error('Error registering user:', error);
    throw error;
  }
}
