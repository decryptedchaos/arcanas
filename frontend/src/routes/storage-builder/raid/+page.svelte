<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { diskAPI } from '$lib/api.js';

  let availableDisks = [];
  let loading = true;
  let error = null;
  let creating = false;

  let raidConfig = {
    name: '',
    level: 'raid1',
    devices: []
  };

  async function loadAvailableDisks() {
    try {
      loading = true;
      const diskStats = await diskAPI.getDiskStats();
      availableDisks = diskStats.filter((disk) => {
        return disk.available !== false && !disk.mountpoint;
      });

      // Filter out RAID member disks
      const raidArrays = await diskAPI.getRAIDArrays();
      const raidMemberPaths = raidArrays.flatMap((array) => array.devices || []);
      availableDisks = availableDisks.filter((disk) => {
        return !raidMemberPaths.includes(disk.path);
      });
    } catch (err) {
      error = err.message || 'Failed to load available disks';
      console.error('Failed to load available disks:', err);
    } finally {
      loading = false;
    }
  }

  function toggleDevice(devicePath) {
    if (raidConfig.devices.includes(devicePath)) {
      raidConfig.devices = raidConfig.devices.filter(d => d !== devicePath);
    } else {
      raidConfig.devices = [...raidConfig.devices, devicePath];
    }
  }

  function cancel() {
    goto('/storage-builder');
  }

  async function createRAID() {
    try {
      if (raidConfig.devices.length === 0) {
        error = 'Please select at least one disk';
        return;
      }

      creating = true;
      await diskAPI.createRAIDArray(raidConfig);

      // Success - redirect to storage page
      goto('/storage?tab=raid');
    } catch (err) {
      error = err.message || 'Failed to create RAID array';
      console.error('Failed to create RAID array:', err);
    } finally {
      creating = false;
    }
  }

  onMount(() => {
    loadAvailableDisks();
  });

  const raidLevels = [
    { value: 'raid0', label: 'RAID 0', description: 'Striping - Fastest, no redundancy' },
    { value: 'raid1', label: 'RAID 1', description: 'Mirroring - Redundancy, 50% capacity' },
    { value: 'raid5', label: 'RAID 5', description: 'Parity - Good balance, min 3 disks' },
    { value: 'raid6', label: 'RAID 6', description: 'Dual Parity - High redundancy, min 4 disks' },
    { value: 'raid10', label: 'RAID 10', description: 'Mirrored Striping - Best of both, min 4 disks' }
  ];
</script>

<div class="p-6" role="main" tabindex="-1">
  <!-- Header -->
  <div class="mb-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-gray-900 dark:text-white mb-2">
          Create RAID Array
        </h1>
        <p class="text-sm text-gray-600 dark:text-gray-300">
          Combine physical disks into a RAID array for redundancy or performance
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
        <p class="text-gray-600 dark:text-gray-400">Creating RAID array...</p>
      </div>
    </div>
  {:else}
    <!-- Form -->
    <div class="max-w-3xl">
      <div class="bg-white dark:bg-card rounded-lg border border-gray-200 dark:border-border p-6">
        <!-- RAID Level -->
        <div class="mb-6">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
            RAID Level
          </label>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            {#each raidLevels as level}
              <button
                type="button"
                on:click={() => raidConfig.level = level.value}
                class="p-4 text-left rounded-lg border-2 transition-colors
                  {raidConfig.level === level.value
                    ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                    : 'border-gray-200 dark:border-border hover:border-gray-300 dark:hover:border-border'}"
              >
                <div class="font-medium text-gray-900 dark:text-white">{level.label}</div>
                <div class="text-sm text-gray-600 dark:text-gray-400 mt-1">{level.description}</div>
              </button>
            {/each}
          </div>
        </div>

        <!-- Array Name -->
        <div class="mb-6">
          <label for="raidName" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Array Name (optional)
          </label>
          <input
            id="raidName"
            type="text"
            bind:value={raidConfig.name}
            placeholder="Leave empty to auto-generate (e.g., md0, md1)"
            class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-muted dark:text-white"
          />
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
            If not specified, the system will automatically assign the next available device number.
          </p>
        </div>

        <!-- Disk Selection -->
        <div class="mb-6">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Select Disks
          </label>
          {#if availableDisks.length === 0}
            <div class="p-4 border border-gray-300 dark:border rounded-md bg-gray-50 dark:bg-muted text-gray-500 dark:text-gray-400 text-center">
              No disks available. All devices are either mounted or part of a RAID array.
            </div>
          {:else}
            <div class="border border-gray-300 dark:border rounded-md p-4 max-h-60 overflow-y-auto bg-gray-50 dark:bg-muted">
              {#each availableDisks as disk}
                <label class="flex items-center space-x-3 p-2 hover:bg-gray-100 dark:hover:bg-muted rounded cursor-pointer">
                  <input
                    type="checkbox"
                    checked={raidConfig.devices.includes(disk.path)}
                    on:change={() => toggleDevice(disk.path)}
                    class="rounded border-gray-300 dark:border text-blue-600 focus:ring-blue-500 dark:bg-card"
                  />
                  <div class="flex-1">
                    <div class="flex items-center space-x-2">
                      <span class="font-medium text-gray-900 dark:text-white">
                        {disk.name || disk.path}
                      </span>
                      <span class="text-xs text-gray-500 dark:text-gray-400">
                        ({disk.model || 'Unknown model'})
                      </span>
                    </div>
                    <div class="text-xs text-gray-500 dark:text-gray-400">
                      {disk.path} • {disk.size} bytes
                      {#if disk.filesystem}
                        • {disk.filesystem}
                      {/if}
                    </div>
                  </div>
                </label>
              {/each}
            </div>
            <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
              {raidConfig.devices.length} disk(s) selected
            </p>
          {/if}
        </div>

        <!-- Warning -->
        <div class="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-md p-3 mb-6">
          <div class="flex">
            <svg class="w-5 h-5 text-yellow-400 mr-2 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
            </svg>
            <div class="text-sm text-yellow-800 dark:text-yellow-200">
              <strong>Warning:</strong> Creating a RAID array will erase all data on the selected disks.
            </div>
          </div>
        </div>

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
            on:click={createRAID}
            class="px-4 py-2 bg-blue-600 text-white rounded-md shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={creating || raidConfig.devices.length === 0}
          >
            Create RAID Array
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>
