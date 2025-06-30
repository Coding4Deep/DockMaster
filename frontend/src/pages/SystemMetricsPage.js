import React, { useState, useEffect, useRef } from 'react';
import { 
  CpuChipIcon,
  CircleStackIcon,
  ServerIcon,
  WifiIcon,
  ClockIcon
} from '@heroicons/react/24/outline';
import LoadingSpinner from '../components/common/LoadingSpinner';
import api from '../services/api';

const SystemMetricsPage = () => {
  const [metrics, setMetrics] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const intervalRef = useRef(null);

  useEffect(() => {
    fetchMetrics();
    
    if (autoRefresh) {
      intervalRef.current = setInterval(fetchMetrics, 2000); // Refresh every 2 seconds
    }

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [autoRefresh]);

  const fetchMetrics = async () => {
    try {
      const response = await api.getSystemMetrics();
      setMetrics(response.data);
      setError(null);
      if (loading) setLoading(false);
    } catch (err) {
      console.error('Failed to fetch system metrics:', err);
      setError('Failed to load system metrics');
      setLoading(false);
    }
  };

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatUptime = (seconds) => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    
    if (days > 0) {
      return `${days}d ${hours}h ${minutes}m`;
    } else if (hours > 0) {
      return `${hours}h ${minutes}m`;
    } else {
      return `${minutes}m`;
    }
  };

  const getProgressColor = (percentage) => {
    if (percentage < 50) return 'bg-green-500';
    if (percentage < 80) return 'bg-yellow-500';
    return 'bg-red-500';
  };

  const ProgressBar = ({ percentage, label, color }) => (
    <div className="w-full">
      <div className="flex justify-between text-sm text-gray-600 dark:text-gray-400 mb-1">
        <span>{label}</span>
        <span>{percentage.toFixed(1)}%</span>
      </div>
      <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
        <div
          className={`h-2 rounded-full transition-all duration-300 ${color || getProgressColor(percentage)}`}
          style={{ width: `${Math.min(percentage, 100)}%` }}
        ></div>
      </div>
    </div>
  );

  const MetricCard = ({ title, icon: Icon, children, className = "" }) => (
    <div className={`bg-white dark:bg-gray-800 rounded-lg shadow p-6 ${className}`}>
      <div className="flex items-center mb-4">
        <Icon className="h-6 w-6 text-blue-600 dark:text-blue-400 mr-2" />
        <h3 className="text-lg font-medium text-gray-900 dark:text-white">{title}</h3>
      </div>
      {children}
    </div>
  );

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <LoadingSpinner />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
            System Metrics
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            Real-time system performance monitoring
          </p>
        </div>
        <div className="flex items-center space-x-4">
          <label className="flex items-center">
            <input
              type="checkbox"
              checked={autoRefresh}
              onChange={(e) => setAutoRefresh(e.target.checked)}
              className="rounded border-gray-300 text-blue-600 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50"
            />
            <span className="ml-2 text-sm text-gray-700 dark:text-gray-300">
              Auto-refresh
            </span>
          </label>
          <button
            onClick={fetchMetrics}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            Refresh
          </button>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 dark:bg-red-900 border border-red-200 dark:border-red-700 rounded-md p-4">
          <div className="text-red-800 dark:text-red-200">{error}</div>
        </div>
      )}

      {metrics && (
        <>
          {/* Overview Cards */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
              <div className="flex items-center">
                <CpuChipIcon className="h-8 w-8 text-blue-600 dark:text-blue-400" />
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">CPU Usage</p>
                  <p className="text-2xl font-bold text-gray-900 dark:text-white">
                    {metrics.cpu.usage.toFixed(1)}%
                  </p>
                </div>
              </div>
            </div>

            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
              <div className="flex items-center">
                <CircleStackIcon className="h-8 w-8 text-green-600 dark:text-green-400" />
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Memory Usage</p>
                  <p className="text-2xl font-bold text-gray-900 dark:text-white">
                    {metrics.memory.usage.toFixed(1)}%
                  </p>
                </div>
              </div>
            </div>

            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
              <div className="flex items-center">
                <ServerIcon className="h-8 w-8 text-yellow-600 dark:text-yellow-400" />
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Disk Usage</p>
                  <p className="text-2xl font-bold text-gray-900 dark:text-white">
                    {metrics.disk.usage.toFixed(1)}%
                  </p>
                </div>
              </div>
            </div>

            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
              <div className="flex items-center">
                <ClockIcon className="h-8 w-8 text-purple-600 dark:text-purple-400" />
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Uptime</p>
                  <p className="text-2xl font-bold text-gray-900 dark:text-white">
                    {formatUptime(metrics.uptime)}
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Detailed Metrics */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* CPU Details */}
            <MetricCard title="CPU Details" icon={CpuChipIcon}>
              <div className="space-y-4">
                <ProgressBar 
                  percentage={metrics.cpu.usage} 
                  label="Overall Usage"
                />
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-gray-600 dark:text-gray-400">Cores:</span>
                    <span className="ml-2 font-medium text-gray-900 dark:text-white">
                      {metrics.cpu.cores}
                    </span>
                  </div>
                  <div>
                    <span className="text-gray-600 dark:text-gray-400">User Time:</span>
                    <span className="ml-2 font-medium text-gray-900 dark:text-white">
                      {metrics.cpu.user_time.toFixed(1)}%
                    </span>
                  </div>
                  <div>
                    <span className="text-gray-600 dark:text-gray-400">System Time:</span>
                    <span className="ml-2 font-medium text-gray-900 dark:text-white">
                      {metrics.cpu.system_time.toFixed(1)}%
                    </span>
                  </div>
                  <div>
                    <span className="text-gray-600 dark:text-gray-400">Idle Time:</span>
                    <span className="ml-2 font-medium text-gray-900 dark:text-white">
                      {metrics.cpu.idle_time.toFixed(1)}%
                    </span>
                  </div>
                </div>
              </div>
            </MetricCard>

            {/* Memory Details */}
            <MetricCard title="Memory Details" icon={CircleStackIcon}>
              <div className="space-y-4">
                <ProgressBar 
                  percentage={metrics.memory.usage} 
                  label="Memory Usage"
                />
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-gray-600 dark:text-gray-400">Total:</span>
                    <span className="ml-2 font-medium text-gray-900 dark:text-white">
                      {formatBytes(metrics.memory.total)}
                    </span>
                  </div>
                  <div>
                    <span className="text-gray-600 dark:text-gray-400">Used:</span>
                    <span className="ml-2 font-medium text-gray-900 dark:text-white">
                      {formatBytes(metrics.memory.used)}
                    </span>
                  </div>
                  <div>
                    <span className="text-gray-600 dark:text-gray-400">Free:</span>
                    <span className="ml-2 font-medium text-gray-900 dark:text-white">
                      {formatBytes(metrics.memory.free)}
                    </span>
                  </div>
                  <div>
                    <span className="text-gray-600 dark:text-gray-400">Available:</span>
                    <span className="ml-2 font-medium text-gray-900 dark:text-white">
                      {formatBytes(metrics.memory.available)}
                    </span>
                  </div>
                  <div>
                    <span className="text-gray-600 dark:text-gray-400">Buffers:</span>
                    <span className="ml-2 font-medium text-gray-900 dark:text-white">
                      {formatBytes(metrics.memory.buffers)}
                    </span>
                  </div>
                  <div>
                    <span className="text-gray-600 dark:text-gray-400">Cached:</span>
                    <span className="ml-2 font-medium text-gray-900 dark:text-white">
                      {formatBytes(metrics.memory.cached)}
                    </span>
                  </div>
                </div>
              </div>
            </MetricCard>

            {/* Disk Details */}
            <MetricCard title="Disk Details" icon={ServerIcon}>
              <div className="space-y-4">
                <ProgressBar 
                  percentage={metrics.disk.usage} 
                  label="Disk Usage"
                />
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-gray-600 dark:text-gray-400">Total:</span>
                    <span className="ml-2 font-medium text-gray-900 dark:text-white">
                      {formatBytes(metrics.disk.total)}
                    </span>
                  </div>
                  <div>
                    <span className="text-gray-600 dark:text-gray-400">Used:</span>
                    <span className="ml-2 font-medium text-gray-900 dark:text-white">
                      {formatBytes(metrics.disk.used)}
                    </span>
                  </div>
                  <div>
                    <span className="text-gray-600 dark:text-gray-400">Free:</span>
                    <span className="ml-2 font-medium text-gray-900 dark:text-white">
                      {formatBytes(metrics.disk.free)}
                    </span>
                  </div>
                  <div>
                    <span className="text-gray-600 dark:text-gray-400">Read Ops:</span>
                    <span className="ml-2 font-medium text-gray-900 dark:text-white">
                      {metrics.disk.read_ops.toLocaleString()}
                    </span>
                  </div>
                  <div>
                    <span className="text-gray-600 dark:text-gray-400">Write Ops:</span>
                    <span className="ml-2 font-medium text-gray-900 dark:text-white">
                      {metrics.disk.write_ops.toLocaleString()}
                    </span>
                  </div>
                  <div>
                    <span className="text-gray-600 dark:text-gray-400">I/O:</span>
                    <span className="ml-2 font-medium text-gray-900 dark:text-white">
                      {formatBytes(metrics.disk.read_bytes + metrics.disk.write_bytes)}
                    </span>
                  </div>
                </div>
              </div>
            </MetricCard>

            {/* Network & Load Details */}
            <MetricCard title="Network & Load" icon={WifiIcon}>
              <div className="space-y-4">
                <div>
                  <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Network Traffic</h4>
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <span className="text-gray-600 dark:text-gray-400">Received:</span>
                      <span className="ml-2 font-medium text-gray-900 dark:text-white">
                        {formatBytes(metrics.network.bytes_received)}
                      </span>
                    </div>
                    <div>
                      <span className="text-gray-600 dark:text-gray-400">Sent:</span>
                      <span className="ml-2 font-medium text-gray-900 dark:text-white">
                        {formatBytes(metrics.network.bytes_sent)}
                      </span>
                    </div>
                    <div>
                      <span className="text-gray-600 dark:text-gray-400">RX Packets:</span>
                      <span className="ml-2 font-medium text-gray-900 dark:text-white">
                        {metrics.network.packets_received.toLocaleString()}
                      </span>
                    </div>
                    <div>
                      <span className="text-gray-600 dark:text-gray-400">TX Packets:</span>
                      <span className="ml-2 font-medium text-gray-900 dark:text-white">
                        {metrics.network.packets_sent.toLocaleString()}
                      </span>
                    </div>
                  </div>
                </div>
                
                <div>
                  <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Load Average</h4>
                  <div className="grid grid-cols-3 gap-4 text-sm">
                    <div>
                      <span className="text-gray-600 dark:text-gray-400">1 min:</span>
                      <span className="ml-2 font-medium text-gray-900 dark:text-white">
                        {metrics.load.load1.toFixed(2)}
                      </span>
                    </div>
                    <div>
                      <span className="text-gray-600 dark:text-gray-400">5 min:</span>
                      <span className="ml-2 font-medium text-gray-900 dark:text-white">
                        {metrics.load.load5.toFixed(2)}
                      </span>
                    </div>
                    <div>
                      <span className="text-gray-600 dark:text-gray-400">15 min:</span>
                      <span className="ml-2 font-medium text-gray-900 dark:text-white">
                        {metrics.load.load15.toFixed(2)}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </MetricCard>
          </div>
        </>
      )}
    </div>
  );
};

export default SystemMetricsPage;
