<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
    import { sambaAPI, storageAPI } from "$lib/api.js";
    import { onMount } from "svelte";

    let shares = [];
    let loading = true;
    let error = null;
    let showCreateModal = false;
    let showEditModal = false;
    let selectedShare = null;

    // Storage pools data
    let storagePools = [];
    let poolsLoading = true;

    // Form data for new share
    let newShare = {
        name: "",
        path: "",
        comment: "",
        users: "",
        groups: "",
        guest_ok: false,
        read_only: false,
        browseable: true,
    };

    // Form data for editing share
    let editShare = {
        name: "",
        path: "",
        comment: "",
        users: "",
        groups: "",
        guest_ok: false,
        read_only: false,
        browseable: true,
        available: true,
    };

    async function loadShares() {
        try {
            loading = true;
            shares = await sambaAPI.getShares();
            error = null;
        } catch (err) {
            error = err.message;
            console.error("Failed to load Samba shares:", err);
        } finally {
            loading = false;
        }
    }

    onMount(() => {
        loadShares();
        loadStoragePools();
    });

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

    async function toggleShare(share) {
        try {
            await sambaAPI.toggleShare(share.name);
            share.available = !share.available;
            shares = [...shares];
        } catch (err) {
            console.error("Failed to toggle share:", err);
            share.available = !share.available;
            shares = [...shares];
        }
    }

    async function deleteShare(shareId) {
        if (confirm("Are you sure you want to delete this Samba share?")) {
            try {
                await sambaAPI.deleteShare(shareId);
                shares = shares.filter((s) => s.name !== shareId);
            } catch (err) {
                console.error("Failed to delete share:", err);
            }
        }
    }

    async function createShare() {
        try {
            // Debug logging
            console.log("Creating share with data:", newShare);
            console.log("Selected path:", newShare.path);
            console.log("Available storage pools:", storagePools);

            // Validate that a storage pool is selected
            if (!newShare.path || newShare.path.trim() === "") {
                alert("Please select a storage pool for the share path.");
                return;
            }

            const shareData = {
                ...newShare,
                users: newShare.users
                    .split(",")
                    .map((u) => u.trim())
                    .filter(Boolean),
                groups: newShare.groups
                    .split(",")
                    .map((g) => g.trim())
                    .filter(Boolean),
            };

            console.log("Sending share data:", shareData);

            const createdShare = await sambaAPI.createShare(shareData);
            shares = [...shares, createdShare];
            showCreateModal = false;
            resetForm();
        } catch (err) {
            console.error("Failed to create share:", err);
            alert(`Failed to create share: ${err.message || err}`);
        }
    }

    function openEditModal(share) {
        selectedShare = share;
        editShare = {
            name: share.name,
            path: share.path,
            comment: share.comment || "",
            users: share.users.join(", "),
            groups: share.groups.join(", "),
            guest_ok: share.guest_ok,
            read_only: share.read_only,
            browseable: share.browseable,
            available: share.available,
        };
        showEditModal = true;
    }

    async function updateShare() {
        try {
            const shareData = {
                ...editShare,
                users: editShare.users
                    .split(",")
                    .map((u) => u.trim())
                    .filter(Boolean),
                groups: editShare.groups
                    .split(",")
                    .map((g) => g.trim())
                    .filter(Boolean),
            };

            const updatedShare = await sambaAPI.updateShare(
                selectedShare.name,
                shareData,
            );
            shares = shares.map((s) =>
                s.name === selectedShare.name ? updatedShare : s,
            );
            showEditModal = false;
            selectedShare = null;
        } catch (err) {
            console.error("Failed to update share:", err);
        }
    }

    function resetForm() {
        newShare = {
            name: "",
            path: "",
            comment: "",
            users: "",
            groups: "",
            guest_ok: false,
            read_only: false,
            browseable: true,
        };
    }
</script>

<div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
        <div class="flex items-center space-x-4">
            <!-- Samba Icon with glow effect -->
            <div class="relative">
                <div class="absolute inset-0 bg-gradient-to-br from-orange-400 to-amber-500 rounded-xl blur opacity-25"></div>
                <div class="relative w-14 h-14 bg-gradient-to-br from-orange-100 to-amber-100 dark:from-orange-900/40 dark:to-amber-900/40 rounded-xl flex items-center justify-center shadow-lg">
                    <svg class="w-7 h-7 text-orange-600 dark:text-orange-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z"/>
                    </svg>
                </div>
            </div>
            <div>
                <h2 class="text-xl font-bold text-gray-900 dark:text-white">
                    Samba Shares
                </h2>
                <p class="text-sm text-gray-600 dark:text-gray-300 mt-1">
                    Manage SMB/CIFS file shares
                </p>
            </div>
        </div>
        <div class="flex items-center space-x-3">
            <button
                on:click={loadShares}
                class="btn btn-secondary"
                disabled={loading}
            >
                Refresh
            </button>
            <button
                on:click={() => (showCreateModal = true)}
                class="btn btn-primary"
            >
                Create Share
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
    {:else if shares.length === 0}
        <div class="text-center py-12">
            <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">
                No Samba shares found
            </h3>
            <p class="text-gray-600 dark:text-gray-300">
                No Samba shares are configured.
            </p>
        </div>
    {:else}
        <div class="space-y-4">
            {#each shares as share}
                <div class="bg-white dark:bg-card shadow-lg hover:shadow-xl transition-shadow duration-200 rounded-lg p-6 border border-gray-100 dark:border-border">
                    <!-- Header Section -->
                    <div class="flex items-start justify-between mb-6">
                        <div class="flex items-center space-x-4">
                            <!-- Share Icon with glow effect -->
                            <div class="relative">
                                <div class="absolute inset-0 {share.available
                                    ? 'bg-gradient-to-br from-orange-400 to-amber-500'
                                    : 'bg-gradient-to-br from-gray-400 to-gray-500'} rounded-xl blur opacity-25"></div>
                                <div class="relative w-14 h-14 {share.available
                                    ? 'bg-gradient-to-br from-orange-100 to-amber-100 dark:from-orange-900/40 dark:to-amber-900/40'
                                    : 'bg-gradient-to-br from-gray-100 to-gray-200 dark:from-gray-800/40 dark:to-gray-700/40'} rounded-xl flex items-center justify-center shadow-lg">
                                    <svg class="w-7 h-7 {share.available ? 'text-orange-600 dark:text-orange-400' : 'text-gray-500 dark:text-gray-400'}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z"/>
                                    </svg>
                                </div>
                            </div>

                            <div>
                                <h3 class="text-xl font-bold text-gray-900 dark:text-white">
                                    {share.name}
                                </h3>
                                <div class="flex items-center flex-wrap gap-2 mt-1">
                                    <!-- Status Badge -->
                                    <span class="px-2 py-0.5 text-xs font-semibold rounded-full {share.available
                                        ? 'bg-gradient-to-r from-orange-500 to-amber-500 text-white'
                                        : 'bg-gradient-to-r from-gray-500 to-gray-600 text-white'}">
                                        {share.available ? "AVAILABLE" : "UNAVAILABLE"}
                                    </span>
                                    <!-- Read Only Badge -->
                                    {#if share.read_only}
                                        <span class="px-2 py-0.5 text-xs font-semibold rounded-full bg-gradient-to-r from-blue-500 to-indigo-500 text-white">
                                            READ ONLY
                                        </span>
                                    {/if}
                                    <!-- Guest Access Badge -->
                                    {#if share.guest_ok}
                                        <span class="px-2 py-0.5 text-xs font-semibold rounded-full bg-gradient-to-r from-green-500 to-emerald-500 text-white">
                                            GUEST OK
                                        </span>
                                    {/if}
                                </div>
                            </div>
                        </div>

                        <!-- Action Buttons -->
                        <div class="flex space-x-2">
                            <button
                                on:click={() => toggleShare(share)}
                                class="p-2 rounded-lg {share.available
                                    ? 'text-green-500 hover:bg-green-50 dark:hover:bg-green-900/20'
                                    : 'text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800/20'}"
                                title="Toggle share availability"
                                aria-label="Toggle share availability"
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
                                        d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                                    />
                                </svg>
                            </button>
                            <button
                                on:click={() => openEditModal(share)}
                                class="p-2 rounded-lg text-gray-400 hover:text-blue-500 hover:bg-blue-50 dark:hover:bg-blue-900/20"
                                title="Edit share"
                                aria-label="Edit share"
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
                                on:click={() => deleteShare(share.name)}
                                class="p-2 rounded-lg text-gray-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20"
                                title="Delete share"
                                aria-label="Delete share"
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

                    <!-- Share Details -->
                    <div class="space-y-4">
                        <!-- Comment -->
                        {#if share.comment}
                            <div>
                                <p class="text-sm text-gray-600 dark:text-gray-300">
                                    {share.comment}
                                </p>
                            </div>
                        {/if}

                        <!-- Details Grid -->
                        <div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                            <div>
                                <p class="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1">
                                    Path
                                </p>
                                <p class="font-mono text-sm text-gray-900 dark:text-white flex items-center">
                                    <svg class="w-4 h-4 mr-1 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>
                                    </svg>
                                    {share.path}
                                </p>
                            </div>
                            <div>
                                <p class="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1">
                                    Browseable
                                </p>
                                <div class="flex items-center space-x-2">
                                    <div class="w-2 h-2 rounded-full {share.browseable ? 'bg-green-500' : 'bg-gray-400'}"></div>
                                    <span class="text-sm font-medium text-gray-900 dark:text-white">
                                        {share.browseable ? 'Yes' : 'No'}
                                    </span>
                                </div>
                            </div>
                            <div>
                                <p class="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1">
                                    Users ({share.users.length})
                                </p>
                                <p class="text-sm font-medium text-gray-900 dark:text-white">
                                    {share.users.length > 0 ? share.users.join(", ") : 'None'}
                                </p>
                            </div>
                            <div>
                                <p class="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1">
                                    Groups ({share.groups.length})
                                </p>
                                <p class="text-sm font-medium text-gray-900 dark:text-white">
                                    {share.groups.length > 0 ? share.groups.join(", ") : 'None'}
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
            {/each}
        </div>
    {/if}
</div>

<!-- Create Modal -->
{#if showCreateModal}
    <div class="fixed inset-0 z-50 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
            <div
                class="fixed inset-0 bg-gray-500 bg-opacity-75"
                role="button"
                tabindex="0"
                on:click={() => {
                    showCreateModal = false;
                    resetForm();
                }}
                on:keydown={(e) =>
                    e.key === "Escape" && (showCreateModal = false)}
                aria-label="Close modal"
            ></div>
            <div
                class="relative bg-white dark:bg-card rounded-lg max-w-2xl w-full p-6"
            >
                <h3
                    class="text-lg font-medium text-gray-900 dark:text-white mb-4"
                >
                    Create Samba Share
                </h3>
                <form on:submit|preventDefault={createShare} class="space-y-4">
                    <input
                        type="text"
                        bind:value={newShare.name}
                        placeholder="Share Name"
                        class="w-full p-2 border rounded bg-white dark:bg-muted text-gray-900 dark:text-white border-gray-300 dark:border"
                        required
                    />
                    {#if poolsLoading}
                        <div
                            class="w-full p-2 border rounded bg-gray-100 dark:bg-muted text-gray-500 dark:text-gray-300"
                        >
                            Loading storage pools...
                        </div>
                    {:else if storagePools.length === 0}
                        <div
                            class="w-full p-2 border rounded bg-yellow-50 dark:bg-yellow-900/20 text-yellow-800 dark:text-yellow-200"
                        >
                            No storage pools available. Please create a storage
                            pool first.
                        </div>
                    {:else}
                        <select
                            bind:value={newShare.path}
                            class="w-full p-2 border rounded bg-white dark:bg-muted text-gray-900 dark:text-white border-gray-300 dark:border"
                            required
                        >
                            <option value="">Select a storage pool</option>
                            {#each storagePools as pool}
                                <option value={pool.mount_point}>
                                    {pool.name} ({pool.type}) - {pool.mount_point}
                                </option>
                            {/each}
                        </select>
                    {/if}
                    <input
                        type="text"
                        bind:value={newShare.comment}
                        placeholder="Comment"
                        class="w-full p-2 border rounded bg-white dark:bg-muted text-gray-900 dark:text-white border-gray-300 dark:border"
                    />
                    <input
                        type="text"
                        bind:value={newShare.users}
                        placeholder="Users (comma separated)"
                        class="w-full p-2 border rounded bg-white dark:bg-muted text-gray-900 dark:text-white border-gray-300 dark:border"
                    />
                    <input
                        type="text"
                        bind:value={newShare.groups}
                        placeholder="Groups (comma separated)"
                        class="w-full p-2 border rounded bg-white dark:bg-muted text-gray-900 dark:text-white border-gray-300 dark:border"
                    />
                    <div class="space-y-2">
                        <label class="flex items-center space-x-2">
                            <input
                                type="checkbox"
                                bind:checked={newShare.guest_ok}
                                class="rounded border-gray-300 dark:border"
                            />
                            <span class="text-sm text-gray-700 dark:text-gray-300"
                                >Allow guest access</span
                            >
                        </label>
                        <label class="flex items-center space-x-2">
                            <input
                                type="checkbox"
                                bind:checked={newShare.read_only}
                                class="rounded border-gray-300 dark:border"
                            />
                            <span class="text-sm text-gray-700 dark:text-gray-300"
                                >Read only</span
                            >
                        </label>
                        <label class="flex items-center space-x-2">
                            <input
                                type="checkbox"
                                bind:checked={newShare.browseable}
                                class="rounded border-gray-300 dark:border"
                            />
                            <span class="text-sm text-gray-700 dark:text-gray-300"
                                >Browseable</span
                            >
                        </label>
                    </div>
                    <div class="flex space-x-2">
                        <button
                            type="button"
                            on:click={() => {
                                showCreateModal = false;
                                resetForm();
                            }}
                            class="btn btn-secondary">Cancel</button
                        >
                        <button type="submit" class="btn btn-primary"
                            >Create</button
                        >
                    </div>
                </form>
            </div>
        </div>
    </div>
{/if}

<!-- Edit Modal -->
{#if showEditModal}
    <div class="fixed inset-0 z-50 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
            <div
                class="fixed inset-0 bg-gray-500 bg-opacity-75"
                role="button"
                tabindex="0"
                on:click={() => {
                    showEditModal = false;
                    selectedShare = null;
                }}
                on:keydown={(e) =>
                    e.key === "Escape" && (showEditModal = false)}
                aria-label="Close modal"
            ></div>
            <div
                class="relative bg-white dark:bg-card rounded-lg max-w-2xl w-full p-6"
            >
                <h3
                    class="text-lg font-medium text-gray-900 dark:text-white mb-4"
                >
                    Edit Samba Share
                </h3>
                <form on:submit|preventDefault={updateShare} class="space-y-4">
                    <input
                        type="text"
                        bind:value={editShare.name}
                        placeholder="Share Name"
                        class="w-full p-2 border rounded bg-white dark:bg-muted text-gray-900 dark:text-white border-gray-300 dark:border"
                        required
                    />
                    {#if poolsLoading}
                        <div
                            class="w-full p-2 border rounded bg-gray-100 dark:bg-muted text-gray-500 dark:text-gray-300"
                        >
                            Loading storage pools...
                        </div>
                    {:else if storagePools.length === 0}
                        <div
                            class="w-full p-2 border rounded bg-yellow-50 dark:bg-yellow-900/20 text-yellow-800 dark:text-yellow-200"
                        >
                            No storage pools available. Please create a storage
                            pool first.
                        </div>
                    {:else}
                        <select
                            bind:value={editShare.path}
                            class="w-full p-2 border rounded bg-white dark:bg-muted text-gray-900 dark:text-white border-gray-300 dark:border"
                            required
                        >
                            <option value="">Select a storage pool</option>
                            {#each storagePools as pool}
                                <option value={pool.mount_point}>
                                    {pool.name} ({pool.type}) - {pool.mount_point}
                                </option>
                            {/each}
                        </select>
                    {/if}
                    <input
                        type="text"
                        bind:value={editShare.comment}
                        placeholder="Comment"
                        class="w-full p-2 border rounded bg-white dark:bg-muted text-gray-900 dark:text-white border-gray-300 dark:border"
                    />
                    <input
                        type="text"
                        bind:value={editShare.users}
                        placeholder="Users (comma separated)"
                        class="w-full p-2 border rounded bg-white dark:bg-muted text-gray-900 dark:text-white border-gray-300 dark:border"
                    />
                    <input
                        type="text"
                        bind:value={editShare.groups}
                        placeholder="Groups (comma separated)"
                        class="w-full p-2 border rounded bg-white dark:bg-muted text-gray-900 dark:text-white border-gray-300 dark:border"
                    />
                    <div class="space-y-2">
                        <label class="flex items-center space-x-2">
                            <input
                                type="checkbox"
                                bind:checked={editShare.guest_ok}
                                class="rounded border-gray-300 dark:border"
                            />
                            <span class="text-sm text-gray-700 dark:text-gray-300"
                                >Allow guest access</span
                            >
                        </label>
                        <label class="flex items-center space-x-2">
                            <input
                                type="checkbox"
                                bind:checked={editShare.read_only}
                                class="rounded border-gray-300 dark:border"
                            />
                            <span class="text-sm text-gray-700 dark:text-gray-300"
                                >Read only</span
                            >
                        </label>
                        <label class="flex items-center space-x-2">
                            <input
                                type="checkbox"
                                bind:checked={editShare.browseable}
                                class="rounded border-gray-300 dark:border"
                            />
                            <span class="text-sm text-gray-700 dark:text-gray-300"
                                >Browseable</span
                            >
                        </label>
                        <label class="flex items-center space-x-2">
                            <input
                                type="checkbox"
                                bind:checked={editShare.available}
                                class="rounded border-gray-300 dark:border"
                            />
                            <span class="text-sm text-gray-700 dark:text-gray-300"
                                >Available</span
                            >
                        </label>
                    </div>
                    <div class="flex space-x-2">
                        <button
                            type="button"
                            on:click={() => {
                                showEditModal = false;
                                selectedShare = null;
                            }}
                            class="btn btn-secondary">Cancel</button
                        >
                        <button type="submit" class="btn btn-primary"
                            >Update</button
                        >
                    </div>
                </form>
            </div>
        </div>
    </div>
{/if}
