<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
    import { formatBytes } from "$lib/utils/byteUtils.js";

    export let raidArrays = [];
    export let loading = true;
    export let error = null;
    export let onDelete = null;
</script>

{#if loading}
    <div class="flex justify-center items-center h-64">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>
{:else if error}
    <div class="bg-red-50 border border-red-200 rounded-md p-4">
        <div class="text-red-700">Error: {error}</div>
    </div>
{:else if raidArrays.length === 0}
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
        {#each raidArrays as array}
            <div class="bg-white dark:bg-card shadow-lg hover:shadow-xl transition-shadow duration-200 rounded-lg p-6 border border-gray-100 dark:border-border">
                <!-- Header Section -->
                <div class="flex items-start justify-between mb-6">
                    <div class="flex items-center space-x-4">
                        <!-- RAID Icon with glow effect -->
                        <div class="relative">
                            <div class="absolute inset-0 bg-gradient-to-br from-orange-400 to-purple-500 rounded-xl blur opacity-25"></div>
                            <div class="relative w-14 h-14 bg-gradient-to-br from-orange-100 to-orange-200 dark:from-orange-900/40 dark:to-purple-900/40 rounded-xl flex items-center justify-center shadow-lg">
                                <svg class="w-7 h-7 text-orange-600 dark:text-orange-400" fill="currentColor" viewBox="0 0 20 20">
                                    <path d="M5 3a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2V5a2 2 0 00-2-2H5zM5 11a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2H5zM11 5a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V5zM13 11a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2h-2z" />
                                </svg>
                            </div>
                        </div>

                        <div>
                            <h3 class="text-xl font-bold text-gray-900 dark:text-white">{array.name}</h3>
                            <div class="flex items-center space-x-2 mt-1">
                                <span class="px-2 py-0.5 text-xs font-semibold rounded-full bg-gradient-to-r from-orange-500 to-purple-500 text-white">
                                    {array.level.toUpperCase()}
                                </span>
                                <span class="text-sm text-gray-500 dark:text-gray-400">
                                    {array.devices?.length || 0} device{(array.devices?.length || 0) !== 1 ? 's' : ''}
                                </span>
                            </div>
                        </div>
                    </div>

                    <!-- Status Badge -->
                    <div class="flex items-center space-x-2 px-3 py-1.5 rounded-full {array.state === 'clean' || array.state === 'active'
                        ? 'bg-green-100 dark:bg-green-900/30'
                        : array.state === 'degraded'
                        ? 'bg-yellow-100 dark:bg-yellow-900/30'
                        : 'bg-red-100 dark:bg-red-900/30'}">
                        <div class="w-2.5 h-2.5 rounded-full {array.state === 'clean' || array.state === 'active'
                            ? 'bg-green-500 animate-pulse'
                            : array.state === 'degraded'
                            ? 'bg-yellow-500'
                            : 'bg-red-500'}"></div>
                        <span class="text-sm font-medium {array.state === 'clean' || array.state === 'active'
                            ? 'text-green-700 dark:text-green-400'
                            : array.state === 'degraded'
                            ? 'text-yellow-700 dark:text-yellow-400'
                            : 'text-red-700 dark:text-red-400'}">
                            {array.state || "Unknown"}
                        </span>
                    </div>
                </div>

                <!-- Stats Grid -->
                <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
                    <!-- Size Card -->
                    <div class="bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 rounded-xl p-4 border border-blue-100 dark:border-blue-800">
                        <div class="flex items-center space-x-3">
                            <div class="w-10 h-10 bg-blue-100 dark:bg-blue-800 rounded-lg flex items-center justify-center">
                                <svg class="w-5 h-5 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
                                </svg>
                            </div>
                            <div>
                                <p class="text-xs text-blue-600 dark:text-blue-400 font-medium uppercase tracking-wide">Total Size</p>
                                <p class="text-lg font-bold text-gray-900 dark:text-white">{formatBytes(array.size)}</p>
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
                                <p class="text-lg font-bold text-gray-900 dark:text-white">{array.used ? formatBytes(array.used) : "N/A"}</p>
                            </div>
                        </div>
                        <!-- Usage Bar -->
                        {#if array.used && array.size}
                            <div class="mt-3 w-full bg-gray-200 dark:bg-muted rounded-full h-2 overflow-hidden">
                                <div class="h-full bg-gradient-to-r from-purple-500 to-pink-500 rounded-full transition-all duration-300" style="width: {Math.min((array.used / array.size) * 100, 100)}%"></div>
                            </div>
                        {/if}
                    </div>

                    <!-- Health Card -->
                    <div class="bg-gradient-to-br from-emerald-50 to-teal-50 dark:from-emerald-900/20 dark:to-teal-900/20 rounded-xl p-4 border border-emerald-100 dark:border-emerald-800">
                        <div class="flex items-center space-x-3">
                            <div class="w-10 h-10 bg-emerald-100 dark:bg-emerald-800 rounded-lg flex items-center justify-center">
                                <svg class="w-5 h-5 text-emerald-600 dark:text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                                </svg>
                            </div>
                            <div>
                                <p class="text-xs text-emerald-600 dark:text-emerald-400 font-medium uppercase tracking-wide">Health</p>
                                <p class="text-lg font-bold text-gray-900 dark:text-white">{array.health}%</p>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Devices List -->
                <div>
                    <div class="flex items-center justify-between mb-3">
                        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide">Member Devices</h4>
                        <span class="text-xs text-gray-500 dark:text-gray-400">{array.devices?.length || 0} disk{(array.devices?.length || 0) !== 1 ? 's' : ''}</span>
                    </div>
                    <div class="flex flex-wrap gap-2">
                        {#each array.devices || [] as device}
                            <div class="inline-flex items-center space-x-2 px-3 py-2 bg-gray-100 dark:bg-card rounded-lg border border-gray-200 dark:border-border">
                                <svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
                                </svg>
                                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{device}</span>
                            </div>
                        {/each}
                    </div>
                </div>

                <!-- Actions Footer -->
                <div class="flex items-center justify-between pt-4 mt-4 border-t border-gray-200 dark:border-border">
                    <div class="flex items-center space-x-2 text-sm text-gray-500 dark:text-gray-400">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                        </svg>
                        <span>{array.mount_point || "Not mounted"}</span>
                    </div>
                    {#if onDelete}
                        <button
                            on:click={() => onDelete(array)}
                            class="inline-flex items-center space-x-2 px-4 py-2 text-sm font-medium text-red-600 bg-red-50 hover:bg-red-100 dark:bg-red-900/20 dark:hover:bg-red-900/40 dark:text-red-400 rounded-lg transition-colors"
                            title="Delete RAID Array"
                        >
                            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                            </svg>
                            <span>Delete Array</span>
                        </button>
                    {/if}
                </div>
            </div>
        {/each}
    </div>
{/if}
