import { API_BASE_URL } from '@/config';
import { getAccessToken, getRefreshToken, setAccessToken, removeAccessToken } from '@/utils/token-storage';

async function refreshAccessToken() {
  const refreshToken = getRefreshToken();
  if (!refreshToken) {
    throw new Error('No refresh token');
  }

  const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ refreshToken }),
  });

  if (!response.ok) {
    throw new Error('Failed to refresh token');
  }

  const data = await response.json();
  setAccessToken(data.accessToken);
  return data.accessToken;
}

export async function apiClient(path, options = {}, retry = true) {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${getAccessToken() || ''}`,
      ...(options.headers || {}),
    },
  });

  if (response.status === 401 && retry) {
    try {
      const newToken = await refreshAccessToken();
      return apiClient(path, options, false);
    } catch (e) {
      removeAccessToken();
      window.location.href = '/login';
      return;
    }
  }

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(errorText || 'API error');
  }

  return response;
}
