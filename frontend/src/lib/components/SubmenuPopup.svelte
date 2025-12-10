<!--
  This file is part of the Arcanas project.
  Licensed under the Mozilla Public License, v. 2.0.
  See the LICENSE file at the project root for details.
-->

<script>
    import { page } from "$app/stores";

    export let show = false;
    export let position = { top: 0, left: 0 };
    export let items = [];
    export let onClose = () => {};

    function handleClickOutside() {
        onClose();
    }

    function handleItemClick() {
        onClose();
    }
</script>

{#if show}
    <!-- Click outside to close -->
    <div
        class="fixed inset-0 z-40"
        role="button"
        tabindex="0"
        aria-label="Close submenu"
        on:click={handleClickOutside}
        on:keydown={(e) => e.key === "Enter" && handleClickOutside()}
    ></div>

    <!-- Submenu Popup -->
    <div
        class="fixed z-50 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 py-2 min-w-48"
        style="top: {position.top}px; left: {position.left}px;"
        role="menu"
        tabindex="-1"
        on:click|stopPropagation
        on:keydown={(e) => {
            if (e.key === "Escape") {
                handleClickOutside();
            }
        }}
    >
        {#each items as item}
            <a
                href={item.href}
                class="flex items-center px-3 py-2 text-sm font-medium transition-colors duration-200 {$page
                    .url.pathname === item.href
                    ? 'bg-primary-50 dark:bg-primary-900/20 text-primary-700 dark:text-primary-300'
                    : 'text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700'}"
                role="menuitem"
                on:click={handleItemClick}
            >
                <div class="flex-shrink-0 mr-3">
                    {@html item.icon}
                </div>
                {item.name}
            </a>
        {/each}
    </div>
{/if}
