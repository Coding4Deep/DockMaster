import React, { useState, useEffect } from 'react';
import { 
  PlayIcon, 
  StopIcon, 
  ArrowPathIcon, 
  TrashIcon,
  EyeIcon,
  DocumentTextIcon
} from '@heroicons/react/24/outline';
import LoadingSpinner from '../components/common/LoadingSpinner';
import Modal from '../components/common/Modal';
import api from '../services/api';

const ContainersPage = () => {
  const [containers, setContainers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showAll, setShowAll] = useState(true);
  const [selectedContainer, setSelectedContainer] = useState(null);
  const [showStatsModal, setShowStatsModal] = useState(false);
  const [showLogsModal, setShowLogsModal] = useState(false);
  const [containerStats, setContainerStats] = useState(null);
  const [containerLogs, setContainerLogs] = useState([]);

  useEffect(() => {
    fetchContainers();
  }, [showAll]);

  const fetchContainers = async () => {
    try {
      setLoading(true);
      const response = await api.getContainers(showAll);
      setContainers(response.data);
      setError(null);
    } catch (err) {
      console.error('Failed to fetch containers:', err);
      setError('Failed to load containers');
    } finally {
      setLoading(false);
    }
  };

  const handleContainerAction = async (action, containerId) => {
    try {
      switch (action) {
        case 'start':
          await api.startContainer(containerId);
          break;
        case 'stop':
          await api.stopContainer(containerId);
          break;
        case 'restart':
          await api.restartContainer(containerId);
          break;
        case 'delete':
          if (window.confirm('Are you sure you want to delete this container?')) {
            await api.deleteContainer(containerId, true);
          } else {
            return;
          }
          break;
        default:
          return;
      }
      await fetchContainers();
    } catch (err) {
      console.error(`Failed to ${action} container:`, err);
      alert(`Failed to ${action} container: ${err.response?.data?.message || err.message}`);
    }
  };

  const showContainerStats = async (container) => {
    try {
      setSelectedContainer(container);
      const response = await api.getContainerStats(container.Id);
      setContainerStats(response.data);
      setShowStatsModal(true);
    } catch (err) {
      console.error('Failed to fetch container stats:', err);
      alert('Failed to fetch container stats');
    }
  };

  const showContainerLogs = async (container) => {
    try {
      setSelectedContainer(container);
      const response = await api.getContainerLogs(container.Id);
      setContainerLogs(response.data);
      setShowLogsModal(true);
    } catch (err) {
      console.error('Failed to fetch container logs:', err);
      alert('Failed to fetch container logs');
    }
  };

  const getStatusColor = (state) => {
    switch (state.toLowerCase()) {
      case 'running':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
      case 'exited':
        return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200';
      case 'paused':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200';
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200';
    }
  };

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
            Containers
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            Manage your Docker containers
          </p>
        </div>
        <div className="flex items-center space-x-4">
          <label className="flex items-center">
            <input
              type="checkbox"
              checked={showAll}
              onChange={(e) => setShowAll(e.target.checked)}
              className="rounded border-gray-300 text-blue-600 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50"
            />
            <span className="ml-2 text-sm text-gray-700 dark:text-gray-300">
              Show all containers
            </span>
          </label>
          <button
            onClick={fetchContainers}
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

      {/* Containers Table */}
      <div className="bg-white dark:bg-gray-800 shadow overflow-hidden sm:rounded-md">
        <div className="px-4 py-5 sm:p-6">
          {containers.length === 0 ? (
            <div className="text-center py-12">
              <div className="text-gray-500 dark:text-gray-400">
                No containers found
              </div>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                <thead className="bg-gray-50 dark:bg-gray-700">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Name
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Image
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Status
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Ports
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                  {containers.map((container) => (
                    <tr key={container.Id}>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-gray-900 dark:text-white">
                          {container.Names?.[0]?.replace('/', '') || container.Id.substring(0, 12)}
                        </div>
                        <div className="text-sm text-gray-500 dark:text-gray-400">
                          {container.Id.substring(0, 12)}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm text-gray-900 dark:text-white">
                          {container.Image}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getStatusColor(container.State)}`}>
                          {container.State}
                        </span>
                        <div className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                          {container.Status}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                        {container.Ports?.map((port, index) => (
                          <div key={index}>
                            {port.PublicPort ? `${port.PublicPort}:${port.PrivatePort}` : port.PrivatePort}
                            /{port.Type}
                          </div>
                        )) || 'None'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                        <div className="flex space-x-2">
                          {container.State === 'running' ? (
                            <>
                              <button
                                onClick={() => handleContainerAction('stop', container.Id)}
                                className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300"
                                title="Stop"
                              >
                                <StopIcon className="h-4 w-4" />
                              </button>
                              <button
                                onClick={() => handleContainerAction('restart', container.Id)}
                                className="text-yellow-600 hover:text-yellow-900 dark:text-yellow-400 dark:hover:text-yellow-300"
                                title="Restart"
                              >
                                <ArrowPathIcon className="h-4 w-4" />
                              </button>
                              <button
                                onClick={() => showContainerStats(container)}
                                className="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300"
                                title="Stats"
                              >
                                <EyeIcon className="h-4 w-4" />
                              </button>
                              <button
                                onClick={() => showContainerLogs(container)}
                                className="text-purple-600 hover:text-purple-900 dark:text-purple-400 dark:hover:text-purple-300"
                                title="Logs"
                              >
                                <DocumentTextIcon className="h-4 w-4" />
                              </button>
                            </>
                          ) : (
                            <button
                              onClick={() => handleContainerAction('start', container.Id)}
                              className="text-green-600 hover:text-green-900 dark:text-green-400 dark:hover:text-green-300"
                              title="Start"
                            >
                              <PlayIcon className="h-4 w-4" />
                            </button>
                          )}
                          <button
                            onClick={() => handleContainerAction('delete', container.Id)}
                            className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300"
                            title="Delete"
                          >
                            <TrashIcon className="h-4 w-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {/* Stats Modal */}
      <Modal
        isOpen={showStatsModal}
        onClose={() => setShowStatsModal(false)}
        title={`Container Stats - ${selectedContainer?.Names?.[0]?.replace('/', '') || 'Unknown'}`}
      >
        {containerStats && (
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">CPU Usage</dt>
                <dd className="text-lg font-semibold text-gray-900 dark:text-white">
                  {containerStats.cpuPerc?.toFixed(2) || 0}%
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Memory Usage</dt>
                <dd className="text-lg font-semibold text-gray-900 dark:text-white">
                  {containerStats.memPerc?.toFixed(2) || 0}%
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Network RX</dt>
                <dd className="text-lg font-semibold text-gray-900 dark:text-white">
                  {Math.round((containerStats.netRx || 0) / 1024)} KB
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Network TX</dt>
                <dd className="text-lg font-semibold text-gray-900 dark:text-white">
                  {Math.round((containerStats.netTx || 0) / 1024)} KB
                </dd>
              </div>
            </div>
          </div>
        )}
      </Modal>

      {/* Logs Modal */}
      <Modal
        isOpen={showLogsModal}
        onClose={() => setShowLogsModal(false)}
        title={`Container Logs - ${selectedContainer?.Names?.[0]?.replace('/', '') || 'Unknown'}`}
      >
        <div className="bg-black text-green-400 p-4 rounded-md font-mono text-sm max-h-96 overflow-y-auto">
          {containerLogs.length > 0 ? (
            containerLogs.map((log, index) => (
              <div key={index} className="mb-1">
                <span className="text-gray-500">[{new Date(log.timestamp).toLocaleTimeString()}]</span> {log.log}
              </div>
            ))
          ) : (
            <div>No logs available</div>
          )}
        </div>
      </Modal>
    </div>
  );
};

export default ContainersPage;
