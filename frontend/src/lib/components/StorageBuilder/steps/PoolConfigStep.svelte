<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { createEventDispatcher } from 'svelte';

  export let raidConfig = { enabled: false, name: '', level: '' };
  export let lvmConfig = { enabled: false, vgName: '', lvName: '', sizeGB: null };
  export let poolConfig = {
    name: '',
    mountPoint: '',
    filesystem: 'ext4'
  };

  const dispatch = createEventDispatcher();

  const filesystems = [
    { value: 'ext4', name: 'ext4', description: 'Most compatible, reliable for most use cases' },
    { value: 'xfs', name: 'XFS', description: 'Better for large files, high performance' },
    { value: 'btrfs', name: 'Btrfs', description: 'Advanced features: snapshots, compression' }
  ];

  function handleNext() {
    dispatch('next');
  }

  function handlePrevious() {
    dispatch('previous');
  }

  function getSourceDescription() {
    if (lvmConfig.enabled) {
      return `LVM Volume: ${lvmConfig.vgName}/${lvmConfig.lvName}`;
    } else if (raidConfig.enabled) {
      return `RAID Array: ${raidConfig.name}`;
    } else {
      return 'Direct disk access';
    }
  }

  function getDefaultMountPoint() {
    if (poolConfig.name) {
      return `/srv/${poolConfig.name}`;
    }
    return '/srv/storage';
  }

  $: mountPoint = poolConfig.mountPoint || getDefaultMountPoint();
  $: sourceDescription = getSourceDescription();
</script>

<div class="bg-white dark:bg-card rounded-lg shadow-lg overflow-hidden">
  <!-- Step Header -->
  <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-xl font-semibold text-gray-900 dark:text-white">
          Storage Pool Configuration
        </h2>
        <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
          Configure your storage pool settings
        </p>
      </div>
      <div class="text-sm text-gray-600 dark:text-gray-400">
        Source: {sourceDescription}
      </div>
    </div>
  </div>

  <!-- Pool Configuration Form -->
  <div class="p-6 space-y-6">
    <!-- Pool Name -->
    <div>
      <label for="pool-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
        Pool Name
      </label>
      <input
        id="pool-name"
        type="text"
        bind:value={poolConfig.name}
        class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md dark:bg-gray-800 dark:text-white"
        placeholder="e.g., data, backups, media"
        pattern="[a-z0-9\-]+"
        title="Use lowercase letters, numbers, and hyphens"
        required
      />
      <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
        A unique name for this storage pool
      </p>
    </div>

    <!-- Mount Point -->
    <div>
      <label for="mount-point" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
        Mount Point
      </label>
      <div class="flex items-center space-x-2">
        <span class="text-gray-500 dark:text-gray-400">/srv/</span>
        <input
          id="mount-point"
          type="text"
          bind:value={poolConfig.mountPoint}
          class="flex-1 px-3 py-2 border border-gray-300 dark:border rounded-md dark:bg-gray-800 dark:text-white"
          placeholder="data"
          pattern="[a-z0-9\-]+"
          title="Use lowercase letters, numbers, and hyphens"
        />
      </div>
      <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
        The pool will be mounted at: <span class="font-mono">{mountPoint}</span>
      </p>
    </div>

    <!-- Filesystem Selection -->
    <div>
      <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
        Filesystem
      </label>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
        {#each filesystems as fs}
          {@const isSelected = poolConfig.filesystem === fs.value}

          <button
            on:click={() => poolConfig.filesystem = fs.value}
            type="button"
            class="p-3 rounded-lg border-2 text-left transition-all
              {isSelected
                ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-900/20'
                : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'}"
          >
            <div class="flex items-center justify-between mb-1">
              <span class="font-medium text-gray-900 dark:text-white">{fs.name}</span>
              <div
                class="w-4 h-4 rounded-full border-2 flex items-center justify-center
                  {isSelected
                    ? 'border-indigo-500 bg-indigo-500'
                    : 'border-gray-300 dark:border-gray-600'}"
              >
                {#if isSelected}
                  <svg class="w-2.5 h-2.5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
                  </svg>
                {/if}
              </div>
            </div>
            <p class="text-xs text-gray-600 dark:text-gray-400">{fs.description}</p>
          </button>
        {/each}
      </div>
    </div>

    <!-- Configuration Summary -->
    <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
      <h4 class="text-sm font-medium text-gray-900 dark:text-white mb-3">Configuration Summary</h4>
      <div class="space-y-2 text-sm">
        <div class="flex justify-between">
          <span class="text-gray-600 dark:text-gray-400">Pool name:</span>
          <span class="font-mono text-gray-900 dark:text-white">{poolConfig.name || '(not set)'}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-gray-600 dark:text-gray-400">Mount point:</span>
          <span class="font-mono text-gray-900 dark:text-white">{mountPoint}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-gray-600 dark:text-gray-400">Filesystem:</span>
          <span class="font-medium text-gray-900 dark:text-white">{poolConfig.filesystem.toUpperCase()}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-gray-600 dark:text-gray-400">Source:</span>
          <span class="text-gray-900 dark:text-white">{sourceDescription}</span>
        </div>
      </div>
    </div>

    <!-- Information box -->
    <div class="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-4">
      <div class="flex items-start space-x-3">
        <svg class="w-5 h-5 text-yellow-600 dark:text-yellow-400 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
        <div class="flex-1 text-sm text-yellow-800 dark:text-yellow-200">
          <p class="font-medium mb-1">Before creating your pool</p>
          <ul class="list-disc list-inside space-y-1 text-yellow-700 dark:text-yellow-300">
            <li>The pool will be formatted with {poolConfig.filesystem.toUpperCase()}</li>
            <li>Any existing data on the target will be erased</li>
            <li>The pool will be automatically mounted and added to fstab</li>
          </ul>
        </div>
      </div>
    </div>
  </div>

  <!-- Actions -->
  <div class="px-6 py-4 bg-gray-50 dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 flex items-center justify-between">
    <button
      on:click={handlePrevious}
      class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"
    >
      ← Previous
    </button>

    <div class="flex items-center space-x-3">
      <p class="text-sm text-gray-600 dark:text-gray-400">
        {poolConfig.name
          ? `Pool "${poolConfig.name}" at ${mountPoint}`
          : 'Enter pool name to continue'}
      </p>

      <button
        on:click={handleNext}
        disabled={!poolConfig.name}
        class="px-6 py-2 text-sm font-medium text-white bg-indigo-600 rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        Next Step →
      </button>
    </div>
  </div>
</div>
