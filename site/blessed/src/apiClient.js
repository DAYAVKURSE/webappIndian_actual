import { API_BASE_URL } from '@/config';
import {
  getAccessToken,
  getRefreshToken,
  setAccessToken,
  setRefreshToken,
  removeAccessToken,
  removeRefreshToken,
} from '@/utils/token-storage';

async function refreshToken() {
  const refresh_token = getRefreshToken();
  if (!refresh_token) {
    throw new Error('No refresh token available');
  }

  const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token }),
  });

  if (!response.ok) {
    throw new Error('Refresh token expired or invalid');
  }

  const data = await response.json();
  if (data.access_token) {
    setAccessToken(data.access_token);
  }
  if (data.refresh_token) {
    setRefreshToken(data.refresh_token);
  }

  return data.access_token;
}

export async function apiClient(path, options = {}) {
  let token = getAccessToken();

  const headers = {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${token || ''}`,
    ...(options.headers || {}),
  };

  let response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers,
  });

  if (response.status === 401) {
    // Попытка обновить токен
    try {
      token = await refreshToken();

      // Повторяем исходный запрос с новым токеном
      response = await fetch(`${API_BASE_URL}${path}`, {
        ...options,
        headers: {
          ...headers,
          Authorization: `Bearer ${token}`,
        },
      });
    } catch (e) {
      // Если обновление не удалось, очищаем токены и редиректим
      removeAccessToken();
      removeRefreshToken();
      window.location.href = '/login';
      return;
    }
  }

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(errorText || 'API error');
  }

  return await response.json();
}
