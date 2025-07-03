import {apiClient} from "@/apiClient";
import useStore from '@/store';

export async function getBPC() {
  try {
    const response = await apiClient(`/clicker`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
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
