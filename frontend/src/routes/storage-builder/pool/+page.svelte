<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { diskAPI, storageAPI, lvmAPI } from '$lib/api.js';

  let availableDevices = [];
  let logicalVolumes = [];
  let loading = true;
  let error = null;
  let creating = false;

  let poolConfig = {
    name: '',
    type: 'bind',
    devices: [],
    filesystem: 'ext4',
    config: ''
  };

  // Pool types
  const poolTypes = [
    { value: 'bind', label: 'Bind Mount', description: 'Mount a single device directly', icon: 'link' },
    { value: 'mergerfs', label: 'MergerFS', description: 'Pool multiple devices together', icon: 'layers' },
    { value: 'lvm', label: 'LVM Logical Volume', description: 'Use an existing LVM logical volume', icon: 'cube' }
  ];

  async function loadAvailableResources() {
    try {
      loading = true;

      // Load physical disks and RAID arrays
      const diskStats = await diskAPI.getDiskStats();
      const raidArrays = await diskAPI.getRAIDArrays();

      // Filter available disks
      const availableDisks = diskStats.filter((disk) => {
        return disk.available !== false && !disk.mountpoint;
      });

      // Filter out RAID members
      const raidMemberPaths = raidArrays.flatMap((array) => array.devices || []);
      const availableDisksFiltered = availableDisks.filter((disk) => {
        return !raidMemberPaths.includes(disk.path);
      });

      // Load LVM logical volumes
      const vgs = await lvmAPI.getVolumeGroups();
      const lvs = await lvmAPI.getLogicalVolumes();

      // Filter unmounted LVs
      logicalVolumes = lvs.filter(lv => !lv.mounted);

      // Combine all available devices
      availableDevices = [
        ...availableDisksFiltered.map(d => ({
          type: 'disk',
          path: d.path,
          name: d.name || d.path,
          size: d.size,
          model: d.model,
          filesystem: d.filesystem
        })),
        ...raidArrays.map(r => ({
          type: 'raid',
          path: r.device,
          name: r.name,
          size: r.size,
          devices: r.devices?.length || 0
        }))
      ];
    } catch (err) {
      error = err.message || 'Failed to load available resources';
      console.error('Failed to load available resources:', err);
    } finally {
      loading = false;
    }
  }

  function toggleDevice(devicePath) {
    if (poolConfig.devices.includes(devicePath)) {
      poolConfig.devices = poolConfig.devices.filter(d => d !== devicePath);
    } else {
      poolConfig.devices = [...poolConfig.devices, devicePath];
    }
  }

  function toggleLV(lvPath) {
    if (poolConfig.devices.includes(lvPath)) {
      poolConfig.devices = poolConfig.devices.filter(d => d !== lvPath);
    } else {
      poolConfig.devices = [...poolConfig.devices, lvPath];
    }
  }

  function cancel() {
    goto('/storage-builder');
  }

  async function createPool() {
    try {
      if (!poolConfig.name) {
        error = 'Please enter a pool name';
        return;
      }

      if (poolConfig.type === 'lvm_lv') {
        if (poolConfig.devices.length === 0) {
          error = 'Please select a logical volume';
          return;
        }
      } else if (poolConfig.devices.length === 0) {
        error = 'Please select at least one device';
        return;
      }

      if (poolConfig.type === 'mergerfs' && poolConfig.devices.length < 2) {
        error = 'MergerFS requires at least 2 devices';
        return;
      }

      creating = true;
      await storageAPI.createPool(poolConfig);

      // Success - redirect to storage page
      goto('/storage?tab=pools');
    } catch (err) {
      error = err.message || 'Failed to create storage pool';
      console.error('Failed to create storage pool:', err);
    } finally {
      creating = false;
    }
  }

  onMount(() => {
    loadAvailableResources();
  });

  function getPoolTypeIcon(type) {
    const icons = {
      bind: '<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" /></svg>',
      mergerfs: '<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" /></svg>',
      lvm: '<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" /></svg>'
    };
    return icons[type] || icons.bind;
  }
</script>

<div class="p-6" role="main" tabindex="-1">
  <!-- Header -->
  <div class="mb-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-gray-900 dark:text-white mb-2">
          Create Storage Pool
        </h1>
        <p class="text-sm text-gray-600 dark:text-gray-300">
          Create a storage pool from devices, RAID arrays, or LVM volumes
        </p>
      </div>
      <button
        on:click={cancel}
        class="text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
        aria-label="Cancel"
        title="Cancel"
      >
        <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
  </div>

  <!-- Error -->
  {#if error}
    <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4 mb-6">
      <p class="text-red-800 dark:text-red-200">{error}</p>
    </div>
  {/if}

  <!-- Loading -->
  {#if loading}
    <div class="flex justify-center items-center py-12">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
    </div>
  {:else if creating}
    <div class="flex justify-center items-center py-12">
      <div class="text-center">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
        <p class="text-gray-600 dark:text-gray-400">Creating storage pool...</p>
      </div>
    </div>
  {:else}
    <!-- Form -->
    <div class="max-w-4xl">
      <div class="bg-white dark:bg-card rounded-lg border border-gray-200 dark:border-border p-6">
        <!-- Pool Type Selection -->
        <div class="mb-6">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
            Pool Type
          </label>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            {#each poolTypes as type}
              <button
                type="button"
                on:click={() => {
                  poolConfig.type = type.value;
                  poolConfig.devices = [];
                }}
                class="p-4 text-left rounded-lg border-2 transition-colors
                  {poolConfig.type === type.value
                    ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                    : 'border-gray-200 dark:border-border hover:border-gray-300 dark:hover:border-border'}"
              >
                <div class="flex items-center space-x-3">
                  <div class="text-blue-600 dark:text-blue-400">
                    {@html getPoolTypeIcon(type.value)}
                  </div>
                  <div>
                    <div class="font-medium text-gray-900 dark:text-white">{type.label}</div>
                    <div class="text-sm text-gray-600 dark:text-gray-400 mt-1">{type.description}</div>
                  </div>
                </div>
              </button>
            {/each}
          </div>
        </div>

        <!-- Pool Name -->
        <div class="mb-6">
          <label for="poolName" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Pool Name
          </label>
          <input
            id="poolName"
            type="text"
            bind:value={poolConfig.name}
            placeholder="my-pool"
            class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-muted dark:text-white"
            required
          />
        </div>

        <!-- Device/LV Selection -->
        {#if poolConfig.type === 'lvm'}
          <!-- LVM Logical Volume Selection -->
          <div class="mb-6">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Select Logical Volume
            </label>
            {#if logicalVolumes.length === 0}
              <div class="p-4 border border-gray-300 dark:border rounded-md bg-gray-50 dark:bg-muted text-gray-500 dark:text-gray-400 text-center">
                No unmounted logical volumes available. Create one in the
                <a href="/lvm" class="text-blue-600 dark:text-blue-400 hover:underline">LVM Volumes</a>
                page first.
              </div>
            {:else}
              <div class="border border-gray-300 dark:border rounded-md p-4 max-h-60 overflow-y-auto bg-gray-50 dark:bg-muted">
                {#each logicalVolumes as lv}
                  <label class="flex items-center space-x-3 p-2 hover:bg-gray-100 dark:hover:bg-muted rounded cursor-pointer">
                    <input
                      type="radio"
                      name="lv-selection"
                      checked={poolConfig.devices.includes(lv.path)}
                      on:change={() => {
                        poolConfig.devices = [lv.path];
                      }}
                      class="border-gray-300 dark:border text-blue-600 focus:ring-blue-500 dark:bg-card"
                    />
                    <div class="flex-1">
                      <div class="font-medium text-gray-900 dark:text-white">
                        {lv.name} ({lv.vg_name})
                      </div>
                      <div class="text-xs text-gray-500 dark:text-gray-400">
                        {lv.path} • {lv.size_gb} GB • {lv.filesystem || 'No filesystem'}
                      </div>
                    </div>
                  </label>
                {/each}
              </div>
            {/if}
          </div>
        {:else}
          <!-- Device Selection -->
          <div class="mb-6">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Select {poolConfig.type === 'mergerfs' ? 'Devices' : 'Device'}
            </label>
            {#if availableDevices.length === 0}
              <div class="p-4 border border-gray-300 dark:border rounded-md bg-gray-50 dark:bg-muted text-gray-500 dark:text-gray-400 text-center">
                No devices available. All devices are either mounted or in use.
              </div>
            {:else}
              <div class="border border-gray-300 dark:border rounded-md p-4 max-h-60 overflow-y-auto bg-gray-50 dark:bg-muted">
                {#each availableDevices as device}
                  <label class="flex items-center space-x-3 p-2 hover:bg-gray-100 dark:hover:bg-muted rounded cursor-pointer">
                    <input
                      type="checkbox"
                      checked={poolConfig.devices.includes(device.path)}
                      on:change={() => toggleDevice(device.path)}
                      class="rounded border-gray-300 dark:border text-blue-600 focus:ring-blue-500 dark:bg-card"
                      disabled={poolConfig.type === 'bind' && poolConfig.devices.length > 0 && poolConfig.devices[0] !== device.path}
                    />
                    <div class="flex-1">
                      <div class="flex items-center space-x-2">
                        <span class="font-medium text-gray-900 dark:text-white">
                          {device.name}
                        </span>
                        <span class="text-xs px-2 py-0.5 rounded bg-gray-200 dark:bg-muted text-gray-700 dark:text-gray-300">
                          {device.type}
                        </span>
                      </div>
                      <div class="text-xs text-gray-500 dark:text-gray-400">
                        {device.path} • {device.size}
                        {#if device.model}
                          • {device.model}
                        {/if}
                        {#if device.filesystem}
                          • {device.filesystem}
                        {/if}
                      </div>
                    </div>
                  </label>
                {/each}
              </div>
              <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
                {poolConfig.devices.length} device(s) selected
                {#if poolConfig.type === 'bind' && poolConfig.devices.length > 1}
                  <span class="text-yellow-600 dark:text-yellow-400">(Bind mount only supports 1 device)</span>
                {/if}
              </p>
            {/if}
          </div>
        {/if}

        <!-- Filesystem -->
        <div class="mb-6">
          <label for="filesystem" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Filesystem
          </label>
          <select
            id="filesystem"
            bind:value={poolConfig.filesystem}
            class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-muted dark:text-white"
          >
            <option value="ext4">ext4 (Recommended)</option>
            <option value="xfs">XFS</option>
            <option value="btrfs">Btrfs</option>
          </select>
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
            {#if poolConfig.type === 'lvm_lv'}
      Note: If the LV already has a filesystem, this will be skipped.
    {:else}
      The selected device(s) will be formatted with this filesystem.
    {/if}
          </p>
        </div>

        <!-- Mount Options (optional) -->
        <div class="mb-6">
          <label for="mountOptions" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Mount Options (optional)
          </label>
          <input
            id="mountOptions"
            type="text"
            bind:value={poolConfig.config}
            placeholder="defaults,allow_other,use_ino"
            class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-muted dark:text-white"
          />
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
            Comma-separated list of mount options (e.g., defaults,allow_other)
          </p>
        </div>

        <!-- Warning -->
        {#if poolConfig.type !== 'lvm_lv'}
          <div class="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-md p-3 mb-6">
            <div class="flex">
              <svg class="w-5 h-5 text-yellow-400 mr-2 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
              </svg>
              <div class="text-sm text-yellow-800 dark:text-yellow-200">
                <strong>Warning:</strong> Creating a storage pool will format the selected device(s) and erase all data.
              </div>
            </div>
          </div>
        {/if}

        <!-- Actions -->
        <div class="flex justify-end space-x-3">
          <button
            type="button"
            on:click={cancel}
            class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-muted border border-gray-300 dark:border rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-muted"
            disabled={creating}
          >
            Cancel
          </button>
          <button
            type="button"
            on:click={createPool}
            class="px-4 py-2 bg-blue-600 text-white rounded-md shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={creating || !poolConfig.name || poolConfig.devices.length === 0}
          >
            Create Pool
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>
