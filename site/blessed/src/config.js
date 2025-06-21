const isDev = import.meta.env.DEV;

export const ACCESS_TOKEN_KEY = 'accessToken';
export const REFRESH_TOKEN_KEY = 'refreshToken';

export const API_BASE_URL = isDev
  ? 'https://testfakeserver.com/api'
  : 'https://testfakeserver.com/api';
