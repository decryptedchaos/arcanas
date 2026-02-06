<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";
    import { diskAPI, storageAPI } from "$lib/api.js";
    import { onMount } from "svelte";
    import DisksTab from "$lib/components/DisksTab.svelte";
    import RAIDTab from "$lib/components/RAIDTab.svelte";
    import PoolsTab from "$lib/components/PoolsTab.svelte";
    import StorageBuilderCard from "$lib/components/StorageBuilderCard.svelte";

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

    // Modal state for pool actions
    let showDeleteModal = false;
    let poolToDelete = null;
    let deleteConfirmation = "";
    let showDetailsModal = false;
    let selectedPoolDetails = null;
    let showCreatePoolModal = false;
    let poolCreating = false;
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

    // Form data for editing storage pool
    let editingPool = null;
    let editedPool = {
        name: "",
        config: ""
    };
    let showEditPoolModal = false;
    let poolEditing = false;

    // Form data for creating storage pool
    let newPool = {
        name: "",
        type: "mergerfs",
        devices: [],
        config: ""
    };

    // Notification system
    let notifications = [];
    let notificationId = 0;

    function showNotification(message, type = 'info') {
        const id = ++notificationId;
        notifications = [...notifications, { id, message, type }];

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
            const result = await storageAPI.getPools();
            if (result === null || result === undefined) {
                storagePools = [];
            } else if (Array.isArray(result)) {
                storagePools = result;
            } else {
                storagePools = [];
            }
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
                console.log("Toggle mount for pool:", pool.name);
                break;
            case "edit":
                editingPool = pool;
                editedPool = {
                    name: pool.name,
                    config: pool.config || ""
                };
                showEditPoolModal = true;
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

    async function confirmDeletePool() {
        if (!poolToDelete) return;

        try {
            await storageAPI.deletePool(poolToDelete.name);
            await loadStoragePools();
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

    // Get available (unmounted) devices for pool creation
    $: availableDevices = diskStats.filter(disk => {
        if (!disk.device || disk.device === "") return false;
        if (disk.filesystem === "linux_raid_member") return false;
        if (disk.mountpoint && disk.mountpoint !== "") return false;
        if (disk.device && disk.device.startsWith("/dev/md")) return true;
        if (disk.filesystem && disk.filesystem !== "unknown") return true;
        return true;
    });

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
            await loadDiskStats();
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
        loadDiskStats();
        loadRAIDArrays();
        loadStoragePools();
    });

    $: disks = Array.isArray(diskStats) ? diskStats : [];
    $: raidArraysSafe = Array.isArray(raidArrays) ? raidArrays : [];
    $: storagePoolsSafe = Array.isArray(storagePools) ? storagePools : [];
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

    <!-- Storage Builder Card -->
    <div class="mb-6">
        <StorageBuilderCard />
    </div>

    <!-- Tab Navigation -->
    <div class="border-b border-gray-200 dark:border-border mb-6">
        <nav class="-mb-px flex space-x-8">
            <button
                class="py-2 px-1 border-b-2 font-medium text-sm {activeTab === 'disks'
                    ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                    : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-border'}"
                on:click={() => goto("?tab=disks", { replaceState: true })}
            >
                Disks
            </button>
            <button
                class="py-2 px-1 border-b-2 font-medium text-sm {activeTab === 'raid'
                    ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                    : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-border'}"
                on:click={() => goto("?tab=raid", { replaceState: true })}
            >
                RAID Arrays
            </button>
            <button
                class="py-2 px-1 border-b-2 font-medium text-sm {activeTab === 'pools'
                    ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                    : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-border'}"
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
            <a
                href="/storage-builder/raid"
                class="btn btn-primary inline-flex items-center"
            >
                Create RAID Array
            </a>
        {:else if activeTab === "pools"}
            <a
                href="/storage-builder/pool"
                class="btn btn-primary inline-flex items-center"
            >
                Create Storage Pool
            </a>
        {/if}
    </div>

    <!-- Tab Content -->
    <div role="region">
        {#if activeTab === "disks"}
            <DisksTab disks={disks} {loading} {error} />
        {:else if activeTab === "raid"}
            <RAIDTab raidArrays={raidArraysSafe} loading={raidLoading} error={raidError} onDelete={handleRAIDDelete} />
        {:else if activeTab === "pools"}
            <PoolsTab
                storagePools={storagePoolsSafe}
                loading={poolsLoading}
                error={poolsError}
                onPoolAction={handlePoolAction}
            />
        {/if}
    </div>
</div>

<!-- Delete Confirmation Modal -->
{#if showDeleteModal}
    <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
        <div class="bg-white dark:bg-card rounded-lg p-6 max-w-md w-full mx-4">
            <div class="flex items-center mb-4">
                <svg class="w-6 h-6 text-red-600 mr-3" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
                </svg>
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                    Delete Storage Pool
                </h3>
            </div>
            <p class="text-gray-600 dark:text-gray-300 mb-4">
                Are you sure you want to delete the storage pool "{poolToDelete?.name}"?
                This action cannot be undone.
            </p>
            <div class="mb-4">
                <label for="deleteConfirmation" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
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
                    disabled={deleteConfirmation !== `DELETE ${poolToDelete?.name.toUpperCase()}`}
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
    <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
        <div class="bg-white dark:bg-card rounded-lg p-6 max-w-2xl w-full mx-4">
            <div class="flex items-center justify-between mb-4">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                    Storage Pool Details
                </h3>
                <button
                    type="button"
                    class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                    on:click={closeDetailsModal}
                    aria-label="Close details modal"
                >
                    <svg class="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                    </svg>
                </button>
            </div>
            <div class="space-y-4">
                <div class="grid grid-cols-2 gap-4">
                    <div>
                        <p class="text-sm font-medium text-gray-500 dark:text-gray-400">Name</p>
                        <p class="text-gray-900 dark:text-white">{selectedPoolDetails?.name}</p>
                    </div>
                    <div>
                        <p class="text-sm font-medium text-gray-500 dark:text-gray-400">Type</p>
                        <p class="text-gray-900 dark:text-white">Storage Pool</p>
                    </div>
                    <div>
                        <p class="text-sm font-medium text-gray-500 dark:text-gray-400">State</p>
                        <p class="text-gray-900 dark:text-white">{selectedPoolDetails?.state || "Unknown"}</p>
                    </div>
                    <div>
                        <p class="text-sm font-medium text-gray-500 dark:text-gray-400">Mount Point</p>
                        <p class="text-gray-900 dark:text-white">{selectedPoolDetails?.mount_point || "Not mounted"}</p>
                    </div>
                </div>
                <div class="grid grid-cols-2 gap-4">
                    <div>
                        <p class="text-sm font-medium text-gray-500 dark:text-gray-400">Size</p>
                        <p class="text-gray-900 dark:text-white">{selectedPoolDetails?.size || "N/A"}</p>
                    </div>
                    <div>
                        <p class="text-sm font-medium text-gray-500 dark:text-gray-400">Used</p>
                        <p class="text-gray-900 dark:text-white">{selectedPoolDetails?.used || "N/A"}</p>
                    </div>
                    <div>
                        <p class="text-sm font-medium text-gray-500 dark:text-gray-400">Available</p>
                        <p class="text-gray-900 dark:text-white">{selectedPoolDetails?.available || "N/A"}</p>
                    </div>
                    <div>
                        <p class="text-sm font-medium text-gray-500 dark:text-gray-400">Created</p>
                        <p class="text-gray-900 dark:text-white">
                            {selectedPoolDetails?.created_at
                                ? new Date(selectedPoolDetails.created_at).toLocaleDateString()
                                : "N/A"}
                        </p>
                    </div>
                </div>
                {#if selectedPoolDetails?.config}
                    <div>
                        <p class="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Configuration</p>
                        <pre class="bg-gray-100 dark:bg-muted p-3 rounded text-xs overflow-auto">{JSON.stringify(selectedPoolDetails.config, null, 2)}</pre>
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
                on:keydown={(e) => e.key === "Escape" && (showCreatePoolModal = false)}
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
                <h3 id="create-pool-title" class="text-lg font-medium text-gray-900 dark:text-white mb-4">
                    Create Storage Pool
                </h3>
                <form on:submit|preventDefault={createStoragePool} class="space-y-4">
                    <div>
                        <label for="poolName" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
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
                        <label for="poolType" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
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
                                                {disk.device} • {disk.size} bytes
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
                        <label for="poolConfig" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
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
                                <svg class="animate-spin h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                                </svg>
                                <span>Creating...</span>
                            {:else}
                                <span>Create Pool</span>
                            {/if}
                        </button>
                    </div>
                </form>

                {#if poolCreating}
                    <div class="absolute inset-0 bg-white dark:bg-card bg-opacity-90 rounded-lg flex flex-col items-center justify-center z-10">
                        <div class="flex flex-col items-center space-y-4">
                            <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
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
    <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
        <div class="bg-white dark:bg-card rounded-lg p-6 max-w-md w-full mx-4">
            <div class="flex items-center mb-4">
                <svg class="w-6 h-6 text-red-600 mr-3" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
                </svg>
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                    Delete RAID Array
                </h3>
            </div>
            <p class="text-gray-600 dark:text-gray-300 mb-4">
                Are you sure you want to delete the RAID array "{raidToDelete?.name}"?
                This will stop the array and zero the superblock on all member devices.
                This action cannot be undone.
            </p>
            <div class="mb-4">
                <label for="raidDeleteConfirmation" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
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
                    disabled={raidDeleteConfirmation !== `DELETE ${raidToDelete?.name.toUpperCase()}`}
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
                <h3 id="edit-pool-title" class="text-lg font-medium text-gray-900 dark:text-white mb-4">
                    Edit Storage Pool
                </h3>
                <form on:submit|preventDefault={saveEditedPool} class="space-y-4">
                    <div>
                        <label for="editPoolName" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
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
                        <label for="editPoolConfig" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
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
                                <svg class="animate-spin h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
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
            <div class="relative bg-white dark:bg-card rounded-lg max-w-2xl w-full p-6" on:click|stopPropagation>
                <h3 id="create-raid-title" class="text-lg font-medium text-gray-900 dark:text-white mb-4">
                    Create RAID Array
                </h3>
                <form on:submit|preventDefault={createRAIDArray} class="space-y-4">
                    <div>
                        <label for="raidName" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
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
                        <label for="raidLevel" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
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
                                                    {disk.device} • {disk.size} bytes
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
                            <svg class="w-5 h-5 text-yellow-400 mr-2" fill="currentColor" viewBox="0 0 20 20">
                                <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
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
