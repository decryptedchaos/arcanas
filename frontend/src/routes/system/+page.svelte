<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { systemAPI } from "$lib/api.js";
  import { calculateScale, formatBytes } from "$lib/utils/byteUtils.js";
  import { onMount } from "svelte";

  let systemStats = null;
  let loading = true;
  let error = null;

  // History data for graphs - make them reactive
  let cpuHistory = [{ time: new Date(), value: 0 }];
  let diskIOHistory = [{ time: new Date(), read: 0, write: 0 }];
  let networkIOHistory = [{ time: new Date(), rx: 0, tx: 0 }];
  const maxHistoryPoints = 60; // 5 minutes of data at 5-second intervals

  async function loadSystemStats() {
    try {
      const [newStats, diskIORates, networkIORates] = await Promise.all([
        systemAPI.getOverview(),
        systemAPI.getDiskIORates(),
        systemAPI.getNetworkIORates(),
      ]);

      console.log("Disk I/O data:", diskIORates);
      console.log("Network I/O data:", networkIORates);

      // Update history arrays - use assignment for Svelte reactivity
      const now = new Date();
      cpuHistory = [
        ...cpuHistory,
        {
          time: now,
          value: newStats?.cpu?.usage || 0,
        },
      ];

      // Use actual disk I/O rates from backend (convert MB/s to bytes/sec for consistent scaling)
      diskIOHistory = [
        ...diskIOHistory,
        {
          time: now,
          read: (diskIORates?.read_rate || 0) * 1024 * 1024, // Convert MB/s to bytes/sec
          write: (diskIORates?.write_rate || 0) * 1024 * 1024, // Convert MB/s to bytes/sec
        },
      ];

      // Use actual network I/O rates from backend (convert Mbps to bytes/sec for consistent scaling)
      networkIOHistory = [
        ...networkIOHistory,
        {
          time: now,
          rx: (networkIORates?.rx_rate || 0) * 1000000, // Convert Mbps to bytes/sec
          tx: (networkIORates?.tx_rate || 0) * 1000000, // Convert Mbps to bytes/sec
        },
      ];

      // Debug logging
      console.log(
        "Raw disk rates - Read:",
        diskIORates?.read_rate,
        "Write:",
        diskIORates?.write_rate,
      );
      console.log(
        "Raw network rates - RX:",
        networkIORates?.rx_rate,
        "TX:",
        networkIORates?.tx_rate,
      );
      console.log(
        "Formatted disk - Read:",
        formatDataRate(diskIORates?.read_rate),
        "Write:",
        formatDataRate(diskIORates?.write_rate),
      );
      console.log(
        "Formatted network - RX:",
        formatNetworkRate(networkIORates?.rx_rate),
        "TX:",
        formatNetworkRate(networkIORates?.tx_rate),
      );

      // Keep only recent history
      if (cpuHistory.length > maxHistoryPoints)
        cpuHistory = cpuHistory.slice(-maxHistoryPoints);
      if (diskIOHistory.length > maxHistoryPoints)
        diskIOHistory = diskIOHistory.slice(-maxHistoryPoints);
      if (networkIOHistory.length > maxHistoryPoints)
        networkIOHistory = networkIOHistory.slice(-maxHistoryPoints);

      systemStats = newStats;
      error = null;
    } catch (err) {
      error = err.message;
      console.error("Failed to load system stats:", err);
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadSystemStats();
    // Refresh stats every 2 seconds for more responsive updates
    const interval = setInterval(loadSystemStats, 2000);
    return () => clearInterval(interval);
  });

  function formatDataRate(rate) {
    if (!rate && rate !== 0) return "0 B/s";
    if (rate === 0 || isNaN(rate)) return "0 B/s";

    // Backend returns MB/s, convert to bytes/s first
    const bytesPerSecond = rate * 1024 * 1024;
    return formatBytes(bytesPerSecond) + "/s";
  }

  function formatNetworkRate(rate) {
    if (!rate && rate !== 0) return "0 bps";
    if (rate === 0 || isNaN(rate)) return "0 bps";

    // Backend returns Mbps, convert to bytes/s first
    const bytesPerSecond = rate * 1000000;
    return formatBytes(bytesPerSecond) + "/s";
  }

  function formatLoad(load) {
    return load.toFixed(2);
  }

  function formatPercent(value) {
    return (value || 0).toFixed(1);
  }

  function formatTemp(temp) {
    return (temp || 0).toFixed(1);
  }

  function formatUptime(seconds) {
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

  function getLoadColor(load) {
    if (load > 2) return "text-red-600";
    if (load > 1) return "text-yellow-600";
    return "text-green-600";
  }

  function getTempColor(temp) {
    if (temp > 70) return "text-red-600";
    if (temp > 60) return "text-yellow-600";
    return "text-green-600";
  }
</script>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div>
      <h2
        class="text-xl font-bold text-gray-900 dark:text-white"
      >
        System Statistics
      </h2>
      <p
        class="text-sm text-gray-600 dark:text-gray-300 dark:text-gray-300 mt-1"
      >
        Real-time system performance monitoring
      </p>
    </div>
    <div class="flex items-center space-x-3">
      <button
        on:click={loadSystemStats}
        class="btn btn-primary"
        disabled={loading}
      >
        {loading ? "Loading..." : "Refresh"}
      </button>
    </div>
  </div>

  <!-- Error State -->
  {#if error}
    <div class="bg-red-50 border border-red-200 rounded-md p-4">
      <div class="flex">
        <div class="flex-shrink-0">
          <svg
            class="h-5 w-5 text-red-400"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fill-rule="evenodd"
              d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
              clip-rule="evenodd"
            />
          </svg>
        </div>
        <div class="ml-3">
          <h3 class="text-sm font-medium text-red-800">
            Error loading system stats
          </h3>
          <div class="mt-2 text-sm text-red-700">{error}</div>
        </div>
      </div>
    </div>
  {/if}

  <!-- Loading State -->
  {#if loading && !systemStats}
    <div class="text-center py-8">
      <div
        class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"
      ></div>
      <p class="mt-2 text-sm text-gray-600 dark:text-gray-300">
        Loading system statistics...
      </p>
    </div>
  {/if}

  <!-- System Stats -->
  {#if systemStats && !loading}
    <!-- System Info -->
    <div class="card">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
        System Information
      </h3>
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
        <div>
          <p class="text-gray-600 dark:text-gray-300">Hostname</p>
          <p class="font-medium text-gray-900 dark:text-white">
            {systemStats?.system?.hostname || "N/A"}
          </p>
        </div>
        <div>
          <p class="text-gray-600 dark:text-gray-300">Operating System</p>
          <p class="font-medium text-gray-900 dark:text-white">
            {systemStats?.system?.os || "N/A"}
          </p>
        </div>
        <div>
          <p class="text-gray-600 dark:text-gray-300">Kernel</p>
          <p class="font-medium text-gray-900 dark:text-white">
            {systemStats?.system?.kernel || "N/A"}
          </p>
        </div>
        <div>
          <p class="text-gray-600 dark:text-gray-300">Uptime</p>
          <p class="font-medium text-gray-900 dark:text-white">
            {formatUptime(systemStats?.system?.uptime || 0)}
          </p>
        </div>
      </div>
    </div>

    <!-- CPU & Memory Stats -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- CPU Stats with Integrated Graph -->
      <div class="card">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
            CPU Performance
          </h3>
          <span class="text-2xl font-bold text-gray-900 dark:text-white"
            >{formatPercent(systemStats?.cpu?.usage)}%</span
          >
        </div>

        <div class="space-y-4">
          <!-- CPU Usage Bar -->
          <div>
            <div class="w-full bg-gray-200 dark:bg-muted rounded-full h-4">
              <div
                class="bg-blue-600 h-4 rounded-full transition-all duration-500"
                style="width: {formatPercent(systemStats?.cpu?.usage)}%"
              ></div>
            </div>
          </div>

          <!-- CPU Graph -->
          <div
            class="h-32 bg-gray-50 dark:bg-muted rounded-lg p-2 relative flex"
          >
            <!-- Y-axis labels -->
            <div
              class="flex flex-col justify-between text-xs text-gray-500 mr-2 w-8"
            >
              {#each calculateScale(cpuHistory, "value").steps as step}
                <span>{step}%</span>
              {/each}
            </div>

            <!-- Graph container -->
            <div class="flex-1 relative">
              <svg class="absolute inset-0 w-full h-full">
                <!-- Grid lines -->
                <g stroke="#6B7280" stroke-width="0.5">
                  <!-- Horizontal grid lines -->
                  <line x1="0%" y1="25%" x2="100%" y2="25%" />
                  <line x1="0%" y1="50%" x2="100%" y2="50%" />
                  <line x1="0%" y1="75%" x2="100%" y2="75%" />
                  <!-- Vertical grid lines -->
                  <line x1="25%" y1="0%" x2="25%" y2="100%" />
                  <line x1="50%" y1="0%" x2="50%" y2="100%" />
                  <line x1="75%" y1="0%" x2="75%" y2="100%" />
                </g>

                {#each cpuHistory as point, i}
                  {@const cpuScale = calculateScale(cpuHistory, "value")}
                  {@const xPos = (i / cpuHistory.length) * 100 + "%"}
                  <circle
                    cx={xPos}
                    cy={100 - (point.value / cpuScale.max) * 100 + "%"}
                    r="2"
                    fill="#3B82F6"
                  />
                  {#if i > 0}
                    {@const prevXPos =
                      ((i - 1) / cpuHistory.length) * 100 + "%"}
                    <line
                      x1={prevXPos}
                      y1={100 -
                        (cpuHistory[i - 1].value / cpuScale.max) * 100 +
                        "%"}
                      x2={xPos}
                      y2={100 - (point.value / cpuScale.max) * 100 + "%"}
                      stroke="#3B82F6"
                      stroke-width="2"
                      stroke-linejoin="round"
                      stroke-linecap="round"
                    />
                  {/if}
                {/each}
              </svg>
            </div>
          </div>

          <!-- CPU Details -->
          <div class="grid grid-cols-2 gap-4 text-sm">
            <div>
              <p class="text-gray-600 dark:text-gray-300">Model</p>
              <p class="font-medium text-gray-900 dark:text-white">
                {systemStats?.cpu?.model || "Unknown"}
              </p>
            </div>
            <div>
              <p class="text-gray-600 dark:text-gray-300">Cores</p>
              <p class="font-medium text-gray-900 dark:text-white">
                {systemStats?.cpu?.cores || 0} @ {systemStats?.cpu?.frequency ||
                  "Unknown"}
              </p>
            </div>
            <div>
              <p class="text-gray-600 dark:text-gray-300">Temperature</p>
              <p
                class="font-medium {getTempColor(
                  systemStats?.cpu?.temperature || 0,
                )}"
              >
                {formatTemp(systemStats?.cpu?.temperature)}°C
              </p>
            </div>
            <div>
              <p class="text-gray-600 dark:text-gray-300">Load Average</p>
              <p class="font-medium text-gray-900 dark:text-white">
                {(systemStats?.cpu?.load_average || [])
                  .map((load) => formatLoad(load))
                  .join(", ")}
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Memory Stats -->
      <div class="card">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
            Memory Usage
          </h3>
          <span class="text-2xl font-bold text-gray-900 dark:text-white"
            >{formatPercent(systemStats?.memory?.usage)}%</span
          >
        </div>

        <div class="space-y-4">
          <!-- Memory Usage Bar -->
          <div>
            <div class="w-full bg-gray-200 dark:bg-muted rounded-full h-4">
              <div
                class="bg-green-600 h-4 rounded-full transition-all duration-500"
                style="width: {formatPercent(systemStats?.memory?.usage)}%"
              ></div>
            </div>
          </div>

          <!-- Memory Details -->
          <div class="space-y-3">
            <div class="flex justify-between text-sm">
              <span class="text-gray-600 dark:text-gray-300">Used</span>
              <span class="font-medium text-gray-900 dark:text-white"
                >{formatBytes(systemStats?.memory?.used || 0)}</span
              >
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-600 dark:text-gray-300">Available</span>
              <span class="font-medium text-gray-900 dark:text-white"
                >{formatBytes(systemStats?.memory?.available || 0)}</span
              >
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-600 dark:text-gray-300">Total</span>
              <span class="font-medium text-gray-900 dark:text-white"
                >{formatBytes(systemStats?.memory?.total || 0)}</span
              >
            </div>
          </div>

          <!-- Swap Usage -->
          <div class="pt-3 border-t border-gray-200">
            <p class="text-sm text-gray-600 dark:text-gray-300 mb-2">
              Swap Memory
            </p>
            <div class="flex justify-between text-sm">
              <span class="text-gray-600 dark:text-gray-300">Used</span>
              <span class="font-medium text-gray-900 dark:text-white"
                >{formatBytes(systemStats?.memory?.swap?.used || 0)}</span
              >
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-600 dark:text-gray-300">Total</span>
              <span class="font-medium text-gray-900 dark:text-white"
                >{formatBytes(systemStats?.memory?.swap?.total || 0)}</span
              >
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Performance Graphs -->
    <div class="space-y-6">
      <!-- Disk I/O Graph -->
      <div class="card">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
          Disk I/O
        </h3>
        <div class="space-y-4">
          <div class="flex justify-around">
            <!-- Read Gauge -->
            <div class="text-center">
              <div class="relative w-32 h-32 mx-auto">
                <svg class="w-full h-full transform -rotate-90">
                  <circle
                    cx="64"
                    cy="64"
                    r="56"
                    stroke="#E5E7EB"
                    stroke-width="8"
                    fill="none"
                    class="dark:stroke-gray-600"
                  />
                  <circle
                    cx="64"
                    cy="64"
                    r="56"
                    stroke="#10B981"
                    stroke-width="8"
                    fill="none"
                    stroke-dasharray={`${((diskIOHistory[diskIOHistory.length - 1]?.read || 0) / calculateScale(diskIOHistory, "all").max) * 351.86} 351.86`}
                    stroke-linecap="round"
                  />
                </svg>
                <div
                  class="absolute inset-0 flex flex-col items-center justify-center"
                >
                  <span class="text-lg font-bold text-gray-900 dark:text-white"
                    >{formatDataRate(
                      (diskIOHistory[diskIOHistory.length - 1]?.read ?? 0) /
                        (1024 * 1024), // Convert bytes/sec back to MB/s
                    )}</span
                  >
                  <span class="text-xs text-gray-500 dark:text-gray-400"
                    >Read</span
                  >
                </div>
              </div>
            </div>

            <!-- Write Gauge -->
            <div class="text-center">
              <div class="relative w-32 h-32 mx-auto">
                <svg class="w-full h-full transform -rotate-90">
                  <circle
                    cx="64"
                    cy="64"
                    r="56"
                    stroke="#E5E7EB"
                    stroke-width="8"
                    fill="none"
                    class="dark:stroke-gray-600"
                  />
                  <circle
                    cx="64"
                    cy="64"
                    r="56"
                    stroke="#F59E0B"
                    stroke-width="8"
                    fill="none"
                    stroke-dasharray={`${((diskIOHistory[diskIOHistory.length - 1]?.write || 0) / calculateScale(diskIOHistory, "all").max) * 351.86} 351.86`}
                    stroke-linecap="round"
                  />
                </svg>
                <div
                  class="absolute inset-0 flex flex-col items-center justify-center"
                >
                  <span class="text-lg font-bold text-gray-900 dark:text-white"
                    >{formatDataRate(
                      (diskIOHistory[diskIOHistory.length - 1]?.write ?? 0) /
                        (1024 * 1024), // Convert bytes/sec back to MB/s
                    )}</span
                  >
                  <span class="text-xs text-gray-500 dark:text-gray-400"
                    >Write</span
                  >
                </div>
              </div>
            </div>
          </div>
          <div
            class="h-32 bg-gray-50 dark:bg-muted rounded-lg p-2 relative flex"
          >
            <!-- Y-axis labels -->
            <div
              class="flex flex-col justify-between text-xs text-gray-500 mr-2 w-16"
            >
              {#each calculateScale(diskIOHistory, "all").steps as step}
                <span>{(step / (1024 * 1024)).toFixed(1)} MB/s</span>
              {/each}
            </div>

            <!-- Graph container -->
            <div class="flex-1 relative">
              <svg class="absolute inset-0 w-full h-full">
                <!-- Grid lines -->
                <g stroke="#6B7280" stroke-width="0.5">
                  <!-- Horizontal grid lines -->
                  <line x1="0%" y1="25%" x2="100%" y2="25%" />
                  <line x1="0%" y1="50%" x2="100%" y2="50%" />
                  <line x1="0%" y1="75%" x2="100%" y2="75%" />
                  <!-- Vertical grid lines -->
                  <line x1="25%" y1="0%" x2="25%" y2="100%" />
                  <line x1="50%" y1="0%" x2="50%" y2="100%" />
                  <line x1="75%" y1="0%" x2="75%" y2="100%" />
                </g>

                {#each diskIOHistory as point, i}
                  {@const diskScale = calculateScale(diskIOHistory, "all")}
                  {@const xPos = (i / diskIOHistory.length) * 100 + "%"}
                  <circle
                    cx={xPos}
                    cy={100 - (point.read / diskScale.max) * 100 + "%"}
                    r="2"
                    fill="#10B981"
                  />
                  <circle
                    cx={xPos}
                    cy={100 - (point.write / diskScale.max) * 100 + "%"}
                    r="2"
                    fill="#F59E0B"
                  />
                  {#if i > 0}
                    {@const prevXPos =
                      ((i - 1) / diskIOHistory.length) * 100 + "%"}
                    <line
                      x1={prevXPos}
                      y1={100 -
                        (diskIOHistory[i - 1].read / diskScale.max) * 100 +
                        "%"}
                      x2={xPos}
                      y2={100 - (point.read / diskScale.max) * 100 + "%"}
                      stroke="#10B981"
                      stroke-width="2"
                      stroke-linejoin="round"
                      stroke-linecap="round"
                    />
                    <line
                      x1={prevXPos}
                      y1={100 -
                        (diskIOHistory[i - 1].write / diskScale.max) * 100 +
                        "%"}
                      x2={xPos}
                      y2={100 - (point.write / diskScale.max) * 100 + "%"}
                      stroke="#F59E0B"
                      stroke-width="2"
                      stroke-linejoin="round"
                      stroke-linecap="round"
                    />
                  {/if}
                {/each}
              </svg>
            </div>
          </div>
          <div class="flex items-center space-x-4 text-xs">
            <div class="flex items-center space-x-1">
              <div class="w-3 h-3 bg-green-500 rounded-full"></div>
              <span class="text-gray-600 dark:text-gray-300">Read</span>
            </div>
            <div class="flex items-center space-x-1">
              <div class="w-3 h-3 bg-amber-500 rounded-full"></div>
              <span class="text-gray-600 dark:text-gray-300">Write</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Network I/O Graph -->
      <div class="card">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
          Network I/O
        </h3>
        <div class="space-y-4">
          <div class="flex justify-around">
            <!-- Download Gauge -->
            <div class="text-center">
              <div class="relative w-32 h-32 mx-auto">
                <svg class="w-full h-full transform -rotate-90">
                  <circle
                    cx="64"
                    cy="64"
                    r="56"
                    stroke="#E5E7EB"
                    stroke-width="8"
                    fill="none"
                    class="dark:stroke-gray-600"
                  />
                  <circle
                    cx="64"
                    cy="64"
                    r="56"
                    stroke="#8B5CF6"
                    stroke-width="8"
                    fill="none"
                    stroke-dasharray={`${((networkIOHistory[networkIOHistory.length - 1]?.rx || 0) / calculateScale(networkIOHistory, "net").max) * 351.86} 351.86`}
                    stroke-linecap="round"
                  />
                </svg>
                <div
                  class="absolute inset-0 flex flex-col items-center justify-center"
                >
                  <span class="text-lg font-bold text-gray-900 dark:text-white"
                    >{(
                      (networkIOHistory[networkIOHistory.length - 1]?.rx ?? 0) /
                      1000000
                    ).toFixed(1)} Mbps</span
                  >
                  <span class="text-xs text-gray-500 dark:text-gray-400"
                    >Download</span
                  >
                </div>
              </div>
            </div>

            <!-- Upload Gauge -->
            <div class="text-center">
              <div class="relative w-32 h-32 mx-auto">
                <svg class="w-full h-full transform -rotate-90">
                  <circle
                    cx="64"
                    cy="64"
                    r="56"
                    stroke="#E5E7EB"
                    stroke-width="8"
                    fill="none"
                    class="dark:stroke-gray-600"
                  />
                  <circle
                    cx="64"
                    cy="64"
                    r="56"
                    stroke="#EC4899"
                    stroke-width="8"
                    fill="none"
                    stroke-dasharray={`${((networkIOHistory[networkIOHistory.length - 1]?.tx || 0) / calculateScale(networkIOHistory, "net").max) * 351.86} 351.86`}
                    stroke-linecap="round"
                  />
                </svg>
                <div
                  class="absolute inset-0 flex flex-col items-center justify-center"
                >
                  <span class="text-lg font-bold text-gray-900 dark:text-white"
                    >{(
                      (networkIOHistory[networkIOHistory.length - 1]?.tx ?? 0) /
                      1000000
                    ).toFixed(1)} Mbps</span
                  >
                  <span class="text-xs text-gray-500 dark:text-gray-400"
                    >Upload</span
                  >
                </div>
              </div>
            </div>
          </div>
          <div
            class="h-32 bg-gray-50 dark:bg-muted rounded-lg p-2 relative flex"
          >
            <!-- Y-axis labels -->
            <div
              class="flex flex-col justify-between text-xs text-gray-500 mr-2 w-16"
            >
              {#each calculateScale(networkIOHistory, "net").steps as step}
                <span>{(step / 1000000).toFixed(1)} Mbps</span>
              {/each}
            </div>

            <!-- Graph container -->
            <div class="flex-1 relative">
              <svg class="absolute inset-0 w-full h-full">
                <!-- Grid lines -->
                <g stroke="#6B7280" stroke-width="0.5">
                  <!-- Horizontal grid lines -->
                  <line x1="0%" y1="25%" x2="100%" y2="25%" />
                  <line x1="0%" y1="50%" x2="100%" y2="50%" />
                  <line x1="0%" y1="75%" x2="100%" y2="75%" />
                  <!-- Vertical grid lines -->
                  <line x1="25%" y1="0%" x2="25%" y2="100%" />
                  <line x1="50%" y1="0%" x2="50%" y2="100%" />
                  <line x1="75%" y1="0%" x2="75%" y2="100%" />
                </g>

                {#each networkIOHistory as point, i}
                  {@const networkScale = calculateScale(networkIOHistory, "net")}
                  {@const xPos = (i / networkIOHistory.length) * 100 + "%"}
                  <circle
                    cx={xPos}
                    cy={100 - (point.rx / networkScale.max) * 100 + "%"}
                    r="2"
                    fill="#8B5CF6"
                  />
                  <circle
                    cx={xPos}
                    cy={100 - (point.tx / networkScale.max) * 100 + "%"}
                    r="2"
                    fill="#EC4899"
                  />
                  {#if i > 0}
                    {@const prevXPos =
                      ((i - 1) / networkIOHistory.length) * 100 + "%"}
                    <line
                      x1={prevXPos}
                      y1={100 -
                        (networkIOHistory[i - 1].rx / networkScale.max) * 100 +
                        "%"}
                      x2={xPos}
                      y2={100 - (point.rx / networkScale.max) * 100 + "%"}
                      stroke="#8B5CF6"
                      stroke-width="2"
                      stroke-linejoin="round"
                      stroke-linecap="round"
                    />
                    <line
                      x1={prevXPos}
                      y1={100 -
                        (networkIOHistory[i - 1].tx / networkScale.max) * 100 +
                        "%"}
                      x2={xPos}
                      y2={100 - (point.tx / networkScale.max) * 100 + "%"}
                      stroke="#EC4899"
                      stroke-width="2"
                      stroke-linejoin="round"
                      stroke-linecap="round"
                    />
                  {/if}
                {/each}
              </svg>
            </div>
          </div>
          <div class="flex items-center space-x-4 text-xs">
            <div class="flex items-center space-x-1">
              <div class="w-3 h-3 bg-purple-500 rounded-full"></div>
              <span class="text-gray-600 dark:text-gray-300">Download</span>
            </div>
            <div class="flex items-center space-x-1">
              <div class="w-3 h-3 bg-pink-500 rounded-full"></div>
              <span class="text-gray-600 dark:text-gray-300">Upload</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Network Interfaces -->
    <div class="card">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
        Network Interfaces
      </h3>
      <div class="space-y-4">
        {#each systemStats?.network?.interfaces || [] as iface}
          <div class="p-4 bg-gray-50 dark:bg-muted rounded-lg">
            <div class="flex items-center justify-between mb-3">
              <div class="flex items-center space-x-3">
                <h4 class="font-medium text-gray-900 dark:text-white">
                  {iface.name}
                </h4>
                <span
                  class="px-2 py-1 text-xs font-medium rounded-full {iface.status ===
                  'up'
                    ? 'bg-green-100 text-green-600'
                    : 'bg-red-100 text-red-600'}"
                >
                  {iface.status}
                </span>
                <span class="text-sm text-gray-600 dark:text-gray-300"
                  >{iface.speed}</span
                >
              </div>
            </div>

            <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
              <div>
                <p class="text-gray-600 dark:text-gray-300">IP Address</p>
                <p class="font-medium text-gray-900 dark:text-white">
                  {iface.ip}
                </p>
              </div>
              <div>
                <p class="text-gray-600 dark:text-gray-300">Netmask</p>
                <p class="font-medium text-gray-900 dark:text-white">
                  {iface.netmask}
                </p>
              </div>
              <div>
                <p class="text-gray-600 dark:text-gray-300">Gateway</p>
                <p class="font-medium text-gray-900 dark:text-white">
                  {iface.gateway}
                </p>
              </div>
              <div>
                <p class="text-gray-600 dark:text-gray-300">Traffic</p>
                <p class="font-medium text-gray-900 dark:text-white">
                  ↓ {formatBytes(iface.rx)} / ↑ {formatBytes(iface.tx)}
                </p>
              </div>
            </div>
          </div>
        {/each}
      </div>
    </div>

    <!-- Storage Health -->
    <div class="card">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
        Storage Health
      </h3>
      <div class="space-y-4">
        {#each systemStats?.storage?.disks || [] as disk, index}
          <!-- RAID Array Separator -->
          {#if disk && disk.type && disk.type === 'raid' && (index === 0 || !systemStats?.storage?.disks[index - 1] || systemStats?.storage?.disks[index - 1].type !== 'raid')}
            <div class="relative my-6">
              <div class="absolute inset-0 flex items-center">
                <div class="w-full border-t-2 border-indigo-300 dark:border-indigo-700"></div>
              </div>
              <div class="relative flex justify-center">
                <span class="px-4 py-2 bg-indigo-100 dark:bg-indigo-900/50 text-indigo-700 dark:text-indigo-300 text-sm font-semibold rounded-full border-2 border-indigo-300 dark:border-indigo-700 flex items-center space-x-2">
                  <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M5 4a2 2 0 012-2h6a2 2 0 012 2v14l-5-2.5L5 18V4z"/>
                  </svg>
                  <span>RAID Arrays</span>
                </span>
              </div>
            </div>
          {/if}

          <!-- Disk Card -->
          <div class="bg-white dark:bg-card shadow-lg hover:shadow-xl transition-shadow duration-200 rounded-lg p-6 border border-gray-100 dark:border-border {disk?.type === 'raid' ? 'ring-2 ring-indigo-200 dark:ring-indigo-800' : ''}">
            <!-- Header Section -->
            <div class="flex items-start justify-between mb-6">
              <div class="flex items-center space-x-4">
                <!-- Disk Icon with glow effect -->
                <div class="relative">
                  <div class="absolute inset-0 {disk?.type === 'raid'
                    ? 'bg-gradient-to-br from-indigo-400 to-purple-500'
                    : 'bg-gradient-to-br from-green-400 to-emerald-500'} rounded-xl blur opacity-25"></div>
                  <div class="relative w-14 h-14 {disk?.type === 'raid'
                    ? 'bg-gradient-to-br from-indigo-100 to-purple-100 dark:from-indigo-900/40 dark:to-purple-900/40'
                    : 'bg-gradient-to-br from-green-100 to-emerald-100 dark:from-green-900/40 dark:to-emerald-900/40'} rounded-xl flex items-center justify-center shadow-lg">
                    <svg class="w-7 h-7 {disk?.type === 'raid' ? 'text-indigo-600 dark:text-indigo-400' : 'text-green-600 dark:text-green-400'}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5h4M4 7h16"/>
                    </svg>
                  </div>
                </div>

                <div>
                  <h3 class="text-xl font-bold text-gray-900 dark:text-white">
                    {disk?.device || "Unknown"}
                  </h3>
                  <div class="flex items-center flex-wrap gap-2 mt-1">
                    <!-- Health Badge -->
                    {#if disk?.smart_status}
                      <span class="px-2 py-0.5 text-xs font-semibold rounded-full {disk?.smart_status === 'healthy'
                        ? 'bg-gradient-to-r from-emerald-500 to-green-500 text-white'
                        : 'bg-gradient-to-r from-yellow-500 to-orange-500 text-white'}">
                        {disk.smart_status.toUpperCase()}
                      </span>
                    {/if}
                    <!-- Temperature Badge -->
                    {#if disk?.temperature}
                      <span class="text-xs text-gray-500 dark:text-gray-400">
                        {disk.temperature}°C
                      </span>
                    {/if}
                    <!-- RAID Badge -->
                    {#if disk?.type === 'raid'}
                      <span class="px-2 py-0.5 text-xs font-semibold rounded-full bg-gradient-to-r from-indigo-500 to-purple-500 text-white">
                        RAID
                      </span>
                    {/if}
                  </div>
                </div>
              </div>
            </div>

            <!-- Disk Details -->
            <div class="space-y-4">
              <!-- Model -->
              <div>
                <p class="text-sm text-gray-600 dark:text-gray-300">
                  {disk?.model || "Unknown Model"}
                </p>
              </div>

              <!-- Capacity & Usage -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <p class="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1">
                    Capacity
                  </p>
                  <p class="text-lg font-semibold text-gray-900 dark:text-white">
                    {formatBytes(disk?.size || 0)}
                  </p>
                </div>
                <div>
                  <p class="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1">
                    Used Space
                  </p>
                  <p class="text-lg font-semibold text-gray-900 dark:text-white">
                    {formatBytes(disk?.used || 0)}
                  </p>
                </div>
              </div>

              <!-- Usage Progress Bar -->
              <div>
                <div class="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400 mb-1">
                  <span>Usage</span>
                  <span>{disk?.size && disk.size > 0 ? ((disk?.used || 0) / disk.size * 100).toFixed(1) : 0}%</span>
                </div>
                <div class="w-full bg-gray-200 dark:bg-muted rounded-full h-3">
                  <div class="{disk?.type === 'raid' ? 'bg-indigo-600' : 'bg-gradient-to-r from-emerald-500 to-green-500'} h-3 rounded-full transition-all duration-500"
                    style="width: {disk?.size && disk.size > 0 ? ((disk?.used || 0) / disk.size * 100) : 0}%"></div>
                </div>
              </div>

              <!-- Health Status -->
              {#if disk?.health !== undefined}
                <div>
                  <p class="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1">
                    Health Status
                  </p>
                  <div class="flex items-center space-x-2">
                    <div class="w-2 h-2 rounded-full {disk.health >= 90
                      ? 'bg-green-500'
                      : disk.health >= 70
                      ? 'bg-yellow-500'
                      : 'bg-red-500'}"></div>
                    <span class="text-sm font-medium text-gray-900 dark:text-white">
                      {disk.health}% Good
                    </span>
                  </div>
                </div>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>
