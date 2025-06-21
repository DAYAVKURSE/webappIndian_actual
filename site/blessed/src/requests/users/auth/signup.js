import { apiClient } from '../../../apiClient';

export async function signUp(payload) {
  try {
    const response = apiClient('/auth/signup', {
      method: 'POST',
      body: JSON.stringify(payload),
    });

    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Error logging in user:', error);
    throw error;
  }
}
