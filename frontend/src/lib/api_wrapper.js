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