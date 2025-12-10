<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { nfsAPI } from "$lib/api.js";
  import { onMount } from "svelte";

  let exports = [];
  let loading = true;
  let showCreateModal = false;
  let error = null;

  // Form data for new export
  let newExport = {
    path: "",
    clientNetwork: "192.168.1.0/24",
    options: {
      rw: true,
      sync: true,
      no_subtree_check: false,
      no_root_squash: false,
      async: false,
      crossmnt: false,
    },
  };

  async function loadExports() {
    try {
      loading = true;
      error = null;
      exports = await nfsAPI.getExports();
    } catch (err) {
      error = err.message;
      console.error("Failed to load NFS exports:", err);
      // Use mock data as fallback
      exports = [
        {
          id: 1,
          path: "/data/media",
          clients: [
            {
              network: "192.168.1.0/24",
              options: "rw,sync,no_subtree_check",
              access: "read-write",
            },
            {
              network: "10.0.0.100",
              options: "ro,sync,no_subtree_check",
              access: "read-only",
            },
          ],
          filesystem: "/dev/sdb1",
          size: "2TB",
          used: "1.8TB",
          active_connections: 3,
          created: "2024-01-15",
          last_modified: "2024-12-07 09:30:00",
        },
      ];
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadExports();
    // Refresh exports every 30 seconds
    const interval = setInterval(loadExports, 30000);
    return () => clearInterval(interval);
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

  function getUsageColor(percentage) {
    if (percentage >= 90) return "bg-red-500";
    if (percentage >= 75) return "bg-yellow-500";
    return "bg-green-500";
  }

  async function createExport() {
    try {
      // Build options string
      const options = [];
      if (newExport.options.rw) options.push("rw");
      else options.push("ro");
      if (newExport.options.sync) options.push("sync");
      if (newExport.options.no_subtree_check) options.push("no_subtree_check");
      if (newExport.options.no_root_squash) options.push("no_root_squash");
      if (newExport.options.async) options.push("async");
      if (newExport.options.crossmnt) options.push("crossmnt");

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

  function resetForm() {
    newExport = {
      path: "",
      clientNetwork: "192.168.1.0/24",
      options: {
        rw: true,
        sync: true,
        no_subtree_check: false,
        no_root_squash: false,
        async: false,
        crossmnt: false,
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
</script>

<div class="p-6">
  <div class="mb-6">
    <h2 class="text-xl font-bold text-gray-900 dark:text-white mb-2">
      NFS Exports
    </h2>
    <p class="text-sm text-gray-600 dark:text-gray-300">
      Manage Network File System exports
    </p>
  </div>

  <div class="flex justify-between items-center mb-6">
    <button on:click={loadExports} class="btn btn-secondary" disabled={loading}>
      Refresh
    </button>
    <button on:click={() => (showCreateModal = true)} class="btn btn-primary">
      <svg
        class="h-4 w-4 mr-2"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M12 6v6m0 0v6m0-6h6m-6 0H6"
        />
      </svg>
      Create Export
    </button>
  </div>

  {#if loading}
    <div class="flex items-center justify-center py-12">
      <div
        class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"
      ></div>
    </div>
  {:else if exports.length === 0}
    <div class="text-center py-12">
      <svg
        class="h-12 w-12 text-gray-400 mx-auto mb-4"
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
      <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">
        No NFS exports found
      </h3>
      <p class="text-gray-600 dark:text-gray-300 mb-4">
        Create your first NFS export to get started.
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
              6.5TB
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Exports List -->
    <div class="space-y-4">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
        Export Details
      </h3>

      {#each exports as export_item}
        <div class="card hover:shadow-lg transition-shadow">
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="flex items-center space-x-3 mb-3">
                <h4 class="text-lg font-semibold text-gray-900 dark:text-white">
                  {export_item.path}
                </h4>
                <span
                  class="px-2 py-1 text-xs font-medium rounded-full bg-blue-100 text-blue-600"
                >
                  {export_item.active_connections} connections
                </span>
              </div>

              <!-- Usage Bar -->
              <div class="mb-4">
                <div class="flex items-center justify-between text-sm mb-2">
                  <span class="text-gray-600 dark:text-gray-300"
                    >Usage: {export_item.used} / {export_item.size}</span
                  >
                  <span class="font-medium text-gray-900 dark:text-white"
                    >{calculateUsage(export_item.used, export_item.size)}%</span
                  >
                </div>
                <div
                  class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3"
                >
                  <div
                    class="{getUsageColor(
                      calculateUsage(export_item.used, export_item.size),
                    )} h-3 rounded-full transition-all duration-500"
                    style="width: {calculateUsage(
                      export_item.used,
                      export_item.size,
                    )}%"
                  ></div>
                </div>
              </div>

              <!-- Client Access -->
              <div class="mb-4">
                <p
                  class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3"
                >
                  Client Access Rules:
                </p>
                <div class="space-y-2">
                  {#each export_item.clients as client}
                    <div
                      class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700 rounded-lg"
                    >
                      <div class="flex items-center space-x-3">
                        <span
                          class="px-2 py-1 text-xs font-medium rounded-full {getAccessColor(
                            client.access,
                          )}"
                        >
                          {client.access}
                        </span>
                        <span
                          class="text-sm font-medium text-gray-900 dark:text-white"
                          >{client.network}</span
                        >
                      </div>
                      <div class="flex flex-wrap gap-1">
                        {#each parseOptions(client.options) as option}
                          <span
                            class="px-2 py-1 bg-white dark:bg-gray-700 text-gray-600 dark:text-gray-300 text-xs rounded border border-gray-200 dark:border-gray-600"
                          >
                            {option}
                          </span>
                        {/each}
                      </div>
                    </div>
                  {/each}
                </div>
              </div>

              <!-- Details Grid -->
              <div class="grid grid-cols-2 md:grid-cols-3 gap-4 text-sm mb-4">
                <div>
                  <p class="text-gray-600 dark:text-gray-300">Filesystem</p>
                  <p class="font-medium text-gray-900 dark:text-white">
                    {export_item.filesystem}
                  </p>
                </div>
                <div>
                  <p class="text-gray-600 dark:text-gray-300">Total Clients</p>
                  <p class="font-medium text-gray-900 dark:text-white">
                    {export_item.clients.length}
                  </p>
                </div>
                <div>
                  <p class="text-gray-600 dark:text-gray-300">
                    Active Connections
                  </p>
                  <p class="font-medium text-gray-900 dark:text-white">
                    {export_item.active_connections}
                  </p>
                </div>
              </div>

              <!-- Timestamps -->
              <div class="flex items-center space-x-6 text-xs text-gray-500">
                <span>Created: {export_item.created}</span>
                <span>Modified: {export_item.last_modified}</span>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex space-x-2 ml-4">
              <button
                class="p-2 text-gray-400 hover:text-gray-600 dark:text-gray-300"
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
                class="p-2 text-gray-400 hover:text-red-600"
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
        class="relative bg-white dark:bg-gray-800 rounded-lg max-w-2xl w-full p-6"
      >
        <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">
          Create NFS Export
        </h3>

        <form on:submit|preventDefault={createExport} class="space-y-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label
                for="export-path"
                class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                >Export Path</label
              >
              <input
                id="export-path"
                type="text"
                bind:value={newExport.path}
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
                placeholder="/data/share"
              />
            </div>

            <div>
              <label
                for="client-network"
                class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                >Client Network</label
              >
              <input
                id="client-network"
                type="text"
                bind:value={newExport.clientNetwork}
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
                placeholder="192.168.1.0/24"
              />
            </div>
          </div>

          <div class="space-y-3">
            <fieldset class="space-y-3">
              <legend
                class="block text-sm font-medium text-gray-700 dark:text-gray-300"
                >Export Options:</legend
              >
              <div class="grid grid-cols-2 gap-3">
                <label class="flex items-center">
                  <input
                    type="checkbox"
                    class="mr-2 dark:bg-gray-700 dark:border-gray-600"
                    bind:checked={newExport.options.rw}
                  />
                  <span class="text-sm text-gray-700 dark:text-gray-300"
                    >rw (read-write)</span
                  >
                </label>
                <label class="flex items-center">
                  <input
                    type="checkbox"
                    class="mr-2 dark:bg-gray-700 dark:border-gray-600"
                    bind:checked={newExport.options.sync}
                  />
                  <span class="text-sm text-gray-700 dark:text-gray-300"
                    >sync</span
                  >
                </label>
                <label class="flex items-center">
                  <input
                    type="checkbox"
                    class="mr-2 dark:bg-gray-700 dark:border-gray-600"
                    bind:checked={newExport.options.no_subtree_check}
                  />
                  <span class="text-sm text-gray-700 dark:text-gray-300"
                    >no_subtree_check</span
                  >
                </label>
                <label class="flex items-center">
                  <input
                    type="checkbox"
                    class="mr-2 dark:bg-gray-700 dark:border-gray-600"
                    bind:checked={newExport.options.no_root_squash}
                  />
                  <span class="text-sm text-gray-700 dark:text-gray-300"
                    >no_root_squash</span
                  >
                </label>
                <label class="flex items-center">
                  <input
                    type="checkbox"
                    class="mr-2 dark:bg-gray-700 dark:border-gray-600"
                    bind:checked={newExport.options.async}
                  />
                  <span class="text-sm text-gray-700 dark:text-gray-300"
                    >async</span
                  >
                </label>
                <label class="flex items-center">
                  <input
                    type="checkbox"
                    class="mr-2 dark:bg-gray-700 dark:border-gray-600"
                    bind:checked={newExport.options.crossmnt}
                  />
                  <span class="text-sm text-gray-700 dark:text-gray-300"
                    >crossmnt</span
                  >
                </label>
              </div>
            </fieldset>
          </div>

          <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg">
            <p class="text-sm text-gray-600 dark:text-gray-300 mb-2">
              <strong>Preview export line:</strong>
            </p>
            <code
              class="text-sm bg-gray-800 dark:bg-gray-900 text-gray-100 px-3 py-2 rounded block"
            >
              {newExport.path}
              {newExport.clientNetwork}({#if newExport.options.rw}rw{:else}ro{/if}{#if newExport.options.sync},sync{/if}{#if newExport.options.no_subtree_check},no_subtree_check{/if}{#if newExport.options.no_root_squash},no_root_squash{/if}{#if newExport.options.async},async{/if}{#if newExport.options.crossmnt},crossmnt{/if})
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
