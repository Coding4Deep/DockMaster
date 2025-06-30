import React, { useState } from 'react';
import api from '../../services/api';

const ImageSearchModal = ({ isOpen, onClose, onSuccess }) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState({ local: [], docker_hub: [] });
  const [loading, setLoading] = useState(false);
  const [pulling, setPulling] = useState({});
  const [selectedImage, setSelectedImage] = useState(null);
  const [imageDetails, setImageDetails] = useState(null);
  const [error, setError] = useState('');

  const handleSearch = async (e) => {
    e.preventDefault();
    if (!searchQuery.trim()) return;

    setLoading(true);
    setError('');

    try {
      const response = await api.searchImages(searchQuery);
      setSearchResults(response.data);
    } catch (error) {
      setError('Failed to search images');
      console.error('Search failed:', error);
    } finally {
      setLoading(false);
    }
  };

  const handlePullImage = async (imageName, tag = 'latest') => {
    const pullKey = `${imageName}:${tag}`;
    setPulling(prev => ({ ...prev, [pullKey]: true }));

    try {
      await api.pullImage(imageName, tag);
      onSuccess();
      // Remove from docker hub results since it's now local
      setSearchResults(prev => ({
        ...prev,
        docker_hub: prev.docker_hub.filter(img => img.name !== imageName)
      }));
    } catch (error) {
      setError(`Failed to pull ${pullKey}: ${error.response?.data?.message || error.message}`);
    } finally {
      setPulling(prev => ({ ...prev, [pullKey]: false }));
    }
  };

  const handleViewDetails = async (imageName) => {
    try {
      const response = await api.getImageDetails(imageName);
      setImageDetails(response.data);
      setSelectedImage(imageName);
    } catch (error) {
      setError('Failed to get image details');
    }
  };

  const handleClose = () => {
    setSearchQuery('');
    setSearchResults({ local: [], docker_hub: [] });
    setSelectedImage(null);
    setImageDetails(null);
    setError('');
    onClose();
  };

  const formatSize = (bytes) => {
    if (!bytes) return 'Unknown';
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${sizes[i]}`;
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 w-full max-w-6xl max-h-screen overflow-y-auto m-4">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-semibold">Search & Pull Images</h2>
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

        {/* Search Form */}
        <form onSubmit={handleSearch} className="mb-6">
          <div className="flex gap-2">
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search for images (e.g., nginx, ubuntu, node)"
              className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <button
              type="submit"
              disabled={loading}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
            >
              {loading ? 'Searching...' : 'Search'}
            </button>
          </div>
        </form>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Local Images */}
          <div>
            <h3 className="text-lg font-medium mb-3">Local Images</h3>
            <div className="space-y-2 max-h-96 overflow-y-auto">
              {searchResults.local.length === 0 ? (
                <p className="text-gray-500">No local images found</p>
              ) : (
                searchResults.local.map((image, index) => (
                  <div key={index} className="border rounded-lg p-3">
                    <div className="flex justify-between items-start">
                      <div>
                        <h4 className="font-medium">{image.RepoTags?.[0] || 'Unknown'}</h4>
                        <p className="text-sm text-gray-600">Size: {formatSize(image.Size)}</p>
                        <p className="text-sm text-gray-600">
                          Created: {new Date(image.Created * 1000).toLocaleDateString()}
                        </p>
                      </div>
                      <span className="px-2 py-1 bg-green-100 text-green-800 text-xs rounded">
                        Local
                      </span>
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>

          {/* Docker Hub Images */}
          <div>
            <h3 className="text-lg font-medium mb-3">Docker Hub Images</h3>
            <div className="space-y-2 max-h-96 overflow-y-auto">
              {searchResults.docker_hub.length === 0 ? (
                <p className="text-gray-500">No Docker Hub images found</p>
              ) : (
                searchResults.docker_hub.map((image, index) => (
                  <div key={index} className="border rounded-lg p-3">
                    <div className="flex justify-between items-start mb-2">
                      <div className="flex-1">
                        <h4 className="font-medium">{image.name}</h4>
                        <p className="text-sm text-gray-600 line-clamp-2">
                          {image.short_description || image.description}
                        </p>
                        <div className="flex items-center gap-4 mt-1">
                          <span className="text-xs text-gray-500">
                            ⭐ {image.star_count}
                          </span>
                          <span className="text-xs text-gray-500">
                            📥 {image.pull_count.toLocaleString()}
                          </span>
                          {image.is_official && (
                            <span className="px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded">
                              Official
                            </span>
                          )}
                        </div>
                      </div>
                    </div>
                    <div className="flex gap-2">
                      <button
                        onClick={() => handleViewDetails(image.name)}
                        className="px-3 py-1 text-sm border border-gray-300 rounded hover:bg-gray-50"
                      >
                        Details
                      </button>
                      <button
                        onClick={() => handlePullImage(image.name)}
                        disabled={pulling[`${image.name}:latest`]}
                        className="px-3 py-1 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
                      >
                        {pulling[`${image.name}:latest`] ? 'Pulling...' : 'Pull'}
                      </button>
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>

        {/* Image Details Modal */}
        {selectedImage && imageDetails && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-60">
            <div className="bg-white rounded-lg p-6 w-full max-w-2xl max-h-screen overflow-y-auto m-4">
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-lg font-semibold">{selectedImage}</h3>
                <button
                  onClick={() => {
                    setSelectedImage(null);
                    setImageDetails(null);
                  }}
                  className="text-gray-500 hover:text-gray-700"
                >
                  ×
                </button>
              </div>

              <div className="space-y-4">
                <div>
                  <h4 className="font-medium mb-2">Description</h4>
                  <p className="text-gray-700">{imageDetails.full_description || imageDetails.description}</p>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <h4 className="font-medium mb-2">Stars</h4>
                    <p>{imageDetails.star_count}</p>
                  </div>
                  <div>
                    <h4 className="font-medium mb-2">Pulls</h4>
                    <p>{imageDetails.pull_count?.toLocaleString()}</p>
                  </div>
                  <div>
                    <h4 className="font-medium mb-2">Last Updated</h4>
                    <p>{new Date(imageDetails.last_updated).toLocaleDateString()}</p>
                  </div>
                  <div>
                    <h4 className="font-medium mb-2">Official</h4>
                    <p>{imageDetails.is_official ? 'Yes' : 'No'}</p>
                  </div>
                </div>

                <div className="flex justify-end gap-2">
                  <button
                    onClick={() => {
                      setSelectedImage(null);
                      setImageDetails(null);
                    }}
                    className="px-4 py-2 text-gray-600 border border-gray-300 rounded-md hover:bg-gray-50"
                  >
                    Close
                  </button>
                  <button
                    onClick={() => handlePullImage(selectedImage)}
                    disabled={pulling[`${selectedImage}:latest`]}
                    className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
                  >
                    {pulling[`${selectedImage}:latest`] ? 'Pulling...' : 'Pull Image'}
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default ImageSearchModal;
