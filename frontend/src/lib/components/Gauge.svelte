<!--
  Scalable circular gauge component
  Uses CSS variables for responsive scaling
-->

<script>
  export let value = 0; // Current value (0-100 or actual value)
  export let max = 100; // Maximum value for percentage calculation
  export let color = "#3B82F6"; // Stroke color
  export let label = ""; // Label text
  export let showValue = true; // Whether to display the value in center
  export let valueFormatter = (v) => v.toFixed(1); // Formatter for displayed value

  // Calculate percentage
  $: percentage = Math.min((value / max) * 100, 100);

  // SVG dimensions based on viewport units for scaling
  const viewBox = "0 0 100 100";
  const cx = 50;
  const cy = 50;
  const radius = 42;
  const circumference = 2 * Math.PI * radius; // ~263.89
  const strokeWidth = 8;
</script>

<div class="gauge-container">
  <svg class="gauge-svg" viewBox={viewBox}>
    <!-- Background circle -->
    <circle
      cx={cx}
      cy={cy}
      r={radius}
      stroke="#E5E7EB"
      stroke-width={strokeWidth}
      fill="none"
      class="dark:stroke-gray-600"
    />
    <!-- Value circle -->
    <circle
      cx={cx}
      cy={cy}
      r={radius}
      stroke={color}
      stroke-width={strokeWidth}
      fill="none"
      stroke-dasharray={circumference}
      stroke-dashoffset={circumference - (percentage / 100) * circumference}
      stroke-linecap="round"
      class="gauge-value"
      style="transition: stroke-dashoffset 0.5s ease"
    />
  </svg>

  {#if showValue || label}
    <div class="gauge-content">
      {#if showValue}
        <span class="gauge-value-text">{valueFormatter(value)}</span>
      {/if}
      {#if label}
        <span class="gauge-label">{label}</span>
      {/if}
    </div>
  {/if}
</div>

<style>
  .gauge-container {
    position: relative;
    width: var(--gauge-size);
    height: var(--gauge-size);
  }

  .gauge-svg {
    width: 100%;
    height: 100%;
    transform: rotate(-90deg);
  }

  .gauge-content {
    position: absolute;
    inset: 15%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: calc(var(--spacing-xs) * 0.5);
  }

  .gauge-value-text {
    font-size: var(--font-lg);
    font-weight: 700;
    color: inherit;
    line-height: 1.1;
  }

  .gauge-label {
    font-size: var(--font-sm);
    color: #6B7280;
    line-height: 2;
  }

  .dark .gauge-label {
    color: #9CA3AF;
  }
</style>
