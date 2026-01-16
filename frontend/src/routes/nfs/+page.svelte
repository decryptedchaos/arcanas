<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { nfsAPI, storageAPI } from "$lib/api.js";
  import { onMount } from "svelte";

  let exports = [];
  let loading = true;
  let showCreateModal = false;
  let showEditModal = false;
  let error = null;

  // Storage pools data
  let storagePools = [];
  let poolsLoading = true;

  // Form data for new export
  let newExport = {
    path: "",
    clientNetwork: "192.168.1.0/24",
    options: {
      rw: true,
      sync: false,
      no_subtree_check: false,
      no_root_squash: false,
      async: true,
      crossmnt: false,
      no_wdelay: false,
    },
  };

  // Editing state
  let editingExport = null;
  let editClients = [];
  let editNewClientNetwork = "192.168.1.0/24";
  let editNewClientOptions = {
    rw: true,
    sync: false,
    no_subtree_check: false,
    no_root_squash: false,
    async: true,
    crossmnt: false,
    no_wdelay: false,
  };

  async function loadExports() {
    try {
      loading = true;
      error = null;
      exports = await nfsAPI.getExports();
    } catch (err) {
      error = err.message;
      console.error("Failed to load NFS exports:", err);
    } finally {
      loading = false;
    }
  }

  async function loadStoragePools() {
    try {
      poolsLoading = true;
      const result = await storageAPI.getPools();

      // Handle null or undefined responses
      if (result === null || result === undefined) {
        storagePools = [];
      } else if (Array.isArray(result)) {
        storagePools = result;
      } else {
        storagePools = [];
      }
    } catch (err) {
      console.error("Failed to load storage pools:", err);
      storagePools = [];
    } finally {
      poolsLoading = false;
    }
  }

  onMount(() => {
    loadExports();
    loadStoragePools();
  });

  function getAccessColor(access) {
    switch (access) {
      case "read-write":
        return "text-green-600 bg-green-100";
      case "read-only":
        return "text-blue-600 bg-blue-100";
      default:
        return "text-gray-600 dark:text-gray-300 bg-gray-100";
    }
  }

  async function createExport() {
    try {
      // Build options string (sync and async are mutually exclusive)
      const options = [];
      if (newExport.options.rw) options.push("rw");
      else options.push("ro");
      if (newExport.options.sync && !newExport.options.async) options.push("sync");
      if (newExport.options.no_subtree_check) options.push("no_subtree_check");
      if (newExport.options.no_root_squash) options.push("no_root_squash");
      if (newExport.options.async) options.push("async");
      if (newExport.options.crossmnt) options.push("crossmnt");
      if (newExport.options.no_wdelay) options.push("no_wdelay");

      const exportData = {
        path: newExport.path,
        clients: [
          {
            network: newExport.clientNetwork,
            options: options.join(","),
            access: newExport.options.rw ? "read-write" : "read-only",
          },
        ],
      };

      await nfsAPI.createExport(exportData);
      showCreateModal = false;
      resetForm();
      await loadExports();
    } catch (err) {
      console.error("Failed to create export:", err);
      alert("Failed to create NFS export");
    }
  }

  async function deleteExport(exportPath) {
    if (!confirm("Are you sure you want to delete this NFS export?")) return;

    try {
      await nfsAPI.deleteExport(exportPath);
      await loadExports();
    } catch (err) {
      console.error("Failed to delete export:", err);
      alert("Failed to delete NFS export");
    }
  }

  function openEditModal(export_item) {
    editingExport = export_item;
    // Deep copy clients so we don't modify the original until save
    editClients = export_item.clients.map(client => ({
      network: client.network,
      options: client.options,
      access: client.access
    }));
    showEditModal = true;
  }

  function closeEditModal() {
    editingExport = null;
    editClients = [];
    showEditModal = false;
  }

  function removeClient(index) {
    editClients.splice(index, 1);
  }

  function addClient() {
    const options = [];
    if (editNewClientOptions.rw) options.push("rw");
    else options.push("ro");
    // sync and async are mutually exclusive
    if (editNewClientOptions.sync && !editNewClientOptions.async) options.push("sync");
    if (editNewClientOptions.no_subtree_check) options.push("no_subtree_check");
    if (editNewClientOptions.no_root_squash) options.push("no_root_squash");
    if (editNewClientOptions.async) options.push("async");
    if (editNewClientOptions.crossmnt) options.push("crossmnt");
    if (editNewClientOptions.no_wdelay) options.push("no_wdelay");

    editClients.push({
      network: editNewClientNetwork,
      options: options.join(","),
      access: editNewClientOptions.rw ? "read-write" : "read-only",
    });

    // Reset form
    editNewClientNetwork = "192.168.1.0/24";
    editNewClientOptions = {
      rw: true,
      sync: false,
      no_subtree_check: false,
      no_root_squash: false,
      async: true,
      crossmnt: false,
      no_wdelay: false,
    };
  }

  function parseOptionsToObject(optionsStr) {
    const opts = {
      rw: false,
      sync: false,
      no_subtree_check: false,
      no_root_squash: false,
      async: false,
      crossmnt: false,
      no_wdelay: false,
    };
    const parts = optionsStr.split(",");
    for (const part of parts) {
      const opt = part.trim();
      if (opt === "rw") opts.rw = true;
      if (opt === "ro") opts.rw = false;
      if (opt === "sync") opts.sync = true;
      if (opt === "async") opts.async = true;
      if (opt === "no_subtree_check") opts.no_subtree_check = true;
      if (opt === "no_root_squash") opts.no_root_squash = true;
      if (opt === "crossmnt") opts.crossmnt = true;
      if (opt === "no_wdelay") opts.no_wdelay = true;
    }
    return opts;
  }

  async function updateExport() {
    try {
      if (!editingExport) return;

      const exportData = {
        path: editingExport.path,
        clients: editClients,
      };

      await nfsAPI.updateExport(editingExport.path, exportData);
      closeEditModal();
      await loadExports();
    } catch (err) {
      console.error("Failed to update export:", err);
      alert("Failed to update NFS export");
    }
  }

  function resetForm() {
    newExport = {
      path: "",
      clientNetwork: "192.168.1.0/24",
      options: {
        rw: true,
        sync: false,
        no_subtree_check: false,
        no_root_squash: false,
        async: true,
        crossmnt: false,
        no_wdelay: false,
      },
    };
  }

  function parseOptions(options) {
    return options.split(",").map((opt) => opt.trim());
  }

  function calculateUsage(used, size) {
    const usedGB = parseFloat(used.replace(/[^\d.]/g, ""));
    const sizeGB = parseFloat(size.replace(/[^\d.]/g, ""));
    return Math.round((usedGB / sizeGB) * 100);
  }

  function parseSizeToGB(sizeStr) {
    // Parse strings like "6.5TB", "500GB", "2.3TB", "100GB", "6.5T", "500G"
    const match = sizeStr.match(/^([\d.]+)(\w*)$/);
    if (!match) return 0;

    const value = parseFloat(match[1]);
    let unit = match[2].toUpperCase();

    // Handle both "TB"/"GB" and "T"/"G" formats
    if (unit === 'T') unit = 'TB';
    if (unit === 'G') unit = 'GB';
    if (unit === 'M') unit = 'MB';

    switch (unit) {
      case 'TB': return value * 1024;
      case 'GB': return value;
      case 'MB': return value / 1024;
      default: return 0;
    }
  }

  function formatGBtoSize(gb) {
    if (gb >= 1024) {
      return (gb / 1024).toFixed(1) + 'TB';
    }
    return Math.round(gb) + 'GB';
  }

  function getTotalSize() {
    let totalGB = 0;
    for (const exp of exports) {
      totalGB += parseSizeToGB(exp.size);
    }
    return formatGBtoSize(totalGB);
  }

  function generateMountCommand(export_item) {
    // Get the server's IP from the current page
    const serverIP = window.location.hostname;
    return `mount -t nfs ${serverIP}:${export_item.path} /mnt/mountpoint -o noac,wsize=1048576,rsize=1048576`;
  }

  function generateFstabLine(export_item) {
    const serverIP = window.location.hostname;
    return `${serverIP}:${export_item.path} /mnt/mountpoint nfs _netdev,noac,wsize=1048576,rsize=1048576 0 0`;
  }

  async function copyToClipboard(text) {
    try {
      await navigator.clipboard.writeText(text);
    } catch (err) {
      // Fallback for older browsers
      const textArea = document.createElement('textarea');
      textArea.value = text;
      textArea.style.position = 'fixed';
      textArea.style.left = '-999999px';
      document.body.appendChild(textArea);
      textArea.select();
      try {
        document.execCommand('copy');
      } catch (err) {
        console.error('Failed to copy:', err);
      }
      document.body.removeChild(textArea);
    }
  }
</script>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center space-x-4">
      <!-- NFS Icon with glow effect -->
      <div class="relative">
        <div class="absolute inset-0 bg-gradient-to-br from-green-400 to-emerald-500 rounded-xl blur opacity-25"></div>
        <div class="relative w-14 h-14 bg-gradient-to-br from-green-100 to-emerald-100 dark:from-green-900/40 dark:to-emerald-900/40 rounded-xl flex items-center justify-center shadow-lg">
          <svg class="w-7 h-7 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"/>
          </svg>
        </div>
      </div>
      <div>
        <h2 class="text-xl font-bold text-gray-900 dark:text-white">
          NFS Exports
        </h2>
        <p class="text-sm text-gray-600 dark:text-gray-300 mt-1">
          Manage Network File System exports
        </p>
      </div>
    </div>
    <div class="flex items-center space-x-3">
      <button
        on:click={loadExports}
        class="btn btn-secondary"
        disabled={loading}
      >
        Refresh
      </button>
      <button
        on:click={() => (showCreateModal = true)}
        class="btn btn-primary"
      >
        Create Export
      </button>
    </div>
  </div>

  {#if loading}
    <div class="flex justify-center items-center h-64">
      <div
        class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"
      ></div>
    </div>
  {:else if error}
    <div
      class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4"
    >
      <div class="text-red-700 dark:text-red-400">Error: {error}</div>
    </div>
  {:else if exports.length === 0}
    <div class="text-center py-12">
      <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">
        No NFS exports found
      </h3>
      <p class="text-gray-600 dark:text-gray-300">
        No NFS exports are configured.
      </p>
    </div>
  {:else}
    <!-- Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-6">
      <div class="stat-card">
        <div class="flex items-center">
          <div class="p-3 bg-blue-100 rounded-lg mr-4">
            <svg
              class="h-6 w-6 text-blue-600"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"
              />
            </svg>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-600 dark:text-gray-300">
              Total Exports
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {exports.length}
            </p>
          </div>
        </div>
      </div>

      <div class="stat-card">
        <div class="flex items-center">
          <div class="p-3 bg-purple-100 rounded-lg mr-4">
            <svg
              class="h-6 w-6 text-purple-600"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
              />
            </svg>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-600 dark:text-gray-300">
              Active Connections
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {exports.reduce((sum, e) => sum + e.active_connections, 0)}
            </p>
          </div>
        </div>
      </div>

      <div class="stat-card">
        <div class="flex items-center">
          <div class="p-3 bg-green-100 rounded-lg mr-4">
            <svg
              class="h-6 w-6 text-green-600"
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
              Read-Write Access
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {exports.reduce(
                (sum, e) =>
                  sum +
                  e.clients.filter((c) => c.access === "read-write").length,
                0,
              )}
            </p>
          </div>
        </div>
      </div>

      <div class="stat-card">
        <div class="flex items-center">
          <div class="p-3 bg-orange-100 rounded-lg mr-4">
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
                d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"
              />
            </svg>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-600 dark:text-gray-300">
              Total Size
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {getTotalSize()}
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Exports List -->
    <div class="space-y-4">
      {#each exports as export_item}
        <div class="bg-white dark:bg-card shadow-lg hover:shadow-xl transition-shadow duration-200 rounded-lg p-6 border border-gray-100 dark:border-border">
          <!-- Header Section -->
          <div class="flex items-start justify-between mb-6">
            <div class="flex items-center space-x-4">
              <!-- NFS Icon with glow effect -->
              <div class="relative">
                <div class="absolute inset-0 bg-gradient-to-br from-green-400 to-emerald-500 rounded-xl blur opacity-25"></div>
                <div class="relative w-14 h-14 bg-gradient-to-br from-green-100 to-emerald-100 dark:from-green-900/40 dark:to-emerald-900/40 rounded-xl flex items-center justify-center shadow-lg">
                  <svg class="w-7 h-7 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"/>
                  </svg>
                </div>
              </div>

              <div>
                <h3 class="text-xl font-bold text-gray-900 dark:text-white">
                  {export_item.path}
                </h3>
                <div class="flex items-center flex-wrap gap-2 mt-1">
                  <!-- Active Connections Badge -->
                  <span class="px-2 py-0.5 text-xs font-semibold rounded-full bg-gradient-to-r from-green-500 to-emerald-500 text-white">
                    {export_item.active_connections} connections
                  </span>
                </div>
              </div>
            </div>

            <!-- Action Buttons -->
            <div class="flex space-x-2">
              <button
                on:click={() => openEditModal(export_item)}
                class="p-2 rounded-lg text-gray-400 hover:text-blue-500 hover:bg-blue-50 dark:hover:bg-blue-900/20"
                title="Edit"
              >
                <svg
                  class="h-5 w-5"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                  />
                </svg>
              </button>
              <button
                on:click={() => deleteExport(export_item.path)}
                class="p-2 rounded-lg text-gray-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20"
                title="Delete"
              >
                <svg
                  class="h-5 w-5"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                  />
                </svg>
              </button>
            </div>
          </div>

          <!-- Stats Grid -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
            <!-- Usage Card -->
            <div class="bg-gradient-to-br from-purple-50 to-pink-50 dark:from-purple-900/20 dark:to-pink-900/20 rounded-xl p-4 border border-purple-100 dark:border-purple-800">
              <div class="flex items-center space-x-3">
                <div class="w-10 h-10 bg-purple-100 dark:bg-purple-800 rounded-lg flex items-center justify-center">
                  <svg class="w-5 h-5 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
                  </svg>
                </div>
                <div>
                  <p class="text-xs text-purple-600 dark:text-purple-400 font-medium uppercase tracking-wide">
                    Usage
                  </p>
                  <p class="text-lg font-bold text-gray-900 dark:text-white">
                    {export_item.used} / {export_item.size}
                  </p>
                </div>
              </div>
              <!-- Usage Bar -->
              <div class="mt-3 w-full bg-gray-200 dark:bg-muted rounded-full h-2 overflow-hidden">
                <div
                  class="h-full bg-gradient-to-r from-purple-500 to-pink-500 rounded-full transition-all duration-300"
                  style="width: {calculateUsage(export_item.used, export_item.size)}%"
                ></div>
              </div>
            </div>

            <!-- Filesystem Card -->
            <div class="bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 rounded-xl p-4 border border-blue-100 dark:border-blue-800">
              <div class="flex items-center space-x-3">
                <div class="w-10 h-10 bg-blue-100 dark:bg-blue-800 rounded-lg flex items-center justify-center">
                  <svg class="w-5 h-5 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5h4M4 7h16"/>
                  </svg>
                </div>
                <div>
                  <p class="text-xs text-blue-600 dark:text-blue-400 font-medium uppercase tracking-wide">
                    Filesystem
                  </p>
                  <p class="text-sm font-medium text-gray-900 dark:text-white truncate">
                    {export_item.filesystem}
                  </p>
                </div>
              </div>
            </div>

            <!-- Clients Card -->
            <div class="bg-gradient-to-br from-emerald-50 to-teal-50 dark:from-emerald-900/20 dark:to-teal-900/20 rounded-xl p-4 border border-emerald-100 dark:border-emerald-800">
              <div class="flex items-center space-x-3">
                <div class="w-10 h-10 bg-emerald-100 dark:bg-emerald-800 rounded-lg flex items-center justify-center">
                  <svg class="w-5 h-5 text-emerald-600 dark:text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"/>
                  </svg>
                </div>
                <div>
                  <p class="text-xs text-emerald-600 dark:text-emerald-400 font-medium uppercase tracking-wide">
                    Total Clients
                  </p>
                  <p class="text-lg font-bold text-gray-900 dark:text-white">
                    {export_item.clients.length}
                  </p>
                </div>
              </div>
            </div>
          </div>

          <!-- Mount/Fstab Reference -->
          <div class="bg-gray-50 dark:bg-muted rounded-lg p-3 mb-4">
            <p class="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-3">
              Quick Reference
            </p>
            <div class="space-y-3">
              <!-- Mount Command -->
              <div>
                <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">Mount:</p>
                <div class="flex items-start space-x-2">
                  <code class="text-xs font-mono text-gray-800 dark:text-gray-200 bg-white dark:bg-gray-900 px-3 py-2 rounded border border-gray-200 dark:border flex-1 select-all">
                    {generateMountCommand(export_item)}
                  </code>
                  <button
                    on:click={() => copyToClipboard(generateMountCommand(export_item))}
                    class="p-1.5 rounded hover:bg-gray-200 dark:hover:bg-gray-700 flex-shrink-0"
                    title="Copy mount command"
                  >
                    <svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                    </svg>
                  </button>
                </div>
              </div>
              <!-- Fstab Line -->
              <div>
                <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">Fstab:</p>
                <div class="flex items-start space-x-2">
                  <code class="text-xs font-mono text-gray-800 dark:text-gray-200 bg-white dark:bg-gray-900 px-3 py-2 rounded border border-gray-200 dark:border flex-1 select-all">
                    {generateFstabLine(export_item)}
                  </code>
                  <button
                    on:click={() => copyToClipboard(generateFstabLine(export_item))}
                    class="p-1.5 rounded hover:bg-gray-200 dark:hover:bg-gray-700 flex-shrink-0"
                    title="Copy fstab line"
                  >
                    <svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Client Access Rules -->
          <div>
            <p class="text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide mb-3">
              Client Access Rules
            </p>
            <div class="space-y-2">
              {#each export_item.clients as client}
                <div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-muted rounded-lg">
                  <div class="flex items-center space-x-3">
                    <span class="px-2 py-1 text-xs font-medium rounded-full {getAccessColor(client.access)}">
                      {client.access}
                    </span>
                    <span class="text-sm font-medium text-gray-900 dark:text-white">
                      {client.network}
                    </span>
                  </div>
                  <div class="flex flex-wrap gap-1">
                    {#each parseOptions(client.options) as option}
                      <span class="px-2 py-1 bg-white dark:bg-card text-gray-600 dark:text-gray-300 text-xs rounded border border-gray-200 dark:border">
                        {option}
                      </span>
                    {/each}
                  </div>
                </div>
              {/each}
            </div>
          </div>

          <!-- Timestamps -->
          <div class="flex items-center space-x-6 text-xs text-gray-500 pt-4 mt-4 border-t border-gray-200 dark:border-border">
            <span>Created: {export_item.created}</span>
            <span>Modified: {export_item.last_modified}</span>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Create Export Modal -->
{#if showCreateModal}
  <div class="fixed inset-0 z-50 overflow-y-auto">
    <div class="flex items-center justify-center min-h-screen px-4">
      <div
        class="fixed inset-0 bg-gray-500 bg-opacity-75"
        on:click={() => {
          showCreateModal = false;
          resetForm();
        }}
        on:keydown={(e) => e.key === "Escape" && (showCreateModal = false)}
        role="button"
        tabindex="0"
        aria-label="Close modal"
      ></div>

      <div
        class="relative bg-white dark:bg-card rounded-lg max-w-2xl w-full p-6"
      >
        <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">
          Create NFS Export
        </h3>

        <form on:submit|preventDefault={createExport} class="space-y-4">
          <!-- Storage Pool Picker -->
          {#if poolsLoading}
            <div class="w-full p-2 border rounded bg-gray-100 dark:bg-muted text-gray-500 dark:text-gray-300">
              Loading storage pools...
            </div>
          {:else if storagePools.length === 0}
            <div class="w-full p-2 border rounded bg-yellow-50 dark:bg-yellow-900/20 text-yellow-800 dark:text-yellow-200">
              No storage pools available. Please create a storage pool first.
            </div>
          {:else}
            <div>
              <label for="export-pool" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Storage Pool
              </label>
              <select
                id="export-pool"
                bind:value={newExport.path}
                class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-muted dark:text-white"
                required
              >
                <option value="">Select a storage pool</option>
                {#each storagePools as pool}
                  <option value={pool.mount_point}>
                    {pool.name} ({pool.type}) - {pool.mount_point}
                  </option>
                {/each}
              </select>
            </div>
          {/if}

          <div>
            <label for="client-network" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Client Network
            </label>
            <input
              id="client-network"
              type="text"
              bind:value={newExport.clientNetwork}
              class="w-full px-3 py-2 border border-gray-300 dark:border bg-white dark:bg-card rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
              placeholder="192.168.1.0/24"
            />
          </div>

          <div class="space-y-3">
            <fieldset class="space-y-3">
              <legend class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                Export Options:
              </legend>
              <div class="grid grid-cols-2 gap-3">
                <label class="flex items-center">
                  <input type="checkbox" class="mr-2 dark:bg-muted dark:border" bind:checked={newExport.options.rw} />
                  <span class="text-sm text-gray-700 dark:text-gray-300">rw (read-write)</span>
                </label>
                <label class="flex items-center">
                  <input type="checkbox" class="mr-2 dark:bg-muted dark:border" bind:checked={newExport.options.sync} />
                  <span class="text-sm text-gray-700 dark:text-gray-300">sync</span>
                </label>
                <label class="flex items-center">
                  <input type="checkbox" class="mr-2 dark:bg-muted dark:border" bind:checked={newExport.options.no_subtree_check} />
                  <span class="text-sm text-gray-700 dark:text-gray-300">no_subtree_check</span>
                </label>
                <label class="flex items-center">
                  <input type="checkbox" class="mr-2 dark:bg-muted dark:border" bind:checked={newExport.options.no_root_squash} />
                  <span class="text-sm text-gray-700 dark:text-gray-300">no_root_squash</span>
                </label>
                <label class="flex items-center">
                  <input type="checkbox" class="mr-2 dark:bg-muted dark:border" bind:checked={newExport.options.async} />
                  <span class="text-sm text-gray-700 dark:text-gray-300">async</span>
                </label>
                <label class="flex items-center">
                  <input type="checkbox" class="mr-2 dark:bg-muted dark:border" bind:checked={newExport.options.crossmnt} />
                  <span class="text-sm text-gray-700 dark:text-gray-300">crossmnt</span>
                </label>
                <label class="flex items-center">
                  <input type="checkbox" class="mr-2 dark:bg-muted dark:border" bind:checked={newExport.options.no_wdelay} />
                  <span class="text-sm text-gray-700 dark:text-gray-300">no_wdelay</span>
                </label>
              </div>
            </fieldset>
          </div>

          {#if newExport.options.sync && !newExport.options.async}
            <div class="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 p-3 rounded-lg mb-4">
              <div class="flex items-start">
                <svg class="h-5 w-5 text-amber-500 mr-2 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
                <div>
                  <p class="text-sm font-medium text-amber-800 dark:text-amber-200">Sync mode enabled</p>
                  <p class="text-xs text-amber-700 dark:text-amber-300 mt-1">
                    Sync mode significantly reduces write performance (2-5 MB/s on typical setups) and may cause inaccurate speed reporting. Use async mode for better performance unless you require absolute data safety.
                  </p>
                </div>
              </div>
            </div>
          {/if}

          <div class="bg-gray-50 dark:bg-muted p-4 rounded-lg">
            <p class="text-sm text-gray-600 dark:text-gray-300 mb-2">
              <strong>Preview export line:</strong>
            </p>
            <code class="text-sm bg-gray-800 dark:bg-muted text-gray-100 px-3 py-2 rounded block">
              {newExport.path}
              {newExport.clientNetwork}({#if newExport.options.rw}rw{:else}ro{/if}{#if newExport.options.sync},sync{/if}{#if newExport.options.no_subtree_check},no_subtree_check{/if}{#if newExport.options.no_root_squash},no_root_squash{/if}{#if newExport.options.async},async{/if}{#if newExport.options.crossmnt},crossmnt{/if}{#if newExport.options.no_wdelay},no_wdelay{/if})
            </code>
          </div>

          <div class="flex justify-end space-x-3 pt-4">
            <button
              type="button"
              on:click={() => {
                showCreateModal = false;
                resetForm();
              }}
              class="btn btn-secondary"
            >
              Cancel
            </button>
            <button type="submit" class="btn btn-primary">
              Create Export
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
{/if}

<!-- Edit Export Modal -->
{#if showEditModal && editingExport}
  <div class="fixed inset-0 z-50 overflow-y-auto">
    <div class="flex items-center justify-center min-h-screen px-4">
      <div
        class="fixed inset-0 bg-gray-500 bg-opacity-75"
        on:click={closeEditModal}
        on:keydown={(e) => e.key === "Escape" && closeEditModal()}
        role="button"
        tabindex="0"
        aria-label="Close modal"
      ></div>

      <div
        class="relative bg-white dark:bg-card rounded-lg max-w-3xl w-full p-6 max-h-[90vh] overflow-y-auto"
      >
        <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">
          Edit NFS Export: {editingExport.path}
        </h3>

        <div class="space-y-4">
          <!-- Export Path (read-only for now) -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Export Path
            </label>
            <input
              type="text"
              value={editingExport.path}
              disabled
              class="w-full px-3 py-2 border border-gray-300 dark:border bg-gray-100 dark:bg-muted rounded-md text-gray-500 dark:text-gray-400"
            />
            <p class="text-xs text-gray-500 mt-1">Path cannot be changed. Delete and recreate to change path.</p>
          </div>

          <!-- Existing Clients -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Client Access Rules
            </label>
            {#if editClients.length === 0}
              <div class="p-4 bg-yellow-50 dark:bg-yellow-900/20 text-yellow-800 dark:text-yellow-200 rounded-lg mb-3">
                No clients configured. Please add at least one client.
              </div>
            {:else}
              <div class="space-y-4 mb-4">
                {#each editClients as client, index}
                  <div class="p-4 bg-white dark:bg-card border border-gray-200 dark:border rounded-lg">
                    <div class="flex items-center justify-between mb-3">
                      <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                        Client {index + 1}
                      </span>
                      <button
                        on:click={() => removeClient(index)}
                        class="p-1 text-red-500 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-900/20 rounded"
                        type="button"
                        title="Remove client"
                      >
                        <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      </button>
                    </div>

                    <!-- Network Input -->
                    <div class="mb-3">
                      <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
                        Network
                      </label>
                      <input
                        type="text"
                        bind:value={client.network}
                        class="w-full px-3 py-2 border border-gray-300 dark:border bg-white dark:bg-card rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white text-sm"
                        placeholder="192.168.1.0/24"
                      />
                    </div>

                    <!-- Options -->
                    <div>
                      <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-2">
                        Options
                      </label>
                      <div class="grid grid-cols-3 gap-2">
                        <label class="flex items-center">
                          <input
                            type="checkbox"
                            class="mr-2 dark:bg-muted dark:border"
                            checked={client.options.includes('rw') || !client.options.includes('ro')}
                            on:change={(e) => {
                              if (e.target.checked) {
                                client.options = client.options.replace(',ro', '').replace('ro,', '').replace('ro', '');
                                if (!client.options.includes('rw')) client.options = 'rw,' + client.options;
                              } else {
                                client.options = client.options.replace(',rw', '').replace('rw,', '').replace('rw', '');
                                if (!client.options.includes('ro')) client.options = 'ro,' + client.options;
                              }
                            }}
                          />
                          <span class="text-xs text-gray-700 dark:text-gray-300">rw/ro</span>
                        </label>
                        <label class="flex items-center">
                          <input
                            type="checkbox"
                            class="mr-2 dark:bg-muted dark:border"
                            checked={client.options.includes('sync')}
                            on:change={(e) => {
                              if (e.target.checked && !client.options.includes('sync')) {
                                client.options += ',sync';
                              } else if (!e.target.checked) {
                                client.options = client.options.replace(',sync', '').replace('sync', '');
                              }
                            }}
                          />
                          <span class="text-xs text-gray-700 dark:text-gray-300">sync</span>
                        </label>
                        <label class="flex items-center">
                          <input
                            type="checkbox"
                            class="mr-2 dark:bg-muted dark:border"
                            checked={client.options.includes('no_subtree_check')}
                            on:change={(e) => {
                              if (e.target.checked && !client.options.includes('no_subtree_check')) {
                                client.options += ',no_subtree_check';
                              } else if (!e.target.checked) {
                                client.options = client.options.replace(',no_subtree_check', '').replace('no_subtree_check', '');
                              }
                            }}
                          />
                          <span class="text-xs text-gray-700 dark:text-gray-300">no_subtree_check</span>
                        </label>
                        <label class="flex items-center">
                          <input
                            type="checkbox"
                            class="mr-2 dark:bg-muted dark:border"
                            checked={client.options.includes('no_root_squash')}
                            on:change={(e) => {
                              if (e.target.checked && !client.options.includes('no_root_squash')) {
                                client.options += ',no_root_squash';
                              } else if (!e.target.checked) {
                                client.options = client.options.replace(',no_root_squash', '').replace('no_root_squash', '');
                              }
                            }}
                          />
                          <span class="text-xs text-gray-700 dark:text-gray-300">no_root_squash</span>
                        </label>
                        <label class="flex items-center">
                          <input
                            type="checkbox"
                            class="mr-2 dark:bg-muted dark:border"
                            checked={client.options.includes('async')}
                            on:change={(e) => {
                              if (e.target.checked && !client.options.includes('async')) {
                                client.options += ',async';
                              } else if (!e.target.checked) {
                                client.options = client.options.replace(',async', '').replace('async', '');
                              }
                            }}
                          />
                          <span class="text-xs text-gray-700 dark:text-gray-300">async</span>
                        </label>
                        <label class="flex items-center">
                          <input
                            type="checkbox"
                            class="mr-2 dark:bg-muted dark:border"
                            checked={client.options.includes('crossmnt')}
                            on:change={(e) => {
                              if (e.target.checked && !client.options.includes('crossmnt')) {
                                client.options += ',crossmnt';
                              } else if (!e.target.checked) {
                                client.options = client.options.replace(',crossmnt', '').replace('crossmnt', '');
                              }
                            }}
                          />
                          <span class="text-xs text-gray-700 dark:text-gray-300">crossmnt</span>
                        </label>
                        <label class="flex items-center">
                          <input
                            type="checkbox"
                            class="mr-2 dark:bg-muted dark:border"
                            checked={client.options.includes('no_wdelay')}
                            on:change={(e) => {
                              if (e.target.checked && !client.options.includes('no_wdelay')) {
                                client.options += ',no_wdelay';
                              } else if (!e.target.checked) {
                                client.options = client.options.replace(',no_wdelay', '').replace('no_wdelay', '');
                              }
                            }}
                          />
                          <span class="text-xs text-gray-700 dark:text-gray-300">no_wdelay</span>
                        </label>
                      </div>
                      <!-- Preview current options -->
                      <div class="mt-2 p-2 bg-gray-50 dark:bg-muted rounded text-xs font-mono text-gray-600 dark:text-gray-300">
                        {client.options}
                      </div>
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
          </div>

          <!-- Add New Client -->
          <div class="border-t border-gray-200 dark:border-border pt-4">
            <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">Add New Client</h4>

            <div class="space-y-3">
              <div>
                <label for="edit-client-network" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Client Network
                </label>
                <input
                  id="edit-client-network"
                  type="text"
                  bind:value={editNewClientNetwork}
                  class="w-full px-3 py-2 border border-gray-300 dark:border bg-white dark:bg-card rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
                  placeholder="192.168.1.0/24"
                />
              </div>

              <div class="space-y-3">
                <fieldset class="space-y-3">
                  <legend class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                    Export Options:
                  </legend>
                  <div class="grid grid-cols-2 gap-3">
                    <label class="flex items-center">
                      <input type="checkbox" class="mr-2 dark:bg-muted dark:border" bind:checked={editNewClientOptions.rw} />
                      <span class="text-sm text-gray-700 dark:text-gray-300">rw (read-write)</span>
                    </label>
                    <label class="flex items-center">
                      <input type="checkbox" class="mr-2 dark:bg-muted dark:border" bind:checked={editNewClientOptions.sync} />
                      <span class="text-sm text-gray-700 dark:text-gray-300">sync</span>
                    </label>
                    <label class="flex items-center">
                      <input type="checkbox" class="mr-2 dark:bg-muted dark:border" bind:checked={editNewClientOptions.no_subtree_check} />
                      <span class="text-sm text-gray-700 dark:text-gray-300">no_subtree_check</span>
                    </label>
                    <label class="flex items-center">
                      <input type="checkbox" class="mr-2 dark:bg-muted dark:border" bind:checked={editNewClientOptions.no_root_squash} />
                      <span class="text-sm text-gray-700 dark:text-gray-300">no_root_squash</span>
                    </label>
                    <label class="flex items-center">
                      <input type="checkbox" class="mr-2 dark:bg-muted dark:border" bind:checked={editNewClientOptions.async} />
                      <span class="text-sm text-gray-700 dark:text-gray-300">async</span>
                    </label>
                    <label class="flex items-center">
                      <input type="checkbox" class="mr-2 dark:bg-muted dark:border" bind:checked={editNewClientOptions.crossmnt} />
                      <span class="text-sm text-gray-700 dark:text-gray-300">crossmnt</span>
                    </label>
                    <label class="flex items-center">
                      <input type="checkbox" class="mr-2 dark:bg-muted dark:border" bind:checked={editNewClientOptions.no_wdelay} />
                      <span class="text-sm text-gray-700 dark:text-gray-300">no_wdelay</span>
                    </label>
                  </div>
                </fieldset>
              </div>

              <button
                on:click={addClient}
                disabled={!editNewClientNetwork}
                class="w-full py-2 px-4 bg-blue-600 hover:bg-blue-700 text-white rounded-md disabled:opacity-50 disabled:cursor-not-allowed"
                type="button"
              >
                + Add Client
              </button>
            </div>
          </div>

          <div class="flex justify-end space-x-3 pt-4 border-t border-gray-200 dark:border-border">
            <button
              type="button"
              on:click={closeEditModal}
              class="btn btn-secondary"
            >
              Cancel
            </button>
            <button
              on:click={updateExport}
              disabled={editClients.length === 0}
              class="btn btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Save Changes
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}
