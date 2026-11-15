const API_BASE = ''; // Uses Vite proxy (or http://localhost:8080)

export const api = {
  getToken() {
    return localStorage.getItem('token');
  },
  getUser() {
    const u = localStorage.getItem('user');
    return u ? JSON.parse(u) : null;
  },
  setAuth(token, user) {
    localStorage.setItem('token', token);
    localStorage.setItem('user', JSON.stringify(user));
  },
  clearAuth() {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
  },
  async request(endpoint, options = {}) {
    const token = this.getToken();
    const headers = {
      'Content-Type': 'application/json',
      ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
      ...(options.headers || {})
    };

    const res = await fetch(`${API_BASE}${endpoint}`, {
      ...options,
      headers
    });

    if (res.status === 204) return null;
    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.result?.error || data.error || 'Request failed');
    }
    return data.result !== undefined ? data.result : data;
  },

  // Auth
  register(body) {
    return this.request('/auth/register', { method: 'POST', body: JSON.stringify(body) });
  },
  login(body) {
    return this.request('/auth/login', { method: 'POST', body: JSON.stringify(body) });
  },

  // Products
  getProducts() {
    return this.request('/products');
  },
  getProduct(id) {
    return this.request(`/products/${id}`);
  },

  // Orders & Puzzle Checkout
  requestChallenge(items) {
    return this.request('/orders/checkout-challenge', { method: 'POST', body: JSON.stringify({ items }) });
  },
  confirmOrder(challengeId, answer, items) {
    return this.request('/orders/confirm', { method: 'POST', body: JSON.stringify({ challengeId, answer, items }) });
  },
  getMyOrders() {
    return this.request('/orders/my-orders');
  },

  // Admin / Store Owner
  getAdminProducts() {
    return this.request('/admin/products');
  },
  createProduct(body) {
    return this.request('/admin/products', { method: 'POST', body: JSON.stringify(body) });
  },
  updateProduct(id, body) {
    return this.request(`/admin/products/${id}`, { method: 'PUT', body: JSON.stringify(body) });
  },
  deleteProduct(id) {
    return this.request(`/admin/products/${id}`, { method: 'DELETE' });
  },
  getAdminOrders() {
    return this.request('/admin/orders');
  },
  updateOrderStatus(id, status) {
    return this.request(`/admin/orders/${id}/status`, { method: 'PUT', body: JSON.stringify({ status }) });
  }
};
