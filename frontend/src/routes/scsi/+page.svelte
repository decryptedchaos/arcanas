<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { onMount } from "svelte";
  import { formatBytes } from "$lib/utils/byteUtils.js";

  // State
  let target = null;
  let luns = [];
  let backends = [];
  let acls = [];
  let loading = true;
  let error = null;
  let showCreateModal = false;
  let creatingLUN = false;

  // ACL management state
  let showACLModal = false;
  let showMapLUNModal = false;
  let creatingACL = false;
  let selectedACLForMapping = null;
  let aclsExpanded = false;
  let newACL = {
    initiator_iqn: "",
    name: ""
  };
  let newLUNMapping = {
    source_lun: 0,
    target_lun: 0
  };

  // Quick create VG modal state
  let showCreateVGModal = false;
  let vgCreating = false;
  let availableVGDevices = [];
  let newVG = {
    name: "",
    devices: []
  };

  // Form data for new LUN
  let newLUN = {
    name: "",
    size_gb: 10,
    backend_type: "lvm",
    volume_group: "",
    device_path: "",
    allowed_iqns: []
  };

  // Notification system
  let notifications = [];
  let notificationId = 0;

  function showNotification(message, type = 'info') {
    const id = ++notificationId;
    notifications = [...notifications, { id, message, type }];
    setTimeout(() => {
      notifications = notifications.filter(n => n.id !== id);
    }, 5000);
  }

  // API calls
  async function loadTarget() {
    try {
      const response = await fetch('/api/iscsi/target');
      if (!response.ok) throw new Error('Failed to load target');
      target = await response.json();
    } catch (err) {
      // Target might not exist yet, that's ok
      target = {
        iqn: "iqn.2024-01.com.nas:storage",
        name: "Arcanas NAS Storage",
        status: "inactive",
        luns: [],
        sessions: 0
      };
    }
  }

  async function loadLUNs() {
    try {
      loading = true;
      const response = await fetch('/api/iscsi/luns');
      if (!response.ok) throw new Error('Failed to load LUNs');
      luns = (await response.json()) || [];
    } catch (err) {
      error = err.message;
      showNotification(err.message, 'error');
      luns = []; // Ensure luns is always an array
    } finally {
      loading = false;
    }
  }

  async function loadBackends() {
    try {
      const response = await fetch('/api/iscsi/backends');
      if (!response.ok) throw new Error('Failed to load backends');
      backends = (await response.json()) || [];
    } catch (err) {
      showNotification('Failed to load backend options', 'warning');
      backends = []; // Ensure backends is always an array
    }
  }

  async function createLUN() {
    try {
      creatingLUN = true;
      const response = await fetch('/api/iscsi/luns', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newLUN)
      });
      if (!response.ok) {
        const err = await response.text();
        throw new Error(err || 'Failed to create LUN');
      }
      showNotification('LUN created successfully', 'success');
      showCreateModal = false;
      resetCreateForm();
      await loadLUNs();
    } catch (err) {
      showNotification(err.message, 'error');
    } finally {
      creatingLUN = false;
    }
  }

  async function deleteLUN(lunNum) {
    if (!confirm(`Are you sure you want to delete LUN ${lunNum}?`)) return;

    try {
      const response = await fetch(`/api/iscsi/luns/${lunNum}`, {
        method: 'DELETE'
      });
      if (!response.ok) throw new Error('Failed to delete LUN');
      showNotification(`LUN ${lunNum} deleted`, 'success');
      await loadLUNs();
    } catch (err) {
      showNotification(err.message, 'error');
    }
  }

  // ACL management functions
  async function loadACLs() {
    try {
      const response = await fetch('/api/iscsi/acls');
      if (!response.ok) throw new Error('Failed to load ACLs');
      acls = (await response.json()) || [];
    } catch (err) {
      showNotification('Failed to load ACLs', 'warning');
      acls = [];
    }
  }

  async function createACL() {
    try {
      creatingACL = true;
      const response = await fetch('/api/iscsi/acls', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newACL)
      });
      if (!response.ok) {
        const err = await response.text();
        throw new Error(err || 'Failed to create ACL');
      }
      showNotification('ACL created successfully', 'success');
      showACLModal = false;
      newACL = { initiator_iqn: "", name: "" };
      await loadACLs();
    } catch (err) {
      showNotification(err.message, 'error');
    } finally {
      creatingACL = false;
    }
  }

  async function deleteACL(iqn) {
    if (!confirm(`Are you sure you want to delete ACL for ${iqn}?`)) return;

    try {
      const response = await fetch(`/api/iscsi/acls/${encodeURIComponent(iqn)}`, {
        method: 'DELETE'
      });
      if (!response.ok) throw new Error('Failed to delete ACL');
      showNotification(`ACL deleted`, 'success');
      await loadACLs();
    } catch (err) {
      showNotification(err.message, 'error');
    }
  }

  async function mapLUNToACL(iqn, sourceLUN, targetLUN) {
    try {
      const response = await fetch(`/api/iscsi/acls/${encodeURIComponent(iqn)}/luns`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ source_lun: sourceLUN, target_lun: targetLUN })
      });
      if (!response.ok) {
        const err = await response.text();
        throw new Error(err || 'Failed to map LUN');
      }
      showNotification('LUN mapped successfully', 'success');
      await loadACLs();
      return true;
    } catch (err) {
      showNotification(err.message, 'error');
      return false;
    }
  }

  async function unmapLUNFromACL(iqn, targetLUN) {
    try {
      const response = await fetch(`/api/iscsi/acls/${encodeURIComponent(iqn)}/luns/${targetLUN}`, {
        method: 'DELETE'
      });
      if (!response.ok) throw new Error('Failed to unmap LUN');
      showNotification('LUN unmapped successfully', 'success');
      await loadACLs();
    } catch (err) {
      showNotification(err.message, 'error');
    }
  }

  function openMapLUNModal(acl) {
    selectedACLForMapping = acl;
    newLUNMapping = {
      source_lun: 0,
      target_lun: 0
    };
    showMapLUNModal = true;
  }

  async function submitLUNMapping() {
    if (!selectedACLForMapping) return;
    const success = await mapLUNToACL(
      selectedACLForMapping.initiator_iqn,
      newLUNMapping.source_lun,
      newLUNMapping.target_lun
    );
    if (success) {
      showMapLUNModal = false;
      selectedACLForMapping = null;
    }
  }

  function getACLDisplayName(acl) {
    return acl.name && acl.name !== acl.initiator_iqn ? acl.name : acl.initiator_iqn;
  }

  function resetCreateForm() {
    newLUN = {
      name: "",
      size_gb: 10,
      backend_type: "lvm",
      volume_group: backends.find(b => b.type === 'lvm')?.resources?.[0] || "",
      device_path: "",
      allowed_iqns: []
    };
  }

  function getBackendInfo(type) {
    return backends.find(b => b.type === type);
  }

  function getBackendColor(type) {
    switch(type) {
      case 'lvm': return 'bg-green-100 dark:bg-green-900/40 text-green-700 dark:text-green-400';
      case 'block': return 'bg-orange-100 dark:bg-orange-900/40 text-orange-700 dark:text-orange-400';
      case 'fileio': return 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300';
      default: return 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300';
    }
  }

  // Generate connection command for iSCSI initiator
  function getConnectCommand(iqn) {
    const serverIP = window.location.hostname;
    return `sudo iscsiadm -m node -T ${iqn} -p ${serverIP} -l`;
  }

  async function copyToClipboard(text) {
    try {
      await navigator.clipboard.writeText(text);
    } catch (err) {
      const textArea = document.createElement('textarea');
      textArea.value = text;
      textArea.style.position = 'fixed';
      textArea.style.left = '-999999px';
      document.body.appendChild(textArea);
      textArea.select();
      try {
        document.execCommand('copy');
      } catch (err) {
        console.error('Failed to copy:', err);
      }
      document.body.removeChild(textArea);
    }
  }

  // Quick create VG functions
  async function loadAvailableVGDevices() {
    try {
      const result = await fetch('/api/volume-groups/available-devices');
      if (!result.ok) throw new Error('Failed to load available devices');
      availableVGDevices = (await result.json()) || [];
    } catch (err) {
      showNotification('Failed to load available devices', 'error');
      availableVGDevices = [];
    }
  }

  async function quickCreateVG() {
    try {
      vgCreating = true;
      const result = await fetch('/api/volume-groups', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newVG)
      });
      if (!result.ok) {
        const err = await result.text();
        throw new Error(err || 'Failed to create volume group');
      }
      showNotification('Volume group created successfully', 'success');
      showCreateVGModal = false;
      showCreateModal = false; // Close the LUN modal too
      newVG = { name: "", devices: [] };
      await loadBackends(); // Refresh backends to show new VG
      // Open LUN modal again with new VG pre-selected
      showCreateModal = true;
      const lvmBackend = backends.find(b => b.type === 'lvm');
      if (lvmBackend && lvmBackend.resources.length > 0) {
        newLUN.volume_group = lvmBackend.resources[0];
      }
    } catch (err) {
      showNotification(err.message, 'error');
    } finally {
      vgCreating = false;
    }
  }

  onMount(() => {
    loadTarget();
    loadLUNs();
    loadBackends();
    loadACLs();
  });
</script>

<div class="p-6">
  <!-- Notifications -->
  {#if notifications.length > 0}
    <div class="fixed top-4 right-4 z-50 space-y-2">
      {#each notifications as notification (notification.id)}
        <div
          class="p-4 rounded-md shadow-lg flex items-center justify-between max-w-sm
            {notification.type === 'error' ? 'bg-red-50 border border-red-200 text-red-700' : ''}
            {notification.type === 'warning' ? 'bg-yellow-50 border border-yellow-200 text-yellow-700' : ''}
            {notification.type === 'success' ? 'bg-green-50 border border-green-200 text-green-700' : ''}
            {notification.type === 'info' ? 'bg-blue-50 border border-blue-200 text-blue-700' : ''}"
        >
          <span class="text-sm font-medium">{notification.message}</span>
          <button
            on:click={() => notifications = notifications.filter(n => n.id !== notification.id)}
            class="ml-4 text-sm underline hover:no-underline"
          >
            Dismiss
          </button>
        </div>
      {/each}
    </div>
  {/if}

  <!-- Header -->
  <div class="mb-6">
    <h1 class="text-xl font-bold text-gray-900 dark:text-white mb-2">
      iSCSI LUN Management
    </h1>
    <p class="text-sm text-gray-600 dark:text-gray-300">
      Single target: <code class="text-blue-600 dark:text-blue-400">{target?.iqn || 'Loading...'}</code>
    </p>
  </div>

  <!-- Summary Stats -->
  <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
    <div class="bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 rounded-xl p-4 border border-blue-100 dark:border-blue-800">
      <p class="text-xs text-blue-600 dark:text-blue-400 font-medium uppercase tracking-wide">Target Status</p>
      <p class="text-lg font-bold text-gray-900 dark:text-white flex items-center">
        {#if target?.status === 'active'}
          <span class="flex items-center">
            <span class="w-2.5 h-2.5 bg-green-500 rounded-full mr-2"></span>
            Active
          </span>
        {:else}
          <span class="text-gray-600 dark:text-gray-300">Inactive</span>
        {/if}
      </p>
    </div>

    <div class="bg-gradient-to-br from-purple-50 to-pink-50 dark:from-purple-900/20 dark:to-pink-900/20 rounded-xl p-4 border border-purple-100 dark:border-purple-800">
      <p class="text-xs text-purple-600 dark:text-purple-400 font-medium uppercase tracking-wide">Total LUNs</p>
      <p class="text-2xl font-bold text-gray-900 dark:text-white">{luns.length}</p>
    </div>

    <div class="bg-gradient-to-br from-green-50 to-emerald-50 dark:from-green-900/20 dark:to-emerald-900/20 rounded-xl p-4 border border-green-100 dark:border-green-800">
      <p class="text-xs text-green-600 dark:text-green-400 font-medium uppercase tracking-wide">Active Sessions</p>
      <p class="text-2xl font-bold text-gray-900 dark:text-white">{target?.sessions || 0}</p>
    </div>

    <div class="bg-gradient-to-br from-orange-50 to-amber-50 dark:from-orange-900/20 dark:to-amber-900/20 rounded-xl p-4 border border-orange-100 dark:border-orange-800">
      <p class="text-xs text-orange-600 dark:text-orange-400 font-medium uppercase tracking-wide">Total Capacity</p>
      <p class="text-2xl font-bold text-gray-900 dark:text-white">
        {luns.reduce((sum, lun) => sum + (lun.size_gb || 0), 0).toFixed(1)} GB
      </p>
    </div>
  </div>

  <!-- Actions -->
  <div class="flex justify-between items-center mb-6">
    <div class="flex space-x-2">
      <button
        on:click={loadLUNs}
        class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-card border border-gray-300 dark:border rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-muted"
        disabled={loading}
      >
        <svg class="w-4 h-4 mr-2 inline" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
        Refresh
      </button>
      <button
        on:click={loadACLs}
        class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-card border border-gray-300 dark:border rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-muted"
      >
        <svg class="w-4 h-4 mr-2 inline" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
        </svg>
        Refresh ACLs
      </button>
    </div>
    <div class="flex space-x-2">
      <button
        on:click={() => { showACLModal = true; }}
        class="px-4 py-2 bg-purple-600 text-white rounded-md shadow-sm hover:bg-purple-700"
      >
        <svg class="w-4 h-4 mr-2 inline" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
        </svg>
        Add Client ACL
      </button>
      <button
        on:click={() => { showCreateModal = true; resetCreateForm(); }}
        class="px-4 py-2 bg-blue-600 text-white rounded-md shadow-sm hover:bg-blue-700"
      >
        <svg class="w-4 h-4 mr-2 inline" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0v6h6m-6 0v6m0-6h6" />
        </svg>
        Create LUN
      </button>
    </div>
  </div>

  <!-- Compact ACL Bar -->
  {#if acls.length > 0}
    <div class="mb-6 border border-purple-200 dark:border-purple-800 rounded-lg overflow-hidden">
      <button
        on:click={() => aclsExpanded = !aclsExpanded}
        class="w-full px-4 py-3 bg-gradient-to-r from-purple-50 to-indigo-50 dark:from-purple-900/20 dark:to-indigo-900/20 flex items-center justify-between hover:from-purple-100 hover:to-indigo-100 dark:hover:from-purple-900/30 dark:hover:to-indigo-900/30 transition-colors"
      >
        <div class="flex items-center">
          <svg class="w-4 h-4 mr-2 text-purple-600 dark:text-purple-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
          </svg>
          <span class="text-sm font-semibold text-gray-700 dark:text-gray-200">
            {acls.length} Client{acls.length === 1 ? '' : 's'}
          </span>
          <span class="ml-3 text-xs text-gray-500 dark:text-gray-400">
            {acls.reduce((sum, acl) => sum + (acl.mapped_luns?.length || 0), 0)} LUN{acls.reduce((sum, acl) => sum + (acl.mapped_luns?.length || 0), 0) === 1 ? '' : 's'} mapped
          </span>
        </div>
        <svg class="w-4 h-4 text-gray-400 transform transition-transform {aclsExpanded ? 'rotate-180' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>
      {#if aclsExpanded}
        <div class="border-t border-purple-200 dark:border-purple-800 bg-white dark:bg-card p-3">
          <div class="space-y-3">
            {#each acls as acl (acl.initiator_iqn)}
              <div class="rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
                <!-- ACL Header -->
                <div class="flex items-center justify-between px-3 py-2 bg-gray-50 dark:bg-gray-800/50">
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2">
                      <span class="text-sm font-medium text-gray-900 dark:text-white truncate" title={acl.initiator_iqn}>
                        {getACLDisplayName(acl)}
                      </span>
                      <span class="text-xs text-gray-500 dark:text-gray-400">
                        {acl.mapped_luns?.length || 0} LUN{acl.mapped_luns?.length === 1 ? '' : 's'} mapped
                      </span>
                    </div>
                  </div>
                  <div class="flex items-center gap-1">
                    <button
                      on:click={() => openMapLUNModal(acl)}
                      class="p-1.5 text-gray-400 hover:text-purple-600 hover:bg-purple-50 dark:hover:bg-purple-900/20 rounded"
                      title="Map LUN"
                    >
                      <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0v6h6m-6 0v6m0-6h6" />
                      </svg>
                    </button>
                    <button
                      on:click={() => deleteACL(acl.initiator_iqn)}
                      class="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded"
                      title="Delete ACL"
                    >
                      <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1 10V4a1 1 0 011-1h-4a1 1 0 01-1 1v3m0 0h-1a2 2 0 00-2 2v-2a2 2 0 012 2h12a2 2 0 012 2v2a2 2 0 002 2h-3m-3 0h4m-3 0v6" />
                      </svg>
                    </button>
                  </div>
                </div>

                <!-- Mapped LUNs List -->
                <div class="p-2 bg-white dark:bg-card">
                  {#if acl.mapped_luns && acl.mapped_luns.length > 0}
                    <div class="space-y-1">
                      {#each acl.mapped_luns as lunNum}
                        {@const mappedLUN = luns.find(l => l.lun === lunNum)}
                        {#if mappedLUN}
                          <div class="flex items-center justify-between px-2 py-1.5 rounded text-sm hover:bg-gray-50 dark:hover:bg-gray-800/30 group">
                            <div class="flex items-center gap-2 min-w-0">
                              <span class="px-1.5 py-0.5 text-xs font-medium rounded bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-400">
                                LUN {mappedLUN.lun}
                              </span>
                              <span class="text-gray-700 dark:text-gray-300 truncate">
                                {mappedLUN.name || `LUN ${mappedLUN.lun}`}
                              </span>
                              <span class="text-xs text-gray-500 dark:text-gray-400">
                                {mappedLUN.size_gb || 0} GB
                              </span>
                            </div>
                            <button
                              on:click={() => unmapLUNFromACL(acl.initiator_iqn, lunNum)}
                              class="p-1 text-gray-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded opacity-0 group-hover:opacity-100 transition-opacity"
                              title="Unmap LUN"
                            >
                              <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                              </svg>
                            </button>
                          </div>
                        {:else}
                          <div class="flex items-center justify-between px-2 py-1.5 rounded text-sm bg-gray-50 dark:bg-gray-800/30">
                            <div class="flex items-center gap-2">
                              <span class="px-1.5 py-0.5 text-xs font-medium rounded bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400">
                                LUN {lunNum}
                              </span>
                              <span class="text-gray-500 dark:text-gray-400 text-xs italic">
                                (LUN not found)
                              </span>
                            </div>
                            <button
                              on:click={() => unmapLUNFromACL(acl.initiator_iqn, lunNum)}
                              class="p-1 text-gray-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded opacity-0 group-hover:opacity-100 transition-opacity"
                              title="Unmap LUN"
                            >
                              <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                              </svg>
                            </button>
                          </div>
                        {/if}
                      {/each}
                    </div>
                  {:else}
                    <div class="text-center py-3 text-sm text-gray-500 dark:text-gray-400 italic">
                      No LUNs mapped to this client
                    </div>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/if}
    </div>
  {/if}

  <!-- Loading -->
  {#if loading}
    <div class="flex justify-center items-center h-64">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>
  {:else if luns.length === 0}
    <!-- Empty State -->
    <div class="text-center py-12 bg-white dark:bg-card shadow-lg hover:shadow-xl transition-shadow duration-200 rounded-lg border border-gray-100 dark:border-border">
      <svg class="h-16 w-16 text-gray-400 mx-auto mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4m4 0l4 4m4-4h4m-4 4v6" />
      </svg>
      <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">No LUNs configured</h3>
      <p class="text-gray-600 dark:text-gray-300 mb-4">
        Create LUNs to share storage via iSCSI
      </p>
      <button
        on:click={() => { showCreateModal = true; resetCreateForm(); }}
        class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
      >
        Create LUN
      </button>
    </div>
  {:else}
    <!-- LUNs List -->
    <div class="space-y-4">
      {#each luns as lun (lun.id)}
        <div class="bg-white dark:bg-card shadow-lg hover:shadow-xl transition-shadow duration-200 rounded-lg border border-gray-100 dark:border-border">
          <!-- Header Section -->
          <div class="p-6">
            <div class="flex items-start justify-between">
              <div class="flex items-center space-x-4">
                <!-- LUN Icon -->
                <div class="relative">
                  <div class="absolute inset-0 bg-gradient-to-br from-blue-400 to-indigo-500 rounded-xl blur opacity-25"></div>
                  <div class="relative w-14 h-14 bg-gradient-to-br from-blue-100 to-indigo-100 dark:from-blue-900/40 dark:to-indigo-900/40 rounded-xl flex items-center justify-center shadow-lg">
                    <svg class="w-7 h-7 text-blue-600 dark:text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
                    </svg>
                  </div>
                </div>

                <!-- LUN Info -->
                <div>
                  <div class="flex items-center space-x-2">
                    <h3 class="text-xl font-bold text-gray-900 dark:text-white">
                      {lun.name || `LUN ${lun.lun}`}
                    </h3>
                    <span class="px-2 py-0.5 text-xs font-semibold rounded-full bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-400">
                      LUN {lun.lun}
                    </span>
                    <span class="px-2 py-0.5 text-xs font-semibold rounded-full {getBackendColor(lun.backend_type)}">
                      {lun.backend_type?.toUpperCase() || 'BLOCK'}
                    </span>
                  </div>
                  <div class="mt-1 flex items-center space-x-4 text-sm text-gray-600 dark:text-gray-300">
                    <span>{lun.size_gb || 0} GB</span>
                    <span class="px-2 py-0.5 rounded-full bg-green-100 dark:bg-green-900/40 text-green-700 dark:text-green-400">
                      {lun.status === 'active' ? 'Active' : 'Inactive'}
                    </span>
                  </div>
                  {#if lun.backend_path}
                    <p class="mt-1 text-xs font-mono text-gray-500 dark:text-gray-400">
                      {lun.backend_path}
                    </p>
                  {/if}
                </div>
              </div>

              <!-- Actions -->
              <div class="flex items-center space-x-2">
                <button
                  on:click={() => deleteLUN(lun.lun)}
                  class="p-2 rounded-lg text-gray-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20"
                  title="Delete LUN"
                >
                  <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1 10V4a1 1 0 011-1h-4a1 1 0 01-1 1v3m0 0h-1a2 2 0 00-2 2v-2a2 2 0 012 2h12a2 2 0 012 2v2a2 2 0 002 2h-3m-3 0h4m-3 0v6" />
                  </svg>
                </button>
              </div>
            </div>
          </div>

          <!-- Stats Section -->
          <div class="bg-gray-50 dark:bg-muted rounded-b-lg p-4 border-t border-gray-200 dark:border-border">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
              <!-- Allocated Size -->
              <div class="bg-gradient-to-br from-orange-50 to-amber-50 dark:from-orange-900/20 dark:to-amber-900/20 rounded-xl p-4 border border-orange-100 dark:border-orange-800">
                <div class="flex items-center space-x-3">
                  <div class="w-10 h-10 bg-orange-100 dark:bg-orange-800 rounded-lg flex items-center justify-center">
                    <svg class="w-5 h-5 text-orange-600 dark:text-orange-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"/>
                    </svg>
                  </div>
                  <div>
                    <p class="text-xs text-orange-600 dark:text-orange-400 font-medium uppercase tracking-wide">
                      Allocated Size
                    </p>
                    <p class="text-lg font-bold text-gray-900 dark:text-white">
                      {#if lun.backend_type === 'lvm' && lun.lv_size_bytes}
                        {formatBytes(lun.lv_size_bytes)}
                      {:else}
                        {lun.size_gb || 0} GB
                      {/if}
                    </p>
                  </div>
                </div>
              </div>

              <!-- Backend Type -->
              <div class="bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 rounded-xl p-4 border border-blue-100 dark:border-blue-800">
                <div class="flex items-center space-x-3">
                  <div class="w-10 h-10 bg-blue-100 dark:bg-blue-800 rounded-lg flex items-center justify-center">
                    <svg class="w-5 h-5 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"/>
                    </svg>
                  </div>
                  <div>
                    <p class="text-xs text-blue-600 dark:text-blue-400 font-medium uppercase tracking-wide">
                      Backend
                    </p>
                    <p class="text-lg font-bold text-gray-900 dark:text-white">
                      {lun.backend_type?.toUpperCase() || 'BLOCK'}
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <!-- LV Info for LVM-backed LUNs -->
            {#if lun.backend_type === 'lvm' && lun.lv_path}
              <div class="mt-4">
                <div class="bg-gradient-to-br from-green-50 to-emerald-50 dark:from-green-900/20 dark:to-emerald-900/20 rounded-xl p-4 border border-green-100 dark:border-green-800">
                  <div class="flex items-center space-x-3">
                    <div class="w-10 h-10 bg-green-100 dark:bg-green-800 rounded-lg flex items-center justify-center">
                      <svg class="w-5 h-5 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"/>
                      </svg>
                    </div>
                    <div>
                      <p class="text-xs text-green-600 dark:text-green-400 font-medium uppercase tracking-wide">
                        LV Path
                      </p>
                      <p class="text-sm font-mono text-gray-900 dark:text-white">
                        {lun.lv_path}
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            {/if}

            <!-- Connection Command -->
            <div>
              <p class="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-2">
                Connection Command
              </p>
              <div class="flex items-start space-x-2">
                <code class="text-xs font-mono text-gray-800 dark:text-gray-200 bg-white dark:bg-gray-900 px-3 py-2 rounded border border-gray-200 dark:border flex-1 select-all">
                  {getConnectCommand(target?.iqn || 'iqn.2024-01.com.nas:storage')}
                </code>
                <button
                  on:click={() => copyToClipboard(getConnectCommand(target?.iqn || 'iqn.2024-01.com.nas:storage'))}
                  class="p-1.5 rounded hover:bg-gray-200 dark:hover:bg-gray-700 flex-shrink-0"
                  title="Copy connection command"
                >
                  <svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Create LUN Modal -->
{#if showCreateModal}
  <div
    class="fixed inset-0 z-50 overflow-y-auto"
    on:click={() => showCreateModal = false}
    on:keydown={(e) => e.key === 'Escape' && (showCreateModal = false)}
    role="dialog"
    aria-modal="true"
    aria-labelledby="create-lun-title"
    tabindex="-1"
  >
    <div class="flex items-center justify-center min-h-screen px-4">
      <div
        class="fixed inset-0 bg-gray-500 bg-opacity-75"
        aria-hidden="true"
        on:click={() => showCreateModal = false}
      ></div>
      <div
        class="relative bg-white dark:bg-card rounded-lg max-w-lg w-full p-6"
        on:click|stopPropagation
      >
        <h3 id="create-lun-title" class="text-lg font-medium text-gray-900 dark:text-white mb-4">
          Create iSCSI LUN
        </h3>

        <form on:submit|preventDefault={createLUN} class="space-y-4">
          <!-- Name -->
          <div>
            <label for="lun-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              LUN Name
            </label>
            <input
              id="lun-name"
              type="text"
              bind:value={newLUN.name}
              class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md dark:bg-muted dark:text-white"
              placeholder="e.g., Client 1 Storage"
              required
            />
          </div>

          <!-- Size -->
          <div>
            <label for="lun-size" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Size (GB)
            </label>
            <input
              id="lun-size"
              type="number"
              bind:value={newLUN.size_gb}
              min="1"
              step="0.1"
              class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md dark:bg-muted dark:text-white"
              required
            />
          </div>

          <!-- Backend Type -->
          <div>
            <label for="lun-backend" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Storage Backend
            </label>
            <select
              id="lun-backend"
              bind:value={newLUN.backend_type}
              class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md dark:bg-muted dark:text-white"
              required
              on:change={() => {
                // Reset device/volume group fields when backend type changes
                if (newLUN.backend_type !== 'block') {
                  newLUN.device_path = "";
                }
                if (newLUN.backend_type !== 'lvm') {
                  newLUN.volume_group = "";
                }
              }}
            >
              {#each backends as backend}
                {#if backend.available}
                  <option value={backend.type}>{backend.name}</option>
                {/if}
              {/each}
            </select>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              {getBackendInfo(newLUN.backend_type)?.description || ''}
            </p>
          </div>

          <!-- Backend-specific fields -->
          {#if newLUN.backend_type === 'lvm'}
            <div>
              <label for="lun-vg" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Volume Group
              </label>
              {#if getBackendInfo('lvm')?.resources && getBackendInfo('lvm').resources.length > 0}
                <select
                  id="lun-vg"
                  bind:value={newLUN.volume_group}
                  class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md dark:bg-muted dark:text-white"
                  required
                >
                  <option value="">Select volume group...</option>
                  {#each getBackendInfo('lvm').resources || [] as vg}
                    <option value={vg}>{vg}</option>
                  {/each}
                </select>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  Select a volume group to create flexible LUNs from
                </p>
              {:else}
                <div class="p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-md">
                  <p class="text-sm text-blue-800 dark:text-blue-200 mb-3">
                    No volume groups available. Create a volume group to use LVM flexible LUNs.
                  </p>
                  <button
                    on:click={async () => {
                      await loadAvailableVGDevices();
                      showCreateVGModal = true;
                    }}
                    class="px-3 py-1.5 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700"
                  >
                    Create Volume Group
                  </button>
                </div>
              {/if}
            </div>
          {:else if newLUN.backend_type === 'block'}
            <div>
              <label for="lun-device" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Block Device
              </label>
              <select
                id="lun-device"
                bind:value={newLUN.device_path}
                class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md dark:bg-muted dark:text-white"
                required
              >
                <option value="">Select device...</option>
                {#each getBackendInfo('block')?.resources || [] as dev}
                  <option value={dev}>{dev}</option>
                {/each}
              </select>
            </div>
          {/if}

          <!-- Buttons -->
          <div class="flex justify-end space-x-3 pt-4">
            <button
              type="button"
              on:click={() => showCreateModal = false}
              class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-card border border-gray-300 dark:border rounded-md hover:bg-gray-50 dark:hover:bg-muted"
            >
              Cancel
            </button>
            <button
              type="submit"
              class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
              disabled={creatingLUN}
            >
              {#if creatingLUN}
                <svg class="animate-spin h-4 w-4 mr-2 inline" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8 0V0C5.373 0 0 1 4 0z"></path>
                </svg>
                <span>Creating...</span>
              {:else}
                <span>Create LUN</span>
              {/if}
              </button>
          </div>
        </form>
      </div>
    </div>
  </div>
{/if}

<!-- Quick Create VG Modal -->
{#if showCreateVGModal}
  <div
    class="fixed inset-0 z-50 overflow-y-auto"
    on:click={() => showCreateVGModal = false}
    on:keydown={(e) => e.key === 'Escape' && (showCreateVGModal = false)}
    role="dialog"
    aria-modal="true"
    aria-labelledby="create-vg-title-scsi"
    tabindex="-1"
  >
    <div class="flex items-center justify-center min-h-screen px-4">
      <div
        class="fixed inset-0 bg-gray-500 bg-opacity-75"
        aria-hidden="true"
        on:click={() => showCreateVGModal = false}
      ></div>
      <div
        class="relative bg-white dark:bg-card rounded-lg max-w-lg w-full p-6"
        on:click|stopPropagation
      >
        <h3 id="create-vg-title-scsi" class="text-lg font-medium text-gray-900 dark:text-white mb-4">
          Create Volume Group for iSCSI
        </h3>

        <form on:submit|preventDefault={quickCreateVG} class="space-y-4">
          <div>
            <label for="vg-name-scsi" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Volume Group Name
            </label>
            <input
              id="vg-name-scsi"
              type="text"
              bind:value={newVG.name}
              class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md dark:bg-muted dark:text-white"
              placeholder="e.g., vg-iscsi"
              pattern="[a-z0-9-]+"
              title="Only lowercase letters, numbers, and hyphens"
              required
            />
          </div>

          <fieldset class="border-0 p-0 m-0">
            <legend class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Physical Devices
            </legend>
            <div class="space-y-2 max-h-48 overflow-y-auto border border-gray-300 dark:border rounded-md p-2 dark:bg-muted">
              {#each availableVGDevices as device (device.path)}
                <label class="flex items-center space-x-2 p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded cursor-pointer">
                  <input
                    type="checkbox"
                    bind:group={newVG.devices}
                    value={device.path}
                    disabled={!device.available}
                    class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  />
                  <div class="flex-1">
                    <div class="text-sm font-medium text-gray-900 dark:text-white">{device.path}</div>
                    {#if device.available}
                      <div class="text-xs text-green-600 dark:text-green-400">Available</div>
                    {:else}
                      <div class="text-xs text-red-600 dark:text-red-400">{device.reason}</div>
                    {/if}
                  </div>
                </label>
              {/each}
            </div>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
              Select unmounted devices (typically RAID arrays) to include in the volume group
            </p>
          </fieldset>

          <div class="flex justify-end space-x-3 pt-4">
            <button
              type="button"
              on:click={() => showCreateVGModal = false}
              class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-card border border-gray-300 dark:border rounded-md hover:bg-gray-50 dark:hover:bg-muted"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={vgCreating || newVG.devices.length === 0}
              class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
            >
              {#if vgCreating}
                <svg class="animate-spin h-4 w-4 mr-2 inline" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8 0V0C5.373 0 0 1.4 0z"></path>
                </svg>
                Creating...
              {:else}
                Create Volume Group
              {/if}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
{/if}

<!-- Create ACL Modal -->
{#if showACLModal}
  <div
    class="fixed inset-0 z-50 overflow-y-auto"
    on:click={() => showACLModal = false}
    on:keydown={(e) => e.key === 'Escape' && (showACLModal = false)}
    role="dialog"
    aria-modal="true"
    aria-labelledby="create-acl-title"
    tabindex="-1"
  >
    <div class="flex items-center justify-center min-h-screen px-4">
      <div
        class="fixed inset-0 bg-gray-500 bg-opacity-75"
        aria-hidden="true"
        on:click={() => showACLModal = false}
      ></div>
      <div
        class="relative bg-white dark:bg-card rounded-lg max-w-lg w-full p-6"
        on:click|stopPropagation
      >
        <h3 id="create-acl-title" class="text-lg font-medium text-gray-900 dark:text-white mb-4">
          Add Client ACL
        </h3>

        <form on:submit|preventDefault={createACL} class="space-y-4">
          <div>
            <label for="acl-iqn" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Initiator IQN
            </label>
            <input
              id="acl-iqn"
              type="text"
              bind:value={newACL.initiator_iqn}
              class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md dark:bg-muted dark:text-white font-mono text-sm"
              placeholder="iqn.1993-08.org.debian:01:abc123"
              pattern="^iqn\.[a-z0-9\-\.]+:[a-z0-9\-\.]+$"
              title="Must be a valid IQN format (e.g., iqn.yyyy-mm.domain:identifier)"
              required
            />
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              The iSCSI Qualified Name of the client. Find it on the client: cat /etc/iscsi/initiatorname.iscsi
            </p>
          </div>

          <div>
            <label for="acl-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Client Name (Optional)
            </label>
            <input
              id="acl-name"
              type="text"
              bind:value={newACL.name}
              class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md dark:bg-muted dark:text-white"
              placeholder="e.g., Backup Server, Client 1"
            />
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              A friendly name to identify this client. Defaults to IQN if not provided.
            </p>
          </div>

          <div class="flex justify-end space-x-3 pt-4">
            <button
              type="button"
              on:click={() => showACLModal = false}
              class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-card border border-gray-300 dark:border rounded-md hover:bg-gray-50 dark:hover:bg-muted"
            >
              Cancel
            </button>
            <button
              type="submit"
              class="px-4 py-2 bg-purple-600 text-white rounded-md hover:bg-purple-700"
              disabled={creatingACL}
            >
              {#if creatingACL}
                <svg class="animate-spin h-4 w-4 mr-2 inline" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8 0V0C5.373 0 0 1.4 0z"></path>
                </svg>
                <span>Creating...</span>
              {:else}
                <span>Create ACL</span>
              {/if}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
{/if}

<!-- Map LUN to ACL Modal -->
{#if showMapLUNModal && selectedACLForMapping}
  <div
    class="fixed inset-0 z-50 overflow-y-auto"
    on:click={() => showMapLUNModal = false}
    on:keydown={(e) => e.key === 'Escape' && (showMapLUNModal = false)}
    role="dialog"
    aria-modal="true"
    aria-labelledby="map-lun-title"
    tabindex="-1"
  >
    <div class="flex items-center justify-center min-h-screen px-4">
      <div
        class="fixed inset-0 bg-gray-500 bg-opacity-75"
        aria-hidden="true"
        on:click={() => showMapLUNModal = false}
      ></div>
      <div
        class="relative bg-white dark:bg-card rounded-lg max-w-lg w-full p-6"
        on:click|stopPropagation
      >
        <h3 id="map-lun-title" class="text-lg font-medium text-gray-900 dark:text-white mb-4">
          Map LUN to {getACLDisplayName(selectedACLForMapping)}
        </h3>

        <form on:submit|preventDefault={submitLUNMapping} class="space-y-4">
          <div>
            <label for="map-source-lun" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Source LUN
            </label>
            <select
              id="map-source-lun"
              bind:value={newLUNMapping.source_lun}
              class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md dark:bg-muted dark:text-white"
              required
            >
              <option value={-1}>Select a LUN...</option>
              {#each luns as lun}
                <option value={lun.lun}>LUN {lun.lun}: {lun.name || 'Unnamed'} ({lun.size_gb} GB)</option>
              {/each}
            </select>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              The actual LUN number to share with this client
            </p>
          </div>

          <div>
            <label for="map-target-lun" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Target LUN (Optional)
            </label>
            <input
              id="map-target-lun"
              type="number"
              min="0"
              max="255"
              bind:value={newLUNMapping.target_lun}
              class="w-full px-3 py-2 border border-gray-300 dark:border rounded-md dark:bg-muted dark:text-white"
              placeholder="Leave empty to use same as source"
            />
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              The LUN number as seen by the client. Leave empty to use the same number as source.
            </p>
          </div>

          <div class="flex justify-end space-x-3 pt-4">
            <button
              type="button"
              on:click={() => showMapLUNModal = false}
              class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-white dark:bg-card border border-gray-300 dark:border rounded-md hover:bg-gray-50 dark:hover:bg-muted"
            >
              Cancel
            </button>
            <button
              type="submit"
              class="px-4 py-2 bg-purple-600 text-white rounded-md hover:bg-purple-700"
            >
              Map LUN
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
{/if}
