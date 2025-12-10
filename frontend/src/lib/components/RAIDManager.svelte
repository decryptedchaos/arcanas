<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { onMount } from 'svelte';
  
  let raidArrays = [];
  let loading = true;
  let showCreateModal = false;
  let showAddDiskModal = false;
  let selectedArray = null;
  let availableDisks = [];

  $: safeRaidArrays = raidArrays || [];
  $: safeAvailableDisks = availableDisks || [];
  
  let createForm = {
    name: '',
    level: 'raid1',
    devices: []
  };
  
  let addDiskForm = {
    device: ''
  };

  function loadRAIDArrays() {
    loading = true;
    fetch('/api/raid-arrays')
      .then(response => response.json())
      .then(data => {
        raidArrays = data;
        loading = false;
      })
      .catch(error => {
        console.error('Error loading RAID arrays:', error);
        loading = false;
      });
  }

  function loadAvailableDisks() {
    fetch('/api/disk-stats')
      .then(response => response.json())
      .then(data => {
        availableDisks = data || [];
        console.log('Available disks:', availableDisks);
      })
      .catch(error => console.error('Error loading disks:', error));
  }

  function createRAIDArray() {
    fetch('/api/raid-arrays', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(createForm)
    })
    .then(response => {
      if (response.ok) {
        showCreateModal = false;
        createForm = { name: '', level: 'raid1', devices: [] };
        loadRAIDArrays();
        loadAvailableDisks();
      } else {
        throw new Error('Failed to create RAID array');
      }
    })
    .catch(error => {
      console.error('Error creating RAID array:', error);
      alert('Failed to create RAID array');
    });
  }

  function deleteRAIDArray(arrayName) {
    if (!confirm(`Are you sure you want to delete RAID array ${arrayName}?`)) {
      return;
    }

    fetch(`/api/raid-arrays/${arrayName}`, {
      method: 'DELETE'
    })
    .then(response => {
      if (response.ok) {
        loadRAIDArrays();
        loadAvailableDisks();
      } else {
        throw new Error('Failed to delete RAID array');
      }
    })
    .catch(error => {
      console.error('Error deleting RAID array:', error);
      alert('Failed to delete RAID array');
    });
  }

  function addDiskToArray() {
    fetch(`/api/raid-arrays/${selectedArray.name}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ device: addDiskForm.device })
    })
    .then(response => {
      if (response.ok) {
        showAddDiskModal = false;
        selectedArray = null;
        addDiskForm = { device: '' };
        loadRAIDArrays();
        loadAvailableDisks();
      } else {
        throw new Error('Failed to add disk to RAID array');
      }
    })
    .catch(error => {
      console.error('Error adding disk to RAID array:', error);
      alert('Failed to add disk to RAID array');
    });
  }

  function getRAIDLevelDescription(level) {
    const descriptions = {
      'raid0': 'Striping - No redundancy, maximum performance',
      'raid1': 'Mirroring - Full redundancy, 50% capacity',
      'raid5': 'Striping with parity - Good redundancy, good performance',
      'raid6': 'Striping with double parity - High redundancy',
      'raid10': 'Striped mirrors - Best performance and redundancy'
    };
    return descriptions[level] || level;
  }

  function getStateColor(state) {
    switch (state) {
      case 'active': return 'text-green-600';
      case 'degraded': return 'text-yellow-600';
      case 'failed': return 'text-red-600';
      case 'syncing': return 'text-blue-600';
      default: return 'text-gray-600';
    }
  }

  function getHealthColor(health) {
    if (health >= 90) return 'text-green-600';
    if (health >= 70) return 'text-yellow-600';
    return 'text-red-600';
  }

  onMount(() => {
    loadRAIDArrays();
    loadAvailableDisks();
  });
</script>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div>
      <h2 class="text-xl font-bold text-gray-900 dark:text-white">RAID Arrays</h2>
      <p class="text-sm text-gray-600 dark:text-gray-300">Manage RAID storage arrays</p>
    </div>
    <button 
      class="btn btn-primary"
      on:click={() => { showCreateModal = true; loadAvailableDisks(); }}
    >
      Create RAID Array
    </button>
  </div>

  <!-- Loading State -->
  {#if loading}
    <div class="text-center py-8">
      <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      <p class="mt-2 text-sm text-gray-600 dark:text-gray-300">Loading RAID arrays...</p>
    </div>
  {/if}

  <!-- RAID Arrays List -->
  {#if !loading && safeRaidArrays.length === 0}
    <div class="text-center py-12">
      <div class="mx-auto max-w-md">
        <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
        <h3 class="mt-2 text-sm font-medium text-gray-900 dark:text-white">No RAID arrays</h3>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Create your first RAID array to get started</p>
      </div>
    </div>
  {/if}

  <!-- RAID Array Cards -->
  {#if !loading}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {#each raidArrays as array}
        <div class="card">
          <div class="flex items-center justify-between mb-4">
            <div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{array.name}</h3>
              <p class="text-sm text-gray-600 dark:text-gray-300">{array.level.toUpperCase()}</p>
            </div>
            <div class="flex items-center space-x-2">
              <span class="px-2 py-1 text-xs font-medium rounded-full {array.state === 'active' ? 'bg-green-100 text-green-600' : 'bg-yellow-100 text-yellow-600'}">
                {array.state}
              </span>
            </div>
          </div>

          <div class="space-y-3">
            <div class="flex justify-between text-sm">
              <span class="text-gray-600 dark:text-gray-300">Size</span>
              <span class="font-medium text-gray-900 dark:text-white">{(array.size / 1024 / 1024 / 1024).toFixed(1)} GB</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-600 dark:text-gray-300">Used</span>
              <span class="font-medium text-gray-900 dark:text-white">{(array.used / 1024 / 1024 / 1024).toFixed(1)} GB</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-600 dark:text-gray-300">Health</span>
              <span class="font-medium {getHealthColor(array.health)}">{array.health}%</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-600 dark:text-gray-300">Devices</span>
              <span class="font-medium text-gray-900 dark:text-white">{array.devices.length}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-600 dark:text-gray-300">Mount Point</span>
              <span class="font-medium text-gray-900 dark:text-white">{array.mount_point || 'Not mounted'}</span>
            </div>
          </div>

          <div class="mt-4 flex space-x-2">
            <button 
              class="btn btn-secondary btn-sm"
              on:click={() => { selectedArray = array; showAddDiskModal = true; loadAvailableDisks(); }}
            >
              Add Disk
            </button>
            <button 
              class="btn btn-danger btn-sm"
              on:click={() => deleteRAIDArray(array.name)}
            >
              Delete
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}

  <!-- Create RAID Modal -->
  {#if showCreateModal}
    <div class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div class="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white dark:bg-gray-800">
        <div class="mt-3">
          <h3 class="text-lg font-medium text-gray-900 dark:text-white">Create RAID Array</h3>
          <div class="mt-4 space-y-4">
            <div>
              <label for="raid-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Name</label>
              <input 
                id="raid-name"
                type="text" 
                bind:value={createForm.name}
                class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600"
                placeholder="Enter RAID array name"
              />
            </div>
            <div>
              <label for="raid-level" class="block text-sm font-medium text-gray-700 dark:text-gray-300">RAID Level</label>
              <select 
                id="raid-level"
                bind:value={createForm.level}
                class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600">
                <option value="raid0">RAID 0 - Striping</option>
                <option value="raid1">RAID 1 - Mirroring</option>
                <option value="raid5">RAID 5 - Parity</option>
                <option value="raid6">RAID 6 - Double Parity</option>
                <option value="raid10">RAID 10 - Striped Mirrors</option>
              </select>
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{getRAIDLevelDescription(createForm.level)}</p>
            </div>
            <div>
              <label for="raid-devices" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Devices</label>
              <div id="raid-devices" class="mt-2 space-y-2 max-h-32 overflow-y-auto" role="group" aria-label="Available devices for RAID">
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
          </div>
          <div class="mt-6 flex space-x-3">
            <button 
              class="btn btn-primary"
              on:click={createRAIDArray}
              disabled={createForm.name === '' || createForm.devices.length === 0}
            >
              Create RAID Array
            </button>
            <button 
              class="btn btn-secondary"
              on:click={() => { showCreateModal = false; createForm = { name: '', level: 'raid1', devices: [] }; }}
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>
  {/if}

  <!-- Add Disk Modal -->
  {#if showAddDiskModal && selectedArray}
    <div class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div class="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white dark:bg-gray-800">
        <div class="mt-3">
          <h3 class="text-lg font-medium text-gray-900 dark:text-white">Add Disk to {selectedArray.name}</h3>
          <div class="mt-4 space-y-4">
            <div>
              <label for="add-disk-device" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Device</label>
              <select 
                id="add-disk-device"
                bind:value={addDiskForm.device}
                class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600">
                <option value="">Select a device...</option>
                {#each safeAvailableDisks as disk}
                  <option value={disk.device}>
                    {disk.device} ({(disk.size / 1024 / 1024 / 1024).toFixed(1)} GB)
                  </option>
                {/each}
              </select>
            </div>
          </div>
          <div class="mt-6 flex space-x-3">
            <button 
              class="btn btn-primary"
              on:click={addDiskToArray}
              disabled={addDiskForm.device === ''}
            >
              Add Disk
            </button>
            <button 
              class="btn btn-secondary"
              on:click={() => { showAddDiskModal = false; selectedArray = null; addDiskForm = { device: '' }; }}
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>
  {/if}
</div>
