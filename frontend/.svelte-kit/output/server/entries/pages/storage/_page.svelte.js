import { V as ensure_array_like, U as attr_class, W as stringify, a0 as attr_style } from "../../../chunks/index2.js";
import { a as attr } from "../../../chunks/attributes.js";
import { e as escape_html } from "../../../chunks/escaping.js";
function Storage($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let disks;
    let diskStats = [];
    let loading = true;
    let expandedDisks = /* @__PURE__ */ new Set();
    function getStatusColor(status) {
      switch (status) {
        case "healthy":
          return "text-green-600 bg-green-100";
        case "warning":
          return "text-yellow-600 bg-yellow-100";
        case "critical":
          return "text-red-600 bg-red-100";
        default:
          return "text-gray-600 dark:text-gray-300 bg-gray-100";
      }
    }
    function getUsageColor(percentage) {
      if (percentage >= 90) return "bg-red-500";
      if (percentage >= 75) return "bg-yellow-500";
      return "bg-green-500";
    }
    function formatBytes(bytes) {
      if (bytes === 0) return "0 B";
      const k = 1024;
      const sizes = ["B", "KB", "MB", "GB", "TB"];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
    }
    disks = diskStats || [];
    $$renderer2.push(`<div class="space-y-6"><div class="flex items-center justify-between"><div><h2 class="text-xl font-bold text-gray-900 dark:text-white">Arcanas Storage Management</h2> <p class="text-sm text-gray-600 dark:text-gray-300 mt-1">Manage disks, RAID arrays, and storage pools</p></div> <button class="btn btn-primary"${attr(
      "disabled",
      // Show RAID creation modal or redirect to dedicated RAID page
      // Show Pool creation modal or redirect to dedicated Pool page  
      // TODO: Rename this function - it returns disk info, not stats
      // TODO: Rename API call - returns disk info, not stats
      // Only update if data actually changed to prevent flashing
      // Trigger reactivity
      // Refresh stats every 10 seconds
      loading,
      true
    )}>${escape_html("Loading...")}</button></div> `);
    {
      $$renderer2.push("<!--[!-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (!disks.length) {
      $$renderer2.push("<!--[-->");
      $$renderer2.push(`<div class="text-center py-8"><div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div> <p class="mt-2 text-sm text-gray-600 dark:text-gray-300">Loading storage information...</p></div>`);
    } else {
      $$renderer2.push("<!--[!-->");
    }
    $$renderer2.push(`<!--]-->  <div class="grid grid-cols-1 md:grid-cols-3 gap-6"><div class="stat-card"><div class="flex items-center"><div class="p-3 bg-blue-100 rounded-lg mr-4"><svg class="h-6 w-6 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"></path></svg></div> <div><p class="text-sm font-medium text-gray-600 dark:text-gray-300">Total Storage</p> <p class="text-2xl font-bold text-gray-900 dark:text-white">${escape_html(disks.length)} disks</p> <p class="text-xs text-gray-500 dark:text-gray-400">Connected</p></div></div></div> <div class="stat-card"><div class="flex items-center"><div class="p-3 bg-green-100 rounded-lg mr-4"><svg class="h-6 w-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg></div> <div><p class="text-sm font-medium text-gray-600 dark:text-gray-300">Available</p> <p class="text-2xl font-bold text-gray-900 dark:text-white">${escape_html(disks.filter((d) => d.smart.status === "healthy").length)}</p> <p class="text-xs text-gray-500 dark:text-gray-400">Healthy disks</p></div></div></div> <div class="stat-card"><div class="flex items-center"><div class="p-3 bg-yellow-100 rounded-lg mr-4"><svg class="h-6 w-6 text-yellow-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg></div> <div><p class="text-sm font-medium text-gray-600 dark:text-gray-300">Health Alerts</p> <p class="text-2xl font-bold text-gray-900 dark:text-white">${escape_html(disks.filter((d) => d.smart.status !== "healthy").length)}</p> <p class="text-xs text-gray-500 dark:text-gray-400">Needs attention</p></div></div></div></div> <div class="space-y-3"><h3 class="text-lg font-semibold text-gray-900 dark:text-white">Disk Details</h3> <!--[-->`);
    const each_array = ensure_array_like(disks);
    for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
      let disk = each_array[$$index];
      $$renderer2.push(`<div class="card hover:shadow-lg transition-shadow"><div class="flex items-center justify-between p-4 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700" role="button" tabindex="0"${attr("aria-label", `Toggle details for ${stringify(disk.device)}`)}${attr("aria-expanded", expandedDisks.has(disk.device))}><div class="flex items-center space-x-4 flex-1"><div${attr_class("transition-transform duration-200", void 0, { "rotate-90": expandedDisks.has(disk.device) })}><svg class="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path></svg></div> <div class="flex-1"><div class="flex items-center space-x-3"><h4 class="text-lg font-semibold text-gray-900 dark:text-white">${escape_html(disk.device)}</h4> <span${attr_class(`px-2 py-1 text-xs font-medium rounded-full ${stringify(getStatusColor(disk.smart.status))}`)}>${escape_html(disk.smart.status)}</span></div> <p class="text-sm text-gray-600 dark:text-gray-300">${escape_html(disk.model)}</p></div> <div class="text-right"><div class="text-lg font-semibold text-gray-900 dark:text-white">${escape_html(disk.usage.toFixed(1))}%</div> <div class="text-sm text-gray-600 dark:text-gray-300">${escape_html(formatBytes(disk.used))} / ${escape_html(formatBytes(disk.size))}</div></div></div></div> `);
      if (expandedDisks.has(disk.device)) {
        $$renderer2.push("<!--[-->");
        $$renderer2.push(`<div class="border-t border-gray-200 dark:border-gray-700 p-4 space-y-4"><div class="grid grid-cols-1 md:grid-cols-3 gap-4"><div><label${attr("for", `storage-type-${stringify(disk.device)}`)} class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Storage Type</label> <select${attr("id", `storage-type-${stringify(disk.device)}`)} class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 text-gray-900 dark:text-white">`);
        $$renderer2.option({ value: "jbod" }, ($$renderer3) => {
          $$renderer3.push(`JBOD (Independent)`);
        });
        $$renderer2.option({ value: "raid" }, ($$renderer3) => {
          $$renderer3.push(`RAID Array`);
        });
        $$renderer2.option({ value: "zfs" }, ($$renderer3) => {
          $$renderer3.push(`ZFS Pool`);
        });
        $$renderer2.push(`</select></div> <div><label${attr("for", `mount-point-${stringify(disk.device)}`)} class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Mount Point</label> <input${attr("id", `mount-point-${stringify(disk.device)}`)} type="text"${attr("value", disk.mountpoint || "")} placeholder="e.g., /data" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 text-gray-900 dark:text-white"/></div> <div><label${attr("for", `filesystem-${stringify(disk.device)}`)} class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Filesystem</label> <select${attr("id", `filesystem-${stringify(disk.device)}`)} class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 text-gray-900 dark:text-white">`);
        $$renderer2.option({ value: "ext4", selected: disk.filesystem === "ext4" }, ($$renderer3) => {
          $$renderer3.push(`ext4`);
        });
        $$renderer2.option({ value: "xfs", selected: disk.filesystem === "xfs" }, ($$renderer3) => {
          $$renderer3.push(`XFS`);
        });
        $$renderer2.option({ value: "btrfs", selected: disk.filesystem === "btrfs" }, ($$renderer3) => {
          $$renderer3.push(`Btrfs`);
        });
        $$renderer2.option({ value: "zfs", selected: disk.filesystem === "zfs" }, ($$renderer3) => {
          $$renderer3.push(`ZFS`);
        });
        $$renderer2.option({ value: "mergerfs", selected: disk.filesystem === "mergerfs" }, ($$renderer3) => {
          $$renderer3.push(`MergerFS`);
        });
        $$renderer2.push(`</select></div></div> <div><div class="flex items-center justify-between text-sm mb-2"><span class="text-gray-600 dark:text-gray-300">Usage: ${escape_html(formatBytes(disk.used))} / ${escape_html(formatBytes(disk.size))}</span> <span class="font-medium text-gray-900 dark:text-white">${escape_html(disk.usage.toFixed(1))}%</span></div> <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3"><div${attr_class(`${stringify(getUsageColor(disk.usage))} h-3 rounded-full transition-all duration-500`)}${attr_style(`width: ${stringify(disk.usage)}%`)}></div></div></div> <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm"><div><p class="text-gray-600 dark:text-gray-300">Temperature</p> <p class="font-medium text-gray-900 dark:text-white">${escape_html(disk.smart.temperature)}°C</p></div> <div><p class="text-gray-600 dark:text-gray-300">Health</p> <p class="font-medium text-gray-900 dark:text-white">${escape_html(disk.smart.health)}%</p></div> <div><p class="text-gray-600 dark:text-gray-300">Available</p> <p class="font-medium text-gray-900 dark:text-white">${escape_html(formatBytes(disk.available))}</p></div> <div><p class="text-gray-600 dark:text-gray-300">Read-Only</p> <p class="font-medium text-gray-900 dark:text-white">${escape_html(disk.read_only ? "Yes" : "No")}</p></div></div> <div class="flex space-x-2 pt-2"><button class="btn btn-primary">Apply Changes</button> <button class="btn btn-secondary">Format Disk</button> <button class="btn btn-secondary">SMART Test</button> <button class="btn btn-secondary">Unmount</button></div></div>`);
      } else {
        $$renderer2.push("<!--[!-->");
      }
      $$renderer2.push(`<!--]--></div>`);
    }
    $$renderer2.push(`<!--]--></div></div>`);
  });
}
function _page($$renderer) {
  $$renderer.push(`<div class="space-y-6"><div class="border-b border-gray-200 dark:border-gray-700"><nav class="-mb-px flex space-x-8"><button${attr_class(`py-2 px-1 border-b-2 font-medium text-sm ${stringify(
    "border-blue-500 text-blue-600"
  )}`)}>Disks</button> <button${attr_class(`py-2 px-1 border-b-2 font-medium text-sm ${stringify("border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300")}`)}>RAID Arrays</button> <button${attr_class(`py-2 px-1 border-b-2 font-medium text-sm ${stringify("border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300")}`)}>Storage Pools</button></nav></div> `);
  {
    $$renderer.push("<!--[-->");
    Storage($$renderer);
  }
  $$renderer.push(`<!--]--></div>`);
}
export {
  _page as default
};
