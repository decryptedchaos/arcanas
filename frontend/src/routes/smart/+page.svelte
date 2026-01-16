<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { smartAPI } from '$lib/api.js';
  import { onMount } from 'svelte';
  import Gauge from '$lib/components/Gauge.svelte';

  let disks = [];
  let loading = true;
  let error = null;
  let expandedDisk = null;
  let selectedDiskDetails = null;
  let showTestModal = false;
  let testType = 'short';

  onMount(async () => {
    await loadDisks();
  });

  async function loadDisks() {
    try {
      loading = true;
      error = null;
      disks = await smartAPI.getAllStatus();
    } catch (err) {
      error = err.message || 'Failed to load SMART data';
      console.error('Error loading SMART data:', err);
    } finally {
      loading = false;
    }
  }

  function getStatusColor(status) {
    switch (status.toLowerCase()) {
      case 'healthy':
      case 'passed':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
      case 'warning':
      case 'caution':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200';
      case 'failed':
      case 'error':
        return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200';
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200';
    }
  }

  function getTempColor(temp) {
    if (temp >= 60) return 'text-red-600 dark:text-red-400';
    if (temp >= 50) return 'text-yellow-600 dark:text-yellow-400';
    return 'text-green-600 dark:text-green-400';
  }

  async function toggleDetails(disk) {
    if (expandedDisk === disk.device) {
      expandedDisk = null;
      selectedDiskDetails = null;
    } else {
      expandedDisk = disk.device;
      await loadDiskDetails(disk);
    }
  }

  async function loadDiskDetails(disk) {
    try {
      const [attributes, errors] = await Promise.all([
        smartAPI.getAttributes(disk.device),
        smartAPI.getErrors(disk.device)
      ]);
      selectedDiskDetails = {
        ...disk,
        attributes,
        errors
      };
    } catch (err) {
      console.error('Error loading disk details:', err);
      selectedDiskDetails = disk;
    }
  }

  async function runTest() {
    if (!expandedDisk) return;

    try {
      await smartAPI.runTest(expandedDisk, testType);
      alert(`SMART ${testType} test started on ${expandedDisk}`);
      showTestModal = false;
      await loadDisks(); // Refresh data
    } catch (err) {
      alert('Failed to start test: ' + err.message);
    }
  }

  function getHealthLabel(health) {
    if (health >= 90) return 'Excellent';
    if (health >= 70) return 'Good';
    if (health >= 50) return 'Fair';
    if (health >= 30) return 'Poor';
    return 'Critical';
  }
</script>

<div class="p-6">
  <div class="mb-6">
    <h1 class="text-xl font-bold text-gray-900 dark:text-white mb-2">
      SMART Status
    </h1>
    <p class="text-sm text-gray-600 dark:text-gray-300">
      Monitor disk health and manage SMART self-tests
    </p>
  </div>

  {#if loading}
    <div class="flex justify-center items-center h-64">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>
  {:else if error}
    <div class="bg-red-50 border border-red-200 rounded-md p-4">
      <div class="flex">
        <div class="flex-shrink-0">
          <svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
          </svg>
        </div>
        <div class="ml-3">
          <h3 class="text-sm font-medium text-red-800">Error</h3>
          <div class="mt-2 text-sm text-red-700">{error}</div>
        </div>
      </div>
    </div>
  {:else}
    <!-- Overview Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <div class="card">
        <div class="text-sm text-gray-500 dark:text-gray-400">Total Disks</div>
        <div class="text-2xl font-bold text-gray-900 dark:text-white">{disks.length}</div>
      </div>
      <div class="card">
        <div class="text-sm text-gray-500 dark:text-gray-400">Healthy</div>
        <div class="text-2xl font-bold text-green-600 dark:text-green-400">
          {disks.filter(d => d.status === 'healthy').length}
        </div>
      </div>
      <div class="card">
        <div class="text-sm text-gray-500 dark:text-gray-400">Warning</div>
        <div class="text-2xl font-bold text-yellow-600 dark:text-yellow-400">
          {disks.filter(d => d.status === 'warning' || d.status === 'caution').length}
        </div>
      </div>
      <div class="card">
        <div class="text-sm text-gray-500 dark:text-gray-400">Failed</div>
        <div class="text-2xl font-bold text-red-600 dark:text-red-400">
          {disks.filter(d => d.status === 'failed').length}
        </div>
      </div>
    </div>

    <!-- Disk List -->
    <div class="space-y-4">
      {#each disks as disk (disk.device)}
        <div class="card">
          <div class="flex items-start justify-between">
            <!-- Basic Info -->
            <div class="flex items-center space-x-4 flex-1">
              <!-- Health Gauge -->
              <div class="flex-shrink-0">
                <Gauge
                  value={disk.health}
                  max={100}
                  color={disk.status === 'healthy' ? '#10B981' : disk.status === 'warning' ? '#F59E0B' : '#EF4444'}
                  label={getHealthLabel(disk.health)}
                  valueFormatter={(v) => Math.round(v).toString()}
                />
              </div>

              <!-- Device Details -->
              <div class="flex-1 min-w-0">
                <div class="flex items-center space-x-2 mb-1">
                  <h3 class="text-lg font-semibold text-gray-900 dark:text-white truncate">
                    {disk.device}
                  </h3>
                  <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium {getStatusColor(disk.status)}">
                    {disk.status}
                  </span>
                  {#if !disk.enabled}
                    <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400">
                      SMART Disabled
                    </span>
                  {/if}
                </div>

                {#if disk.model}
                  <p class="text-sm text-gray-600 dark:text-gray-300 truncate">
                    {disk.model}
                    {#if disk.serial}
                      <span class="text-gray-400">• S/N: {disk.serial}</span>
                    {/if}
                  </p>
                {/if}

                <div class="mt-2 flex items-center space-x-4 text-sm">
                  <div class="flex items-center space-x-1">
                    <span class="text-gray-500 dark:text-gray-400">Temperature:</span>
                    <span class="font-medium {getTempColor(disk.temperature)}">
                      {disk.temperature}°C
                    </span>
                  </div>
                  {#if disk.power_on_hours > 0}
                    <div class="flex items-center space-x-1">
                      <span class="text-gray-500 dark:text-gray-400">Power On:</span>
                      <span class="font-medium text-gray-900 dark:text-white">
                        {(disk.power_on_hours / 24).toFixed(0)} days
                      </span>
                    </div>
                  {/if}
                  {#if disk.power_cycles > 0}
                    <div class="flex items-center space-x-1">
                      <span class="text-gray-500 dark:text-gray-400">Cycles:</span>
                      <span class="font-medium text-gray-900 dark:text-white">
                        {disk.power_cycles}
                      </span>
                    </div>
                  {/if}
                </div>
              </div>

              <!-- Actions -->
              <div class="flex items-center space-x-2">
                <button
                  on:click={() => toggleDetails(disk)}
                  class="btn btn-secondary"
                >
                  {expandedDisk === disk.device ? 'Hide' : 'Details'}
                </button>
              </div>
            </div>
          </div>

          <!-- Expandable Details -->
          {#if expandedDisk === disk.device && selectedDiskDetails}
            <div class="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
              <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                <!-- SMART Attributes -->
                <div class="md:col-span-2">
                  <h4 class="text-sm font-semibold text-gray-900 dark:text-white mb-3">
                    SMART Attributes
                  </h4>
                  {#if selectedDiskDetails.attributes && selectedDiskDetails.attributes.length > 0}
                    <div class="overflow-x-auto">
                      <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                        <thead class="bg-gray-50 dark:bg-gray-800">
                          <tr>
                            <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-300">ID</th>
                            <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-300">Attribute</th>
                            <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-300">Value</th>
                            <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-300">Worst</th>
                            <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-300">Threshold</th>
                            <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-300">Raw</th>
                          </tr>
                        </thead>
                        <tbody class="bg-white dark:bg-card divide-y divide-gray-200 dark:divide-gray-700">
                          {#each selectedDiskDetails.attributes.slice(0, 10) as attr}
                            <tr class:color-red-500={attr.failed}>
                              <td class="px-3 py-2 text-sm text-gray-900 dark:text-white">{attr.id}</td>
                              <td class="px-3 py-2 text-sm text-gray-900 dark:text-white">{attr.name}</td>
                              <td class="px-3 py-2 text-sm text-gray-900 dark:text-white">{attr.value}</td>
                              <td class="px-3 py-2 text-sm text-gray-900 dark:text-white">{attr.worst}</td>
                              <td class="px-3 py-2 text-sm text-gray-900 dark:text-white">{attr.threshold}</td>
                              <td class="px-3 py-2 text-sm text-gray-900 dark:text-white font-mono">{attr.raw_value}</td>
                            </tr>
                          {/each}
                        </tbody>
                      </table>
                    </div>
                  {:else}
                    <p class="text-sm text-gray-500 dark:text-gray-400">No attributes available</p>
                  {/if}
                </div>

                <!-- Self-Tests & Actions -->
                <div>
                  <h4 class="text-sm font-semibold text-gray-900 dark:text-white mb-3">
                    Self-Tests ({disk.passed_tests} passed, {disk.failed_tests} failed)
                  </h4>

                  <div class="mb-4 space-y-2">
                    <button
                      on:click={() => { testType = 'short'; showTestModal = true; }}
                      class="w-full btn btn-secondary text-sm"
                    >
                      Run Short Test
                    </button>
                    <button
                      on:click={() => { testType = 'long'; showTestModal = true; }}
                      class="w-full btn btn-secondary text-sm"
                    >
                      Run Long Test
                    </button>
                  </div>

                  {#if disk.self_tests && disk.self_tests.length > 0}
                    <div class="border-t border-gray-200 dark:border-gray-700 pt-3">
                      <h5 class="text-xs font-medium text-gray-500 dark:text-gray-400 mb-2">Recent Tests</h5>
                      <div class="space-y-1">
                        {#each disk.self_tests.slice(0, 3) as test}
                          <div class="text-xs">
                            <span class="font-medium text-gray-900 dark:text-white">{test.type}</span>
                            <span class="text-gray-500 dark:text-gray-400 mx-1">•</span>
                            <span class="text-gray-600 dark:text-gray-300">{test.status}</span>
                          </div>
                        {/each}
                      </div>
                    </div>
                  {/if}
                </div>
              </div>
            </div>
          {/if}
        </div>
      {:else}
        <div class="card text-center py-12">
          <p class="text-gray-500 dark:text-gray-400">No disks found</p>
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Test Confirmation Modal -->
{#if showTestModal}
  <div class="fixed inset-0 z-50 overflow-y-auto">
    <div class="flex items-center justify-center min-h-screen px-4">
      <div
        class="fixed inset-0 bg-gray-500 bg-opacity-75"
        on:click={() => (showTestModal = false)}
      ></div>
      <div class="relative bg-white dark:bg-card rounded-lg text-left overflow-hidden shadow-xl max-w-md w-full">
        <div class="p-6">
          <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">
            Run SMART Self-Test
          </h3>
          <p class="text-sm text-gray-600 dark:text-gray-300 mb-4">
            Are you sure you want to run a <span class="font-semibold">{testType}</span> SMART test on {expandedDisk}?
          </p>
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-4">
            {testType === 'short' ? 'Short test typically takes a few minutes.' : 'Long test can take several hours to complete.'}
          </p>
          <div class="flex justify-end space-x-3">
            <button
              on:click={() => (showTestModal = false)}
              class="btn btn-secondary"
            >
              Cancel
            </button>
            <button
              on:click={runTest}
              class="btn btn-primary"
            >
              Start Test
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  .btn {
    display: inline-flex;
    align-items: center;
    padding: 0.5rem 1rem;
    border: 1px solid transparent;
    border-radius: 0.375rem;
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
    font-size: 0.875rem;
    font-weight: 500;
    color: white;
    transition: color 0.2s, background-color 0.2s, border-color 0.2s;
    outline: none;
  }

  .btn:focus {
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.5);
  }

  .btn-primary {
    background-color: rgb(37, 99, 235);
    border-color: rgb(37, 99, 235);
  }

  .btn-primary:hover {
    background-color: rgb(29, 78, 216);
  }

  .btn-secondary {
    background-color: rgb(75, 85, 99);
    border-color: rgb(75, 85, 99);
  }

  .btn-secondary:hover {
    background-color: rgb(55, 65, 81);
  }
</style>
