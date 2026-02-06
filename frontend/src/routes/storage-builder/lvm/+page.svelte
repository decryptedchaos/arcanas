<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { lvmAPI } from '$lib/api.js';

  let volumeGroups = [];
  let logicalVolumes = [];
  let loading = true;
  let error = null;
  let creating = false;

  let lvConfig = {
    name: '',
    vgName: '',
    sizeGB: 10
  };

  async function loadData() {
    try {
      loading = true;
      const [vgs, lvs] = await Promise.all([
        lvmAPI.getVolumeGroups(),
        lvmAPI.getLogicalVolumes()
      ]);
      volumeGroups = vgs || [];
      logicalVolumes = lvs || [];

      // Check for vg query parameter to pre-select
      const urlVG = $page.url.searchParams.get('vg');
      if (urlVG && volumeGroups.find(vg => vg.name === urlVG)) {
        lvConfig.vgName = urlVG;
      } else if (volumeGroups.length > 0 && !lvConfig.vgName) {
        // Auto-select first VG if available
        lvConfig.vgName = volumeGroups[0].name;
      }
    } catch (err) {
      error = err.message || 'Failed to load LVM data';
      console.error('Failed to load LVM data:', err);
    } finally {
      loading = false;
    }
  }

  function cancel() {
    goto('/storage-builder');
  }

  async function createLogicalVolume() {
    try {
      if (!lvConfig.name) {
        error = 'Please enter a logical volume name';
        return;
      }

      if (!lvConfig.vgName) {
        error = 'Please select a volume group';
        return;
      }

      if (lvConfig.sizeGB < 1) {
        error = 'Size must be at least 1 GB';
        return;
      }

      creating = true;
      const lvData = {
        name: lvConfig.name,
        vg_name: lvConfig.vgName,
        size_gb: lvConfig.sizeGB
      };
      await lvmAPI.createLogicalVolume(lvData);

      // Success - redirect to LVM page
      goto('/lvm');
    } catch (err) {
      error = err.message || 'Failed to create logical volume';
      console.error('Failed to create logical volume:', err);
    } finally {
      creating = false;
    }
  }

  function goToCreateVG() {
    goto('/storage-builder/vg');
  }

  function formatBytes(bytes) {
    if (!bytes) return '0 B';
    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    let i = 0;
    while (bytes >= 1024 && i < units.length - 1) {
      bytes /= 1024;
      i++;
    }
    return `${bytes.toFixed(bytes < 10 && i > 0 ? 1 : 0)} ${units[i]}`;
  }

  onMount(() => {
    loadData();
  });

  $: selectedVG = volumeGroups.find(vg => vg.name === lvConfig.vgName);
  $: maxLVSizeGB = selectedVG ? (selectedVG.free / (1024 * 1024 * 1024)) : 0;
</script>

<div class="p-6" role="main" tabindex="-1">
  <!-- Header -->
  <div class="mb-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-gray-900 dark:text-white mb-2">
          Create Logical Volume
        </h1>
        <p class="text-sm text-gray-600 dark:text-gray-300">
          Create a new logical volume in an existing volume group
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
        <p class="text-gray-600 dark:text-gray-400">Creating logical volume...</p>
      </div>
    </div>
  {:else}
    <!-- No Volume Groups Alert -->
    {#if volumeGroups.length === 0}
      <div class="max-w-2xl">
        <div class="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-6 text-center">
          <svg class="w-16 h-16 text-yellow-500 mx-auto mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
          </svg>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
            No Volume Groups Found
          </h3>
          <p class="text-gray-600 dark:text-gray-300 mb-4">
            You need to create a volume group before you can create logical volumes.
            Volume groups combine physical devices (like RAID arrays or disks) into a storage pool.
          </p>
          <div class="flex justify-center space-x-3">
            <button
              on:click={cancel}
              class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-card border border-gray-300 dark:border rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-muted"
            >
              Cancel
            </button>
            <button
              on:click={goToCreateVG}
              class="px-4 py-2 bg-indigo-600 text-white rounded-md shadow-sm hover:bg-indigo-700"
            >
              Create Volume Group
            </button>
          </div>
        </div>
      </div>
    {:else}
      <!-- Form -->
      <div class="max-w-3xl">
        <div class="bg-white dark:bg-card rounded-lg border border-gray-200 dark:border-border p-6">
          <!-- Volume Group Selection -->
          <div class="mb-6">
            <label for="vg-select" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Volume Group
            </label>
            <select
              id="vg-select"
              bind:value={lvConfig.vgName}
              class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 dark:bg-muted dark:text-white"
            >
              {#each volumeGroups as vg}
                <option value={vg.name}>{vg.name}</option>
              {/each}
            </select>
            {#if selectedVG}
              <div class="mt-2 flex items-center space-x-4 text-xs text-gray-500 dark:text-gray-400">
                <span>Total: {formatBytes(selectedVG.size)}</span>
                <span class="text-green-600 dark:text-green-400">Free: {formatBytes(selectedVG.free)}</span>
                <span>{Math.round((selectedVG.free / selectedVG.size) * 100)}% available</span>
              </div>
              <div class="mt-2 w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                <div
                  class="h-full bg-green-500 rounded-full"
                  style="width: {(selectedVG.free / selectedVG.size) * 100}%"
                ></div>
              </div>
            {/if}
          </div>

          <!-- LV Name -->
          <div class="mb-6">
            <label for="lv-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Logical Volume Name
            </label>
            <input
              id="lv-name"
              type="text"
              bind:value={lvConfig.name}
              placeholder="e.g., data"
              pattern="[a-z0-9\-]+"
              title="Only lowercase letters, numbers, and hyphens"
              class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 dark:bg-muted dark:text-white"
              required
            />
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              The LV will be created at <code class="bg-gray-100 dark:bg-gray-800 px-1 rounded">/dev/{lvConfig.vgName || 'vg'}/{lvConfig.name || 'name'}</code>
            </p>
          </div>

          <!-- LV Size -->
          <div class="mb-6">
            <label for="lv-size" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Size (GB)
            </label>
            <div class="flex items-center space-x-4">
              <input
                id="lv-size"
                type="number"
                bind:value={lvConfig.sizeGB}
                min="1"
                max={maxLVSizeGB}
                step="0.1"
                class="flex-1 px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 dark:bg-muted dark:text-white"
                required
              />
              <div class="text-sm text-gray-600 dark:text-gray-400 whitespace-nowrap">
                of {maxLVSizeGB.toFixed(1)} GB available
              </div>
            </div>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              You can resize this logical volume later if needed
            </p>

            <!-- Quick size options -->
            {#if selectedVG}
              <div class="mt-2 flex flex-wrap gap-2">
                {#if maxLVSizeGB >= 10}
                  <button
                    type="button"
                    on:click={() => lvConfig.sizeGB = Math.min(10, maxLVSizeGB)}
                    class="px-3 py-1 text-xs bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-200 dark:hover:bg-gray-600"
                  >
                    10 GB
                  </button>
                {/if}
                {#if maxLVSizeGB >= 100}
                  <button
                    type="button"
                    on:click={() => lvConfig.sizeGB = Math.min(100, maxLVSizeGB)}
                    class="px-3 py-1 text-xs bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-200 dark:hover:bg-gray-600"
                  >
                    100 GB
                  </button>
                {/if}
                {#if maxLVSizeGB >= 500}
                  <button
                    type="button"
                    on:click={() => lvConfig.sizeGB = Math.min(500, maxLVSizeGB)}
                    class="px-3 py-1 text-xs bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-200 dark:hover:bg-gray-600"
                  >
                    500 GB
                  </button>
                {/if}
                <button
                  type="button"
                  on:click={() => lvConfig.sizeGB = Math.round(maxLVSizeGB * 0.5)}
                  class="px-3 py-1 text-xs bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-200 dark:hover:bg-gray-600"
                >
                  50%
                </button>
                <button
                  type="button"
                  on:click={() => lvConfig.sizeGB = Math.round(maxLVSizeGB * 0.9)}
                  class="px-3 py-1 text-xs bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-200 dark:hover:bg-gray-600"
                >
                  90%
                </button>
                <button
                  type="button"
                  on:click={() => lvConfig.sizeGB = Math.round(maxLVSizeGB)}
                  class="px-3 py-1 text-xs bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-200 dark:hover:bg-gray-600"
                >
                  100%
                </button>
              </div>
            {/if}
          </div>

          <!-- Existing LVs in selected VG -->
          {#if logicalVolumes.filter(lv => lv.vg_name === lvConfig.vgName).length > 0}
            <div class="mb-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
              <p class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Existing Logical Volumes in {lvConfig.vgName}:
              </p>
              <div class="space-y-1">
                {#each logicalVolumes.filter(lv => lv.vg_name === lvConfig.vgName) as lv}
                  <div class="flex items-center justify-between text-sm">
                    <span class="text-gray-900 dark:text-white">{lv.name}</span>
                    <span class="text-gray-600 dark:text-gray-400">{formatBytes(lv.size)}</span>
                  </div>
                {/each}
              </div>
            </div>
          {/if}

          <!-- Info -->
          <div class="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-md p-3 mb-6">
            <div class="flex">
              <svg class="w-5 h-5 text-blue-400 mr-2 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
              </svg>
              <div class="text-sm text-blue-800 dark:text-blue-200">
                <strong>Logical Volumes</strong> are flexible partitions that can be resized independently. After creating, you can mount it as a storage pool for use with NFS/Samba.
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
              on:click={createLogicalVolume}
              class="px-4 py-2 bg-indigo-600 text-white rounded-md shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
              disabled={creating || !lvConfig.name || !lvConfig.vgName || lvConfig.sizeGB < 1 || lvConfig.sizeGB > maxLVSizeGB}
            >
              Create Logical Volume
            </button>
          </div>
        </div>
      </div>
    {/if}
  {/if}
</div>
