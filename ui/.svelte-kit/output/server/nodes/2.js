

export const index = 2;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/_page.svelte.js')).default;
export const universal = {
  "ssr": false
};
export const universal_id = "src/routes/+page.ts";
export const imports = ["_app/immutable/nodes/2.CQmKVgES.js","_app/immutable/chunks/DsnmJJEf.js","_app/immutable/chunks/DO59wc33.js","_app/immutable/chunks/ik-EZS2y.js","_app/immutable/chunks/CszUx5kC.js","_app/immutable/chunks/DWFGZehV.js","_app/immutable/chunks/DQqIutZC.js","_app/immutable/chunks/PXDOYhLw.js","_app/immutable/chunks/DYffSaQ-.js","_app/immutable/chunks/D7RuEkZb.js","_app/immutable/chunks/Bc2WQhPO.js","_app/immutable/chunks/jJgo0auY.js"];
export const stylesheets = ["_app/immutable/assets/chat.Bnxpj5AT.css"];
export const fonts = [];
