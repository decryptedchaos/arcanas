<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { createEventDispatcher } from 'svelte';
  import { diskAPI } from '$lib/api.js';

  export let availableDisks = [];
  export let selectedDisks = [];

  const dispatch = createEventDispatcher();

  let disks = availableDisks;
  let loading = false;
  let smartData = {};

  // Load SMART data for disks on mount
  async function loadSmartData() {
    for (const disk of availableDisks) {
      try {
        const data = await diskAPI.getSmartStatus(disk.path);
        smartData[disk.path] = data;
      } catch (err) {
        // SMART not available or error - that's okay
        smartData[disk.path] = null;
      }
    }
  }

  function toggleDisk(diskPath) {
    const index = selectedDisks.indexOf(diskPath);
    if (index === -1) {
      selectedDisks = [...selectedDisks, diskPath];
    } else {
      selectedDisks = selectedDisks.filter((d) => d !== diskPath);
    }
    dispatch('select', { detail: selectedDisks });
  }

  function handleNext() {
    if (selectedDisks.length > 0) {
      dispatch('next');
    }
  }

  function handlePrevious() {
    dispatch('previous');
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

  function getDiskHealth(disk) {
    const smart = smartData[disk.path];
    if (!smart) return { status: 'unknown', text: 'Unknown', color: 'gray' };

    if (smart.smart_supported === false) {
      return { status: 'no-smart', text: 'No SMART', color: 'gray' };
    }

    if (smart.smart_enabled === false) {
      return { status: 'disabled', text: 'Disabled', color: 'yellow' };
    }

    if (smart.overall_status === 'PASSED') {
      return { status: 'passed', text: 'Healthy', color: 'green' };
    }

    if (smart.overall_status === 'FAILED') {
      return { status: 'failed', text: 'Failed', color: 'red' };
    }

    return { status: 'unknown', text: 'Unknown', color: 'gray' };
  }
</script>

<div class="bg-white dark:bg-card rounded-lg shadow-lg overflow-hidden">
  <!-- Step Header -->
  <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-xl font-semibold text-gray-900 dark:text-white">
          Select Physical Disks
        </h2>
        <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
          Choose the disks you want to use for your storage. You can select multiple disks for RAID.
        </p>
      </div>
      <div class="text-right">
        <p class="text-sm text-gray-600 dark:text-gray-400">
          Selected: <span class="font-semibold text-indigo-600 dark:text-indigo-400">{selectedDisks.length}</span> /
          {availableDisks.length}
        </p>
        <p class="text-xs text-gray-500 dark:text-gray-500 mt-1">
          {selectedDisks.length === 0
            ? 'Select at least one disk'
            : selectedDisks.length === 1
            ? 'Single disk selected'
            : selectedDisks.length === 2
            ? 'RAID 1 or RAID 0 available'
            : 'RAID 5, 6, or 10 available'}
        </p>
      </div>
    </div>
  </div>

  <!-- Disk List -->
  <div class="p-6">
    {#if availableDisks.length === 0}
      <div class="text-center py-12 bg-gray-50 dark:bg-gray-800 rounded-lg">
        <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
        <h3 class="mt-2 text-sm font-medium text-gray-900 dark:text-white">No available disks</h3>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          All disks are either mounted or in use. You need to free up a disk to create storage.
        </p>
      </div>
    {:else}
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        {#each availableDisks as disk}
          {@const health = getDiskHealth(disk)}
          {@const isSelected = selectedDisks.includes(disk.path)}
          {@const smart = smartData[disk.path]}

          <button
            on:click={() => toggleDisk(disk.path)}
            class="relative p-4 rounded-lg border-2 text-left transition-all
              {isSelected
                ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-900/20'
                : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'}"
            type="button"
          >
            <!-- Checkbox indicator -->
            <div class="absolute top-4 right-4">
              <div
                class="w-6 h-6 rounded-full border-2 flex items-center justify-center
                  {isSelected
                    ? 'border-indigo-500 bg-indigo-500'
                    : 'border-gray-300 dark:border-gray-600'}"
              >
                {#if isSelected}
                  <svg class="w-4 h-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
                  </svg>
                {/if}
              </div>
            </div>

            <!-- Disk Info -->
            <div class="space-y-2">
              <div class="flex items-center space-x-2">
                <svg class="w-5 h-5 text-gray-600 dark:text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
                </svg>
                <span class="font-mono text-sm font-medium text-gray-900 dark:text-white">
                  {disk.path}
                </span>
              </div>

              <div class="flex items-center justify-between">
                <span class="text-sm text-gray-600 dark:text-gray-400">
                  {formatBytes(disk.size)}
                </span>
                <span class="px-2 py-0.5 text-xs font-medium rounded-full
                  {health.color === 'green'
                    ? 'bg-green-100 dark:bg-green-900/40 text-green-700 dark:text-green-400'
                    : health.color === 'red'
                    ? 'bg-red-100 dark:bg-red-900/40 text-red-700 dark:text-red-400'
                    : health.color === 'yellow'
                    ? 'bg-yellow-100 dark:bg-yellow-900/40 text-yellow-700 dark:text-yellow-400'
                    : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300'}">
                  {health.text}
                </span>
              </div>

              {#if smart && smart.temperature}
                <div class="text-xs text-gray-500 dark:text-gray-400">
                  Temp: {smart.temperature}°C
                </div>
              {/if}

              {#if disk.model}
                <div class="text-xs text-gray-500 dark:text-gray-400 truncate" title={disk.model}>
                  {disk.model}
                </div>
              {/if}
            </div>
          </button>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Actions -->
  <div class="px-6 py-4 bg-gray-50 dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 flex items-center justify-between">
    <button
      on:click={handlePrevious}
      class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"
    >
      ← Cancel
    </button>

    <div class="flex items-center space-x-3">
      <p class="text-sm text-gray-600 dark:text-gray-400">
        {selectedDisks.length === 0
          ? 'Select at least one disk to continue'
          : selectedDisks.length === 1
          ? 'Ready to create single disk storage'
          : `${selectedDisks.length} disks selected - RAID available`}
      </p>
      <button
        on:click={handleNext}
        disabled={selectedDisks.length === 0}
        class="px-6 py-2 text-sm font-medium text-white bg-indigo-600 rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {selectedDisks.length === 0
          ? 'Select Disks'
          : selectedDisks.length === 1
          ? 'Next Step →'
          : `Next (${selectedDisks.length} disks) →`}
      </button>
    </div>
  </div>
</div>
