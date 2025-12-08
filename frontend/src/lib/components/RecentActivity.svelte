<script>
  let activities = [
    {
      id: 1,
      type: 'success',
      title: 'Samba share created',
      description: 'Share "media" created successfully',
      time: '2 minutes ago',
      icon: 'share'
    },
    {
      id: 2,
      type: 'warning',
      title: 'Disk space low',
      description: 'Disk /dev/sdb1 at 85% capacity',
      time: '15 minutes ago',
      icon: 'disk'
    },
    {
      id: 3,
      type: 'info',
      title: 'System backup completed',
      description: 'Daily backup finished successfully',
      time: '1 hour ago',
      icon: 'backup'
    },
    {
      id: 4,
      type: 'error',
      title: 'SCSI target disconnected',
      description: 'Target iqn.2024-01.com.nas:target1 lost connection',
      time: '2 hours ago',
      icon: 'target'
    },
    {
      id: 5,
      type: 'success',
      title: 'NFS export added',
      description: 'Export "/data/backups" added for 192.168.1.0/24',
      time: '3 hours ago',
      icon: 'network'
    }
  ];
  
  function getTypeClasses(type) {
    const types = {
      success: 'bg-green-100 text-green-600 border-green-200',
      warning: 'bg-yellow-100 text-yellow-600 border-yellow-200',
      error: 'bg-red-100 text-red-600 border-red-200',
      info: 'bg-blue-100 text-blue-600 border-blue-200'
    };
    return types[type] || types.info;
  }
  
  function getIcon(iconName) {
    const icons = {
      share: '<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m9.032 4.026a3 3 0 10-4.732 2.684m4.732-2.684a3 3 0 00-4.732-2.684M3 12a3 3 0 104.732 2.684M3 12a3 3 0 014.732-2.684" /></svg>',
      disk: '<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" /></svg>',
      backup: '<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>',
      target: '<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>',
      network: '<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" /></svg>'
    };
    return icons[iconName] || icons.info;
  }
</script>

<div class="card">
  <div class="flex items-center justify-between mb-4">
    <h3 class="text-lg font-semibold text-gray-900 dark:text-white dark:text-white">Recent Activity</h3>
    <button class="text-sm text-primary-600 hover:text-primary-700 font-medium">
      View All
    </button>
  </div>
  
  <div class="space-y-3">
    {#each activities as activity}
      <div class="flex items-start space-x-3 p-3 rounded-lg border border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
        <div class="flex-shrink-0 p-2 {getTypeClasses(activity.type)} rounded-lg border">
          {@html getIcon(activity.icon)}
        </div>
        <div class="flex-1 min-w-0">
          <p class="text-sm font-medium text-gray-900 dark:text-white">{activity.title}</p>
          <p class="text-sm text-gray-600 dark:text-gray-300 mt-1">{activity.description}</p>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-2">{activity.time}</p>
        </div>
      </div>
    {/each}
  </div>
</div>
