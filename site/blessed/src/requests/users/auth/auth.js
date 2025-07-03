import { apiClient } from '../../../apiClient';
import {
  setAccessToken,
  setRefreshToken,
  removeAccessToken,
  removeRefreshToken,
} from '@/utils/token-storage';

export async function login(payload) {
  try {
    const response = await apiClient('/auth/login', {
      method: 'POST',
      body: JSON.stringify(payload),
    });

    const data = await response.json();

    if (data?.access_token) {
      setAccessToken(data.access_token);
    }
    if (data?.refresh_token) {
      setRefreshToken(data.refresh_token);
    }

    return data;
  } catch (error) {
    console.error('Error logging in user:', error);
    throw error;
  }
}


export async function logout() {
  removeAccessToken();
  removeRefreshToken();
  return apiClient('/users/logout', {
    method: 'POST',
  });
}
