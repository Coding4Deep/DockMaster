import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

import { AuthProvider } from './contexts/AuthContext';
import ProtectedRoute from './components/auth/ProtectedRoute';
import Layout from './components/common/Layout';
import DashboardPage from './pages/DashboardPage';
import ContainersPage from './pages/ContainersPage';
import ImagesPage from './pages/ImagesPage';
import VolumesPage from './pages/VolumesPage';
import NetworksPage from './pages/NetworksPage';
import SystemMetricsPage from './pages/SystemMetricsPage';
import { ThemeProvider } from './hooks/useTheme';

function App() {
  return (
    <AuthProvider>
      <ThemeProvider>
        <Router>
          <div className="min-h-screen bg-gray-50 dark:bg-gray-900 transition-colors duration-200">
            <ProtectedRoute>
              <Layout>
                <Routes>
                  <Route path="/" element={<DashboardPage />} />
                  <Route path="/containers" element={<ContainersPage />} />
                  <Route path="/images" element={<ImagesPage />} />
                  <Route path="/volumes" element={<VolumesPage />} />
                  <Route path="/networks" element={<NetworksPage />} />
                  <Route path="/metrics" element={<SystemMetricsPage />} />
                </Routes>
              </Layout>
            </ProtectedRoute>
            <ToastContainer
              position="top-right"
              autoClose={5000}
              hideProgressBar={false}
              newestOnTop={false}
              closeOnClick
              rtl={false}
              pauseOnFocusLoss
              draggable
              pauseOnHover
              theme="colored"
            />
          </div>
        </Router>
      </ThemeProvider>
    </AuthProvider>
  );
}

export default App;
