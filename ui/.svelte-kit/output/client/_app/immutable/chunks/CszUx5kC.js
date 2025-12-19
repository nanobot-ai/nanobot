import"./DsnmJJEf.js";import{p as g,d as m,f as y,an as f,a2 as I,g as w,c as C,i as o,a8 as h,ao as c,ad as l}from"./DO59wc33.js";import{s as b,r as T}from"./ik-EZS2y.js";import{I as F,S as p,g as x,a as P}from"./DWFGZehV.js";function A(n,t){g(t,!0);/**
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
 */let s=T(t,["$$slots","$$events","$$legacy"]);const e=[["rect",{width:"14",height:"14",x:"8",y:"8",rx:"2",ry:"2"}],["path",{d:"M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"}]];F(n,b({name:"copy"},()=>s,{get iconNode(){return e},children:(i,a)=>{var r=m(),d=y(r);f(d,()=>t.children??I),w(i,r)},$$slots:{default:!0}})),C()}function E(n,t){g(t,!0);/**
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
 */let s=T(t,["$$slots","$$events","$$legacy"]);const e=[["path",{d:"m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"}],["path",{d:"M12 9v4"}],["path",{d:"M12 17h.01"}]];F(n,b({name:"triangle-alert"},()=>s,{get iconNode(){return e},children:(i,a)=>{var r=m(),d=y(r);f(d,()=>t.children??I),w(i,r)},$$slots:{default:!0}})),C()}class U{baseUrl;mcpClient;constructor(t="",s){this.baseUrl=t,this.mcpClient=new p({baseUrl:t,fetcher:s?.fetcher,sessionId:s?.sessionId})}#t(t){return t?new p({baseUrl:this.baseUrl,sessionId:t}):this.mcpClient}async reply(t,s,e){await this.#t(e?.sessionId).reply(t,s)}async exchange(t,s,e){return await this.#t(e?.sessionId).exchange(t,s)}async callMCPTool(t,s){const e=this.#t(s?.sessionId);try{const i=await e.exchange("tools/call",{name:t,arguments:s?.payload||{},...s?.async&&{_meta:{"ai.nanobot.async":!0,progressToken:s?.progressToken}}},{abort:s?.abort});return s?.parseResponse?s.parseResponse(i):i&&typeof i=="object"&&"structuredContent"in i?i.structuredContent:i}catch(i){try{const a=x(),r=i instanceof Error?i.message:String(i);a.error("API Error",r)}catch{console.error("MCP Tool Error:",i)}throw i}}async capabilities(){const t=this.#t(),{initializeResult:s}=await t.getSessionDetails();return s?.capabilities?.experimental?.["ai.nanobot"]?.session??{}}async deleteThread(t){return this.#t(t).deleteSession()}async renameThread(t,s){return await this.callMCPTool("update_chat",{payload:{chatId:t,title:s}})}async listAgents(t){return await this.callMCPTool("list_agents",t)}async getThreads(){return(await this.callMCPTool("list_chats")).chats}async createThread(){const t=this.#t("new"),{id:s}=await t.getSessionDetails();return{id:s,title:"New Chat",created:new P().toISOString()}}async createResource(t,s,e,i){return await this.callMCPTool("create_resource",{payload:{blob:e,mimeType:s,name:t,...i?.description&&{description:i.description}},sessionId:i?.sessionId,abort:i?.abort,parseResponse:a=>a.content?.[0]?.type==="resource_link"?{uri:a.content[0].uri}:{uri:""}})}async sendMessage(t){return await this.callMCPTool("chat",{payload:{prompt:t.message,attachments:t.attachments?.map(e=>({name:e.name,url:e.uri,mimeType:e.mimeType}))},sessionId:t.threadId,progressToken:t.id,async:!0}),{message:{id:t.id,role:"user",created:S(),items:[{id:t.id+"_0",type:"text",text:t.message}]}}}subscribe(t,s,e){console.log("Subscribing to thread:",t);const i=new EventSource(`${this.baseUrl}/api/events/${t}`);i.onmessage=a=>{const r=JSON.parse(a.data);s({type:"message",message:r})};for(const a of e?.events??[])i.addEventListener(a,r=>{const d=parseInt(r.lastEventId);s({id:d||r.lastEventId,type:a,data:JSON.parse(r.data)})});return i.onerror=a=>{s({type:"error",error:String(a)}),console.error("EventSource failed:",a),i.close()},i.onopen=()=>{console.log("EventSource connected for thread:",t)},()=>i.close()}}function u(n,t){let s=!1;return t.id&&(n=n.map(e=>e.id===t.id?(s=!0,t):e)),s||(n=[...n,t]),n}const $=new U;class M{#t;get messages(){return o(this.#t)}set messages(t){h(this.#t,t,!0)}#s;get history(){return o(this.#s)}set history(t){h(this.#s,t,!0)}#e;get isLoading(){return o(this.#e)}set isLoading(t){h(this.#e,t,!0)}#i;get elicitations(){return o(this.#i)}set elicitations(t){h(this.#i,t,!0)}#a;get prompts(){return o(this.#a)}set prompts(t){h(this.#a,t,!0)}#r;get resources(){return o(this.#r)}set resources(t){h(this.#r,t,!0)}#n;get chatId(){return o(this.#n)}set chatId(t){h(this.#n,t,!0)}#o;get agent(){return o(this.#o)}set agent(t){h(this.#o,t,!0)}#h;get uploadedFiles(){return o(this.#h)}set uploadedFiles(t){h(this.#h,t,!0)}#c;get uploadingFiles(){return o(this.#c)}set uploadingFiles(t){h(this.#c,t,!0)}api;closer=()=>{};onChatDone=[];constructor(t){this.api=t?.api||$,this.#t=c(l([])),this.#s=c(),this.#e=c(!1),this.#i=c(l([])),this.#a=c(l([])),this.#r=c(l([])),this.#n=c(""),this.#o=c(l({})),this.#h=c(l([])),this.#c=c(l([])),this.setChatId(t?.chatId)}close=()=>{this.closer(),this.setChatId("")};setChatId=async t=>{t!==this.chatId&&(this.messages=[],this.prompts=[],this.resources=[],this.elicitations=[],this.history=void 0,this.isLoading=!1,this.uploadedFiles=[],this.uploadingFiles=[],t&&(this.chatId=t,this.subscribe(t)),this.listResources().then(s=>{s&&s.resources&&(this.resources=s.resources)}),this.listPrompts().then(s=>{s&&s.prompts&&(this.prompts=s.prompts)}),await this.reloadAgent())};reloadAgent=async()=>{const t=await this.api.listAgents({sessionId:this.chatId});t.agents?.length>0&&(this.agent=t.agents[0])};listPrompts=async()=>await this.api.exchange("prompts/list",{},{sessionId:this.chatId});listResources=async()=>await this.api.exchange("resources/list",{},{sessionId:this.chatId});subscribe(t){this.closer(),t&&(this.closer=this.api.subscribe(t,s=>{if(s.type=="message"&&s.message?.id)this.history?this.history=u(this.history,s.message):this.messages=u(this.messages,s.message);else if(s.type=="history-start")this.history=[];else if(s.type=="history-end")this.messages=this.history||[],this.history=void 0;else if(s.type=="chat-in-progress")this.isLoading=!0;else if(s.type=="chat-done"){this.isLoading=!1;for(const e of this.onChatDone)e();this.onChatDone=[]}else s.type=="elicitation/create"&&(this.elicitations=[...this.elicitations,{id:s.id,...s.data}]);console.debug("Received event:",s)},{events:["history-start","history-end","chat-in-progress","chat-done","elicitation/create"]}))}replyToElicitation=async(t,s)=>{await this.api.reply(t.id,s,{sessionId:this.chatId}),this.elicitations=this.elicitations.filter(e=>e.id!==t.id)};newChat=async()=>{const t=await this.api.createThread();await this.setChatId(t.id)};sendMessage=async(t,s)=>{if(!(!t.trim()||this.isLoading)){this.isLoading=!0,this.chatId||await this.newChat();try{const e=await this.api.sendMessage({id:crypto.randomUUID(),threadId:this.chatId,message:t,attachments:[...this.uploadedFiles,...s||[]]});return this.uploadedFiles=[],this.messages=u(this.messages,e.message),new Promise(i=>{this.onChatDone.push(()=>{this.isLoading=!1;const a=this.messages.findIndex(r=>r.id===e.message.id);a!==-1&&a<=this.messages.length?i({message:this.messages[a+1]}):i()})})}catch(e){this.isLoading=!1,this.messages=u(this.messages,{id:crypto.randomUUID(),role:"assistant",created:S(),items:[{id:crypto.randomUUID(),type:"text",text:`Sorry, I couldn't send your message. Please try again. Error: ${e}`}]})}}};cancelUpload=t=>{this.uploadingFiles=this.uploadingFiles.filter(s=>s.id!==t?!0:(s.controller&&s.controller.abort(),!1)),this.uploadedFiles=this.uploadedFiles.filter(s=>s.id!==t)};uploadFile=async(t,s)=>{if(!this.chatId){const a=await this.api.createThread();await this.setChatId(a.id)}const e=crypto.randomUUID(),i=s?.controller||new AbortController;this.uploadingFiles.push({file:t,id:e,controller:i});try{const a=await this.doUploadFile(t,i);return this.uploadedFiles.push({file:t,uri:a.uri,id:e,mimeType:a.mimeType}),a}finally{this.uploadingFiles=this.uploadingFiles.filter(a=>a.id!==e)}};doUploadFile=async(t,s)=>{const e=new FileReader;e.readAsDataURL(t),await new Promise((a,r)=>{e.onloadend=a,e.onerror=r});const i=e.result.split(",")[1];if(!this.chatId)throw new Error("Chat ID not set");return await this.api.createResource(t.name,t.type,i,{description:t.name,sessionId:this.chatId,abort:s})}}function S(){return new Date().toISOString()}export{M as C,E as T,A as a,$ as d};
