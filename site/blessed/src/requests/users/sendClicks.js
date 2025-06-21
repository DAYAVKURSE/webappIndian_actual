import useStore from '@/store';
import { apiClient } from '@/apiClient';

export async function sendClicks(Clicks, BPC) {
  try {
    const response = await apiClient(`/clicker`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        ClicksCount: Clicks,
        BiPerClick: BPC,
      }),
    });

    const contentLength = response.headers.get('Content-Length');
    if (contentLength && parseInt(contentLength) > 0) {
      const data = await response.json();
      if (data && data.BiPerClick) {
        useStore.setState({ BiPerClick: data.BiPerClick * data.BonusMultiplier });
      }
      return data;
    } else {
      return null;
    }
  } catch (error) {
    console.error('Error registering user:', error);
    throw error;
  }
}
