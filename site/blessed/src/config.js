const isDev = import.meta.env.DEV;

export const ACCESS_TOKEN_KEY = 'accessToken';
export const API_BASE_URL = isDev
  ? 'https://testfakeserver.com/api'   // dev-сервер
  : 'https://testfakeserver.com/api';  // продакшен