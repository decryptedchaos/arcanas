<script>
  import { page } from '$app/stores';
  import { createEventDispatcher } from 'svelte';
  
  export let sidebarOpen = true;
  
  const dispatch = createEventDispatcher();
  
  let servicesExpanded = false;
  
  const navigation = [
    { name: 'Dashboard', href: '/', icon: 'home' },
    { name: 'Storage', href: '/storage', icon: 'disk' },
    { name: 'Sharing', href: '#', icon: 'share', isExpandable: true },
    { name: 'System Stats', href: '/system', icon: 'cpu' },
    { name: 'SMART Status', href: '/smart', icon: 'health' },
    { name: 'Settings', href: '/settings', icon: 'settings' }
  ];
  
  const services = [
    { name: 'SCSI Targets', href: '/scsi', icon: 'target' },
    { name: 'Samba Shares', href: '/samba', icon: 'share' },
    { name: 'NFS Exports', href: '/nfs', icon: 'network' }
  ];
  
  function getIcon(iconName) {
    const icons = {
      home: '<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" /></svg>',
      disk: '<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" /></svg>',
      target: '<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>',
      share: '<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m9.032 4.026a3 3 0 10-4.732 2.684m4.732-2.684a3 3 0 00-4.732-2.684M3 12a3 3 0 104.732 2.684M3 12a3 3 0 014.732-2.684" /></svg>',
      network: '<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" /></svg>',
      cpu: '<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" /></svg>',
      health: '<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>',
      settings: '<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /></svg>',
      chevron: '<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>'
    };
    return icons[iconName] || icons.home;
  }
</script>

<!-- Mobile sidebar overlay -->
{#if sidebarOpen}
  <div 
    class="fixed inset-0 z-40 bg-gray-600 bg-opacity-75 md:hidden"
    on:click={() => sidebarOpen = false}
    on:keydown={(e) => e.key === 'Escape' && (sidebarOpen = false)}
    role="button"
    tabindex="0"
    aria-label="Close sidebar"
  ></div>
{/if}

<!-- Sidebar -->
<div class="fixed inset-y-0 left-0 z-50 w-64 bg-white dark:bg-gray-800 shadow-lg transform transition-transform duration-300 ease-in-out md:translate-x-0 md:static md:inset-0 {sidebarOpen ? 'translate-x-0' : '-translate-x-full'}">
  <div class="flex flex-col h-full">
    <!-- Logo -->
    <div class="flex items-center h-16 px-6 border-b border-gray-200 dark:border-gray-700">
      <div class="flex items-center">
        <div class="flex-shrink-0">
          <div class="w-8 h-8 bg-primary-600 rounded-lg flex items-center justify-center">
            <span class="text-white font-bold text-lg">A</span>
          </div>
        </div>
        <div class="ml-3">
          <p class="text-sm font-medium text-gray-900 dark:text-white">Arcanas Manager</p>
          <p class="text-xs text-gray-500 dark:text-gray-400">v1.0.0</p>
        </div>
      </div>
    </div>
    
    <!-- Navigation -->
    <nav class="flex-1 px-4 py-6 space-y-2">
      {#each navigation as item}
        {#if item.isExpandable}
          <!-- Sharing Expandable Menu -->
          <div>
            <button
              on:click={() => servicesExpanded = !servicesExpanded}
              class="w-full group flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors duration-200 text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-white"
            >
              <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
              </svg>
              <span class="ml-3">{item.name}</span>
              <span class="ml-auto transition-transform duration-200 {servicesExpanded ? 'rotate-90' : ''}">
                {@html getIcon('chevron')}
              </span>
            </button>
            
            {#if servicesExpanded}
              <div class="ml-6 mt-1 space-y-1">
                {#each services as service}
                  <a
                    href={service.href}
                    class="group flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors duration-200 {
                      $page.url.pathname === service.href 
                        ? 'bg-primary-50 dark:bg-primary-900/20 text-primary-700 dark:text-primary-300 border-r-2 border-primary-700 dark:border-primary-400' 
                        : 'text-gray-500 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 hover:text-gray-700 dark:hover:text-gray-200'
                    }"
                  >
                    {@html getIcon(service.icon)}
                    <span class="ml-3">{service.name}</span>
                  </a>
                {/each}
              </div>
            {/if}
          </div>
        {:else}
          <!-- Regular Navigation Item -->
          <a
            href={item.href}
            class="group flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors duration-200 {
              $page.url.pathname === item.href 
                ? 'bg-primary-50 dark:bg-primary-900/20 text-primary-700 dark:text-primary-300 border-r-2 border-primary-700 dark:border-primary-400' 
                : 'text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-white'
            }"
          >
            {@html getIcon(item.icon)}
            <span class="ml-3">{item.name}</span>
          </a>
        {/if}
      {/each}
    </nav>
  </div>
</div>
