// Base API configuration
const API_BASE = '/api';

// Generic API request helper
async function apiRequest(endpoint, options = {}) {
  const url = `${API_BASE}${endpoint}`;
  const config = {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  };

  try {
    const response = await fetch(url, config);
    
    if (!response.ok) {
      throw new Error(`API Error: ${response.status} ${response.statusText}`);
    }
    
    return await response.json();
  } catch (error) {
    console.error('API request failed:', error);
    throw error;
  }
}

// Disk Storage API
export const diskAPI = {
  // Get all disk statistics
  getDiskStats: () => apiRequest('/disk-stats'),
  
  // Get SMART status for specific disk
  getSmartStatus: (disk) => apiRequest(`/disk/smart?disk=${disk}`),
  
  // Get disk partitions
  getPartitions: (disk) => apiRequest(`/disk/partitions?disk=${disk}`),
  
  // Mount/unmount disk
  mountDisk: (disk, mountpoint) => apiRequest(`/disk/${disk}/mount`, {
    method: 'POST',
    body: JSON.stringify({ mountpoint }),
  }),
  
  unmountDisk: (disk) => apiRequest(`/disk/${disk}/unmount`, {
    method: 'POST',
  }),
};

// SCSI Targets API
export const scsiAPI = {
  // Get all SCSI targets
  getTargets: () => apiRequest('/scsi-targets'),
  
  // Create new SCSI target
  createTarget: (targetData) => apiRequest('/scsi-targets', {
    method: 'POST',
    body: JSON.stringify(targetData),
  }),
  
  // Update SCSI target
  updateTarget: (targetId, targetData) => apiRequest(`/scsi-targets`, {
    method: 'PUT',
    body: JSON.stringify({ ...targetData, name: targetId }),
  }),
  
  // Delete SCSI target
  deleteTarget: (targetId) => apiRequest(`/scsi-targets`, {
    method: 'DELETE',
    body: JSON.stringify({ name: targetId }),
  }),
  
  // Toggle target status
  toggleTarget: (targetId) => apiRequest(`/scsi-targets`, {
    method: 'POST',
    body: JSON.stringify({ name: targetId }),
  }),
  
  // Get target sessions
  getSessions: (targetId) => apiRequest(`/scsi-targets/${targetId}/sessions`),
};

// Samba Shares API
export const sambaAPI = {
  // Get all Samba shares
  getShares: () => apiRequest('/samba-shares'),
  
  // Create new Samba share
  createShare: (shareData) => apiRequest('/samba-shares', {
    method: 'POST',
    body: JSON.stringify(shareData),
  }),
  
  // Update Samba share
  updateShare: (shareId, shareData) => apiRequest(`/samba-shares/${shareId}`, {
    method: 'PUT',
    body: JSON.stringify(shareData),
  }),
  
  // Delete Samba share
  deleteShare: (shareId) => apiRequest(`/samba-shares`, {
    method: 'DELETE',
    body: JSON.stringify({ name: shareId }),
  }),
  
  // Toggle share availability
  toggleShare: (shareId) => apiRequest(`/samba-shares`, {
    method: 'POST',
    body: JSON.stringify({ name: shareId }),
  }),
  
  // Get share connections
  getConnections: (shareId) => apiRequest(`/samba-shares/${shareId}/connections`),
  
  // Test share configuration
  testConfig: (shareData) => apiRequest('/samba-shares/test', {
    method: 'POST',
    body: JSON.stringify(shareData),
  }),
};

// NFS Exports API
export const nfsAPI = {
  // Get all NFS exports
  getExports: () => apiRequest('/nfs-exports'),
  
  // Create new NFS export
  createExport: (exportData) => apiRequest('/nfs-exports', {
    method: 'POST',
    body: JSON.stringify(exportData),
  }),
  
  // Update NFS export
  updateExport: (exportId, exportData) => apiRequest(`/nfs-exports/${exportId}`, {
    method: 'PUT',
    body: JSON.stringify(exportData),
  }),
  
  // Delete NFS export
  async deleteExport(exportPath) {
    return apiRequest(`/nfs-exports?path=${encodeURIComponent(exportPath)}`, {
      method: 'DELETE',
    });
  },
  
  // Get export status
  getExportStatus: (exportId) => apiRequest(`/nfs-exports/${exportId}/status`),
  
  // Reload NFS configuration
  reloadConfig: () => apiRequest('/nfs-exports/reload', {
    method: 'POST',
  }),
};

// System Statistics API
export const systemAPI = {
  // Get system overview
  getOverview: () => apiRequest('/system/overview'),
  
  // Get CPU statistics
  getCpuStats: () => apiRequest('/system/cpu'),
  
  // Get memory statistics
  getMemoryStats: () => apiRequest('/system/memory'),
  
  // Get network statistics
  getNetworkStats: () => apiRequest('/system/network'),
  
  // Get storage health
  getStorageHealth: () => apiRequest('/system/storage-health'),
  
  // Get system processes
  getProcesses: () => apiRequest('/system/processes'),
  
  // Get system logs
  getLogs: (options = {}) => apiRequest('/system/logs', {
    method: 'POST',
    body: JSON.stringify(options),
  }),
  
  // Get disk I/O rates
  getDiskIORates: () => apiRequest('/system/disk-io'),
  
  // Get network I/O rates
  getNetworkIORates: () => apiRequest('/system/network-io'),
  
  // Reboot system
  reboot: () => apiRequest('/system/reboot', { method: 'POST' }),
  
  // Shutdown system
  shutdown: () => apiRequest('/system/shutdown', { method: 'POST' }),
};

// Services API
export const servicesAPI = {
  // Get all services status
  getServices: () => apiRequest('/services'),
  
  // Get specific service status
  getService: (serviceName) => apiRequest(`/services/${serviceName}`),
  
  // Start service
  startService: (serviceName) => apiRequest(`/services/${serviceName}/start`, {
    method: 'POST',
  }),
  
  // Stop service
  stopService: (serviceName) => apiRequest(`/services/${serviceName}/stop`, {
    method: 'POST',
  }),
  
  // Restart service
  restartService: (serviceName) => apiRequest(`/services/${serviceName}/restart`, {
    method: 'POST',
  }),
  
  // Enable service (boot startup)
  enableService: (serviceName) => apiRequest(`/services/${serviceName}/enable`, {
    method: 'POST',
  }),
  
  // Disable service (boot startup)
  disableService: (serviceName) => apiRequest(`/services/${serviceName}/disable`, {
    method: 'POST',
  }),
  
  // Get service logs
  getServiceLogs: (serviceName, options = {}) => apiRequest(`/services/${serviceName}/logs`, {
    method: 'POST',
    body: JSON.stringify(options),
  }),
};

// Users and Groups API
export const usersAPI = {
  // Get all users
  getUsers: () => apiRequest('/users'),
  
  // Create user
  createUser: (userData) => apiRequest('/users', {
    method: 'POST',
    body: JSON.stringify(userData),
  }),
  
  // Update user
  updateUser: (username, userData) => apiRequest(`/users/${username}`, {
    method: 'PUT',
    body: JSON.stringify(userData),
  }),
  
  // Delete user
  deleteUser: (username) => apiRequest(`/users/${username}`, {
    method: 'DELETE',
  }),
  
  // Get all groups
  getGroups: () => apiRequest('/groups'),
  
  // Create group
  createGroup: (groupData) => apiRequest('/groups', {
    method: 'POST',
    body: JSON.stringify(groupData),
  }),
  
  // Add user to group
  addUserToGroup: (username, groupname) => apiRequest(`/users/${username}/groups/${groupname}`, {
    method: 'POST',
  }),
  
  // Remove user from group
  removeUserFromGroup: (username, groupname) => apiRequest(`/users/${username}/groups/${groupname}`, {
    method: 'DELETE',
  }),
};

// Backup API
export const backupAPI = {
  // Get all backup jobs
  getBackupJobs: () => apiRequest('/backup/jobs'),
  
  // Create backup job
  createBackupJob: (jobData) => apiRequest('/backup/jobs', {
    method: 'POST',
    body: JSON.stringify(jobData),
  }),
  
  // Run backup job
  runBackupJob: (jobId) => apiRequest(`/backup/jobs/${jobId}/run`, {
    method: 'POST',
  }),
  
  // Get backup history
  getBackupHistory: (jobId) => apiRequest(`/backup/jobs/${jobId}/history`),
  
  // Restore from backup
  restoreBackup: (backupId) => apiRequest(`/backup/restore/${backupId}`, {
    method: 'POST',
  }),
};

// Settings API
export const settingsAPI = {
  // Get system settings
  getSettings: () => apiRequest('/settings'),
  
  // Update system settings
  updateSettings: (settings) => apiRequest('/settings', {
    method: 'PUT',
    body: JSON.stringify(settings),
  }),
  
  // Get network configuration
  getNetworkConfig: () => apiRequest('/settings/network'),
  
  // Update network configuration
  updateNetworkConfig: (config) => apiRequest('/settings/network', {
    method: 'PUT',
    body: JSON.stringify(config),
  }),
  
  // Get timezone settings
  getTimezone: () => apiRequest('/settings/timezone'),
  
  // Update timezone
  updateTimezone: (timezone) => apiRequest('/settings/timezone', {
    method: 'PUT',
    body: JSON.stringify({ timezone }),
  }),
};

// Real-time updates (WebSocket/SSE)
export class RealtimeAPI {
  constructor() {
    this.eventSource = null;
    this.websocket = null;
  }
  
  // Server-Sent Events for system stats
  connectSSE(onMessage, onError) {
    this.eventSource = new EventSource(`${API_BASE}/events`);
    
    this.eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        onMessage(data);
      } catch (error) {
        console.error('Error parsing SSE data:', error);
      }
    };
    
    this.eventSource.onerror = (error) => {
      console.error('SSE connection error:', error);
      if (onError) onError(error);
    };
    
    return this.eventSource;
  }
  
  // WebSocket for real-time updates
  connectWebSocket(onMessage, onError, onClose) {
    const wsUrl = API_BASE.replace('http', 'ws') + '/ws';
    this.websocket = new WebSocket(wsUrl);
    
    this.websocket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        onMessage(data);
      } catch (error) {
        console.error('Error parsing WebSocket data:', error);
      }
    };
    
    this.websocket.onerror = (error) => {
      console.error('WebSocket connection error:', error);
      if (onError) onError(error);
    };
    
    this.websocket.onclose = (event) => {
      console.log('WebSocket connection closed:', event);
      if (onClose) onClose(event);
    };
    
    return this.websocket;
  }
  
  // Close connections
  disconnect() {
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
    
    if (this.websocket) {
      this.websocket.close();
      this.websocket = null;
    }
  }
}

// Export a single instance for real-time updates
export const realtimeAPI = new RealtimeAPI();

// Legacy function for backward compatibility
export async function hello() {
  return apiRequest('/hello');
}
