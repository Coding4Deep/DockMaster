import React, { useState, useEffect } from 'react';
import api from '../../services/api';

const CreateContainerModal = ({ isOpen, onClose, onSuccess }) => {
  const [formData, setFormData] = useState({
    name: '',
    image: '',
    ports: [{ host_port: '', container_port: '', protocol: 'tcp' }],
    environment: [{ key: '', value: '' }],
    volumes: [{ host_path: '', container_path: '', read_only: false }],
    command: [],
    working_dir: '',
    restart_policy: 'no'
  });
  const [images, setImages] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (isOpen) {
      fetchImages();
    }
  }, [isOpen]);

  const fetchImages = async () => {
    try {
      const response = await api.getImages();
      setImages(response.data);
    } catch (error) {
      console.error('Failed to fetch images:', error);
    }
  };

  const handleInputChange = (field, value) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const handleArrayChange = (field, index, key, value) => {
    setFormData(prev => ({
      ...prev,
      [field]: prev[field].map((item, i) => 
        i === index ? { ...item, [key]: value } : item
      )
    }));
  };

  const addArrayItem = (field) => {
    const newItem = field === 'ports' 
      ? { host_port: '', container_port: '', protocol: 'tcp' }
      : field === 'environment'
      ? { key: '', value: '' }
      : { host_path: '', container_path: '', read_only: false };

    setFormData(prev => ({
      ...prev,
      [field]: [...prev[field], newItem]
    }));
  };

  const removeArrayItem = (field, index) => {
    setFormData(prev => ({
      ...prev,
      [field]: prev[field].filter((_, i) => i !== index)
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      // Clean up the data
      const cleanData = {
        ...formData,
        ports: formData.ports.filter(p => p.host_port && p.container_port),
        environment: formData.environment.reduce((acc, env) => {
          if (env.key && env.value) {
            acc[env.key] = env.value;
          }
          return acc;
        }, {}),
        volumes: formData.volumes.filter(v => v.host_path && v.container_path),
        command: formData.command.filter(cmd => cmd.trim())
      };

      await api.createContainer(cleanData);
      onSuccess();
      handleClose();
    } catch (error) {
      setError(error.response?.data?.message || 'Failed to create container');
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setFormData({
      name: '',
      image: '',
      ports: [{ host_port: '', container_port: '', protocol: 'tcp' }],
      environment: [{ key: '', value: '' }],
      volumes: [{ host_path: '', container_path: '', read_only: false }],
      command: [],
      working_dir: '',
      restart_policy: 'no'
    });
    setError('');
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 overflow-y-auto">
      <div className="bg-white rounded-lg p-6 w-full max-w-4xl max-h-screen overflow-y-auto m-4">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-semibold">Create Container</h2>
          <button
            onClick={handleClose}
            className="text-gray-500 hover:text-gray-700"
          >
            ×
          </button>
        </div>

        {error && (
          <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Basic Information */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Container Name
              </label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => handleInputChange('name', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="my-container"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Image *
              </label>
              <select
                value={formData.image}
                onChange={(e) => handleInputChange('image', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                required
              >
                <option value="">Select an image</option>
                {images.map((image, index) => (
                  <option key={index} value={image.RepoTags?.[0] || image.Id}>
                    {image.RepoTags?.[0] || image.Id.substring(0, 12)}
                  </option>
                ))}
              </select>
            </div>
          </div>

          {/* Port Mappings */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Port Mappings
            </label>
            {formData.ports.map((port, index) => (
              <div key={index} className="flex gap-2 mb-2">
                <input
                  type="text"
                  placeholder="Host Port"
                  value={port.host_port}
                  onChange={(e) => handleArrayChange('ports', index, 'host_port', e.target.value)}
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <input
                  type="text"
                  placeholder="Container Port"
                  value={port.container_port}
                  onChange={(e) => handleArrayChange('ports', index, 'container_port', e.target.value)}
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <select
                  value={port.protocol}
                  onChange={(e) => handleArrayChange('ports', index, 'protocol', e.target.value)}
                  className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="tcp">TCP</option>
                  <option value="udp">UDP</option>
                </select>
                <button
                  type="button"
                  onClick={() => removeArrayItem('ports', index)}
                  className="px-3 py-2 text-red-600 hover:text-red-800"
                >
                  Remove
                </button>
              </div>
            ))}
            <button
              type="button"
              onClick={() => addArrayItem('ports')}
              className="text-blue-600 hover:text-blue-800"
            >
              + Add Port
            </button>
          </div>

          {/* Environment Variables */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Environment Variables
            </label>
            {formData.environment.map((env, index) => (
              <div key={index} className="flex gap-2 mb-2">
                <input
                  type="text"
                  placeholder="Key"
                  value={env.key}
                  onChange={(e) => handleArrayChange('environment', index, 'key', e.target.value)}
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <input
                  type="text"
                  placeholder="Value"
                  value={env.value}
                  onChange={(e) => handleArrayChange('environment', index, 'value', e.target.value)}
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <button
                  type="button"
                  onClick={() => removeArrayItem('environment', index)}
                  className="px-3 py-2 text-red-600 hover:text-red-800"
                >
                  Remove
                </button>
              </div>
            ))}
            <button
              type="button"
              onClick={() => addArrayItem('environment')}
              className="text-blue-600 hover:text-blue-800"
            >
              + Add Environment Variable
            </button>
          </div>

          {/* Volume Mappings */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Volume Mappings
            </label>
            {formData.volumes.map((volume, index) => (
              <div key={index} className="flex gap-2 mb-2 items-center">
                <input
                  type="text"
                  placeholder="Host Path"
                  value={volume.host_path}
                  onChange={(e) => handleArrayChange('volumes', index, 'host_path', e.target.value)}
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <input
                  type="text"
                  placeholder="Container Path"
                  value={volume.container_path}
                  onChange={(e) => handleArrayChange('volumes', index, 'container_path', e.target.value)}
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={volume.read_only}
                    onChange={(e) => handleArrayChange('volumes', index, 'read_only', e.target.checked)}
                    className="mr-1"
                  />
                  Read Only
                </label>
                <button
                  type="button"
                  onClick={() => removeArrayItem('volumes', index)}
                  className="px-3 py-2 text-red-600 hover:text-red-800"
                >
                  Remove
                </button>
              </div>
            ))}
            <button
              type="button"
              onClick={() => addArrayItem('volumes')}
              className="text-blue-600 hover:text-blue-800"
            >
              + Add Volume
            </button>
          </div>

          {/* Additional Options */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Working Directory
              </label>
              <input
                type="text"
                value={formData.working_dir}
                onChange={(e) => handleInputChange('working_dir', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="/app"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Restart Policy
              </label>
              <select
                value={formData.restart_policy}
                onChange={(e) => handleInputChange('restart_policy', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="no">No</option>
                <option value="always">Always</option>
                <option value="unless-stopped">Unless Stopped</option>
                <option value="on-failure">On Failure</option>
              </select>
            </div>
          </div>

          <div className="flex justify-end space-x-3">
            <button
              type="button"
              onClick={handleClose}
              className="px-4 py-2 text-gray-600 border border-gray-300 rounded-md hover:bg-gray-50"
              disabled={loading}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
              disabled={loading}
            >
              {loading ? 'Creating...' : 'Create Container'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default CreateContainerModal;
