import"./DsnmJJEf.js";import{p as I,d as w,f as C,an as b,a2 as T,g as F,c as S,ao as h,ap as g,i as n,aq as m,j as U,a8 as c,ad as d}from"./TV9P7kG6.js";import{s as v,r as x}from"./DJGld0fm.js";import{I as P,S as y,g as D}from"./CnB9Cxot.js";function N(o,t){I(t,!0);/**
 * @license @lucide/svelte v0.540.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let e=x(t,["$$slots","$$events","$$legacy"]);const i=[["rect",{width:"14",height:"14",x:"8",y:"8",rx:"2",ry:"2"}],["path",{d:"M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"}]];P(o,v({name:"copy"},()=>e,{get iconNode(){return i},children:(s,a)=>{var r=w(),l=C(r);b(l,()=>t.children??T),F(s,r)},$$slots:{default:!0}})),S()}function O(o,t){I(t,!0);/**
 * @license @lucide/svelte v0.540.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let e=x(t,["$$slots","$$events","$$legacy"]);const i=[["path",{d:"m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"}],["path",{d:"M12 9v4"}],["path",{d:"M12 17h.01"}]];P(o,v({name:"triangle-alert"},()=>e,{get iconNode(){return i},children:(s,a)=>{var r=w(),l=C(r);b(l,()=>t.children??T),F(s,r)},$$slots:{default:!0}})),S()}var f=!1;class p extends Date{#t=h(super.getTime());#e=new Map;#s=g;constructor(...t){super(...t),f||this.#i()}#i(){f=!0;var t=p.prototype,e=Date.prototype,i=Object.getOwnPropertyNames(e);for(const s of i)(s.startsWith("get")||s.startsWith("to")||s==="valueOf")&&(t[s]=function(...a){if(a.length>0)return n(this.#t),e[s].apply(this,a);var r=this.#e.get(s);if(r===void 0){const l=g;m(this.#s),r=U(()=>(n(this.#t),e[s].apply(this,a))),this.#e.set(s,r),m(l)}return n(r)}),s.startsWith("set")&&(t[s]=function(...a){var r=e[s].apply(this,a);return c(this.#t,e.getTime.call(this)),r})}}class ${baseUrl;mcpClient;constructor(t="",e){this.baseUrl=t,this.mcpClient=new y({baseUrl:t,fetcher:e?.fetcher,sessionId:e?.sessionId})}#t(t){return t?new y({baseUrl:this.baseUrl,sessionId:t}):this.mcpClient}async reply(t,e,i){await this.#t(i?.sessionId).reply(t,e)}async exchange(t,e,i){return await this.#t(i?.sessionId).exchange(t,e)}async callMCPTool(t,e){const i=this.#t(e?.sessionId);try{const s=await i.exchange("tools/call",{name:t,arguments:e?.payload||{},...e?.async&&{_meta:{"ai.nanobot.async":!0,progressToken:e?.progressToken}}},{abort:e?.abort});return e?.parseResponse?e.parseResponse(s):s&&typeof s=="object"&&"structuredContent"in s?s.structuredContent:s}catch(s){try{const a=D(),r=s instanceof Error?s.message:String(s);a.error("API Error",r)}catch{console.error("MCP Tool Error:",s)}throw s}}async capabilities(){const t=this.#t(),{initializeResult:e}=await t.getSessionDetails();return e?.capabilities?.experimental?.["ai.nanobot"]?.session??{}}async deleteThread(t){return this.#t(t).deleteSession()}async renameThread(t,e){return await this.callMCPTool("update_chat",{payload:{chatId:t,title:e}})}async listAgents(t){return await this.callMCPTool("list_agents",t)}async getThreads(){return(await this.callMCPTool("list_chats")).chats}async createThread(){const t=this.#t("new"),{id:e}=await t.getSessionDetails();return{id:e,title:"New Chat",created:new p().toISOString()}}async createResource(t,e,i,s){return await this.callMCPTool("create_resource",{payload:{blob:i,mimeType:e,name:t,...s?.description&&{description:s.description}},sessionId:s?.sessionId,abort:s?.abort,parseResponse:a=>a.content?.[0]?.type==="resource_link"?{uri:a.content[0].uri}:{uri:""}})}async sendMessage(t){return await this.callMCPTool("chat",{payload:{prompt:t.message,attachments:t.attachments?.map(i=>({name:i.name,url:i.uri,mimeType:i.mimeType}))},sessionId:t.threadId,progressToken:t.id,async:!0}),{message:{id:t.id,role:"user",created:_(),items:[{id:t.id+"_0",type:"text",text:t.message}]}}}subscribe(t,e,i){console.log("Subscribing to thread:",t);const s=new EventSource(`${this.baseUrl}/api/events/${t}`);s.onmessage=a=>{const r=JSON.parse(a.data);e({type:"message",message:r})};for(const a of i?.events??[])s.addEventListener(a,r=>{const l=parseInt(r.lastEventId);e({id:l||r.lastEventId,type:a,data:JSON.parse(r.data)})});return s.onerror=a=>{e({type:"error",error:String(a)}),console.error("EventSource failed:",a),s.close()},s.onopen=()=>{console.log("EventSource connected for thread:",t)},()=>s.close()}}function u(o,t){let e=!1;return t.id&&(o=o.map(i=>i.id===t.id?(e=!0,t):i)),e||(o=[...o,t]),o}const L=new $;class k{#t;get messages(){return n(this.#t)}set messages(t){c(this.#t,t,!0)}#e;get history(){return n(this.#e)}set history(t){c(this.#e,t,!0)}#s;get isLoading(){return n(this.#s)}set isLoading(t){c(this.#s,t,!0)}#i;get elicitations(){return n(this.#i)}set elicitations(t){c(this.#i,t,!0)}#a;get prompts(){return n(this.#a)}set prompts(t){c(this.#a,t,!0)}#r;get resources(){return n(this.#r)}set resources(t){c(this.#r,t,!0)}#n;get chatId(){return n(this.#n)}set chatId(t){c(this.#n,t,!0)}#o;get agent(){return n(this.#o)}set agent(t){c(this.#o,t,!0)}#h;get uploadedFiles(){return n(this.#h)}set uploadedFiles(t){c(this.#h,t,!0)}#c;get uploadingFiles(){return n(this.#c)}set uploadingFiles(t){c(this.#c,t,!0)}api;closer=()=>{};onChatDone=[];constructor(t){this.api=t?.api||L,this.#t=h(d([])),this.#e=h(),this.#s=h(!1),this.#i=h(d([])),this.#a=h(d([])),this.#r=h(d([])),this.#n=h(""),this.#o=h(d({})),this.#h=h(d([])),this.#c=h(d([])),this.setChatId(t?.chatId)}close=()=>{this.closer(),this.setChatId("")};setChatId=async t=>{t!==this.chatId&&(this.messages=[],this.prompts=[],this.resources=[],this.elicitations=[],this.history=void 0,this.isLoading=!1,this.uploadedFiles=[],this.uploadingFiles=[],t&&(this.chatId=t,this.subscribe(t)),this.listResources().then(e=>{e&&e.resources&&(this.resources=e.resources)}),this.listPrompts().then(e=>{e&&e.prompts&&(this.prompts=e.prompts)}),await this.reloadAgent())};reloadAgent=async()=>{const t=await this.api.listAgents({sessionId:this.chatId});t.agents?.length>0&&(this.agent=t.agents[0])};listPrompts=async()=>await this.api.exchange("prompts/list",{},{sessionId:this.chatId});listResources=async()=>await this.api.exchange("resources/list",{},{sessionId:this.chatId});subscribe(t){this.closer(),t&&(this.closer=this.api.subscribe(t,e=>{if(e.type=="message"&&e.message?.id)this.history?this.history=u(this.history,e.message):this.messages=u(this.messages,e.message);else if(e.type=="history-start")this.history=[];else if(e.type=="history-end")this.messages=this.history||[],this.history=void 0;else if(e.type=="chat-in-progress")this.isLoading=!0;else if(e.type=="chat-done"){this.isLoading=!1;for(const i of this.onChatDone)i();this.onChatDone=[]}else e.type=="elicitation/create"&&(this.elicitations=[...this.elicitations,{id:e.id,...e.data}]);console.debug("Received event:",e)},{events:["history-start","history-end","chat-in-progress","chat-done","elicitation/create"]}))}replyToElicitation=async(t,e)=>{await this.api.reply(t.id,e,{sessionId:this.chatId}),this.elicitations=this.elicitations.filter(i=>i.id!==t.id)};newChat=async()=>{const t=await this.api.createThread();await this.setChatId(t.id)};sendMessage=async(t,e)=>{if(!(!t.trim()||this.isLoading)){this.isLoading=!0,this.chatId||await this.newChat();try{const i=await this.api.sendMessage({id:crypto.randomUUID(),threadId:this.chatId,message:t,attachments:[...this.uploadedFiles,...e||[]]});return this.uploadedFiles=[],this.messages=u(this.messages,i.message),new Promise(s=>{this.onChatDone.push(()=>{this.isLoading=!1;const a=this.messages.findIndex(r=>r.id===i.message.id);a!==-1&&a<=this.messages.length?s({message:this.messages[a+1]}):s()})})}catch(i){this.isLoading=!1,this.messages=u(this.messages,{id:crypto.randomUUID(),role:"assistant",created:_(),items:[{id:crypto.randomUUID(),type:"text",text:`Sorry, I couldn't send your message. Please try again. Error: ${i}`}]})}}};cancelUpload=t=>{this.uploadingFiles=this.uploadingFiles.filter(e=>e.id!==t?!0:(e.controller&&e.controller.abort(),!1)),this.uploadedFiles=this.uploadedFiles.filter(e=>e.id!==t)};uploadFile=async(t,e)=>{if(!this.chatId){const a=await this.api.createThread();await this.setChatId(a.id)}const i=crypto.randomUUID(),s=e?.controller||new AbortController;this.uploadingFiles.push({file:t,id:i,controller:s});try{const a=await this.doUploadFile(t,s);return this.uploadedFiles.push({file:t,uri:a.uri,id:i,mimeType:a.mimeType}),a}finally{this.uploadingFiles=this.uploadingFiles.filter(a=>a.id!==i)}};doUploadFile=async(t,e)=>{const i=new FileReader;i.readAsDataURL(t),await new Promise((a,r)=>{i.onloadend=a,i.onerror=r});const s=i.result.split(",")[1];if(!this.chatId)throw new Error("Chat ID not set");return await this.api.createResource(t.name,t.type,s,{description:t.name,sessionId:this.chatId,abort:e})}}function _(){return new Date().toISOString()}export{k as C,p as S,O as T,N as a,L as d};
