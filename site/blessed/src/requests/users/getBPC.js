import { API_BASE_URL } from '@/config';
import useStore from '@/store';
const initData = window.Telegram.WebApp.initData;
export async function getBPC() {
  try {
    const response = await fetch(`https://${API_BASE_URL}/clicker`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'X-Telegram-Init-Data': initData,
      },
    });

    const data = await response.json();
    const BPC = data.BiPerClick * data.BonusMultiplier;
    useStore.setState({ BiPerClick: BPC });
    return data;
    
  } catch (error) {
    console.error('Error registering user:', error);
    throw error;
  }
}
