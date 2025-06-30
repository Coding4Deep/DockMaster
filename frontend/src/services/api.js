import axios from 'axios';

// Get API URL from environment variable or construct from current location
const getApiBaseUrl = () => {
  // If running in development mode, use environment variable or localhost
  if (process.env.NODE_ENV === 'development') {
    return process.env.REACT_APP_API_URL || 'http://localhost:8081';
  }
  
  // In production/container, check if we have an explicit API URL
  if (process.env.REACT_APP_API_URL) {
    return process.env.REACT_APP_API_URL;
  }
  
  // For container networking, try to determine the correct URL
  const { protocol, hostname, port } = window.location;
  
  // If we're running on a custom port, assume backend is on a different port
  if (port && port !== '80' && port !== '443') {
    // Try common backend ports
    const frontendPort = parseInt(port);
    let backendPort = 8081; // default
    
    // Map common frontend ports to backend ports
    if (frontendPort === 3000) backendPort = 8081;
    else if (frontendPort === 4000) backendPort = 9090;
    else backendPort = frontendPort + 1000; // fallback logic
    
    return `${protocol}//${hostname}:${backendPort}`;
  }
  
  // Default fallback
  return `${protocol}//${hostname}:8081`;
};

const API_BASE_URL = getApiBaseUrl();

console.log('API Base URL:', API_BASE_URL); // Debug log

// Create axios instance
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    console.error('Request interceptor error:', error);
    return Promise.reject(error);
  }
);

// Response interceptor to handle auth errors
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error('API Error:', error.response?.data || error.message);
    if (error.response?.status === 401) {
      // Token expired or invalid
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

const api = {
  // Set auth token
  setAuthToken: (token) => {
    if (token) {
      apiClient.defaults.headers.common['Authorization'] = `Bearer ${token}`;
      localStorage.setItem('token', token);
    } else {
      delete apiClient.defaults.headers.common['Authorization'];
      localStorage.removeItem('token');
    }
  },

  // Authentication
  login: (username, password) => 
    apiClient.post('/auth/login', { username, password }),
  
  logout: () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    return Promise.resolve();
  },
  
  getCurrentUser: () => 
    apiClient.get('/auth/me'),

  changePassword: (currentPassword, newPassword) =>
    apiClient.post('/auth/change-password', { 
      current_password: currentPassword, 
      new_password: newPassword 
    }),

  // System info
  getSystemInfo: () => 
    apiClient.get('/system/info'),

  getSystemMetrics: () =>
    apiClient.get('/system/metrics'),

  // Containers
  getContainers: (all = true) => 
    apiClient.get(`/containers?all=${all}`),
  
  runContainer: (config) =>
    apiClient.post('/containers/run', config),
  
  startContainer: (id) => 
    apiClient.post(`/containers/${id}/start`),
  
  stopContainer: (id) => 
    apiClient.post(`/containers/${id}/stop`),
  
  restartContainer: (id) => 
    apiClient.post(`/containers/${id}/restart`),
  
  deleteContainer: (id, force = false) => 
    apiClient.delete(`/containers/${id}?force=${force}`),
  
  getContainerStats: (id) => 
    apiClient.get(`/containers/${id}/stats`),
  
  getContainerLogs: (id, tail = 100) => 
    apiClient.get(`/containers/${id}/logs?tail=${tail}`),

  // Images
  getImages: () => 
    apiClient.get('/images'),
  
  searchImages: (query) =>
    apiClient.get(`/images/search?q=${encodeURIComponent(query)}`),

  pullImage: (image) =>
    apiClient.post('/images/pull', { image }),

  inspectImage: (id) =>
    apiClient.get(`/images/${id}/inspect`),
  
  deleteImage: (id, force = false) => 
    apiClient.delete(`/images/${id}?force=${force}`),

  // Volumes
  getVolumes: () => 
    apiClient.get('/volumes'),
  
  deleteVolume: (name, force = false) => 
    apiClient.delete(`/volumes/${name}?force=${force}`),

  // Networks
  getNetworks: () => 
    apiClient.get('/networks'),
  
  deleteNetwork: (id) => 
    apiClient.delete(`/networks/${id}`),

  // Health check
  healthCheck: () => 
    apiClient.get('/health'),
};

export default api;
