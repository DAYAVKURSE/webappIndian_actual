import { apiClient } from '@/apiClient';

export async function getOutcome() {
  try {
    const response = await apiClient(`/games/binary/outcome`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });
    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Error registering user:', error);
    throw error;
  }
}
