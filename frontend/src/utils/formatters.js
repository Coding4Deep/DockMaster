// Format bytes to human readable format
export const formatBytes = (bytes, decimals = 2) => {
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
};

// Format timestamp to relative time
export const formatRelativeTime = (timestamp) => {
  const now = new Date();
  const date = new Date(timestamp * 1000); // Convert Unix timestamp to milliseconds
  const diffInSeconds = Math.floor((now - date) / 1000);

  if (diffInSeconds < 60) {
    return `${diffInSeconds} seconds ago`;
  } else if (diffInSeconds < 3600) {
    const minutes = Math.floor(diffInSeconds / 60);
    return `${minutes} minute${minutes > 1 ? 's' : ''} ago`;
  } else if (diffInSeconds < 86400) {
    const hours = Math.floor(diffInSeconds / 3600);
    return `${hours} hour${hours > 1 ? 's' : ''} ago`;
  } else if (diffInSeconds < 2592000) {
    const days = Math.floor(diffInSeconds / 86400);
    return `${days} day${days > 1 ? 's' : ''} ago`;
  } else {
    const months = Math.floor(diffInSeconds / 2592000);
    return `${months} month${months > 1 ? 's' : ''} ago`;
  }
};

// Format container status
export const formatContainerStatus = (state, status) => {
  const statusMap = {
    running: { text: 'Running', class: 'status-running' },
    exited: { text: 'Stopped', class: 'status-stopped' },
    paused: { text: 'Paused', class: 'status-paused' },
    restarting: { text: 'Restarting', class: 'status-warning' },
    removing: { text: 'Removing', class: 'status-warning' },
    dead: { text: 'Dead', class: 'status-stopped' },
    created: { text: 'Created', class: 'status-paused' },
  };

  const statusInfo = statusMap[state.toLowerCase()] || { text: state, class: 'status-paused' };
  return statusInfo;
};

// Format port mappings
export const formatPorts = (ports) => {
  if (!ports || ports.length === 0) return 'None';
  
  return ports.map(port => {
    if (port.PublicPort) {
      return `${port.PublicPort}:${port.PrivatePort}/${port.Type}`;
    }
    return `${port.PrivatePort}/${port.Type}`;
  }).join(', ');
};

// Format container names (remove leading slash)
export const formatContainerName = (names) => {
  if (!names || names.length === 0) return 'Unknown';
  return names[0].replace(/^\//, '');
};

// Format image tags
export const formatImageTags = (repoTags) => {
  if (!repoTags || repoTags.length === 0) return '<none>';
  return repoTags.join(', ');
};

// Format network driver
export const formatNetworkDriver = (driver) => {
  const driverMap = {
    bridge: 'Bridge',
    host: 'Host',
    overlay: 'Overlay',
    macvlan: 'MACVLAN',
    none: 'None',
  };
  
  return driverMap[driver] || driver;
};

// Format volume driver
export const formatVolumeDriver = (driver) => {
  const driverMap = {
    local: 'Local',
    nfs: 'NFS',
    cifs: 'CIFS',
  };
  
  return driverMap[driver] || driver;
};

// Truncate text with ellipsis
export const truncateText = (text, maxLength = 50) => {
  if (!text) return '';
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength) + '...';
};

// Format CPU percentage
export const formatCPUPercent = (percent) => {
  return `${percent.toFixed(2)}%`;
};

// Format memory usage
export const formatMemoryUsage = (usage, limit) => {
  const usageFormatted = formatBytes(usage);
  const limitFormatted = formatBytes(limit);
  const percent = ((usage / limit) * 100).toFixed(1);
  return `${usageFormatted} / ${limitFormatted} (${percent}%)`;
};

// Format uptime
export const formatUptime = (startedAt) => {
  if (!startedAt) return 'N/A';
  
  const now = new Date();
  const started = new Date(startedAt);
  const diffInSeconds = Math.floor((now - started) / 1000);
  
  if (diffInSeconds < 60) {
    return `${diffInSeconds}s`;
  } else if (diffInSeconds < 3600) {
    const minutes = Math.floor(diffInSeconds / 60);
    const seconds = diffInSeconds % 60;
    return `${minutes}m ${seconds}s`;
  } else if (diffInSeconds < 86400) {
    const hours = Math.floor(diffInSeconds / 3600);
    const minutes = Math.floor((diffInSeconds % 3600) / 60);
    return `${hours}h ${minutes}m`;
  } else {
    const days = Math.floor(diffInSeconds / 86400);
    const hours = Math.floor((diffInSeconds % 86400) / 3600);
    return `${days}d ${hours}h`;
  }
};
