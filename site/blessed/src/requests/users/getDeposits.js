import { API_BASE_URL, WS_PROTOCOL, API_PROTOCOL } from '@/config';
const initData = window.Telegram.WebApp.initData;
export async function getDeposits() {
  try {
    const response = await fetch(`${API_PROTOCOL}://${API_BASE_URL}/users/deposits`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'X-Telegram-Init-Data': initData,
      },
    });

    const data = await response.json();
    return data;
    
  } catch (error) {
    console.error('Error registering user:', error);
    throw error;
  }
}
