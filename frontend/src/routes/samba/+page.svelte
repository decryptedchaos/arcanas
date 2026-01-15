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

    function getStatusColor(available) {
        return available
            ? "text-green-600 bg-green-100"
            : "text-gray-600 dark:text-gray-300 bg-gray-100";
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

<div class="p-6">
    <div class="mb-6">
        <h1 class="text-xl font-bold text-gray-900 dark:text-white mb-2">
            Samba Shares
        </h1>
        <p class="text-sm text-gray-600 dark:text-gray-300">
            Manage SMB/CIFS file shares
        </p>
    </div>

    <div class="flex justify-between items-center mb-6">
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
                <div class="bg-white dark:bg-card shadow rounded-lg p-6">
                    <div class="flex justify-between items-start">
                        <div>
                            <div class="flex items-center space-x-3 mb-3">
                                <h3
                                    class="text-lg font-semibold text-gray-900 dark:text-white"
                                >
                                    {share.name}
                                </h3>
                                <span
                                    class="px-2 py-1 text-xs font-medium rounded-full {getStatusColor(
                                        share.available,
                                    )}"
                                >
                                    {share.available
                                        ? "Available"
                                        : "Unavailable"}
                                </span>
                            </div>
                            <p class="text-gray-600 dark:text-gray-300 mb-2">
                                {share.comment}
                            </p>
                            <p class="text-sm text-gray-500 dark:text-gray-400">
                                Path: {share.path}
                            </p>
                            <p class="text-sm text-gray-500 dark:text-gray-400">
                                Users: {share.users.join(", ")}
                            </p>
                            <p class="text-sm text-gray-500 dark:text-gray-400">
                                Groups: {share.groups.join(", ")}
                            </p>
                        </div>
                        <div class="flex space-x-2">
                            <button
                                on:click={() => toggleShare(share)}
                                class="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
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
                                class="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
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
                                class="p-2 text-gray-400 hover:text-red-600"
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
