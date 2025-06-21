export function getAccessToken() {
  return localStorage.getItem('accessToken');
}

export function getRefreshToken() {
  return localStorage.getItem('refreshToken');
}

export function setAccessToken(token) {
  localStorage.setItem('accessToken', token);
}

export function setRefreshToken(token) {
  localStorage.setItem('refreshToken', token);
}

export function removeAccessToken() {
  localStorage.removeItem('accessToken');
  localStorage.removeItem('refreshToken');
}
