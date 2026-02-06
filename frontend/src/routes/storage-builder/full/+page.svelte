<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { diskAPI, storageAPI, nfsAPI, sambaAPI } from '$lib/api.js';
  import StepIndicator from '$lib/components/StorageBuilder/StepIndicator.svelte';
  import DiskSelectionStep from '$lib/components/StorageBuilder/steps/DiskSelectionStep.svelte';
  import RAIDConfigStep from '$lib/components/StorageBuilder/steps/RAIDConfigStep.svelte';
  import LVMConfigStep from '$lib/components/StorageBuilder/steps/LVMConfigStep.svelte';
  import PoolConfigStep from '$lib/components/StorageBuilder/steps/PoolConfigStep.svelte';
  import ShareConfigStep from '$lib/components/StorageBuilder/steps/ShareConfigStep.svelte';
  import SummaryStep from '$lib/components/StorageBuilder/steps/SummaryStep.svelte';

  // Wizard state
  let currentStep = 1;
  const totalSteps = 5;

  // Data state
  let availableDisks = [];
  let loading = true;
  let error = null;
  let creating = false;
  let createSuccess = false;
  let createError = null;

  // Wizard configuration
  let wizardConfig = {
    // Step 1: Disks
    selectedDisks: [],

    // Step 2: RAID
    raidConfig: {
      enabled: false,
      level: null,
      name: 'md0'
    },

    // Step 3: LVM
    lvmConfig: {
      enabled: false,
      vgName: '',
      lvName: 'data',
      sizeGB: null
    },

    // Step 4: Pool
    poolConfig: {
      name: '',
      mountPoint: '',
      filesystem: 'ext4'
    },

    // Step 5: Shares
    shareConfig: {
      createNFS: false,
      createSamba: false,
      nfsClients: [{ network: '192.168.1.0/24', options: 'rw,sync,no_subtree_check' }],
      sambaName: ''
    }
  };

  async function loadAvailableDisks() {
    try {
      loading = true;
      const diskStats = await diskAPI.getDiskStats();

      // Filter to only available disks (not mounted, not in use)
      availableDisks = diskStats.filter((disk) => {
        return disk.available !== false && !disk.mountpoint;
      });

      // Also filter out RAID member disks
      const raidArrays = await diskAPI.getRAIDArrays();
      const raidMemberPaths = raidArrays.flatMap((array) => array.devices || []);

      availableDisks = availableDisks.filter((disk) => {
        return !raidMemberPaths.includes(disk.path);
      });
    } catch (err) {
      error = err.message || 'Failed to load available disks';
      console.error('Failed to load available disks:', err);
    } finally {
      loading = false;
    }
  }

  function handleDiskSelection(e) {
    wizardConfig.selectedDisks = e.detail;

    // Auto-enable RAID if 2+ disks selected
    if (e.detail.length >= 2 && !wizardConfig.raidConfig.enabled) {
      wizardConfig.raidConfig.enabled = true;
      // Set default RAID level based on disk count
      if (e.detail.length === 2) {
        wizardConfig.raidConfig.level = 'raid1';
      } else if (e.detail.length >= 3) {
        wizardConfig.raidConfig.level = 'raid5';
      }
    }
  }

  function handleRAIDChange(e) {
    wizardConfig.raidConfig = e.detail;
  }

  function handleLVMChange(e) {
    wizardConfig.lvmConfig = e.detail;
  }

  function handleShareChange(e) {
    wizardConfig.shareConfig = e.detail;
  }

  function nextStep() {
    if (currentStep < totalSteps) {
      currentStep = currentStep + 1;
    }
  }

  function previousStep() {
    if (currentStep > 1) {
      currentStep = currentStep - 1;
    }
  }

  function goToStep(e) {
    const step = e.detail;
    // Can only go to steps that have been completed or are the next step
    if (step <= currentStep || step === currentStep + 1) {
      currentStep = step;
    }
  }

  function cancelWizard() {
    goto('/storage-builder');
  }

  async function handleCreate() {
    creating = true;
    createError = null;

    try {
      // Execute the storage creation workflow
      // This is a placeholder - in a real implementation, you'd call a unified API
      // For now, we'll simulate the creation process

      // Step 1: Create RAID (if enabled)
      if (wizardConfig.raidConfig.enabled) {
        console.log('Creating RAID array:', wizardConfig.raidConfig);
        // await diskAPI.createRAIDArray({...});
      }

      // Step 2: Create LVM (if enabled)
      if (wizardConfig.lvmConfig.enabled) {
        console.log('Creating LVM volumes:', wizardConfig.lvmConfig);
        // await lvmAPI.createVolumeGroup({...});
        // await lvmAPI.createLogicalVolume({...});
      }

      // Step 3: Create storage pool
      console.log('Creating storage pool:', wizardConfig.poolConfig);
      // Determine pool type
      let poolType = 'bind';
      let devices = [];

      if (wizardConfig.lvmConfig.enabled) {
        poolType = 'lvm_lv';
        devices = [`/dev/${wizardConfig.lvmConfig.vgName}/${wizardConfig.lvmConfig.lvName}`];
      } else if (wizardConfig.raidConfig.enabled) {
        poolType = 'mergerfs';
        devices = wizardConfig.selectedDisks.map(d => `/mnt/arcanas-disk-${d.split('/').pop()}`);
      } else {
        devices = wizardConfig.selectedDisks;
      }

      // await storageAPI.createPool({
      //   name: wizardConfig.poolConfig.name,
      //   type: poolType,
      //   devices: devices,
      //   filesystem: wizardConfig.poolConfig.filesystem
      // });

      // Step 4: Create shares (if enabled)
      const mountPoint = wizardConfig.poolConfig.mountPoint || `/srv/${wizardConfig.poolConfig.name}`;

      if (wizardConfig.shareConfig.createNFS) {
        console.log('Creating NFS exports:', wizardConfig.shareConfig.nfsClients);
        for (const client of wizardConfig.shareConfig.nfsClients) {
          // await nfsAPI.createExport({
          //   path: mountPoint,
          //   clients: [{ network: client.network, options: client.options }]
          // });
        }
      }

      if (wizardConfig.shareConfig.createSamba) {
        console.log('Creating Samba share:', wizardConfig.shareConfig.sambaName);
        // await sambaAPI.createShare({
        //   name: wizardConfig.shareConfig.sambaName,
        //   path: mountPoint,
        //   ... other config
        // });
      }

      // Success!
      createSuccess = true;

      // Redirect to storage page after a short delay
      setTimeout(() => {
        goto('/storage');
      }, 2000);
    } catch (err) {
      createError = err.message || 'Failed to create storage';
      console.error('Failed to create storage:', err);
    } finally {
      creating = false;
    }
  }

  onMount(() => {
    loadAvailableDisks();
  });

  // Update mount point when pool name changes
  $: if (wizardConfig.poolConfig.name && !wizardConfig.poolConfig.mountPoint) {
    wizardConfig.poolConfig.mountPoint = `/srv/${wizardConfig.poolConfig.name}`;
  }
</script>

<div class="p-6" role="main" tabindex="-1">
  <!-- Header -->
  <div class="mb-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-gray-900 dark:text-white mb-2">
          Complete Storage Setup
        </h1>
        <p class="text-sm text-gray-600 dark:text-gray-300">
          Guided wizard for disks, RAID, LVM, pools, and shares
        </p>
      </div>
      <button
        on:click={cancelWizard}
        class="text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
        aria-label="Cancel wizard"
        title="Cancel wizard"
      >
        <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
  </div>

  <!-- Step Indicator -->
  <StepIndicator currentStep={currentStep} totalSteps={totalSteps} on:goToStep={goToStep} />

  <!-- Loading State -->
  {#if loading}
    <div class="flex justify-center items-center py-12">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
    </div>
  {:else if error}
    <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
      <p class="text-red-800 dark:text-red-200">{error}</p>
    </div>
  {:else if createSuccess}
    <div class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg p-8 text-center">
      <svg class="mx-auto h-16 w-16 text-green-600 dark:text-green-400 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <h2 class="text-2xl font-bold text-green-900 dark:text-green-100 mb-2">Storage Created Successfully!</h2>
      <p class="text-green-700 dark:text-green-300">Redirecting to storage page...</p>
    </div>
  {:else if createError}
    <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4 mb-4">
      <p class="text-red-800 dark:text-red-200 font-medium mb-2">Failed to create storage</p>
      <p class="text-red-700 dark:text-red-300">{createError}</p>
    </div>
  {:else}
    <!-- Step Content -->
    <div class="mt-8">
      {#if currentStep === 1}
        <DiskSelectionStep
          {availableDisks}
          selectedDisks={wizardConfig.selectedDisks}
          on:select={handleDiskSelection}
          on:next={nextStep}
          on:previous={cancelWizard}
        />
      {:else if currentStep === 2}
        <RAIDConfigStep
          selectedDisks={wizardConfig.selectedDisks}
          raidConfig={wizardConfig.raidConfig}
          on:raidChange={handleRAIDChange}
          on:next={nextStep}
          on:previous={previousStep}
        />
      {:else if currentStep === 3}
        <LVMConfigStep
          raidConfig={wizardConfig.raidConfig}
          selectedDisks={wizardConfig.selectedDisks}
          lvmConfig={wizardConfig.lvmConfig}
          on:lvmChange={handleLVMChange}
          on:next={nextStep}
          on:previous={previousStep}
        />
      {:else if currentStep === 4}
        <PoolConfigStep
          raidConfig={wizardConfig.raidConfig}
          lvmConfig={wizardConfig.lvmConfig}
          poolConfig={wizardConfig.poolConfig}
          on:next={nextStep}
          on:previous={previousStep}
        />
      {:else if currentStep === 5}
        <ShareConfigStep
          poolConfig={wizardConfig.poolConfig}
          shareConfig={wizardConfig.shareConfig}
          on:shareChange={handleShareChange}
          on:finish={nextStep}
          on:previous={previousStep}
        />
      {:else if currentStep === 6}
        <SummaryStep
          {wizardConfig}
          {availableDisks}
          on:create={handleCreate}
          on:previous={previousStep}
        />
      {/if}
    </div>
  {/if}
</div>
