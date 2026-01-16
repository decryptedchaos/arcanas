<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { settingsAPI } from '$lib/api.js';
  import { onMount } from 'svelte';

  let activeTab = 'general';
  const tabs = [
    { id: 'general', label: 'General', icon: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z' },
    { id: 'network', label: 'Network', icon: 'M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0' }
  ];

  // General settings
  let settings = {
    hostname: '',
    timezone: '',
    locale: ''
  };
  let timezoneList = [];
  let loading = true;
  let saving = false;
  let error = null;
  let success = null;

  // Network settings
  let networkInterfaces = [];
  let selectedInterface = null;
  let showNetworkModal = false;
  let networkForm = {
    method: 'dhcp',
    address: '',
    netmask: '',
    gateway: '',
    dns: ['']
  };

  onMount(async () => {
    await loadSettings();
    await loadNetworkInterfaces();
  });

  async function loadSettings() {
    try {
      loading = true;
      const [settingsData, timezoneData] = await Promise.all([
        settingsAPI.getSettings(),
        settingsAPI.getTimezone()
      ]);
      settings = settingsData;
      timezoneList = timezoneData.available;
    } catch (err) {
      error = err.message || 'Failed to load settings';
    } finally {
      loading = false;
    }
  }

  async function loadNetworkInterfaces() {
    try {
      networkInterfaces = await settingsAPI.getNetworkConfig();
    } catch (err) {
      console.error('Failed to load network interfaces:', err);
    }
  }

  async function saveSettings() {
    try {
      saving = true;
      error = null;
      await settingsAPI.updateSettings(settings);
      success = 'Settings saved successfully';
      setTimeout(() => success = null, 3000);
    } catch (err) {
      error = err.message || 'Failed to save settings';
    } finally {
      saving = false;
    }
  }

  function openNetworkModal(iface) {
    selectedInterface = iface;
    networkForm = {
      method: iface.ipv4.address ? 'static' : 'dhcp',
      address: iface.ipv4.address || '',
      netmask: iface.ipv4.netmask || '24',
      gateway: iface.ipv4.gateway || '',
      dns: iface.dns && iface.dns.length > 0 ? [...iface.dns] : ['']
    };
    showNetworkModal = true;
  }

  async function saveNetworkConfig() {
    try {
      const config = {
        interface: selectedInterface.name,
        config: {
          name: selectedInterface.name,
          ipv4: {
            method: networkForm.method,
            address: networkForm.address,
            netmask: networkForm.netmask,
            gateway: networkForm.gateway
          },
          ipv6: { method: 'auto' },
          dns: networkForm.dns.filter(d => d),
          dhcp: networkForm.method === 'dhcp'
        }
      };

      await settingsAPI.updateNetworkConfig(config);
      showNetworkModal = false;
      await loadNetworkInterfaces();
    } catch (err) {
      alert('Failed to update network config: ' + err.message);
    }
  }

  function getStatusColor(up) {
    return up ? 'bg-green-500' : 'bg-red-500';
  }
</script>

<div class="p-6">
  <div class="mb-6">
    <h1 class="text-xl font-bold text-gray-900 dark:text-white mb-2">
      Settings
    </h1>
    <p class="text-sm text-gray-600 dark:text-gray-300">
      Configure system and network settings
    </p>
  </div>

  {#if error}
    <div class="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
      <p class="text-sm text-red-600">{error}</p>
    </div>
  {/if}

  {#if success}
    <div class="mb-4 p-3 bg-green-50 border border-green-200 rounded-md">
      <p class="text-sm text-green-600">{success}</p>
    </div>
  {/if}

  <!-- Tabs -->
  <div class="border-b border-gray-200 dark:border-gray-700 mb-6">
    <nav class="-mb-px flex space-x-8">
      {#each tabs as tab}
        <button
          class="py-2 px-1 border-b-2 font-medium text-sm transition-colors flex items-center space-x-2"
          class:border-blue-500={activeTab === tab.id}
          class:text-blue-600={activeTab === tab.id}
          class:border-transparent={activeTab !== tab.id}
          class:text-gray-500={activeTab !== tab.id}
          class:hover:text-gray-700={activeTab !== tab.id}
          class:dark:text-gray-400={activeTab !== tab.id}
          class:dark:hover:text-gray-300={activeTab !== tab.id}
          class:hover:border-gray-300={activeTab !== tab.id}
          class:dark:hover:border-gray-600={activeTab !== tab.id}
          on:click={() => activeTab = tab.id}
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={tab.icon} />
          </svg>
          {tab.label}
        </button>
      {/each}
    </nav>
  </div>

  <!-- General Tab -->
  {#if activeTab === 'general'}
    <div class="max-w-2xl">
      <div class="card">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
          General Settings
        </h2>

        {#if loading}
          <div class="flex justify-center py-8">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          </div>
        {:else}
          <div class="space-y-4">
            <!-- Hostname -->
            <div>
              <label for="hostname" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Hostname
              </label>
              <input
                id="hostname"
                type="text"
                bind:value={settings.hostname}
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-muted text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500"
                placeholder="e.g. nas-server"
              />
            </div>

            <!-- Timezone -->
            <div>
              <label for="timezone" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Timezone
              </label>
              <select
                id="timezone"
                bind:value={settings.timezone}
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-muted text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500"
              >
                <option value="">Select timezone...</option>
                {#each timezoneList as tz}
                  <option value={tz}>{tz}</option>
                {/each}
              </select>
            </div>

            <!-- Locale -->
            <div>
              <label for="locale" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Locale
              </label>
              <select
                id="locale"
                bind:value={settings.locale}
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-muted text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500"
              >
                <option value="en_US.UTF-8">English (US)</option>
                <option value="en_GB.UTF-8">English (UK)</option>
                <option value="de_DE.UTF-8">German (Germany)</option>
                <option value="fr_FR.UTF-8">French (France)</option>
                <option value="es_ES.UTF-8">Spanish (Spain)</option>
                <option value="ja_JP.UTF-8">Japanese (Japan)</option>
                <option value="zh_CN.UTF-8">Chinese (Simplified)</option>
              </select>
            </div>

            <!-- Save Button -->
            <div class="pt-4 border-t border-gray-200 dark:border-gray-700">
              <button
                on:click={saveSettings}
                disabled={saving}
                class="btn btn-primary"
              >
                {#if saving}
                  Saving...
                {:else}
                  Save Settings
                {/if}
              </button>
            </div>
          </div>
        {/if}
      </div>
    </div>
  {:else if activeTab === 'network'}
    <!-- Network Tab -->
    <div>
      <div class="mb-4 flex justify-between items-center">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
          Network Interfaces
        </h2>
        <button
          on:click={loadNetworkInterfaces}
          class="btn btn-secondary text-sm"
        >
          Refresh
        </button>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {#each networkInterfaces as iface}
          <div class="card">
            <div class="flex items-start justify-between mb-4">
              <div class="flex items-center space-x-2">
                <div class="w-3 h-3 rounded-full {getStatusColor(iface.up)}"></div>
                <h3 class="font-semibold text-gray-900 dark:text-white">
                  {iface.name}
                </h3>
              </div>
              <span class="text-xs px-2 py-1 rounded bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300">
                {iface.type}
              </span>
            </div>

            <div class="space-y-2 text-sm">
              <div class="flex justify-between">
                <span class="text-gray-500 dark:text-gray-400">Status:</span>
                <span class="font-medium text-gray-900 dark:text-white">
                  {iface.up ? 'Up' : 'Down'}
                </span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500 dark:text-gray-400">MAC:</span>
                <span class="font-mono text-gray-900 dark:text-white text-xs">
                  {iface.mac}
                </span>
              </div>
              {#if iface.ipv4.address}
                <div class="flex justify-between">
                  <span class="text-gray-500 dark:text-gray-400">IPv4:</span>
                  <span class="font-mono text-gray-900 dark:text-white text-xs">
                    {iface.ipv4.address}
                  </span>
                </div>
              {/if}
              {#if iface.ipv6.address}
                <div class="flex justify-between">
                  <span class="text-gray-500 dark:text-gray-400">IPv6:</span>
                  <span class="font-mono text-gray-900 dark:text-white text-xs">
                    {iface.ipv6.address}
                  </span>
                </div>
              {/if}
              {#if iface.ipv4.gateway}
                <div class="flex justify-between">
                  <span class="text-gray-500 dark:text-gray-400">Gateway:</span>
                  <span class="font-mono text-gray-900 dark:text-white text-xs">
                    {iface.ipv4.gateway}
                  </span>
                </div>
              {/if}
              {#if iface.dns && iface.dns.length > 0}
                <div class="flex justify-between">
                  <span class="text-gray-500 dark:text-gray-400">DNS:</span>
                  <span class="font-mono text-gray-900 dark:text-white text-xs">
                    {iface.dns[0]}
                  </span>
                </div>
              {/if}
            </div>

            <div class="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
              <button
                on:click={() => openNetworkModal(iface)}
                class="btn btn-secondary text-sm w-full"
              >
                Configure
              </button>
            </div>
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>

<!-- Network Configuration Modal -->
{#if showNetworkModal && selectedInterface}
  <div class="fixed inset-0 z-50 overflow-y-auto" role="dialog" aria-modal="true" aria-labelledby="modal-title">
    <div class="flex items-center justify-center min-h-screen px-4">
      <div
        class="fixed inset-0 bg-gray-500 bg-opacity-75"
        role="button"
        tabindex="0"
        aria-label="Close modal"
        on:click={() => (showNetworkModal = false)}
        on:keydown={(e) => e.key === 'Enter' && (showNetworkModal = false)}
      ></div>
      <div class="relative bg-white dark:bg-card rounded-lg text-left overflow-hidden shadow-xl max-w-md w-full">
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <h3 id="modal-title" class="text-lg font-medium text-gray-900 dark:text-white">
              Configure {selectedInterface.name}
            </h3>
            <button
              on:click={() => (showNetworkModal = false)}
              class="text-gray-400 hover:text-gray-500"
              aria-label="Close modal"
            >
              <svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <div class="space-y-4">
            <!-- Configuration Method -->
            <div>
              <span class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Configuration Method
              </span>
              <div class="flex space-x-4">
                <label class="flex items-center">
                  <input
                    type="radio"
                    bind:group={networkForm.method}
                    value="dhcp"
                    class="mr-2"
                  />
                  <span class="text-sm text-gray-700 dark:text-gray-300">DHCP</span>
                </label>
                <label class="flex items-center">
                  <input
                    type="radio"
                    bind:group={networkForm.method}
                    value="static"
                    class="mr-2"
                  />
                  <span class="text-sm text-gray-700 dark:text-gray-300">Static</span>
                </label>
              </div>
            </div>

            {#if networkForm.method === 'static'}
              <!-- Static IP Configuration -->
              <div class="space-y-3">
                <div>
                  <label for="net-address" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    IP Address
                  </label>
                  <input
                    id="net-address"
                    type="text"
                    bind:value={networkForm.address}
                    placeholder="192.168.1.100"
                    class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-muted text-gray-900 dark:text-white"
                  />
                </div>
                <div>
                  <label for="net-netmask" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Netmask
                  </label>
                  <input
                    id="net-netmask"
                    type="text"
                    bind:value={networkForm.netmask}
                    placeholder="255.255.255.0"
                    class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-muted text-gray-900 dark:text-white"
                  />
                </div>
                <div>
                  <label for="net-gateway" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Gateway
                  </label>
                  <input
                    id="net-gateway"
                    type="text"
                    bind:value={networkForm.gateway}
                    placeholder="192.168.1.1"
                    class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-muted text-gray-900 dark:text-white"
                  />
                </div>
              </div>
            {/if}

            <!-- DNS Servers -->
            <div>
              <span class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                DNS Servers
              </span>
              {#each networkForm.dns as dns, index}
                <div class="flex space-x-2 mb-2">
                  <input
                    type="text"
                    bind:value={networkForm.dns[index]}
                    placeholder="8.8.8.8"
                    class="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-muted text-gray-900 dark:text-white text-sm"
                  />
                  <button
                    on:click={() => networkForm.dns = networkForm.dns.filter((_, i) => i !== index)}
                    class="px-2 py-1 text-red-600 hover:text-red-800"
                  >
                    Remove
                  </button>
                </div>
              {/each}
              <button
                on:click={() => networkForm.dns = [...networkForm.dns, '']}
                class="text-sm text-blue-600 hover:text-blue-800"
              >
                + Add DNS Server
              </button>
            </div>
          </div>

          <div class="mt-6 flex justify-end space-x-3">
            <button
              on:click={() => (showNetworkModal = false)}
              class="btn btn-secondary"
            >
              Cancel
            </button>
            <button
              on:click={saveNetworkConfig}
              class="btn btn-primary"
            >
              Apply
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  .btn {
    display: inline-flex;
    align-items: center;
    padding: 0.5rem 1rem;
    border: 1px solid transparent;
    border-radius: 0.375rem;
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
    font-size: 0.875rem;
    font-weight: 500;
    color: white;
    transition: color 0.2s, background-color 0.2s, border-color 0.2s;
    outline: none;
  }

  .btn:focus {
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.5);
  }

  .btn-primary {
    background-color: rgb(37, 99, 235);
    border-color: rgb(37, 99, 235);
  }

  .btn-primary:hover {
    background-color: rgb(29, 78, 216);
  }

  .btn-secondary {
    background-color: rgb(75, 85, 99);
    border-color: rgb(75, 85, 99);
  }

  .btn-secondary:hover {
    background-color: rgb(55, 65, 81);
  }
</style>
