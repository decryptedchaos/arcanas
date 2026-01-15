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
    let showCreatePoolModal = false;
    let showCreateRAIDModal = false;

    // RAID form data
    let newRAID = {
        name: "",
        level: "raid0",
        devices: []
    };

    // RAID array actions
    let showRAIDDeleteModal = false;
    let raidToDelete = null;
    let raidDeleteConfirmation = "";

    // Notification system
    let notifications = [];
    let notificationId = 0;

    function showNotification(message, type = 'info') {
        const id = ++notificationId;
        notifications = [...notifications, { id, message, type }];
        
        // Auto-remove after 5 seconds
        setTimeout(() => {
            notifications = notifications.filter(n => n.id !== id);
        }, 5000);
    }

    function clearNotification(id) {
        notifications = notifications.filter(n => n.id !== id);
    }

    async function loadDiskStats() {
        try {
            error = null;
            const newStats = await diskAPI.getDiskStats();
            // More efficient comparison - check length first, then individual items
            if (newStats.length !== diskStats.length || 
                newStats.some((disk, index) => JSON.stringify(disk) !== JSON.stringify(diskStats[index]))) {
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
            showNotification("Failed to delete pool: " + err.message, "error");
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

    // Form data for creating storage pool
    let newPool = {
        name: "",
        type: "mergerfs",
        devices: [], // Array of selected device paths
        config: ""
    };

    async function createStoragePool() {
        try {
            if (!newPool.name) {
                showNotification("Please enter a pool name", "warning");
                return;
            }
            if (!newPool.devices || newPool.devices.length === 0) {
                showNotification("Please select at least one disk", "warning");
                return;
            }

            // devices is already an array
            const poolData = {
                name: newPool.name,
                type: newPool.type,
                devices: newPool.devices,
                config: newPool.config
            };

            await storageAPI.createPool(poolData);
            showCreatePoolModal = false;
            resetPoolForm();
            await loadStoragePools();
        } catch (err) {
            console.error("Failed to create storage pool:", err);
            showNotification("Failed to create storage pool: " + err.message, "error");
        }
    }

    function resetPoolForm() {
        newPool = {
            name: "",
            type: "mergerfs",
            devices: [],
            config: ""
        };
    }

    function toggleDevice(devicePath) {
        if (newPool.devices.includes(devicePath)) {
            newPool.devices = newPool.devices.filter(d => d !== devicePath);
        } else {
            newPool.devices = [...newPool.devices, devicePath];
        }
    }

    // RAID Functions
    async function createRAIDArray() {
        try {
            if (!newRAID.name) {
                showNotification("Please enter a RAID array name", "warning");
                return;
            }
            if (!newRAID.devices || newRAID.devices.length === 0) {
                showNotification("Please select at least one disk", "warning");
                return;
            }

            const raidData = {
                name: newRAID.name,
                level: newRAID.level,
                devices: newRAID.devices
            };

            await diskAPI.createRAIDArray(raidData);
            showCreateRAIDModal = false;
            resetRAIDForm();
            await loadRAIDArrays();
            await loadDiskStats(); // Refresh to show new md device
            showNotification("RAID array created successfully", "success");
        } catch (err) {
            console.error("Failed to create RAID array:", err);
            showNotification("Failed to create RAID array: " + err.message, "error");
        }
    }

    function resetRAIDForm() {
        newRAID = {
            name: "",
            level: "raid0",
            devices: []
        };
    }

    function toggleRAIDDevice(devicePath) {
        if (newRAID.devices.includes(devicePath)) {
            newRAID.devices = newRAID.devices.filter(d => d !== devicePath);
        } else {
            newRAID.devices = [...newRAID.devices, devicePath];
        }
    }

    function handleRAIDDelete(array) {
        raidToDelete = array;
        raidDeleteConfirmation = `DELETE ${array.name.toUpperCase()}`;
        showRAIDDeleteModal = true;
    }

    async function confirmRAIDDelete() {
        if (!raidToDelete) return;

        try {
            await diskAPI.deleteRAIDArray(raidToDelete.name);
            showRAIDDeleteModal = false;
            raidToDelete = null;
            raidDeleteConfirmation = "";
            await loadRAIDArrays();
            await loadDiskStats();
            showNotification("RAID array deleted successfully", "success");
        } catch (err) {
            console.error("Failed to delete RAID array:", err);
            showNotification("Failed to delete RAID array: " + err.message, "error");
        }
    }

    function cancelRAIDDelete() {
        showRAIDDeleteModal = false;
        raidToDelete = null;
        raidDeleteConfirmation = "";
    }

    function toggleDiskExpansion(diskName) {
        if (!diskName) return; // Guard against undefined
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

    // Close dropdowns on outside click
    function handleOutsideClick(e) {
        closeDropdowns();
    }

    // Close dropdowns on Escape key
    function handleEscapeKey(e) {
        if (e.key === "Escape") {
            closeDropdowns();
        }
    }
</script>

<div class="p-6" role="main" tabindex="-1">
    <!-- Notifications -->
    {#if notifications.length > 0}
        <div class="fixed top-4 right-4 z-50 space-y-2">
            {#each notifications as notification (notification.id)}
                <div 
                    class="p-4 rounded-md shadow-lg flex items-center justify-between max-w-sm animate-pulse
                        {notification.type === 'error' ? 'bg-red-50 border border-red-200 text-red-700' : ''}
                        {notification.type === 'warning' ? 'bg-yellow-50 border border-yellow-200 text-yellow-700' : ''}
                        {notification.type === 'success' ? 'bg-green-50 border border-green-200 text-green-700' : ''}
                        {notification.type === 'info' ? 'bg-blue-50 border border-blue-200 text-blue-700' : ''}"
                >
                    <span class="text-sm font-medium">{notification.message}</span>
                    <button 
                        on:click={() => clearNotification(notification.id)}
                        class="ml-4 text-sm underline hover:no-underline"
                    >
                        Dismiss
                    </button>
                </div>
            {/each}
        </div>
    {/if}

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
            <button
                on:click={() => (showCreateRAIDModal = true)}
                class="btn btn-primary"
            >
                Create RAID Array
            </button>
        {:else if activeTab === "pools"}
            <button
                on:click={() => (showCreatePoolModal = true)}
                class="btn btn-primary"
            >
                Create Storage Pool
            </button>
        {/if}
    </div>

    <!-- Tab Content -->
    <div role="region">
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
                                                {disk.name || disk.device || "Unknown"}
                                            </h3>
                                            <span
                                                class="px-2 py-1 text-xs font-medium rounded-full {getHealthColor(
                                                    disk.smart?.status ||
                                                        disk.health ||
                                                        "Unknown",
                                                )}"
                                            >
                                                {disk.smart?.status || disk.health || "Unknown"}
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
                                                    {formatBytes(disk.size || 0)}
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
                                                    {disk.usage
                                                        ? disk.usage.toFixed(1) + "%"
                                                        : "N/A"}
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
                                                    {disk.smart?.temperature
                                                        ? disk.smart.temperature +
                                                            "°C"
                                                        : "N/A"}
                                                </p>
                                            </div>
                                            <div>
                                                <p
                                                    class="text-gray-500 dark:text-gray-400"
                                                >
                                                    Filesystem
                                                </p>
                                                <p
                                                    class="font-medium text-gray-900 dark:text-white"
                                                >
                                                    {disk.filesystem || "Unknown"}
                                                </p>
                                            </div>
                                        </div>
                                    </div>
                                    <button
                                        on:click={() =>
                                            toggleDiskExpansion(disk.name || disk.device)}
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
                            {#if expandedDisks.has(disk.name || disk.device)}
                                <div
                                    class="border-t border-gray-200 dark:border-gray-700 p-6 bg-gray-50 dark:bg-gray-900/50"
                                >
                                    <h4 class="font-medium text-gray-900 dark:text-white mb-4">
                                        Detailed Information
                                    </h4>
                                    <div
                                        class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm"
                                    >
                                        <div>
                                            <p
                                                class="text-gray-500 dark:text-gray-400"
                                            >
                                                Device Path
                                            </p>
                                            <p class="font-medium text-gray-900 dark:text-white">
                                                {disk.device || "Unknown"}
                                            </p>
                                        </div>
                                        <div>
                                            <p
                                                class="text-gray-500 dark:text-gray-400"
                                            >
                                                Mount Point
                                            </p>
                                            <p class="font-medium text-gray-900 dark:text-white">
                                                {disk.mountpoint || "Not mounted"}
                                            </p>
                                        </div>
                                        <div>
                                            <p
                                                class="text-gray-500 dark:text-gray-400"
                                            >
                                                Filesystem
                                            </p>
                                            <p class="font-medium text-gray-900 dark:text-white">
                                                {disk.filesystem || "Unknown"}
                                            </p>
                                        </div>
                                        <div>
                                            <p
                                                class="text-gray-500 dark:text-gray-400"
                                            >
                                                Used Space
                                            </p>
                                            <p class="font-medium text-gray-900 dark:text-white">
                                                {formatBytes(disk.used || 0)}
                                            </p>
                                        </div>
                                        <div>
                                            <p
                                                class="text-gray-500 dark:text-gray-400"
                                            >
                                                Available Space
                                            </p>
                                            <p class="font-medium text-gray-900 dark:text-white">
                                                {formatBytes(disk.available || 0)}
                                            </p>
                                        </div>
                                        <div>
                                            <p
                                                class="text-gray-500 dark:text-gray-400"
                                            >
                                                Usage Percentage
                                            </p>
                                            <p class="font-medium text-gray-900 dark:text-white">
                                                {disk.usage
                                                    ? disk.usage.toFixed(2) + "%"
                                                    : "N/A"}
                                            </p>
                                        </div>
                                        <div>
                                            <p
                                                class="text-gray-500 dark:text-gray-400"
                                            >
                                                Read-Only
                                            </p>
                                            <p class="font-medium text-gray-900 dark:text-white">
                                                {disk.read_only
                                                    ? "Yes"
                                                    : "No"}
                                            </p>
                                        </div>
                                        <div>
                                            <p
                                                class="text-gray-500 dark:text-gray-400"
                                            >
                                                SMART Health
                                            </p>
                                            <p class="font-medium text-gray-900 dark:text-white">
                                                {disk.smart?.health
                                                    ? disk.smart.health + "%"
                                                    : "N/A"}
                                            </p>
                                        </div>
                                        <div>
                                            <p
                                                class="text-gray-500 dark:text-gray-400"
                                            >
                                                SMART Status
                                            </p>
                                            <p class="font-medium text-gray-900 dark:text-white">
                                                {disk.smart?.status || "Unknown"}
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
                                                        class="w-2 h-2 rounded-full {array.state ===
                                                        'active'
                                                            ? 'bg-green-500'
                                                            : 'bg-yellow-500'}"
                                                    ></div>
                                                    <p
                                                        class="font-medium text-gray-900 dark:text-white"
                                                    >
                                                        {array.state || "Unknown"}
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

                                <!-- RAID Actions -->
                                <button
                                    on:click={() => handleRAIDDelete(array)}
                                    class="text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300 p-2 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
                                    title="Delete RAID Array"
                                >
                                    <svg
                                        class="w-5 h-5"
                                        fill="none"
                                        stroke="currentColor"
                                        viewBox="0 0 24 24"
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
                                            Storage Pool
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
                                d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z"
                                clip-rule="evenodd"
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
                                    Storage Pool
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

        <!-- Create Storage Pool Modal -->
        {#if showCreatePoolModal}
            <div class="fixed inset-0 z-50 overflow-y-auto">
                <div class="flex items-center justify-center min-h-screen px-4">
                    <div
                        class="fixed inset-0 bg-gray-500 bg-opacity-75"
                        role="button"
                        tabindex="0"
                        on:click={() => {
                            showCreatePoolModal = false;
                            resetPoolForm();
                        }}
                        on:keydown={(e) =>
                            e.key === "Escape" && (showCreatePoolModal = false)}
                        aria-label="Close modal"
                    ></div>
                    <div
                        class="relative bg-white dark:bg-gray-800 rounded-lg max-w-2xl w-full p-6"
                        role="dialog"
                        aria-modal="true"
                        aria-labelledby="create-pool-title"
                        tabindex="-1"
                        on:click|stopPropagation
                        on:keydown|stopPropagation
                    >
                        <h3
                            id="create-pool-title"
                            class="text-lg font-medium text-gray-900 dark:text-white mb-4"
                        >
                            Create Storage Pool
                        </h3>
                        <form on:submit|preventDefault={createStoragePool} class="space-y-4">
                            <div>
                                <label
                                    for="poolName"
                                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                                >
                                    Pool Name
                                </label>
                                <input
                                    id="poolName"
                                    type="text"
                                    bind:value={newPool.name}
                                    placeholder="my-pool"
                                    class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
                                    required
                                />
                            </div>

                            <div>
                                <label
                                    for="poolType"
                                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                                >
                                    Pool Type
                                </label>
                                <select
                                    id="poolType"
                                    bind:value={newPool.type}
                                    class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
                                >
                                    <option value="mergerfs">Combined Storage</option>
                                </select>
                            </div>

                            <fieldset class="border-0 p-0 m-0">
                                <legend class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                    Select Disks
                                </legend>
                                {#if diskStats.length === 0}
                                    <div class="w-full p-4 border border-gray-300 dark:border-gray-600 rounded-md bg-gray-50 dark:bg-gray-700 text-gray-500 dark:text-gray-400 text-center">
                                        No disks available. Please add disks to the system first.
                                    </div>
                                {:else}
                                    <div class="border border-gray-300 dark:border-gray-600 rounded-md p-4 max-h-60 overflow-y-auto bg-gray-50 dark:bg-gray-700">
                                        {#each diskStats as disk}
                                            {#if disk.device && disk.device !== ""}
                                                <label class="flex items-center space-x-3 p-2 hover:bg-gray-100 dark:hover:bg-gray-600 rounded cursor-pointer">
                                                    <input
                                                        type="checkbox"
                                                        checked={newPool.devices.includes(disk.device)}
                                                        on:change={() => toggleDevice(disk.device)}
                                                        class="rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500 dark:bg-gray-800"
                                                    />
                                                    <div class="flex-1">
                                                        <div class="flex items-center space-x-2">
                                                            <span class="font-medium text-gray-900 dark:text-white">
                                                                {disk.name || disk.device}
                                                            </span>
                                                            <span class="text-xs text-gray-500 dark:text-gray-400">
                                                                ({disk.model || "Unknown model"})
                                                            </span>
                                                        </div>
                                                        <div class="text-xs text-gray-500 dark:text-gray-400">
                                                            {disk.device} • {formatBytes(disk.size || 0)}
                                                            {#if disk.filesystem}
                                                                • {disk.filesystem}
                                                            {/if}
                                                        </div>
                                                    </div>
                                                </label>
                                            {/if}
                                        {/each}
                                    </div>
                                    <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
                                        {newPool.devices.length} disk(s) selected
                                    </p>
                                {/if}
                            </fieldset>

                            <div>
                                <label
                                    for="poolConfig"
                                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                                >
                                    Mount Options (optional)
                                </label>
                                <input
                                    id="poolConfig"
                                    type="text"
                                    bind:value={newPool.config}
                                    placeholder="defaults,allow_other,use_ino"
                                    class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
                                />
                            </div>

                            <div class="flex justify-end space-x-3 mt-6">
                                <button
                                    type="button"
                                    on:click={() => {
                                        showCreatePoolModal = false;
                                        resetPoolForm();
                                    }}
                                    class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-gray-600"
                                >
                                    Cancel
                                </button>
                                <button
                                    type="submit"
                                    class="px-4 py-2 bg-blue-600 text-white rounded-md shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                                >
                                    Create Pool
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        {/if}

        <!-- RAID Delete Confirmation Modal -->
        {#if showRAIDDeleteModal}
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
                                d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z"
                                clip-rule="evenodd"
                            />
                        </svg>
                        <h3
                            class="text-lg font-semibold text-gray-900 dark:text-white"
                        >
                            Delete RAID Array
                        </h3>
                    </div>
                    <p class="text-gray-600 dark:text-gray-300 mb-4">
                        Are you sure you want to delete the RAID array "{raidToDelete?.name}"?
                        This will stop the array and zero the superblock on all member devices.
                        This action cannot be undone.
                    </p>
                    <div class="mb-4">
                        <label
                            for="raidDeleteConfirmation"
                            class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                        >
                            Type <span class="text-red-600">*</span> to confirm:
                        </label>
                        <input
                            id="raidDeleteConfirmation"
                            type="text"
                            bind:value={raidDeleteConfirmation}
                            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white sm:text-sm"
                            placeholder="DELETE {raidToDelete?.name.toUpperCase()}"
                        />
                    </div>
                    <div class="flex justify-end space-x-3">
                        <button
                            type="button"
                            class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-gray-600"
                            on:click={cancelRAIDDelete}
                        >
                            Cancel
                        </button>
                        <button
                            type="button"
                            class="px-4 py-2 bg-red-600 text-white rounded-md shadow-sm hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50"
                            disabled={raidDeleteConfirmation !==
                                `DELETE ${raidToDelete?.name.toUpperCase()}`}
                            on:click={confirmRAIDDelete}
                        >
                            Delete RAID Array
                        </button>
                    </div>
                </div>
            </div>
        {/if}

        <!-- Create RAID Array Modal -->
        {#if showCreateRAIDModal}
            <div
                class="fixed inset-0 z-50 overflow-y-auto"
                on:click={() => {
                    showCreateRAIDModal = false;
                    resetRAIDForm();
                }}
            >
                <div class="flex items-center justify-center min-h-screen px-4">
                    <div
                        class="fixed inset-0 bg-gray-500 bg-opacity-75"
                        role="button"
                        tabindex="0"
                        on:click={() => {
                            showCreateRAIDModal = false;
                            resetRAIDForm();
                        }}
                        on:keydown={(e) =>
                            e.key === "Escape" && (showCreateRAIDModal = false)}
                        aria-label="Close modal"
                    ></div>
                    <div
                        class="relative bg-white dark:bg-gray-800 rounded-lg max-w-2xl w-full p-6"
                        role="dialog"
                        aria-modal="true"
                        aria-labelledby="create-raid-title"
                        tabindex="-1"
                        on:click|stopPropagation
                        on:keydown|stopPropagation
                    >
                        <h3
                            id="create-raid-title"
                            class="text-lg font-medium text-gray-900 dark:text-white mb-4"
                        >
                            Create RAID Array
                        </h3>
                        <form on:submit|preventDefault={createRAIDArray} class="space-y-4">
                            <div>
                                <label
                                    for="raidName"
                                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                                >
                                    Array Name (e.g., md0, md1)
                                </label>
                                <input
                                    id="raidName"
                                    type="text"
                                    bind:value={newRAID.name}
                                    placeholder="md0"
                                    class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
                                    required
                                />
                            </div>

                            <div>
                                <label
                                    for="raidLevel"
                                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                                >
                                    RAID Level
                                </label>
                                <select
                                    id="raidLevel"
                                    bind:value={newRAID.level}
                                    class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
                                >
                                    <option value="raid0">RAID 0 (Striping)</option>
                                    <option value="raid1">RAID 1 (Mirroring)</option>
                                    <option value="raid5">RAID 5 (Parity)</option>
                                    <option value="raid6">RAID 6 (Dual Parity)</option>
                                    <option value="raid10">RAID 10 (Mirroring + Striping)</option>
                                </select>
                            </div>

                            <fieldset class="border-0 p-0 m-0">
                                <legend class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                    Select Disks
                                </legend>
                                {#if diskStats.length === 0}
                                    <div class="w-full p-4 border border-gray-300 dark:border-gray-600 rounded-md bg-gray-50 dark:bg-gray-700 text-gray-500 dark:text-gray-400 text-center">
                                        No disks available. Please add disks to the system first.
                                    </div>
                                {:else}
                                    <div class="border border-gray-300 dark:border-gray-600 rounded-md p-4 max-h-60 overflow-y-auto bg-gray-50 dark:bg-gray-700">
                                        {#each diskStats as disk}
                                            {#if disk.device && disk.device !== "" && !disk.device.includes("/dev/md")}
                                                <label class="flex items-center space-x-3 p-2 hover:bg-gray-100 dark:hover:bg-gray-600 rounded cursor-pointer">
                                                    <input
                                                        type="checkbox"
                                                        checked={newRAID.devices.includes(disk.device)}
                                                        on:change={() => toggleRAIDDevice(disk.device)}
                                                        class="rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500 dark:bg-gray-800"
                                                    />
                                                    <div class="flex-1">
                                                        <div class="flex items-center space-x-2">
                                                            <span class="font-medium text-gray-900 dark:text-white">
                                                                {disk.name || disk.device}
                                                            </span>
                                                            <span class="text-xs text-gray-500 dark:text-gray-400">
                                                                ({disk.model || "Unknown model"})
                                                            </span>
                                                        </div>
                                                        <div class="text-xs text-gray-500 dark:text-gray-400">
                                                            {disk.device} • {formatBytes(disk.size || 0)}
                                                            {#if disk.filesystem}
                                                                • {disk.filesystem}
                                                            {/if}
                                                        </div>
                                                    </div>
                                                </label>
                                            {/if}
                                        {/each}
                                    </div>
                                    <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
                                        {newRAID.devices.length} disk(s) selected
                                    </p>
                                {/if}
                            </fieldset>

                            <div class="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-md p-3">
                                <div class="flex">
                                    <svg
                                        class="w-5 h-5 text-yellow-400 mr-2"
                                        fill="currentColor"
                                        viewBox="0 0 20 20"
                                    >
                                        <path
                                            fill-rule="evenodd"
                                            d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
                                            clip-rule="evenodd"
                                        />
                                    </svg>
                                    <div class="text-sm text-yellow-800 dark:text-yellow-200">
                                        <strong>Warning:</strong> Creating a RAID array will erase all data on the selected disks.
                                        Make sure you have backups if needed.
                                    </div>
                                </div>
                            </div>

                            <div class="flex justify-end space-x-3 mt-6">
                                <button
                                    type="button"
                                    on:click={() => {
                                        showCreateRAIDModal = false;
                                        resetRAIDForm();
                                    }}
                                    class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-gray-600"
                                >
                                    Cancel
                                </button>
                                <button
                                    type="submit"
                                    class="px-4 py-2 bg-blue-600 text-white rounded-md shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                                >
                                    Create RAID Array
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        {/if}
    </div>
</div>

<svelte:body on:click={handleOutsideClick} on:keydown={handleEscapeKey} />
