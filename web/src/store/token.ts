const STORAGE_KEYS = {
  name: "Name",
  accessToken: "AccessToken",
  refreshToken: "RefreshToken",
} as const;

class TokenStore {
  /**
   * Save token to localStorage
   */
  setToken(token: TokenResult) {
    if (token) {
      localStorage.setItem(STORAGE_KEYS.name, token.name);
      localStorage.setItem(STORAGE_KEYS.accessToken, token.accessToken);
      localStorage.setItem(STORAGE_KEYS.refreshToken, token.refreshToken);
    }
  }

  /**
   * Clear token and reload page
   */
  removeToken() {
    localStorage.removeItem(STORAGE_KEYS.accessToken);
    localStorage.removeItem(STORAGE_KEYS.refreshToken);
    location.reload();
  }

  /**
   * Get username
   */
  getName() {
    return localStorage.getItem(STORAGE_KEYS.name);
  }

  /**
   * Get access token
   */
  getAccessToken() {
    return localStorage.getItem(STORAGE_KEYS.accessToken);
  }

  /**
   * Get refresh token
   */
  getRefreshToken() {
    return localStorage.getItem(STORAGE_KEYS.refreshToken);
  }
}

export default new TokenStore();
