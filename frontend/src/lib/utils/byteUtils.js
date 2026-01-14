/**
 * Byte conversion utilities for entire frontend
 * Provides consistent formatting across all components
 */

/**
 * Convert bytes to human-readable format
 * @param {number} bytes - Number of bytes to convert
 * @param {number} decimals - Number of decimal places (default: 1)
 * @returns {string} Human-readable string (e.g., "2.5 GB")
 */
export function formatBytes(bytes, decimals = 1) {
  if (bytes === 0 || bytes === null || bytes === undefined) return '0 B';

  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(decimals)) + ' ' + sizes[i];
}

/**
 * Convert bytes to human-readable format with full words
 * @param {number} bytes - Number of bytes to convert
 * @param {number} decimals - Number of decimal places (default: 1)
 * @returns {string} Human-readable string with full words (e.g., "2.5 Gigabytes")
 */
export function formatBytesFull(bytes, decimals = 1) {
  if (bytes === 0 || bytes === null || bytes === undefined) return '0 Bytes';

  const k = 1024;
  const sizes = ['Bytes', 'Kilobytes', 'Megabytes', 'Gigabytes', 'Terabytes', 'Petabytes', 'Exabytes'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(decimals)) + ' ' + sizes[i];
}

/**
 * Convert bytes to human-readable format with proper suffix
 * @param {number} bytes - Number of bytes to convert
 * @param {number} decimals - Number of decimal places (default: 1)
 * @returns {string} Human-readable string with proper suffix (e.g., "2.5GB")
 */
export function formatBytesCompact(bytes, decimals = 1) {
  if (bytes === 0 || bytes === null || bytes === undefined) return '0B';

  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(decimals)) + sizes[i];
}

/**
 * Get the unit for a given number of bytes
 * @param {number} bytes - Number of bytes
 * @returns {string} Unit string (e.g., "GB")
 */
export function getByteUnit(bytes) {
  if (bytes === 0 || bytes === null || bytes === undefined) return 'B';

  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return sizes[i];
}

/**
 * Convert bytes to specified unit
 * @param {number} bytes - Number of bytes to convert
 * @param {string} targetUnit - Target unit ('B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB')
 * @param {number} decimals - Number of decimal places (default: 1)
 * @returns {number} Value in target unit
 */
export function convertToUnit(bytes, targetUnit, decimals = 1) {
  if (bytes === 0 || bytes === null || bytes === undefined) return 0;

  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB'];
  const targetIndex = sizes.indexOf(targetUnit.toUpperCase());

  if (targetIndex === -1) {
    throw new Error(`Invalid target unit: ${targetUnit}. Must be one of: ${sizes.join(', ')}`);
  }

  const value = bytes / Math.pow(k, targetIndex);
  return parseFloat(value.toFixed(decimals));
}

/**
 * Calculate percentage of used space
 * @param {number} used - Used bytes
 * @param {number} total - Total bytes
 * @returns {number} Percentage (0-100)
 */
export function calculateUsagePercentage(used, total) {
  if (!total || total === 0) return 0;
  return Math.round((used / total) * 100);
}

/**
 * Format storage usage with percentage
 * @param {number} used - Used bytes
 * @param {number} total - Total bytes
 * @param {number} decimals - Number of decimal places for size (default: 1)
 * @returns {string} Formatted string (e.g., "2.5 GB of 10 GB (25%)")
 */
export function formatStorageUsage(used, total, decimals = 1) {
  const usedStr = formatBytes(used, decimals);
  const totalStr = formatBytes(total, decimals);
  const percentage = calculateUsagePercentage(used, total);

  return `${usedStr} of ${totalStr} (${percentage}%)`;
}

/**
 * Get appropriate color class for usage percentage
 * @param {number} percentage - Usage percentage (0-100)
 * @returns {string} Tailwind CSS color class
 */
export function getUsageColorClass(percentage) {
  if (percentage >= 90) return 'text-red-600 dark:text-red-400';
  if (percentage >= 75) return 'text-yellow-600 dark:text-yellow-400';
  return 'text-green-600 dark:text-green-400';
}

/**
 * Get appropriate background color class for usage percentage
 * @param {number} percentage - Usage percentage (0-100)
 * @returns {string} Tailwind CSS background color class
 */
export function getUsageBgColorClass(percentage) {
  if (percentage >= 90) return 'bg-red-500';
  if (percentage >= 75) return 'bg-yellow-500';
  return 'bg-green-500';
}

/**
 * Calculate Y-axis scale for graphs with dynamic padding
 * @param {Array} history - Array of data points with timestamp and values
 * @param {string} key - Key to extract from data points (e.g., 'rx', 'tx', 'read', 'write', 'value', 'all', 'net')
 * @returns {Object} Scale object with max and steps array
 */
export function calculateScale(history, key) {
  if (!history || history.length === 0)
    return { max: 100, steps: [100, 75, 50, 25, 0] };

  let values;
  let maxValue;

  // Handle 'all' key for unified disk I/O scaling
  if (key === 'all') {
    // Combine read and write values to find the maximum
    values = history.flatMap((point) => [point.read || 0, point.write || 0]);
  } else if (key === 'net') {
    // Combine rx and tx values to find the maximum
    values = history.flatMap((point) => [point.rx || 0, point.tx || 0]);
  } else {
    values = history.map((point) => point[key] || 0);
  }

  maxValue = Math.max(...values, 1);

  // Handle CPU percentage values differently (0-100 scale)
  if (key === 'value') {
    // CPU percentages - simple scaling with clean steps
    const scaledMax = Math.ceil(maxValue * 1.2); // 20% padding
    const finalMax = Math.min(scaledMax, 100); // Cap at 100%

    // Generate clean round steps for percentages
    const step = finalMax / 4;
    const steps = [
      Math.round(finalMax),
      Math.round(finalMax * 0.75),
      Math.round(finalMax * 0.5),
      Math.round(finalMax * 0.25),
      0,
    ];

    return { max: finalMax, steps };
  }

  // For network/disk I/O values (bytes/sec), convert to display units for scaling
  const isNetwork = key === 'rx' || key === 'tx' || key === 'net';
  const displayValue = isNetwork ? maxValue / 1000000 : maxValue / (1024 * 1024);

  // Find nearest power of 10 or clean number for scaling
  let scaledMax = calculateNiceScale(displayValue);

  // Convert back to original units for scale
  const finalMax = isNetwork ? scaledMax * 1000000 : scaledMax * 1024 * 1024;

  // Generate nice round steps
  const step = finalMax / 4;
  const steps = [
    Math.round(finalMax),
    Math.round(finalMax * 0.75),
    Math.round(finalMax * 0.5),
    Math.round(finalMax * 0.25),
    0,
  ];

  return { max: finalMax, steps };
}

/**
 * Calculate a nice round scale value for graphs
 * @param {number} value - Maximum value to scale
 * @returns {number} Nice round scale value
 */
function calculateNiceScale(value) {
  if (value === 0) return 10;

  // Add 10-20% padding
  const paddedValue = value * 1.1;

  // Find the power of 10
  const exponent = Math.floor(Math.log10(paddedValue));
  const base = Math.pow(10, exponent);

  // Find the nice step
  const fraction = paddedValue / base;

  if (fraction <= 1) {
    return 1 * base;
  } else if (fraction <= 2) {
    return 2 * base;
  } else if (fraction <= 5) {
    return 5 * base;
  } else {
    return 10 * base;
  }
}
