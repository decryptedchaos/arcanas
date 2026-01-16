<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import { auth } from '$lib/stores/auth.js';
  import { goto } from '$app/navigation';

  let username = '';
  let password = '';
  let errorMessage = '';

  async function handleLogin(event) {
    event.preventDefault();
    errorMessage = '';

    if (!username || !password) {
      errorMessage = 'Please enter both username and password';
      return;
    }

    const success = await auth.login(username, password);
    if (success) {
      // Redirect to dashboard
      goto('/');
    } else {
      errorMessage = $auth.error || 'Invalid username or password';
    }
  }

  function clearError() {
    errorMessage = '';
    auth.clearError();
  }
</script>

<div class="min-h-screen bg-gray-50 dark:bg-surface flex items-center justify-center px-4">
  <div class="max-w-md w-full">
    <!-- Logo/Branding -->
    <div class="text-center mb-8">
      <div class="inline-flex items-center justify-center w-16 h-16 bg-primary-600 rounded-lg shadow-lg mb-4">
        <span class="text-3xl font-bold text-white">A</span>
      </div>
      <h1 class="text-3xl font-bold text-gray-900 dark:text-white">Arcanas</h1>
      <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
        Storage Management Dashboard
      </p>
    </div>

    <!-- Login Form -->
    <div class="bg-white dark:bg-card rounded-lg shadow-md p-8">
      <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-6">
        Sign in to your account
      </h2>

      {#if errorMessage}
        <div class="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
          <p class="text-sm text-red-600">{errorMessage}</p>
        </div>
      {/if}

      <form on:submit={handleLogin}>
        <div class="space-y-4">
          <!-- Username -->
          <div>
            <label
              for="username"
              class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
            >
              Username
            </label>
            <input
              id="username"
              type="text"
              bind:value={username}
              on:input={clearError}
              autocomplete="username"
              required
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm bg-white dark:bg-muted text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="Enter your username"
            />
          </div>

          <!-- Password -->
          <div>
            <label
              for="password"
              class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
            >
              Password
            </label>
            <input
              id="password"
              type="password"
              bind:value={password}
              on:input={clearError}
              autocomplete="current-password"
              required
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm bg-white dark:bg-muted text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="Enter your password"
            />
          </div>

          <!-- Submit Button -->
          <button
            type="submit"
            disabled={$auth.loading}
            class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {#if $auth.loading}
              <svg
                class="animate-spin -ml-1 mr-3 h-5 w-5 text-white"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle
                  class="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  stroke-width="4"
                ></circle>
                <path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path>
              </svg>
              Signing in...
            {:else}
              Sign in
            {/if}
          </button>
        </div>
      </form>

      <div class="mt-6">
        <div class="relative">
          <div class="absolute inset-0 flex items-center">
            <div class="w-full border-t border-gray-300 dark:border-gray-600"></div>
          </div>
          <div class="relative flex justify-center text-sm">
            <span class="px-2 bg-white dark:bg-card text-gray-500 dark:text-gray-400">
              Use your system user account
            </span>
          </div>
        </div>

        <p class="mt-4 text-xs text-center text-gray-500 dark:text-gray-400">
          Root users have full access. Regular users can only manage shares they have
          access to.
        </p>
      </div>
    </div>

    <!-- Footer -->
    <p class="mt-8 text-center text-sm text-gray-500 dark:text-gray-400">
      Arcanas v{__APP_VERSION__} &bull; Mozilla Public License v2.0
    </p>
  </div>
</div>

<style>
  :global(body) {
    margin: 0;
  }
</style>
