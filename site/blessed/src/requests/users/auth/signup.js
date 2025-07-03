import { apiClient } from '../../../apiClient';

export async function signUp(payload) {
  try {
    // Ждём, пока apiClient вернёт ответ (уже распарсенный)
    const response = await apiClient('/auth/signup', {
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

