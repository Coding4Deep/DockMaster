import React, { createContext, useContext, useState, useEffect } from 'react';
import api from '../services/api';

const AuthContext = createContext();

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [token, setToken] = useState(localStorage.getItem('token'));

  useEffect(() => {
    if (token) {
      api.setAuthToken(token);
      fetchUser();
    } else {
      setLoading(false);
    }
  }, [token]);

  const fetchUser = async () => {
    console.log('AuthContext: fetchUser called');
    try {
      console.log('AuthContext: Making API call to getCurrentUser...');
      const response = await api.getCurrentUser();
      console.log('AuthContext: getCurrentUser response:', response.data);
      setUser(response.data);
    } catch (error) {
      console.error('AuthContext: Failed to fetch user:', error);
      console.error('AuthContext: fetchUser error details:', {
        message: error.message,
        response: error.response?.data,
        status: error.response?.status
      });
      logout();
    } finally {
      setLoading(false);
    }
  };

  const login = async (username, password) => {
    console.log('AuthContext: login called with username:', username);
    try {
      console.log('AuthContext: Making API call to login...');
      const response = await api.login(username, password);
      console.log('AuthContext: login response:', response.data);
      const { token: newToken, user: userData } = response.data;
      
      localStorage.setItem('token', newToken);
      setToken(newToken);
      setUser(userData);
      api.setAuthToken(newToken);
      
      console.log('AuthContext: login successful');
      return { success: true };
    } catch (error) {
      console.error('AuthContext: Login failed:', error);
      console.error('AuthContext: Login error details:', {
        message: error.message,
        response: error.response?.data,
        status: error.response?.status,
        config: error.config
      });
      return { 
        success: false, 
        error: error.response?.data?.message || error.message || 'Login failed' 
      };
    }
  };

  const logout = async () => {
    try {
      if (token) {
        await api.logout();
      }
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      localStorage.removeItem('token');
      setToken(null);
      setUser(null);
      api.setAuthToken(null);
    }
  };

  const changePassword = async (currentPassword, newPassword) => {
    try {
      await api.changePassword(currentPassword, newPassword);
      return { success: true };
    } catch (error) {
      console.error('Change password failed:', error);
      return { 
        success: false, 
        error: error.response?.data?.message || 'Failed to change password' 
      };
    }
  };

  const value = {
    user,
    login,
    logout,
    changePassword,
    loading,
    isAuthenticated: !!user,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};
