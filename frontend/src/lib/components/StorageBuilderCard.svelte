<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { goto } from '$app/navigation';
</script>

<div
  class="flex items-center p-4 bg-white dark:bg-card rounded-lg border border-gray-200 dark:border-border hover:shadow-md transition-all duration-200 cursor-pointer group"
  on:click={() => goto('/storage-builder')}
  on:keydown={(e) => e.key === 'Enter' && goto('/storage-builder')}
  role="button"
  tabindex="0"
>
  <div class="flex-shrink-0 p-3 bg-indigo-100 dark:bg-indigo-900/40 text-indigo-600 dark:text-indigo-400 rounded-lg group-hover:scale-105 transition-transform">
    <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" />
    </svg>
  </div>
  <div class="ml-4 flex-1">
    <h3 class="text-base font-semibold text-gray-900 dark:text-white">Storage Builder</h3>
    <p class="text-sm text-gray-600 dark:text-gray-300 mt-1">
      Guided setup for RAID, LVM, pools & shares
    </p>
  </div>
  <svg class="w-5 h-5 text-gray-400 group-hover:text-gray-600 dark:group-hover:text-gray-300 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
  </svg>
</div>
