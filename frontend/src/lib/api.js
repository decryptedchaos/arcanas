/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

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

// Storage Pools API
export const storageAPI = {
  // Get all storage pools
  getPools: () => apiRequest('/storage-pools'),

  // Create storage pool
  createPool: (poolData) => apiRequest('/storage-pools', {
    method: 'POST',
    body: JSON.stringify(poolData),
  }),

  // Update storage pool
  updatePool: (poolName, poolData) => apiRequest(`/storage-pools/${poolName}`, {
    method: 'PUT',
    body: JSON.stringify(poolData),
  }),

  // Delete storage pool
  deletePool: (poolName) => apiRequest(`/storage-pools/${poolName}`, {
    method: 'DELETE',
  }),

  // Cleanup legacy pool (from /var/lib/arcanas/)
  // TODO: Remove this method after migration period (v1.0.0 or later)
  // DEPRECATED: This is temporary migration helper code
  cleanupLegacyPool: (poolName) => apiRequest(`/storage-pools/cleanup/${poolName}`, {
    method: 'DELETE',
  }),

  // Format disk
  formatDisk: (diskData) => apiRequest('/disk/format', {
    method: 'POST',
    body: JSON.stringify(diskData),
  }),
};

// Disk Storage API
export const diskAPI = {
  // TODO: Rename this function - it returns disk info, not stats
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

  // RAID arrays
  getRAIDArrays: () => apiRequest('/raid-arrays'),

  // Create RAID array
  createRAIDArray: (raidData) => apiRequest('/raid-arrays', {
    method: 'POST',
    body: JSON.stringify(raidData),
  }),

  // Update RAID array
  updateRAIDArray: (raidId, raidData) => apiRequest(`/raid-arrays/${raidId}`, {
    method: 'PUT',
    body: JSON.stringify(raidData),
  }),

  // Delete RAID array
  deleteRAIDArray: (raidId) => apiRequest(`/raid-arrays/${raidId}`, {
    method: 'DELETE',
  }),

  // Get RAID array status
  getRAIDStatus: (raidId) => apiRequest(`/raid-arrays/${raidId}/status`),
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
  updateShare: (shareName, shareData) => apiRequest(`/samba-shares/${shareName}/`, {
    method: 'PUT',
    body: JSON.stringify(shareData),
  }),

  // Delete Samba share
  deleteShare: (shareName) => apiRequest(`/samba-shares/`, {
    method: 'DELETE',
    body: JSON.stringify({ name: shareName }),
  }),

  // Toggle share availability
  toggleShare: (shareName) => apiRequest(`/samba-shares/`, {
    method: 'POST',
    body: JSON.stringify({ name: shareName }),
  }),

  // Get share connections
  getConnections: (shareId) => apiRequest(`/samba-shares/${shareId}/connections`),

  // Test Samba configuration
  testConfig: (configData) => apiRequest('/samba-shares/test', {
    method: 'POST',
    body: JSON.stringify(configData),
  }),

  // User management
  getUsers: () => apiRequest('/users'),
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
  updateExport: (exportPath, exportData) => apiRequest(`/nfs-exports/?path=${encodeURIComponent(exportPath)}`, {
    method: 'PUT',
    body: JSON.stringify(exportData),
  }),

  // Delete NFS export
  async deleteExport(exportPath) {
    return apiRequest(`/nfs-exports/?path=${encodeURIComponent(exportPath)}`, {
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

  // Get disk I/O rates (physical disks only, excludes md devices)
  getDiskIORates: () => apiRequest('/system/disk-io'),

  // Get array I/O rates (RAID arrays, actual data throughput)
  getArrayIORates: () => apiRequest('/system/array-io'),

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

  // Update user services
  updateUserServices: (username, services) => apiRequest(`/users/${username}/services`, {
    method: 'PUT',
    body: JSON.stringify({ services }),
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

// Authentication API
export const authAPI = {
  // Login with username and password
  login: (username, password) => apiRequest('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  }),

  // Logout (clears session cookie)
  logout: () => apiRequest('/auth/logout', {
    method: 'POST',
  }),

  // Validate current session
  validate: () => apiRequest('/auth/validate'),
};

// SMART monitoring API
export const smartAPI = {
  // Get SMART status for all disks
  getAllStatus: () => apiRequest('/smart/status'),

  // Get SMART status for a specific disk
  getStatus: (disk) => apiRequest(`/smart/status?disk=${disk}`),

  // Run SMART self-test
  runTest: (disk, testType) => apiRequest('/smart/test', {
    method: 'POST',
    body: JSON.stringify({ disk, test_type: testType }),
  }),

  // Get detailed SMART attributes for a disk
  getAttributes: (disk) => apiRequest(`/smart/attributes?disk=${disk}`),

  // Get error log for a disk
  getErrors: (disk) => apiRequest(`/smart/errors?disk=${disk}`),

  // Modify SMART settings
  updateSetting: (disk, setting, value) => apiRequest('/smart/setting', {
    method: 'PUT',
    body: JSON.stringify({ disk, setting, value }),
  }),
};

// Legacy function for backward compatibility
export async function hello() {
  return apiRequest('/hello');
}
