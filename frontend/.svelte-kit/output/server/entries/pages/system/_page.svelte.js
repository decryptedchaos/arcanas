import "clsx";
import { a as attr } from "../../../chunks/attributes.js";
import { e as escape_html } from "../../../chunks/escaping.js";
function SystemStats($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let loading = true;
    $$renderer2.push(`<div class="space-y-6"><div class="flex items-center justify-between"><div><h2 class="text-xl font-bold text-gray-900 dark:text-white dark:text-white">System Statistics</h2> <p class="text-sm text-gray-600 dark:text-gray-300 dark:text-gray-300 mt-1">Real-time system performance monitoring</p></div> <div class="flex items-center space-x-3"><button class="btn btn-primary"${attr("disabled", loading, true)}>${escape_html("Loading...")}</button></div></div> `);
    {
      $$renderer2.push("<!--[!-->");
    }
    $$renderer2.push(`<!--]--> `);
    {
      $$renderer2.push("<!--[-->");
      $$renderer2.push(`<div class="text-center py-8"><div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div> <p class="mt-2 text-sm text-gray-600 dark:text-gray-300">Loading system statistics...</p></div>`);
    }
    $$renderer2.push(`<!--]--> `);
    {
      $$renderer2.push("<!--[!-->");
    }
    $$renderer2.push(`<!--]--></div>`);
  });
}
function _page($$renderer) {
  SystemStats($$renderer);
}
export {
  _page as default
};
