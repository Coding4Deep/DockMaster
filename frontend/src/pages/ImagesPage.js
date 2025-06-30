import React, { useState, useEffect } from 'react';
import { 
  TrashIcon,
  PhotoIcon,
  MagnifyingGlassIcon,
  CloudArrowDownIcon,
  EyeIcon,
  InformationCircleIcon
} from '@heroicons/react/24/outline';
import LoadingSpinner from '../components/common/LoadingSpinner';
import api from '../services/api';
import { toast } from 'react-toastify';

const ImagesPage = () => {
  const [images, setImages] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState(null);
  const [searching, setSearching] = useState(false);
  const [selectedImage, setSelectedImage] = useState(null);
  const [imageDetails, setImageDetails] = useState(null);
  const [showImageModal, setShowImageModal] = useState(false);
  const [pulling, setPulling] = useState(false);

  useEffect(() => {
    fetchImages();
  }, []);

  const fetchImages = async () => {
    try {
      setLoading(true);
      const response = await api.getImages();
      setImages(response.data);
      setError(null);
    } catch (err) {
      console.error('Failed to fetch images:', err);
      setError('Failed to load images');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteImage = async (imageId) => {
    if (window.confirm('Are you sure you want to delete this image?')) {
      try {
        await api.deleteImage(imageId, true);
        await fetchImages();
        toast.success('Image deleted successfully');
      } catch (err) {
        console.error('Failed to delete image:', err);
        toast.error(`Failed to delete image: ${err.response?.data?.message || err.message}`);
      }
    }
  };

  const handleSearch = async () => {
    if (!searchQuery.trim()) {
      toast.error('Please enter a search query');
      return;
    }

    try {
      setSearching(true);
      const response = await api.searchImages(searchQuery);
      setSearchResults(response.data);
    } catch (err) {
      console.error('Failed to search images:', err);
      toast.error('Failed to search images');
    } finally {
      setSearching(false);
    }
  };

  const handleViewImageDetails = async (imageName) => {
    try {
      setSelectedImage(imageName);
      setShowImageModal(true);
      
      // Try to get details from local images first
      const localImage = images.find(img => 
        img.RepoTags && img.RepoTags.some(tag => tag.includes(imageName))
      );
      
      if (localImage) {
        const response = await api.inspectImage(localImage.Id);
        setImageDetails(response.data);
      } else {
        // For Docker Hub images, show basic info
        const hubImage = searchResults?.dockerhub?.find(img => img.name === imageName);
        if (hubImage) {
          setImageDetails({
            name: hubImage.name,
            description: hubImage.description,
            stars: hubImage.star_count,
            official: hubImage.is_official,
            automated: hubImage.is_automated,
            isHubImage: true
          });
        }
      }
    } catch (err) {
      console.error('Failed to get image details:', err);
      toast.error('Failed to get image details');
    }
  };

  const handlePullImage = async (imageName) => {
    try {
      setPulling(true);
      await api.pullImage(imageName);
      toast.success(`Image ${imageName} pulled successfully`);
      await fetchImages();
      setSearchResults(null);
      setSearchQuery('');
    } catch (err) {
      console.error('Failed to pull image:', err);
      toast.error(`Failed to pull image: ${err.response?.data?.message || err.message}`);
    } finally {
      setPulling(false);
    }
  };

  const formatSize = (bytes) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatDate = (timestamp) => {
    return new Date(timestamp * 1000).toLocaleDateString();
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
            Images
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            Manage your Docker images
          </p>
        </div>
        <button
          onClick={fetchImages}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          Refresh
        </button>
      </div>

      {/* Search Section */}
      <div className="bg-white dark:bg-gray-800 shadow rounded-lg p-6">
        <h2 className="text-lg font-medium text-gray-900 dark:text-white mb-4">
          Search Images
        </h2>
        <div className="flex gap-4">
          <div className="flex-1">
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search for images (e.g., nginx, ubuntu, node)"
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
              onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
            />
          </div>
          <button
            onClick={handleSearch}
            disabled={searching}
            className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-green-500 disabled:opacity-50 flex items-center gap-2"
          >
            <MagnifyingGlassIcon className="h-4 w-4" />
            {searching ? 'Searching...' : 'Search'}
          </button>
        </div>
      </div>

      {/* Search Results */}
      {searchResults && (
        <div className="bg-white dark:bg-gray-800 shadow rounded-lg p-6">
          <h2 className="text-lg font-medium text-gray-900 dark:text-white mb-4">
            Search Results
          </h2>
          
          {/* Local Results */}
          {searchResults.local && searchResults.local.length > 0 && (
            <div className="mb-6">
              <h3 className="text-md font-medium text-gray-700 dark:text-gray-300 mb-2">
                Local Images
              </h3>
              <div className="grid gap-2">
                {searchResults.local.map((image) => (
                  <div key={image.id} className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700 rounded-md">
                    <div>
                      <div className="font-medium text-gray-900 dark:text-white">
                        {image.repo_tags?.[0] || '<none>:<none>'}
                      </div>
                      <div className="text-sm text-gray-500 dark:text-gray-400">
                        {formatSize(image.size)} • {formatDate(image.created)}
                      </div>
                    </div>
                    <div className="flex gap-2">
                      <button
                        onClick={() => handleViewImageDetails(image.repo_tags?.[0])}
                        className="p-1 text-blue-600 hover:text-blue-800 dark:text-blue-400"
                        title="View Details"
                      >
                        <EyeIcon className="h-4 w-4" />
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Docker Hub Results */}
          {searchResults.dockerhub && searchResults.dockerhub.length > 0 && (
            <div>
              <h3 className="text-md font-medium text-gray-700 dark:text-gray-300 mb-2">
                Docker Hub Images
              </h3>
              <div className="grid gap-2">
                {searchResults.dockerhub.map((image) => (
                  <div key={image.name} className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700 rounded-md">
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <span className="font-medium text-gray-900 dark:text-white">
                          {image.name}
                        </span>
                        {image.is_official && (
                          <span className="px-2 py-1 text-xs bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200 rounded">
                            Official
                          </span>
                        )}
                        {image.is_automated && (
                          <span className="px-2 py-1 text-xs bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200 rounded">
                            Automated
                          </span>
                        )}
                      </div>
                      <div className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                        {image.description || 'No description available'}
                      </div>
                      <div className="text-xs text-gray-400 dark:text-gray-500 mt-1">
                        ⭐ {image.star_count} stars
                      </div>
                    </div>
                    <div className="flex gap-2">
                      <button
                        onClick={() => handleViewImageDetails(image.name)}
                        className="p-1 text-blue-600 hover:text-blue-800 dark:text-blue-400"
                        title="View Details"
                      >
                        <InformationCircleIcon className="h-4 w-4" />
                      </button>
                      <button
                        onClick={() => handlePullImage(image.name)}
                        disabled={pulling}
                        className="p-1 text-green-600 hover:text-green-800 dark:text-green-400 disabled:opacity-50"
                        title="Pull Image"
                      >
                        <CloudArrowDownIcon className="h-4 w-4" />
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {searchResults.local?.length === 0 && searchResults.dockerhub?.length === 0 && (
            <div className="text-center py-8 text-gray-500 dark:text-gray-400">
              No images found for "{searchQuery}"
            </div>
          )}
        </div>
      )}

      {error && (
        <div className="bg-red-50 dark:bg-red-900 border border-red-200 dark:border-red-700 rounded-md p-4">
          <div className="text-red-800 dark:text-red-200">{error}</div>
        </div>
      )}

      {/* Local Images Table */}
      <div className="bg-white dark:bg-gray-800 shadow overflow-hidden sm:rounded-md">
        <div className="px-4 py-5 sm:p-6">
          <h2 className="text-lg font-medium text-gray-900 dark:text-white mb-4">
            Local Images
          </h2>
          {images.length === 0 ? (
            <div className="text-center py-12">
              <PhotoIcon className="mx-auto h-12 w-12 text-gray-400" />
              <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-white">No images</h3>
              <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                No Docker images found on this system.
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                <thead className="bg-gray-50 dark:bg-gray-700">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Repository
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Tag
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Image ID
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Created
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Size
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                  {images.map((image) => {
                    const repoTag = image.RepoTags?.[0] || '<none>:<none>';
                    const [repository, tag] = repoTag.split(':');
                    
                    return (
                      <tr key={image.Id}>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <div className="text-sm font-medium text-gray-900 dark:text-white">
                            {repository || '<none>'}
                          </div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
                            {tag || '<none>'}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <div className="text-sm text-gray-900 dark:text-white font-mono">
                            {image.Id.replace('sha256:', '').substring(0, 12)}
                          </div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                          {formatDate(image.Created)}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                          {formatSize(image.Size)}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                          <div className="flex gap-2">
                            <button
                              onClick={() => handleViewImageDetails(repoTag)}
                              className="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300"
                              title="View Details"
                            >
                              <EyeIcon className="h-4 w-4" />
                            </button>
                            <button
                              onClick={() => handleDeleteImage(image.Id)}
                              className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300"
                              title="Delete Image"
                            >
                              <TrashIcon className="h-4 w-4" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {/* Image Details Modal */}
      {showImageModal && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
          <div className="relative top-20 mx-auto p-5 border w-11/12 md:w-3/4 lg:w-1/2 shadow-lg rounded-md bg-white dark:bg-gray-800">
            <div className="mt-3">
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-lg font-medium text-gray-900 dark:text-white">
                  Image Details: {selectedImage}
                </h3>
                <button
                  onClick={() => {
                    setShowImageModal(false);
                    setImageDetails(null);
                    setSelectedImage(null);
                  }}
                  className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
                >
                  ✕
                </button>
              </div>
              
              {imageDetails ? (
                <div className="space-y-4">
                  {imageDetails.isHubImage ? (
                    // Docker Hub image details
                    <div className="space-y-2">
                      <div><strong>Name:</strong> {imageDetails.name}</div>
                      <div><strong>Description:</strong> {imageDetails.description || 'N/A'}</div>
                      <div><strong>Stars:</strong> {imageDetails.stars}</div>
                      <div><strong>Official:</strong> {imageDetails.official ? 'Yes' : 'No'}</div>
                      <div><strong>Automated:</strong> {imageDetails.automated ? 'Yes' : 'No'}</div>
                    </div>
                  ) : (
                    // Local image details
                    <div className="space-y-2 text-sm">
                      {imageDetails.Config && (
                        <>
                          <div><strong>Architecture:</strong> {imageDetails.Architecture}</div>
                          <div><strong>OS:</strong> {imageDetails.Os}</div>
                          <div><strong>Created:</strong> {new Date(imageDetails.Created).toLocaleString()}</div>
                          <div><strong>Size:</strong> {formatSize(imageDetails.Size)}</div>
                          {imageDetails.Config.ExposedPorts && (
                            <div>
                              <strong>Exposed Ports:</strong> {Object.keys(imageDetails.Config.ExposedPorts).join(', ')}
                            </div>
                          )}
                          {imageDetails.Config.Env && (
                            <div>
                              <strong>Environment:</strong>
                              <ul className="list-disc list-inside ml-4 mt-1">
                                {imageDetails.Config.Env.slice(0, 5).map((env, idx) => (
                                  <li key={idx} className="text-xs">{env}</li>
                                ))}
                                {imageDetails.Config.Env.length > 5 && (
                                  <li className="text-xs">... and {imageDetails.Config.Env.length - 5} more</li>
                                )}
                              </ul>
                            </div>
                          )}
                        </>
                      )}
                    </div>
                  )}
                </div>
              ) : (
                <div className="flex justify-center py-4">
                  <LoadingSpinner />
                </div>
              )}
              
              <div className="flex justify-end mt-6">
                <button
                  onClick={() => {
                    setShowImageModal(false);
                    setImageDetails(null);
                    setSelectedImage(null);
                  }}
                  className="px-4 py-2 bg-gray-300 text-gray-700 rounded-md hover:bg-gray-400 focus:outline-none focus:ring-2 focus:ring-gray-500"
                >
                  Close
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default ImagesPage;
