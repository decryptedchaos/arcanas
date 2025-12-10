/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

export async function fetchSCSITargets() {
    const response = await fetch('/api/scsi-targets');
    if (!response.ok) throw new Error('Failed to fetch data');
    return response.json();
}

export async function fetchDiskStats() {
    const response = await fetch('/api/disk-stats');
    if (!response.ok) throw new Error('Failed to fetch data');
    return response.json();
}