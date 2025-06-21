import { apiClient } from '../../../apiClient';

export async function signUp(payload) {
  try {
    // Ждём, пока apiClient вернёт ответ (уже распарсенный)
    const data = await apiClient('/auth/signup', {
      method: 'POST',
      body: JSON.stringify(payload),
    });

    return data;
  } catch (error) {
    console.error('Error logging in user:', error);
    throw error;
  }
}

