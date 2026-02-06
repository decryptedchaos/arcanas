<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { goto } from '$app/navigation';

  const workflows = [
    {
      id: 'full',
      title: 'Complete Storage Setup',
      description: 'Guided wizard for disks, RAID, LVM, pools, and shares',
      icon: 'wand',
      href: '/storage-builder/full',
      color: 'indigo',
      featured: true
    },
    {
      id: 'raid',
      title: 'RAID Array',
      description: 'Create RAID from physical disks',
      icon: 'shield',
      href: '/storage-builder/raid',
      color: 'blue'
    },
    {
      id: 'vg',
      title: 'Volume Group',
      description: 'Create LVM volume group from devices',
      icon: 'cube',
      href: '/storage-builder/vg',
      color: 'purple'
    },
    {
      id: 'lvm',
      title: 'Logical Volume',
      description: 'Create LVM logical volume in existing VG',
      icon: 'layers',
      href: '/storage-builder/lvm',
      color: 'cyan'
    },
    {
      id: 'pool',
      title: 'Storage Pool',
      description: 'Create pool from device, RAID, or LVM',
      icon: 'disk',
      href: '/storage-builder/pool',
      color: 'green'
    },
    {
      id: 'shares',
      title: 'Network Shares',
      description: 'NFS or Samba from existing storage',
      icon: 'share',
      href: '/storage-builder/shares',
      color: 'orange'
    }
  ];

  function getIcon(iconName) {
    const icons = {
      wand: '<svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" /></svg>',
      shield: '<svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" /></svg>',
      cube: '<svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" /></svg>',
      layers: '<svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" /></svg>',
      disk: '<svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" /></svg>',
      share: '<svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" /></svg>'
    };
    return icons[iconName] || icons.wand;
  }

  function getColorClasses(color, featured = false) {
    const colors = {
      indigo: featured
        ? 'bg-indigo-100 dark:bg-indigo-900/40 text-indigo-600 dark:text-indigo-400 hover:bg-indigo-200 dark:hover:bg-indigo-900/60 border-indigo-200 dark:border-indigo-700'
        : 'bg-indigo-100 dark:bg-indigo-900/40 text-indigo-600 dark:text-indigo-400 hover:bg-indigo-200 dark:hover:bg-indigo-900/60',
      blue: 'bg-blue-100 dark:bg-blue-900/40 text-blue-600 dark:text-blue-400 hover:bg-blue-200 dark:hover:bg-blue-900/60',
      purple: 'bg-purple-100 dark:bg-purple-900/40 text-purple-600 dark:text-purple-400 hover:bg-purple-200 dark:hover:bg-purple-900/60',
      cyan: 'bg-cyan-100 dark:bg-cyan-900/40 text-cyan-600 dark:text-cyan-400 hover:bg-cyan-200 dark:hover:bg-cyan-900/60',
      green: 'bg-green-100 dark:bg-green-900/40 text-green-600 dark:text-green-400 hover:bg-green-200 dark:hover:bg-green-900/60',
      orange: 'bg-orange-100 dark:bg-orange-900/40 text-orange-600 dark:text-orange-400 hover:bg-orange-200 dark:hover:bg-orange-900/60'
    };
    return colors[color] || colors.indigo;
  }

  function getCardClasses(featured = false) {
    return featured
      ? 'border-2 border-indigo-300 dark:border-indigo-600 shadow-lg hover:shadow-xl'
      : 'border border-gray-200 dark:border-border hover:shadow-md';
  }
</script>

<div class="space-y-6">
  <div class="text-center">
    <h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">
      What would you like to create?
    </h2>
    <p class="text-gray-600 dark:text-gray-400">
      Choose a workflow or use the complete guided wizard
    </p>
  </div>

  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
    {#each workflows as workflow}
      <a
        href={workflow.href}
        class="flex items-start p-5 bg-white dark:bg-card rounded-lg {getCardClasses(workflow.featured)} transition-all duration-200 group"
        class:ring-2={workflow.featured}
        class:ring-indigo-300={workflow.featured}
        class:dark:ring-indigo-700={workflow.featured}
      >
        <div class="flex-shrink-0 p-3 {getColorClasses(workflow.color, workflow.featured)} rounded-lg group-hover:scale-105 transition-transform">
          {@html getIcon(workflow.icon)}
        </div>
        <div class="ml-4 flex-1">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
              {workflow.title}
            </h3>
            {#if workflow.featured}
              <span class="px-2 py-1 text-xs font-medium bg-indigo-100 dark:bg-indigo-900/40 text-indigo-600 dark:text-indigo-400 rounded-full">
                Recommended
              </span>
            {/if}
          </div>
          <p class="text-sm text-gray-600 dark:text-gray-300 mt-1">
            {workflow.description}
          </p>
          <div class="mt-3 flex items-center text-sm text-indigo-600 dark:text-indigo-400 group-hover:text-indigo-700 dark:group-hover:text-indigo-300">
            <span>Start workflow</span>
            <svg class="ml-1 w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </div>
        </div>
      </a>
    {/each}
  </div>

  <div class="text-center">
    <p class="text-sm text-gray-500 dark:text-gray-400">
      You can also configure each component individually from their respective pages
    </p>
  </div>
</div>
