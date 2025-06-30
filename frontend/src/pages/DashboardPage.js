import React, { useState, useEffect } from 'react';
import { 
  CubeIcon, 
  PhotoIcon, 
  CircleStackIcon, 
  GlobeAltIcon,
  ServerIcon,
  CpuChipIcon,
  ComputerDesktopIcon
} from '@heroicons/react/24/outline';
import LoadingSpinner from '../components/common/LoadingSpinner';
import api from '../services/api';

const DashboardPage = () => {
  const [systemInfo, setSystemInfo] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    fetchSystemInfo();
  }, []);

  const fetchSystemInfo = async () => {
    try {
      setLoading(true);
      const response = await api.getSystemInfo();
      setSystemInfo(response.data);
      setError(null);
    } catch (err) {
      console.error('Failed to fetch system info:', err);
      setError('Failed to load system information');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <LoadingSpinner />
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-12">
        <div className="text-red-600 dark:text-red-400 mb-4">{error}</div>
        <button
          onClick={fetchSystemInfo}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          Retry
        </button>
      </div>
    );
  }

  const stats = [
    {
      name: 'Total Containers',
      value: systemInfo?.Containers || 0,
      icon: CubeIcon,
      color: 'bg-blue-500',
      details: `${systemInfo?.ContainersRunning || 0} running, ${systemInfo?.ContainersStopped || 0} stopped`
    },
    {
      name: 'Images',
      value: systemInfo?.Images || 0,
      icon: PhotoIcon,
      color: 'bg-green-500',
      details: 'Docker images'
    },
    {
      name: 'CPU Cores',
      value: systemInfo?.NCPU || 0,
      icon: CpuChipIcon,
      color: 'bg-purple-500',
      details: systemInfo?.Architecture || 'Unknown'
    },
    {
      name: 'Memory',
      value: systemInfo?.MemTotal ? `${Math.round(systemInfo.MemTotal / 1024 / 1024 / 1024)}GB` : 'N/A',
      icon: ComputerDesktopIcon,
      color: 'bg-yellow-500',
      details: 'Total system memory'
    }
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
            Dashboard
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            Docker system overview and statistics
          </p>
        </div>
        <button
          onClick={fetchSystemInfo}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          Refresh
        </button>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {stats.map((stat) => (
          <div
            key={stat.name}
            className="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg"
          >
            <div className="p-5">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <div className={`${stat.color} p-3 rounded-md`}>
                    <stat.icon className="h-6 w-6 text-white" />
                  </div>
                </div>
                <div className="ml-5 w-0 flex-1">
                  <dl>
                    <dt className="text-sm font-medium text-gray-500 dark:text-gray-400 truncate">
                      {stat.name}
                    </dt>
                    <dd className="text-lg font-medium text-gray-900 dark:text-white">
                      {stat.value}
                    </dd>
                    <dd className="text-sm text-gray-500 dark:text-gray-400">
                      {stat.details}
                    </dd>
                  </dl>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* System Information */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Docker Info */}
        <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg leading-6 font-medium text-gray-900 dark:text-white mb-4">
              Docker Information
            </h3>
            <div className="space-y-3">
              <div className="flex justify-between">
                <span className="text-sm text-gray-500 dark:text-gray-400">Version:</span>
                <span className="text-sm font-medium text-gray-900 dark:text-white">
                  {systemInfo?.ServerVersion || 'N/A'}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm text-gray-500 dark:text-gray-400">Architecture:</span>
                <span className="text-sm font-medium text-gray-900 dark:text-white">
                  {systemInfo?.Architecture || 'N/A'}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm text-gray-500 dark:text-gray-400">OS Type:</span>
                <span className="text-sm font-medium text-gray-900 dark:text-white">
                  {systemInfo?.OSType || 'N/A'}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Container Status */}
        <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg leading-6 font-medium text-gray-900 dark:text-white mb-4">
              Container Status
            </h3>
            <div className="space-y-3">
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-500 dark:text-gray-400">Running:</span>
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
                  {systemInfo?.ContainersRunning || 0}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-500 dark:text-gray-400">Paused:</span>
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200">
                  {systemInfo?.ContainersPaused || 0}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-500 dark:text-gray-400">Stopped:</span>
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200">
                  {systemInfo?.ContainersStopped || 0}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-500 dark:text-gray-400">Total:</span>
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
                  {systemInfo?.Containers || 0}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* System Details */}
      <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <h3 className="text-lg leading-6 font-medium text-gray-900 dark:text-white mb-4">
            System Details
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div>
              <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Operating System</dt>
              <dd className="mt-1 text-sm text-gray-900 dark:text-white">
                {systemInfo?.OperatingSystem || 'Unknown'}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Architecture</dt>
              <dd className="mt-1 text-sm text-gray-900 dark:text-white">
                {systemInfo?.Architecture || 'Unknown'}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">CPU Cores</dt>
              <dd className="mt-1 text-sm text-gray-900 dark:text-white">
                {systemInfo?.NCPU || 'Unknown'}
              </dd>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default DashboardPage;
