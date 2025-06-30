import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

// Create axios instance
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
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
    return Promise.reject(error);
  }
);

// Response interceptor to handle auth errors
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Token expired or invalid
      localStorage.removeItem('token');
      window.location.reload();
    }
    return Promise.reject(error);
  }
);

const api = {
  // Set auth token
  setAuthToken: (token) => {
    if (token) {
      apiClient.defaults.headers.common['Authorization'] = `Bearer ${token}`;
    } else {
      delete apiClient.defaults.headers.common['Authorization'];
    }
  },

  // Authentication
  login: (username, password) => 
    apiClient.post('/auth/login', { username, password }),
  
  logout: () => 
    apiClient.post('/auth/logout'),
  
  getCurrentUser: () => 
    apiClient.get('/auth/me'),

  // System info
  getSystemInfo: () => 
    apiClient.get('/system/info'),

  // Containers
  getContainers: (all = false) => 
    apiClient.get(`/containers?all=${all}`),
  
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
