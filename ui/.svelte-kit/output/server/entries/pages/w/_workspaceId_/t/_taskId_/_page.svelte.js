import { g as getContext, p as push, b as ensure_array_like, c as attr, d as stringify, e as escape_html, a as pop, f as store_get, u as unsubscribe_stores } from "../../../../../../chunks/index.js";
import "@sveltejs/kit/internal";
import "../../../../../../chunks/exports.js";
import "../../../../../../chunks/utils.js";
import "clsx";
import "../../../../../../chunks/state.svelte.js";
import { w as workspaceStore } from "../../../../../../chunks/workspaces.svelte.js";
import "marked";
const getStores = () => {
  const stores$1 = getContext("__svelte__");
  return {
    /** @type {typeof page} */
    page: {
      subscribe: stores$1.page.subscribe
    },
    /** @type {typeof navigating} */
    navigating: {
      subscribe: stores$1.navigating.subscribe
    },
    /** @type {typeof updated} */
    updated: stores$1.updated
  };
};
const page = {
  subscribe(fn) {
    const store = getStores().page;
    return store.subscribe(fn);
  }
};
function TaskFlowchart($$payload, $$props) {
  push();
  let { flowchart, selectedNodeId } = $$props;
  const svgBounds = () => {
    const nodes = flowchart.nodes;
    if (nodes.length === 0) return { minX: 0, minY: 0, width: 800, height: 600 };
    const minX = Math.min(...nodes.map((n) => n.position.x)) - 50;
    const minY = Math.min(...nodes.map((n) => n.position.y)) - 50;
    const maxX = Math.max(...nodes.map((n) => n.position.x)) + 200;
    const maxY = Math.max(...nodes.map((n) => n.position.y)) + 100;
    return { minX, minY, width: maxX - minX, height: maxY - minY };
  };
  function getNodeShape(node) {
    switch (node.type) {
      case "start":
      case "end":
        return "rounded";
      case "decision":
        return "diamond";
      case "process":
      default:
        return "rectangle";
    }
  }
  function getNodeColor(node) {
    if (node.completed) return "#10b981";
    switch (node.type) {
      case "start":
        return "#3b82f6";
      case // blue
      "end":
        return "#8b5cf6";
      case // purple
      "decision":
        return "#f59e0b";
      case // orange
      "process":
      default:
        return "#6b7280";
    }
  }
  function getEdgePath(edge) {
    const sourceNode = flowchart.nodes.find((n) => n.id === edge.source);
    const targetNode = flowchart.nodes.find((n) => n.id === edge.target);
    if (!sourceNode || !targetNode) return "";
    const sourcePoint = getNodeAnchor(sourceNode, targetNode);
    const targetPoint = getNodeAnchor(targetNode, sourceNode);
    return `M ${sourcePoint.x} ${sourcePoint.y} L ${targetPoint.x} ${targetPoint.y}`;
  }
  function getNodeAnchor(node, otherNode) {
    const nodeWidth = 150;
    const nodeHeight = node.type === "decision" ? 80 : 60;
    const centerX = node.position.x + nodeWidth / 2;
    const centerY = node.position.y + nodeHeight / 2;
    const dx = otherNode.position.x + 75 - centerX;
    const dy = otherNode.position.y + 30 - centerY;
    const angle = Math.atan2(dy, dx) * (180 / Math.PI);
    if (angle >= -45 && angle < 45) {
      return { x: node.position.x + nodeWidth, y: centerY };
    } else if (angle >= 45 && angle < 135) {
      return { x: centerX, y: node.position.y + nodeHeight };
    } else if (angle >= 135 || angle < -135) {
      return { x: node.position.x, y: centerY };
    } else {
      return { x: centerX, y: node.position.y };
    }
  }
  function getEdgeLabelPosition(edge) {
    if (!edge.label) return null;
    const sourceNode = flowchart.nodes.find((n) => n.id === edge.source);
    const targetNode = flowchart.nodes.find((n) => n.id === edge.target);
    if (!sourceNode || !targetNode) return null;
    const sourcePoint = getNodeAnchor(sourceNode, targetNode);
    const targetPoint = getNodeAnchor(targetNode, sourceNode);
    return {
      x: (sourcePoint.x + targetPoint.x) / 2,
      y: (sourcePoint.y + targetPoint.y) / 2
    };
  }
  const each_array = ensure_array_like(flowchart.edges);
  const each_array_1 = ensure_array_like(flowchart.nodes);
  $$payload.out.push(`<svg${attr("viewBox", `${stringify(svgBounds().minX)} ${stringify(svgBounds().minY)} ${stringify(svgBounds().width)} ${stringify(svgBounds().height)}`)}${attr("width", svgBounds().width)}${attr("height", svgBounds().height)}><defs><marker id="arrowhead" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto"><polygon points="0 0, 10 3, 0 6" fill="#6b7280"></polygon></marker></defs><!--[-->`);
  for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
    let edge = each_array[$$index];
    $$payload.out.push(`<g><path${attr("d", getEdgePath(edge))} stroke="#6b7280" stroke-width="2" fill="none" marker-end="url(#arrowhead)"></path>`);
    if (edge.label) {
      $$payload.out.push("<!--[-->");
      const labelPos = getEdgeLabelPosition(edge);
      if (labelPos) {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`<text${attr("x", labelPos.x)}${attr("y", labelPos.y)} text-anchor="middle" class="fill-base-content text-sm font-medium" dominant-baseline="middle">${escape_html(edge.label)}</text>`);
      } else {
        $$payload.out.push("<!--[!-->");
      }
      $$payload.out.push(`<!--]-->`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--></g>`);
  }
  $$payload.out.push(`<!--]--><!--[-->`);
  for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
    let node = each_array_1[$$index_1];
    const shape = getNodeShape(node);
    const color = getNodeColor(node);
    const isSelected = selectedNodeId === node.id;
    const hasAssignments = node.tools && node.tools.length > 0 || node.agents && node.agents.length > 0 || node.tasks && node.tasks.length > 0;
    $$payload.out.push(`<g role="button" class="cursor-pointer transition-opacity hover:opacity-80" tabindex="0">`);
    if (shape === "diamond") {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<polygon${attr("points", `${stringify(node.position.x + 75)},
${stringify(node.position.y)}
${stringify(node.position.x + 150)},${stringify(node.position.y + 40)}
${stringify(node.position.x + 75)},${stringify(node.position.y + 80)}
${stringify(node.position.x)},${stringify(node.position.y + 40)}`)}${attr("fill", color)}${attr("stroke", isSelected ? "#000" : color)}${attr("stroke-width", isSelected ? "3" : "1")}${attr("opacity", node.completed ? "0.7" : "1")}></polygon><text${attr("x", node.position.x + 75)}${attr("y", node.position.y + 40)} text-anchor="middle" dominant-baseline="middle" class="pointer-events-none fill-white text-sm font-medium">${escape_html(node.label)}</text>`);
    } else {
      $$payload.out.push("<!--[!-->");
      if (shape === "rounded") {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`<rect${attr("x", node.position.x)}${attr("y", node.position.y)} width="150" height="60" rx="30" ry="30"${attr("fill", color)}${attr("stroke", isSelected ? "#000" : color)}${attr("stroke-width", isSelected ? "3" : "1")}${attr("opacity", node.completed ? "0.7" : "1")}></rect><text${attr("x", node.position.x + 75)}${attr("y", node.position.y + 30)} text-anchor="middle" dominant-baseline="middle" class="pointer-events-none fill-white text-sm font-medium">${escape_html(node.label)}</text>`);
      } else {
        $$payload.out.push("<!--[!-->");
        $$payload.out.push(`<rect${attr("x", node.position.x)}${attr("y", node.position.y)} width="150" height="60" rx="4" ry="4"${attr("fill", color)}${attr("stroke", isSelected ? "#000" : color)}${attr("stroke-width", isSelected ? "3" : "1")}${attr("opacity", node.completed ? "0.7" : "1")}></rect><text${attr("x", node.position.x + 75)}${attr("y", node.position.y + 30)} text-anchor="middle" dominant-baseline="middle" class="pointer-events-none fill-white text-sm font-medium">${escape_html(node.label)}</text>`);
      }
      $$payload.out.push(`<!--]-->`);
    }
    $$payload.out.push(`<!--]-->`);
    if (node.completed) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<circle${attr("cx", node.position.x + 140)}${attr("cy", node.position.y + 10)} r="12" fill="#10b981"></circle><text${attr("x", node.position.x + 140)}${attr("y", node.position.y + 10)} text-anchor="middle" dominant-baseline="middle" class="pointer-events-none fill-white text-xs font-bold">✓</text>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]-->`);
    if (hasAssignments) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<g${attr("transform", `translate(${stringify(node.position.x + 5)}, ${stringify(node.position.y + 5)})`)}>`);
      if (node.tools && node.tools.length > 0) {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`<circle cx="0" cy="0" r="8" fill="#f59e0b"></circle><text x="0" y="0" text-anchor="middle" dominant-baseline="middle" class="pointer-events-none fill-white text-[10px] font-bold">T</text>`);
      } else {
        $$payload.out.push("<!--[!-->");
      }
      $$payload.out.push(`<!--]-->`);
      if (node.agents && node.agents.length > 0) {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`<circle cx="18" cy="0" r="8" fill="#3b82f6"></circle><text x="18" y="0" text-anchor="middle" dominant-baseline="middle" class="pointer-events-none fill-white text-[10px] font-bold">A</text>`);
      } else {
        $$payload.out.push("<!--[!-->");
      }
      $$payload.out.push(`<!--]-->`);
      if (node.tasks && node.tasks.length > 0) {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`<circle cx="36" cy="0" r="8" fill="#8b5cf6"></circle><text x="36" y="0" text-anchor="middle" dominant-baseline="middle" class="pointer-events-none fill-white text-[10px] font-bold">K</text>`);
      } else {
        $$payload.out.push("<!--[!-->");
      }
      $$payload.out.push(`<!--]--></g>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--></g>`);
  }
  $$payload.out.push(`<!--]--></svg>`);
  pop();
}
function _page($$payload, $$props) {
  push();
  var $$store_subs;
  const workspaceId = store_get($$store_subs ??= {}, "$page", page).params.workspaceId ?? "";
  const taskId = store_get($$store_subs ??= {}, "$page", page).params.taskId ?? "";
  const itemStore = workspaceStore.getItemStore(workspaceId);
  const task = itemStore.items.find((item) => item.id === taskId && item.type === "task");
  const workspace = workspaceStore.workspaces.find((w) => w.id === workspaceId);
  const flowchart = taskId ? itemStore.getTaskFlowchart(taskId) : void 0;
  let selectedNode = null;
  $$payload.out.push(`<div class="flex h-full flex-col bg-base-100">`);
  if (task && workspace && flowchart) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="flex-shrink-0 bg-base-200 p-4"><div class="mb-2 text-xs text-base-content/60"><button class="hover:text-base-content">Workspaces</button> <span class="mx-1">/</span> <span>${escape_html(workspace.name)}</span> <span class="mx-1">/</span> <span>Tasks</span></div> <h1 class="text-xl font-bold">${escape_html(task.title)}</h1></div> <div class="flex flex-1 overflow-hidden"><div class="flex flex-1 flex-col overflow-hidden border-r border-base-200 bg-base-200"><div class="flex-1 overflow-y-auto p-4">`);
    {
      $$payload.out.push("<!--[!-->");
      $$payload.out.push(`<div class="flex h-full items-center justify-center text-center"><div class="text-base-content/60"><p class="mb-2 text-lg font-medium">Select a Node</p> <p class="text-sm">Click on any node in the flowchart to view its details</p></div></div>`);
    }
    $$payload.out.push(`<!--]--></div></div> <div class="flex-shrink-0 overflow-auto bg-base-100"><div class="mx-auto" style="width: fit-content;">`);
    TaskFlowchart($$payload, {
      flowchart,
      selectedNodeId: selectedNode?.id
    });
    $$payload.out.push(`<!----></div></div></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
    $$payload.out.push(`<div class="flex h-full w-full items-center justify-center"><div class="text-center"><h2 class="mb-2 text-2xl font-bold">Task Not Found</h2> <p class="mb-4 text-base-content/60">The task you're looking for doesn't exist or has been deleted.</p> <button class="btn btn-primary">Back to Home</button></div></div>`);
  }
  $$payload.out.push(`<!--]--></div>`);
  if ($$store_subs) unsubscribe_stores($$store_subs);
  pop();
}
export {
  _page as default
};
