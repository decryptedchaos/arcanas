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

    // RAID data
    let raidArrays = [];
    let raidLoading = true;
    let raidError = null;

    // Storage pools data
    let storagePools = [];
    let poolsLoading = true;
    let poolsError = null;

    // Volume Groups data (for iSCSI LUNs)
    let volumeGroups = [];
    let vgsLoading = true;
    let vgsError = null;
    let showCreateVGModal = false;
    let showVGDeleteModal = false;
    let vgToDelete = null;
    let vgCreating = false;
    let availableVGDevices = [];

    // VG form data
    let newVG = {
        name: "",
        devices: []
    };

    // Modal state for pool actions
    let showDeleteModal = false;
    let poolToDelete = null;
    let deleteConfirmation = "";
    let showDetailsModal = false;
    let selectedPoolDetails = null;
    let openDropdown = null; // Track which dropdown is open
    let showCreatePoolModal = false;
    let poolCreating = false; // Loading state for pool creation
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

    async function loadVolumeGroups() {
        try {
            vgsError = null;
            const result = await fetch('/api/volume-groups');
            if (!result.ok) throw new Error('Failed to load volume groups');
            volumeGroups = (await result.json()) || [];
        } catch (err) {
            vgsError = err.message;
            console.error("Failed to load volume groups:", err);
            volumeGroups = [];
        } finally {
            vgsLoading = false;
        }
    }

    async function loadAvailableVGDevices() {
        try {
            const result = await fetch('/api/volume-groups/available-devices');
            if (!result.ok) throw new Error('Failed to load available devices');
            availableVGDevices = (await result.json()) || [];
        } catch (err) {
            console.error("Failed to load available devices:", err);
            availableVGDevices = [];
        }
    }

    async function createVG() {
        try {
            vgCreating = true;
            const result = await fetch('/api/volume-groups', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(newVG)
            });
            if (!result.ok) {
                const err = await result.text();
                throw new Error(err || 'Failed to create volume group');
            }
            showNotification('Volume group created successfully', 'success');
            showCreateVGModal = false;
            resetVGForm();
            await loadVolumeGroups();
        } catch (err) {
            showNotification(err.message, 'error');
        } finally {
            vgCreating = false;
        }
    }

    async function deleteVG() {
        if (!vgToDelete) return;

        try {
            const result = await fetch(`/api/volume-groups/${vgToDelete.name}`, {
                method: 'DELETE'
            });
            if (!result.ok) {
                const err = await result.text();
                throw new Error(err || 'Failed to delete volume group');
            }
            showNotification(`Volume group ${vgToDelete.name} deleted`, 'success');
            showVGDeleteModal = false;
            vgToDelete = null;
            await loadVolumeGroups();
        } catch (err) {
            showNotification(err.message, 'error');
        }
    }

    function resetVGForm() {
        newVG = {
            name: "",
            devices: []
        };
    }

    function handlePoolAction(action, pool) {
        switch (action) {
            case "mount":
                // Toggle mount/unmount
                console.log("Toggle mount for pool:", pool.name);
                // TODO: Implement mount/unmount logic
                break;
            case "edit":
                editingPool = pool;
                editedPool = {
                    name: pool.name,
                    config: pool.config || ""
                };
                showEditPoolModal = true;
                closeDropdowns();
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

    async function saveEditedPool() {
        if (!editingPool) return;

        try {
            if (!editedPool.name) {
                showNotification("Pool name is required", "warning");
                return;
            }

            const updateData = {
                name: editedPool.name,
                config: editedPool.config
            };

            poolEditing = true;
            await storageAPI.updatePool(editingPool.name, updateData);
            showEditPoolModal = false;
            editingPool = null;
            editedPool = { name: "", config: "" };
            await loadStoragePools();
            showNotification("Storage pool updated successfully", "success");
        } catch (err) {
            console.error("Failed to update pool:", err);
            showNotification("Failed to update pool: " + err.message, "error");
        } finally {
            poolEditing = false;
        }
    }

    function cancelEditPool() {
        showEditPoolModal = false;
        editingPool = null;
        editedPool = { name: "", config: "" };
    }

    async function setPoolUsage(pool, newMode) {
        try {
            await fetch(`/api/storage-pools/${pool.name}/export-mode`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ export_mode: newMode })
            });

            // Update the pool in the local state
            const poolIndex = storagePools.findIndex(p => p.name === pool.name);
            if (poolIndex !== -1) {
                storagePools[poolIndex].export_mode = newMode;
            }

            showNotification(`Pool "${pool.name}" export mode changed to ${newMode}`, "success");
        } catch (err) {
            console.error("Failed to set export mode:", err);
            showNotification("Failed to change export mode: " + err.message, "error");
            // Revert the select by reloading pools
            await loadStoragePools();
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

    function closeDropdowns() {
        openDropdown = null;
    }

    // Get available (unmounted) devices for pool creation
    $: availableDevices = diskStats.filter(disk => {
        // Must have a device path
        if (!disk.device || disk.device === "") return false;

        // Exclude RAID member disks (they're part of an array)
        if (disk.filesystem === "linux_raid_member") return false;

        // Exclude devices that are already mounted
        if (disk.mountpoint && disk.mountpoint !== "") return false;

        // Include MD RAID devices (they're virtual, not members)
        if (disk.device && disk.device.startsWith("/dev/md")) return true;

        // Include regular disks with filesystems
        if (disk.filesystem && disk.filesystem !== "unknown") return true;

        // Include raw disks without filesystems (can be formatted)
        return true;
    });

    // Form data for creating storage pool
    let newPool = {
        name: "",
        type: "mergerfs",
        devices: [], // Array of selected device paths
        config: ""
    };

    // Form data for editing storage pool
    let editingPool = null;
    let editedPool = {
        name: "",
        config: ""
    };
    let showEditPoolModal = false;
    let poolEditing = false;

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

            poolCreating = true;
            await storageAPI.createPool(poolData);
            showCreatePoolModal = false;
            resetPoolForm();
            await loadStoragePools();
        } catch (err) {
            console.error("Failed to create storage pool:", err);
            showNotification("Failed to create storage pool: " + err.message, "error");
        } finally {
            poolCreating = false;
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
            await diskAPI.deleteRAIDArray(raidToDelete.device);
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

    onMount(() => {
        // Initial load only
        loadDiskStats();
        loadRAIDArrays();
        loadStoragePools();
        loadVolumeGroups();
    });

    $: disks = Array.isArray(diskStats) ? diskStats : [];
    $: raidArraysSafe = Array.isArray(raidArrays) ? raidArrays : [];
    $: storagePoolsSafe = Array.isArray(storagePools) ? storagePools : [];

    // Close dropdowns on outside click
    function handleOutsideClick() {
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
    <div class="border-b border-gray-200 dark:border-border mb-6">
        <nav class="-mb-px flex space-x-8">
            <button
                class="py-2 px-1 border-b-2 font-medium text-sm {activeTab ===
                'disks'
                    ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                    : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-border'}"
                on:click={() => goto("?tab=disks", { replaceState: true })}
            >
                Disks
            </button>
            <button
                class="py-2 px-1 border-b-2 font-medium text-sm {activeTab ===
                'raid'
                    ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                    : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-border'}"
                on:click={() => goto("?tab=raid", { replaceState: true })}
            >
                RAID Arrays
            </button>
            <button
                class="py-2 px-1 border-b-2 font-medium text-sm {activeTab ===
                'pools'
                    ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                    : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-border'}"
                on:click={() => goto("?tab=pools", { replaceState: true })}
            >
                Storage Pools
            </button>
            <button
                class="py-2 px-1 border-b-2 font-medium text-sm {activeTab ===
                'vgs'
                    ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                    : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-border'}"
                on:click={() => goto("?tab=vgs", { replaceState: true })}
            >
                Volume Groups
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
                loadVolumeGroups();
            }}
            class="btn btn-secondary"
            disabled={loading || raidLoading || poolsLoading || vgsLoading}
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
        {:else if activeTab === "vgs"}
            <button
                on:click={async () => {
                    await loadAvailableVGDevices();
                    showCreateVGModal = true;
                }}
                class="btn btn-primary"
            >
                Create Volume Group
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
                    {#each disks as disk, index}
                        <!-- RAID Array Separator -->
                        {#if disk.device && disk.device.startsWith('/dev/md') && (index === 0 || !disks[index - 1].device || !disks[index - 1].device.startsWith('/dev/md'))}
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

                        <div
                            class="bg-white dark:bg-card shadow-lg hover:shadow-xl transition-shadow duration-200 rounded-lg p-6 border border-gray-100 dark:border-border {disk.device && disk.device.startsWith('/dev/md') ? 'ring-2 ring-indigo-200 dark:ring-indigo-800' : ''}"
                        >
                            <!-- Header Section -->
                            <div class="flex items-start justify-between mb-6">
                                <div class="flex items-center space-x-4">
                                    <!-- Disk Icon with glow effect -->
                                    <div
                                        class="relative"
                                    >
                                        <div
                                            class="absolute inset-0 bg-gradient-to-br from-green-400 to-emerald-500 rounded-xl blur opacity-25"
                                        ></div>
                                        <div
                                            class="relative w-14 h-14 bg-gradient-to-br from-green-100 to-emerald-100 dark:from-green-900/40 dark:to-emerald-900/40 rounded-xl flex items-center justify-center shadow-lg"
                                        >
                                            <svg
                                                class="w-7 h-7 text-green-600 dark:text-green-400"
                                                fill="none"
                                                stroke="currentColor"
                                                viewBox="0 0 24 24"
                                            >
                                                <path
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    stroke-width="2"
                                                    d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5h4M4 7h16"
                                                />
                                            </svg>
                                        </div>
                                    </div>

                                    <div>
                                        <h3
                                            class="text-xl font-bold text-gray-900 dark:text-white"
                                        >
                                            {disk.name || disk.device || "Unknown"}
                                        </h3>
                                        <div
                                            class="flex items-center flex-wrap gap-2 mt-1"
                                        >
                                            <!-- Filesystem Badge -->
                                            {#if disk.filesystem && disk.filesystem !== "Unknown"}
                                                <span
                                                    class="px-2 py-0.5 text-xs font-semibold rounded-full bg-gradient-to-r from-slate-500 to-slate-600 text-white"
                                                >
                                                    {disk.filesystem.toUpperCase()}
                                                </span>
                                            {/if}
                                            <!-- Health Badge -->
                                            {#if disk.smart?.status || disk.health}
                                                <span
                                                    class="px-2 py-0.5 text-xs font-semibold rounded-full {(disk.smart?.status === 'healthy' || disk.health === 'healthy')
                                                        ? 'bg-gradient-to-r from-emerald-500 to-green-500 text-white'
                                                        : (disk.smart?.status === 'warning' || disk.health === 'warning')
                                                        ? 'bg-gradient-to-r from-yellow-500 to-orange-500 text-white'
                                                        : 'bg-gradient-to-r from-red-500 to-rose-500 text-white'}"
                                                >
                                                    {(disk.smart?.status || disk.health || "").toUpperCase()}
                                                </span>
                                            {/if}
                                            <!-- Temperature Badge -->
                                            {#if disk.smart?.temperature}
                                                <span
                                                    class="text-xs text-gray-500 dark:text-gray-400"
                                                >
                                                    {disk.smart.temperature}°C
                                                </span>
                                            {/if}
                                            <!-- RAID Badge -->
                                            {#if disk.device && disk.device.startsWith('/dev/md')}
                                                <span
                                                    class="px-2 py-0.5 text-xs font-semibold rounded-full bg-gradient-to-r from-indigo-500 to-purple-500 text-white"
                                                >
                                                    RAID
                                                </span>
                                            {/if}
                                        </div>
                                    </div>
                                </div>

                                <!-- Usage Badge -->
                                {#if disk.size}
                                    <div
                                        class="flex items-center space-x-2 px-3 py-1.5 rounded-full bg-gray-100 dark:bg-muted"
                                    >
                                        <span
                                            class="text-sm font-medium text-gray-700 dark:text-gray-300"
                                        >
                                            {Math.round(
                                                ((disk.used || 0) / disk.size) *
                                                    100,
                                            )}%
                                        </span>
                                    </div>
                                {/if}
                            </div>

                            <!-- Stats Grid -->
                            <div
                                class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6"
                            >
                                <!-- Total Size Card -->
                                <div
                                    class="bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 rounded-xl p-4 border border-blue-100 dark:border-blue-800"
                                >
                                    <div class="flex items-center space-x-3">
                                        <div
                                            class="w-10 h-10 bg-blue-100 dark:bg-blue-800 rounded-lg flex items-center justify-center"
                                        >
                                            <svg
                                                class="w-5 h-5 text-blue-600 dark:text-blue-400"
                                                fill="none"
                                                stroke="currentColor"
                                                viewBox="0 0 24 24"
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
                                            <p
                                                class="text-xs text-blue-600 dark:text-blue-400 font-medium uppercase tracking-wide"
                                            >
                                                Capacity
                                            </p>
                                            <p
                                                class="text-lg font-bold text-gray-900 dark:text-white"
                                            >
                                                {formatBytes(disk.size || 0)}
                                            </p>
                                        </div>
                                    </div>
                                </div>

                                <!-- Used Card -->
                                <div
                                    class="bg-gradient-to-br from-purple-50 to-pink-50 dark:from-purple-900/20 dark:to-pink-900/20 rounded-xl p-4 border border-purple-100 dark:border-purple-800"
                                >
                                    <div class="flex items-center space-x-3">
                                        <div
                                            class="w-10 h-10 bg-purple-100 dark:bg-purple-800 rounded-lg flex items-center justify-center"
                                        >
                                            <svg
                                                class="w-5 h-5 text-purple-600 dark:text-purple-400"
                                                fill="none"
                                                stroke="currentColor"
                                                viewBox="0 0 24 24"
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
                                            <p
                                                class="text-xs text-purple-600 dark:text-purple-400 font-medium uppercase tracking-wide"
                                            >
                                                Used
                                            </p>
                                            <p
                                                class="text-lg font-bold text-gray-900 dark:text-white"
                                            >
                                                {formatBytes(disk.used || 0)}
                                            </p>
                                        </div>
                                    </div>
                                    <!-- Usage Bar -->
                                    {#if disk.size > 0}
                                        <div
                                            class="mt-3 w-full bg-gray-200 dark:bg-muted rounded-full h-2 overflow-hidden"
                                        >
                                            <div
                                                class="h-full bg-gradient-to-r from-purple-500 to-pink-500 rounded-full transition-all duration-300"
                                                style="width: {Math.min(
                                                    ((disk.used ||
                                                        0) /
                                                        disk.size) *
                                                        100,
                                                    100,
                                                )}%"
                                            ></div>
                                        </div>
                                    {/if}
                                </div>

                                <!-- Available Card -->
                                <div
                                    class="bg-gradient-to-br from-emerald-50 to-teal-50 dark:from-emerald-900/20 dark:to-teal-900/20 rounded-xl p-4 border border-emerald-100 dark:border-emerald-800"
                                >
                                    <div class="flex items-center space-x-3">
                                        <div
                                            class="w-10 h-10 bg-emerald-100 dark:bg-emerald-800 rounded-lg flex items-center justify-center"
                                        >
                                            <svg
                                                class="w-5 h-5 text-emerald-600 dark:text-emerald-400"
                                                fill="none"
                                                stroke="currentColor"
                                                viewBox="0 0 24 24"
                                            >
                                                <path
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    stroke-width="2"
                                                    d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"
                                                />
                                            </svg>
                                        </div>
                                        <div>
                                            <p
                                                class="text-xs text-emerald-600 dark:text-emerald-400 font-medium uppercase tracking-wide"
                                            >
                                                Available
                                            </p>
                                            <p
                                                class="text-lg font-bold text-gray-900 dark:text-white"
                                            >
                                                {formatBytes(
                                                    disk.available ||
                                                        (disk.size || 0) -
                                                            (disk.used ||
                                                                0),
                                                )}
                                            </p>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <!-- Info Footer -->
                            <div
                                class="flex items-center justify-between pt-4 mt-4 border-t border-gray-200 dark:border-border"
                            >
                                <div class="space-y-1">
                                    <!-- Device Path -->
                                    <div
                                        class="flex items-center space-x-2 text-sm"
                                    >
                                        <svg
                                            class="w-4 h-4 text-gray-400"
                                            fill="none"
                                            stroke="currentColor"
                                            viewBox="0 0 24 24"
                                        >
                                            <path
                                                stroke-linecap="round"
                                                stroke-linejoin="round"
                                                stroke-width="2"
                                                d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"
                                            />
                                        </svg>
                                        <span
                                            class="text-gray-500 dark:text-gray-400 font-mono text-xs"
                                        >
                                            {disk.device || "Unknown"}
                                        </span>
                                    </div>
                                    <!-- Mount Point -->
                                    <div
                                        class="flex items-center space-x-2 text-sm"
                                    >
                                        <svg
                                            class="w-4 h-4 text-gray-400"
                                            fill="none"
                                            stroke="currentColor"
                                            viewBox="0 0 24 24"
                                        >
                                            <path
                                                stroke-linecap="round"
                                                stroke-linejoin="round"
                                                stroke-width="2"
                                                d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"
                                            />
                                            <path
                                                stroke-linecap="round"
                                                stroke-linejoin="round"
                                                stroke-width="2"
                                                d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"
                                            />
                                        </svg>
                                        <span
                                            class="text-gray-500 dark:text-gray-400"
                                        >
                                            {disk.mountpoint || "Not mounted"}
                                        </span>
                                    </div>
                                </div>

                                <!-- Model info -->
                                {#if disk.model && disk.model !== "Unknown"}
                                    <div
                                        class="text-sm text-gray-500 dark:text-gray-400 text-right"
                                    >
                                        {disk.model}
                                    </div>
                                {/if}
                            </div>
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
                            class="bg-white dark:bg-card shadow-lg hover:shadow-xl transition-shadow duration-200 rounded-lg p-6 border border-gray-100 dark:border-border"
                        >
                            <!-- Header Section -->
                            <div class="flex items-start justify-between mb-6">
                                <div class="flex items-center space-x-4">
                                    <!-- RAID Icon with glow effect -->
                                    <div
                                        class="relative"
                                    >
                                        <div
                                            class="absolute inset-0 bg-gradient-to-br from-orange-400 to-purple-500 rounded-xl blur opacity-25"
                                        ></div>
                                        <div
                                            class="relative w-14 h-14 bg-gradient-to-br from-orange-100 to-orange-200 dark:from-orange-900/40 dark:to-purple-900/40 rounded-xl flex items-center justify-center shadow-lg"
                                        >
                                            <svg
                                                class="w-7 h-7 text-orange-600 dark:text-orange-400"
                                                fill="currentColor"
                                                viewBox="0 0 20 20"
                                            >
                                                <path
                                                    d="M5 3a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2V5a2 2 0 00-2-2H5zM5 11a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2H5zM11 5a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V5zM13 11a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2h-2z"
                                                />
                                            </svg>
                                        </div>
                                    </div>

                                    <div>
                                        <h3
                                            class="text-xl font-bold text-gray-900 dark:text-white"
                                        >
                                            {array.name}
                                        </h3>
                                        <div
                                            class="flex items-center space-x-2 mt-1"
                                        >
                                            <span
                                                class="px-2 py-0.5 text-xs font-semibold rounded-full bg-gradient-to-r from-orange-500 to-purple-500 text-white"
                                            >
                                                {array.level.toUpperCase()}
                                            </span>
                                            <span
                                                class="text-sm text-gray-500 dark:text-gray-400"
                                            >
                                                {array.devices?.length || 0} device{(array.devices?.length || 0) !== 1 ? 's' : ''}
                                            </span>
                                        </div>
                                    </div>
                                </div>

                                <!-- Status Badge -->
                                <div
                                    class="flex items-center space-x-2 px-3 py-1.5 rounded-full {array.state === 'clean' || array.state === 'active'
                                        ? 'bg-green-100 dark:bg-green-900/30'
                                        : array.state === 'degraded'
                                        ? 'bg-yellow-100 dark:bg-yellow-900/30'
                                        : 'bg-red-100 dark:bg-red-900/30'}"
                                >
                                    <div
                                        class="w-2.5 h-2.5 rounded-full {array.state === 'clean' || array.state === 'active'
                                            ? 'bg-green-500 animate-pulse'
                                            : array.state === 'degraded'
                                            ? 'bg-yellow-500'
                                            : 'bg-red-500'}"
                                    ></div>
                                    <span
                                        class="text-sm font-medium {array.state === 'clean' || array.state === 'active'
                                            ? 'text-green-700 dark:text-green-400'
                                            : array.state === 'degraded'
                                            ? 'text-yellow-700 dark:text-yellow-400'
                                            : 'text-red-700 dark:text-red-400'}"
                                    >
                                        {array.state || "Unknown"}
                                    </span>
                                </div>
                            </div>

                            <!-- Stats Grid -->
                            <div
                                class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6"
                            >
                                <!-- Size Card -->
                                <div
                                    class="bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 rounded-xl p-4 border border-blue-100 dark:border-blue-800"
                                >
                                    <div class="flex items-center space-x-3">
                                        <div
                                            class="w-10 h-10 bg-blue-100 dark:bg-blue-800 rounded-lg flex items-center justify-center"
                                        >
                                            <svg
                                                class="w-5 h-5 text-blue-600 dark:text-blue-400"
                                                fill="none"
                                                stroke="currentColor"
                                                viewBox="0 0 24 24"
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
                                            <p
                                                class="text-xs text-blue-600 dark:text-blue-400 font-medium uppercase tracking-wide"
                                            >
                                                Total Size
                                            </p>
                                            <p
                                                class="text-lg font-bold text-gray-900 dark:text-white"
                                            >
                                                {formatBytes(array.size)}
                                            </p>
                                        </div>
                                    </div>
                                </div>

                                <!-- Used Card -->
                                <div
                                    class="bg-gradient-to-br from-purple-50 to-pink-50 dark:from-purple-900/20 dark:to-pink-900/20 rounded-xl p-4 border border-purple-100 dark:border-purple-800"
                                >
                                    <div class="flex items-center space-x-3">
                                        <div
                                            class="w-10 h-10 bg-purple-100 dark:bg-purple-800 rounded-lg flex items-center justify-center"
                                        >
                                            <svg
                                                class="w-5 h-5 text-purple-600 dark:text-purple-400"
                                                fill="none"
                                                stroke="currentColor"
                                                viewBox="0 0 24 24"
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
                                            <p
                                                class="text-xs text-purple-600 dark:text-purple-400 font-medium uppercase tracking-wide"
                                            >
                                                Used Space
                                            </p>
                                            <p
                                                class="text-lg font-bold text-gray-900 dark:text-white"
                                            >
                                                {array.used
                                                    ? formatBytes(
                                                          array.used,
                                                      )
                                                    : "N/A"}
                                            </p>
                                        </div>
                                    </div>
                                    <!-- Usage Bar -->
                                    {#if array.used && array.size}
                                        <div
                                            class="mt-3 w-full bg-gray-200 dark:bg-muted rounded-full h-2 overflow-hidden"
                                        >
                                            <div
                                                class="h-full bg-gradient-to-r from-purple-500 to-pink-500 rounded-full transition-all duration-300"
                                                style="width: {Math.min(
                                                    (array.used / array.size) * 100,
                                                    100,
                                                )}%"
                                            ></div>
                                        </div>
                                    {/if}
                                </div>

                                <!-- Health Card -->
                                <div
                                    class="bg-gradient-to-br from-emerald-50 to-teal-50 dark:from-emerald-900/20 dark:to-teal-900/20 rounded-xl p-4 border border-emerald-100 dark:border-emerald-800"
                                >
                                    <div class="flex items-center space-x-3">
                                        <div
                                            class="w-10 h-10 bg-emerald-100 dark:bg-emerald-800 rounded-lg flex items-center justify-center"
                                        >
                                            <svg
                                                class="w-5 h-5 text-emerald-600 dark:text-emerald-400"
                                                fill="none"
                                                stroke="currentColor"
                                                viewBox="0 0 24 24"
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
                                            <p
                                                class="text-xs text-emerald-600 dark:text-emerald-400 font-medium uppercase tracking-wide"
                                            >
                                                Health
                                            </p>
                                            <p
                                                class="text-lg font-bold text-gray-900 dark:text-white"
                                            >
                                                {array.health}%
                                            </p>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <!-- Devices List -->
                            <div>
                                <div
                                    class="flex items-center justify-between mb-3"
                                >
                                    <h4
                                        class="text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide"
                                    >
                                        Member Devices
                                    </h4>
                                    <span
                                        class="text-xs text-gray-500 dark:text-gray-400"
                                    >
                                        {array.devices?.length || 0} disk{(array.devices?.length || 0) !== 1 ? 's' : ''}
                                    </span>
                                </div>
                                <div
                                    class="flex flex-wrap gap-2"
                                >
                                    {#each array.devices || [] as device}
                                        <div
                                            class="inline-flex items-center space-x-2 px-3 py-2 bg-gray-100 dark:bg-card rounded-lg border border-gray-200 dark:border-border"
                                        >
                                            <svg
                                                class="w-4 h-4 text-gray-500 dark:text-gray-400"
                                                fill="none"
                                                stroke="currentColor"
                                                viewBox="0 0 24 24"
                                            >
                                                <path
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    stroke-width="2"
                                                    d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"
                                                />
                                            </svg>
                                            <span
                                                class="text-sm font-medium text-gray-700 dark:text-gray-300"
                                            >
                                                {device}
                                            </span>
                                        </div>
                                    {/each}
                                </div>
                            </div>

                            <!-- Actions Footer -->
                            <div
                                class="flex items-center justify-between pt-4 mt-4 border-t border-gray-200 dark:border-border"
                            >
                                <div
                                    class="flex items-center space-x-2 text-sm text-gray-500 dark:text-gray-400"
                                >
                                    <svg
                                        class="w-4 h-4"
                                        fill="none"
                                        stroke="currentColor"
                                        viewBox="0 0 24 24"
                                    >
                                        <path
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            stroke-width="2"
                                            d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"
                                        />
                                        <path
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            stroke-width="2"
                                            d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"
                                        />
                                    </svg>
                                    <span>{array.mount_point || "Not mounted"}</span>
                                </div>
                                <button
                                    on:click={() => handleRAIDDelete(array)}
                                    class="inline-flex items-center space-x-2 px-4 py-2 text-sm font-medium text-red-600 bg-red-50 hover:bg-red-100 dark:bg-red-900/20 dark:hover:bg-red-900/40 dark:text-red-400 rounded-lg transition-colors"
                                    title="Delete RAID Array"
                                >
                                    <svg
                                        class="w-4 h-4"
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
                                    <span>Delete Array</span>
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
                            class="bg-white dark:bg-card shadow-lg hover:shadow-xl transition-shadow duration-200 rounded-lg p-6 border border-gray-100 dark:border-border"
                        >
                            <!-- Header Section -->
                            <div class="flex items-start justify-between mb-6">
                                <div class="flex items-center space-x-4">
                                    <!-- Pool Icon with glow effect -->
                                    <div
                                        class="relative"
                                    >
                                        <div
                                            class="absolute inset-0 bg-gradient-to-br from-blue-400 to-cyan-500 rounded-xl blur opacity-25"
                                        ></div>
                                        <div
                                            class="relative w-14 h-14 bg-gradient-to-br from-blue-100 to-cyan-100 dark:from-blue-900/40 dark:to-cyan-900/40 rounded-xl flex items-center justify-center shadow-lg"
                                        >
                                            <svg
                                                class="w-7 h-7 text-blue-600 dark:text-blue-400"
                                                fill="currentColor"
                                                viewBox="0 0 20 20"
                                            >
                                                <path
                                                    d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zM2 11a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z"
                                                />
                                            </svg>
                                        </div>
                                    </div>

                                    <div>
                                        <h3
                                            class="text-xl font-bold text-gray-900 dark:text-white"
                                        >
                                            {pool.name}
                                        </h3>
                                        <div
                                            class="flex items-center space-x-2 mt-1"
                                        >
                                            <span
                                                class="px-2 py-0.5 text-xs font-semibold rounded-full bg-gradient-to-r from-blue-500 to-cyan-500 text-white"
                                            >
                                                STORAGE POOL
                                            </span>
                                            <span
                                                class="text-sm text-gray-500 dark:text-gray-400"
                                            >
                                                {pool.devices?.length || 0} volume{(pool.devices?.length || 0) !== 1 ? 's' : ''}
                                            </span>
                                        </div>
                                    </div>
                                </div>

                                <!-- Status Badge -->
                                <div
                                    class="flex items-center space-x-2 px-3 py-1.5 rounded-full {pool.state === 'active'
                                        ? 'bg-green-100 dark:bg-green-900/30'
                                        : 'bg-gray-100 dark:bg-muted'}"
                                >
                                    <div
                                        class="w-2.5 h-2.5 rounded-full {pool.state === 'active'
                                            ? 'bg-green-500 animate-pulse'
                                            : 'bg-gray-400'}"
                                    ></div>
                                    <span
                                        class="text-sm font-medium {pool.state === 'active'
                                            ? 'text-green-700 dark:text-green-400'
                                            : 'text-gray-600 dark:text-gray-400'}"
                                    >
                                        {pool.state || "Inactive"}
                                    </span>
                                </div>
                            </div>

                            <!-- Usage Mode -->
                            <div class="mb-4">
                                <div class="flex items-center justify-between">
                                    <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Usage</span>
                                    <select
                                        value={pool.export_mode || "file"}
                                        on:change={(e) => setPoolUsage(pool, e.target.value)}
                                        class="text-sm border border-gray-300 dark:border rounded-md px-3 py-1.5 bg-white dark:bg-muted dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                    >
                                        <option value="file">📁 File Share</option>
                                        <option value="iscsi">💾 iSCSI Block Device</option>
                                        <option value="available">⏸ Available</option>
                                    </select>
                                </div>
                                <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                                    {pool.export_mode === 'file' ? 'Mounted for NFS/Samba sharing' : ''}
                                    {pool.export_mode === 'iscsi' ? 'Unmounted, for iSCSI/LVM use' : ''}
                                    {pool.export_mode === 'available' ? 'Unmounted, reserved' : ''}
                                </p>
                            </div>

                            <!-- Stats Grid -->
                            <div
                                class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6"
                            >
                                <!-- Total Size Card -->
                                <div
                                    class="bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 rounded-xl p-4 border border-blue-100 dark:border-blue-800"
                                >
                                    <div class="flex items-center space-x-3">
                                        <div
                                            class="w-10 h-10 bg-blue-100 dark:bg-blue-800 rounded-lg flex items-center justify-center"
                                        >
                                            <svg
                                                class="w-5 h-5 text-blue-600 dark:text-blue-400"
                                                fill="none"
                                                stroke="currentColor"
                                                viewBox="0 0 24 24"
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
                                            <p
                                                class="text-xs text-blue-600 dark:text-blue-400 font-medium uppercase tracking-wide"
                                            >
                                                Total Size
                                            </p>
                                            <p
                                                class="text-lg font-bold text-gray-900 dark:text-white"
                                            >
                                                {formatBytes(pool.size)}
                                            </p>
                                        </div>
                                    </div>
                                </div>

                                <!-- Used Card -->
                                <div
                                    class="bg-gradient-to-br from-purple-50 to-pink-50 dark:from-purple-900/20 dark:to-pink-900/20 rounded-xl p-4 border border-purple-100 dark:border-purple-800"
                                >
                                    <div class="flex items-center space-x-3">
                                        <div
                                            class="w-10 h-10 bg-purple-100 dark:bg-purple-800 rounded-lg flex items-center justify-center"
                                        >
                                            <svg
                                                class="w-5 h-5 text-purple-600 dark:text-purple-400"
                                                fill="none"
                                                stroke="currentColor"
                                                viewBox="0 0 24 24"
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
                                            <p
                                                class="text-xs text-purple-600 dark:text-purple-400 font-medium uppercase tracking-wide"
                                            >
                                                Used Space
                                            </p>
                                            <p
                                                class="text-lg font-bold text-gray-900 dark:text-white"
                                            >
                                                {formatBytes(pool.used)}
                                            </p>
                                        </div>
                                    </div>
                                    <!-- Usage Bar -->
                                    {#if pool.size > 0}
                                        <div
                                            class="mt-3 w-full bg-gray-200 dark:bg-muted rounded-full h-2 overflow-hidden"
                                        >
                                            <div
                                                class="h-full bg-gradient-to-r from-purple-500 to-pink-500 rounded-full transition-all duration-300"
                                                style="width: {Math.min(
                                                    (pool.used / pool.size) * 100,
                                                    100,
                                                )}%"
                                            ></div>
                                        </div>
                                    {/if}
                                </div>

                                <!-- Available Card -->
                                <div
                                    class="bg-gradient-to-br from-emerald-50 to-teal-50 dark:from-emerald-900/20 dark:to-teal-900/20 rounded-xl p-4 border border-emerald-100 dark:border-emerald-800"
                                >
                                    <div class="flex items-center space-x-3">
                                        <div
                                            class="w-10 h-10 bg-emerald-100 dark:bg-emerald-800 rounded-lg flex items-center justify-center"
                                        >
                                            <svg
                                                class="w-5 h-5 text-emerald-600 dark:text-emerald-400"
                                                fill="none"
                                                stroke="currentColor"
                                                viewBox="0 0 24 24"
                                            >
                                                <path
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    stroke-width="2"
                                                    d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"
                                                />
                                            </svg>
                                        </div>
                                        <div>
                                            <p
                                                class="text-xs text-emerald-600 dark:text-emerald-400 font-medium uppercase tracking-wide"
                                            >
                                                Available
                                            </p>
                                            <p
                                                class="text-lg font-bold text-gray-900 dark:text-white"
                                            >
                                                {formatBytes(
                                                    pool.available,
                                                )}
                                            </p>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <!-- Devices List -->
                            {#if pool.devices && pool.devices.length > 0}
                                <div class="mb-6">
                                    <div
                                        class="flex items-center justify-between mb-3"
                                    >
                                        <h4
                                            class="text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide"
                                        >
                                            Volumes
                                        </h4>
                                        <span
                                            class="text-xs text-gray-500 dark:text-gray-400"
                                        >
                                            {pool.devices?.length || 0} volume{(pool.devices?.length || 0) !== 1 ? 's' : ''}
                                        </span>
                                    </div>
                                    <div
                                        class="flex flex-wrap gap-2"
                                    >
                                        {#each pool.devices || [] as device}
                                            <div
                                                class="inline-flex items-center space-x-2 px-3 py-2 bg-gray-100 dark:bg-card rounded-lg border border-gray-200 dark:border-border"
                                            >
                                                <svg
                                                    class="w-4 h-4 text-gray-500 dark:text-gray-400"
                                                    fill="none"
                                                    stroke="currentColor"
                                                    viewBox="0 0 24 24"
                                                >
                                                    <path
                                                        stroke-linecap="round"
                                                        stroke-linejoin="round"
                                                        stroke-width="2"
                                                        d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"
                                                    />
                                                </svg>
                                                <span
                                                    class="text-sm font-medium text-gray-700 dark:text-gray-300"
                                                >
                                                    {device}
                                                </span>
                                            </div>
                                        {/each}
                                    </div>
                                </div>
                            {/if}

                            <!-- Actions Footer -->
                            <div
                                class="flex items-center justify-between pt-4 mt-4 border-t border-gray-200 dark:border-border"
                            >
                                <div
                                    class="flex items-center space-x-2 text-sm text-gray-500 dark:text-gray-400"
                                >
                                    <svg
                                        class="w-4 h-4"
                                        fill="none"
                                        stroke="currentColor"
                                        viewBox="0 0 24 24"
                                    >
                                        <path
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            stroke-width="2"
                                            d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"
                                        />
                                        <path
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            stroke-width="2"
                                            d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"
                                        />
                                    </svg>
                                    <span>{pool.mount_point || "Not mounted"}</span>
                                </div>
                                <div class="flex items-center space-x-2">
                                    <button
                                        on:click={() => {
                                            handlePoolAction(
                                                "mount",
                                                pool,
                                            );
                                        }}
                                        class="inline-flex items-center space-x-2 px-4 py-2 text-sm font-medium text-blue-600 bg-blue-50 hover:bg-blue-100 dark:bg-blue-900/20 dark:hover:bg-blue-900/40 dark:text-blue-400 rounded-lg transition-colors"
                                    >
                                        <svg
                                            class="w-4 h-4"
                                            fill="none"
                                            stroke="currentColor"
                                            viewBox="0 0 24 24"
                                        >
                                            <path
                                                stroke-linecap="round"
                                                stroke-linejoin="round"
                                                stroke-width="2"
                                                d="M5 13l4 4L19 7"
                                            />
                                        </svg>
                                        <span>{pool.state === "active"
                                            ? "Unmount"
                                            : "Mount"}</span>
                                    </button>
                                    <button
                                        on:click={() =>
                                            handlePoolAction(
                                                "edit",
                                                pool,
                                            )}
                                        class="inline-flex items-center space-x-2 px-4 py-2 text-sm font-medium text-gray-600 bg-gray-50 hover:bg-gray-100 dark:bg-gray-900/20 dark:hover:bg-gray-900/40 dark:text-gray-400 rounded-lg transition-colors"
                                    >
                                        <svg
                                            class="w-4 h-4"
                                            fill="none"
                                            stroke="currentColor"
                                            viewBox="0 0 24 24"
                                        >
                                            <path
                                                stroke-linecap="round"
                                                stroke-linejoin="round"
                                                stroke-width="2"
                                                d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                                            />
                                        </svg>
                                        <span>Edit</span>
                                    </button>
                                    <button
                                        on:click={() =>
                                            handlePoolAction(
                                                "delete",
                                                pool,
                                            )}
                                        class="inline-flex items-center space-x-2 px-4 py-2 text-sm font-medium text-red-600 bg-red-50 hover:bg-red-100 dark:bg-red-900/20 dark:hover:bg-red-900/40 dark:text-red-400 rounded-lg transition-colors"
                                    >
                                        <svg
                                            class="w-4 h-4"
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
                                        <span>Delete</span>
                                    </button>
                                </div>
                            </div>
                        </div>
                    {/each}
                </div>
            {/if}
        {:else if activeTab === "vgs"}
            <!-- Volume Groups Tab -->
            {#if vgsLoading}
                <div class="flex justify-center items-center h-64">
                    <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                </div>
            {:else if volumeGroups.length === 0}
                <div class="text-center py-12 bg-white dark:bg-card shadow-lg hover:shadow-xl transition-shadow duration-200 rounded-lg border border-gray-100 dark:border-border">
                    <svg class="h-16 w-16 text-gray-400 mx-auto mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                    </svg>
                    <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">No volume groups configured</h3>
                    <p class="text-gray-600 dark:text-gray-300 mb-4">
                        Create volume groups to manage flexible iSCSI LUNs
                    </p>
                    <button
                        on:click={async () => {
                            await loadAvailableVGDevices();
                            showCreateVGModal = true;
                        }}
                        class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                    >
                        Create Volume Group
                    </button>
                </div>
            {:else}
                <div class="space-y-4">
                    {#each volumeGroups as vg (vg.name)}
                        <div class="bg-white dark:bg-card shadow-lg hover:shadow-xl transition-shadow duration-200 rounded-lg border border-gray-100 dark:border-border">
                            <div class="flex items-start justify-between">
                                <div class="flex items-center space-x-4">
                                    <!-- VG Icon -->
                                    <div class="relative">
                                        <div class="absolute inset-0 bg-gradient-to-br from-purple-400 to-pink-500 rounded-xl blur opacity-25"></div>
                                        <div class="relative w-14 h-14 bg-gradient-to-br from-purple-100 to-pink-100 dark:from-purple-900/40 dark:to-pink-900/40 rounded-xl flex items-center justify-center shadow-lg">
                                            <svg class="w-7 h-7 text-purple-600 dark:text-purple-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
                                            </svg>
                                        </div>
                                    </div>

                                    <!-- VG Info -->
                                    <div>
                                        <div class="flex items-center space-x-2">
                                            <h3 class="text-xl font-bold text-gray-900 dark:text-white">
                                                {vg.name}
                                            </h3>
                                            <span class="px-2 py-0.5 text-xs font-semibold rounded-full bg-purple-100 dark:bg-purple-900/40 text-purple-700 dark:text-purple-400">
                                                LVM
                                            </span>
                                        </div>
                                        {#if vg.devices && vg.devices.length > 0}
                                            <p class="mt-1 text-xs font-mono text-gray-500 dark:text-gray-400">
                                                {vg.devices.join(', ')}
                                            </p>
                                        {/if}
                                    </div>
                                </div>

                                <!-- Actions -->
                                <div class="flex items-center space-x-2">
                                    <button
                                        on:click={() => {
                                            vgToDelete = vg;
                                            showVGDeleteModal = true;
                                        }}
                                        class="p-2 rounded-lg text-gray-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20"
                                        title="Delete Volume Group"
                                    >
                                        <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                        </svg>
                                    </button>
                                </div>
                            </div>

                            <!-- Stats Grid -->
                            <div class="bg-gray-50 dark:bg-muted rounded-b-lg p-4 border-t border-gray-200 dark:border-border">
                                <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                                    <!-- Used Space -->
                                    <div class="bg-gradient-to-br from-purple-50 to-pink-50 dark:from-purple-900/20 dark:to-pink-900/20 rounded-xl p-4 border border-purple-100 dark:border-purple-800">
                                        <div class="flex items-center space-x-3">
                                            <div class="w-10 h-10 bg-purple-100 dark:bg-purple-800 rounded-lg flex items-center justify-center">
                                                <svg class="w-5 h-5 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
                                                </svg>
                                            </div>
                                            <div>
                                                <p class="text-xs text-purple-600 dark:text-purple-400 font-medium uppercase tracking-wide">
                                                    Used
                                                </p>
                                                <p class="text-lg font-bold text-gray-900 dark:text-white">
                                                    {formatBytes(vg.size - vg.free)}
                                                </p>
                                            </div>
                                        </div>
                                        <!-- Usage Bar -->
                                        {#if vg.size > 0}
                                            <div class="mt-3 w-full bg-gray-200 dark:bg-muted rounded-full h-2 overflow-hidden">
                                                <div
                                                    class="h-full bg-gradient-to-r from-purple-500 to-pink-500 rounded-full transition-all duration-300"
                                                    style="width: {((vg.size - vg.free) / vg.size * 100).toFixed(1)}%"
                                                ></div>
                                            </div>
                                        {/if}
                                    </div>

                                    <!-- Available Space -->
                                    <div class="bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 rounded-xl p-4 border border-blue-100 dark:border-blue-800">
                                        <div class="flex items-center space-x-3">
                                            <div class="w-10 h-10 bg-blue-100 dark:bg-blue-800 rounded-lg flex items-center justify-center">
                                                <svg class="w-5 h-5 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5h4M4 7h16"/>
                                                </svg>
                                            </div>
                                            <div>
                                                <p class="text-xs text-blue-600 dark:text-blue-400 font-medium uppercase tracking-wide">
                                                    Available
                                                </p>
                                                <p class="text-lg font-bold text-gray-900 dark:text-white">
                                                    {formatBytes(vg.free)}
                                                </p>
                                            </div>
                                        </div>
                                    </div>

                                    <!-- LUNs -->
                                    <div class="bg-gradient-to-br from-emerald-50 to-teal-50 dark:from-emerald-900/20 dark:to-teal-900/20 rounded-xl p-4 border border-emerald-100 dark:border-emerald-800">
                                        <div class="flex items-center space-x-3">
                                            <div class="w-10 h-10 bg-emerald-100 dark:bg-emerald-800 rounded-lg flex items-center justify-center">
                                                <svg class="w-5 h-5 text-emerald-600 dark:text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
                                                </svg>
                                            </div>
                                            <div>
                                                <p class="text-xs text-emerald-600 dark:text-emerald-400 font-medium uppercase tracking-wide">
                                                    Total LUNs
                                                </p>
                                                <p class="text-lg font-bold text-gray-900 dark:text-white">
                                                    {vg.lun_count}
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
        {/if}

        <!-- Delete Confirmation Modal -->
        {#if showDeleteModal}
            <div
                class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
            >
                <div
                    class="bg-white dark:bg-card rounded-lg p-6 max-w-md w-full mx-4"
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
                            class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-muted dark:text-white sm:text-sm"
                            placeholder="DELETE POOL_NAME"
                        />
                    </div>
                    <div class="flex justify-end space-x-3">
                        <button
                            type="button"
                            class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-muted border border-gray-300 dark:border rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-muted"
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
                    class="bg-white dark:bg-card rounded-lg p-6 max-w-2xl w-full mx-4"
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
                                    class="bg-gray-100 dark:bg-muted p-3 rounded text-xs overflow-auto">{JSON.stringify(
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
                        class="relative bg-white dark:bg-card rounded-lg max-w-2xl w-full p-6"
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
                                    class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-muted dark:text-white"
                                    required
                                    disabled={poolCreating}
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
                                    class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-muted dark:text-white"
                                    disabled={poolCreating}
                                >
                                    <option value="mergerfs">Combined Storage</option>
                                </select>
                            </div>

                            <fieldset class="border-0 p-0 m-0" disabled={poolCreating}>
                                <legend class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                    Select Disks
                                </legend>
                                {#if availableDevices.length === 0}
                                    <div class="w-full p-4 border border-gray-300 dark:border rounded-md bg-gray-50 dark:bg-muted text-gray-500 dark:text-gray-400 text-center">
                                        No disks available. All devices are either mounted or part of a RAID array.
                                    </div>
                                {:else}
                                    <div class="border border-gray-300 dark:border rounded-md p-4 max-h-60 overflow-y-auto bg-gray-50 dark:bg-muted">
                                        {#each availableDevices as disk}
                                                <label class="flex items-center space-x-3 p-2 hover:bg-gray-100 dark:hover:bg-muted rounded cursor-pointer">
                                                    <input
                                                        type="checkbox"
                                                        checked={newPool.devices.includes(disk.device)}
                                                        on:change={() => toggleDevice(disk.device)}
                                                        class="rounded border-gray-300 dark:border text-blue-600 focus:ring-blue-500 dark:bg-card"
                                                        disabled={poolCreating}
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
                                    class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-muted dark:text-white"
                                    disabled={poolCreating}
                                />
                            </div>

                            <div class="flex justify-end space-x-3 mt-6">
                                <button
                                    type="button"
                                    on:click={() => {
                                        showCreatePoolModal = false;
                                        resetPoolForm();
                                    }}
                                    class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-muted border border-gray-300 dark:border rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-muted disabled:opacity-50 disabled:cursor-not-allowed"
                                    disabled={poolCreating}
                                >
                                    Cancel
                                </button>
                                <button
                                    type="submit"
                                    class="px-4 py-2 bg-blue-600 text-white rounded-md shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed flex items-center space-x-2"
                                    disabled={poolCreating}
                                >
                                    {#if poolCreating}
                                        <svg
                                            class="animate-spin h-4 w-4 text-white"
                                            fill="none"
                                            viewBox="0 0 24 24"
                                        >
                                            <circle
                                                class="opacity-25"
                                                cx="12"
                                                cy="12"
                                                r="10"
                                                stroke="currentColor"
                                                stroke-width="4"
                                            ></circle>
                                            <path
                                                class="opacity-75"
                                                fill="currentColor"
                                                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                                            ></path>
                                        </svg>
                                        <span>Creating...</span>
                                    {:else}
                                        <span>Create Pool</span>
                                    {/if}
                                </button>
                            </div>
                        </form>

                        <!-- Loading overlay -->
                        {#if poolCreating}
                            <div class="absolute inset-0 bg-white dark:bg-card bg-opacity-90 rounded-lg flex flex-col items-center justify-center z-10">
                                <div class="flex flex-col items-center space-y-4">
                                    <div
                                        class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"
                                    ></div>
                                    <div class="text-center space-y-2">
                                        <p class="text-lg font-medium text-gray-900 dark:text-white">
                                            Creating Storage Pool
                                        </p>
                                        <p class="text-sm text-gray-500 dark:text-gray-400 max-w-xs">
                                            This may take a while depending on disk sizes. Formatting and mounting disks...
                                        </p>
                                    </div>
                                </div>
                            </div>
                        {/if}
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
                    class="bg-white dark:bg-card rounded-lg p-6 max-w-md w-full mx-4"
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
                            class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-muted dark:text-white sm:text-sm"
                            placeholder="DELETE {raidToDelete?.name.toUpperCase()}"
                        />
                    </div>
                    <div class="flex justify-end space-x-3">
                        <button
                            type="button"
                            class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-muted border border-gray-300 dark:border rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-muted"
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

        <!-- Edit Storage Pool Modal -->
        {#if showEditPoolModal}
            <div class="fixed inset-0 z-50 overflow-y-auto">
                <div class="flex items-center justify-center min-h-screen px-4">
                    <div
                        class="fixed inset-0 bg-gray-500 bg-opacity-75"
                        role="button"
                        tabindex="0"
                        on:click={cancelEditPool}
                        on:keydown={(e) => e.key === "Escape" && cancelEditPool()}
                        aria-label="Close modal"
                    ></div>
                    <div
                        class="relative bg-white dark:bg-card rounded-lg max-w-lg w-full p-6"
                        role="dialog"
                        aria-modal="true"
                        aria-labelledby="edit-pool-title"
                        tabindex="-1"
                        on:click|stopPropagation
                        on:keydown|stopPropagation
                    >
                        <h3
                            id="edit-pool-title"
                            class="text-lg font-medium text-gray-900 dark:text-white mb-4"
                        >
                            Edit Storage Pool
                        </h3>
                        <form on:submit|preventDefault={saveEditedPool} class="space-y-4">
                            <div>
                                <label
                                    for="editPoolName"
                                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                                >
                                    Pool Name
                                </label>
                                <input
                                    id="editPoolName"
                                    type="text"
                                    bind:value={editedPool.name}
                                    placeholder="my-pool"
                                    class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-muted dark:text-white"
                                    required
                                    disabled={poolEditing}
                                />
                            </div>

                            <div>
                                <label
                                    for="editPoolConfig"
                                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                                >
                                    Mount Options (optional)
                                </label>
                                <input
                                    id="editPoolConfig"
                                    type="text"
                                    bind:value={editedPool.config}
                                    placeholder="defaults,allow_other,use_ino"
                                    class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-muted dark:text-white"
                                    disabled={poolEditing}
                                />
                                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                                    Note: Changing mount options requires remounting the pool.
                                </p>
                            </div>

                            <div class="flex justify-end space-x-3 mt-6">
                                <button
                                    type="button"
                                    on:click={cancelEditPool}
                                    class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-muted border border-gray-300 dark:border rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-muted disabled:opacity-50 disabled:cursor-not-allowed"
                                    disabled={poolEditing}
                                >
                                    Cancel
                                </button>
                                <button
                                    type="submit"
                                    class="px-4 py-2 bg-blue-600 text-white rounded-md shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed flex items-center space-x-2"
                                    disabled={poolEditing}
                                >
                                    {#if poolEditing}
                                        <svg
                                            class="animate-spin h-4 w-4 text-white"
                                            fill="none"
                                            viewBox="0 0 24 24"
                                        >
                                            <circle
                                                class="opacity-25"
                                                cx="12"
                                                cy="12"
                                                r="10"
                                                stroke="currentColor"
                                                stroke-width="4"
                                            ></circle>
                                            <path
                                                class="opacity-75"
                                                fill="currentColor"
                                                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                                            ></path>
                                        </svg>
                                        <span>Saving...</span>
                                    {:else}
                                        <span>Save Changes</span>
                                    {/if}
                                </button>
                            </div>
                        </form>
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
                on:keydown={(e) => e.key === "Escape" && (showCreateRAIDModal = false)}
                role="dialog"
                aria-modal="true"
                aria-labelledby="create-raid-title"
                tabindex="-1"
            >
                <div class="flex items-center justify-center min-h-screen px-4">
                    <div
                        class="fixed inset-0 bg-gray-500 bg-opacity-75"
                        aria-hidden="true"
                        on:click={() => {
                            showCreateRAIDModal = false;
                            resetRAIDForm();
                        }}
                    ></div>
                    <div
                        class="relative bg-white dark:bg-card rounded-lg max-w-2xl w-full p-6"
                        on:click|stopPropagation
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
                                    Array Name (optional - auto-generated if empty)
                                </label>
                                <input
                                    id="raidName"
                                    type="text"
                                    bind:value={newRAID.name}
                                    placeholder="Leave empty to auto-generate (e.g., md0, md1)"
                                    class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-muted dark:text-white"
                                />
                                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                                    If not specified, the system will automatically assign the next available device number.
                                </p>
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
                                    class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-muted dark:text-white"
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
                                    <div class="w-full p-4 border border-gray-300 dark:border rounded-md bg-gray-50 dark:bg-muted text-gray-500 dark:text-gray-400 text-center">
                                        No disks available. Please add disks to the system first.
                                    </div>
                                {:else}
                                    <div class="border border-gray-300 dark:border rounded-md p-4 max-h-60 overflow-y-auto bg-gray-50 dark:bg-muted">
                                        {#each diskStats as disk}
                                            {#if disk.device && disk.device !== "" && !disk.device.includes("/dev/md")}
                                                <label class="flex items-center space-x-3 p-2 hover:bg-gray-100 dark:hover:bg-muted rounded cursor-pointer">
                                                    <input
                                                        type="checkbox"
                                                        checked={newRAID.devices.includes(disk.device)}
                                                        on:change={() => toggleRAIDDevice(disk.device)}
                                                        class="rounded border-gray-300 dark:border text-blue-600 focus:ring-blue-500 dark:bg-card"
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
                                    class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-muted border border-gray-300 dark:border rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-muted"
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

        <!-- Create Volume Group Modal -->
        {#if showCreateVGModal}
            <div
                class="fixed inset-0 z-50 overflow-y-auto"
                on:click={() => showCreateVGModal = false}
                on:keydown={(e) => e.key === 'Escape' && (showCreateVGModal = false)}
                role="dialog"
                aria-modal="true"
                aria-labelledby="create-vg-title-storage"
                tabindex="-1"
            >
                <div class="flex items-center justify-center min-h-screen px-4">
                    <div
                        class="fixed inset-0 bg-gray-500 bg-opacity-75"
                        aria-hidden="true"
                        on:click={() => showCreateVGModal = false}
                    ></div>
                    <div
                        class="relative bg-white dark:bg-card rounded-lg max-w-lg w-full p-6"
                        on:click|stopPropagation
                    >
                        <h3 id="create-vg-title-storage" class="text-lg font-medium text-gray-900 dark:text-white mb-4">
                            Create Volume Group
                        </h3>

                        <form on:submit|preventDefault={createVG} class="space-y-4">
                            <div>
                                <label for="vg-name-storage" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                    Volume Group Name
                                </label>
                                <input
                                    id="vg-name-storage"
                                    type="text"
                                    bind:value={newVG.name}
                                    class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md dark:bg-muted dark:text-white"
                                    placeholder="e.g., vg-iscsi"
                                    pattern="[a-z0-9-]+"
                                    title="Only lowercase letters, numbers, and hyphens"
                                    required
                                />
                            </div>

                            <fieldset class="border-0 p-0 m-0">
                                <legend class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                    Physical Devices
                                </legend>
                                <div class="space-y-2 max-h-48 overflow-y-auto border border-gray-300 dark:border rounded-md p-2 dark:bg-muted">
                                    {#each availableVGDevices as device (device.path)}
                                        <label class="flex items-center space-x-2 p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded cursor-pointer">
                                            <input
                                                type="checkbox"
                                                bind:group={newVG.devices}
                                                value={device.path}
                                                disabled={!device.available}
                                                class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                                            />
                                            <div class="flex-1">
                                                <div class="text-sm font-medium text-gray-900 dark:text-white">{device.path}</div>
                                                {#if device.available}
                                                    <div class="text-xs text-green-600 dark:text-green-400">Available</div>
                                                {:else}
                                                    <div class="text-xs text-red-600 dark:text-red-400">{device.reason}</div>
                                                {/if}
                                            </div>
                                        </label>
                                    {/each}
                                </div>
                                <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                                    Select unmounted devices to include in the volume group
                                </p>
                            </fieldset>

                            <div class="flex justify-end space-x-3 pt-4">
                                <button
                                    type="button"
                                    on:click={() => showCreateVGModal = false}
                                    class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-card border border-gray-300 dark:border rounded-md hover:bg-gray-50 dark:hover:bg-muted"
                                >
                                    Cancel
                                </button>
                                <button
                                    type="submit"
                                    disabled={vgCreating || newVG.devices.length === 0}
                                    class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
                                >
                                    {#if vgCreating}
                                        <svg class="animate-spin h-4 w-4 mr-2 inline" fill="none" viewBox="0 0 24 24">
                                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                                            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8 0V0C5.373 0 0 1.4 0z"></path>
                                        </svg>
                                        Creating...
                                    {:else}
                                        Create Volume Group
                                    {/if}
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        {/if}

        <!-- Delete VG Modal -->
        {#if showVGDeleteModal && vgToDelete}
            <div
                class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
                on:click={() => {
                    showVGDeleteModal = false;
                    vgToDelete = null;
                }}
                on:keydown={(e) => e.key === 'Escape' && (showVGDeleteModal = false)}
                role="dialog"
                aria-modal="true"
                aria-labelledby="delete-vg-title"
                tabindex="-1"
            >
                <div
                    class="bg-white dark:bg-card rounded-lg p-6 max-w-md w-full mx-4"
                >
                    <h3 id="delete-vg-title" class="text-lg font-medium text-gray-900 dark:text-white mb-4">
                        Delete Volume Group
                    </h3>
                    <p class="text-sm text-gray-600 dark:text-gray-300 mb-4">
                        Are you sure you want to delete <span class="font-bold text-gray-900 dark:text-white">{vgToDelete.name}</span>?
                        This will also delete all LUNs created from this volume group. This action cannot be undone.
                    </p>
                    {#if vgToDelete.lun_count > 0}
                        <div class="mb-4 p-3 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-md">
                            <p class="text-sm text-yellow-800 dark:text-yellow-200">
                                Warning: This volume group contains {vgToDelete.lun_count} LUN{vgToDelete.lun_count !== 1 ? 's' : ''} that will be deleted.
                            </p>
                        </div>
                    {/if}
                    <div class="flex justify-end space-x-3">
                        <button
                            on:click={() => {
                                showVGDeleteModal = false;
                                vgToDelete = null;
                            }}
                            class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-card border border-gray-300 dark:border rounded-md hover:bg-gray-50 dark:hover:bg-muted"
                        >
                            Cancel
                        </button>
                        <button
                            on:click={deleteVG}
                            class="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700"
                        >
                            Delete Volume Group
                        </button>
                    </div>
                </div>
            </div>
        {/if}
    </div>
</div>

<svelte:body on:click={handleOutsideClick} on:keydown={handleEscapeKey} />
