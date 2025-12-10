<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  const actions = [
    {
      title: 'Create Share',
      description: 'Create new Samba or NFS share',
      icon: 'share',
      color: 'blue',
      href: '/shares/create'
    },
    {
      title: 'Add Disk',
      description: 'Configure new storage disk',
      icon: 'disk',
      color: 'green',
      href: '/storage/add'
    },
    {
      title: 'SCSI Target',
      description: 'Manage iSCSI targets',
      icon: 'target',
      color: 'purple',
      href: '/scsi'
    },
    {
      title: 'System Backup',
      description: 'Configure system backups',
      icon: 'backup',
      color: 'orange',
      href: '/backup'
    }
  ];
  
  function getIcon(iconName) {
    const icons = {
      share: '<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m9.032 4.026a3 3 0 10-4.732 2.684m4.732-2.684a3 3 0 00-4.732-2.684M3 12a3 3 0 104.732 2.684M3 12a3 3 0 014.732-2.684" /></svg>',
      disk: '<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" /></svg>',
      target: '<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>',
      backup: '<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>'
    };
    return icons[iconName] || icons.share;
  }
  
  function getColorClasses(color) {
    const colors = {
      blue: 'bg-blue-100 text-blue-600 hover:bg-blue-200',
      green: 'bg-green-100 text-green-600 hover:bg-green-200',
      purple: 'bg-purple-100 text-purple-600 hover:bg-purple-200',
      orange: 'bg-orange-100 text-orange-600 hover:bg-orange-200'
    };
    return colors[color] || colors.blue;
  }
</script>

<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
  {#each actions as action}
    <a 
      href={action.href}
      class="flex items-center p-4 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 hover:shadow-md transition-all duration-200 group"
    >
      <div class="flex-shrink-0 p-3 {getColorClasses(action.color)} rounded-lg group-hover:scale-105 transition-transform">
        {@html getIcon(action.icon)}
      </div>
      <div class="ml-4">
        <h3 class="text-base font-semibold text-gray-900 dark:text-white">{action.title}</h3>
        <p class="text-sm text-gray-600 dark:text-gray-300 mt-1">{action.description}</p>
      </div>
    </a>
  {/each}
</div>
