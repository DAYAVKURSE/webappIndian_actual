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
    const errorText = await response.text();
    console.error('Failed to refresh token:', errorText);
    throw new Error('Refresh token expired or invalid');
  }

  const data = await response.json();
  console.log('Refresh token response:', data);

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

  async function makeRequest(tokenToUse) {
    const headers = {
      'Content-Type': 'application/json',
      ...(options.headers || {}),
      Authorization: `Bearer ${tokenToUse}`,
    };

    const response = await fetch(`${API_BASE_URL}${path}`, {
      ...options,
      headers,
    });

    return response;
  }

  let response = await makeRequest(token || '');

  if (response.status === 401) {
    try {
      console.log('Token refresh triggered');

      const newAccessToken = await refreshToken();

      response = await makeRequest(newAccessToken);
    } catch (error) {
      console.error('Token refresh failed:', error);
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

  return response;
}
