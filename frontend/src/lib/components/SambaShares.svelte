<script>
  import { onMount } from 'svelte';
  import { sambaAPI } from '$lib/api.js';
  
  let shares = [];
  let loading = true;
  let error = null;
  let showCreateModal = false;
  let showEditModal = false;
  let selectedShare = null;
  
  // Form data for new share
  let newShare = {
    name: '',
    path: '',
    comment: '',
    users: '',
    groups: '',
    guest_ok: false,
    read_only: false,
    browseable: true
  };
  
  // Form data for editing share
  let editShare = {
    name: '',
    path: '',
    comment: '',
    users: '',
    groups: '',
    guest_ok: false,
    read_only: false,
    browseable: true
  };
  
  async function loadShares() {
    try {
      loading = true;
      shares = await sambaAPI.getShares();
      error = null;
    } catch (err) {
      error = err.message;
      console.error('Failed to load Samba shares:', err);
    } finally {
      loading = false;
    }
  }
  
  onMount(() => {
    loadShares();
  });
  
  function getStatusColor(available) {
    return available ? 'text-green-600 bg-green-100' : 'text-gray-600 dark:text-gray-300 bg-gray-100';
  }
  
  function getUsageColor(percentage) {
    if (percentage >= 90) return 'bg-red-500';
    if (percentage >= 75) return 'bg-yellow-500';
    return 'bg-green-500';
  }
  
  async function toggleShare(share) {
    try {
      await sambaAPI.toggleShare(share.id);
      share.available = !share.available;
      shares = [...shares];
    } catch (err) {
      console.error('Failed to toggle share:', err);
      // Revert the change on error
      share.available = !share.available;
      shares = [...shares];
    }
  }
  
  async function deleteShare(shareId) {
    if (confirm('Are you sure you want to delete this Samba share?')) {
      try {
        await sambaAPI.deleteShare(shareId);
        shares = shares.filter(s => s.id !== shareId);
      } catch (err) {
        console.error('Failed to delete share:', err);
      }
    }
  }
  
  async function createShare() {
    try {
      const shareData = {
        ...newShare,
        users: newShare.users.split(',').map(u => u.trim()).filter(u => u),
        groups: newShare.groups.split(',').map(g => g.trim()).filter(g => g)
      };
      
      const createdShare = await sambaAPI.createShare(shareData);
      shares = [...shares, createdShare];
      showCreateModal = false;
      resetForm();
    } catch (err) {
      console.error('Failed to create share:', err);
    }
  }
  
  function openEditModal(share) {
    selectedShare = share;
    editShare = {
      name: share.name,
      path: share.path,
      comment: share.comment || '',
      users: share.users.join(', '),
      groups: share.groups.join(', '),
      guest_ok: share.guest_ok,
      read_only: share.read_only,
      browseable: share.browseable
    };
    showEditModal = true;
  }
  
  async function updateShare() {
    try {
      const shareData = {
        ...editShare,
        users: editShare.users.split(',').map(u => u.trim()).filter(u => u),
        groups: editShare.groups.split(',').map(g => g.trim()).filter(g => g)
      };
      
      const updatedShare = await sambaAPI.updateShare(selectedShare.name, shareData);
      shares = shares.map(s => s.name === selectedShare.name ? updatedShare : s);
      showEditModal = false;
      selectedShare = null;
    } catch (err) {
      console.error('Failed to update share:', err);
    }
  }
  
  function resetForm() {
    newShare = {
      name: '',
      path: '',
      comment: '',
      users: '',
      groups: '',
      guest_ok: false,
      read_only: false,
      browseable: true
    };
  }
  
  function calculateUsage(used, size) {
    const usedGB = parseFloat(used.replace(/[^\d.]/g, ''));
    const sizeGB = parseFloat(size.replace(/[^\d.]/g, ''));
    return Math.round((usedGB / sizeGB) * 100);
  }
</script>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div>
      <h2 class="text-xl font-bold text-gray-900 dark:text-white">Samba Shares</h2>
      <p class="text-sm text-gray-600 dark:text-gray-300 mt-1">Manage SMB/CIFS file shares</p>
    </div>
    <div class="flex items-center space-x-3">
      <button 
        on:click={loadShares}
        class="btn btn-secondary"
        disabled={loading}
      >
        <svg class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
        Refresh
      </button>
      <button 
        on:click={() => showCreateModal = true}
        class="btn btn-primary"
      >
        <svg class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
        </svg>
        Create Share
      </button>
    </div>
  </div>
  
  <!-- Error State -->
  {#if error}
    <div class="bg-red-50 border border-red-200 rounded-md p-4">
      <div class="flex">
        <div class="flex-shrink-0">
          <svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
          </svg>
        </div>
        <div class="ml-3">
          <h3 class="text-sm font-medium text-red-800">Error loading Samba shares</h3>
          <div class="mt-2 text-sm text-red-700">{error}</div>
        </div>
      </div>
    </div>
  {/if}
  
  {#if loading}
    <div class="flex items-center justify-center py-12">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
    </div>
  {:else if shares.length === 0}
    <div class="text-center py-12">
      <svg class="h-12 w-12 text-gray-400 mx-auto mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m9.032 4.026a3 3 0 10-4.732 2.684m4.732-2.684a3 3 0 00-4.732-2.684M3 12a3 3 0 104.732 2.684M3 12a3 3 0 014.732-2.684" />
      </svg>
      <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">No Samba shares found</h3>
      <p class="text-gray-600 dark:text-gray-300 mb-4">Create your first Samba share to get started.</p>
      <button class="btn btn-primary">Create Share</button>
    </div>
  {:else}
    <!-- Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
      <div class="stat-card">
        <div class="flex items-center">
          <div class="p-3 bg-blue-100 rounded-lg mr-4">
            <svg class="h-6 w-6 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m9.032 4.026a3 3 0 10-4.732 2.684m4.732-2.684a3 3 0 00-4.732-2.684M3 12a3 3 0 104.732 2.684M3 12a3 3 0 014.732-2.684" />
            </svg>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-600 dark:text-gray-300">Total Shares</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{shares.length}</p>
          </div>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="flex items-center">
          <div class="p-3 bg-green-100 rounded-lg mr-4">
            <svg class="h-6 w-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-600 dark:text-gray-300">Active</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{shares.filter(s => s.available).length}</p>
          </div>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="flex items-center">
          <div class="p-3 bg-purple-100 rounded-lg mr-4">
            <svg class="h-6 w-6 text-purple-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-600 dark:text-gray-300">Connections</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{shares.reduce((sum, s) => sum + s.connections, 0)}</p>
          </div>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="flex items-center">
          <div class="p-3 bg-orange-100 rounded-lg mr-4">
            <svg class="h-6 w-6 text-orange-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
            </svg>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-600 dark:text-gray-300">Total Size</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">6.6TB</p>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Shares List -->
    <div class="space-y-4">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Share Details</h3>
      
      {#each shares as share}
        <div class="card hover:shadow-lg transition-shadow">
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="flex items-center space-x-3 mb-3">
                <h4 class="text-lg font-semibold text-gray-900 dark:text-white">{share.name}</h4>
                <span class="px-2 py-1 text-xs font-medium rounded-full {getStatusColor(share.available)}">
                  {share.available ? 'Available' : 'Unavailable'}
                </span>
                {#if share.read_only}
                  <span class="px-2 py-1 text-xs font-medium rounded-full bg-blue-100 text-blue-600">
                    Read-only
                  </span>
                {/if}
              </div>
              
              <p class="text-sm text-gray-600 dark:text-gray-300 mb-4">{share.comment}</p>
              
              <!-- Usage Bar -->
              <div class="mb-4">
                <div class="flex items-center justify-between text-sm mb-2">
                  <span class="text-gray-600 dark:text-gray-300">Usage: {share.used} / {share.size}</span>
                  <span class="font-medium text-gray-900 dark:text-white">{calculateUsage(share.used, share.size)}%</span>
                </div>
                <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3">
                  <div 
                    class="{getUsageColor(calculateUsage(share.used, share.size))} h-3 rounded-full transition-all duration-500" 
                    style="width: {calculateUsage(share.used, share.size)}%"
                  ></div>
                </div>
              </div>
              
              <!-- Details Grid -->
              <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm mb-4">
                <div>
                  <p class="text-gray-600 dark:text-gray-300">Path</p>
                  <p class="font-medium text-gray-900 dark:text-white">{share.path}</p>
                </div>
                <div>
                  <p class="text-gray-600 dark:text-gray-300">Connections</p>
                  <p class="font-medium text-gray-900 dark:text-white">{share.connections}</p>
                </div>
                <div>
                  <p class="text-gray-600 dark:text-gray-300">Guest Access</p>
                  <p class="font-medium text-gray-900 dark:text-white">{share.guest_ok ? 'Allowed' : 'Denied'}</p>
                </div>
                <div>
                  <p class="text-gray-600 dark:text-gray-300">Browseable</p>
                  <p class="font-medium text-gray-900 dark:text-white">{share.browseable ? 'Yes' : 'No'}</p>
                </div>
              </div>
              
              <!-- Users and Groups -->
              <div class="mb-4">
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {#if share.users.length > 0}
                    <div>
                      <p class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Allowed Users:</p>
                      <div class="flex flex-wrap gap-2">
                        {#each share.users as user}
                          <span class="px-2 py-1 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 text-xs rounded-md">
                            {user}
                          </span>
                        {/each}
                      </div>
                    </div>
                  {/if}
                  {#if share.groups.length > 0}
                    <div>
                      <p class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Allowed Groups:</p>
                      <div class="flex flex-wrap gap-2">
                        {#each share.groups as group}
                          <span class="px-2 py-1 bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300 text-xs rounded-md">
                            {group}
                          </span>
                        {/each}
                      </div>
                    </div>
                  {/if}
                </div>
              </div>
              
              <!-- Timestamps -->
              <div class="flex items-center space-x-6 text-xs text-gray-500">
                <span>Created: {share.created}</span>
                <span>Modified: {share.last_modified}</span>
              </div>
            </div>
            
            <!-- Actions -->
            <div class="flex space-x-2 ml-4">
              <button 
                on:click={() => toggleShare(share)}
                class="p-2 text-gray-400 hover:text-gray-600 dark:text-gray-300"
                title={share.available ? 'Disable' : 'Enable'}
              >
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  {#if share.available}
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
                  {:else}
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  {/if}
                </svg>
              </button>
              <button class="p-2 text-gray-400 hover:text-gray-600 dark:text-gray-300" title="Edit" on:click={() => openEditModal(share)}>
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
              </button>
              <button 
                on:click={() => deleteShare(share.id)}
                class="p-2 text-gray-400 hover:text-red-600" 
                title="Delete"
              >
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Create Share Modal -->
{#if showCreateModal}
  <div class="fixed inset-0 z-50 overflow-y-auto">
    <div class="flex items-center justify-center min-h-screen px-4">
      <div class="fixed inset-0 bg-gray-500 bg-opacity-75" on:click={() => { showCreateModal = false; resetForm(); }} on:keydown={(e) => e.key === 'Escape' && (showCreateModal = false)} role="button" tabindex="0" aria-label="Close modal"></div>
      
      <div class="relative bg-white dark:bg-gray-800 rounded-lg max-w-2xl w-full p-6">
        <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Create Samba Share</h3>
        
        <form on:submit|preventDefault={createShare} class="space-y-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label for="share-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Share Name</label>
              <input 
                id="share-name"
                type="text" 
                bind:value={newShare.name}
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
                placeholder="media"
                required
              />
            </div>
            
            <div>
              <label for="share-path" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Path</label>
              <input 
                id="share-path"
                type="text" 
                bind:value={newShare.path}
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
                placeholder="/data/media"
                required
              />
            </div>
          </div>
          
          <div>
            <label for="share-comment" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Comment</label>
            <input 
              id="share-comment"
              type="text" 
              bind:value={newShare.comment}
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
              placeholder="Media files and movies"
            />
          </div>
          
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label for="valid-users" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Valid Users</label>
              <input 
                id="valid-users"
                type="text" 
                bind:value={newShare.users}
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
                placeholder="admin, media_user"
              />
            </div>
            
            <div>
              <label for="valid-groups" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Valid Groups</label>
              <input 
                id="valid-groups"
                type="text" 
                bind:value={newShare.groups}
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
                placeholder="media_group"
              />
            </div>
          </div>
          
          <div class="space-y-3">
            <label class="flex items-center">
              <input type="checkbox" bind:checked={newShare.read_only} class="mr-2 dark:bg-gray-700 dark:border-gray-600" />
              <span class="text-sm text-gray-700 dark:text-gray-300">Read only</span>
            </label>
            
            <label class="flex items-center">
              <input type="checkbox" bind:checked={newShare.guest_ok} class="mr-2 dark:bg-gray-700 dark:border-gray-600" />
              <span class="text-sm text-gray-700 dark:text-gray-300">Allow guest access</span>
            </label>
            
            <label class="flex items-center">
              <input type="checkbox" bind:checked={newShare.browseable} class="mr-2 dark:bg-gray-700 dark:border-gray-600" />
              <span class="text-sm text-gray-700 dark:text-gray-300">Browseable</span>
            </label>
          </div>
          
          <div class="flex justify-end space-x-3 pt-4">
            <button 
              type="button" 
              on:click={() => showCreateModal = false}
              class="btn btn-secondary"
            >
              Cancel
            </button>
            <button type="submit" class="btn btn-primary">
              Create Share
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
{/if}

<!-- Edit Share Modal -->
{#if showEditModal}
  <div class="fixed inset-0 z-50 overflow-y-auto">
    <div class="flex items-center justify-center min-h-screen px-4">
      <div class="fixed inset-0 bg-gray-500 bg-opacity-75" on:click={() => { showEditModal = false; selectedShare = null; }} on:keydown={(e) => e.key === 'Escape' && (showEditModal = false)} role="button" tabindex="0" aria-label="Close modal"></div>
      
      <div class="relative bg-white dark:bg-gray-800 rounded-lg max-w-2xl w-full p-6">
        <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Edit Samba Share</h3>
        
        <form on:submit|preventDefault={updateShare} class="space-y-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label for="edit-share-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Share Name</label>
              <input 
                id="edit-share-name"
                type="text" 
                bind:value={editShare.name}
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
                placeholder="media"
                required
              />
            </div>
            
            <div>
              <label for="edit-share-path" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Path</label>
              <input 
                id="edit-share-path"
                type="text" 
                bind:value={editShare.path}
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
                placeholder="/data/media"
                required
              />
            </div>
          </div>
          
          <div>
            <label for="edit-share-comment" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Comment</label>
            <input 
              id="edit-share-comment"
              type="text" 
              bind:value={editShare.comment}
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
              placeholder="Media files and movies"
            />
          </div>
          
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label for="edit-valid-users" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Valid Users</label>
              <input 
                id="edit-valid-users"
                type="text" 
                bind:value={editShare.users}
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
                placeholder="admin, media_user"
              />
            </div>
            
            <div>
              <label for="edit-valid-groups" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Valid Groups</label>
              <input 
                id="edit-valid-groups"
                type="text" 
                bind:value={editShare.groups}
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400 text-gray-900 dark:text-white"
                placeholder="media_group"
              />
            </div>
          </div>
          
          <div class="space-y-3">
            <label class="flex items-center">
              <input type="checkbox" bind:checked={editShare.read_only} class="mr-2 dark:bg-gray-700 dark:border-gray-600" />
              <span class="text-sm text-gray-700 dark:text-gray-300">Read only</span>
            </label>
            
            <label class="flex items-center">
              <input type="checkbox" bind:checked={editShare.guest_ok} class="mr-2 dark:bg-gray-700 dark:border-gray-600" />
              <span class="text-sm text-gray-700 dark:text-gray-300">Allow guest access</span>
            </label>
            
            <label class="flex items-center">
              <input type="checkbox" bind:checked={editShare.browseable} class="mr-2 dark:bg-gray-700 dark:border-gray-600" />
              <span class="text-sm text-gray-700 dark:text-gray-300">Browseable</span>
            </label>
          </div>
          
          <div class="flex justify-end space-x-3 pt-4">
            <button 
              type="button" 
              on:click={() => { showEditModal = false; selectedShare = null; }}
              class="btn btn-secondary"
            >
              Cancel
            </button>
            <button type="submit" class="btn btn-primary">
              Update Share
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
{/if}
