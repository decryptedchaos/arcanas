/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

import { writable, derived } from 'svelte/store';
import { authAPI } from '$lib/api.js';

// Auth state store
function createAuthStore() {
  const { subscribe, set, update } = writable({
    isAuthenticated: false,
    username: null,
    isRoot: false,
    isAdmin: false,
    loading: true,
    error: null,
  });

  return {
    subscribe,
    login: async (username, password) => {
      update((state) => ({ ...state, loading: true, error: null }));
      try {
        const response = await authAPI.login(username, password);
        set({
          isAuthenticated: true,
          username: response.username,
          isRoot: response.is_root,
          isAdmin: response.is_admin,
          loading: false,
          error: null,
        });
        return true;
      } catch (error) {
        update((state) => ({
          ...state,
          loading: false,
          error: error.message || 'Login failed',
        }));
        return false;
      }
    },

    logout: async () => {
      try {
        await authAPI.logout();
      } catch (error) {
        console.error('Logout error:', error);
      } finally {
        set({
          isAuthenticated: false,
          username: null,
          isRoot: false,
          isAdmin: false,
          loading: false,
          error: null,
        });
      }
    },

    validate: async () => {
      update((state) => ({ ...state, loading: true }));
      try {
        const response = await authAPI.validate();
        set({
          isAuthenticated: response.valid,
          username: response.username,
          isRoot: response.is_root,
          isAdmin: response.is_admin,
          loading: false,
          error: null,
        });
      } catch (error) {
        set({
          isAuthenticated: false,
          username: null,
          isRoot: false,
          isAdmin: false,
          loading: false,
          error: null,
        });
      }
    },

    clearError: () => {
      update((state) => ({ ...state, error: null }));
    },
  };
}

export const auth = createAuthStore();

// Derived stores for convenience
export const isAdmin = derived(auth, ($auth) => $auth.isAdmin);
export const isRoot = derived(auth, ($auth) => $auth.isRoot);
export const canManageAllShares = derived(auth, ($auth) => $auth.isAdmin || $auth.isRoot);
