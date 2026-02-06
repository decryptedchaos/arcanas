<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { createEventDispatcher } from 'svelte';

  export let selectedDisks = [];
  export let raidConfig = {
    enabled: false,
    level: null,
    name: ''
  };

  const dispatch = createEventDispatcher();

  const raidLevels = [
    {
      value: 'raid0',
      name: 'RAID 0 (Striping)',
      description: 'Fastest performance, no redundancy. All disk capacity is used.',
      minDisks: 2,
      redundancy: 'None',
      performance: 'Excellent',
      capacity: '100%'
    },
    {
      value: 'raid1',
      name: 'RAID 1 (Mirroring)',
      description: 'Excellent redundancy, slower writes. Capacity is 50%.',
      minDisks: 2,
      redundancy: 'Excellent',
      performance: 'Good read, Fair write',
      capacity: '50%'
    },
    {
      value: 'raid5',
      name: 'RAID 5',
      description: 'Good balance of performance and redundancy. One disk can fail.',
      minDisks: 3,
      redundancy: 'Good (1 disk)',
      performance: 'Good',
      capacity: `${selectedDisks.length > 0 ? Math.round((1 - 1 / selectedDisks.length) * 100) : 67}%`
    },
    {
      value: 'raid6',
      name: 'RAID 6',
      description: 'Excellent redundancy. Two disks can fail.',
      minDisks: 4,
      redundancy: 'Excellent (2 disks)',
      performance: 'Good',
      capacity: `${selectedDisks.length > 0 ? Math.round((1 - 2 / selectedDisks.length) * 100) : 50}%`
    },
    {
      value: 'raid10',
      name: 'RAID 10',
      description: 'Best of both worlds. Requires even number of disks.',
      minDisks: 4,
      redundancy: 'Excellent',
      performance: 'Excellent',
      capacity: '50%'
    }
  ];

  function toggleRAID(enabled) {
    raidConfig = {
      ...raidConfig,
      enabled
    };

    // Set default RAID level if enabling
    if (enabled && !raidConfig.level) {
      if (selectedDisks.length === 2) {
        raidConfig.level = 'raid1';
      } else if (selectedDisks.length >= 3) {
        raidConfig.level = 'raid5';
      }
    }

    dispatch('raidChange', raidConfig);
  }

  function selectRAIDLevel(level) {
    raidConfig = {
      ...raidConfig,
      level
    };
    dispatch('raidChange', raidConfig);
  }

  function handleNext() {
    dispatch('next');
  }

  function handlePrevious() {
    dispatch('previous');
  }

  function handleSkip() {
    raidConfig = {
      enabled: false,
      level: null,
      name: ''
    };
    dispatch('next');
  }

  function getAvailableRAIDLevels() {
    return raidLevels.filter((level) => selectedDisks.length >= level.minDisks);
  }

  function getCapacityForLevel(level) {
    if (level.value === 'raid0') return '100%';
    if (level.value === 'raid1' || level.value === 'raid10') return '50%';
    if (level.value === 'raid5') return `${Math.round((1 - 1 / selectedDisks.length) * 100)}%`;
    if (level.value === 'raid6') return `${Math.round((1 - 2 / selectedDisks.length) * 100)}%`;
    return 'Unknown';
  }

  $: availableRAIDLevels = getAvailableRAIDLevels();
  $: canUseRAID = selectedDisks.length >= 2;
</script>

<div class="bg-white dark:bg-card rounded-lg shadow-lg overflow-hidden">
  <!-- Step Header -->
  <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-xl font-semibold text-gray-900 dark:text-white">
          RAID Configuration
        </h2>
        <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
          {#if raidConfig.enabled}
            Configure your RAID array for {selectedDisks.length} disk{selectedDisks.length !== 1 ? 's' : ''}
          {:else}
            Optionally create a RAID array from your {selectedDisks.length} disk{selectedDisks.length !== 1 ? 's' : ''}
          {/if}
        </p>
      </div>
      <div class="flex items-center space-x-2">
        <span class="text-sm text-gray-600 dark:text-gray-400">
          {selectedDisks.length} disk{selectedDisks.length !== 1 ? 's' : ''} selected
        </span>
      </div>
    </div>
  </div>

  <!-- RAID Toggle -->
  <div class="p-6 border-b border-gray-200 dark:border-gray-700">
    <div class="flex items-center justify-between">
      <div>
        <h3 class="text-lg font-medium text-gray-900 dark:text-white">
          Create RAID Array
        </h3>
        <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
          RAID provides data redundancy and/or performance improvements
        </p>
      </div>
      <button
        on:click={() => toggleRAID(!raidConfig.enabled)}
        type="button"
        class="relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2
          {raidConfig.enabled ? 'bg-indigo-600' : 'bg-gray-200 dark:bg-gray-700'}"
        role="switch"
        aria-checked={raidConfig.enabled}
        disabled={!canUseRAID}
      >
        <span
          aria-hidden="true"
          class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out
            {raidConfig.enabled ? 'translate-x-5' : 'translate-x-0'}"
        ></span>
      </button>
    </div>

    {#if !canUseRAID}
      <p class="mt-2 text-sm text-yellow-600 dark:text-yellow-400">
        ⚠️ RAID requires at least 2 disks
      </p>
    {/if}
  </div>

  <!-- RAID Level Selection -->
  {#if raidConfig.enabled && canUseRAID}
    <div class="p-6">
      <h4 class="text-sm font-medium text-gray-900 dark:text-white mb-4">
        Select RAID Level
      </h4>

      <div class="grid grid-cols-1 gap-4">
        {#each availableRAIDLevels as level}
          {@const isSelected = raidConfig.level === level.value}
          {@const capacity = getCapacityForLevel(level)}

          <button
            on:click={() => selectRAIDLevel(level.value)}
            type="button"
            class="relative p-4 rounded-lg border-2 text-left transition-all
              {isSelected
                ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-900/20'
                : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'}"
          >
            <!-- Selection indicator -->
            <div class="absolute top-4 right-4">
              <div
                class="w-5 h-5 rounded-full border-2 flex items-center justify-center
                  {isSelected
                    ? 'border-indigo-500 bg-indigo-500'
                    : 'border-gray-300 dark:border-gray-600'}"
              >
                {#if isSelected}
                  <svg class="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
                  </svg>
                {/if}
              </div>
            </div>

            <!-- RAID Info -->
            <div class="space-y-2">
              <div>
                <h5 class="font-semibold text-gray-900 dark:text-white">{level.name}</h5>
                <p class="text-sm text-gray-600 dark:text-gray-400">{level.description}</p>
              </div>

              <div class="flex flex-wrap gap-3 text-xs">
                <div class="flex items-center space-x-1">
                  <span class="text-gray-500 dark:text-gray-400">Redundancy:</span>
                  <span class="font-medium
                    {level.redundancy === 'None' ? 'text-red-600 dark:text-red-400' : level.redundancy === 'Excellent' ? 'text-green-600 dark:text-green-400' : 'text-yellow-600 dark:text-yellow-400'}">
                    {level.redundancy}
                  </span>
                </div>
                <div class="flex items-center space-x-1">
                  <span class="text-gray-500 dark:text-gray-400">Performance:</span>
                  <span class="font-medium text-gray-900 dark:text-white">{level.performance}</span>
                </div>
                <div class="flex items-center space-x-1">
                  <span class="text-gray-500 dark:text-gray-400">Capacity:</span>
                  <span class="font-medium text-gray-900 dark:text-white">{capacity}</span>
                </div>
              </div>
            </div>
          </button>
        {/each}
      </div>

      <!-- RAID Name -->
      {#if raidConfig.level}
        <div class="mt-6">
          <label for="raid-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            RAID Array Name
          </label>
          <input
            id="raid-name"
            type="text"
            bind:value={raidConfig.name}
            class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md dark:bg-gray-800 dark:text-white"
            placeholder="e.g., md0"
            pattern="md[0-9]+"
            title="Use format: md0, md1, etc."
          />
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
            The array will be created as /dev/{raidConfig.name || 'md0'}
          </p>
        </div>
      {/if}
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
      {#if canUseRAID && !raidConfig.enabled}
        <button
          on:click={handleSkip}
          class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"
        >
          Skip RAID →
        </button>
      {:else}
        <p class="text-sm text-gray-600 dark:text-gray-400">
          {#if !raidConfig.enabled}
            RAID disabled - using individual disks
          {:else if !raidConfig.level}
            Select a RAID level to continue
          {:else}
            {raidConfig.level.toUpperCase()} selected
          {/if}
        </p>
      {/if}

      <button
        on:click={handleNext}
        disabled={raidConfig.enabled && !raidConfig.level}
        class="px-6 py-2 text-sm font-medium text-white bg-indigo-600 rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        Next Step →
      </button>
    </div>
  </div>
</div>
