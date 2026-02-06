<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
    import { formatBytes } from "$lib/utils/byteUtils.js";

    export let disks = [];
    export let loading = true;
    export let error = null;
</script>

{#if loading}
    <div class="flex justify-center items-center h-64">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>
{:else if error}
    <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
        <div class="text-red-700 dark:text-red-400">Error: {error}</div>
    </div>
{:else if disks.length === 0}
    <div class="text-center py-12">
        <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">
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

            <div class="bg-white dark:bg-card shadow-lg hover:shadow-xl transition-shadow duration-200 rounded-lg p-6 border border-gray-100 dark:border-border {disk.device && disk.device.startsWith('/dev/md') ? 'ring-2 ring-indigo-200 dark:ring-indigo-800' : ''}">
                <!-- Header Section -->
                <div class="flex items-start justify-between mb-6">
                    <div class="flex items-center space-x-4">
                        <!-- Disk Icon with glow effect -->
                        <div class="relative">
                            <div class="absolute inset-0 bg-gradient-to-br from-green-400 to-emerald-500 rounded-xl blur opacity-25"></div>
                            <div class="relative w-14 h-14 bg-gradient-to-br from-green-100 to-emerald-100 dark:from-green-900/40 dark:to-emerald-900/40 rounded-xl flex items-center justify-center shadow-lg">
                                <svg class="w-7 h-7 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5h4M4 7h16" />
                                </svg>
                            </div>
                        </div>

                        <div>
                            <h3 class="text-xl font-bold text-gray-900 dark:text-white">
                                {disk.name || disk.device || "Unknown"}
                            </h3>
                            <div class="flex items-center flex-wrap gap-2 mt-1">
                                <!-- Filesystem Badge -->
                                {#if disk.filesystem && disk.filesystem !== "Unknown"}
                                    <span class="px-2 py-0.5 text-xs font-semibold rounded-full bg-gradient-to-r from-slate-500 to-slate-600 text-white">
                                        {disk.filesystem.toUpperCase()}
                                    </span>
                                {/if}
                                <!-- Health Badge -->
                                {#if disk.smart?.status || disk.health}
                                    <span class="px-2 py-0.5 text-xs font-semibold rounded-full {(disk.smart?.status === 'healthy' || disk.health === 'healthy')
                                        ? 'bg-gradient-to-r from-emerald-500 to-green-500 text-white'
                                        : (disk.smart?.status === 'warning' || disk.health === 'warning')
                                        ? 'bg-gradient-to-r from-yellow-500 to-orange-500 text-white'
                                        : 'bg-gradient-to-r from-red-500 to-rose-500 text-white'}">
                                        {(disk.smart?.status || disk.health || "").toUpperCase()}
                                    </span>
                                {/if}
                                <!-- Temperature Badge -->
                                {#if disk.smart?.temperature}
                                    <span class="text-xs text-gray-500 dark:text-gray-400">
                                        {disk.smart.temperature}°C
                                    </span>
                                {/if}
                                <!-- RAID Badge -->
                                {#if disk.device && disk.device.startsWith('/dev/md')}
                                    <span class="px-2 py-0.5 text-xs font-semibold rounded-full bg-gradient-to-r from-indigo-500 to-purple-500 text-white">
                                        RAID
                                    </span>
                                {/if}
                            </div>
                        </div>
                    </div>

                    <!-- Usage Badge -->
                    {#if disk.size}
                        <div class="flex items-center space-x-2 px-3 py-1.5 rounded-full bg-gray-100 dark:bg-muted">
                            <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                                {Math.round(((disk.used || 0) / disk.size) * 100)}%
                            </span>
                        </div>
                    {/if}
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
                                <p class="text-xs text-blue-600 dark:text-blue-400 font-medium uppercase tracking-wide">Capacity</p>
                                <p class="text-lg font-bold text-gray-900 dark:text-white">{formatBytes(disk.size || 0)}</p>
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
                                <p class="text-xs text-purple-600 dark:text-purple-400 font-medium uppercase tracking-wide">Used</p>
                                <p class="text-lg font-bold text-gray-900 dark:text-white">{formatBytes(disk.used || 0)}</p>
                            </div>
                        </div>
                        <!-- Usage Bar -->
                        {#if disk.size > 0}
                            <div class="mt-3 w-full bg-gray-200 dark:bg-muted rounded-full h-2 overflow-hidden">
                                <div class="h-full bg-gradient-to-r from-purple-500 to-pink-500 rounded-full transition-all duration-300" style="width: {Math.min(((disk.used || 0) / disk.size) * 100, 100)}%"></div>
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
                                <p class="text-lg font-bold text-gray-900 dark:text-white">{formatBytes(disk.available || (disk.size || 0) - (disk.used || 0))}</p>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Info Footer -->
                <div class="flex items-center justify-between pt-4 mt-4 border-t border-gray-200 dark:border-border">
                    <div class="space-y-1">
                        <!-- Device Path -->
                        <div class="flex items-center space-x-2 text-sm">
                            <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
                            </svg>
                            <span class="text-gray-500 dark:text-gray-400 font-mono text-xs">{disk.device || "Unknown"}</span>
                        </div>
                        <!-- Mount Point -->
                        <div class="flex items-center space-x-2 text-sm">
                            <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                            </svg>
                            <span class="text-gray-500 dark:text-gray-400">{disk.mountpoint || "Not mounted"}</span>
                        </div>
                    </div>

                    <!-- Model info -->
                    {#if disk.model && disk.model !== "Unknown"}
                        <div class="text-sm text-gray-500 dark:text-gray-400 text-right">{disk.model}</div>
                    {/if}
                </div>
            </div>
        {/each}
    </div>
{/if}
