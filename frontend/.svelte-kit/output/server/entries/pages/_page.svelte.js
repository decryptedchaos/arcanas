import "clsx";
import { a0 as attr_style, W as stringify, V as ensure_array_like, U as attr_class } from "../../chunks/index2.js";
import { e as escape_html } from "../../chunks/escaping.js";
import { a as attr } from "../../chunks/attributes.js";
import { h as html } from "../../chunks/html.js";
function DashboardStats($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let stats = {
      memory: {
        usedFormatted: "0.0",
        totalFormatted: "0.0"
      },
      storage: {
        usedFormatted: "0.0",
        totalFormatted: "0.0"
      },
      network: {
        rxFormatted: "0.0",
        rxRateFormatted: "0.0",
        txRateFormatted: "0.0"
      },
      uptime: "0 days, 0 hours",
      temperature: 0,
      services: { total: 3 }
      // Total of 3 NAS services: SCSI, Samba, NFS
    };
    $$renderer2.push(`<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6"><div class="stat-card"><div class="flex items-center justify-between"><div><p class="text-sm font-medium text-gray-600 dark:text-gray-300">CPU Usage</p> <p class="text-2xl font-bold text-gray-900 dark:text-white">${escape_html(0)}%</p> <p class="text-xs text-gray-500 dark:text-gray-400">${escape_html(0)} cores</p></div> <div class="p-3 bg-blue-100 rounded-full"><svg class="h-6 w-6 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"></path></svg></div></div> <div class="mt-4"><div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2"><div class="bg-blue-600 h-2 rounded-full transition-all duration-500"${attr_style(`width: ${stringify(0)}%`)}></div></div></div></div> <div class="stat-card"><div class="flex items-center justify-between"><div><p class="text-sm font-medium text-gray-600 dark:text-gray-300">Memory</p> <p class="text-2xl font-bold text-gray-900 dark:text-white">${escape_html(0)}%</p> <p class="text-xs text-gray-500 dark:text-gray-400">${escape_html(stats?.memory?.usedFormatted)}GB / ${escape_html(stats?.memory?.totalFormatted)}GB</p></div> <div class="p-3 bg-green-100 rounded-full"><svg class="h-6 w-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z"></path></svg></div></div> <div class="mt-4"><div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2"><div class="bg-green-600 h-2 rounded-full transition-all duration-500"${attr_style(`width: ${stringify(0)}%`)}></div></div></div></div> <div class="stat-card"><div class="flex items-center justify-between"><div><p class="text-sm font-medium text-gray-600 dark:text-gray-300">Storage</p> <p class="text-2xl font-bold text-gray-900 dark:text-white">${escape_html(0)}%</p> <p class="text-xs text-gray-500 dark:text-gray-400">${escape_html(stats?.storage?.usedFormatted)}TB / ${escape_html(stats?.storage?.totalFormatted)}TB</p></div> <div class="p-3 bg-purple-100 rounded-full"><svg class="h-6 w-6 text-purple-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"></path></svg></div></div> <div class="mt-4"><div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2"><div class="bg-purple-600 h-2 rounded-full transition-all duration-500"${attr_style(`width: ${stringify(0)}%`)}></div></div></div></div> <div class="stat-card"><div class="flex items-center justify-between"><div><p class="text-sm font-medium text-gray-600 dark:text-gray-300">Network</p> <p class="text-2xl font-bold text-gray-900 dark:text-white">${escape_html(stats?.network?.rxFormatted)}MB</p></div> <div class="p-3 bg-orange-100 rounded-full"><svg class="h-6 w-6 text-orange-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg></div></div> <div class="mt-4"><div class="flex space-x-2 mb-1"><div class="flex-1 text-xs text-gray-500 dark:text-gray-400">↓ ${escape_html(stats?.network?.rxRateFormatted)}KB/s</div> <div class="flex-1 text-xs text-gray-500 dark:text-gray-400 text-right">↑ ${escape_html(stats?.network?.txRateFormatted)}KB/s</div></div> <div class="flex space-x-2"><div class="flex-1"><div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2"><div class="bg-orange-500 h-2 rounded-full transition-all duration-500"${attr_style(`width: ${stringify(Math.min(0 / 10, 100))}%`)}></div></div> <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">Download</div></div> <div class="flex-1"><div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2"><div class="bg-orange-400 h-2 rounded-full transition-all duration-500"${attr_style(`width: ${stringify(Math.min(0 / 10, 100))}%`)}></div></div> <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">Upload</div></div></div></div></div></div> <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mt-6"><div class="stat-card"><div class="flex items-center"><div class="p-2 bg-gray-100 rounded-lg mr-3"><svg class="h-5 w-5 text-gray-600 dark:text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg></div> <div><p class="text-sm font-medium text-gray-600 dark:text-gray-300">Uptime</p> <p class="text-lg font-semibold text-gray-900 dark:text-white">${escape_html(stats.uptime)}</p></div></div></div> <div class="stat-card"><div class="flex items-center"><div class="p-2 bg-red-100 rounded-lg mr-3"><svg class="h-5 w-5 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path></svg></div> <div><p class="text-sm font-medium text-gray-600 dark:text-gray-300">Temperature</p> <p class="text-lg font-semibold text-gray-900 dark:text-white">${escape_html(stats.temperature)}°C</p></div></div></div> <div class="stat-card"><div class="flex items-center"><div class="p-2 bg-green-100 rounded-lg mr-3"><svg class="h-5 w-5 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg></div> <div><p class="text-sm font-medium text-gray-600 dark:text-gray-300">Services</p> <p class="text-lg font-semibold text-gray-900 dark:text-white">${escape_html(0)}/${escape_html(stats?.services?.total)} online</p></div></div></div></div>`);
  });
}
function QuickActions($$renderer) {
  const actions = [
    {
      title: "Create Share",
      description: "Create new Samba or NFS share",
      icon: "share",
      color: "blue",
      href: "/shares/create"
    },
    {
      title: "Add Disk",
      description: "Configure new storage disk",
      icon: "disk",
      color: "green",
      href: "/storage/add"
    },
    {
      title: "SCSI Target",
      description: "Manage iSCSI targets",
      icon: "target",
      color: "purple",
      href: "/scsi"
    },
    {
      title: "System Backup",
      description: "Configure system backups",
      icon: "backup",
      color: "orange",
      href: "/backup"
    }
  ];
  function getIcon(iconName) {
    const icons = {
      share: '<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m9.032 4.026a3 3 0 10-4.732 2.684m4.732-2.684a3 3 0 00-4.732-2.684M3 12a3 3 0 104.732 2.684M3 12a3 3 0 014.732-2.684" /></svg>',
      disk: '<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" /></svg>',
      target: '<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>',
      backup: '<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>'
    };
    return icons[iconName] || icons.share;
  }
  function getColorClasses(color) {
    const colors = {
      blue: "bg-blue-100 text-blue-600 hover:bg-blue-200",
      green: "bg-green-100 text-green-600 hover:bg-green-200",
      purple: "bg-purple-100 text-purple-600 hover:bg-purple-200",
      orange: "bg-orange-100 text-orange-600 hover:bg-orange-200"
    };
    return colors[color] || colors.blue;
  }
  $$renderer.push(`<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4"><!--[-->`);
  const each_array = ensure_array_like(actions);
  for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
    let action = each_array[$$index];
    $$renderer.push(`<a${attr("href", action.href)} class="flex items-center p-4 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 hover:shadow-md transition-all duration-200 group"><div${attr_class(`flex-shrink-0 p-3 ${stringify(getColorClasses(action.color))} rounded-lg group-hover:scale-105 transition-transform`)}>${html(getIcon(action.icon))}</div> <div class="ml-4"><h3 class="text-base font-semibold text-gray-900 dark:text-white">${escape_html(action.title)}</h3> <p class="text-sm text-gray-600 dark:text-gray-300 mt-1">${escape_html(action.description)}</p></div></a>`);
  }
  $$renderer.push(`<!--]--></div>`);
}
function RecentActivity($$renderer) {
  let activities = [
    {
      id: 1,
      type: "success",
      title: "Samba share created",
      description: 'Share "media" created successfully',
      time: "2 minutes ago",
      icon: "share"
    },
    {
      id: 2,
      type: "warning",
      title: "Disk space low",
      description: "Disk /dev/sdb1 at 85% capacity",
      time: "15 minutes ago",
      icon: "disk"
    },
    {
      id: 3,
      type: "info",
      title: "System backup completed",
      description: "Daily backup finished successfully",
      time: "1 hour ago",
      icon: "backup"
    },
    {
      id: 4,
      type: "error",
      title: "SCSI target disconnected",
      description: "Target iqn.2024-01.com.nas:target1 lost connection",
      time: "2 hours ago",
      icon: "target"
    },
    {
      id: 5,
      type: "success",
      title: "NFS export added",
      description: 'Export "/data/backups" added for 192.168.1.0/24',
      time: "3 hours ago",
      icon: "network"
    }
  ];
  function getTypeClasses(type) {
    const types = {
      success: "bg-green-100 text-green-600 border-green-200",
      warning: "bg-yellow-100 text-yellow-600 border-yellow-200",
      error: "bg-red-100 text-red-600 border-red-200",
      info: "bg-blue-100 text-blue-600 border-blue-200"
    };
    return types[type] || types.info;
  }
  function getIcon(iconName) {
    const icons = {
      share: '<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m9.032 4.026a3 3 0 10-4.732 2.684m4.732-2.684a3 3 0 00-4.732-2.684M3 12a3 3 0 104.732 2.684M3 12a3 3 0 014.732-2.684" /></svg>',
      disk: '<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" /></svg>',
      backup: '<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>',
      target: '<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>',
      network: '<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" /></svg>'
    };
    return icons[iconName] || icons.info;
  }
  $$renderer.push(`<div class="card"><div class="flex items-center justify-between mb-4"><h3 class="text-lg font-semibold text-gray-900 dark:text-white dark:text-white">Recent Activity</h3> <button class="text-sm text-primary-600 hover:text-primary-700 font-medium">View All</button></div> <div class="space-y-3"><!--[-->`);
  const each_array = ensure_array_like(activities);
  for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
    let activity = each_array[$$index];
    $$renderer.push(`<div class="flex items-start space-x-3 p-3 rounded-lg border border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"><div${attr_class(`flex-shrink-0 p-2 ${stringify(getTypeClasses(activity.type))} rounded-lg border`)}>${html(getIcon(activity.icon))}</div> <div class="flex-1 min-w-0"><p class="text-sm font-medium text-gray-900 dark:text-white">${escape_html(activity.title)}</p> <p class="text-sm text-gray-600 dark:text-gray-300 mt-1">${escape_html(activity.description)}</p> <p class="text-xs text-gray-500 dark:text-gray-400 mt-2">${escape_html(activity.time)}</p></div></div>`);
  }
  $$renderer.push(`<!--]--></div></div>`);
}
function _page($$renderer) {
  $$renderer.push(`<div class="space-y-6"><div><h2 class="text-2xl font-bold text-gray-900 dark:text-white">Dashboard Overview</h2> <p class="mt-1 text-sm text-gray-600 dark:text-gray-300">Monitor your NAS system status and performance</p></div> `);
  QuickActions($$renderer);
  $$renderer.push(`<!----> `);
  DashboardStats($$renderer);
  $$renderer.push(`<!----> `);
  RecentActivity($$renderer);
  $$renderer.push(`<!----></div>`);
}
export {
  _page as default
};
