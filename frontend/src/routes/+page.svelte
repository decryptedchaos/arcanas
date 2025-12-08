<script>
  import DashboardStats from '$lib/components/DashboardStats.svelte';
  import QuickActions from '$lib/components/QuickActions.svelte';
  import RecentActivity from '$lib/components/RecentActivity.svelte';
</script>

<div class="space-y-6">
  <!-- Header -->
  <div>
    <h2 class="text-2xl font-bold text-gray-900 dark:text-white">Dashboard Overview</h2>
    <p class="mt-1 text-sm text-gray-600 dark:text-gray-300">Monitor your NAS system status and performance</p>
  </div>
  
  <!-- Quick Actions -->
  <QuickActions />
  
  <!-- System Stats Grid -->
  <DashboardStats />
  
  <!-- Recent Activity -->
  <RecentActivity />
</div>