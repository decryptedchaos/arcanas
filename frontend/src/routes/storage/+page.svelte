<script>
  import Storage from '$lib/components/Storage.svelte';
  import RAIDManager from '$lib/components/RAIDManager.svelte';
  import StoragePools from '$lib/components/StoragePools.svelte';
  
  let activeTab = 'disks';
</script>

<div class="space-y-6">
  <!-- Tab Navigation -->
  <div class="border-b border-gray-200 dark:border-gray-700">
    <nav class="-mb-px flex space-x-8">
      <button
        class="py-2 px-1 border-b-2 font-medium text-sm {activeTab === 'disks' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
        on:click={() => activeTab = 'disks'}
      >
        Disks
      </button>
      <button
        class="py-2 px-1 border-b-2 font-medium text-sm {activeTab === 'raid' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
        on:click={() => activeTab = 'raid'}
      >
        RAID Arrays
      </button>
      <button
        class="py-2 px-1 border-b-2 font-medium text-sm {activeTab === 'pools' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
        on:click={() => activeTab = 'pools'}
      >
        Storage Pools
      </button>
    </nav>
  </div>

  <!-- Tab Content -->
  {#if activeTab === 'disks'}
    <Storage />
  {:else if activeTab === 'raid'}
    <RAIDManager />
  {:else if activeTab === 'pools'}
    <StoragePools />
  {/if}
</div>
