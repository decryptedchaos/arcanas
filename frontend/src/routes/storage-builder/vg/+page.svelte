<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { lvmAPI } from '$lib/api.js';

  let availableDevices = [];
  let loading = true;
  let error = null;
  let creating = false;

  let vgConfig = {
    name: '',
    devices: []
  };

  async function loadAvailableDevices() {
    try {
      loading = true;
      const result = await lvmAPI.getAvailableDevices();
      availableDevices = result || [];
    } catch (err) {
      error = err.message || 'Failed to load available devices';
      console.error('Failed to load available devices:', err);
    } finally {
      loading = false;
    }
  }

  function toggleDevice(devicePath) {
    if (vgConfig.devices.includes(devicePath)) {
      vgConfig.devices = vgConfig.devices.filter(d => d !== devicePath);
    } else {
      vgConfig.devices = [...vgConfig.devices, devicePath];
    }
  }

  function cancel() {
    goto('/storage-builder');
  }

  async function createVolumeGroup() {
    try {
      if (!vgConfig.name) {
        error = 'Please enter a volume group name';
        return;
      }

      if (vgConfig.devices.length === 0) {
        error = 'Please select at least one device';
        return;
      }

      creating = true;
      await lvmAPI.createVolumeGroup(vgConfig);

      // Success - redirect to LVM page
      goto('/lvm');
    } catch (err) {
      error = err.message || 'Failed to create volume group';
      console.error('Failed to create volume group:', err);
    } finally {
      creating = false;
    }
  }

  function goToCreateLV() {
    goto('/storage-builder/lvm');
  }

  onMount(() => {
    loadAvailableDevices();
  });
</script>

<div class="p-6" role="main" tabindex="-1">
  <!-- Header -->
  <div class="mb-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-gray-900 dark:text-white mb-2">
          Create Volume Group
        </h1>
        <p class="text-sm text-gray-600 dark:text-gray-300">
          Combine physical devices into a storage pool for logical volumes
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
        <p class="text-gray-600 dark:text-gray-400">Creating volume group...</p>
      </div>
    </div>
  {:else}
    <!-- Form -->
    <div class="max-w-3xl">
      <div class="bg-white dark:bg-card rounded-lg border border-gray-200 dark:border-border p-6">
        <!-- VG Name -->
        <div class="mb-6">
          <label for="vg-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Volume Group Name
          </label>
          <input
            id="vg-name"
            type="text"
            bind:value={vgConfig.name}
            placeholder="e.g., vg-data"
            pattern="[a-z0-9\-]+"
            title="Only lowercase letters, numbers, and hyphens"
            class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 dark:bg-muted dark:text-white"
            required
          />
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
            Use lowercase letters, numbers, and hyphens only
          </p>
        </div>

        <!-- Device Selection -->
        <div class="mb-6">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Select Physical Devices
          </label>
          {#if availableDevices.length === 0}
            <div class="p-4 border border-gray-300 dark:border rounded-md bg-gray-50 dark:bg-muted text-gray-500 dark:text-gray-400 text-center">
              <p class="mb-2">No devices available.</p>
              <p class="text-xs">Devices must be unmounted and not in use. Create a RAID array first if needed.</p>
              <a
                href="/storage-builder/raid"
                class="inline-block mt-3 px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 text-sm"
              >
                Create RAID Array
              </a>
            </div>
          {:else}
            <div class="border border-gray-300 dark:border rounded-md p-4 max-h-60 overflow-y-auto bg-gray-50 dark:bg-muted">
              {#each availableDevices as device}
                <label class="flex items-center space-x-3 p-2 hover:bg-gray-100 dark:hover:bg-muted rounded cursor-pointer">
                  <input
                    type="checkbox"
                    checked={vgConfig.devices.includes(device.path)}
                    on:change={() => toggleDevice(device.path)}
                    disabled={!device.available}
                    class="rounded border-gray-300 dark:border text-indigo-600 focus:ring-indigo-500 dark:bg-card"
                  />
                  <div class="flex-1">
                    <div class="font-medium text-gray-900 dark:text-white">{device.path}</div>
                    {#if device.available}
                      <div class="text-xs text-green-600 dark:text-green-400">Available</div>
                    {:else}
                      <div class="text-xs text-red-600 dark:text-red-400">{device.reason}</div>
                    {/if}
                  </div>
                </label>
              {/each}
            </div>
            <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
              {vgConfig.devices.length} device(s) selected
            </p>
          {/if}
        </div>

        <!-- Info -->
        <div class="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-md p-3 mb-6">
          <div class="flex">
            <svg class="w-5 h-5 text-blue-400 mr-2 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
            </svg>
            <div class="text-sm text-blue-800 dark:text-blue-200">
              <strong>What is a Volume Group?</strong> A volume group combines physical devices (like RAID arrays or disks) into a storage pool. You can then create logical volumes from this pool.
            </div>
          </div>
        </div>

        <!-- Next Steps -->
        <div class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-md p-3 mb-6">
          <div class="flex items-center">
            <svg class="w-5 h-5 text-green-500 mr-2 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-11a1 1 0 10-2 0v2H7a1 1 0 100 2h2v2a1 1 0 102 0v-2h2a1 1 0 100-2h-2V7z" clip-rule="evenodd" />
            </svg>
            <div class="text-sm text-green-800 dark:text-green-200 flex-1">
              After creating the volume group, you can create logical volumes.
            </div>
          </div>
        </div>

        <!-- Actions -->
        <div class="flex justify-end space-x-3">
          <button
            type="button"
            on:click={cancel}
            class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-muted border border-gray-300 dark:border rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-muted"
          >
            Cancel
          </button>
          <button
            type="button"
            on:click={createVolumeGroup}
            class="px-4 py-2 bg-indigo-600 text-white rounded-md shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={creating || !vgConfig.name || vgConfig.devices.length === 0}
          >
            Create Volume Group
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>
