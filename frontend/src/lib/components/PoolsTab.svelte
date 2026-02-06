<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
    import { formatBytes } from "$lib/utils/byteUtils.js";

    export let storagePools = [];
    export let loading = true;
    export let error = null;
    export let onPoolAction = null;
</script>

{#if loading}
    <div class="flex justify-center items-center h-64">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>
{:else if error}
    <div class="bg-red-50 border border-red-200 rounded-md p-4">
        <div class="text-red-700">Error: {error}</div>
    </div>
{:else if storagePools.length === 0}
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
        {#each storagePools as pool}
            <div class="bg-white dark:bg-card shadow-lg hover:shadow-xl transition-shadow duration-200 rounded-lg p-6 border border-gray-100 dark:border-border">
                <!-- Header Section -->
                <div class="flex items-start justify-between mb-6">
                    <div class="flex items-center space-x-4">
                        <!-- Pool Icon with glow effect -->
                        <div class="relative">
                            <div class="absolute inset-0 bg-gradient-to-br from-blue-400 to-cyan-500 rounded-xl blur opacity-25"></div>
                            <div class="relative w-14 h-14 bg-gradient-to-br from-blue-100 to-cyan-100 dark:from-blue-900/40 dark:to-cyan-900/40 rounded-xl flex items-center justify-center shadow-lg">
                                <svg class="w-7 h-7 text-blue-600 dark:text-blue-400" fill="currentColor" viewBox="0 0 20 20">
                                    <path d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zM2 11a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z" />
                                </svg>
                            </div>
                        </div>

                        <div>
                            <h3 class="text-xl font-bold text-gray-900 dark:text-white">{pool.name}</h3>
                            <div class="flex items-center space-x-2 mt-1">
                                <span class="px-2 py-0.5 text-xs font-semibold rounded-full bg-gradient-to-r from-blue-500 to-cyan-500 text-white">
                                    STORAGE POOL
                                </span>
                                <span class="text-sm text-gray-500 dark:text-gray-400">
                                    {pool.devices?.length || 0} volume{(pool.devices?.length || 0) !== 1 ? 's' : ''}
                                </span>
                            </div>
                        </div>
                    </div>

                    <!-- Status Badge -->
                    <div class="flex items-center space-x-2 px-3 py-1.5 rounded-full {pool.state === 'active'
                        ? 'bg-green-100 dark:bg-green-900/30'
                        : 'bg-gray-100 dark:bg-muted'}">
                        <div class="w-2.5 h-2.5 rounded-full {pool.state === 'active'
                            ? 'bg-green-500 animate-pulse'
                            : 'bg-gray-400'}"></div>
                        <span class="text-sm font-medium {pool.state === 'active'
                            ? 'text-green-700 dark:text-green-400'
                            : 'text-gray-600 dark:text-gray-400'}">
                            {pool.state || "Inactive"}
                        </span>
                    </div>
                </div>

                <!-- Stats Grid -->
                <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
                    <!-- Total Size Card -->
                    <div class="bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 rounded-xl p-4 border border-blue-100 dark:border-blue-800">
                        <div class="flex items-center space-x-3">
                            <div class="w-10 h-10 bg-blue-100 dark:bg-blue-800 rounded-lg flex items-center justify-center">
                                <svg class="w-5 h-5 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
                                </svg>
                            </div>
                            <div>
                                <p class="text-xs text-blue-600 dark:text-blue-400 font-medium uppercase tracking-wide">Total Size</p>
                                <p class="text-lg font-bold text-gray-900 dark:text-white">{formatBytes(pool.size)}</p>
                            </div>
                        </div>
                    </div>

                    <!-- Used Card -->
                    <div class="bg-gradient-to-br from-purple-50 to-pink-50 dark:from-purple-900/20 dark:to-pink-900/20 rounded-xl p-4 border border-purple-100 dark:border-purple-800">
                        <div class="flex items-center space-x-3">
                            <div class="w-10 h-10 bg-purple-100 dark:bg-purple-800 rounded-lg flex items-center justify-center">
                                <svg class="w-5 h-5 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                                </svg>
                            </div>
                            <div>
                                <p class="text-xs text-purple-600 dark:text-purple-400 font-medium uppercase tracking-wide">Used Space</p>
                                <p class="text-lg font-bold text-gray-900 dark:text-white">{formatBytes(pool.used)}</p>
                            </div>
                        </div>
                        <!-- Usage Bar -->
                        {#if pool.size > 0}
                            <div class="mt-3 w-full bg-gray-200 dark:bg-muted rounded-full h-2 overflow-hidden">
                                <div class="h-full bg-gradient-to-r from-purple-500 to-pink-500 rounded-full transition-all duration-300" style="width: {Math.min((pool.used / pool.size) * 100, 100)}%"></div>
                            </div>
                        {/if}
                    </div>

                    <!-- Available Card -->
                    <div class="bg-gradient-to-br from-emerald-50 to-teal-50 dark:from-emerald-900/20 dark:to-teal-900/20 rounded-xl p-4 border border-emerald-100 dark:border-emerald-800">
                        <div class="flex items-center space-x-3">
                            <div class="w-10 h-10 bg-emerald-100 dark:bg-emerald-800 rounded-lg flex items-center justify-center">
                                <svg class="w-5 h-5 text-emerald-600 dark:text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                                </svg>
                            </div>
                            <div>
                                <p class="text-xs text-emerald-600 dark:text-emerald-400 font-medium uppercase tracking-wide">Available</p>
                                <p class="text-lg font-bold text-gray-900 dark:text-white">{formatBytes(pool.available)}</p>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Devices List -->
                {#if pool.devices && pool.devices.length > 0}
                    <div class="mb-6">
                        <div class="flex items-center justify-between mb-3">
                            <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide">Volumes</h4>
                            <span class="text-xs text-gray-500 dark:text-gray-400">{pool.devices?.length || 0} volume{(pool.devices?.length || 0) !== 1 ? 's' : ''}</span>
                        </div>
                        <div class="flex flex-wrap gap-2">
                            {#each pool.devices || [] as device}
                                <div class="inline-flex items-center space-x-2 px-3 py-2 bg-gray-100 dark:bg-card rounded-lg border border-gray-200 dark:border-border">
                                    <svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
                                    </svg>
                                    <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{device}</span>
                                </div>
                            {/each}
                        </div>
                    </div>
                {/if}

                <!-- Actions Footer -->
                <div class="flex items-center justify-between pt-4 mt-4 border-t border-gray-200 dark:border-border">
                    <div class="flex items-center space-x-2 text-sm text-gray-500 dark:text-gray-400">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                        </svg>
                        <span>{pool.mount_point || "Not mounted"}</span>
                    </div>
                    {#if onPoolAction}
                        <div class="flex items-center space-x-2">
                            <button
                                on:click={() => onPoolAction("mount", pool)}
                                class="inline-flex items-center space-x-2 px-4 py-2 text-sm font-medium text-blue-600 bg-blue-50 hover:bg-blue-100 dark:bg-blue-900/20 dark:hover:bg-blue-900/40 dark:text-blue-400 rounded-lg transition-colors"
                            >
                                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                                </svg>
                                <span>{pool.state === "active" ? "Unmount" : "Mount"}</span>
                            </button>
                            <button
                                on:click={() => onPoolAction("edit", pool)}
                                class="inline-flex items-center space-x-2 px-4 py-2 text-sm font-medium text-gray-600 bg-gray-50 hover:bg-gray-100 dark:bg-gray-900/20 dark:hover:bg-gray-900/40 dark:text-gray-400 rounded-lg transition-colors"
                            >
                                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                                </svg>
                                <span>Edit</span>
                            </button>
                            <button
                                on:click={() => onPoolAction("delete", pool)}
                                class="inline-flex items-center space-x-2 px-4 py-2 text-sm font-medium text-red-600 bg-red-50 hover:bg-red-100 dark:bg-red-900/20 dark:hover:bg-red-900/40 dark:text-red-400 rounded-lg transition-colors"
                            >
                                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                </svg>
                                <span>Delete</span>
                            </button>
                        </div>
                    {/if}
                </div>
            </div>
        {/each}
    </div>
{/if}
