<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { onMount } from 'svelte';
  import { systemAPI } from '$lib/api.js';

  let systemStats = null;
  let loading = true;
  let error = null;

  async function loadSystemStats() {
    try {
      // The overview endpoint returns separate stats, we need to get them individually
      const [memStats, cpuStats] = await Promise.all([
        systemAPI.getMemoryStats().catch(() => null),
        systemAPI.getCPUStats().catch(() => null)
      ]);

      systemStats = {
        memory: memStats,
        cpu: cpuStats
      };
    } catch (err) {
      console.error('Failed to load system stats:', err);
      error = err.message;
    } finally {
      loading = false;
    }
  }

  function formatUptime(seconds) {
    if (!seconds || seconds <= 0) return 'Unknown';
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);

    if (days > 0) {
      return `${days}d ${hours}h ${minutes}m`;
    } else if (hours > 0) {
      return `${hours}h ${minutes}m`;
    } else {
      return `${minutes}m`;
    }
  }

  function getLoadColor(value) {
    if (!value || value < 1) return 'text-green-600 dark:text-green-400';
    if (value < 2) return 'text-yellow-600 dark:text-yellow-400';
    return 'text-red-600 dark:text-red-400';
  }

  onMount(() => {
    loadSystemStats();
    const interval = setInterval(loadSystemStats, 5000);
    return () => clearInterval(interval);
  });
</script>

<div class="card">
  <div class="flex items-center justify-between mb-4">
    <h3 class="text-lg font-semibold text-gray-900 dark:text-white">System Status</h3>
    <div class="flex items-center space-x-2">
      <div class="flex items-center space-x-1">
        <div class="w-2 h-2 rounded-full {loading ? 'bg-yellow-400 animate-pulse' : error ? 'bg-red-500' : 'bg-green-500'}"></div>
        <span class="text-xs text-gray-500 dark:text-gray-400">
          {loading ? 'Loading...' : error ? 'Error' : 'Online'}
        </span>
      </div>
    </div>
  </div>

  {#if error}
    <div class="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
      <p class="text-sm text-red-800 dark:text-red-200">Failed to load system status</p>
    </div>
  {:else if loading && !systemStats}
    <div class="flex justify-center items-center py-8">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>
  {:else}
    <div class="space-y-4">
      <!-- CPU Load -->
      <div class="flex items-center justify-between py-2 border-b border-gray-100 dark:border-border">
        <div class="flex items-center space-x-3">
          <div class="p-2 bg-purple-100 dark:bg-purple-900/40 rounded-lg">
            <svg class="h-4 w-4 text-purple-600 dark:text-purple-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
            </svg>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-900 dark:text-white">CPU Load Average</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">1min / 5min / 15min</p>
          </div>
        </div>
        <p class="text-sm font-mono {getLoadColor(systemStats?.cpu?.load?.[0])}">
          {systemStats?.cpu?.load ? systemStats.cpu.load.map(l => l?.toFixed(2) || '0.00').join(' / ') : 'N/A'}
        </p>
      </div>

      <!-- Memory Usage -->
      <div class="flex items-center justify-between py-2 border-b border-gray-100 dark:border-border">
        <div class="flex items-center space-x-3">
          <div class="p-2 bg-green-100 dark:bg-green-900/40 rounded-lg">
            <svg class="h-4 w-4 text-green-600 dark:text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
            </svg>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-900 dark:text-white">Memory Usage</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">RAM utilization</p>
          </div>
        </div>
        <div class="text-right">
          <p class="text-sm font-semibold text-gray-700 dark:text-gray-300">
            {systemStats?.memory?.usage_percent !== undefined ? (systemStats.memory.usage_percent).toFixed(1) : 'N/A'}%
          </p>
          <p class="text-xs text-gray-500 dark:text-gray-400">
            {#if systemStats?.memory?.used && systemStats?.memory?.total}
              {(systemStats.memory.used / (1024 * 1024 * 1024)).toFixed(1)} GB / {(systemStats.memory.total / (1024 * 1024 * 1024)).toFixed(1)} GB
            {:else}
              N/A
            {/if}
          </p>
        </div>
      </div>

      <!-- Storage Pools Quick Link -->
      <div class="py-2">
        <div class="flex items-center space-x-3 mb-3">
          <div class="p-2 bg-indigo-100 dark:bg-indigo-900/40 rounded-lg">
            <svg class="h-4 w-4 text-indigo-600 dark:text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
            </svg>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-900 dark:text-white">Storage Management</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">Configure disks, RAID, and pools</p>
          </div>
        </div>
        <a href="/storage" class="block p-3 bg-gray-50 dark:bg-muted rounded-lg hover:bg-gray-100 dark:hover:bg-muted transition-colors">
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-2">
              <svg class="h-4 w-4 text-gray-500 dark:text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
              <span class="text-sm text-gray-700 dark:text-gray-300">Manage storage</span>
            </div>
          </div>
        </a>
      </div>
    </div>
  {/if}
</div>
