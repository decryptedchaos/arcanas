<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { createEventDispatcher } from 'svelte';

  export let raidConfig = {
    enabled: false,
    name: '',
    level: ''
  };
  export let selectedDisks = [];
  export let lvmConfig = {
    enabled: false,
    vgName: '',
    lvName: '',
    sizeGB: null
  };

  const dispatch = createEventDispatcher();

  // Calculate available space
  let totalSpaceGB = 0;

  function calculateAvailableSpace() {
    // This would come from actual disk sizes in a real implementation
    // For now, use a placeholder calculation
    totalSpaceGB = selectedDisks.length * 1000; // Assume 1TB per disk
  }

  function toggleLVM(enabled) {
    lvmConfig = {
      ...lvmConfig,
      enabled
    };

    // Auto-generate names if enabling
    if (enabled) {
      if (!lvmConfig.vgName) {
        lvmConfig.vgName = 'vg-' + (raidConfig.enabled ? raidConfig.name : 'storage');
      }
      if (!lvmConfig.lvName) {
        lvmConfig.lvName = 'data';
      }
      if (!lvmConfig.sizeGB) {
        lvmConfig.sizeGB = Math.round(totalSpaceGB * 0.9); // Use 90% by default
      }
    }

    dispatch('lvmChange', lvmConfig);
  }

  function handleNext() {
    dispatch('next');
  }

  function handlePrevious() {
    dispatch('previous');
  }

  function handleSkip() {
    lvmConfig = {
      enabled: false,
      vgName: '',
      lvName: '',
      sizeGB: null
    };
    dispatch('next');
  }

  calculateAvailableSpace();

  $: deviceDescription = raidConfig.enabled
    ? `RAID array (${raidConfig.level})`
    : `${selectedDisks.length} disk${selectedDisks.length !== 1 ? 's' : ''}`;

  $: availableSpaceGB = raidConfig.enabled
    ? Math.round(totalSpaceGB * (raidConfig.level === 'raid1' || raidConfig.level === 'raid10' ? 0.5 : raidConfig.level === 'raid5' ? (1 - 1 / selectedDisks.length) : raidConfig.level === 'raid6' ? (1 - 2 / selectedDisks.length) : 1))
    : totalSpaceGB;
</script>

<div class="bg-white dark:bg-card rounded-lg shadow-lg overflow-hidden">
  <!-- Step Header -->
  <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-xl font-semibold text-gray-900 dark:text-white">
          LVM Configuration
        </h2>
        <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
          {#if lvmConfig.enabled}
            Configure flexible volumes with LVM
          {:else}
            Optionally use LVM for flexible volume management
          {/if}
        </p>
      </div>
      <div class="text-sm text-gray-600 dark:text-gray-400">
        From: {deviceDescription}
      </div>
    </div>
  </div>

  <!-- LVM Explanation -->
  <div class="px-6 py-4 bg-blue-50 dark:bg-blue-900/20 border-b border-blue-100 dark:border-blue-800">
    <div class="flex items-start space-x-3">
      <svg class="w-5 h-5 text-blue-600 dark:text-blue-400 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <div class="flex-1 text-sm text-blue-800 dark:text-blue-200">
        <p class="font-medium mb-1">What is LVM?</p>
        <p class="text-blue-700 dark:text-blue-300">
          LVM (Logical Volume Manager) allows you to create flexible volumes that can be resized,
          snapped, and managed independently. Use LVM if you need to resize storage later or want
          to create multiple volumes from the same disks.
        </p>
      </div>
    </div>
  </div>

  <!-- LVM Toggle -->
  <div class="p-6 border-b border-gray-200 dark:border-gray-700">
    <div class="flex items-center justify-between">
      <div>
        <h3 class="text-lg font-medium text-gray-900 dark:text-white">
          Use LVM for Volume Management
        </h3>
        <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
          Create flexible logical volumes that can be resized
        </p>
      </div>
      <button
        on:click={() => toggleLVM(!lvmConfig.enabled)}
        type="button"
        class="relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2
          {lvmConfig.enabled ? 'bg-indigo-600' : 'bg-gray-200 dark:bg-gray-700'}"
        role="switch"
        aria-checked={lvmConfig.enabled}
      >
        <span
          aria-hidden="true"
          class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out
            {lvmConfig.enabled ? 'translate-x-5' : 'translate-x-0'}"
        ></span>
      </button>
    </div>
  </div>

  <!-- LVM Configuration -->
  {#if lvmConfig.enabled}
    <div class="p-6 space-y-6">
      <!-- Volume Group Name -->
      <div>
        <label for="vg-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
          Volume Group Name
        </label>
        <input
          id="vg-name"
          type="text"
          bind:value={lvmConfig.vgName}
          class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md dark:bg-gray-800 dark:text-white"
          placeholder="e.g., vg-storage"
          pattern="[a-z0-9\-]+"
          title="Use lowercase letters, numbers, and hyphens"
        />
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
          The VG will contain all your logical volumes
        </p>
      </div>

      <!-- Logical Volume Name -->
      <div>
        <label for="lv-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
          Logical Volume Name
        </label>
        <input
          id="lv-name"
          type="text"
          bind:value={lvmConfig.lvName}
          class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md dark:bg-gray-800 dark:text-white"
          placeholder="e.g., data"
          pattern="[a-z0-9\-]+"
          title="Use lowercase letters, numbers, and hyphens"
        />
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
          This is the volume that will be mounted as your storage pool
        </p>
      </div>

      <!-- Volume Size -->
      <div>
        <label for="lv-size" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
          Volume Size (GB)
        </label>
        <div class="flex items-center space-x-4">
          <input
            id="lv-size"
            type="number"
            bind:value={lvmConfig.sizeGB}
            min="1"
            max={availableSpaceGB}
            class="flex-1 px-3 py-2 border border-gray-300 dark:border rounded-md dark:bg-gray-800 dark:text-white"
            placeholder="Size in GB"
          />
          <div class="text-sm text-gray-600 dark:text-gray-400">
            of {Math.round(availableSpaceGB)} GB available
          </div>
        </div>
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
          You can resize this later if needed
        </p>

        <!-- Quick size options -->
        <div class="mt-2 flex flex-wrap gap-2">
          <button
            on:click={() => lvmConfig.sizeGB = Math.round(availableSpaceGB * 0.5)}
            type="button"
            class="px-3 py-1 text-xs bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-200 dark:hover:bg-gray-600"
          >
            50%
          </button>
          <button
            on:click={() => lvmConfig.sizeGB = Math.round(availableSpaceGB * 0.75)}
            type="button"
            class="px-3 py-1 text-xs bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-200 dark:hover:bg-gray-600"
          >
            75%
          </button>
          <button
            on:click={() => lvmConfig.sizeGB = Math.round(availableSpaceGB * 0.9)}
            type="button"
            class="px-3 py-1 text-xs bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-200 dark:hover:bg-gray-600"
          >
            90%
          </button>
          <button
            on:click={() => lvmConfig.sizeGB = Math.round(availableSpaceGB)}
            type="button"
            class="px-3 py-1 text-xs bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-200 dark:hover:bg-gray-600"
          >
            100%
          </button>
        </div>
      </div>

      <!-- Preview -->
      <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
        <h4 class="text-sm font-medium text-gray-900 dark:text-white mb-2">Preview</h4>
        <div class="space-y-1 text-sm text-gray-600 dark:text-gray-400">
          <p>• Volume Group: <span class="font-mono text-gray-900 dark:text-white">/dev/{lvmConfig.vgName || 'vg-storage'}</span></p>
          <p>• Logical Volume: <span class="font-mono text-gray-900 dark:text-white">/dev/{lvmConfig.vgName || 'vg-storage'}/{lvmConfig.lvName || 'data'}</span></p>
          <p>• Size: <span class="font-medium text-gray-900 dark:text-white">{lvmConfig.sizeGB || 0} GB</span></p>
        </div>
      </div>
    </div>
  {/if}

  <!-- Actions -->
  <div class="px-6 py-4 bg-gray-50 dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 flex items-center justify-between">
    <button
      on:click={handlePrevious}
      class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"
    >
      ← Previous
    </button>

    <div class="flex items-center space-x-3">
      {#if !lvmConfig.enabled}
        <button
          on:click={handleSkip}
          class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"
        >
          Skip LVM →
        </button>
      {:else}
        <p class="text-sm text-gray-600 dark:text-gray-400">
          LVM enabled - {lvmConfig.lvName || 'data'} volume
        </p>
      {/if}

      <button
        on:click={handleNext}
        disabled={lvmConfig.enabled && (!lvmConfig.vgName || !lvmConfig.lvName || !lvmConfig.sizeGB)}
        class="px-6 py-2 text-sm font-medium text-white bg-indigo-600 rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        Next Step →
      </button>
    </div>
  </div>
</div>
