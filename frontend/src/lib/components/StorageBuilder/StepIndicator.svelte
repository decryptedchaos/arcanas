<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { createEventDispatcher } from 'svelte';

  export let currentStep = 1;
  export let totalSteps = 5;

  const dispatch = createEventDispatcher();

  const steps = [
    { number: 1, name: 'Disks', icon: '📀' },
    { number: 2, name: 'RAID', icon: '🛡️' },
    { number: 3, name: 'LVM', icon: '📦' },
    { number: 4, name: 'Pool', icon: '💾' },
    { number: 5, name: 'Shares', icon: '🔗' }
  ];

  function handleStepClick(step) {
    dispatch('goToStep', step);
  }
</script>

<div class="mb-8">
  <nav aria-label="Progress">
    <ol class="flex items-center justify-between">
      {#each steps as step, index}
        <li class="flex items-center {index < steps.length - 1 ? 'flex-1' : ''}">
          <div class="flex items-center w-full">
            <button
              on:click={() => handleStepClick(step.number)}
              class="flex items-center justify-center w-10 h-10 rounded-full flex-shrink-0
                {step.number === currentStep
                  ? 'bg-indigo-600 text-white'
                  : step.number < currentStep
                  ? 'bg-green-600 text-white cursor-pointer hover:bg-green-700'
                  : 'bg-gray-200 dark:bg-gray-700 text-gray-500 dark:text-gray-400 cursor-not-allowed'}
                focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              aria-current={step.number === currentStep ? 'step' : undefined}
              disabled={step.number > currentStep}
              title={step.name}
            >
              {step.number < currentStep ? '✓' : step.icon}
            </button>

            <span class="ml-2 text-sm font-medium {step.number === currentStep ? 'text-indigo-600 dark:text-indigo-400' : step.number < currentStep ? 'text-green-600 dark:text-green-400' : 'text-gray-500 dark:text-gray-400'} hidden sm:block">
              {step.name}
            </span>

            {#if index < steps.length - 1}
              <div class="flex-1 mx-4 h-1 bg-gray-200 dark:bg-gray-700 rounded {step.number < currentStep ? 'bg-green-600' : ''}" style="min-width: 3rem; max-width: 6rem;"></div>
            {/if}
          </div>
        </li>
      {/each}
    </ol>
  </nav>

  <!-- Mobile step counter -->
  <div class="mt-4 text-center sm:hidden">
    <p class="text-sm text-gray-600 dark:text-gray-400">
      Step {currentStep} of {totalSteps}
    </p>
  </div>
</div>

<style>
  /* Progress line animation */
  ol li .flex-1 {
    transition: background-color 0.3s ease;
  }
</style>
