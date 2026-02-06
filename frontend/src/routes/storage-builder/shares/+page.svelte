<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { storageAPI } from '$lib/api.js';

  let storagePools = [];
  let logicalVolumes = [];
  let loading = true;
  let error = null;
  let creatingNFS = false;
  let creatingSamba = false;

  // NFS configuration
  let nfsConfig = {
    path: '',
    clients: [{ network: '192.168.1.0/24', options: 'rw,sync,no_subtree_check' }]
  };

  // Samba configuration
  let sambaConfig = {
    name: '',
    path: '',
    comment: '',
    readOnly: false,
    browseable: true
  };

  async function loadAvailableStorage() {
    try {
      loading = true;

      // Load storage pools
      const pools = await storageAPI.getPools();
      storagePools = Array.isArray(pools) ? pools : [];

      // For now, we'll just use pools. LVM LVs would need to be loaded separately
      // and mounted first before they can be shared.
    } catch (err) {
      error = err.message || 'Failed to load available storage';
      console.error('Failed to load available storage:', err);
    } finally {
      loading = false;
    }
  }

  function addNFSClient() {
    nfsConfig.clients = [...nfsConfig.clients, { network: '', options: 'rw,sync,no_subtree_check' }];
  }

  function removeNFSClient(index) {
    if (nfsConfig.clients.length > 1) {
      nfsConfig.clients = nfsConfig.clients.filter((_, i) => i !== index);
    }
  }

  function selectStoragePath(path) {
    nfsConfig.path = path;
    sambaConfig.path = path;
  }

  function cancel() {
    goto('/storage-builder');
  }

  async function createNFSExport() {
    try {
      if (!nfsConfig.path) {
        error = 'Please select a storage path';
        return;
      }

      if (nfsConfig.clients.length === 0 || nfsConfig.clients.some(c => !c.network)) {
        error = 'Please configure at least one NFS client';
        return;
      }

      creatingNFS = true;
      // await nfsAPI.createExport(nfsConfig);

      // Success - redirect to NFS page
      goto('/nfs');
    } catch (err) {
      error = err.message || 'Failed to create NFS export';
      console.error('Failed to create NFS export:', err);
    } finally {
      creatingNFS = false;
    }
  }

  async function createSambaShare() {
    try {
      if (!sambaConfig.name) {
        error = 'Please enter a share name';
        return;
      }

      if (!sambaConfig.path) {
        error = 'Please select a storage path';
        return;
      }

      creatingSamba = true;
      // await sambaAPI.createShare(sambaConfig);

      // Success - redirect to Samba page
      goto('/samba');
    } catch (err) {
      error = err.message || 'Failed to create Samba share';
      console.error('Failed to create Samba share:', err);
    } finally {
      creatingSamba = false;
    }
  }

  onMount(() => {
    loadAvailableStorage();
  });
</script>

<div class="p-6" role="main" tabindex="-1">
  <!-- Header -->
  <div class="mb-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-gray-900 dark:text-white mb-2">
          Create Network Shares
        </h1>
        <p class="text-sm text-gray-600 dark:text-gray-300">
          Create NFS or Samba shares from existing storage
        </p>
      </div>
      <button
        on:click={cancel}
        class="text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
        aria-label="Cancel"
        title="Cancel"
      >
        <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
  </div>

  <!-- Error -->
  {#if error}
    <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4 mb-6">
      <p class="text-red-800 dark:text-red-200">{error}</p>
    </div>
  {/if}

  <!-- Loading -->
  {#if loading}
    <div class="flex justify-center items-center py-12">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
    </div>
  {:else}
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- NFS Export -->
      <div class="bg-white dark:bg-card rounded-lg border border-gray-200 dark:border-border p-6">
        <div class="flex items-center space-x-3 mb-4">
          <div class="p-2 bg-orange-100 dark:bg-orange-900/40 rounded-lg">
            <svg class="w-6 h-6 text-orange-600 dark:text-orange-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
            </svg>
          </div>
          <div>
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">NFS Export</h2>
            <p class="text-sm text-gray-600 dark:text-gray-400">Unix/Linux network file sharing</p>
          </div>
        </div>

        <!-- Select Storage Path -->
        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Select Storage Path
          </label>
          {#if storagePools.length === 0}
            <div class="p-3 bg-gray-50 dark:bg-muted border border-gray-200 dark:border rounded-md text-sm text-gray-500 dark:text-gray-400 text-center">
              No storage pools available. Create one first.
            </div>
          {:else}
            <select
              class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-muted dark:text-white"
              on:change={(e) => selectStoragePath(e.target.value)}
            >
              <option value="">-- Select a pool --</option>
              {#each storagePools as pool}
                <option value={pool.mount_point}>
                  {pool.name} ({pool.mount_point})
                </option>
              {/each}
            </select>
          {/if}
        </div>

        <!-- NFS Clients -->
        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Client Access Rules
          </label>
          <div class="space-y-2">
            {#each nfsConfig.clients as client, index}
              <div class="flex space-x-2">
                <input
                  type="text"
                  bind:value={client.network}
                  placeholder="192.168.1.0/24"
                  class="flex-1 px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-muted dark:text-white text-sm"
                />
                <input
                  type="text"
                  bind:value={client.options}
                  placeholder="rw,sync,no_subtree_check"
                  class="flex-1 px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-muted dark:text-white text-sm"
                />
                {#if nfsConfig.clients.length > 1}
                  <button
                    type="button"
                    on:click={() => removeNFSClient(index)}
                    class="px-2 py-2 text-red-600 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300"
                    title="Remove client"
                  >
                    <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                {/if}
              </div>
            {/each}
          </div>
          <button
            type="button"
            on:click={addNFSClient}
            class="mt-2 text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300"
          >
            + Add another client
          </button>
        </div>

        <!-- Create NFS Export Button -->
        <button
          type="button"
          on:click={createNFSExport}
          disabled={creatingNFS || !nfsConfig.path}
          class="w-full px-4 py-2 bg-orange-600 text-white rounded-md shadow-sm hover:bg-orange-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-orange-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {#if creatingNFS}
            Creating...
          {:else}
            Create NFS Export
          {/if}
        </button>
      </div>

      <!-- Samba Share -->
      <div class="bg-white dark:bg-card rounded-lg border border-gray-200 dark:border-border p-6">
        <div class="flex items-center space-x-3 mb-4">
          <div class="p-2 bg-blue-100 dark:bg-blue-900/40 rounded-lg">
            <svg class="w-6 h-6 text-blue-600 dark:text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
            </svg>
          </div>
          <div>
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">Samba Share</h2>
            <p class="text-sm text-gray-600 dark:text-gray-400">Windows/SMB network file sharing</p>
          </div>
        </div>

        <!-- Select Storage Path -->
        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Select Storage Path
          </label>
          {#if storagePools.length === 0}
            <div class="p-3 bg-gray-50 dark:bg-muted border border-gray-200 dark:border rounded-md text-sm text-gray-500 dark:text-gray-400 text-center">
              No storage pools available. Create one first.
            </div>
          {:else}
            <select
              class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-muted dark:text-white"
              on:change={(e) => sambaConfig.path = e.target.value}
            >
              <option value="">-- Select a pool --</option>
              {#each storagePools as pool}
                <option value={pool.mount_point}>
                  {pool.name} ({pool.mount_point})
                </option>
              {/each}
            </select>
          {/if}
        </div>

        <!-- Share Name -->
        <div class="mb-4">
          <label for="sambaName" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Share Name
          </label>
          <input
            id="sambaName"
            type="text"
            bind:value={sambaConfig.name}
            placeholder="my-share"
            class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-muted dark:text-white"
          />
        </div>

        <!-- Description (optional) -->
        <div class="mb-4">
          <label for="sambaComment" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Description (optional)
          </label>
          <input
            id="sambaComment"
            type="text"
            bind:value={sambaConfig.comment}
            placeholder="My network share"
            class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-muted dark:text-white"
          />
        </div>

        <!-- Options -->
        <div class="mb-4 space-y-2">
          <label class="flex items-center">
            <input
              type="checkbox"
              bind:checked={sambaConfig.readOnly}
              class="rounded border-gray-300 dark:border text-blue-600 focus:ring-blue-500 dark:bg-card"
            />
            <span class="ml-2 text-sm text-gray-700 dark:text-gray-300">Read-only</span>
          </label>
          <label class="flex items-center">
            <input
              type="checkbox"
              bind:checked={sambaConfig.browseable}
              class="rounded border-gray-300 dark:border text-blue-600 focus:ring-blue-500 dark:bg-card"
            />
            <span class="ml-2 text-sm text-gray-700 dark:text-gray-300">Browseable (visible in network)</span>
          </label>
        </div>

        <!-- Create Samba Share Button -->
        <button
          type="button"
          on:click={createSambaShare}
          disabled={creatingSamba || !sambaConfig.name || !sambaConfig.path}
          class="w-full px-4 py-2 bg-blue-600 text-white rounded-md shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {#if creatingSamba}
            Creating...
          {:else}
            Create Samba Share
          {/if}
        </button>
      </div>
    </div>

    <!-- Info Note -->
    <div class="mt-6 p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg">
      <div class="flex">
        <svg class="w-5 h-5 text-blue-400 mr-2 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
        </svg>
        <div class="text-sm text-blue-800 dark:text-blue-200">
          <p class="font-medium mb-1">Need more options?</p>
          <p>You can also configure advanced NFS and Samba settings from their respective pages in the Sharing section.</p>
        </div>
      </div>
    </div>
  {/if}
</div>
