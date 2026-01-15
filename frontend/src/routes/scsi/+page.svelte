<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { scsiAPI } from "$lib/api.js";
  import { onMount } from "svelte";

  let targets = [];
  let loading = true;
  let error = null;
  let showCreateModal = false;
  let showEditModal = false;
  let selectedTarget = null;
  let initiatorIPsText = "";

  // Form data for new target
  let newTarget = {
    name: "",
    status: "active",
    backing_store: "",
    initiator_ips: "",
  };

  async function loadTargets() {
    try {
      loading = true;
      error = null;
      targets = await scsiAPI.getTargets();
    } catch (err) {
      error = err.message;
      console.error("Failed to load SCSI targets:", err);
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadTargets();
    // Refresh targets every 15 seconds
    const interval = setInterval(loadTargets, 15000);
    return () => clearInterval(interval);
  });

  function getStatusColor(status) {
    switch (status) {
      case "active":
        return "text-green-600 bg-green-100";
      case "inactive":
        return "text-gray-600 dark:text-gray-300 bg-gray-100";
      case "error":
        return "text-red-600 bg-red-100";
      default:
        return "text-gray-600 dark:text-gray-300 bg-gray-100";
    }
  }

  async function toggleTarget(target) {
    try {
      const newStatus = target.status === "active" ? "inactive" : "active";
      await scsiAPI.updateTarget(target.id, { ...target, status: newStatus });
      await loadTargets(); // Refresh data
    } catch (err) {
      console.error("Failed to toggle target:", err);
      alert("Failed to toggle target status");
    }
  }

  async function deleteTarget(targetId) {
    if (confirm("Are you sure you want to delete this SCSI target?")) {
      try {
        await scsiAPI.deleteTarget(targetId);
        await loadTargets(); // Refresh data
      } catch (err) {
        console.error("Failed to delete target:", err);
        alert("Failed to delete target");
      }
    }
  }

  function openEditModal(target) {
    selectedTarget = target;
    initiatorIPsText = target.initiator_ips.join(", ");
    showEditModal = true;
  }

  function generateIQN() {
    const now = new Date();
    const year = now.getFullYear();
    const month = String(now.getMonth() + 1).padStart(2, "0");
    const timestamp = now.getTime();

    // Generate reverse domain format (com.example.nas)
    const domain = "com.nas.server";

    // Generate IQN: iqn.year-month.domain:target.timestamp
    return `iqn.${year}-${month}.${domain}:target.${timestamp}`;
  }

  async function createTarget() {
    try {
      const targetData = {
        ...newTarget,
        initiator_ips: newTarget.initiator_ips
          .split(",")
          .map((ip) => ip.trim())
          .filter((ip) => ip),
      };

      await scsiAPI.createTarget(targetData);
      await loadTargets(); // Refresh data
      showCreateModal = false;
      resetCreateForm();
    } catch (err) {
      console.error("Failed to create target:", err);
      alert("Failed to create target");
    }
  }

  function resetCreateForm() {
    newTarget = {
      name: "",
      status: "active",
      backing_store: "",
      initiator_ips: "",
    };
  }

  async function updateTarget() {
    try {
      await scsiAPI.updateTarget(selectedTarget.name, selectedTarget);
      await loadTargets(); // Refresh data
      showEditModal = false;
      selectedTarget = null;
    } catch (err) {
      console.error("Failed to update target:", err);
      alert("Failed to update target");
    }
  }
</script>

<div class="p-6">
  <div class="mb-6">
    <h2 class="text-xl font-bold text-gray-900 dark:text-white mb-2">
      iSCSI Targets
    </h2>
    <p class="text-sm text-gray-600 dark:text-gray-300">
      Manage iSCSI targets and LUNs
    </p>
  </div>

  <div class="flex justify-between items-center mb-6">
    <button on:click={loadTargets} class="btn btn-secondary" disabled={loading}>
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
      Create Target
    </button>
  </div>

  {#if loading}
    <div class="flex items-center justify-center py-12">
      <div
        class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"
      ></div>
    </div>
  {:else if targets.length === 0}
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
          d="M13 10V3L4 14h7v7l9-11h-7z"
        />
      </svg>
      <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">
        No SCSI targets found
      </h3>
      <p class="text-gray-600 dark:text-gray-300">
        Create your first iSCSI target to get started.
      </p>
    </div>
  {:else}
    <!-- Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
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
                d="M13 10V3L4 14h7v7l9-11h-7z"
              />
            </svg>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-600 dark:text-gray-300">
              Total Targets
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {targets.length}
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
              Active
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {targets.filter((t) => t.status === "active").length}
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
              Sessions
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {targets.reduce((sum, t) => sum + t.sessions, 0)}
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

    <!-- Targets List -->
    <div class="space-y-4">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
        Target Details
      </h3>

      {#each targets as target}
        <div class="card hover:shadow-lg transition-shadow">
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="flex items-center space-x-3 mb-3">
                <h4 class="text-lg font-semibold text-gray-900 dark:text-white">
                  {target.name}
                </h4>
                <span
                  class="px-2 py-1 text-xs font-medium rounded-full {getStatusColor(
                    target.status,
                  )}"
                >
                  {target.status}
                </span>
              </div>

              <!-- Details Grid -->
              <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm mb-4">
                <div>
                  <p class="text-gray-600 dark:text-gray-300">Sessions</p>
                  <p class="font-medium text-gray-900 dark:text-white">
                    {target.sessions}
                  </p>
                </div>
                <div>
                  <p class="text-gray-600 dark:text-gray-300">LUNs</p>
                  <p class="font-medium text-gray-900 dark:text-white">
                    {target.lun_count}
                  </p>
                </div>
                <div>
                  <p class="text-gray-600 dark:text-gray-300">Size</p>
                  <p class="font-medium text-gray-900 dark:text-white">
                    {target.size}
                  </p>
                </div>
                <div>
                  <p class="text-gray-600 dark:text-gray-300">Backing Store</p>
                  <p class="font-medium text-gray-900 dark:text-white">
                    {target.backing_store}
                  </p>
                </div>
              </div>

              <!-- Initiators -->
              {#if target.initiator_ips.length > 0}
                <div class="mb-4">
                  <p
                    class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                  >
                    Connected Initiators:
                  </p>
                  <div class="flex flex-wrap gap-2">
                    {#each target.initiator_ips as ip}
                      <span
                        class="px-2 py-1 bg-gray-100 dark:bg-muted text-gray-700 dark:text-gray-300 text-xs rounded-md"
                      >
                        {ip}
                      </span>
                    {/each}
                  </div>
                </div>
              {/if}

              <!-- Timestamps -->
              <div
                class="flex items-center space-x-6 text-xs text-gray-500 dark:text-gray-400"
              >
                <span>Created: {target.created}</span>
                <span>Last access: {target.last_access}</span>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex space-x-2 ml-4">
              <button
                on:click={() => toggleTarget(target)}
                class="p-2 text-gray-400 hover:text-gray-600 dark:text-gray-300"
                title={target.status === "active" ? "Deactivate" : "Activate"}
              >
                <svg
                  class="h-5 w-5"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  {#if target.status === "active"}
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  {:else}
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
                    />
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  {/if}
                </svg>
              </button>
              <button
                class="p-2 text-gray-400 hover:text-gray-600 dark:text-gray-300"
                title="Edit"
                on:click={() => openEditModal(target)}
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
                on:click={() => deleteTarget(target.id)}
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

<!-- Create Target Modal -->
{#if showCreateModal}
  <div class="fixed inset-0 z-50 overflow-y-auto">
    <div class="flex items-center justify-center min-h-screen px-4">
      <div
        class="fixed inset-0 bg-gray-500 bg-opacity-75"
        on:click={() => (showCreateModal = false)}
        on:keydown={(e) => e.key === "Escape" && (showCreateModal = false)}
        role="button"
        tabindex="0"
        aria-label="Close modal"
      ></div>

      <div
        class="relative bg-white dark:bg-card rounded-lg max-w-md w-full p-6"
      >
        <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">
          Create iSCSI Target
        </h3>

        <form on:submit|preventDefault={createTarget} class="space-y-4">
          <div>
            <label
              for="target-name"
              class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
              >Target IQN</label
            >
            <div class="flex space-x-2">
              <input
                id="target-name"
                type="text"
                bind:value={newTarget.name}
                class="flex-1 px-3 py-2 border border-gray-300 dark:border bg-white dark:bg-card rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
                placeholder="iqn.2024-01.com.nas:target-name"
                required
              />
              <button
                type="button"
                on:click={() => (newTarget.name = generateIQN())}
                class="btn btn-secondary"
                title="Generate IQN"
              >
                <svg
                  class="h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                  />
                </svg>
              </button>
            </div>
          </div>

          <div>
            <label
              for="backing-store"
              class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
              >Backing Store</label
            >
            <select
              id="backing-store"
              bind:value={newTarget.backing_store}
              class="w-full px-3 py-2 border border-gray-300 dark:border bg-white dark:bg-card rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
            >
              <option value="">Select backing store</option>
              <option value="/dev/sdb1">/dev/sdb1</option>
              <option value="/dev/sdc1">/dev/sdc1</option>
              <option value="/dev/sda2">/dev/sda2</option>
            </select>
          </div>

          <div>
            <label
              for="initiator-ips"
              class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
              >Initiator IPs (comma separated)</label
            >
            <input
              id="initiator-ips"
              type="text"
              bind:value={newTarget.initiator_ips}
              class="w-full px-3 py-2 border border-gray-300 dark:border bg-white dark:bg-card rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
              placeholder="192.168.1.100, 192.168.1.101"
            />
          </div>

          <div class="flex justify-end space-x-3 pt-4">
            <button
              type="button"
              on:click={() => {
                showCreateModal = false;
                resetCreateForm();
              }}
              class="btn btn-secondary"
            >
              Cancel
            </button>
            <button type="submit" class="btn btn-primary">
              Create Target
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
{/if}

<!-- Edit Target Modal -->
{#if showEditModal && selectedTarget}
  <div class="fixed inset-0 z-50 overflow-y-auto">
    <div class="flex items-center justify-center min-h-screen px-4">
      <div
        class="fixed inset-0 bg-gray-500 bg-opacity-75"
        on:click={() => {
          showEditModal = false;
          selectedTarget = null;
        }}
        on:keydown={(e) => e.key === "Escape" && (showEditModal = false)}
        role="button"
        tabindex="0"
        aria-label="Close modal"
      ></div>

      <div
        class="relative bg-white dark:bg-card rounded-lg max-w-2xl w-full p-6"
      >
        <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">
          Edit iSCSI Target
        </h3>

        <form on:submit|preventDefault={updateTarget} class="space-y-4">
          <div>
            <label
              for="edit-target-name"
              class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
              >Target IQN</label
            >
            <div class="flex space-x-2">
              <input
                id="edit-target-name"
                type="text"
                bind:value={selectedTarget.name}
                class="flex-1 px-3 py-2 border border-gray-300 dark:border bg-white dark:bg-card rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
                placeholder="iqn.2024-01.com.nas:storage.target1"
                required
              />
              <button
                type="button"
                on:click={() => (selectedTarget.name = generateIQN())}
                class="btn btn-secondary"
                title="Generate IQN"
              >
                <svg
                  class="h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                  />
                </svg>
              </button>
            </div>
          </div>

          <div>
            <label
              for="edit-target-status"
              class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
              >Status</label
            >
            <select
              id="edit-target-status"
              bind:value={selectedTarget.status}
              class="w-full px-3 py-2 border border-gray-300 dark:border bg-white dark:bg-card rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
            >
              <option value="active">Active</option>
              <option value="inactive">Inactive</option>
            </select>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label
                for="edit-sessions"
                class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                >Sessions</label
              >
              <input
                id="edit-sessions"
                type="number"
                bind:value={selectedTarget.sessions}
                class="w-full px-3 py-2 border border-gray-300 dark:border bg-white dark:bg-card rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
                readonly
              />
            </div>

            <div>
              <label
                for="edit-luns"
                class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                >LUNs</label
              >
              <input
                id="edit-luns"
                type="number"
                bind:value={selectedTarget.lun_count}
                class="w-full px-3 py-2 border border-gray-300 dark:border bg-white dark:bg-card rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
                readonly
              />
            </div>
          </div>

          <div>
            <label
              for="edit-size"
              class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
              >Size</label
            >
            <input
              id="edit-size"
              type="text"
              bind:value={selectedTarget.size}
              class="w-full px-3 py-2 border border-gray-300 dark:border bg-white dark:bg-card rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
              readonly
            />
          </div>

          <div>
            <label
              for="edit-backing-store"
              class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
              >Backing Store</label
            >
            <input
              id="edit-backing-store"
              type="text"
              bind:value={selectedTarget.backing_store}
              class="w-full px-3 py-2 border border-gray-300 dark:border bg-white dark:bg-card rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
              placeholder="/dev/sdb1"
              readonly
            />
          </div>

          <div>
            <label
              for="edit-initiator-ips"
              class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
              >Allowed Initiator IPs</label
            >
            <input
              id="edit-initiator-ips"
              type="text"
              bind:value={initiatorIPsText}
              class="w-full px-3 py-2 border border-gray-300 dark:border bg-white dark:bg-card rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
              placeholder="192.168.1.100, 192.168.1.101"
              readonly
            />
          </div>

          <div class="flex justify-end space-x-3 pt-4">
            <button
              type="button"
              on:click={() => {
                showEditModal = false;
                selectedTarget = null;
              }}
              class="btn btn-secondary"
            >
              Cancel
            </button>
            <button type="submit" class="btn btn-primary">
              Update Target
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
{/if}
