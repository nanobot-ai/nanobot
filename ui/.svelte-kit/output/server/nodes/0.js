

export const index = 0;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/_layout.svelte.js')).default;
export const universal = {
  "ssr": false
};
export const universal_id = "src/routes/+layout.ts";
export const imports = ["_app/immutable/nodes/0.CD3cMGvk.js","_app/immutable/chunks/DsnmJJEf.js","_app/immutable/chunks/DO59wc33.js","_app/immutable/chunks/ik-EZS2y.js","_app/immutable/chunks/DWFGZehV.js","_app/immutable/chunks/CszUx5kC.js","_app/immutable/chunks/D7RuEkZb.js","_app/immutable/chunks/jJgo0auY.js","_app/immutable/chunks/ByRzogd4.js","_app/immutable/chunks/Bc2WQhPO.js"];
export const stylesheets = ["_app/immutable/assets/chat.Bnxpj5AT.css","_app/immutable/assets/0.t5HqxhbQ.css"];
export const fonts = [];
