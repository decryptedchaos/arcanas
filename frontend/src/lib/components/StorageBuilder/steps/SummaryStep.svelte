<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { createEventDispatcher } from 'svelte';
  import { diskAPI } from '$lib/api.js';

  export let wizardConfig = {
    selectedDisks: [],
    raidConfig: { enabled: false, level: '', name: '' },
    lvmConfig: { enabled: false, vgName: '', lvName: '', sizeGB: null },
    poolConfig: { name: '', mountPoint: '', filesystem: 'ext4' },
    shareConfig: { createNFS: false, createSamba: false, nfsClients: [], sambaName: '' }
  };

  export let disks = [];
  let creating = false;

  const dispatch = createEventDispatcher();

  function getSourceDescription() {
    if (wizardConfig.lvmConfig.enabled) {
      return `LVM: ${wizardConfig.lvmConfig.vgName}/${wizardConfig.lvmConfig.lvName}`;
    } else if (wizardConfig.raidConfig.enabled) {
      return `RAID ${wizardConfig.raidConfig.level}: ${wizardConfig.raidConfig.name}`;
    } else {
      return `${wizardConfig.selectedDisks.length} disk(s)`;
    }
  }

  function getStorageSteps() {
    const steps = [];
    steps.push({ title: 'Disks', description: `${wizardConfig.selectedDisks.length} disk(s) selected` });

    if (wizardConfig.raidConfig.enabled) {
      steps.push({
        title: 'RAID',
        description: `${wizardConfig.raidConfig.level.toUpperCase()} - ${wizardConfig.raidConfig.name}`
      });
    }

    if (wizardConfig.lvmConfig.enabled) {
      steps.push({
        title: 'LVM',
        description: `VG: ${wizardConfig.lvmConfig.vgName}, LV: ${wizardConfig.lvmConfig.lvName} (${wizardConfig.lvmConfig.sizeGB} GB)`
      });
    }

    steps.push({
      title: 'Pool',
      description: `${wizardConfig.poolConfig.name} at ${wizardConfig.poolConfig.mountPoint || '/srv/' + wizardConfig.poolConfig.name} (${wizardConfig.poolConfig.filesystem.toUpperCase()})`
    });

    const shareTypes = [];
    if (wizardConfig.shareConfig.createNFS) shareTypes.push('NFS');
    if (wizardConfig.shareConfig.createSamba) shareTypes.push('Samba');

    steps.push({
      title: 'Shares',
      description: shareTypes.length > 0 ? shareTypes.join(' + ') : 'None (configured later)'
    });

    return steps;
  }

  async function handleCreate() {
    creating = true;
    try {
      // Dispatch the create event with all config
      dispatch('create', { detail: wizardConfig });
    } finally {
      creating = false;
    }
  }

  function handleBack() {
    dispatch('previous');
  }

  $: storageSteps = getStorageSteps();
  $: sourceDescription = getSourceDescription();
</script>

<div class="bg-white dark:bg-card rounded-lg shadow-lg overflow-hidden">
  <!-- Step Header -->
  <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-xl font-semibold text-gray-900 dark:text-white">
          Review & Create Storage
        </h2>
        <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
          Review your configuration before creating
        </p>
      </div>
    </div>
  </div>

  <div class="p-6 space-y-6">
    <!-- Storage Stack Visualization -->
    <div>
      <h3 class="text-sm font-medium text-gray-900 dark:text-white mb-4">Storage Stack</h3>
      <div class="relative">
        <!-- Timeline -->
        <div class="absolute left-4 top-0 bottom-0 w-0.5 bg-gray-200 dark:bg-gray-700"></div>

        <div class="space-y-4">
          {#each storageSteps as step, index}
            <div class="relative flex items-start space-x-4">
              <div class="relative z-10 w-8 h-8 rounded-full bg-indigo-600 flex items-center justify-center flex-shrink-0">
                <span class="text-white text-xs font-bold">{index + 1}</span>
              </div>
              <div class="flex-1 pt-1">
                <h4 class="text-sm font-semibold text-gray-900 dark:text-white">{step.title}</h4>
                <p class="text-sm text-gray-600 dark:text-gray-400">{step.description}</p>
              </div>
            </div>
          {/each}
        </div>
      </div>
    </div>

    <!-- Configuration Details -->
    <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
      <h3 class="text-sm font-medium text-gray-900 dark:text-white mb-3">Configuration Details</h3>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
        <div>
          <span class="text-gray-600 dark:text-gray-400">Pool name:</span>
          <span class="ml-2 font-mono text-gray-900 dark:text-white">{wizardConfig.poolConfig.name}</span>
        </div>
        <div>
          <span class="text-gray-600 dark:text-gray-400">Mount point:</span>
          <span class="ml-2 font-mono text-gray-900 dark:text-white">{wizardConfig.poolConfig.mountPoint || '/srv/' + wizardConfig.poolConfig.name}</span>
        </div>
        <div>
          <span class="text-gray-600 dark:text-gray-400">Filesystem:</span>
          <span class="ml-2 text-gray-900 dark:text-white">{wizardConfig.poolConfig.filesystem.toUpperCase()}</span>
        </div>
        <div>
          <span class="text-gray-600 dark:text-gray-400">Source:</span>
          <span class="ml-2 text-gray-900 dark:text-white">{sourceDescription}</span>
        </div>
      </div>
    </div>

    <!-- Warnings -->
    <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
      <div class="flex items-start space-x-3">
        <svg class="w-5 h-5 text-red-600 dark:text-red-400 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
        <div class="flex-1 text-sm text-red-800 dark:text-red-200">
          <p class="font-medium mb-1">Warning: Data will be erased</p>
          <ul class="list-disc list-inside space-y-1 text-red-700 dark:text-red-300">
            <li>All selected disks will be formatted</li>
            <li>Any existing data on these devices will be permanently lost</li>
            <li>This action cannot be undone</li>
          </ul>
        </div>
      </div>
    </div>

    <!-- What happens next -->
    <div class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg p-4">
      <div class="flex items-start space-x-3">
        <svg class="w-5 h-5 text-green-600 dark:text-green-400 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <div class="flex-1 text-sm text-green-800 dark:text-green-200">
          <p class="font-medium mb-1">After creation:</p>
          <ul class="list-disc list-inside space-y-1 text-green-700 dark:text-green-300">
            <li>Your storage pool will be mounted and ready to use</li>
            {#if wizardConfig.shareConfig.createNFS}
              <li>NFS export will be available immediately</li>
            {/if}
            {#if wizardConfig.shareConfig.createSamba}
              <li>Samba share will be visible on the network</li>
            {/if}
            <li>You can add more shares or configure settings later</li>
          </ul>
        </div>
      </div>
    </div>
  </div>

  <!-- Actions -->
  <div class="px-6 py-4 bg-gray-50 dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 flex items-center justify-between">
    <button
      on:click={handleBack}
      disabled={creating}
      class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white disabled:opacity-50"
    >
      ← Previous
    </button>

    <div class="flex items-center space-x-3">
      <button
        on:click={handleCreate}
        disabled={creating}
        class="px-6 py-2 text-sm font-medium text-white bg-green-600 rounded-md hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 disabled:opacity-50 disabled:cursor-not-allowed flex items-center space-x-2"
      >
        {#if creating}
          <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8 0V0C5.373 0 0 1.4 0z"></path>
          </svg>
          <span>Creating...</span>
        {:else}
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
          <span>Create Storage</span>
        {/if}
      </button>
    </div>
  </div>
</div>
