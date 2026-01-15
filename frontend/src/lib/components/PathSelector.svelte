<script>
    import { storageAPI } from "$lib/api.js";
    import { onMount } from "svelte";

    export let path = "";
    export let label = "Path";
    export let required = true;
    export let id = "path-selector";

    let pools = [];
    let loading = true;
    let version = Date.now(); // Force cache busting

    onMount(async () => {
        console.log("PathSelector mounting at:", new Date().toISOString());
        try {
            const poolsData = await storageAPI.getPools();
            console.log("Loaded pools:", poolsData);
            pools = poolsData;
        } catch (e) {
            console.error("Failed to load paths", e);
        } finally {
            loading = false;
            console.log("PathSelector loading complete");
        }
    });
</script>

<div>
    <label
        for={id}
        class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
        >{label}</label
    >

    {#if loading}
        <div
            class="animate-pulse h-10 bg-gray-200 dark:bg-muted rounded-md"
        ></div>
        <div class="text-xs text-red-500 mt-1">DEBUG: Loading pools...</div>
    {:else}
        <div class="text-xs text-blue-500 mb-1">
            DEBUG: Found {pools.length} pools - UPDATED VERSION 2.0!
        </div>
        <select
            {id}
            bind:value={path}
            class="w-full px-3 py-2 border-4 border-red-500 bg-red-100 dark:bg-red-900 rounded-md focus:outline-none focus:ring-4 focus:ring-red-500 text-red-900 dark:text-red-100"
            {required}
        >
            <option value="" disabled>Select a storage pool...</option>

            {#if pools.length > 0}
                {#each pools as pool}
                    <option value={pool.mount_point}>
                        {pool.name} ({pool.type}) - {(
                            pool.size /
                            1024 /
                            1024 /
                            1024
                        ).toFixed(1)} GB
                    </option>
                {/each}
            {:else}
                <option value="" disabled>No storage pools available</option>
            {/if}
        </select>
    {/if}
</div>
