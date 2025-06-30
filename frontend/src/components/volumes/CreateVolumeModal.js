import React, { useState } from 'react';
import api from '../../services/api';

const CreateVolumeModal = ({ isOpen, onClose, onSuccess }) => {
  const [formData, setFormData] = useState({
    name: '',
    driver: 'local',
    driver_opts: [{ key: '', value: '' }],
    labels: [{ key: '', value: '' }]
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

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
    setFormData(prev => ({
      ...prev,
      [field]: [...prev[field], { key: '', value: '' }]
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
        name: formData.name,
        driver: formData.driver,
        driver_opts: formData.driver_opts.reduce((acc, opt) => {
          if (opt.key && opt.value) {
            acc[opt.key] = opt.value;
          }
          return acc;
        }, {}),
        labels: formData.labels.reduce((acc, label) => {
          if (label.key && label.value) {
            acc[label.key] = label.value;
          }
          return acc;
        }, {})
      };

      await api.createVolume(cleanData);
      onSuccess();
      handleClose();
    } catch (error) {
      setError(error.response?.data?.message || 'Failed to create volume');
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setFormData({
      name: '',
      driver: 'local',
      driver_opts: [{ key: '', value: '' }],
      labels: [{ key: '', value: '' }]
    });
    setError('');
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 w-full max-w-2xl max-h-screen overflow-y-auto m-4">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-semibold">Create Volume</h2>
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
                Volume Name
              </label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => handleInputChange('name', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="my-volume"
              />
              <p className="text-xs text-gray-500 mt-1">Leave empty for auto-generated name</p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Driver
              </label>
              <select
                value={formData.driver}
                onChange={(e) => handleInputChange('driver', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="local">local</option>
                <option value="nfs">nfs</option>
                <option value="cifs">cifs</option>
              </select>
            </div>
          </div>

          {/* Driver Options */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Driver Options
            </label>
            {formData.driver_opts.map((opt, index) => (
              <div key={index} className="flex gap-2 mb-2">
                <input
                  type="text"
                  placeholder="Option Key"
                  value={opt.key}
                  onChange={(e) => handleArrayChange('driver_opts', index, 'key', e.target.value)}
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <input
                  type="text"
                  placeholder="Option Value"
                  value={opt.value}
                  onChange={(e) => handleArrayChange('driver_opts', index, 'value', e.target.value)}
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <button
                  type="button"
                  onClick={() => removeArrayItem('driver_opts', index)}
                  className="px-3 py-2 text-red-600 hover:text-red-800"
                >
                  Remove
                </button>
              </div>
            ))}
            <button
              type="button"
              onClick={() => addArrayItem('driver_opts')}
              className="text-blue-600 hover:text-blue-800"
            >
              + Add Driver Option
            </button>
          </div>

          {/* Labels */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Labels
            </label>
            {formData.labels.map((label, index) => (
              <div key={index} className="flex gap-2 mb-2">
                <input
                  type="text"
                  placeholder="Label Key"
                  value={label.key}
                  onChange={(e) => handleArrayChange('labels', index, 'key', e.target.value)}
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <input
                  type="text"
                  placeholder="Label Value"
                  value={label.value}
                  onChange={(e) => handleArrayChange('labels', index, 'value', e.target.value)}
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <button
                  type="button"
                  onClick={() => removeArrayItem('labels', index)}
                  className="px-3 py-2 text-red-600 hover:text-red-800"
                >
                  Remove
                </button>
              </div>
            ))}
            <button
              type="button"
              onClick={() => addArrayItem('labels')}
              className="text-blue-600 hover:text-blue-800"
            >
              + Add Label
            </button>
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
              {loading ? 'Creating...' : 'Create Volume'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default CreateVolumeModal;
