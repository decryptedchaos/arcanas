<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
  import Sidebar from "$lib/components/Sidebar.svelte";
  import { onMount } from "svelte";
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { auth } from "$lib/stores/auth.js";
  import "../app.css";

  let sidebarOpen = true;
  // Initialize dark mode IMMEDIATELY from localStorage, not in onMount
  // This prevents flash by having correct value on first render
  const savedDarkMode = typeof window !== 'undefined' ? localStorage.getItem("darkMode") : null;
  let darkMode = savedDarkMode === null ? true : savedDarkMode === "true";

  // Save default preference if none exists
  if (typeof window !== 'undefined' && savedDarkMode === null) {
    localStorage.setItem("darkMode", "true");
  }

  // Public routes that don't require authentication
  const publicRoutes = ["/login"];

  $: darkModeClass = darkMode ? "dark" : "";

  $: if (darkMode) {
    document.documentElement.classList.add("dark");
  } else {
    document.documentElement.classList.remove("dark");
  }

  $: if (sidebarOpen !== undefined) {
    localStorage.setItem("sidebarOpen", sidebarOpen.toString());
  }

  onMount(async () => {
    // Validate session on mount
    await auth.validate();

    // Check if route requires authentication
    const currentPath = $page.url.pathname;
    const isPublicRoute = publicRoutes.includes(currentPath);

    // Redirect to login if not authenticated and not on public route
    if (!$auth.isAuthenticated && !isPublicRoute) {
      goto('/login');
      return;
    }

    // Redirect to dashboard if already authenticated and on login page
    if ($auth.isAuthenticated && currentPath === '/login') {
      goto('/');
      return;
    }

    // Check for saved sidebar preference
    const savedSidebarState = localStorage.getItem("sidebarOpen");
    if (savedSidebarState !== null) {
      sidebarOpen = savedSidebarState === "true";
    } else {
      // Auto-close sidebar on mobile only if no saved preference
      if (window.innerWidth < 768) {
        sidebarOpen = false;
      }
    }

    // Dark mode is now initialized at component level, not here
  });

  function toggleDarkMode() {
    darkMode = !darkMode;
    localStorage.setItem("darkMode", darkMode.toString());
  }
</script>

<svelte:head>
  <title>Arcanas</title>
  <meta
    name="description"
    content="Arcanas - Modern Storage Management Dashboard"
  />
</svelte:head>

<div class="h-screen bg-gray-50 dark:bg-surface flex overflow-hidden {darkModeClass}">
  <!-- Sidebar -->
  <Sidebar bind:sidebarOpen />

  <!-- Main Content -->
  <div class="flex-1 flex flex-col">
    <!-- Top Navigation -->
    <header
      class="bg-white dark:bg-card shadow-sm border-b border-gray-200 dark:border-gray-700 z-10 flex-shrink-0"
    >
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between items-center h-16">
          <div class="flex items-center">
            <button
              on:click={() => (sidebarOpen = !sidebarOpen)}
              class="md:hidden p-2 rounded-md text-gray-400 dark:text-gray-300 hover:text-gray-500 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-muted"
              aria-label="Toggle sidebar"
            >
              <svg
                class="h-6 w-6"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M4 6h16M4 12h16M4 18h16"
                />
              </svg>
            </button>
            <h1
              class="ml-4 text-2xl font-bold text-gray-900 dark:text-white tracking-wider font-mono"
            >
              ARCANAS
            </h1>
          </div>

          <div class="flex items-center space-x-4">
            <!-- Dark Mode Toggle -->
            <button
              on:click={toggleDarkMode}
              class="p-2 rounded-lg bg-gray-100 dark:bg-muted hover:bg-gray-200 dark:hover:bg-border transition-colors"
              title="Toggle dark mode"
            >
              {#if darkMode}
                <svg
                  class="w-5 h-5 text-yellow-500"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fill-rule="evenodd"
                    d="M10 2a1 1 0 011 1v1a1 1 0 11-2 0V3a1 1 0 011-1zm4 8a4 4 0 11-8 0 4 4 0 018 0zm-.464 4.95l.707.707a1 1 0 001.414-1.414l-.707-.707a1 1 0 00-1.414 1.414zm2.12-10.607a1 1 0 010 1.414l-.706.707a1 1 0 11-1.414-1.414l.707-.707a1 1 0 011.414 0zM17 11a1 1 0 100-2h-1a1 1 0 100 2h1zm-7 4a1 1 0 011 1v1a1 1 0 11-2 0v-1a1 1 0 011-1zM5.05 6.464A1 1 0 106.465 5.05l-.708-.707a1 1 0 00-1.414 1.414l.707.707zm1.414 8.486l-.707.707a1 1 0 01-1.414-1.414l.707-.707a1 1 0 011.414 1.414zM4 11a1 1 0 100-2H3a1 1 0 000 2h1z"
                    clip-rule="evenodd"
                  />
                </svg>
              {:else}
                <svg
                  class="w-5 h-5 text-gray-700 dark:text-gray-300"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    d="M17.293 13.293A8 8 0 016.707 2.707a8.001 8.001 0 1010.586 10.586z"
                  />
                </svg>
              {/if}
            </button>

            <div
              class="flex items-center space-x-2 text-sm text-gray-600 dark:text-gray-300"
            >
              <div
                class="w-2 h-2 bg-green-500 rounded-full animate-pulse"
              ></div>
              <span>System Online</span>
            </div>

            <!-- User Menu -->
            <div class="flex items-center space-x-3">
              <span
                class="text-sm font-medium text-gray-700 dark:text-gray-300"
              >
                {$auth.username}
              </span>
              <button
                on:click={async () => await auth.logout()}
                class="btn btn-secondary text-xs"
                title="Logout"
              >
                <svg
                  class="w-4 h-4"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
                  />
                </svg>
              </button>
            </div>

            <div class="relative">
              <button
                class="p-2 rounded-full hover:bg-gray-100 dark:hover:bg-muted"
                aria-label="Notifications"
              >
                <svg
                  class="h-5 w-5 text-gray-600 dark:text-gray-300"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
                  />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>
    </header>

    <!-- Page Content -->
    <main class="flex-1 overflow-auto">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <slot />
      </div>
    </main>
  </div>
</div>

<style>
  :global(body) {
    margin: 0;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto",
      "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans",
      "Helvetica Neue", sans-serif;
  }
</style>
