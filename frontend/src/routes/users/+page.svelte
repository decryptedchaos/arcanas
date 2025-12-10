<script>
    import { usersAPI } from "$lib/api.js";
    import { onMount } from "svelte";

    let users = [];
    let loading = true;
    let error = null;
    let selectedService = "all";
    let showCreateModal = false;
    let showEditModal = false;
    let showServicesModal = false;
    let selectedUser = null;

    // Form data for new user
    let newUser = {
        username: "",
        name: "",
        password: "",
        groups: "",
        shell: "/bin/bash",
    };

    onMount(async () => {
        await loadUsers();
    });

    async function loadUsers() {
        try {
            loading = true;
            error = null;
            users = await usersAPI.getUsers();
        } catch (err) {
            error = err.message || "Failed to load users";
            console.error("Error loading users:", err);
        } finally {
            loading = false;
        }
    }

    function filteredUsers() {
        if (selectedService === "all") return users;

        return users.filter((user) => {
            switch (selectedService) {
                case "samba":
                    return user.services.samba;
                case "nfs":
                    return user.services.nfs;
                case "ssh":
                    return user.services.ssh;
                default:
                    return true;
            }
        });
    }

    function getServiceBadge(service, enabled) {
        const colors = {
            samba: enabled
                ? "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
                : "bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300",
            nfs: enabled
                ? "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200"
                : "bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300",
            ssh: enabled
                ? "bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200"
                : "bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300",
        };
        return colors[service] || colors.samba;
    }

    function getServiceCount(service) {
        if (service === "all") return users.length;
        return users.filter((user) => user.services[service]).length;
    }

    function openEditModal(user) {
        selectedUser = {
            ...user,
            groups: Array.isArray(user.groups)
                ? user.groups.join(", ")
                : user.groups,
            password: "",
            passwordConfirm: "",
        };
        showEditModal = true;
    }

    function openServicesModal(user) {
        selectedUser = user;
        showServicesModal = true;
    }

    async function createUser() {
        try {
            console.log("Creating user:", newUser);
            const userData = {
                username: newUser.username,
                name: newUser.name,
                password: newUser.password,
                groups: newUser.groups
                    .split(",")
                    .map((g) => g.trim())
                    .filter((g) => g),
                shell: newUser.shell,
            };

            console.log("Sending user data:", userData);
            await usersAPI.createUser(userData);
            console.log("User created successfully");
            showCreateModal = false;
            resetNewUserForm();
            await loadUsers();
        } catch (err) {
            console.error("Failed to create user:", err);
            error = err.message || "Failed to create user";
            // Don't close modal on error so user can try again
        }
    }

    async function updateUser() {
        try {
            console.log("Updating user:", selectedUser);

            // Password validation
            if (
                selectedUser.password &&
                selectedUser.password !== selectedUser.passwordConfirm
            ) {
                error = "Passwords do not match";
                return;
            }

            const userData = {
                name: selectedUser.name,
                groups: Array.isArray(selectedUser.groups)
                    ? selectedUser.groups
                    : selectedUser.groups
                          .split(",")
                          .map((g) => g.trim())
                          .filter((g) => g),
            };

            // Only include password if it's provided (not blank)
            if (selectedUser.password) {
                userData.password = selectedUser.password;
            }

            console.log("Sending update data:", userData);
            await usersAPI.updateUser(selectedUser.username, userData);
            console.log("User updated successfully");
            showEditModal = false;
            await loadUsers();
        } catch (err) {
            console.error("Failed to update user:", err);
            error = err.message || "Failed to update user";
            // Don't close modal on error so user can try again
        }
    }

    async function updateUserServices() {
        try {
            console.log(
                "Updating services for:",
                selectedUser.username,
                selectedUser.services,
            );
            await usersAPI.updateUserServices(
                selectedUser.username,
                selectedUser.services,
            );
            console.log("Services updated successfully");
            showServicesModal = false;
            await loadUsers();
        } catch (err) {
            console.error("Failed to update user services:", err);
            error = err.message || "Failed to update user services";
            // Don't close modal on error so user can try again
        }
    }

    function resetNewUserForm() {
        newUser = {
            username: "",
            name: "",
            password: "",
            groups: "",
            shell: "/bin/bash",
        };
    }
</script>

<div class="p-6">
    <div class="mb-6">
        <h1 class="text-xl font-bold text-gray-900 dark:text-white mb-2">
            User Management
        </h1>
        <p class="text-sm text-gray-600 dark:text-gray-300">
            Manage system users and their service access
        </p>
    </div>

    <div class="flex justify-between items-center mb-6">
        <button
            on:click={loadUsers}
            class="btn btn-secondary"
            disabled={loading}
        >
            Refresh
        </button>
        <button
            on:click={() => (showCreateModal = true)}
            class="btn btn-primary"
        >
            Create User
        </button>
    </div>

    {#if loading}
        <div class="flex justify-center items-center h-64">
            <div
                class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"
            ></div>
        </div>
    {:else if error}
        <div class="bg-red-50 border border-red-200 rounded-md p-4">
            <div class="flex">
                <div class="flex-shrink-0">
                    <svg
                        class="h-5 w-5 text-red-400"
                        viewBox="0 0 20 20"
                        fill="currentColor"
                    >
                        <path
                            fill-rule="evenodd"
                            d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                            clip-rule="evenodd"
                        />
                    </svg>
                </div>
                <div class="ml-3">
                    <h3 class="text-sm font-medium text-red-800">Error</h3>
                    <div class="mt-2 text-sm text-red-700">{error}</div>
                </div>
            </div>
        </div>
    {:else}
        <!-- Service Filter Tabs -->
        <div class="mb-6 border-b border-gray-200 dark:border-gray-700">
            <nav class="-mb-px flex space-x-8">
                <button
                    class="py-2 px-1 border-b-2 font-medium text-sm transition-colors"
                    class:border-blue-500={selectedService === "all"}
                    class:text-blue-600={selectedService === "all"}
                    class:border-transparent={selectedService !== "all"}
                    class:text-gray-500={selectedService !== "all"}
                    class:dark:text-gray-400={selectedService !== "all"}
                    class:hover:text-gray-700={selectedService !== "all"}
                    class:dark:hover:text-gray-300={selectedService !== "all"}
                    class:hover:border-gray-300={selectedService !== "all"}
                    class:dark:hover:border-gray-600={selectedService !== "all"}
                    on:click={() => (selectedService = "all")}
                >
                    All Users ({getServiceCount("all")})
                </button>
                <button
                    class="py-2 px-1 border-b-2 font-medium text-sm transition-colors"
                    class:border-blue-500={selectedService === "samba"}
                    class:text-blue-600={selectedService === "samba"}
                    class:border-transparent={selectedService !== "samba"}
                    class:text-gray-500={selectedService !== "samba"}
                    class:hover:text-gray-700={selectedService !== "samba"}
                    class:hover:border-gray-300={selectedService !== "samba"}
                    on:click={() => (selectedService = "samba")}
                >
                    Samba Users ({getServiceCount("samba")})
                </button>
                <button
                    class="py-2 px-1 border-b-2 font-medium text-sm transition-colors"
                    class:border-blue-500={selectedService === "nfs"}
                    class:text-blue-600={selectedService === "nfs"}
                    class:border-transparent={selectedService !== "nfs"}
                    class:text-gray-500={selectedService !== "nfs"}
                    class:hover:text-gray-700={selectedService !== "nfs"}
                    class:hover:border-gray-300={selectedService !== "nfs"}
                    on:click={() => (selectedService = "nfs")}
                >
                    NFS Users ({getServiceCount("nfs")})
                </button>
                <button
                    class="py-2 px-1 border-b-2 font-medium text-sm transition-colors"
                    class:border-blue-500={selectedService === "ssh"}
                    class:text-blue-600={selectedService === "ssh"}
                    class:border-transparent={selectedService !== "ssh"}
                    class:text-gray-500={selectedService !== "ssh"}
                    class:hover:text-gray-700={selectedService !== "ssh"}
                    class:hover:border-gray-300={selectedService !== "ssh"}
                    on:click={() => (selectedService = "ssh")}
                >
                    SSH Users ({getServiceCount("ssh")})
                </button>
            </nav>
        </div>

        <!-- Users Table -->
        <div
            class="bg-white dark:bg-gray-800 shadow overflow-hidden sm:rounded-md"
        >
            <ul class="divide-y divide-gray-200 dark:divide-gray-700">
                {#each filteredUsers() as user (user.username)}
                    <li class="hover:bg-gray-50 dark:hover:bg-gray-700">
                        <div class="px-4 py-4 sm:px-6">
                            <div class="flex items-center justify-between">
                                <div class="flex items-center">
                                    <div class="flex-shrink-0">
                                        <div
                                            class="h-10 w-10 rounded-full bg-gray-300 dark:bg-gray-600 flex items-center justify-center"
                                        >
                                            <span
                                                class="text-sm font-medium text-gray-700 dark:text-gray-200"
                                            >
                                                {user.name
                                                    .charAt(0)
                                                    .toUpperCase()}
                                            </span>
                                        </div>
                                    </div>
                                    <div class="ml-4">
                                        <div
                                            class="text-sm font-medium text-gray-900 dark:text-white"
                                        >
                                            {user.name}
                                        </div>
                                        <div
                                            class="text-sm text-gray-500 dark:text-gray-400"
                                        >
                                            @{user.username} • UID: {user.uid} •
                                            {user.home_dir}
                                        </div>
                                        <div class="mt-1 flex flex-wrap gap-1">
                                            {#each user.groups as group}
                                                <span
                                                    class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200"
                                                >
                                                    {group}
                                                </span>
                                            {/each}
                                        </div>
                                    </div>
                                </div>
                                <div class="flex items-center space-x-2">
                                    <!-- Service Badges -->
                                    <span
                                        class="inline-flex items-center px-2 py-1 rounded text-xs font-medium {getServiceBadge(
                                            'samba',
                                            user.services.samba,
                                        )}"
                                    >
                                        Samba
                                    </span>
                                    <span
                                        class="inline-flex items-center px-2 py-1 rounded text-xs font-medium {getServiceBadge(
                                            'nfs',
                                            user.services.nfs,
                                        )}"
                                    >
                                        NFS
                                    </span>
                                    <span
                                        class="inline-flex items-center px-2 py-1 rounded text-xs font-medium {getServiceBadge(
                                            'ssh',
                                            user.services.ssh,
                                        )}"
                                    >
                                        SSH
                                    </span>

                                    <!-- Actions -->
                                    <div
                                        class="flex items-center space-x-1 ml-4"
                                    >
                                        <button
                                            class="p-1 text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 transition-colors"
                                            title="Edit user"
                                            on:click={() => openEditModal(user)}
                                        >
                                            <svg
                                                class="h-4 w-4"
                                                fill="none"
                                                stroke="currentColor"
                                                viewBox="0 0 24 24"
                                            >
                                                <path
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    stroke-width="2"
                                                    d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                                                />
                                            </svg>
                                        </button>
                                        <button
                                            class="p-1 text-gray-400 hover:text-green-600 dark:hover:text-green-400 transition-colors"
                                            title="Manage services"
                                            on:click={() =>
                                                openServicesModal(user)}
                                        >
                                            <svg
                                                class="h-4 w-4"
                                                fill="none"
                                                stroke="currentColor"
                                                viewBox="0 0 24 24"
                                            >
                                                <path
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    stroke-width="2"
                                                    d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
                                                />
                                                <path
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    stroke-width="2"
                                                    d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                                                />
                                            </svg>
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </li>
                {:else}
                    <li class="p-8 text-center">
                        <div class="text-gray-500 dark:text-gray-400">
                            {selectedService === "all"
                                ? "No users found"
                                : `No users found for ${selectedService}`}
                        </div>
                    </li>
                {/each}
            </ul>
        </div>
    {/if}
</div>

<!-- Create User Modal -->
{#if showCreateModal}
    <div class="fixed inset-0 z-50 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
            <div
                class="fixed inset-0 bg-gray-500 bg-opacity-75"
                role="dialog"
                aria-modal="true"
                aria-labelledby="modal-title"
                on:click={() => (showCreateModal = false)}
                on:keydown={(e) =>
                    e.key === "Escape" && (showCreateModal = false)}
                tabindex="-1"
            ></div>
            <div
                class="relative bg-white dark:bg-gray-800 rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:w-full sm:max-w-lg"
            >
                <div
                    class="bg-white dark:bg-gray-800 px-4 pt-5 pb-4 sm:p-6 sm:pb-4"
                >
                    <div class="flex items-center justify-between mb-4">
                        <h3
                            id="modal-title"
                            class="text-lg font-medium text-gray-900 dark:text-white"
                        >
                            Create New User
                        </h3>
                        <button
                            on:click={() => (showCreateModal = false)}
                            class="text-gray-400 hover:text-gray-500"
                            aria-label="Close create user modal"
                        >
                            <svg
                                class="h-6 w-6"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                            >
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="2"
                                    d="M6 18L18 6M6 6l12 12"
                                />
                            </svg>
                        </button>
                    </div>
                    <form on:submit|preventDefault={createUser}>
                        <div class="space-y-4">
                            <div>
                                <label
                                    for="username"
                                    class="block text-sm font-medium text-gray-700 dark:text-gray-300"
                                    >Username</label
                                >
                                <input
                                    id="username"
                                    type="text"
                                    bind:value={newUser.username}
                                    required
                                    class="mt-1 block w-full border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                                />
                            </div>
                            <div>
                                <label
                                    for="fullname"
                                    class="block text-sm font-medium text-gray-700 dark:text-gray-300"
                                    >Full Name</label
                                >
                                <input
                                    id="fullname"
                                    type="text"
                                    bind:value={newUser.name}
                                    required
                                    class="mt-1 block w-full border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                                />
                            </div>
                            <div>
                                <label
                                    for="password"
                                    class="block text-sm font-medium text-gray-700 dark:text-gray-300"
                                    >Password</label
                                >
                                <input
                                    id="password"
                                    type="password"
                                    bind:value={newUser.password}
                                    required
                                    class="mt-1 block w-full border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                                />
                            </div>
                            <div>
                                <label
                                    for="groups"
                                    class="block text-sm font-medium text-gray-700 dark:text-gray-300"
                                    >Groups (comma separated)</label
                                >
                                <input
                                    id="groups"
                                    type="text"
                                    bind:value={newUser.groups}
                                    class="mt-1 block w-full border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                                />
                            </div>
                        </div>
                        <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                            <button
                                type="button"
                                on:click={() => (showCreateModal = false)}
                                class="w-full inline-flex justify-center rounded-md border border-gray-300 dark:border-gray-600 shadow-sm px-4 py-2 text-base font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:ml-3 sm:w-auto sm:text-sm"
                            >
                                Cancel
                            </button>
                            <button
                                type="submit"
                                class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 text-base font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:w-auto sm:text-sm"
                            >
                                Create User
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    </div>
{/if}

<!-- Edit User Modal -->
{#if showEditModal && selectedUser}
    <div class="fixed inset-0 z-50 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
            <div
                class="fixed inset-0 bg-gray-500 bg-opacity-75"
                role="dialog"
                aria-modal="true"
                aria-labelledby="edit-modal-title"
                on:click={() => (showEditModal = false)}
                on:keydown={(e) =>
                    e.key === "Escape" && (showEditModal = false)}
                tabindex="-1"
            ></div>
            <div
                class="relative bg-white dark:bg-gray-800 rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:w-full sm:max-w-lg"
            >
                <div
                    class="bg-white dark:bg-gray-800 px-4 pt-5 pb-4 sm:p-6 sm:pb-4"
                >
                    <div class="flex items-center justify-between mb-4">
                        <h3
                            id="edit-modal-title"
                            class="text-lg font-medium text-gray-900 dark:text-white"
                        >
                            Edit User: {selectedUser.username}
                        </h3>
                        <button
                            on:click={() => (showEditModal = false)}
                            class="text-gray-400 hover:text-gray-500"
                            aria-label="Close edit user modal"
                        >
                            <svg
                                class="h-6 w-6"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                            >
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="2"
                                    d="M6 18L18 6M6 6l12 12"
                                />
                            </svg>
                        </button>
                    </div>
                    <div class="space-y-4">
                        <div>
                            <label
                                for="edit-name"
                                class="block text-sm font-medium text-gray-700 dark:text-gray-300"
                                >Full Name</label
                            >
                            <input
                                id="edit-name"
                                type="text"
                                bind:value={selectedUser.name}
                                required
                                class="mt-1 block w-full border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                            />
                        </div>
                        <div>
                            <label
                                for="edit-groups"
                                class="block text-sm font-medium text-gray-700 dark:text-gray-300"
                                >Groups (comma separated)</label
                            >
                            <input
                                id="edit-groups"
                                type="text"
                                bind:value={selectedUser.groups}
                                class="mt-1 block w-full border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                            />
                        </div>
                        <div>
                            <label
                                for="edit-password"
                                class="block text-sm font-medium text-gray-700 dark:text-gray-300"
                                >New Password (leave blank to keep current)</label
                            >
                            <input
                                id="edit-password"
                                type="password"
                                bind:value={selectedUser.password}
                                placeholder="Enter new password"
                                class="mt-1 block w-full border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                            />
                        </div>
                        <div>
                            <label
                                for="edit-password-confirm"
                                class="block text-sm font-medium text-gray-700 dark:text-gray-300"
                                >Confirm New Password</label
                            >
                            <input
                                id="edit-password-confirm"
                                type="password"
                                bind:value={selectedUser.passwordConfirm}
                                placeholder="Confirm new password"
                                class="mt-1 block w-full border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                            />
                        </div>
                    </div>
                    <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                        <button
                            type="button"
                            on:click={() => (showEditModal = false)}
                            class="w-full inline-flex justify-center rounded-md border border-gray-300 dark:border-gray-600 shadow-sm px-4 py-2 text-base font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:ml-3 sm:w-auto sm:text-sm"
                        >
                            Cancel
                        </button>
                        <button
                            type="button"
                            on:click={updateUser}
                            class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 text-base font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:w-auto sm:text-sm"
                        >
                            Update User
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
{/if}

<!-- Services Modal -->
{#if showServicesModal && selectedUser}
    <div class="fixed inset-0 z-50 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
            <div
                class="fixed inset-0 bg-gray-500 bg-opacity-75"
                role="dialog"
                aria-modal="true"
                aria-labelledby="services-modal-title"
                on:click={() => (showServicesModal = false)}
                on:keydown={(e) =>
                    e.key === "Escape" && (showServicesModal = false)}
                tabindex="-1"
            ></div>
            <div
                class="relative bg-white dark:bg-gray-800 rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:w-full sm:max-w-lg"
            >
                <div
                    class="bg-white dark:bg-gray-800 px-4 pt-5 pb-4 sm:p-6 sm:pb-4"
                >
                    <div class="flex items-center justify-between mb-4">
                        <h3
                            class="text-lg font-medium text-gray-900 dark:text-white"
                        >
                            Manage Services for {selectedUser.username}
                        </h3>
                        <button
                            on:click={() => (showServicesModal = false)}
                            class="text-gray-400 hover:text-gray-500"
                            aria-label="Close services modal"
                        >
                            <svg
                                class="h-6 w-6"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                            >
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="2"
                                    d="M6 18L18 6M6 6l12 12"
                                />
                            </svg>
                        </button>
                    </div>
                    <div class="space-y-4">
                        <div class="flex items-center">
                            <input
                                id="samba-access"
                                type="checkbox"
                                bind:checked={selectedUser.services.samba}
                                class="h-4 w-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                            />
                            <label
                                for="samba-access"
                                class="ml-2 text-sm font-medium text-gray-700 dark:text-gray-300"
                                >Samba Access</label
                            >
                        </div>
                        <div class="flex items-center">
                            <input
                                id="nfs-access"
                                type="checkbox"
                                bind:checked={selectedUser.services.nfs}
                                class="h-4 w-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                            />
                            <label
                                for="nfs-access"
                                class="ml-2 text-sm font-medium text-gray-700 dark:text-gray-300"
                                >NFS Access</label
                            >
                        </div>
                        <div class="flex items-center">
                            <input
                                id="ssh-access"
                                type="checkbox"
                                bind:checked={selectedUser.services.ssh}
                                class="h-4 w-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                            />
                            <label
                                for="ssh-access"
                                class="ml-2 text-sm font-medium text-gray-700 dark:text-gray-300"
                                >SSH Access</label
                            >
                        </div>
                    </div>
                    <div class="mt-5 sm:mt-4">
                        <button
                            type="button"
                            on:click={updateUserServices}
                            class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 text-base font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:w-auto sm:text-sm"
                        >
                            Update Services
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
        transition:
            color 0.2s,
            background-color 0.2s,
            border-color 0.2s;
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
