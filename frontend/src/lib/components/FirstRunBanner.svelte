<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { createEventDispatcher } from 'svelte';
  import { onMount } from 'svelte';

  export let show = false;

  const dispatch = createEventDispatcher();

  let dismissed = false;

  // Check localStorage on mount
  onMount(() => {
    const dontShowAgain = localStorage.getItem('arcanas-hide-first-run');
    if (dontShowAgain === 'true') {
      dismissed = true;
      show = false;
    }
  });

  function dismiss() {
    dismissed = true;
    show = false;
  }

  function dontShowAgain() {
    localStorage.setItem('arcanas-hide-first-run', 'true');
    dismiss();
  }

  function launchWizard() {
    dispatch('launch-wizard');
    dismiss();
  }
</script>

{#if show && !dismissed}
  <div class="fixed top-4 left-1/2 transform -translate-x-1/2 z-50 max-w-2xl w-full px-4">
    <div
      class="bg-gradient-to-r from-indigo-50 to-purple-50 dark:from-indigo-900/40 dark:to-purple-900/40 border border-indigo-200 dark:border-indigo-700 rounded-lg shadow-lg p-6"
      role="alert"
    >
      <div class="flex items-start justify-between">
        <div class="flex items-start space-x-4">
          <!-- Icon -->
          <div class="flex-shrink-0">
            <div class="w-12 h-12 bg-gradient-to-br from-indigo-500 to-purple-500 rounded-full flex items-center justify-center">
              <svg class="w-6 h-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
              </svg>
            </div>
          </div>

          <!-- Content -->
          <div class="flex-1">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
              Welcome to Arcanas! Let's set up your storage.
            </h3>
            <p class="text-sm text-gray-600 dark:text-gray-300 mb-4">
              We detected you haven't configured any storage yet. Our Storage Builder can guide you
              through the whole process in just a few minutes.
            </p>

            <!-- Actions -->
            <div class="flex flex-wrap items-center gap-3">
              <button
                on:click={launchWizard}
                class="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
              >
                Launch Storage Builder
              </button>
              <button
                on:click={dismiss}
                class="px-4 py-2 text-gray-700 dark:text-gray-300 text-sm font-medium hover:bg-gray-100 dark:hover:bg-gray-700 rounded-md focus:outline-none"
              >
                I'll configure manually →
              </button>
            </div>
          </div>
        </div>

        <!-- Close button -->
        <div class="flex items-start space-x-2">
          <button
            on:click={dismiss}
            type="button"
            class="text-gray-400 hover:text-gray-500 dark:hover:text-gray-300 focus:outline-none"
            aria-label="Dismiss"
          >
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>

      <!-- Don't show again checkbox -->
      <div class="mt-4 pt-4 border-t border-indigo-200 dark:border-indigo-700">
        <label class="flex items-center space-x-2 cursor-pointer">
          <input
            type="checkbox"
            on:change={dontShowAgain}
            class="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500 dark:bg-gray-700 dark:border-gray-600"
          />
          <span class="text-xs text-gray-600 dark:text-gray-400">
            Don't show this again (I know what I'm doing)
          </span>
        </label>
      </div>
    </div>
  </div>
{/if}
