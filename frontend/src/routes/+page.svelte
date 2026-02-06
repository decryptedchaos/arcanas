<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { storageAPI, diskAPI, systemAPI } from '$lib/api.js';
  import DashboardStats from '$lib/components/DashboardStats.svelte';
  import StorageBuilderCard from '$lib/components/StorageBuilderCard.svelte';
  import Gauge from '$lib/components/Gauge.svelte';
  import FirstRunBanner from '$lib/components/FirstRunBanner.svelte';

  let showFirstRunBanner = false;
  let diskIORates = null;
  let networkIORates = null;
  let diskIOHistory = [];
  let networkIOHistory = [];

  async function checkFirstRun() {
    try {
      const [pools, raidArrays] = await Promise.all([
        storageAPI.getPools(),
        diskAPI.getRAIDArrays()
      ]);

      // Show first-run if no storage configured
      showFirstRunBanner = pools.length === 0 && raidArrays.length === 0;
    } catch (err) {
      console.error('Failed to check first-run status:', err);
    }
  }

  function handleLaunchWizard() {
    goto('/storage-builder');
  }

  async function loadIOStats() {
    try {
      const [diskRates, networkRates] = await Promise.all([
        systemAPI.getDiskIORates().catch(() => null),
        systemAPI.getNetworkIORates().catch(() => null)
      ]);

      diskIORates = diskRates;
      networkIORates = networkRates;

      console.log('Network rates:', networkRates);

      // Update history for gauges
      if (diskRates) {
        diskIOHistory = [...diskIOHistory.slice(-59), {
          read: diskRates.read_rate * 1024 * 1024,
          write: diskRates.write_rate * 1024 * 1024,
          timestamp: Date.now()
        }];
      }

      if (networkRates) {
        networkIOHistory = [...networkIOHistory.slice(-59), {
          rx: networkRates.rx_rate * 1000000,
          tx: networkRates.tx_rate * 1000000,
          timestamp: Date.now()
        }];
        console.log('Network history updated:', networkIOHistory[networkIOHistory.length - 1]);
      }
    } catch (err) {
      console.error('Failed to load I/O stats:', err);
    }
  }

  function formatDataRate(rate) {
    if (!rate && rate !== 0) return '0 B/s';
    if (rate === 0 || isNaN(rate)) return '0 B/s';

    const bytesPerSecond = rate * 1024 * 1024;

    const units = ['B/s', 'KB/s', 'MB/s', 'GB/s'];
    let value = bytesPerSecond;
    let unitIndex = 0;

    while (value >= 1024 && unitIndex < units.length - 1) {
      value /= 1024;
      unitIndex++;
    }

    return `${value.toFixed(1)} ${units[unitIndex]}`;
  }

  function formatNetworkRate(rate) {
    if (!rate && rate !== 0) return '0 Mbps';
    if (rate === 0 || isNaN(rate)) return '0 Mbps';

    const mbps = rate / 1000000;
    return `${mbps.toFixed(1)} Mbps`;
  }

  onMount(() => {
    checkFirstRun();
    loadIOStats();
    const interval = setInterval(loadIOStats, 2000);
    return () => clearInterval(interval);
  });
</script>

<div class="space-y-6">
  <!-- First Run Banner -->
  <FirstRunBanner show={showFirstRunBanner} on:launch-wizard={handleLaunchWizard} />

  <!-- Header -->
  <div>
    <h2 class="text-2xl font-bold text-gray-900 dark:text-white">Dashboard Overview</h2>
    <p class="mt-1 text-sm text-gray-600 dark:text-gray-300">Monitor your NAS system status and performance</p>
  </div>

  <!-- Storage Builder Card -->
  <StorageBuilderCard />

  <!-- System Stats Grid -->
  <DashboardStats />

  <!-- I/O Gauges -->
  <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
    <!-- Disk I/O -->
    <div class="card">
      <div class="flex items-center justify-between mb-4">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Disk I/O</h3>
        <div class="flex items-center space-x-2">
          <div class="p-2 bg-indigo-100 dark:bg-indigo-900/40 rounded-lg">
            <svg class="h-4 w-4 text-indigo-600 dark:text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
            </svg>
          </div>
        </div>
      </div>
      <p class="text-xs text-gray-500 dark:text-gray-400 mb-4">Physical disk read/write rates</p>
      <div class="flex justify-around">
        <!-- Read Gauge -->
        <Gauge
          value={(diskIOHistory[diskIOHistory.length - 1]?.read ?? 0) / (1024 * 1024)}
          max={100}
          color="#10B981"
          label="Read"
          valueFormatter={(v) => formatDataRate(v)}
        />

        <!-- Write Gauge -->
        <Gauge
          value={(diskIOHistory[diskIOHistory.length - 1]?.write ?? 0) / (1024 * 1024)}
          max={100}
          color="#F59E0B"
          label="Write"
          valueFormatter={(v) => formatDataRate(v)}
        />
      </div>
    </div>

    <!-- Network I/O -->
    <div class="card">
      <div class="flex items-center justify-between mb-4">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Network I/O</h3>
        <div class="flex items-center space-x-2">
          <div class="p-2 bg-cyan-100 dark:bg-cyan-900/40 rounded-lg">
            <svg class="h-4 w-4 text-cyan-600 dark:text-cyan-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          </div>
        </div>
      </div>
      <p class="text-xs text-gray-500 dark:text-gray-400 mb-4">Network download/upload rates</p>
      <div class="flex justify-around">
        <!-- Download Gauge -->
        <Gauge
          value={(networkIOHistory[networkIOHistory.length - 1]?.rx ?? 0) / 1000000}
          max={1000}
          color="#06B6D4"
          label="Download"
          valueFormatter={(v) => formatNetworkRate(v * 1000000)}
        />

        <!-- Upload Gauge -->
        <Gauge
          value={(networkIOHistory[networkIOHistory.length - 1]?.tx ?? 0) / 1000000}
          max={1000}
          color="#10B981"
          label="Upload"
          valueFormatter={(v) => formatNetworkRate(v * 1000000)}
        />
      </div>
    </div>
  </div>
</div>
