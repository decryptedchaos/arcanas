<script>
  import { onMount } from 'svelte';
  import { diskAPI } from '$lib/api.js';
  
  let diskStats = [];
  let loading = true;
  let error = null;
  let selectedDisk = null;
  let expandedDisks = new Set();
  let activeTab = 'disks'; // 'disks', 'raid', 'pools', 'zfs'

  $: activeTabClass = (tab) => {
    return activeTab === tab 
      ? 'border-blue-500 text-blue-600' 
      : 'border-transparent text-gray-500 dark:text-gray-400 dark:text-gray-400 hover:text-gray-700 dark:text-gray-300 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600';
  }

  async function loadDiskStats() {
    try {
      error = null;
      const newStats = await diskAPI.getDiskStats();
      
      // Only update if data actually changed to prevent flashing
      if (JSON.stringify(newStats) !== JSON.stringify(diskStats)) {
        diskStats = newStats;
      }
    } catch (err) {
      error = err.message;
      console.error('Failed to load disk stats:', err);
    } finally {
      if (loading) {
        loading = false;
      }
    }
  }

  function toggleDiskExpanded(device) {
    if (expandedDisks.has(device)) {
      expandedDisks.delete(device);
    } else {
      expandedDisks.add(device);
    }
    expandedDisks = expandedDisks; // Trigger reactivity
  }

  function switchTab(tab) {
    activeTab = tab;
  }

  function createRaidArray() {
    alert('RAID creation functionality coming soon! This will open a dialog to configure RAID levels, select disks, and create the array.');
  }

  function createStoragePool() {
    alert('Storage pool creation coming soon! This will configure JBOD/MergerFS pooling.');
  }

  function createZfsPool() {
    alert('ZFS pool creation coming soon! This will configure ZFS pools with advanced features.');
  }

  onMount(() => {
    loadDiskStats();
    // Refresh stats every 10 seconds
    const interval = setInterval(loadDiskStats, 10000);
    return () => clearInterval(interval);
  });
  
  function getStatusColor(status) {
    switch (status) {
      case 'healthy': return 'text-green-600 bg-green-100';
      case 'warning': return 'text-yellow-600 bg-yellow-100';
      case 'critical': return 'text-red-600 bg-red-100';
      default: return 'text-gray-600 dark:text-gray-300 bg-gray-100';
    }
  }
  
  function getUsageColor(percentage) {
    if (percentage >= 90) return 'bg-red-500';
    if (percentage >= 75) return 'bg-yellow-500';
    return 'bg-green-500';
  }

  function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }
</script>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div>
      <h2 class="text-xl font-bold text-gray-900 dark:text-white">Arcanas Storage Management</h2>
      <p class="text-sm text-gray-600 dark:text-gray-300 mt-1">Manage disks, RAID arrays, and storage pools</p>
    </div>
    <button 
      on:click={loadDiskStats}
      class="btn btn-primary"
      disabled={loading}
    >
      {loading ? 'Loading...' : 'Refresh'}
    </button>
  </div>

  <!-- Tab Navigation -->
  <div class="border-b border-gray-200 dark:border-gray-700">
    <nav class="-mb-px flex space-x-8">
      <button
        on:click={() => switchTab('disks')}
        class="py-2 px-1 border-b-2 font-medium text-sm transition-colors duration-200 {activeTabClass('disks')}"
      >
        Disks
      </button>
      <button
        on:click={() => switchTab('raid')}
        class="py-2 px-1 border-b-2 font-medium text-sm transition-colors duration-200 {activeTabClass('raid')}"
      >
        RAID Arrays
      </button>
      <button
        on:click={() => switchTab('pools')}
        class="py-2 px-1 border-b-2 font-medium text-sm transition-colors duration-200 {activeTabClass('pools')}"
      >
        Storage Pools
      </button>
      <button
        on:click={() => switchTab('zfs')}
        class="py-2 px-1 border-b-2 font-medium text-sm transition-colors duration-200 {activeTabClass('zfs')}"
      >
        ZFS Pools
      </button>
    </nav>
  </div>

  <!-- Error State -->
  {#if error}
    <div class="bg-red-50 border border-red-200 rounded-md p-4">
      <div class="flex">
        <div class="flex-shrink-0">
          <svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
          </svg>
        </div>
        <div class="ml-3">
          <h3 class="text-sm font-medium text-red-800">Error loading disk stats</h3>
          <div class="mt-2 text-sm text-red-700">{error}</div>
        </div>
      </div>
    </div>
  {/if}

  <!-- Loading State -->
  {#if loading && !diskStats.length}
    <div class="text-center py-8">
      <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      <p class="mt-2 text-sm text-gray-600 dark:text-gray-300">Loading storage information...</p>
    </div>
  {/if}

  <!-- Tab Content -->
  {#if activeTab === 'disks'}
    <!-- Disks Tab Content -->
    <!-- Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <div class="stat-card">
        <div class="flex items-center">
          <div class="p-3 bg-blue-100 rounded-lg mr-4">
            <svg class="h-6 w-6 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
            </svg>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-600 dark:text-gray-300">Total Storage</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{diskStats.length} disks</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">Connected</p>
          </div>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="flex items-center">
          <div class="p-3 bg-green-100 rounded-lg mr-4">
            <svg class="h-6 w-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-600 dark:text-gray-300">Available</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{diskStats.filter(d => d.smart.status === 'healthy').length}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">Healthy disks</p>
          </div>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="flex items-center">
          <div class="p-3 bg-yellow-100 rounded-lg mr-4">
            <svg class="h-6 w-6 text-yellow-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-600 dark:text-gray-300">Health Alerts</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{diskStats.filter(d => d.smart.status !== 'healthy').length}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">Needs attention</p>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Disk List -->
    <div class="space-y-3">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Disk Details</h3>
      
      {#each diskStats as disk}
        <div class="card hover:shadow-lg transition-shadow">
          <!-- Compact Header (Always Visible) -->
          <div 
            class="flex items-center justify-between p-4 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700"
            on:click={() => toggleDiskExpanded(disk.device)}
            on:keydown={(e) => e.key === 'Enter' || e.key === ' ' ? toggleDiskExpanded(disk.device) : null}
            role="button"
            tabindex="0"
            aria-label="Toggle details for {disk.device}"
            aria-expanded={expandedDisks.has(disk.device)}
          >
            <div class="flex items-center space-x-4 flex-1">
              <!-- Expand/Collapse Icon -->
              <div class="transition-transform duration-200" class:rotate-90={expandedDisks.has(disk.device)}>
                <svg class="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                </svg>
              </div>
              
              <!-- Device and Status -->
              <div class="flex-1">
                <div class="flex items-center space-x-3">
                  <h4 class="text-lg font-semibold text-gray-900 dark:text-white">{disk.device}</h4>
                  <span class="px-2 py-1 text-xs font-medium rounded-full {getStatusColor(disk.smart.status)}">
                    {disk.smart.status}
                  </span>
                </div>
                <p class="text-sm text-gray-600 dark:text-gray-300">{disk.model}</p>
              </div>
              
              <!-- Usage Summary -->
              <div class="text-right">
                <div class="text-lg font-semibold text-gray-900 dark:text-white">{disk.usage.toFixed(1)}%</div>
                <div class="text-sm text-gray-600 dark:text-gray-300">{formatBytes(disk.used)} / {formatBytes(disk.size)}</div>
              </div>
            </div>
          </div>
          
          <!-- Expanded Details (Collapsible) -->
          {#if expandedDisks.has(disk.device)}
            <div class="border-t border-gray-200 dark:border-gray-700 p-4 space-y-4">
              <!-- Storage Type Selection -->
              <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div>
                  <label for="storage-type-{disk.device}" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Storage Type</label>
                  <select id="storage-type-{disk.device}" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 text-gray-900 dark:text-white">
                    <option value="jbod">JBOD (Independent)</option>
                    <option value="raid">RAID Array</option>
                    <option value="zfs">ZFS Pool</option>
                  </select>
                </div>
                
                <div>
                  <label for="mount-point-{disk.device}" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Mount Point</label>
                  <input 
                    id="mount-point-{disk.device}"
                    type="text" 
                    value={disk.mountpoint || ''} 
                    placeholder="e.g., /data"
                    class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 text-gray-900 dark:text-white"
                  />
                </div>
                
                <div>
                  <label for="filesystem-{disk.device}" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Filesystem</label>
                  <select id="filesystem-{disk.device}" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 text-gray-900 dark:text-white">
                    <option value="ext4" selected={disk.filesystem === 'ext4'}>ext4</option>
                    <option value="xfs" selected={disk.filesystem === 'xfs'}>XFS</option>
                    <option value="btrfs" selected={disk.filesystem === 'btrfs'}>Btrfs</option>
                    <option value="zfs" selected={disk.filesystem === 'zfs'}>ZFS</option>
                    <option value="mergerfs" selected={disk.filesystem === 'mergerfs'}>MergerFS</option>
                  </select>
                </div>
              </div>
              
              <!-- Usage Bar -->
              <div>
                <div class="flex items-center justify-between text-sm mb-2">
                  <span class="text-gray-600 dark:text-gray-300">Usage: {formatBytes(disk.used)} / {formatBytes(disk.size)}</span>
                  <span class="font-medium text-gray-900 dark:text-white">{disk.usage.toFixed(1)}%</span>
                </div>
                <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3">
                  <div 
                    class="{getUsageColor(disk.usage)} h-3 rounded-full transition-all duration-500" 
                    style="width: {disk.usage}%"
                  ></div>
                </div>
              </div>
              
              <!-- Details Grid -->
              <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                <div>
                  <p class="text-gray-600 dark:text-gray-300">Temperature</p>
                  <p class="font-medium text-gray-900 dark:text-white">{disk.smart.temperature}°C</p>
                </div>
                <div>
                  <p class="text-gray-600 dark:text-gray-300">Health</p>
                  <p class="font-medium text-gray-900 dark:text-white">{disk.smart.health}%</p>
                </div>
                <div>
                  <p class="text-gray-600 dark:text-gray-300">Available</p>
                  <p class="font-medium text-gray-900 dark:text-white">{formatBytes(disk.available)}</p>
                </div>
                <div>
                  <p class="text-gray-600 dark:text-gray-300">Read-Only</p>
                  <p class="font-medium text-gray-900 dark:text-white">{disk.read_only ? 'Yes' : 'No'}</p>
                </div>
              </div>
              
              <!-- Action Buttons -->
              <div class="flex space-x-2 pt-2">
                <button class="btn btn-primary">Apply Changes</button>
                <button class="btn btn-secondary">Format Disk</button>
                <button class="btn btn-secondary">SMART Test</button>
                <button class="btn btn-secondary">Unmount</button>
              </div>
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {:else if activeTab === 'raid'}
    <!-- RAID Arrays Tab Content -->
    <div class="text-center py-12">
      <div class="mx-auto max-w-md">
        <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
        <h3 class="mt-2 text-sm font-medium text-gray-900 dark:text-white">RAID Management</h3>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Create and manage RAID arrays</p>
        <div class="mt-6">
          <button class="btn btn-primary" on:click={createRaidArray}>Create RAID Array</button>
        </div>
      </div>
    </div>
  {:else if activeTab === 'pools'}
    <!-- Storage Pools Tab Content -->
    <div class="text-center py-12">
      <div class="mx-auto max-w-md">
        <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
        </svg>
        <h3 class="mt-2 text-sm font-medium text-gray-900 dark:text-white">Storage Pools</h3>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">JBOD and MergerFS pooling</p>
        <div class="mt-6">
          <button class="btn btn-primary" on:click={createStoragePool}>Create Storage Pool</button>
        </div>
      </div>
    </div>
  {:else if activeTab === 'zfs'}
    <!-- ZFS Pools Tab Content -->
    <div class="text-center py-12">
      <div class="mx-auto max-w-md">
        <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
        </svg>
        <h3 class="mt-2 text-sm font-medium text-gray-900 dark:text-white">ZFS Pools</h3>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Advanced ZFS pool management</p>
        <div class="mt-6">
          <button class="btn btn-primary" on:click={createZfsPool}>Create ZFS Pool</button>
        </div>
      </div>
    </div>
  {/if}
</div>