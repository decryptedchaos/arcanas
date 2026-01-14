<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { systemAPI } from "$lib/api.js";
  import { formatBytes } from "$lib/utils/byteUtils.js";
  import { onMount } from "svelte";

  let stats = {
    cpu: { usage: 0, cores: 0 },
    memory: {
      used: 0,
      total: 0,
      percentage: 0,
      usedFormatted: "0.0",
      totalFormatted: "0.0",
    },
    storage: {
      used: 0,
      total: 0,
      percentage: 0,
      usedFormatted: "0.0",
      totalFormatted: "0.0",
    },
    network: {
      rx: 0,
      tx: 0,
      rxFormatted: "0.0",
      txFormatted: "0.0",
      rxRate: 0,
      txRate: 0,
      rxRateFormatted: "0.0",
      txRateFormatted: "0.0",
    },
    uptime: "0 days, 0 hours",
    temperature: 0,
    services: { running: 0, total: 3 }, // Total of 3 NAS services: SCSI, Samba, NFS
  };
  let loading = true;
  let error = null;

  async function loadStats() {
    try {
      // Only show loading on initial load, not on refresh
      if (loading) {
        error = null;
      }
      const systemData = await systemAPI.getOverview();

      // Update stats reactively by updating each property
      stats.cpu.usage = Math.round((systemData?.cpu?.usage || 0) * 10) / 10;
      stats.cpu.cores = systemData?.cpu?.cores || 0;

      // Convert bytes using utility
      const memoryUsedFormatted = formatBytes(systemData?.memory?.used || 0);
      const memoryTotalFormatted = formatBytes(systemData?.memory?.total || 0);
      stats.memory.usedFormatted = memoryUsedFormatted;
      stats.memory.totalFormatted = memoryTotalFormatted;
      stats.memory.percentage = Math.round(systemData?.memory?.usage || 0);

      // Convert storage bytes using utility
      const storageUsedBytes =
        systemData?.storage?.disks?.reduce(
          (acc, disk) => acc + (disk?.used || 0),
          0,
        ) || 0;
      const storageTotalBytes =
        systemData?.storage?.disks?.reduce(
          (acc, disk) => acc + (disk?.size || 0),
          0,
        ) || 0;
      const storageUsedFormatted = formatBytes(storageUsedBytes);
      const storageTotalFormatted = formatBytes(storageTotalBytes);
      stats.storage.usedFormatted = storageUsedFormatted;
      stats.storage.totalFormatted = storageTotalFormatted;
      stats.storage.percentage =
        storageTotalBytes > 0
          ? Math.round((storageUsedBytes / storageTotalBytes) * 100)
          : 0;

      // Convert network bytes using utility
      const networkRxBytes = systemData?.network?.total_rx || 0;
      const networkTxBytes = systemData?.network?.total_tx || 0;
      const networkRxFormatted = formatBytes(networkRxBytes);
      const networkTxFormatted = formatBytes(networkTxBytes);
      stats.network.rxFormatted = networkRxFormatted;
      stats.network.txFormatted = networkTxFormatted;
      // Convert network rates using utility (rates are in bytes/sec)
      const networkRxRateBytes = systemData?.network?.rx_rate || 0;
      const networkTxRateBytes = systemData?.network?.tx_rate || 0;
      const networkRxRateFormatted = formatBytes(networkRxRateBytes) + "/s";
      const networkTxRateFormatted = formatBytes(networkTxRateBytes) + "/s";
      stats.network.rxRateFormatted = networkRxRateFormatted;
      stats.network.txRateFormatted = networkTxRateFormatted;

      stats.uptime = formatUptime(systemData?.system?.uptime || 0);
      stats.temperature = Math.round(systemData?.cpu?.temperature || 0);

      // Check status of NAS services (SCSI, Samba, NFS)
      // For now, we'll simulate checking these services
      // In a real implementation, you'd check actual service status
      const nasServices = ["scsi-target", "smbd", "nfs-server"];
      let runningServices = 0;

      // This would be replaced with actual service status checks
      // For now, assume all services are running
      runningServices = nasServices.length;

      stats.services.running = runningServices;
      stats.services.total = nasServices.length;

      console.log("Stats updated at:", new Date().toLocaleTimeString()); // Debug log
    } catch (err) {
      error = err.message;
      console.error("Failed to load dashboard stats:", err);
    } finally {
      loading = false;
    }
  }

  function formatUptime(seconds) {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    return `${days} days, ${hours} hours`;
  }

  onMount(() => {
    loadStats();
    // Refresh stats every 5 seconds
    const interval = setInterval(loadStats, 5000);
    return () => clearInterval(interval);
  });
</script>

<!-- Dashboard Stats - Always Rendered -->
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
  <!-- CPU Usage -->
  <div class="stat-card group">
    <div class="flex items-center justify-between">
      <div class="flex-1">
        <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
          CPU Usage
        </p>
        <p class="text-2xl font-bold text-gray-900 dark:text-white">
          {stats?.cpu?.usage || 0}%
        </p>
        <p class="text-xs text-gray-500 dark:text-gray-400">
          {stats?.cpu?.cores || 0} cores
        </p>
      </div>
      <div
        class="p-3 bg-blue-100 dark:bg-blue-900/20 rounded-full group-hover:scale-110 transition-transform duration-200"
      >
        <svg
          class="h-6 w-6 text-blue-600 dark:text-blue-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"
          />
        </svg>
      </div>
    </div>
    <div class="mt-4">
      <div
        class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2 overflow-hidden"
      >
        <div
          class="h-full bg-gradient-to-r from-blue-500 to-blue-600 rounded-full transition-all duration-500 ease-out"
          style="width: {stats?.cpu?.usage || 0}%"
        ></div>
      </div>
    </div>
  </div>

  <!-- Memory Usage -->
  <div class="stat-card group">
    <div class="flex items-center justify-between">
      <div class="flex-1">
        <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
          Memory
        </p>
        <p class="text-2xl font-bold text-gray-900 dark:text-white">
          {stats?.memory?.percentage || 0}%
        </p>
        <p class="text-xs text-gray-500 dark:text-gray-400">
          {stats?.memory?.usedFormatted || "0 B"} / {stats?.memory
            ?.totalFormatted || "0 B"}
        </p>
      </div>
      <div
        class="p-3 bg-emerald-100 dark:bg-emerald-900/20 rounded-full group-hover:scale-110 transition-transform duration-200"
      >
        <svg
          class="h-6 w-6 text-emerald-600 dark:text-emerald-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z"
          />
        </svg>
      </div>
    </div>
    <div class="mt-4">
      <div
        class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2 overflow-hidden"
      >
        <div
          class="h-full bg-gradient-to-r from-emerald-500 to-emerald-600 rounded-full transition-all duration-500 ease-out"
          style="width: {stats?.memory?.percentage || 0}%"
        ></div>
      </div>
    </div>
  </div>

  <!-- Storage Usage -->
  <div class="stat-card group">
    <div class="flex items-center justify-between">
      <div class="flex-1">
        <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
          Storage
        </p>
        <p class="text-2xl font-bold text-gray-900 dark:text-white">
          {stats?.storage?.percentage || 0}%
        </p>
        <p class="text-xs text-gray-500 dark:text-gray-400">
          {stats?.storage?.usedFormatted || "0 B"} / {stats?.storage
            ?.totalFormatted || "0 B"}
        </p>
      </div>
      <div
        class="p-3 bg-purple-100 dark:bg-purple-900/20 rounded-full group-hover:scale-110 transition-transform duration-200"
      >
        <svg
          class="h-6 w-6 text-purple-600 dark:text-purple-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"
          />
        </svg>
      </div>
    </div>
    <div class="mt-4">
      <div
        class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2 overflow-hidden"
      >
        <div
          class="h-full bg-gradient-to-r from-purple-500 to-purple-600 rounded-full transition-all duration-500 ease-out"
          style="width: {stats?.storage?.percentage || 0}%"
        ></div>
      </div>
    </div>
  </div>

  <!-- Network Traffic -->
  <div class="stat-card">
    <div class="flex items-center justify-between">
      <div>
        <p class="text-sm font-medium text-gray-600 dark:text-gray-300">
          Network
        </p>
        <p class="text-2xl font-bold text-gray-900 dark:text-white">
          {stats?.network?.rxFormatted || "0 B"}
        </p>
      </div>
      <div class="p-3 bg-orange-100 rounded-full">
        <svg
          class="h-6 w-6 text-orange-600"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M13 10V3L4 14h7v7l9-11h-7z"
          />
        </svg>
      </div>
    </div>
    <div class="mt-4">
      <div class="flex space-x-2 mb-1">
        <div class="flex-1 text-xs text-gray-500 dark:text-gray-400">
          ↓ {stats?.network?.rxRateFormatted || "0 B/s"}
        </div>
        <div class="flex-1 text-xs text-gray-500 dark:text-gray-400 text-right">
          ↑ {stats?.network?.txRateFormatted || "0 B/s"}
        </div>
      </div>
      <div class="flex space-x-2">
        <div class="flex-1">
          <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
            <div
              class="bg-orange-500 h-2 rounded-full transition-all duration-500"
              style="width: {Math.min(
                (stats?.network?.rxRate || 0) / 10,
                100,
              )}%"
            ></div>
          </div>
          <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">
            Download
          </div>
        </div>
        <div class="flex-1">
          <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
            <div
              class="bg-orange-400 h-2 rounded-full transition-all duration-500"
              style="width: {Math.min(
                (stats?.network?.txRate || 0) / 10,
                100,
              )}%"
            ></div>
          </div>
          <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">
            Upload
          </div>
        </div>
      </div>
    </div>
  </div>
</div>

<!-- Additional Info Row -->
<div class="grid grid-cols-1 md:grid-cols-3 gap-6 mt-6">
  <!-- System Uptime -->
  <div class="stat-card">
    <div class="flex items-center">
      <div class="p-2 bg-gray-100 rounded-lg mr-3">
        <svg
          class="h-5 w-5 text-gray-600 dark:text-gray-300"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
      </div>
      <div>
        <p class="text-sm font-medium text-gray-600 dark:text-gray-300">
          Uptime
        </p>
        <p class="text-lg font-semibold text-gray-900 dark:text-white">
          {stats.uptime}
        </p>
      </div>
    </div>
  </div>

  <!-- Temperature -->
  <div class="stat-card">
    <div class="flex items-center">
      <div class="p-2 bg-red-100 rounded-lg mr-3">
        <svg
          class="h-5 w-5 text-red-600"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
          />
        </svg>
      </div>
      <div>
        <p class="text-sm font-medium text-gray-600 dark:text-gray-300">
          Temperature
        </p>
        <p class="text-lg font-semibold text-gray-900 dark:text-white">
          {stats.temperature}°C
        </p>
      </div>
    </div>
  </div>

  <!-- Services Status -->
  <div class="stat-card">
    <div class="flex items-center">
      <div class="p-2 bg-green-100 rounded-lg mr-3">
        <svg
          class="h-5 w-5 text-green-600"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
      </div>
      <div>
        <p class="text-sm font-medium text-gray-600 dark:text-gray-300">
          Services
        </p>
        <p class="text-lg font-semibold text-gray-900 dark:text-white">
          {stats?.services?.running || 0}/{stats?.services?.total || 0} online
        </p>
      </div>
    </div>
  </div>
</div>
