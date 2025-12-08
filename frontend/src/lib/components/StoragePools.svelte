<script>
  import { onMount } from 'svelte';
  
  let storagePools = [];
  let loading = true;
  let showCreateModal = false;
  let availableDisks = [];

  $: safeStoragePools = storagePools || [];
  $: safeAvailableDisks = availableDisks || [];
  
  let createForm = {
    name: '',
    type: 'mergerfs',
    devices: [],
    config: 'defaults'
  };

  function loadStoragePools() {
    loading = true;
    fetch('/api/storage-pools')
      .then(response => response.json())
      .then(data => {
        storagePools = data;
        loading = false;
      })
      .catch(error => {
        console.error('Error loading storage pools:', error);
        loading = false;
      });
  }

  function loadAvailableDisks() {
    fetch('/api/disk-stats')
      .then(response => response.json())
      .then(data => {
        availableDisks = data || [];
        console.log('Available disks for pools:', availableDisks);
      })
      .catch(error => console.error('Error loading disks:', error));
  }

  function createStoragePool() {
    fetch('/api/storage-pools', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(createForm)
    })
    .then(response => {
      if (response.ok) {
        showCreateModal = false;
        createForm = { name: '', type: 'mergerfs', devices: [], config: 'defaults' };
        loadStoragePools();
        loadAvailableDisks();
      } else {
        return response.text().then(text => {
          console.error('Server error:', text);
          throw new Error(`Failed to create storage pool: ${text}`);
        });
      }
    })
    .catch(error => {
      console.error('Error creating storage pool:', error);
      alert('Failed to create storage pool');
    });
  }

  function deleteStoragePool(poolName) {
    if (!confirm(`Are you sure you want to delete storage pool ${poolName}?`)) {
      return;
    }

    fetch(`/api/storage-pools/${poolName}`, {
      method: 'DELETE'
    })
    .then(response => {
      if (response.ok) {
        loadStoragePools();
        loadAvailableDisks();
      } else {
        throw new Error('Failed to delete storage pool');
      }
    })
    .catch(error => {
      console.error('Error deleting storage pool:', error);
      alert('Failed to delete storage pool');
    });
  }

  function getPoolTypeDescription(type) {
    const descriptions = {
      'mergerfs': 'JBOD pool - Disk failure loses ONLY data on that disk, other disks remain safe. 100% space utilization.',
      'lvm': 'High performance striping - One disk failure loses ALL data. Use only for temporary/cached data.',
      'bind': 'Simple bind mount - mounts a single device to a directory',
      'jbod': 'Just a Bunch Of Disks - simple disk concatenation'
    };
    return descriptions[type] || type;
  }

  function getStateColor(state) {
    switch (state) {
      case 'active': return 'text-green-600';
      case 'inactive': return 'text-gray-600';
      case 'error': return 'text-red-600';
      default: return 'text-gray-600';
    }
  }

  onMount(() => {
    loadStoragePools();
    loadAvailableDisks();
  });
</script>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div>
      <h2 class="text-xl font-bold text-gray-900 dark:text-white">Storage Pools</h2>
      <p class="text-sm text-gray-600 dark:text-gray-300">JBOD and MergerFS pooling</p>
    </div>
    <button 
      class="btn btn-primary"
      on:click={() => { showCreateModal = true; loadAvailableDisks(); }}
    >
      Create Storage Pool
    </button>
  </div>

  <!-- Loading State -->
  {#if loading}
    <div class="text-center py-8">
      <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      <p class="mt-2 text-sm text-gray-600 dark:text-gray-300">Loading storage pools...</p>
    </div>
  {/if}

  <!-- Storage Pools List -->
  {#if !loading && safeStoragePools.length === 0}
    <div class="text-center py-12">
      <div class="mx-auto max-w-md">
        <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
        </svg>
        <h3 class="mt-2 text-sm font-medium text-gray-900 dark:text-white">No storage pools</h3>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Create your first storage pool to get started</p>
      </div>
    </div>
  {/if}

  <!-- Storage Pool Cards -->
  {#if !loading}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {#each storagePools as pool}
        <div class="card">
          <div class="flex items-center justify-between mb-4">
            <div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{pool.name}</h3>
              <p class="text-sm text-gray-600 dark:text-gray-300">{pool.type.toUpperCase()}</p>
            </div>
            <div class="flex items-center space-x-2">
              <span class="px-2 py-1 text-xs font-medium rounded-full {pool.state === 'active' ? 'bg-green-100 text-green-600' : 'bg-gray-100 text-gray-600'}">
                {pool.state}
              </span>
            </div>
          </div>

          <div class="space-y-3">
            <div class="flex justify-between text-sm">
              <span class="text-gray-600 dark:text-gray-300">Size</span>
              <span class="font-medium text-gray-900 dark:text-white">{(pool.size / 1024 / 1024 / 1024).toFixed(1)} GB</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-600 dark:text-gray-300">Used</span>
              <span class="font-medium text-gray-900 dark:text-white">{(pool.used / 1024 / 1024 / 1024).toFixed(1)} GB</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-600 dark:text-gray-300">Devices</span>
              <span class="font-medium text-gray-900 dark:text-white">{pool.devices.length}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-600 dark:text-gray-300">Mount Point</span>
              <span class="font-medium text-gray-900 dark:text-white">{pool.mount_point || 'Not mounted'}</span>
            </div>
          </div>

          <!-- Usage Bar -->
          <div class="mt-4">
            <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
              <div 
                class="bg-purple-600 h-2 rounded-full" 
                style="width: {pool.size ? (pool.used / pool.size * 100) : 0}%"
              ></div>
            </div>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              {pool.size ? ((pool.used / pool.size) * 100).toFixed(1) : 0}% used
            </p>
          </div>

          <div class="mt-4 flex space-x-2">
            <button class="btn btn-secondary btn-sm">
              Configure
            </button>
            <button 
              class="btn btn-danger btn-sm"
              on:click={() => deleteStoragePool(pool.name)}
            >
              Delete
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}

  <!-- Create Storage Pool Modal -->
  {#if showCreateModal}
    <div class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div class="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white dark:bg-gray-800">
        <div class="mt-3">
          <h3 class="text-lg font-medium text-gray-900 dark:text-white">Create Storage Pool</h3>
          <div class="mt-4 space-y-4">
            <div>
              <label for="pool-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Name</label>
              <input 
                id="pool-name"
                type="text" 
                bind:value={createForm.name}
                class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600"
                placeholder="Enter storage pool name"
              />
            </div>
            <div>
              <label for="pool-type" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Pool Type</label>
              <select 
                id="pool-type"
                bind:value={createForm.type}
                class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600">
                <option value="mergerfs">MergerFS (Recommended - No data loss on disk failure)</option>
                <option value="lvm">LVM (High Performance - Risk of total data loss)</option>
                <option value="bind">Bind Mount (Simple)</option>
                <option value="jbod">JBOD (Just a Bunch Of Disks)</option>
              </select>
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{getPoolTypeDescription(createForm.type)}</p>
            </div>
            <div>
              <label for="pool-devices" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Devices</label>
              <div id="pool-devices" class="mt-2 space-y-2 max-h-32 overflow-y-auto" role="group" aria-label="Available devices for storage pool">
                {#each safeAvailableDisks as disk}
                  <label class="flex items-center">
                    <input 
                      type="checkbox" 
                      bind:group={createForm.devices}
                      value={disk.device}
                      class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                    />
                    <span class="ml-2 text-sm text-gray-700 dark:text-gray-300">
                      {disk.device} ({(disk.size / 1024 / 1024 / 1024).toFixed(1)} GB)
                    </span>
                  </label>
                {/each}
              </div>
            </div>
            <div>
              <label for="pool-config" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Configuration</label>
              <input 
                id="pool-config"
                type="text" 
                bind:value={createForm.config}
                class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600"
                placeholder="e.g., defaults,allow_other"
              />
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">MergerFS mount options (e.g., defaults, category.create=mfs)</p>
            </div>
          </div>
          <div class="mt-6 flex space-x-3">
            <button 
              class="btn btn-primary"
              on:click={createStoragePool}
              disabled={createForm.name === '' || createForm.devices.length === 0}
            >
              Create
            </button>
            <button 
              class="btn btn-secondary"
              on:click={() => { showCreateModal = false; createForm = { name: '', type: 'mergerfs', devices: [], config: 'defaults' }; }}
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>
  {/if}
</div>
