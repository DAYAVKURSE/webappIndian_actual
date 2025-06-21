// src/shared/api/auth.js
import { apiClient } from '../../../apiClient';
import { setAccessToken } from '@/utils/token-storage';
import { removeAccessToken } from '../../../utils/token-storage';

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
  } catch (error) {
    console.error('Error logging in user:', error);
    throw error;
  }
}

export async function logout() {
  removeAccessToken();
  return apiClient('/users/logout', {
    method: 'POST',
  });
}
