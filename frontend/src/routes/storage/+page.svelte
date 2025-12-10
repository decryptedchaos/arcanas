<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";
    import { diskAPI, storageAPI } from "$lib/api.js";
    import { formatBytes } from "$lib/utils/byteUtils.js";
    import { onMount } from "svelte";

    // Tab persistence with URL params
    $: activeTab = $page.url.searchParams.get("tab") || "disks";

    let diskStats = [];
    let loading = true;
    let error = null;
    let selectedDisk = null;
    let expandedDisks = new Set();

    // RAID data
    let raidArrays = [];
    let raidLoading = true;
    let raidError = null;

    // Storage pools data
    let storagePools = [];
    let poolsLoading = true;
    let poolsError = null;

    // Modal state for pool actions
    let showDeleteModal = false;
    let poolToDelete = null;
    let deleteConfirmation = "";
    let showDetailsModal = false;
    let selectedPoolDetails = null;
    let openDropdown = null; // Track which dropdown is open

    async function loadDiskStats() {
        try {
            error = null;
            const newStats = await diskAPI.getDiskStats();
            if (JSON.stringify(newStats) !== JSON.stringify(diskStats)) {
                diskStats = newStats;
            }
        } catch (err) {
            error = err.message;
            console.error("Failed to load disk stats:", err);
        } finally {
            loading = false;
        }
    }

    async function loadRAIDArrays() {
        try {
            raidError = null;
            raidArrays = await diskAPI.getRAIDArrays();
        } catch (err) {
            raidError = err.message;
            console.error("Failed to load RAID arrays:", err);
        } finally {
            raidLoading = false;
        }
    }

    async function loadStoragePools() {
        try {
            poolsError = null;
            console.log("Loading storage pools...");
            const result = await storageAPI.getPools();
            console.log("Raw API result:", result);

            // Handle null or undefined responses
            if (result === null || result === undefined) {
                console.warn("Storage pools API returned null/undefined");
                storagePools = [];
            } else if (Array.isArray(result)) {
                storagePools = result;
            } else {
                console.warn("Storage pools API returned non-array:", result);
                storagePools = [];
            }

            console.log("Storage pools loaded:", storagePools);
        } catch (err) {
            poolsError = err.message;
            console.error("Failed to load storage pools:", err);
            storagePools = [];
        } finally {
            poolsLoading = false;
        }
    }

    function handlePoolAction(action, pool) {
        switch (action) {
            case "mount":
                // Toggle mount/unmount
                console.log("Toggle mount for pool:", pool.name);
                // TODO: Implement mount/unmount logic
                break;
            case "details":
                selectedPoolDetails = pool;
                showDetailsModal = true;
                break;
            case "delete":
                poolToDelete = pool;
                deleteConfirmation = `DELETE ${pool.name.toUpperCase()}`;
                showDeleteModal = true;
                break;
        }
    }

    async function confirmDeletePool() {
        if (!poolToDelete) return;

        try {
            console.log("Deleting pool:", poolToDelete.name);

            // Call the real delete API
            await storageAPI.deletePool(poolToDelete.name);

            // Refresh pools list
            await loadStoragePools();

            // Close modal
            showDeleteModal = false;
            poolToDelete = null;
            deleteConfirmation = "";
        } catch (err) {
            console.error("Failed to delete pool:", err);
            alert("Failed to delete pool: " + err.message);
        }
    }

    function cancelDelete() {
        showDeleteModal = false;
        poolToDelete = null;
        deleteConfirmation = "";
    }

    function closeDetailsModal() {
        showDetailsModal = false;
        selectedPoolDetails = null;
    }

    function toggleDropdown(poolName) {
        openDropdown = openDropdown === poolName ? null : poolName;
    }

    function closeDropdowns() {
        openDropdown = null;
    }

    function toggleDiskExpansion(diskName) {
        if (expandedDisks.has(diskName)) {
            expandedDisks.delete(diskName);
        } else {
            expandedDisks.add(diskName);
        }
        expandedDisks = expandedDisks;
    }

    function getHealthColor(health) {
        switch (health?.toLowerCase()) {
            case "good":
            case "healthy":
                return "text-green-600 bg-green-100";
            case "warning":
            case "caution":
                return "text-yellow-600 bg-yellow-100";
            case "bad":
            case "failed":
                return "text-red-600 bg-red-100";
            default:
                return "text-gray-600 bg-gray-100";
        }
    }

    onMount(() => {
        // Initial load only
        loadDiskStats();
        loadRAIDArrays();
        loadStoragePools();
    });

    $: disks = Array.isArray(diskStats) ? diskStats : [];
    $: raidArraysSafe = Array.isArray(raidArrays) ? raidArrays : [];
    $: storagePoolsSafe = Array.isArray(storagePools) ? storagePools : [];
</script>

<div class="p-6" role="main" tabindex="-1">
    <div class="mb-6">
        <h1 class="text-xl font-bold text-gray-900 dark:text-white mb-2">
            Storage Management
        </h1>
        <p class="text-sm text-gray-600 dark:text-gray-300">
            Manage disks, RAID arrays, and storage pools
        </p>
    </div>

    <!-- Tab Navigation -->
    <div class="border-b border-gray-200 dark:border-gray-700 mb-6">
        <nav class="-mb-px flex space-x-8">
            <button
                class="py-2 px-1 border-b-2 font-medium text-sm {activeTab ===
                'disks'
                    ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                    : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600'}"
                on:click={() => goto("?tab=disks", { replaceState: true })}
            >
                Disks
            </button>
            <button
                class="py-2 px-1 border-b-2 font-medium text-sm {activeTab ===
                'raid'
                    ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                    : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600'}"
                on:click={() => goto("?tab=raid", { replaceState: true })}
            >
                RAID Arrays
            </button>
            <button
                class="py-2 px-1 border-b-2 font-medium text-sm {activeTab ===
                'pools'
                    ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                    : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600'}"
                on:click={() => goto("?tab=pools", { replaceState: true })}
            >
                Storage Pools
            </button>
        </nav>
    </div>

    <!-- Action Buttons -->
    <div class="flex justify-between items-center mb-6">
        <button
            on:click={() => {
                loadDiskStats();
                loadRAIDArrays();
                loadStoragePools();
            }}
            class="btn btn-secondary"
            disabled={loading || raidLoading || poolsLoading}
        >
            Refresh All
        </button>
        {#if activeTab === "disks"}
            <button class="btn btn-primary">Scan for New Disks</button>
        {:else if activeTab === "raid"}
            <button class="btn btn-primary">Create RAID Array</button>
        {:else if activeTab === "pools"}
            <button class="btn btn-primary">Create Storage Pool</button>
        {/if}
    </div>

    <!-- Tab Content -->
    <div
        on:click={closeDropdowns}
        on:keydown={(e) => {
            if (e.key === "Escape") closeDropdowns();
        }}
    >
        <!-- Disks Tab -->
        {#if activeTab === "disks"}
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
                    <div class="text-red-700 dark:text-red-400">
                        Error: {error}
                    </div>
                </div>
            {:else if disks.length === 0}
                <div class="text-center py-12">
                    <h3
                        class="text-lg font-medium text-gray-900 dark:text-white mb-2"
                    >
                        No disks found
                    </h3>
                    <p class="text-gray-600 dark:text-gray-300">
                        No storage disks are available.
                    </p>
                </div>
            {:else}
                <div class="space-y-4">
                    {#each disks as disk}
                        <div
                            class="bg-white dark:bg-gray-800 shadow rounded-lg"
                        >
                            <div class="p-6">
                                <div class="flex justify-between items-start">
                                    <div>
                                        <div
                                            class="flex items-center space-x-3 mb-2"
                                        >
                                            <h3
                                                class="text-lg font-semibold text-gray-900 dark:text-white"
                                            >
                                                {disk.name}
                                            </h3>
                                            <span
                                                class="px-2 py-1 text-xs font-medium rounded-full {getHealthColor(
                                                    disk.health,
                                                )}"
                                            >
                                                {disk.health || "Unknown"}
                                            </span>
                                        </div>
                                        <p
                                            class="text-gray-600 dark:text-gray-300 mb-2"
                                        >
                                            {disk.model || "Unknown model"}
                                        </p>
                                        <div
                                            class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm"
                                        >
                                            <div>
                                                <p
                                                    class="text-gray-500 dark:text-gray-400"
                                                >
                                                    Size
                                                </p>
                                                <p
                                                    class="font-medium text-gray-900 dark:text-white"
                                                >
                                                    {disk.size}
                                                </p>
                                            </div>
                                            <div>
                                                <p
                                                    class="text-gray-500 dark:text-gray-400"
                                                >
                                                    Interface
                                                </p>
                                                <p
                                                    class="font-medium text-gray-900 dark:text-white"
                                                >
                                                    {disk.interface ||
                                                        "Unknown"}
                                                </p>
                                            </div>
                                            <div>
                                                <p
                                                    class="text-gray-500 dark:text-gray-400"
                                                >
                                                    Temperature
                                                </p>
                                                <p
                                                    class="font-medium text-gray-900 dark:text-white"
                                                >
                                                    {disk.temperature || "N/A"}
                                                </p>
                                            </div>
                                            <div>
                                                <p
                                                    class="text-gray-500 dark:text-gray-400"
                                                >
                                                    Serial
                                                </p>
                                                <p
                                                    class="font-medium text-gray-900 dark:text-white"
                                                >
                                                    {disk.serial_number ||
                                                        "N/A"}
                                                </p>
                                            </div>
                                        </div>
                                    </div>
                                    <button
                                        on:click={() =>
                                            toggleDiskExpansion(disk.name)}
                                        class="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                                        title="Toggle disk details"
                                        aria-label="Toggle disk details"
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
                                                d="M19 9l-7 7-7-7"
                                            />
                                        </svg>
                                    </button>
                                </div>
                            </div>
                            {#if expandedDisks.has(disk.name)}
                                <div
                                    class="border-t border-gray-200 p-6 bg-gray-50"
                                >
                                    <h4 class="font-medium text-gray-900 mb-4">
                                        Detailed Information
                                    </h4>
                                    <div
                                        class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm"
                                    >
                                        <div>
                                            <p
                                                class="text-gray-500 dark:text-gray-400"
                                            >
                                                Mount Point
                                            </p>
                                            <p class="font-medium">
                                                {disk.mount_point ||
                                                    "Not mounted"}
                                            </p>
                                        </div>
                                        <div>
                                            <p
                                                class="text-gray-500 dark:text-gray-400"
                                            >
                                                Filesystem
                                            </p>
                                            <p class="font-medium">
                                                {disk.filesystem || "Unknown"}
                                            </p>
                                        </div>
                                        <div>
                                            <p
                                                class="text-gray-500 dark:text-gray-400"
                                            >
                                                Used Space
                                            </p>
                                            <p class="font-medium">
                                                {disk.used || "N/A"}
                                            </p>
                                        </div>
                                        <div>
                                            <p
                                                class="text-gray-500 dark:text-gray-400"
                                            >
                                                Available Space
                                            </p>
                                            <p class="font-medium">
                                                {disk.available || "N/A"}
                                            </p>
                                        </div>
                                    </div>
                                </div>
                            {/if}
                        </div>
                    {/each}
                </div>
            {/if}

            <!-- RAID Arrays Tab -->
        {:else if activeTab === "raid"}
            {#if raidLoading}
                <div class="flex justify-center items-center h-64">
                    <div
                        class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"
                    ></div>
                </div>
            {:else if raidError}
                <div class="bg-red-50 border border-red-200 rounded-md p-4">
                    <div class="text-red-700">Error: {raidError}</div>
                </div>
            {:else if raidArraysSafe.length === 0}
                <div class="text-center py-12">
                    <h3 class="text-lg font-medium text-gray-900 mb-2">
                        No RAID arrays found
                    </h3>
                    <p class="text-gray-600 dark:text-gray-300 mb-4">
                        Create a RAID array to get started.
                    </p>
                </div>
            {:else}
                <div class="space-y-4">
                    {#each raidArraysSafe as array}
                        <div
                            class="bg-white dark:bg-gray-800 shadow-lg hover:shadow-xl transition-shadow duration-200 rounded-lg p-6 border border-gray-100 dark:border-gray-700"
                        >
                            <div class="flex justify-between items-start">
                                <div class="flex items-start space-x-4">
                                    <!-- RAID Type Icon -->
                                    <div class="flex-shrink-0">
                                        <div
                                            class="w-12 h-12 bg-orange-100 dark:bg-orange-900 rounded-lg flex items-center justify-center"
                                        >
                                            <svg
                                                class="w-6 h-6 text-orange-600 dark:text-orange-400"
                                                fill="currentColor"
                                                viewBox="0 0 20 20"
                                            >
                                                <path
                                                    d="M5 3a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2V5a2 2 0 00-2-2H5zM5 11a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2H5zM11 5a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V5zM13 11a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2h-2z"
                                                />
                                            </svg>
                                        </div>
                                    </div>

                                    <!-- RAID Info -->
                                    <div>
                                        <h3
                                            class="text-lg font-semibold text-gray-900 dark:text-white mb-2"
                                        >
                                            {array.name}
                                        </h3>
                                        <p
                                            class="text-gray-600 dark:text-gray-300 mb-4"
                                        >
                                            RAID Level: {array.level}
                                        </p>
                                        <div
                                            class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm"
                                        >
                                            <div>
                                                <p
                                                    class="text-gray-500 dark:text-gray-400"
                                                >
                                                    Size
                                                </p>
                                                <p
                                                    class="font-medium text-gray-900 dark:text-white"
                                                >
                                                    {formatBytes(array.size)}
                                                </p>
                                            </div>
                                            <div>
                                                <p
                                                    class="text-gray-500 dark:text-gray-400"
                                                >
                                                    Status
                                                </p>
                                                <div
                                                    class="flex items-center space-x-1"
                                                >
                                                    <div
                                                        class="w-2 h-2 rounded-full {array.status ===
                                                        'active'
                                                            ? 'bg-green-500'
                                                            : 'bg-yellow-500'}"
                                                    ></div>
                                                    <p
                                                        class="font-medium text-gray-900 dark:text-white"
                                                    >
                                                        {array.status}
                                                    </p>
                                                </div>
                                            </div>
                                            <div>
                                                <p
                                                    class="text-gray-500 dark:text-gray-400"
                                                >
                                                    Devices
                                                </p>
                                                <p
                                                    class="font-medium text-gray-900 dark:text-white"
                                                >
                                                    {array.devices?.length || 0}
                                                </p>
                                            </div>
                                            <div>
                                                <p
                                                    class="text-gray-500 dark:text-gray-400"
                                                >
                                                    Used
                                                </p>
                                                <p
                                                    class="font-medium text-gray-900 dark:text-white"
                                                >
                                                    {array.used
                                                        ? formatBytes(
                                                              array.used,
                                                          )
                                                        : "N/A"}
                                                </p>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    {/each}
                </div>
            {/if}

            <!-- Storage Pools Tab -->
        {:else if activeTab === "pools"}
            {#if poolsLoading}
                <div class="flex justify-center items-center h-64">
                    <div
                        class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"
                    ></div>
                </div>
            {:else if poolsError}
                <div class="bg-red-50 border border-red-200 rounded-md p-4">
                    <div class="text-red-700">Error: {poolsError}</div>
                </div>
            {:else if storagePoolsSafe.length === 0}
                <div class="text-center py-12">
                    <h3 class="text-lg font-medium text-gray-900 mb-2">
                        No storage pools found
                    </h3>
                    <p class="text-gray-600 dark:text-gray-300 mb-4">
                        Create a storage pool to get started.
                    </p>
                </div>
            {:else}
                <div class="space-y-4">
                    {#each storagePoolsSafe as pool}
                        <div
                            class="bg-white dark:bg-gray-800 shadow-lg hover:shadow-xl transition-shadow duration-200 rounded-lg p-6 border border-gray-100 dark:border-gray-700"
                        >
                            <div class="flex justify-between items-start">
                                <div class="flex items-start space-x-4">
                                    <!-- Pool Type Icon -->
                                    <div class="flex-shrink-0">
                                        {#if pool.type === "mergerfs"}
                                            <div
                                                class="w-12 h-12 bg-blue-100 dark:bg-blue-900 rounded-lg flex items-center justify-center"
                                            >
                                                <svg
                                                    class="w-6 h-6 text-blue-600 dark:text-blue-400"
                                                    fill="currentColor"
                                                    viewBox="0 0 20 20"
                                                >
                                                    <path
                                                        d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zM2 11a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z"
                                                    />
                                                </svg>
                                            </div>
                                        {:else if pool.type === "lvm"}
                                            <div
                                                class="w-12 h-12 bg-purple-100 dark:bg-purple-900 rounded-lg flex items-center justify-center"
                                            >
                                                <svg
                                                    class="w-6 h-6 text-purple-600 dark:text-purple-400"
                                                    fill="currentColor"
                                                    viewBox="0 0 20 20"
                                                >
                                                    <path
                                                        d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z"
                                                    />
                                                </svg>
                                            </div>
                                        {:else}
                                            <div
                                                class="w-12 h-12 bg-gray-100 dark:bg-gray-700 rounded-lg flex items-center justify-center"
                                            >
                                                <svg
                                                    class="w-6 h-6 text-gray-600 dark:text-gray-400"
                                                    fill="currentColor"
                                                    viewBox="0 0 20 20"
                                                >
                                                    <path
                                                        d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z"
                                                    />
                                                </svg>
                                            </div>
                                        {/if}
                                    </div>

                                    <!-- Pool Info -->
                                    <div>
                                        <h3
                                            class="text-lg font-semibold text-gray-900 dark:text-white mb-2"
                                        >
                                            {pool.name}
                                        </h3>
                                        <p
                                            class="text-gray-600 dark:text-gray-300 mb-4"
                                        >
                                            {pool.type || "Unknown type"} Pool
                                        </p>
                                        <div
                                            class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm"
                                        >
                                            <div>
                                                <p
                                                    class="text-gray-500 dark:text-gray-400"
                                                >
                                                    Size
                                                </p>
                                                <p
                                                    class="font-medium text-gray-900 dark:text-white"
                                                >
                                                    {formatBytes(pool.size)}
                                                </p>
                                            </div>
                                            <div>
                                                <p
                                                    class="text-gray-500 dark:text-gray-400"
                                                >
                                                    Used
                                                </p>
                                                <p
                                                    class="font-medium text-gray-900 dark:text-white"
                                                >
                                                    {formatBytes(pool.used)}
                                                </p>
                                            </div>
                                            <div>
                                                <p
                                                    class="text-gray-500 dark:text-gray-400"
                                                >
                                                    Available
                                                </p>
                                                <p
                                                    class="font-medium text-gray-900 dark:text-white"
                                                >
                                                    {formatBytes(
                                                        pool.available,
                                                    )}
                                                </p>
                                            </div>
                                            <div>
                                                <p
                                                    class="text-gray-500 dark:text-gray-400"
                                                >
                                                    Status
                                                </p>
                                                <div
                                                    class="flex items-center space-x-1"
                                                >
                                                    <div
                                                        class="w-2 h-2 rounded-full {pool.state ===
                                                        'active'
                                                            ? 'bg-green-500'
                                                            : 'bg-gray-400'}"
                                                    ></div>
                                                    <p
                                                        class="font-medium text-gray-900 dark:text-white"
                                                    >
                                                        {pool.state ||
                                                            "Inactive"}
                                                    </p>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                <!-- Action Buttons -->
                                <div class="relative">
                                    <button
                                        class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
                                        on:click|stopPropagation={() =>
                                            toggleDropdown(pool.name)}
                                        aria-label="Pool actions menu"
                                    >
                                        <svg
                                            class="w-5 h-5"
                                            fill="currentColor"
                                            viewBox="0 0 20 20"
                                        >
                                            <path
                                                d="M10 6a2 2 0 110-4 2 2 0 010 4zM10 12a2 2 0 110-4 2 2 0 010 4zM10 18a2 2 0 110-4 2 2 0 010 4z"
                                            />
                                        </svg>
                                    </button>
                                    <!-- Dropdown Menu -->
                                    {#if openDropdown === pool.name}
                                        <div
                                            class="absolute right-0 mt-2 w-48 bg-white dark:bg-gray-800 rounded-md shadow-lg z-10"
                                        >
                                            <div class="py-1">
                                                <button
                                                    class="block w-full text-left px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700"
                                                    on:click={() => {
                                                        closeDropdowns();
                                                        handlePoolAction(
                                                            "mount",
                                                            pool,
                                                        );
                                                    }}
                                                >
                                                    {pool.state === "active"
                                                        ? "Unmount"
                                                        : "Mount"}
                                                </button>
                                                <button
                                                    class="block w-full text-left px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700"
                                                    on:click={() => {
                                                        closeDropdowns();
                                                        handlePoolAction(
                                                            "details",
                                                            pool,
                                                        );
                                                    }}
                                                >
                                                    View Details
                                                </button>
                                                <button
                                                    class="block w-full text-left px-4 py-2 text-sm text-red-600 dark:text-red-400 hover:bg-gray-100 dark:hover:bg-gray-700"
                                                    on:click={() => {
                                                        closeDropdowns();
                                                        handlePoolAction(
                                                            "delete",
                                                            pool,
                                                        );
                                                    }}
                                                >
                                                    Delete Pool
                                                </button>
                                            </div>
                                        </div>
                                    {/if}
                                </div>
                            </div>
                        </div>
                    {/each}
                </div>
            {/if}
        {/if}

        <!-- Delete Confirmation Modal -->
        {#if showDeleteModal}
            <div
                class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
            >
                <div
                    class="bg-white dark:bg-gray-800 rounded-lg p-6 max-w-md w-full mx-4"
                >
                    <div class="flex items-center mb-4">
                        <svg
                            class="w-6 h-6 text-red-600 mr-3"
                            fill="currentColor"
                            viewBox="0 0 20 20"
                        >
                            <path
                                fill-rule="evenodd"
                                d="M8.257 3.099c.765-1.36 1.604-2.856 1.604-2.856.765-1.36 1.604.765-1.36-2.856 0-2.856.765-1.36 2.856.765 1.36 2.856-2.856.765-1.36 2.856v-8.48c0-3.732-3.099-8.48-8.48 0-8.48 3.099 8.48 8.48 3.099 8.48z"
                            />
                        </svg>
                        <h3
                            class="text-lg font-semibold text-gray-900 dark:text-white"
                        >
                            Delete Storage Pool
                        </h3>
                    </div>
                    <p class="text-gray-600 dark:text-gray-300 mb-4">
                        Are you sure you want to delete the storage pool "{poolToDelete?.name}"?
                        This action cannot be undone.
                    </p>
                    <div class="mb-4">
                        <label
                            for="deleteConfirmation"
                            class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                        >
                            Type <span class="text-red-600">*</span> to confirm:
                        </label>
                        <input
                            id="deleteConfirmation"
                            type="text"
                            bind:value={deleteConfirmation}
                            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white sm:text-sm"
                            placeholder="DELETE POOL_NAME"
                        />
                    </div>
                    <div class="flex justify-end space-x-3">
                        <button
                            type="button"
                            class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-gray-600"
                            on:click={cancelDelete}
                        >
                            Cancel
                        </button>
                        <button
                            type="button"
                            class="px-4 py-2 bg-red-600 text-white rounded-md shadow-sm hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50"
                            disabled={deleteConfirmation !==
                                `DELETE ${poolToDelete?.name.toUpperCase()}`}
                            on:click={confirmDeletePool}
                        >
                            Delete Pool
                        </button>
                    </div>
                </div>
            </div>
        {/if}

        <!-- Details Modal -->
        {#if showDetailsModal}
            <div
                class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
            >
                <div
                    class="bg-white dark:bg-gray-800 rounded-lg p-6 max-w-2xl w-full mx-4"
                >
                    <div class="flex items-center justify-between mb-4">
                        <h3
                            class="text-lg font-semibold text-gray-900 dark:text-white"
                        >
                            Storage Pool Details
                        </h3>
                        <button
                            type="button"
                            class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                            on:click={closeDetailsModal}
                            aria-label="Close details modal"
                        >
                            <svg
                                class="w-6 h-6"
                                fill="currentColor"
                                viewBox="0 0 20 20"
                            >
                                <path
                                    fill-rule="evenodd"
                                    d="M4.293 4.293a1 1 0 011.414 1.414 1.414 1.414 1.414 1.414 1.414 1.414-1.414 1.414-1.414 1.414-1.414 1.414v-8.48c0-3.732-3.099-8.48-8.48 0-8.48 3.099 8.48 8.48 3.099 8.48z"
                                />
                            </svg>
                        </button>
                    </div>
                    <div class="space-y-4">
                        <div class="grid grid-cols-2 gap-4">
                            <div>
                                <p
                                    class="text-sm font-medium text-gray-500 dark:text-gray-400"
                                >
                                    Name
                                </p>
                                <p class="text-gray-900 dark:text-white">
                                    {selectedPoolDetails?.name}
                                </p>
                            </div>
                            <div>
                                <p
                                    class="text-sm font-medium text-gray-500 dark:text-gray-400"
                                >
                                    Type
                                </p>
                                <p class="text-gray-900 dark:text-white">
                                    {selectedPoolDetails?.type || "Unknown"}
                                </p>
                            </div>
                            <div>
                                <p
                                    class="text-sm font-medium text-gray-500 dark:text-gray-400"
                                >
                                    State
                                </p>
                                <p class="text-gray-900 dark:text-white">
                                    {selectedPoolDetails?.state || "Unknown"}
                                </p>
                            </div>
                            <div>
                                <p
                                    class="text-sm font-medium text-gray-500 dark:text-gray-400"
                                >
                                    Mount Point
                                </p>
                                <p class="text-gray-900 dark:text-white">
                                    {selectedPoolDetails?.mount_point ||
                                        "Not mounted"}
                                </p>
                            </div>
                        </div>
                        <div class="grid grid-cols-2 gap-4">
                            <div>
                                <p
                                    class="text-sm font-medium text-gray-500 dark:text-gray-400"
                                >
                                    Size
                                </p>
                                <p class="text-gray-900 dark:text-white">
                                    {selectedPoolDetails?.size || "N/A"}
                                </p>
                            </div>
                            <div>
                                <p
                                    class="text-sm font-medium text-gray-500 dark:text-gray-400"
                                >
                                    Used
                                </p>
                                <p class="text-gray-900 dark:text-white">
                                    {selectedPoolDetails?.used || "N/A"}
                                </p>
                            </div>
                            <div>
                                <p
                                    class="text-sm font-medium text-gray-500 dark:text-gray-400"
                                >
                                    Available
                                </p>
                                <p class="text-gray-900 dark:text-white">
                                    {selectedPoolDetails?.available || "N/A"}
                                </p>
                            </div>
                            <div>
                                <p
                                    class="text-sm font-medium text-gray-500 dark:text-gray-400"
                                >
                                    Created
                                </p>
                                <p class="text-gray-900 dark:text-white">
                                    {selectedPoolDetails?.created_at
                                        ? new Date(
                                              selectedPoolDetails.created_at,
                                          ).toLocaleDateString()
                                        : "N/A"}
                                </p>
                            </div>
                        </div>
                        {#if selectedPoolDetails?.config}
                            <div>
                                <p
                                    class="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2"
                                >
                                    Configuration
                                </p>
                                <pre
                                    class="bg-gray-100 dark:bg-gray-700 p-3 rounded text-xs overflow-auto">{JSON.stringify(
                                        selectedPoolDetails.config,
                                        null,
                                        2,
                                    )}</pre>
                            </div>
                        {/if}
                    </div>
                    <div class="flex justify-end mt-6">
                        <button
                            type="button"
                            class="px-4 py-2 bg-blue-600 text-white rounded-md shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                            on:click={closeDetailsModal}
                        >
                            Close
                        </button>
                    </div>
                </div>
            </div>
        {/if}
    </div>
</div>
