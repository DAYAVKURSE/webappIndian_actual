import { apiClient } from '@/apiClient';

export async function getLeaders(type) {
  try {
    const response = await apiClient(`/leaders/get?period=${type}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    const data = await response.json();

    return data;
  } catch (error) {
    console.error('Error registering user:', error);
  }
}
