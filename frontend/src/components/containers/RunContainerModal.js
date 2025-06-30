import React, { useState, useEffect } from 'react';
import { XMarkIcon, PlusIcon, TrashIcon } from '@heroicons/react/24/outline';
import api from '../../services/api';
import { toast } from 'react-toastify';

const RunContainerModal = ({ isOpen, onClose, onSuccess }) => {
  const [formData, setFormData] = useState({
    image: '',
    name: '',
    ports: [{ host: '', container: '' }],
    environment: [''],
    volumes: [''],
    command: '',
    workingDir: '',
    restartPolicy: 'no'
  });
  const [images, setImages] = useState([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (isOpen) {
      fetchImages();
    }
  }, [isOpen]);

  const fetchImages = async () => {
    try {
      const response = await api.getImages();
      setImages(response.data);
    } catch (err) {
      console.error('Failed to fetch images:', err);
    }
  };

  const handleInputChange = (field, value) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const handleArrayChange = (field, index, value) => {
    setFormData(prev => ({
      ...prev,
      [field]: prev[field].map((item, i) => i === index ? value : item)
    }));
  };

  const addArrayItem = (field) => {
    setFormData(prev => ({
      ...prev,
      [field]: [...prev[field], field === 'ports' ? { host: '', container: '' } : '']
    }));
  };

  const removeArrayItem = (field, index) => {
    setFormData(prev => ({
      ...prev,
      [field]: prev[field].filter((_, i) => i !== index)
    }));
  };

  const handlePortChange = (index, type, value) => {
    setFormData(prev => ({
      ...prev,
      ports: prev.ports.map((port, i) => 
        i === index ? { ...port, [type]: value } : port
      )
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!formData.image.trim()) {
      toast.error('Image is required');
      return;
    }

    try {
      setLoading(true);
      
      // Prepare the request data
      const requestData = {
        image: formData.image.trim(),
        name: formData.name.trim() || undefined,
        working_dir: formData.workingDir.trim() || undefined,
        restart_policy: formData.restartPolicy
      };

      // Add ports
      const validPorts = formData.ports.filter(p => p.host && p.container);
      if (validPorts.length > 0) {
        requestData.ports = {};
        validPorts.forEach(port => {
          requestData.ports[port.host] = port.container;
        });
      }

      // Add environment variables
      const validEnv = formData.environment.filter(env => env.trim());
      if (validEnv.length > 0) {
        requestData.environment = validEnv;
      }

      // Add volumes
      const validVolumes = formData.volumes.filter(vol => vol.trim());
      if (validVolumes.length > 0) {
        requestData.volumes = validVolumes;
      }

      // Add command
      if (formData.command.trim()) {
        requestData.command = formData.command.trim().split(' ').filter(cmd => cmd);
      }

      await api.runContainer(requestData);
      toast.success('Container created and started successfully');
      onSuccess();
      handleClose();
    } catch (err) {
      console.error('Failed to run container:', err);
      toast.error(`Failed to run container: ${err.response?.data?.message || err.message}`);
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setFormData({
      image: '',
      name: '',
      ports: [{ host: '', container: '' }],
      environment: [''],
      volumes: [''],
      command: '',
      workingDir: '',
      restartPolicy: 'no'
    });
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-10 mx-auto p-5 border w-11/12 md:w-3/4 lg:w-2/3 xl:w-1/2 shadow-lg rounded-md bg-white dark:bg-gray-800 max-h-screen overflow-y-auto">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-medium text-gray-900 dark:text-white">
            Run New Container
          </h3>
          <button
            onClick={handleClose}
            className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
          >
            <XMarkIcon className="h-6 w-6" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Image Selection */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Image *
            </label>
            <select
              value={formData.image}
              onChange={(e) => handleInputChange('image', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
              required
            >
              <option value="">Select an image</option>
              {images.map((image) => {
                const repoTag = image.RepoTags?.[0] || `${image.Id.substring(0, 12)}`;
                return (
                  <option key={image.Id} value={repoTag}>
                    {repoTag}
                  </option>
                );
              })}
            </select>
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
              Or type a custom image name (e.g., nginx:latest)
            </p>
            <input
              type="text"
              value={formData.image}
              onChange={(e) => handleInputChange('image', e.target.value)}
              placeholder="Or enter image name manually"
              className="w-full mt-2 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
            />
          </div>

          {/* Container Name */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Container Name (optional)
            </label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => handleInputChange('name', e.target.value)}
              placeholder="my-container"
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
            />
          </div>

          {/* Port Mappings */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Port Mappings
            </label>
            {formData.ports.map((port, index) => (
              <div key={index} className="flex gap-2 mb-2">
                <input
                  type="text"
                  value={port.host}
                  onChange={(e) => handlePortChange(index, 'host', e.target.value)}
                  placeholder="Host port (e.g., 8080)"
                  className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
                />
                <span className="flex items-center text-gray-500">:</span>
                <input
                  type="text"
                  value={port.container}
                  onChange={(e) => handlePortChange(index, 'container', e.target.value)}
                  placeholder="Container port (e.g., 80)"
                  className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
                />
                {formData.ports.length > 1 && (
                  <button
                    type="button"
                    onClick={() => removeArrayItem('ports', index)}
                    className="p-2 text-red-600 hover:text-red-800 dark:text-red-400"
                  >
                    <TrashIcon className="h-4 w-4" />
                  </button>
                )}
              </div>
            ))}
            <button
              type="button"
              onClick={() => addArrayItem('ports')}
              className="flex items-center gap-1 text-sm text-blue-600 hover:text-blue-800 dark:text-blue-400"
            >
              <PlusIcon className="h-4 w-4" />
              Add Port Mapping
            </button>
          </div>

          {/* Environment Variables */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Environment Variables
            </label>
            {formData.environment.map((env, index) => (
              <div key={index} className="flex gap-2 mb-2">
                <input
                  type="text"
                  value={env}
                  onChange={(e) => handleArrayChange('environment', index, e.target.value)}
                  placeholder="KEY=value"
                  className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
                />
                {formData.environment.length > 1 && (
                  <button
                    type="button"
                    onClick={() => removeArrayItem('environment', index)}
                    className="p-2 text-red-600 hover:text-red-800 dark:text-red-400"
                  >
                    <TrashIcon className="h-4 w-4" />
                  </button>
                )}
              </div>
            ))}
            <button
              type="button"
              onClick={() => addArrayItem('environment')}
              className="flex items-center gap-1 text-sm text-blue-600 hover:text-blue-800 dark:text-blue-400"
            >
              <PlusIcon className="h-4 w-4" />
              Add Environment Variable
            </button>
          </div>

          {/* Volumes */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Volume Mounts
            </label>
            {formData.volumes.map((volume, index) => (
              <div key={index} className="flex gap-2 mb-2">
                <input
                  type="text"
                  value={volume}
                  onChange={(e) => handleArrayChange('volumes', index, e.target.value)}
                  placeholder="/host/path:/container/path or volume_name:/container/path"
                  className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
                />
                {formData.volumes.length > 1 && (
                  <button
                    type="button"
                    onClick={() => removeArrayItem('volumes', index)}
                    className="p-2 text-red-600 hover:text-red-800 dark:text-red-400"
                  >
                    <TrashIcon className="h-4 w-4" />
                  </button>
                )}
              </div>
            ))}
            <button
              type="button"
              onClick={() => addArrayItem('volumes')}
              className="flex items-center gap-1 text-sm text-blue-600 hover:text-blue-800 dark:text-blue-400"
            >
              <PlusIcon className="h-4 w-4" />
              Add Volume Mount
            </button>
          </div>

          {/* Command */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Command (optional)
            </label>
            <input
              type="text"
              value={formData.command}
              onChange={(e) => handleInputChange('command', e.target.value)}
              placeholder="Override default command"
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
            />
          </div>

          {/* Working Directory */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Working Directory (optional)
            </label>
            <input
              type="text"
              value={formData.workingDir}
              onChange={(e) => handleInputChange('workingDir', e.target.value)}
              placeholder="/app"
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
            />
          </div>

          {/* Restart Policy */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Restart Policy
            </label>
            <select
              value={formData.restartPolicy}
              onChange={(e) => handleInputChange('restartPolicy', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
            >
              <option value="no">No</option>
              <option value="always">Always</option>
              <option value="unless-stopped">Unless Stopped</option>
              <option value="on-failure">On Failure</option>
            </select>
          </div>

          {/* Actions */}
          <div className="flex justify-end gap-3 pt-4 border-t border-gray-200 dark:border-gray-600">
            <button
              type="button"
              onClick={handleClose}
              className="px-4 py-2 bg-gray-300 text-gray-700 rounded-md hover:bg-gray-400 focus:outline-none focus:ring-2 focus:ring-gray-500"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
            >
              {loading ? 'Creating...' : 'Run Container'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default RunContainerModal;
