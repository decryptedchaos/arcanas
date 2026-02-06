<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { onMount } from 'svelte';
  import { lvmAPI } from '$lib/api.js';

  // State
  let volumeGroups = [];
  let logicalVolumes = [];
  let loading = true;
  let error = null;

  // Modal states (only for delete and mount)
  let showDeleteVGModal = false;
  let showDeleteLVModal = false;
  let showMountLVModal = false;

  // Action states
  let deletingVG = false;
  let deletingLV = false;
  let mountingLV = false;

  let vgToDelete = null;
  let lvToDelete = null;
  let lvToMount = null;
  let mountPoolName = '';

  // Expanded state for VG details
  let expandedVGs = new Set();

  function toggleVGExpanded(vgName) {
    if (expandedVGs.has(vgName)) {
      expandedVGs.delete(vgName);
    } else {
      expandedVGs.add(vgName);
    }
    expandedVGs = new Set(expandedVGs);
  }

  // Notification system
  let notifications = [];
  let notificationId = 0;

  function showNotification(message, type = 'info') {
    const id = ++notificationId;
    notifications = [...notifications, { id, message, type }];
    setTimeout(() => {
      notifications = notifications.filter(n => n.id !== id);
    }, 5000);
  }

  // Load functions
  async function loadVolumeGroups() {
    try {
      const result = await lvmAPI.getVolumeGroups();
      volumeGroups = result || [];
    } catch (err) {
      console.error('Failed to load volume groups:', err);
      showNotification(err.message || 'Failed to load volume groups', 'error');
    }
  }

  async function loadLogicalVolumes() {
    try {
      const result = await lvmAPI.getLogicalVolumes();
      logicalVolumes = result || [];
    } catch (err) {
      console.error('Failed to load logical volumes:', err);
      logicalVolumes = [];
    }
  }

  async function loadAll() {
    loading = true;
    error = null;
    await Promise.all([loadVolumeGroups(), loadLogicalVolumes()]);
    loading = false;
  }

  // VG operations
  async function deleteVGConfirm(vg) {
    vgToDelete = vg;
    showDeleteVGModal = true;
  }

  async function deleteVG() {
    try {
      deletingVG = true;
      await lvmAPI.deleteVolumeGroup(vgToDelete.name);
      showNotification(`Volume group "${vgToDelete.name}" deleted`, 'success');
      showDeleteVGModal = false;
      vgToDelete = null;
      await loadAll();
    } catch (err) {
      showNotification(err.message || 'Failed to delete volume group', 'error');
    } finally {
      deletingVG = false;
    }
  }

  // LV operations
  async function deleteLVConfirm(lv) {
    lvToDelete = lv;
    showDeleteLVModal = true;
  }

  async function deleteLV() {
    try {
      deletingLV = true;
      await lvmAPI.deleteLogicalVolume(lvToDelete.path);
      showNotification(`Logical volume "${lvToDelete.name}" deleted`, 'success');
      showDeleteLVModal = false;
      lvToDelete = null;
      await loadAll();
    } catch (err) {
      showNotification(err.message || 'Failed to delete logical volume', 'error');
    } finally {
      deletingLV = false;
    }
  }

  async function mountLVConfirm(lv) {
    lvToMount = lv;
    mountPoolName = lv.name;
    showMountLVModal = true;
  }

  async function mountLV() {
    try {
      mountingLV = true;
      await lvmAPI.mountLVAsPool(lvToMount.path, mountPoolName);
      showNotification(`Logical volume "${lvToMount.name}" mounted as pool "${mountPoolName}"`, 'success');
      showMountLVModal = false;
      lvToMount = null;
      mountPoolName = '';
      await loadAll();
    } catch (err) {
      showNotification(err.message || 'Failed to mount logical volume as pool', 'error');
    } finally {
      mountingLV = false;
    }
  }

  // Helper functions
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

  function getLVsForVG(vgName) {
    return logicalVolumes.filter(lv => lv.vg_name === vgName);
  }

  function getVGUsedPercent(vg) {
    if (!vg.size) return 0;
    const used = vg.size - vg.free;
    return Math.round((used / vg.size) * 100);
  }

  onMount(loadAll);
</script>

<div class="p-6">
  <!-- Notifications -->
  {#if notifications.length > 0}
    <div class="fixed top-4 right-4 z-50 space-y-2">
      {#each notifications as notification (notification.id)}
        <div
          class="p-4 rounded-md shadow-lg flex items-center justify-between max-w-sm
            {notification.type === 'error' ? 'bg-red-50 border border-red-200 text-red-700' : ''}
            {notification.type === 'warning' ? 'bg-yellow-50 border border-yellow-200 text-yellow-700' : ''}
            {notification.type === 'success' ? 'bg-green-50 border border-green-200 text-green-700' : ''}
            {notification.type === 'info' ? 'bg-blue-50 border border-blue-200 text-blue-700' : ''}"
        >
          <span class="text-sm font-medium">{notification.message}</span>
          <button
            on:click={() => notifications = notifications.filter(n => n.id !== notification.id)}
            class="ml-4 text-sm underline hover:no-underline"
          >
            Dismiss
          </button>
        </div>
      {/each}
    </div>
  {/if}

  <!-- Header -->
  <div class="mb-6">
    <div class="flex items-center justify-between">
      <div class="flex items-center space-x-4">
        <div class="relative">
          <div class="absolute inset-0 bg-gradient-to-br from-indigo-400 to-purple-500 rounded-xl blur opacity-25"></div>
          <div class="relative w-14 h-14 bg-gradient-to-br from-indigo-100 to-purple-100 dark:from-indigo-900/40 dark:to-purple-900/40 rounded-xl flex items-center justify-center shadow-lg">
            <svg class="w-7 h-7 text-indigo-600 dark:text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
          </div>
        </div>
        <div>
          <h1 class="text-xl font-bold text-gray-900 dark:text-white">
            LVM Volumes
          </h1>
          <p class="text-sm text-gray-600 dark:text-gray-300">
            Manage Logical Volume Manager volume groups and logical volumes
          </p>
        </div>
      </div>
      <div class="flex space-x-2">
        <button
          on:click={loadAll}
          class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-card border border-gray-300 dark:border rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-muted"
          disabled={loading}
        >
          <svg class="w-4 h-4 mr-2 inline" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          Refresh
        </button>
        <a
          href="/storage-builder/vg"
          class="px-4 py-2 bg-indigo-600 text-white rounded-md shadow-sm hover:bg-indigo-700 inline-flex items-center"
        >
          <svg class="w-4 h-4 mr-2 inline" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
          Create Volume Group
        </a>
      </div>
    </div>
  </div>

  <!-- Summary Stats -->
  <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
    <div class="bg-gradient-to-br from-indigo-50 to-purple-50 dark:from-indigo-900/20 dark:to-purple-900/20 rounded-xl p-4 border border-indigo-100 dark:border-indigo-800">
      <p class="text-xs text-indigo-600 dark:text-indigo-400 font-medium uppercase tracking-wide">
        Volume Groups
      </p>
      <p class="text-2xl font-bold text-gray-900 dark:text-white">
        {volumeGroups.length}
      </p>
    </div>

    <div class="bg-gradient-to-br from-blue-50 to-cyan-50 dark:from-blue-900/20 dark:to-cyan-900/20 rounded-xl p-4 border border-blue-100 dark:border-blue-800">
      <p class="text-xs text-blue-600 dark:text-blue-400 font-medium uppercase tracking-wide">
        Logical Volumes
      </p>
      <p class="text-2xl font-bold text-gray-900 dark:text-white">
        {logicalVolumes.length}
      </p>
    </div>

    <div class="bg-gradient-to-br from-emerald-50 to-teal-50 dark:from-emerald-900/20 dark:to-teal-900/20 rounded-xl p-4 border border-emerald-100 dark:border-emerald-800">
      <p class="text-xs text-emerald-600 dark:text-emerald-400 font-medium uppercase tracking-wide">
        Total Capacity
      </p>
      <p class="text-2xl font-bold text-gray-900 dark:text-white">
        {formatBytes(volumeGroups.reduce((sum, vg) => sum + (vg.size || 0), 0))}
      </p>
    </div>
  </div>

  <!-- Loading -->
  {#if loading}
    <div class="flex justify-center items-center h-64">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
    </div>
  {:else if volumeGroups.length === 0}
    <!-- Empty State -->
    <div class="text-center py-12 bg-white dark:bg-card shadow-lg hover:shadow-xl transition-shadow duration-200 rounded-lg border border-gray-100 dark:border-border">
      <svg class="h-16 w-16 text-gray-400 mx-auto mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
      </svg>
      <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">No volume groups found</h3>
      <p class="text-gray-600 dark:text-gray-300 mb-4">
        Create a volume group to start using LVM for flexible storage management
      </p>
      <a
        href="/storage-builder/vg"
        class="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 inline-block"
      >
        Create Volume Group
      </a>
    </div>
  {:else}
    <!-- Volume Groups List -->
    <div class="space-y-4">
      {#each volumeGroups as vg (vg.name)}
        <div class="bg-white dark:bg-card shadow-lg hover:shadow-xl transition-shadow duration-200 rounded-lg border border-gray-100 dark:border-border overflow-hidden">
          <!-- VG Header -->
          <div class="p-6">
            <div class="flex items-start justify-between">
              <div class="flex items-center space-x-4 flex-1">
                <div class="relative">
                  <div class="absolute inset-0 bg-gradient-to-br from-indigo-400 to-purple-500 rounded-xl blur opacity-25"></div>
                  <div class="relative w-14 h-14 bg-gradient-to-br from-indigo-100 to-purple-100 dark:from-indigo-900/40 dark:to-purple-900/40 rounded-xl flex items-center justify-center shadow-lg">
                    <svg class="w-7 h-7 text-indigo-600 dark:text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                    </svg>
                  </div>
                </div>

                <div class="flex-1">
                  <div class="flex items-center space-x-3">
                    <h3 class="text-xl font-bold text-gray-900 dark:text-white">
                      {vg.name}
                    </h3>
                    <span class="px-2 py-0.5 text-xs font-semibold rounded-full bg-indigo-100 dark:bg-indigo-900/40 text-indigo-700 dark:text-indigo-400">
                      VG
                    </span>
                  </div>

                  <!-- Devices -->
                  {#if vg.devices && vg.devices.length > 0}
                    <div class="mt-2 flex items-center space-x-2">
                      <span class="text-xs text-gray-500 dark:text-gray-400">Devices:</span>
                      {#each vg.devices as device}
                        <code class="text-xs bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded text-gray-700 dark:text-gray-300">
                          {device}
                        </code>
                      {/each}
                    </div>
                  {/if}
                </div>
              </div>

              <!-- Actions -->
              <div class="flex items-center space-x-2">
                <a
                  href="/storage-builder/lvm?vg={vg.name}"
                  class="px-3 py-1.5 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 inline-flex items-center"
                  title="Create Logical Volume"
                >
                  <svg class="w-4 h-4 mr-1 inline" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                  </svg>
                  Add LV
                </a>
                <button
                  on:click={() => toggleVGExpanded(vg.name)}
                  class="p-2 rounded-lg text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-50 dark:hover:bg-muted"
                  title={expandedVGs.has(vg.name) ? 'Hide details' : 'Show details'}
                >
                  <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" class:rotate-180={expandedVGs.has(vg.name)}>
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                  </svg>
                </button>
                <button
                  on:click={() => deleteVGConfirm(vg)}
                  class="p-2 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20"
                  title="Delete Volume Group"
                >
                  <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1 10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>
          </div>

          <!-- VG Stats -->
          <div class="bg-gray-50 dark:bg-muted rounded-b-lg p-4 border-t border-gray-200 dark:border-border">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
              <!-- Total Size -->
              <div class="bg-gradient-to-br from-purple-50 to-pink-50 dark:from-purple-900/20 dark:to-pink-900/20 rounded-xl p-4 border border-purple-100 dark:border-purple-800">
                <div class="flex items-center space-x-3">
                  <div class="w-10 h-10 bg-purple-100 dark:bg-purple-800 rounded-lg flex items-center justify-center">
                    <svg class="w-5 h-5 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"/>
                    </svg>
                  </div>
                  <div>
                    <p class="text-xs text-purple-600 dark:text-purple-400 font-medium uppercase tracking-wide">
                      Total Size
                    </p>
                    <p class="text-lg font-bold text-gray-900 dark:text-white">
                      {formatBytes(vg.size)}
                    </p>
                  </div>
                </div>
              </div>

              <!-- Free Space -->
              <div class="bg-gradient-to-br from-green-50 to-emerald-50 dark:from-green-900/20 dark:to-emerald-900/20 rounded-xl p-4 border border-green-100 dark:border-green-800">
                <div class="flex items-center space-x-3">
                  <div class="w-10 h-10 bg-green-100 dark:bg-green-800 rounded-lg flex items-center justify-center">
                    <svg class="w-5 h-5 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                    </svg>
                  </div>
                  <div>
                    <p class="text-xs text-green-600 dark:text-green-400 font-medium uppercase tracking-wide">
                      Free Space
                    </p>
                    <p class="text-lg font-bold text-gray-900 dark:text-white">
                      {formatBytes(vg.free)}
                    </p>
                  </div>
                </div>
              </div>

              <!-- Usage -->
              <div class="bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 rounded-xl p-4 border border-blue-100 dark:border-blue-800">
                <div class="flex items-center space-x-3">
                  <div class="w-10 h-10 bg-blue-100 dark:bg-blue-800 rounded-lg flex items-center justify-center">
                    <svg class="w-5 h-5 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z"/>
                    </svg>
                  </div>
                  <div>
                    <p class="text-xs text-blue-600 dark:text-blue-400 font-medium uppercase tracking-wide">
                      Used
                    </p>
                    <p class="text-lg font-bold text-gray-900 dark:text-white">
                      {getVGUsedPercent(vg)}%
                    </p>
                  </div>
                </div>
                <div class="mt-2 w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                  <div
                    class="h-full bg-gradient-to-r from-blue-500 to-indigo-500 rounded-full transition-all duration-300"
                    style="width: {getVGUsedPercent(vg)}%"
                  ></div>
                </div>
              </div>
            </div>

            <!-- Logical Volumes for this VG -->
            {#if expandedVGs.has(vg.name)}
              <div class="mt-4 pt-4 border-t border-gray-200 dark:border-border">
                <div class="flex items-center justify-between mb-3">
                  <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300">
                    Logical Volumes ({getLVsForVG(vg.name).length})
                  </h4>
                  <a
                    href="/storage-builder/lvm?vg={vg.name}"
                    class="px-2 py-1 text-xs bg-indigo-600 text-white rounded hover:bg-indigo-700 inline-block"
                  >
                    + Add LV
                  </a>
                </div>

                {#if getLVsForVG(vg.name).length === 0}
                  <div class="text-center py-4 text-sm text-gray-500 dark:text-gray-400">
                    No logical volumes in this volume group
                  </div>
                {:else}
                  <div class="space-y-2">
                    {#each getLVsForVG(vg.name) as lv}
                      <div class="flex items-center justify-between p-3 bg-white dark:bg-card rounded-lg border border-gray-200 dark:border">
                        <div class="flex items-center space-x-3">
                          <div class="w-8 h-8 bg-gradient-to-br from-cyan-100 to-blue-100 dark:from-cyan-900/40 dark:to-blue-900/40 rounded-lg flex items-center justify-center">
                            <svg class="w-4 h-4 text-cyan-600 dark:text-cyan-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"/>
                            </svg>
                          </div>
                          <div>
                            <p class="text-sm font-medium text-gray-900 dark:text-white">{lv.name}</p>
                            <p class="text-xs text-gray-500 dark:text-gray-400 font-mono">{lv.path}</p>
                          </div>
                        </div>
                        <div class="flex items-center space-x-4">
                          <span class="text-sm text-gray-600 dark:text-gray-300">{formatBytes(lv.size)}</span>
                          {#if lv.mountPoint}
                            <span class="px-2 py-0.5 text-xs rounded-full bg-green-100 dark:bg-green-900/40 text-green-700 dark:text-green-400">
                              Mounted: {lv.mountPoint}
                            </span>
                          {:else}
                            <span class="px-2 py-0.5 text-xs rounded-full bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-400">
                              Available
                            </span>
                            <button
                              on:click={() => mountLVConfirm(lv)}
                              class="px-2 py-1 text-xs bg-green-600 text-white rounded hover:bg-green-700"
                              title="Mount as storage pool"
                            >
                              Mount as Pool
                            </button>
                          {/if}
                          <button
                            on:click={() => deleteLVConfirm(lv)}
                            class="p-1 text-gray-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded"
                            title="Delete Logical Volume"
                          >
                            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                            </svg>
                          </button>
                        </div>
                      </div>
                    {/each}
                  </div>
                {/if}
              </div>
            {/if}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Delete VG Confirmation Modal -->
{#if showDeleteVGModal && vgToDelete}
  <div
    class="fixed inset-0 z-50 overflow-y-auto"
    on:click={() => showDeleteVGModal = false}
    on:keydown={(e) => e.key === 'Escape' && (showDeleteVGModal = false)}
    role="dialog"
    aria-modal="true"
    aria-labelledby="delete-vg-title"
    tabindex="-1"
  >
    <div class="flex items-center justify-center min-h-screen px-4">
      <div
        class="fixed inset-0 bg-gray-500 bg-opacity-75"
        aria-hidden="true"
        on:click={() => showDeleteVGModal = false}
      ></div>
      <div
        class="relative bg-white dark:bg-card rounded-lg max-w-md w-full p-6"
        on:click|stopPropagation
      >
        <h3 id="delete-vg-title" class="text-lg font-medium text-gray-900 dark:text-white mb-4">
          Delete Volume Group
        </h3>

        <div class="mb-4">
          <p class="text-sm text-gray-600 dark:text-gray-300 mb-4">
            Are you sure you want to delete the volume group <strong>"{vgToDelete.name}"</strong>?
          </p>
          <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-3">
            <p class="text-sm text-red-800 dark:text-red-200 font-medium mb-1">
              Warning: This will delete all logical volumes in this volume group!
            </p>
            <p class="text-xs text-red-700 dark:text-red-300">
              {vgToDelete.lunCount || 0} logical volume(s) will be permanently deleted.
            </p>
          </div>
        </div>

        <div class="flex justify-end space-x-3">
          <button
            type="button"
            on:click={() => showDeleteVGModal = false}
            class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-card border border-gray-300 dark:border rounded-md hover:bg-gray-50 dark:hover:bg-muted"
          >
            Cancel
          </button>
          <button
            on:click={deleteVG}
            disabled={deletingVG}
            class="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 disabled:opacity-50"
          >
            {#if deletingVG}
              <svg class="animate-spin h-4 w-4 mr-2 inline" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8 0V0C5.373 0 0 1.4 0z"></path>
              </svg>
              Deleting...
            {:else}
              Delete Volume Group
            {/if}
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}

<!-- Delete LV Confirmation Modal -->
{#if showDeleteLVModal && lvToDelete}
  <div
    class="fixed inset-0 z-50 overflow-y-auto"
    on:click={() => showDeleteLVModal = false}
    on:keydown={(e) => e.key === 'Escape' && (showDeleteLVModal = false)}
    role="dialog"
    aria-modal="true"
    aria-labelledby="delete-lv-title"
    tabindex="-1"
  >
    <div class="flex items-center justify-center min-h-screen px-4">
      <div
        class="fixed inset-0 bg-gray-500 bg-opacity-75"
        aria-hidden="true"
        on:click={() => showDeleteLVModal = false}
      ></div>
      <div
        class="relative bg-white dark:bg-card rounded-lg max-w-md w-full p-6"
        on:click|stopPropagation
      >
        <h3 id="delete-lv-title" class="text-lg font-medium text-gray-900 dark:text-white mb-4">
          Delete Logical Volume
        </h3>

        <div class="mb-4">
          <p class="text-sm text-gray-600 dark:text-gray-300 mb-4">
            Are you sure you want to delete the logical volume <strong>"{lvToDelete.name}"</strong>?
          </p>
          {#if lvToDelete.mountPoint}
            <div class="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-md p-3 mb-3">
              <p class="text-sm text-yellow-800 dark:text-yellow-200">
                This logical volume is mounted at {lvToDelete.mountPoint}. It will be unmounted before deletion.
              </p>
            </div>
          {/if}
          <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-3">
            <p class="text-sm text-red-800 dark:text-red-200 font-medium">
              Warning: All data on this logical volume will be permanently deleted!
            </p>
          </div>
        </div>

        <div class="flex justify-end space-x-3">
          <button
            type="button"
            on:click={() => showDeleteLVModal = false}
            class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-card border border-gray-300 dark:border rounded-md hover:bg-gray-50 dark:hover:bg-muted"
          >
            Cancel
          </button>
          <button
            on:click={deleteLV}
            disabled={deletingLV}
            class="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 disabled:opacity-50"
          >
            {#if deletingLV}
              <svg class="animate-spin h-4 w-4 mr-2 inline" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8 0V0C5.373 0 0 1.4 0z"></path>
              </svg>
              Deleting...
            {:else}
              Delete Logical Volume
            {/if}
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}

<!-- Mount LV as Pool Modal -->
{#if showMountLVModal && lvToMount}
  <div
    class="fixed inset-0 z-50 overflow-y-auto"
    on:click={() => showMountLVModal = false}
    on:keydown={(e) => e.key === 'Escape' && (showMountLVModal = false)}
    role="dialog"
    aria-modal="true"
    aria-labelledby="mount-lv-title"
    tabindex="-1"
  >
    <div class="flex items-center justify-center min-h-screen px-4">
      <div
        class="fixed inset-0 bg-gray-500 bg-opacity-75"
        aria-hidden="true"
        on:click={() => showMountLVModal = false}
        on:keydown={(e) => e.key === 'Enter' && (showMountLVModal = false)}
      ></div>
      <div
        class="relative bg-white dark:bg-card rounded-lg max-w-md w-full p-6"
        on:click|stopPropagation
        on:keydown={(e) => e.key === 'Escape' && (showMountLVModal = false)}
        role="dialog"
        tabindex="-1"
      >
        <h3 id="mount-lv-title" class="text-lg font-medium text-gray-900 dark:text-white mb-4">
          Mount Logical Volume as Storage Pool
        </h3>

        <div class="mb-4">
          <p class="text-sm text-gray-600 dark:text-gray-300 mb-4">
            Mount logical volume <strong>"{lvToMount.name}"</strong> as a storage pool for use with NFS/Samba.
          </p>
          <div class="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-md p-3 mb-3">
            <p class="text-xs text-blue-800 dark:text-blue-200">
              The LV will be formatted (if needed) and mounted at <strong>/srv/{mountPoolName}</strong>
            </p>
          </div>
        </div>

        <form on:submit|preventDefault={mountLV} class="space-y-4">
          <div>
            <label for="mount-pool-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Pool Name
            </label>
            <input
              id="mount-pool-name"
              type="text"
              bind:value={mountPoolName}
              class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md dark:bg-muted dark:text-white"
              placeholder="e.g., data"
              pattern="[a-z0-9\-]+"
              title="Only lowercase letters, numbers, and hyphens"
              required
            />
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
              The LV will be mounted at /srv/{mountPoolName}
            </p>
          </div>

          <div class="flex justify-end space-x-3">
            <button
              type="button"
              on:click={() => showMountLVModal = false}
              class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-card border border-gray-300 dark:border rounded-md hover:bg-gray-50 dark:hover:bg-muted"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={mountingLV}
              class="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50"
            >
              {#if mountingLV}
                <svg class="animate-spin h-4 w-4 mr-2 inline" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8 0V0C5.373 0 0 1.4 0z"></path>
                </svg>
                Mounting...
              {:else}
                Mount as Pool
              {/if}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
{/if}
