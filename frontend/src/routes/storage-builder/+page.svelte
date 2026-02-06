<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import WorkflowSelector from '$lib/components/StorageBuilder/WorkflowSelector.svelte';
</script>

<div class="p-6" role="main" tabindex="-1">
  <!-- Header -->
  <div class="mb-6">
    <h1 class="text-xl font-bold text-gray-900 dark:text-white mb-2">
      Storage Builder
    </h1>
    <p class="text-sm text-gray-600 dark:text-gray-300">
      Choose a workflow or use the complete guided wizard
    </p>
  </div>

  <!-- Workflow Selector -->
  <WorkflowSelector />
</div>
