import { apiClient } from '@/apiClient';
export async function getDeposits() {
  try {
    const response = await apiClient(`/users/deposits`, {
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
