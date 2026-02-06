<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { createEventDispatcher } from 'svelte';

  export let poolConfig = {
    name: '',
    mountPoint: '/srv/storage'
  };
  export let shareConfig = {
    createNFS: false,
    createSamba: false,
    nfsClients: [],
    sambaName: '',
    sambaUsers: []
  };

  const dispatch = createEventDispatcher();

  // Add initial NFS client
  if (shareConfig.nfsClients.length === 0) {
    shareConfig.nfsClients = [{ network: '192.168.1.0/24', options: 'rw,sync,no_subtree_check' }];
  }

  function toggleNFS(enabled) {
    shareConfig = { ...shareConfig, createNFS: enabled };
    dispatch('shareChange', shareConfig);
  }

  function toggleSamba(enabled) {
    shareConfig = { ...shareConfig, createSamba: enabled };
    if (enabled && !shareConfig.sambaName) {
      shareConfig.sambaName = poolConfig.name || 'storage';
    }
    dispatch('shareChange', shareConfig);
  }

  function addNFSClient() {
    shareConfig = {
      ...shareConfig,
      nfsClients: [
        ...shareConfig.nfsClients,
        { network: '', options: 'rw,sync,no_subtree_check' }
      ]
    };
  }

  function removeNFSClient(index) {
    shareConfig = {
      ...shareConfig,
      nfsClients: shareConfig.nfsClients.filter((_, i) => i !== index)
    };
  }

  function updateNFSClient(index, field, value) {
    const newClients = [...shareConfig.nfsClients];
    newClients[index] = { ...newClients[index], [field]: value };
    shareConfig = { ...shareConfig, nfsClients: newClients };
  }

  function handleFinish() {
    dispatch('finish');
  }

  function handlePrevious() {
    dispatch('previous');
  }

  function handleSkip() {
    dispatch('finish');
  }

  $: mountPoint = poolConfig.mountPoint || `/srv/${poolConfig.name || 'storage'}`;
  $: hasShares = shareConfig.createNFS || shareConfig.createSamba;
</script>

<div class="bg-white dark:bg-card rounded-lg shadow-lg overflow-hidden">
  <!-- Step Header -->
  <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-xl font-semibold text-gray-900 dark:text-white">
          Network Shares (Optional)
        </h2>
        <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
          Optionally create network shares for your storage pool
        </p>
      </div>
      <div class="text-sm text-gray-600 dark:text-gray-400">
        Pool: {poolConfig.name || 'storage'}
      </div>
    </div>
  </div>

  <div class="p-6 space-y-6">
    <!-- NFS Exports -->
    <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
      <div class="flex items-center justify-between mb-4">
        <div class="flex items-center space-x-3">
          <svg class="w-6 h-6 text-gray-600 dark:text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
          </svg>
          <div>
            <h3 class="text-lg font-medium text-gray-900 dark:text-white">NFS Export</h3>
            <p class="text-sm text-gray-600 dark:text-gray-400">Share via NFS for Linux/Unix systems</p>
          </div>
        </div>
        <button
          on:click={() => toggleNFS(!shareConfig.createNFS)}
          type="button"
          class="relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2
            {shareConfig.createNFS ? 'bg-indigo-600' : 'bg-gray-200 dark:bg-gray-700'}"
          role="switch"
          aria-checked={shareConfig.createNFS}
        >
          <span
            aria-hidden="true"
            class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out
              {shareConfig.createNFS ? 'translate-x-5' : 'translate-x-0'}"
          ></span>
        </button>
      </div>

      {#if shareConfig.createNFS}
        <div class="space-y-3">
          <div class="bg-gray-50 dark:bg-gray-800 rounded px-3 py-2 text-sm">
            <span class="text-gray-600 dark:text-gray-400">Export path: </span>
            <span class="font-mono text-gray-900 dark:text-white">{mountPoint}</span>
          </div>

          <div class="space-y-2">
            <h4 class="text-sm font-medium text-gray-900 dark:text-white">Allowed Clients</h4>
            {#each shareConfig.nfsClients as client, index}
              <div class="flex items-center space-x-2">
                <input
                  type="text"
                  bind:value={client.network}
                  on:input={(e) => updateNFSClient(index, 'network', e.target.value)}
                  placeholder="192.168.1.0/24"
                  class="flex-1 px-3 py-2 border border-gray-300 dark:border rounded-md dark:bg-gray-800 dark:text-white text-sm"
                />
                <input
                  type="text"
                  bind:value={client.options}
                  on:input={(e) => updateNFSClient(index, 'options', e.target.value)}
                  placeholder="rw,sync,no_subtree_check"
                  class="flex-1 px-3 py-2 border border-gray-300 dark:border rounded-md dark:bg-gray-800 dark:text-white text-sm"
                />
                {#if shareConfig.nfsClients.length > 1}
                  <button
                    on:click={() => removeNFSClient(index)}
                    type="button"
                    class="p-2 text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                {/if}
              </div>
            {/each}

            <button
              on:click={addNFSClient}
              type="button"
              class="text-sm text-indigo-600 dark:text-indigo-400 hover:text-indigo-800 dark:hover:text-indigo-300"
            >
              + Add another client
            </button>
          </div>
        </div>
      {/if}
    </div>

    <!-- Samba Share -->
    <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
      <div class="flex items-center justify-between mb-4">
        <div class="flex items-center space-x-3">
          <svg class="w-6 h-6 text-gray-600 dark:text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
          </svg>
          <div>
            <h3 class="text-lg font-medium text-gray-900 dark:text-white">Samba Share</h3>
            <p class="text-sm text-gray-600 dark:text-gray-400">Share via SMB for Windows/Mac systems</p>
          </div>
        </div>
        <button
          on:click={() => toggleSamba(!shareConfig.createSamba)}
          type="button"
          class="relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2
            {shareConfig.createSamba ? 'bg-indigo-600' : 'bg-gray-200 dark:bg-gray-700'}"
          role="switch"
          aria-checked={shareConfig.createSamba}
        >
          <span
            aria-hidden="true"
            class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out
              {shareConfig.createSamba ? 'translate-x-5' : 'translate-x-0'}"
          ></span>
        </button>
      </div>

      {#if shareConfig.createSamba}
        <div class="space-y-3">
          <div class="bg-gray-50 dark:bg-gray-800 rounded px-3 py-2 text-sm">
            <span class="text-gray-600 dark:text-gray-400">Share path: </span>
            <span class="font-mono text-gray-900 dark:text-white">{mountPoint}</span>
          </div>

          <div>
            <label for="samba-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Share Name
            </label>
            <input
              id="samba-name"
              type="text"
              bind:value={shareConfig.sambaName}
              class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md dark:bg-gray-800 dark:text-white"
              placeholder="e.g., storage"
            />
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              This is the name users will see when browsing the network
            </p>
          </div>
        </div>
      {/if}
    </div>

    <!-- Skip Information -->
    <div class="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
      <div class="flex items-start space-x-3">
        <svg class="w-5 h-5 text-blue-600 dark:text-blue-400 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <div class="flex-1 text-sm text-blue-800 dark:text-blue-200">
          <p class="font-medium mb-1">You can always add shares later</p>
          <p class="text-blue-700 dark:text-blue-300">
            If you're not sure, skip this step. You can create NFS exports and Samba shares
            anytime from the Sharing menu after setup.
          </p>
        </div>
      </div>
    </div>
  </div>

  <!-- Actions -->
  <div class="px-6 py-4 bg-gray-50 dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 flex items-center justify-between">
    <button
      on:click={handlePrevious}
      class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"
    >
      ← Previous
    </button>

    <div class="flex items-center space-x-3">
      <button
        on:click={handleSkip}
        class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"
      >
        Skip Shares
      </button>

      <button
        on:click={handleFinish}
        class="px-6 py-2 text-sm font-medium text-white bg-green-600 rounded-md hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500"
      >
        Create Storage →
      </button>
    </div>
  </div>
</div>
