import"../chunks/DsnmJJEf.js";import{w as Qt,be as Yt,Q as ea,bf as ta,bg as aa,e as ra,u as na,bh as sa,bi as oa,bj as ia,bk as It,ak as la,q as ca,a2 as G,p as Q,d as V,f as L,an as le,g as c,c as Y,ap as ze,k as m,s as v,l as s,m as n,i as t,ao as _e,bl as ut,a8 as C,t as se,O as be,ar as Ae,j as q,a as da,ad as Tt,o as va,h as ua}from"../chunks/DO59wc33.js";import{s as ce,r as de,p as _a,i as $,b as fa,a as ha}from"../chunks/ik-EZS2y.js";import{I as ve,e as Fe,i as Ft,r as Qe,c as Ye,X as Le,b as Ie,s as xe,g as ba,h as pa,a as ga,j as ma,k as xa}from"../chunks/DWFGZehV.js";import{a as wa,T as ka,d as Ze}from"../chunks/CszUx5kC.js";import{g as ya}from"../chunks/D7RuEkZb.js";import{r as dt}from"../chunks/jJgo0auY.js";import{w as ie,P as $a,L as Ca,B as Na,p as Ma}from"../chunks/ByRzogd4.js";import{p as Sa}from"../chunks/Bc2WQhPO.js";const Ia=()=>performance.now(),Ne={tick:r=>requestAnimationFrame(r),now:()=>Ia(),tasks:new Set};function Et(){const r=Ne.now();Ne.tasks.forEach(e=>{e.c(r)||(Ne.tasks.delete(e),e.f())}),Ne.tasks.size!==0&&Ne.tick(Et)}function Ta(r){let e;return Ne.tasks.size===0&&Ne.tick(Et),{promise:new Promise(a=>{Ne.tasks.add(e={c:r,f:a})}),abort(){Ne.tasks.delete(e)}}}function Ge(r,e){It(()=>{r.dispatchEvent(new CustomEvent(e))})}function Fa(r){if(r==="float")return"cssFloat";if(r==="offset")return"cssOffset";if(r.startsWith("--"))return r;const e=r.split("-");return e.length===1?e[0]:e[0]+e.slice(1).map(a=>a[0].toUpperCase()+a.slice(1)).join("")}function $t(r){const e={},a=r.split(";");for(const o of a){const[l,u]=o.split(":");if(!l||u===void 0)break;const i=Fa(l.trim());e[i]=u.trim()}return e}const Ea=r=>r;function Ct(r,e,a,o){var l=(r&oa)!==0,u=(r&ia)!==0,i=l&&u,d=(r&sa)!==0,F=i?"both":l?"in":"out",_,E=e.inert,R=e.style.overflow,M,Z;function H(){return It(()=>_??=a()(e,o?.()??{},{direction:F}))}var x={is_global:d,in(){if(e.inert=E,!l){Z?.abort(),Z?.reset?.();return}u||M?.abort(),Ge(e,"introstart"),M=vt(e,H(),Z,1,()=>{Ge(e,"introend"),M?.abort(),M=_=void 0,e.style.overflow=R})},out(w){if(!u){w?.(),_=void 0;return}e.inert=!0,Ge(e,"outrostart"),Z=vt(e,H(),M,0,()=>{Ge(e,"outroend"),w?.()})},stop:()=>{M?.abort(),Z?.abort()}},k=Qt;if((k.transitions??=[]).push(x),l&&Yt){var g=d;if(!g){for(var N=k.parent;N&&(N.f&ea)!==0;)for(;(N=N.parent)&&(N.f&ta)===0;);g=!N||(N.f&aa)!==0}g&&ra(()=>{na(()=>x.in())})}}function vt(r,e,a,o,l){var u=o===1;if(la(e)){var i,d=!1;return ca(()=>{if(!d){var k=e({direction:u?"in":"out"});i=vt(r,k,a,o,l)}}),{abort:()=>{d=!0,i?.abort()},deactivate:()=>i.deactivate(),reset:()=>i.reset(),t:()=>i.t()}}if(a?.deactivate(),!e?.duration)return l(),{abort:G,deactivate:G,reset:G,t:()=>o};const{delay:F=0,css:_,tick:E,easing:R=Ea}=e;var M=[];if(u&&a===void 0&&(E&&E(0,1),_)){var Z=$t(_(0,1));M.push(Z,Z)}var H=()=>1-o,x=r.animate(M,{duration:F,fill:"forwards"});return x.onfinish=()=>{x.cancel();var k=a?.t()??1-o;a?.abort();var g=o-k,N=e.duration*Math.abs(g),w=[];if(N>0){var W=!1;if(_)for(var ae=Math.ceil(N/16.666666666666668),J=0;J<=ae;J+=1){var I=k+g*R(J/ae),re=$t(_(I,1-I));w.push(re),W||=re.overflow==="hidden"}W&&(r.style.overflow="hidden"),H=()=>{var B=x.currentTime;return k+g*R(B/N)},E&&Ta(()=>{if(x.playState!=="running")return!1;var B=H();return E(B,1-B),!0})}x=r.animate(w,{duration:N,fill:"forwards"}),x.onfinish=()=>{H=()=>o,E?.(o,1-o),l()}},{abort:()=>{x&&(x.cancel(),x.effect=null,x.onfinish=G)},deactivate:()=>{l=G},reset:()=>{o===0&&E?.(1,0)},t:()=>H()}}const Pa=!1,In=Object.freeze(Object.defineProperty({__proto__:null,ssr:Pa},Symbol.toStringTag,{value:"Module"})),Wa="data:image/svg+xml,%3c?xml%20version='1.0'%20encoding='UTF-8'?%3e%3csvg%20id='Layer_1'%20data-name='Layer%201'%20xmlns='http://www.w3.org/2000/svg'%20viewBox='0%200%20512%20512'%3e%3cdefs%3e%3cstyle%3e%20.cls-1%20{%20fill:%20%23fff;%20}%20%3c/style%3e%3c/defs%3e%3cpath%20class='cls-1'%20d='M348.33,145.39h-14.33v8.86c0,13.77-5.47,23.43-15.21,33.17-9.74,9.74-22.94,8.11-36.71,8.11h-55.69c-13.77,0-23.43,1.62-33.17-8.11-9.74-9.74-15.21-19.4-15.21-33.17v-8.86h-14.32c-24.89,0-44.77,19.88-44.77,44.77v122.85c0,24.88,19.89,44.77,44.77,44.77h184.65c24.88,0,44.77-19.88,44.77-44.77v-122.85c0-24.89-19.88-44.77-44.77-44.77ZM247.33,284.01c0,5.59-4.53,10.13-10.13,10.13h-39.35c-5.6,0-10.13-4.54-10.13-10.13v-40.62c0-5.59,4.53-10.13,10.13-10.13h39.35c5.6,0,10.13,4.54,10.13,10.13v40.62ZM324.29,284.01c0,5.59-4.54,10.13-10.14,10.13h-39.34c-5.59,0-10.14-4.54-10.14-10.13v-40.62c0-5.59,4.54-10.13,10.14-10.13h39.34c5.6,0,10.14,4.54,10.14,10.13v40.62Z'/%3e%3cpath%20d='M330.95,408.78h-48.98c-3.94,0-7.72,1.57-10.51,4.36-2.79,2.79-4.36,6.57-4.36,10.52v48.04c0,18.94,15.36,34.3,34.3,34.3,11.06,0,21.41-5.47,27.65-14.6,6.87-10.06,16.4-24.01,25.67-37.59,6.02-8.82,6.66-20.24,1.68-29.68-4.98-9.44-14.78-15.35-25.45-15.35Z'/%3e%3cpath%20d='M470.99,212.16c-8.36,0-15.23,6.88-15.23,15.23v18.41h-29.34v-55.65c0-42.92-35.17-78.09-78.09-78.09h-9.02l14.22-53.17c11.11-3.46,18.75-13.99,18.76-25.85,0-14.84-11.95-27.04-26.48-27.04h0c-14.53,0-26.48,12.21-26.48,27.04.03,5.8,1.88,11.44,5.28,16.09l-16.82,62.94h-103.55l-16.82-62.94c3.4-4.64,5.26-10.29,5.28-16.09,0-14.84-11.95-27.04-26.48-27.04h0c-14.53,0-26.48,12.21-26.48,27.04,0,11.86,7.65,22.39,18.76,25.85l14.22,53.17h-9.02c-42.92,0-78.09,35.17-78.09,78.09v55.65h-29.34v-18.41c0-8.36-6.88-15.23-15.23-15.23s-15.23,6.88-15.23,15.23v68.81c.12,8.28,6.96,15.02,15.23,15.02s15.11-6.74,15.23-15.02v-19.94h29.34v36.74c0,42.92,35.18,78.09,78.09,78.09h184.65c42.92,0,78.09-35.17,78.09-78.09v-36.74h29.34v19.94c.12,8.28,6.96,15.02,15.23,15.02s15.11-6.74,15.23-15.02v-68.81c0-8.36-6.88-15.23-15.23-15.23ZM393.09,313c0,24.88-19.88,44.77-44.77,44.77h-184.65c-24.89,0-44.77-19.88-44.77-44.77v-122.85c0-24.89,19.89-44.77,44.77-44.77h14.32v8.86c0,13.77,5.47,23.43,15.21,33.17,9.74,9.74,19.4,8.11,33.17,8.11h55.69c13.77,0,26.97,1.62,36.71-8.11,9.74-9.74,15.21-19.4,15.21-33.17v-8.86h14.33c24.88,0,44.77,19.88,44.77,44.77v122.85Z'/%3e%3cpath%20d='M227.28,94.04c1.8,2.37,4.61,3.77,7.59,3.77s5.79-1.4,7.59-3.77l14.21-18.74,13.63,17.98c1.8,2.37,4.61,3.77,7.59,3.77s5.79-1.4,7.59-3.77l17.23-22.74c3.15-4.16,2.33-10.18-1.83-13.34-1.66-1.26-3.68-1.94-5.76-1.94-2.98,0-5.79,1.39-7.59,3.77l-9.65,12.72-13.63-17.98c-1.8-2.37-4.61-3.76-7.59-3.76h0c-2.98,0-5.79,1.39-7.59,3.76l-14.21,18.75-10.39-13.71c-3.16-4.16-9.18-4.98-13.34-1.83-4.16,3.15-4.99,9.17-1.84,13.34l17.98,23.71Z'/%3e%3cpath%20d='M230.04,408.78h-48.98c-10.67,0-20.47,5.91-25.45,15.35-4.98,9.44-4.34,20.86,1.68,29.68,9.27,13.58,18.8,27.53,25.67,37.59,6.24,9.13,16.59,14.6,27.64,14.6h0c18.94,0,34.3-15.36,34.3-34.3v-48.04c0-3.94-1.57-7.73-4.36-10.52s-6.57-4.36-10.52-4.36Z'/%3e%3cpath%20d='M197.84,233.26h39.35c5.59,0,10.13,4.54,10.13,10.13v40.62c0,5.59-4.54,10.13-10.13,10.13h-39.35c-5.59,0-10.13-4.54-10.13-10.13v-40.62c0-5.59,4.54-10.13,10.13-10.13Z'/%3e%3cpath%20d='M314.15,233.26h-39.34c-5.59,0-10.14,4.54-10.14,10.13v40.62c0,5.59,4.54,10.13,10.14,10.13h39.34c5.6,0,10.14-4.54,10.14-10.13v-40.62c0-5.59-4.54-10.13-10.14-10.13Z'/%3e%3c/svg%3e",Nt=""+new URL("../assets/nanobot.Bn3X0Wtr.svg",import.meta.url).href;function et(r,e){Q(e,!0);/**
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
 */let a=de(e,["$$slots","$$events","$$legacy"]);const o=[["path",{d:"M20 6 9 17l-5-5"}]];ve(r,ce({name:"check"},()=>a,{get iconNode(){return o},children:(l,u)=>{var i=V(),d=L(i);le(d,()=>e.children??G),c(l,i)},$$slots:{default:!0}})),Y()}function De(r,e){Q(e,!0);/**
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
 */let a=de(e,["$$slots","$$events","$$legacy"]);const o=[["path",{d:"m9 18 6-6-6-6"}]];ve(r,ce({name:"chevron-right"},()=>a,{get iconNode(){return o},children:(l,u)=>{var i=V(),d=L(i);le(d,()=>e.children??G),c(l,i)},$$slots:{default:!0}})),Y()}function Da(r,e){Q(e,!0);/**
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
 */let a=de(e,["$$slots","$$events","$$legacy"]);const o=[["circle",{cx:"12",cy:"12",r:"10"}],["line",{x1:"12",x2:"12",y1:"8",y2:"12"}],["line",{x1:"12",x2:"12.01",y1:"16",y2:"16"}]];ve(r,ce({name:"circle-alert"},()=>a,{get iconNode(){return o},children:(l,u)=>{var i=V(),d=L(i);le(d,()=>e.children??G),c(l,i)},$$slots:{default:!0}})),Y()}function Ra(r,e){Q(e,!0);/**
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
 */let a=de(e,["$$slots","$$events","$$legacy"]);const o=[["path",{d:"M21.801 10A10 10 0 1 1 17 3.335"}],["path",{d:"m9 11 3 3L22 4"}]];ve(r,ce({name:"circle-check-big"},()=>a,{get iconNode(){return o},children:(l,u)=>{var i=V(),d=L(i);le(d,()=>e.children??G),c(l,i)},$$slots:{default:!0}})),Y()}function _t(r,e){Q(e,!0);/**
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
 */let a=de(e,["$$slots","$$events","$$legacy"]);const o=[["circle",{cx:"12",cy:"12",r:"1"}],["circle",{cx:"12",cy:"5",r:"1"}],["circle",{cx:"12",cy:"19",r:"1"}]];ve(r,ce({name:"ellipsis-vertical"},()=>a,{get iconNode(){return o},children:(l,u)=>{var i=V(),d=L(i);le(d,()=>e.children??G),c(l,i)},$$slots:{default:!0}})),Y()}function Aa(r,e){Q(e,!0);/**
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
 */let a=de(e,["$$slots","$$events","$$legacy"]);const o=[["path",{d:"M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"}],["path",{d:"M14 2v4a2 2 0 0 0 2 2h4"}],["path",{d:"M10 9H8"}],["path",{d:"M16 13H8"}],["path",{d:"M16 17H8"}]];ve(r,ce({name:"file-text"},()=>a,{get iconNode(){return o},children:(l,u)=>{var i=V(),d=L(i);le(d,()=>e.children??G),c(l,i)},$$slots:{default:!0}})),Y()}function La(r,e){Q(e,!0);/**
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
 */let a=de(e,["$$slots","$$events","$$legacy"]);const o=[["path",{d:"M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"}],["path",{d:"M14 2v4a2 2 0 0 0 2 2h4"}]];ve(r,ce({name:"file"},()=>a,{get iconNode(){return o},children:(l,u)=>{var i=V(),d=L(i);le(d,()=>e.children??G),c(l,i)},$$slots:{default:!0}})),Y()}function Pt(r,e){Q(e,!0);/**
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
 */let a=de(e,["$$slots","$$events","$$legacy"]);const o=[["path",{d:"m6 14 1.5-2.9A2 2 0 0 1 9.24 10H20a2 2 0 0 1 1.94 2.5l-1.54 6a2 2 0 0 1-1.95 1.5H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h3.9a2 2 0 0 1 1.69.9l.81 1.2a2 2 0 0 0 1.67.9H18a2 2 0 0 1 2 2v2"}]];ve(r,ce({name:"folder-open"},()=>a,{get iconNode(){return o},children:(l,u)=>{var i=V(),d=L(i);le(d,()=>e.children??G),c(l,i)},$$slots:{default:!0}})),Y()}function Wt(r,e){Q(e,!0);/**
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
 */let a=de(e,["$$slots","$$events","$$legacy"]);const o=[["path",{d:"M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"}]];ve(r,ce({name:"folder"},()=>a,{get iconNode(){return o},children:(l,u)=>{var i=V(),d=L(i);le(d,()=>e.children??G),c(l,i)},$$slots:{default:!0}})),Y()}function za(r,e){Q(e,!0);/**
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
 */let a=de(e,["$$slots","$$events","$$legacy"]);const o=[["circle",{cx:"12",cy:"12",r:"10"}],["path",{d:"M12 16v-4"}],["path",{d:"M12 8h.01"}]];ve(r,ce({name:"info"},()=>a,{get iconNode(){return o},children:(l,u)=>{var i=V(),d=L(i);le(d,()=>e.children??G),c(l,i)},$$slots:{default:!0}})),Y()}function Oa(r,e){Q(e,!0);/**
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
 */let a=de(e,["$$slots","$$events","$$legacy"]);const o=[["path",{d:"M4 12h16"}],["path",{d:"M4 18h16"}],["path",{d:"M4 6h16"}]];ve(r,ce({name:"menu"},()=>a,{get iconNode(){return o},children:(l,u)=>{var i=V(),d=L(i);le(d,()=>e.children??G),c(l,i)},$$slots:{default:!0}})),Y()}function Ua(r,e){Q(e,!0);/**
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
 */let a=de(e,["$$slots","$$events","$$legacy"]);const o=[["path",{d:"M22 17a2 2 0 0 1-2 2H6.828a2 2 0 0 0-1.414.586l-2.202 2.202A.71.71 0 0 1 2 21.286V5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2z"}]];ve(r,ce({name:"message-square"},()=>a,{get iconNode(){return o},children:(l,u)=>{var i=V(),d=L(i);le(d,()=>e.children??G),c(l,i)},$$slots:{default:!0}})),Y()}function Za(r,e){Q(e,!0);/**
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
 */let a=de(e,["$$slots","$$events","$$legacy"]);const o=[["path",{d:"M20.985 12.486a9 9 0 1 1-9.473-9.472c.405-.022.617.46.402.803a6 6 0 0 0 8.268 8.268c.344-.215.825-.004.803.401"}]];ve(r,ce({name:"moon"},()=>a,{get iconNode(){return o},children:(l,u)=>{var i=V(),d=L(i);le(d,()=>e.children??G),c(l,i)},$$slots:{default:!0}})),Y()}function Ha(r,e){Q(e,!0);/**
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
 */let a=de(e,["$$slots","$$events","$$legacy"]);const o=[["rect",{width:"18",height:"18",x:"3",y:"3",rx:"2"}],["path",{d:"M9 3v18"}],["path",{d:"m16 15-3-3 3-3"}]];ve(r,ce({name:"panel-left-close"},()=>a,{get iconNode(){return o},children:(l,u)=>{var i=V(),d=L(i);le(d,()=>e.children??G),c(l,i)},$$slots:{default:!0}})),Y()}function Mt(r,e){Q(e,!0);/**
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
 */let a=de(e,["$$slots","$$events","$$legacy"]);const o=[["rect",{width:"18",height:"18",x:"3",y:"3",rx:"2"}],["path",{d:"M9 3v18"}],["path",{d:"m14 9 3 3-3 3"}]];ve(r,ce({name:"panel-left-open"},()=>a,{get iconNode(){return o},children:(l,u)=>{var i=V(),d=L(i);le(d,()=>e.children??G),c(l,i)},$$slots:{default:!0}})),Y()}function Re(r,e){Q(e,!0);/**
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
 */let a=de(e,["$$slots","$$events","$$legacy"]);const o=[["path",{d:"M12 3H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"}],["path",{d:"M18.375 2.625a1 1 0 0 1 3 3l-9.013 9.014a2 2 0 0 1-.853.505l-2.873.84a.5.5 0 0 1-.62-.62l.84-2.873a2 2 0 0 1 .506-.852z"}]];ve(r,ce({name:"square-pen"},()=>a,{get iconNode(){return o},children:(l,u)=>{var i=V(),d=L(i);le(d,()=>e.children??G),c(l,i)},$$slots:{default:!0}})),Y()}function ja(r,e){Q(e,!0);/**
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
 */let a=de(e,["$$slots","$$events","$$legacy"]);const o=[["circle",{cx:"12",cy:"12",r:"4"}],["path",{d:"M12 2v2"}],["path",{d:"M12 20v2"}],["path",{d:"m4.93 4.93 1.41 1.41"}],["path",{d:"m17.66 17.66 1.41 1.41"}],["path",{d:"M2 12h2"}],["path",{d:"M20 12h2"}],["path",{d:"m6.34 17.66-1.41 1.41"}],["path",{d:"m19.07 4.93-1.41 1.41"}]];ve(r,ce({name:"sun"},()=>a,{get iconNode(){return o},children:(l,u)=>{var i=V(),d=L(i);le(d,()=>e.children??G),c(l,i)},$$slots:{default:!0}})),Y()}function ft(r,e){Q(e,!0);/**
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
 */let a=de(e,["$$slots","$$events","$$legacy"]);const o=[["path",{d:"M10 11v6"}],["path",{d:"M14 11v6"}],["path",{d:"M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6"}],["path",{d:"M3 6h18"}],["path",{d:"M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"}]];ve(r,ce({name:"trash-2"},()=>a,{get iconNode(){return o},children:(l,u)=>{var i=V(),d=L(i);le(d,()=>e.children??G),c(l,i)},$$slots:{default:!0}})),Y()}function qa(r,e,a){r.key==="Enter"?e():r.key==="Escape"&&a()}var Va=m('<div class="flex items-center border-b border-base-200 p-3"><div class="flex-1"><div class="flex items-center justify-between gap-2"><div class="flex min-w-0 flex-1 items-center gap-2"><div class="h-5 w-48 skeleton"></div></div> <div class="h-4 w-8 skeleton"></div></div></div> <div class="w-8"></div></div>'),Ba=(r,e,a)=>e(t(a).id),Ka=r=>r.stopPropagation(),Ja=m('<input type="text" class="input input-sm min-w-0 flex-1"/>'),Xa=m('<h3 class="truncate text-sm font-medium"> </h3>'),Ga=m('<span class="flex-shrink-0 text-xs text-base-content/50"> </span>'),Qa=m('<div class="flex items-center gap-1 px-2"><button class="btn btn-ghost btn-xs" aria-label="Cancel editing"><!></button> <button class="btn text-success btn-ghost btn-xs hover:bg-success/20" aria-label="Save changes"><!></button></div>'),Ya=(r,e,a)=>e(t(a).id,t(a).title),er=(r,e,a)=>e(t(a).id),tr=m('<div class="dropdown dropdown-end opacity-0 transition-opacity group-hover:opacity-100"><div tabindex="0" role="button" class="btn btn-square btn-ghost btn-sm"><!></div> <ul class="dropdown-content menu z-[1] w-32 rounded-box border bg-base-100 p-2 shadow"><li><button class="text-sm"><!> Rename</button></li> <li><button class="text-sm text-error"><!> Delete</button></li></ul></div>'),ar=m('<div class="group flex items-center border-b border-base-200 hover:bg-base-100"><button class="flex-1 truncate p-3 text-left transition-colors focus:outline-none"><div class="flex items-center justify-between gap-2"><div class="flex min-w-0 flex-1 items-center gap-2"><!></div> <!></div></button> <!> <!></div>'),rr=m('<div class="flex h-full flex-col"><div class="flex-shrink-0 p-2"><h2 class="font-semibold text-base-content/60">Conversations</h2></div> <div class="flex-1 overflow-y-auto"><!></div></div>');function nr(r,e){Q(e,!0);let a=_a(e,"isLoading",3,!1),o=_e(null),l=_e("");function u(k){ya(dt(`/c/${k}`)),e.onThreadClick?.()}function i(k){const N=new Date().getTime()-new Date(k).getTime(),w=Math.floor(N/(1e3*60)),W=Math.floor(N/(1e3*60*60)),ae=Math.floor(N/(1e3*60*60*24));return w<1?"now":w<60?`${w}m`:W<24?`${W}h`:`${ae}d`}function d(k,g){C(o,k,!0),C(l,g||"",!0)}function F(){t(o)&&t(l).trim()&&(e.onRename(t(o),t(l).trim()),C(o,null),C(l,""))}function _(){C(o,null),C(l,"")}function E(k){e.onDelete(k)}var R=rr(),M=v(s(R),2),Z=s(M);{var H=k=>{var g=V(),N=L(g);Fe(N,16,()=>Array(5).fill(null),Ft,(w,W)=>{var ae=Va();c(w,ae)}),c(k,g)},x=k=>{var g=V(),N=L(g);Fe(N,17,()=>e.threads,w=>w.id,(w,W)=>{var ae=ar(),J=s(ae);J.__click=[Ba,u,W];var I=s(J),re=s(I),B=s(re);{var ge=D=>{var p=Ja();Qe(p),p.__keydown=[qa,F,_],p.__click=[Ka],ut("focus",p,b=>b.target.select()),Ye(p,()=>t(l),b=>C(l,b)),c(D,p)},we=D=>{var p=Xa(),b=s(p,!0);n(p),se(()=>be(b,t(W).title||"Untitled")),c(D,p)};$(B,D=>{t(o)===t(W).id?D(ge):D(we,!1)})}n(re);var ke=v(re,2);{var ye=D=>{var p=Ga(),b=s(p,!0);n(p),se(z=>be(b,z),[()=>i(t(W).created)]),c(D,p)};$(ke,D=>{t(o)!==t(W).id&&D(ye)})}n(I),n(J);var me=v(J,2);{var f=D=>{var p=Qa(),b=s(p);b.__click=_;var z=s(b);Le(z,{class:"h-3 w-3"}),n(b);var O=v(b,2);O.__click=F;var h=s(O);et(h,{class:"h-3 w-3"}),n(O),n(p),c(D,p)};$(me,D=>{t(o)===t(W).id&&D(f)})}var P=v(me,2);{var ee=D=>{var p=tr(),b=s(p),z=s(b);_t(z,{class:"h-4 w-4"}),n(b);var O=v(b,2),h=s(O),X=s(h);X.__click=[Ya,d,W];var j=s(X);Re(j,{class:"h-4 w-4"}),Ae(),n(X),n(h);var ue=v(h,2),fe=s(ue);fe.__click=[er,E,W];var he=s(fe);ft(he,{class:"h-4 w-4"}),Ae(),n(fe),n(ue),n(O),n(p),c(D,p)};$(P,D=>{t(o)!==t(W).id&&D(ee)})}n(ae),c(w,ae)}),c(k,g)};$(Z,k=>{a()?k(H):k(x,!1)})}n(M),n(R),c(r,R),Y()}ze(["click","keydown"]);function sr(r,e,a){r.key==="Enter"?e():r.key==="Escape"&&a()}var or=m('<div class="px-3 py-2 text-xs text-base-content/40 italic">No items</div>'),ir=(r,e)=>e.onItemClick?.(),lr=r=>r.stopPropagation(),cr=m('<input type="text" class="input input-sm min-w-0 flex-1"/>'),dr=m('<span class="badge badge-xs badge-success">Done</span>'),vr=m('<span class="truncate text-sm"> </span> <!>',1),ur=m('<span class="flex-shrink-0 text-xs text-base-content/50"> </span>'),_r=m('<div class="flex items-center gap-1 px-2"><button class="btn btn-ghost btn-xs" aria-label="Cancel editing"><!></button> <button class="btn text-success btn-ghost btn-xs hover:bg-success/20" aria-label="Save changes"><!></button></div>'),fr=(r,e,a)=>e(t(a).id,t(a).title),hr=(r,e,a)=>e(t(a).id),br=m('<div class="dropdown dropdown-end opacity-0 transition-opacity group-hover:opacity-100"><div tabindex="0" role="button" class="btn btn-square btn-ghost btn-sm"><!></div> <ul class="dropdown-content menu z-[1] w-32 rounded-box border bg-base-100 p-2 shadow"><li><button class="text-sm"><!> Rename</button></li> <li><button class="text-sm text-error"><!> Delete</button></li></ul></div>'),pr=m('<div><a><div class="flex items-center justify-between gap-2"><div class="flex min-w-0 flex-1 items-center gap-2"><!></div> <!></div></a> <!> <!></div>');function ct(r,e){Q(e,!0);const a=q(()=>Sa.params.taskId);let o=_e(null),l=_e("");function u(x){const g=new Date().getTime()-new Date(x).getTime(),N=Math.floor(g/(1e3*60)),w=Math.floor(g/(1e3*60*60)),W=Math.floor(g/(1e3*60*60*24));return N<1?"now":N<60?`${N}m`:w<24?`${w}h`:`${W}d`}function i(x,k){C(o,x,!0),C(l,k||"",!0)}function d(){t(o)&&t(l).trim()&&(e.onRename(t(o),t(l).trim()),C(o,null),C(l,""))}function F(){C(o,null),C(l,"")}function _(x){e.onDelete(x)}function E(x){return x.type==="task"?`/w/${e.workspaceId}/t/${x.id}`:"#"}var R=V(),M=L(R);{var Z=x=>{var k=or();c(x,k)},H=x=>{var k=V(),g=L(k);Fe(g,17,()=>e.items,N=>N.id,(N,w)=>{const W=q(()=>t(w).type==="task"&&t(w).id===t(a)),ae=q(()=>E(t(w)));var J=pr(),I=s(J);I.__click=[ir,e];var re=s(I),B=s(re),ge=s(B);{var we=p=>{var b=cr();Qe(b),b.__keydown=[sr,d,F],b.__click=[lr],ut("focus",b,z=>z.target.select()),Ye(b,()=>t(l),z=>C(l,z)),c(p,b)},ke=p=>{var b=vr(),z=L(b),O=s(z,!0);n(z);var h=v(z,2);{var X=j=>{var ue=dr();c(j,ue)};$(h,j=>{t(w).status==="completed"&&j(X)})}se(()=>be(O,t(w).title||"Untitled")),c(p,b)};$(ge,p=>{t(o)===t(w).id?p(we):p(ke,!1)})}n(B);var ye=v(B,2);{var me=p=>{var b=ur(),z=s(b,!0);n(b),se(O=>be(z,O),[()=>u(t(w).created)]),c(p,b)};$(ye,p=>{t(o)!==t(w).id&&p(me)})}n(re),n(I);var f=v(I,2);{var P=p=>{var b=_r(),z=s(b);z.__click=F;var O=s(z);Le(O,{class:"h-3 w-3"}),n(z);var h=v(z,2);h.__click=d;var X=s(h);et(X,{class:"h-3 w-3"}),n(h),n(b),c(p,b)};$(f,p=>{t(o)===t(w).id&&p(P)})}var ee=v(f,2);{var D=p=>{var b=br(),z=s(b),O=s(z);_t(O,{class:"h-4 w-4"}),n(z);var h=v(z,2),X=s(h),j=s(X);j.__click=[fr,i,w];var ue=s(j);Re(ue,{class:"h-4 w-4"}),Ae(),n(j),n(X);var fe=v(X,2),he=s(fe);he.__click=[hr,_,w];var $e=s(he);ft($e,{class:"h-4 w-4"}),Ae(),n(he),n(fe),n(h),n(b),c(p,b)};$(ee,p=>{t(o)!==t(w).id&&p(D)})}n(J),se(()=>{Ie(J,1,`group flex items-center border-b border-base-200 ${t(W)?"bg-primary/10":"hover:bg-base-100"}`),xe(I,"href",t(ae)),Ie(I,1,`flex-1 truncate py-2 pr-3 pl-6 text-left transition-colors focus:outline-none ${t(W)?"font-semibold":""}`)}),c(N,J)}),c(x,k)};$(M,x=>{e.items.length===0?x(Z):x(H,!1)})}c(r,R),Y()}ze(["click","keydown"]);var gr=(r,e,a)=>e(t(a)),mr=m('<div class="pl-4"><!></div>'),xr=m('<button class="flex w-full items-center gap-1 py-1 pr-2 pl-6 text-left text-sm hover:bg-base-100"><!> <!> <span class="truncate"> </span></button> <!>',1),wr=(r,e,a)=>e(t(a)),kr=m('<button class="flex w-full items-center gap-1 py-1 pr-2 pl-9 text-left text-sm hover:bg-base-100"><!> <span class="truncate"> </span></button>'),yr=m("<div><!></div>");function Dt(r,e){Q(e,!0);function a(u){u.isDirectory?e.itemStore.toggleFilePath(u.path):e.onFileClick&&e.onFileClick(u)}var o=V(),l=L(o);Fe(l,17,()=>e.nodes,u=>u.path,(u,i)=>{var d=yr(),F=s(d);{var _=R=>{const M=q(()=>e.itemStore.isFilePathExpanded(t(i).path));var Z=xr(),H=L(Z);H.__click=[gr,a,i];var x=s(H);{let I=q(()=>t(M)?"rotate-90":"");De(x,{get class(){return`h-3 w-3 flex-shrink-0 transition-transform ${t(I)??""}`}})}var k=v(x,2);{var g=I=>{Pt(I,{class:"h-3.5 w-3.5 flex-shrink-0 text-warning"})},N=I=>{Wt(I,{class:"h-3.5 w-3.5 flex-shrink-0 text-warning"})};$(k,I=>{t(M)?I(g):I(N,!1)})}var w=v(k,2),W=s(w,!0);n(w),n(H);var ae=v(H,2);{var J=I=>{var re=mr(),B=s(re);Dt(B,{get nodes(){return t(i).children},get itemStore(){return e.itemStore},get onFileClick(){return e.onFileClick}}),n(re),c(I,re)};$(ae,I=>{t(M)&&t(i).children&&I(J)})}se(()=>be(W,t(i).name)),c(R,Z)},E=R=>{var M=kr();M.__click=[wr,a,i];var Z=s(M);La(Z,{class:"h-3.5 w-3.5 flex-shrink-0"});var H=v(Z,2),x=s(H,!0);n(H),n(M),se(()=>be(x,t(i).name)),c(R,M)};$(F,R=>{t(i).isDirectory?R(_):R(E,!1)})}n(d),c(u,d)}),c(r,o),Y()}ze(["click"]);function $r(r,e,a){r.key==="Enter"?e():r.key==="Escape"&&a()}function Cr(r,e,a){C(e,!0),C(a,"")}function Nr(r,e,a){r.key==="Enter"?e():r.key==="Escape"&&a()}var Mr=m('<div class="border-b border-base-200 p-3"><div class="h-5 w-48 skeleton"></div></div>'),Sr=m('<div class="flex items-center gap-2 border-b border-base-200 bg-base-100 p-3"><input type="text" placeholder="Workspace name..." class="input input-sm flex-1"/> <button class="btn btn-ghost btn-xs" aria-label="Cancel"><!></button> <button class="btn text-success btn-ghost btn-xs hover:bg-success/20" aria-label="Create"><!></button></div>'),Ir=(r,e,a)=>e(t(a).id),Tr=(r,e,a)=>e(t(a).id),Fr=r=>r.stopPropagation(),Er=m('<input type="text" class="input input-sm min-w-0 flex-1"/>'),Pr=m('<span class="truncate text-sm font-medium"> </span>'),Wr=m('<div class="flex items-center gap-1 px-2"><button class="btn btn-ghost btn-xs" aria-label="Cancel editing"><!></button> <button class="btn text-success btn-ghost btn-xs hover:bg-success/20" aria-label="Save changes"><!></button></div>'),Dr=(r,e,a)=>e(t(a).id,t(a).name),Rr=(r,e,a)=>e(t(a).id),Ar=m('<div class="dropdown dropdown-end opacity-0 transition-opacity group-hover:opacity-100"><div tabindex="0" role="button" class="btn btn-square btn-ghost btn-sm"><!></div> <ul class="dropdown-content menu z-[1] w-32 rounded-box border bg-base-100 p-2 shadow"><li><button class="text-sm"><!> Rename</button></li> <li><button class="text-sm text-error"><!> Delete</button></li></ul></div>'),Lr=(r,e,a)=>e(t(a).id,"task"),zr=m('<span class="badge badge-xs"> </span>'),Or=(r,e,a)=>e(t(a).id,"agent"),Ur=m('<span class="badge badge-xs"> </span>'),Zr=(r,e,a)=>e(t(a).id,"conversation"),Hr=m('<span class="badge badge-xs"> </span>'),jr=(r,e,a)=>e(t(a).id,"files"),qr=m('<span class="badge badge-xs"> </span>'),Vr=m('<div class="max-h-64 overflow-y-auto"><!></div>'),Br=m('<div class="bg-base-50"><div class="border-t border-base-200/50"><button class="flex w-full items-center gap-2 py-2 pr-3 pl-4 text-left text-sm hover:bg-base-100"><!> <!> <span class="flex-1">Tasks</span> <!></button> <!></div> <div class="border-t border-base-200/50"><button class="flex w-full items-center gap-2 py-2 pr-3 pl-4 text-left text-sm hover:bg-base-100"><!> <!> <span class="flex-1">Agents</span> <!></button> <!></div> <div class="border-t border-base-200/50"><button class="flex w-full items-center gap-2 py-2 pr-3 pl-4 text-left text-sm hover:bg-base-100"><!> <!> <span class="flex-1">Conversations</span> <!></button> <!></div> <div class="border-t border-base-200/50"><button class="flex w-full items-center gap-2 py-2 pr-3 pl-4 text-left text-sm hover:bg-base-100"><!> <!> <span class="flex-1">Files</span> <!></button> <!></div></div>'),Kr=m('<div class="border-b border-base-200"><div class="group flex items-center hover:bg-base-100"><button class="btn px-2 btn-ghost btn-xs"><!></button> <button class="flex-1 truncate py-2 text-left transition-colors focus:outline-none"><div class="flex items-center gap-2"><!> <!></div></button> <!> <!></div> <!></div>'),Jr=m('<div class="p-4 text-center text-sm text-base-content/40">No workspaces yet. Click the + button to create one.</div>'),Xr=m("<!> <!> <!>",1),Gr=m('<div class="flex h-full flex-col"><div class="flex flex-shrink-0 items-center justify-between p-2"><h2 class="font-semibold text-base-content/60">Workspaces</h2> <button class="btn btn-ghost btn-xs" aria-label="Create new workspace"><!></button></div> <div class="flex-1 overflow-y-auto"><!></div></div>');function Qr(r,e){Q(e,!0);const a=()=>ha(Ma,"$page",o),[o,l]=fa(),u=q(()=>a().params.workspaceId),i=q(()=>a().params.taskId);da(()=>{if(t(u)&&t(i)){ie.isWorkspaceExpanded(t(u))||ie.toggleWorkspace(t(u));const f=ie.getItemStore(t(u));f.isSectionExpanded("task")||f.toggleSection("task")}});let d=_e(null),F=_e(""),_=_e(!1),E=_e("");function R(f){ie.toggleWorkspace(f)}function M(f,P){ie.getItemStore(f).toggleSection(P)}function Z(f,P){C(d,f,!0),C(F,P||"",!0)}async function H(){if(t(d)&&t(F).trim())try{await ie.updateWorkspace(t(d),{name:t(F).trim()}),C(d,null),C(F,"")}catch(f){console.error("Failed to rename workspace:",f)}}function x(){C(d,null),C(F,"")}async function k(f){try{await ie.deleteWorkspace(f)}catch(P){console.error("Failed to delete workspace:",P)}}async function g(){if(t(E).trim())try{await ie.createWorkspace(t(E).trim()),C(_,!1),C(E,"")}catch(f){console.error("Failed to create workspace:",f)}}function N(){C(_,!1),C(E,"")}async function w(f,P,ee){try{await ie.getItemStore(f).updateItem(P,{title:ee})}catch(D){console.error("Failed to rename item:",D)}}async function W(f,P){try{await ie.getItemStore(f).deleteItem(P)}catch(ee){console.error("Failed to delete item:",ee)}}function ae(){e.onWorkspaceClick?.()}function J(f){console.log("File clicked:",f),e.onWorkspaceClick?.()}var I=Gr(),re=s(I),B=v(s(re),2);B.__click=[Cr,_,E];var ge=s(B);$a(ge,{class:"h-4 w-4"}),n(B),n(re);var we=v(re,2),ke=s(we);{var ye=f=>{var P=V(),ee=L(P);Fe(ee,16,()=>Array(3).fill(null),Ft,(D,p)=>{var b=Mr();c(D,b)}),c(f,P)},me=f=>{var P=Xr(),ee=L(P);{var D=O=>{var h=Sr(),X=s(h);Qe(X),X.__keydown=[Nr,g,N];var j=v(X,2);j.__click=N;var ue=s(j);Le(ue,{class:"h-3 w-3"}),n(j);var fe=v(j,2);fe.__click=g;var he=s(fe);et(he,{class:"h-3 w-3"}),n(fe),n(h),Ye(X,()=>t(E),$e=>C(E,$e)),c(O,h)};$(ee,O=>{t(_)&&O(D)})}var p=v(ee,2);Fe(p,17,()=>ie.workspaces,O=>O.id,(O,h)=>{var X=Kr(),j=s(X),ue=s(j);ue.__click=[Ir,R,h];var fe=s(ue);{let U=q(()=>ie.isWorkspaceExpanded(t(h).id)?"rotate-90":"");De(fe,{get class(){return`h-3 w-3 transition-transform ${t(U)??""}`}})}n(ue);var he=v(ue,2);he.__click=[Tr,R,h];var $e=s(he),He=s($e);{var je=U=>{{let S=q(()=>t(h).color||"#888");Pt(U,{class:"h-4 w-4",get style(){return`color: ${t(S)??""}`}})}},tt=U=>{{let S=q(()=>t(h).color||"#888");Wt(U,{class:"h-4 w-4",get style(){return`color: ${t(S)??""}`}})}};$(He,U=>{ie.isWorkspaceExpanded(t(h).id)?U(je):U(tt,!1)})}var qe=v(He,2);{var at=U=>{var S=Er();Qe(S),S.__keydown=[$r,H,x],S.__click=[Fr],ut("focus",S,oe=>oe.target.select()),Ye(S,()=>t(F),oe=>C(F,oe)),c(U,S)},rt=U=>{var S=Pr(),oe=s(S,!0);n(S),se(()=>be(oe,t(h).name)),c(U,S)};$(qe,U=>{t(d)===t(h).id?U(at):U(rt,!1)})}n($e),n(he);var y=v(he,2);{var A=U=>{var S=Wr(),oe=s(S);oe.__click=x;var Ee=s(oe);Le(Ee,{class:"h-3 w-3"}),n(oe);var Ce=v(oe,2);Ce.__click=H;var Se=s(Ce);et(Se,{class:"h-3 w-3"}),n(Ce),n(S),c(U,S)};$(y,U=>{t(d)===t(h).id&&U(A)})}var te=v(y,2);{var pe=U=>{var S=Ar(),oe=s(S),Ee=s(oe);_t(Ee,{class:"h-4 w-4"}),n(oe);var Ce=v(oe,2),Se=s(Ce),Pe=s(Se);Pe.__click=[Dr,Z,h];var Ve=s(Pe);Re(Ve,{class:"h-4 w-4"}),Ae(),n(Pe),n(Se);var Oe=v(Se,2),We=s(Oe);We.__click=[Rr,k,h];var nt=s(We);ft(nt,{class:"h-4 w-4"}),Ae(),n(We),n(Oe),n(Ce),n(S),c(U,S)};$(te,U=>{t(d)!==t(h).id&&U(pe)})}n(j);var Te=v(j,2);{var Me=U=>{const S=q(()=>ie.getItemStore(t(h).id)),oe=q(()=>t(S).getItemCount("task")),Ee=q(()=>t(S).isSectionExpanded("task")),Ce=q(()=>t(S).getItemCount("agent")),Se=q(()=>t(S).isSectionExpanded("agent")),Pe=q(()=>t(S).getItemCount("conversation")),Ve=q(()=>t(S).isSectionExpanded("conversation")),Oe=q(()=>t(S).getFileCount()),We=q(()=>t(S).isSectionExpanded("files")),nt=q(()=>t(S).buildFileTree());var st=Br(),ot=s(st),Be=s(ot);Be.__click=[Lr,M,h];var ht=s(Be);{let T=q(()=>t(Ee)?"rotate-90":"");De(ht,{get class(){return`h-3 w-3 transition-transform ${t(T)??""}`}})}var bt=v(ht,2);Ca(bt,{class:"h-3.5 w-3.5"});var Rt=v(bt,4);{var At=T=>{var K=zr(),ne=s(K,!0);n(K),se(()=>be(ne,t(oe))),c(T,K)};$(Rt,T=>{t(oe)>0&&T(At)})}n(Be);var Lt=v(Be,2);{var zt=T=>{{let K=q(()=>t(S).getItems("task"));ct(T,{get items(){return t(K)},get workspaceId(){return t(h).id},onRename:(ne,Ue)=>w(t(h).id,ne,Ue),onDelete:ne=>W(t(h).id,ne),onItemClick:ae})}};$(Lt,T=>{t(Ee)&&T(zt)})}n(ot);var it=v(ot,2),Ke=s(it);Ke.__click=[Or,M,h];var pt=s(Ke);{let T=q(()=>t(Se)?"rotate-90":"");De(pt,{get class(){return`h-3 w-3 transition-transform ${t(T)??""}`}})}var gt=v(pt,2);Na(gt,{class:"h-3.5 w-3.5"});var Ot=v(gt,4);{var Ut=T=>{var K=Ur(),ne=s(K,!0);n(K),se(()=>be(ne,t(Ce))),c(T,K)};$(Ot,T=>{t(Ce)>0&&T(Ut)})}n(Ke);var Zt=v(Ke,2);{var Ht=T=>{{let K=q(()=>t(S).getItems("agent"));ct(T,{get items(){return t(K)},get workspaceId(){return t(h).id},onRename:(ne,Ue)=>w(t(h).id,ne,Ue),onDelete:ne=>W(t(h).id,ne),onItemClick:ae})}};$(Zt,T=>{t(Se)&&T(Ht)})}n(it);var lt=v(it,2),Je=s(lt);Je.__click=[Zr,M,h];var mt=s(Je);{let T=q(()=>t(Ve)?"rotate-90":"");De(mt,{get class(){return`h-3 w-3 transition-transform ${t(T)??""}`}})}var xt=v(mt,2);Ua(xt,{class:"h-3.5 w-3.5"});var jt=v(xt,4);{var qt=T=>{var K=Hr(),ne=s(K,!0);n(K),se(()=>be(ne,t(Pe))),c(T,K)};$(jt,T=>{t(Pe)>0&&T(qt)})}n(Je);var Vt=v(Je,2);{var Bt=T=>{{let K=q(()=>t(S).getItems("conversation"));ct(T,{get items(){return t(K)},get workspaceId(){return t(h).id},onRename:(ne,Ue)=>w(t(h).id,ne,Ue),onDelete:ne=>W(t(h).id,ne),onItemClick:ae})}};$(Vt,T=>{t(Ve)&&T(Bt)})}n(lt);var wt=v(lt,2),Xe=s(wt);Xe.__click=[jr,M,h];var kt=s(Xe);{let T=q(()=>t(We)?"rotate-90":"");De(kt,{get class(){return`h-3 w-3 transition-transform ${t(T)??""}`}})}var yt=v(kt,2);Aa(yt,{class:"h-3.5 w-3.5"});var Kt=v(yt,4);{var Jt=T=>{var K=qr(),ne=s(K,!0);n(K),se(()=>be(ne,t(Oe))),c(T,K)};$(Kt,T=>{t(Oe)>0&&T(Jt)})}n(Xe);var Xt=v(Xe,2);{var Gt=T=>{var K=Vr(),ne=s(K);Dt(ne,{get nodes(){return t(nt)},get itemStore(){return t(S)},onFileClick:J}),n(K),c(T,K)};$(Xt,T=>{t(We)&&T(Gt)})}n(wt),n(st),c(U,st)};$(Te,U=>{ie.isWorkspaceExpanded(t(h).id)&&U(Me)})}n(X),se(U=>xe(ue,"aria-label",U),[()=>ie.isWorkspaceExpanded(t(h).id)?"Collapse workspace":"Expand workspace"]),c(O,X)});var b=v(p,2);{var z=O=>{var h=Jr();c(O,h)};$(b,O=>{ie.workspaces.length===0&&!t(_)&&O(z)})}c(f,P)};$(ke,f=>{ie.isLoading?f(ye):f(me,!1)})}n(we),n(I),c(r,I),Y(),l()}ze(["click","keydown"]);function Yr(r){const e=r-1;return e*e*e+1}function St(r,{delay:e=0,duration:a=400,easing:o=Yr,axis:l="y"}={}){const u=getComputedStyle(r),i=+u.opacity,d=l==="y"?"height":"width",F=parseFloat(u[d]),_=l==="y"?["top","bottom"]:["left","right"],E=_.map(g=>`${g[0].toUpperCase()}${g.slice(1)}`),R=parseFloat(u[`padding${E[0]}`]),M=parseFloat(u[`padding${E[1]}`]),Z=parseFloat(u[`margin${E[0]}`]),H=parseFloat(u[`margin${E[1]}`]),x=parseFloat(u[`border${E[0]}Width`]),k=parseFloat(u[`border${E[1]}Width`]);return{delay:e,duration:a,easing:o,css:g=>`overflow: hidden;opacity: ${Math.min(g*20,1)*i};${d}: ${g*F}px;padding-${_[0]}: ${g*R}px;padding-${_[1]}: ${g*M}px;margin-${_[0]}: ${g*Z}px;margin-${_[1]}: ${g*H}px;border-${_[0]}-width: ${g*x}px;border-${_[1]}-width: ${g*k}px;min-${d}: 0`}}var en=m('<div class="mt-1 text-xs break-all opacity-80"> </div>'),tn=(r,e,a)=>e(t(a)),an=(r,e,a)=>e(t(a).id),rn=m('<div class="absolute -top-8 right-1 rounded bg-success px-2 py-1 text-xs text-success-content opacity-100 shadow-lg transition-opacity duration-500">Copied!</div>'),nn=m('<div class="mt-2 h-1 overflow-hidden rounded bg-black/10"><div class="h-full animate-pulse bg-current opacity-60"></div></div>'),sn=m('<div><div class="flex items-start gap-3"><div class="flex-shrink-0"><!></div> <div class="min-w-0 flex-1"><div class="text-sm font-medium break-all"> </div> <!></div></div> <div class="absolute top-1 right-1 flex gap-1 rounded p-1 opacity-0 backdrop-blur-sm transition-opacity group-hover:opacity-100"><button type="button" class="btn btn-ghost btn-xs" title="Copy notification"><!></button> <button class="btn btn-ghost btn-xs" aria-label="Close notification"><!></button></div> <!></div> <!>',1),on=m('<div class="fixed right-4 bottom-4 z-50 flex w-80 flex-col gap-3"></div>');function ln(r,e){Q(e,!0);let a=_e(null);const o=ba();function l(F){const _="alert shadow-lg border";switch(F){case"success":return`${_} alert-success`;case"error":return`${_} alert-error`;case"warning":return`${_} alert-warning`;case"info":return`${_} alert-info`;default:return`${_}`}}function u(F){o.remove(F)}async function i(F){const _=F.message?`${F.title}
${F.message}`:F.title;await navigator.clipboard.writeText(_),C(a,F.id,!0),setTimeout(()=>{C(a,null)},2e3)}var d=on();Fe(d,21,()=>o.notifications,F=>F.id,(F,_)=>{var E=sn(),R=L(E),M=s(R),Z=s(M),H=s(Z);{var x=f=>{Ra(f,{class:"h-5 w-5"})},k=f=>{var P=V(),ee=L(P);{var D=b=>{Da(b,{class:"h-5 w-5"})},p=b=>{var z=V(),O=L(z);{var h=j=>{ka(j,{class:"h-5 w-5"})},X=j=>{za(j,{class:"h-5 w-5"})};$(O,j=>{t(_).type==="warning"?j(h):j(X,!1)},!0)}c(b,z)};$(ee,b=>{t(_).type==="error"?b(D):b(p,!1)},!0)}c(f,P)};$(H,f=>{t(_).type==="success"?f(x):f(k,!1)})}n(Z);var g=v(Z,2),N=s(g),w=s(N,!0);n(N);var W=v(N,2);{var ae=f=>{var P=en(),ee=s(P,!0);n(P),se(()=>be(ee,t(_).message)),c(f,P)};$(W,f=>{t(_).message&&f(ae)})}n(g),n(M);var J=v(M,2),I=s(J);I.__click=[tn,i,_];var re=s(I);wa(re,{class:"h-3 w-3"}),n(I);var B=v(I,2);B.__click=[an,u,_];var ge=s(B);Le(ge,{class:"h-3 w-3"}),n(B),n(J);var we=v(J,2);{var ke=f=>{var P=rn();c(f,P)};$(we,f=>{t(a)===t(_).id&&f(ke)})}n(R);var ye=v(R,2);{var me=f=>{var P=nn(),ee=s(P);n(P),se(()=>pa(ee,`animation: shrink ${t(_).duration??""}ms linear forwards;`)),c(f,P)};$(ye,f=>{t(_).autoClose&&t(_).duration&&t(_).duration>0&&f(me)})}se(f=>{Ie(R,1,`${f??""} group relative`),be(w,t(_).title)},[()=>l(t(_).type)]),Ct(1,R,()=>St,()=>({duration:300})),Ct(2,R,()=>St,()=>({duration:200})),c(F,E)}),n(d),c(r,d),Y()}ze(["click"]);class cn{#e=_e(Tt([]));get notifications(){return t(this.#e)}set notifications(e){C(this.#e,e,!0)}add(e){const a=crypto.randomUUID(),o={...e,id:a,timestamp:new ga,autoClose:typeof e.autoClose=="boolean"?e.autoClose:e.type!=="error",duration:e.duration||(e.type==="error"?0:5e3)};return this.notifications.push(o),o.autoClose&&o.duration&&o.duration>0&&setTimeout(()=>{this.remove(a)},o.duration),a}remove(e){this.notifications=this.notifications.filter(a=>a.id!==e)}clear(){this.notifications=[]}success(e,a,o){return this.add({type:"success",title:e,message:a,duration:o})}error(e,a){return this.add({type:"error",title:e,message:a,autoClose:!1})}warning(e,a,o){return this.add({type:"warning",title:e,message:a,duration:o})}info(e,a,o){return this.add({type:"info",title:e,message:a,duration:o})}}function dn(r,e){C(e,!t(e))}function vn(r,e){C(e,t(e)==="lofi"?"black":"lofi",!0),document.documentElement.setAttribute("data-theme",t(e)),localStorage.setItem("theme",t(e))}var un=m('<link rel="icon"/>'),_n=(r,e,a)=>{window.innerWidth>=1024?e():a()},fn=m('<div class="divider my-0"></div> <div class="flex-1 overflow-hidden"><!></div>',1),hn=(r,e)=>r.key==="Enter"||r.key===" "?e():null,bn=m('<div class="fixed inset-0 z-30 bg-black/50 lg:hidden" role="button" tabindex="0"></div>'),pn=m('<div class="absolute top-0 left-0 z-10 hidden h-15 items-center bg-transparent p-2 lg:flex"><div class="flex items-center gap-2"><a class="flex items-center gap-2 text-xl font-bold hover:opacity-80"><img alt="Nanobot" class="h-12"/></a> <a class="btn p-1 btn-ghost btn-sm" aria-label="New thread"><!></a> <button class="btn p-1 btn-ghost btn-sm" aria-label="Open sidebar"><!></button></div></div>'),gn=m('<div class="absolute top-4 left-4 z-50 flex gap-2 lg:hidden"><a class="btn border border-base-300 bg-base-100/80 btn-ghost backdrop-blur-sm btn-sm" aria-label="New thread"><!></a> <button class="btn border border-base-300 bg-base-100/80 btn-ghost backdrop-blur-sm btn-sm" aria-label="Open sidebar"><!></button></div>'),mn=m('<div class="relative flex h-dvh"><div><div><div><a class="flex items-center gap-2 text-xl font-bold hover:opacity-80"><img alt="Nanobot" class="h-12"/></a> <div class="flex items-center gap-1"><a class="btn p-1 btn-ghost btn-sm" aria-label="New thread"><!></a> <button class="btn p-1 btn-ghost btn-sm"><span class="hidden lg:inline"><!></span> <span class="lg:hidden"><!></span></button></div></div> <div><div class="flex h-full flex-col"><div><!></div> <!></div></div> <div class="absolute bottom-4 left-4 z-50"><button class="btn btn-circle border-base-300 bg-base-100 shadow-lg btn-sm" aria-label="Toggle theme"><!></button></div></div></div> <!> <!> <!> <div class="h-dvh flex-1"><!></div></div> <!>',1);function Tn(r,e){Q(e,!0);let a=_e(Tt([])),o=_e(!0),l=_e(!1),u=_e(!1),i=_e("lofi"),d=_e(!1);const F=dt("/"),_=dt("/"),E=new cn;ma(E),va(async()=>{if(window.innerWidth>=1024){const A=localStorage.getItem("sidebar-collapsed");A!==null&&C(l,JSON.parse(A),!0)}{const A=localStorage.getItem("theme");if(A)C(i,A,!0);else{const te=window.matchMedia("(prefers-color-scheme: dark)").matches;C(i,te?"black":"lofi",!0)}document.documentElement.setAttribute("data-theme",t(i))}console.log("Capabilities:",await Ze.capabilities()),C(d,!!(await Ze.capabilities()).workspace?.supported);const[y]=await Promise.all([Ze.getThreads(),ie.load()]);C(a,y,!0),C(o,!1)});function R(){window.innerWidth>=1024&&(C(l,!t(l)),localStorage.setItem("sidebar-collapsed",JSON.stringify(t(l))))}function M(){C(u,!1)}async function Z(y,A){try{await Ze.renameThread(y,A);const te=t(a).findIndex(pe=>pe.id===y);te!==-1&&(t(a)[te].title=A),E.success("Thread Renamed",`Successfully renamed to "${A}"`)}catch(te){E.error("Rename Failed","Unable to rename the thread. Please try again."),console.error("Failed to rename thread:",te)}}async function H(y){try{await Ze.deleteThread(y);const A=t(a).find(te=>te.id===y);C(a,t(a).filter(te=>te.id!==y),!0),E.success("Thread Deleted",`Deleted "${A?.title||"thread"}"`)}catch(A){E.error("Delete Failed","Unable to delete the thread. Please try again."),console.error("Failed to delete thread:",A)}}var x=mn();ua(y=>{var A=un();se(()=>xe(A,"href",Wa)),c(y,A)});var k=L(x),g=s(k),N=s(g),w=s(N),W=s(w),ae=s(W);n(W);var J=v(W,2),I=s(J),re=s(I);Re(re,{class:"h-5 w-5"}),n(I);var B=v(I,2);B.__click=[_n,R,M];var ge=s(B),we=s(ge);{var ke=y=>{Mt(y,{class:"h-5 w-5"})},ye=y=>{Ha(y,{class:"h-5 w-5"})};$(we,y=>{t(l)?y(ke):y(ye,!1)})}n(ge);var me=v(ge,2),f=s(me);Le(f,{class:"h-5 w-5"}),n(me),n(B),n(J),n(w);var P=v(w,2),ee=s(P),D=s(ee),p=s(D);nr(p,{get threads(){return t(a)},onRename:Z,onDelete:H,get isLoading(){return t(o)},onThreadClick:M}),n(D);var b=v(D,2);{var z=y=>{var A=fn(),te=v(L(A),2),pe=s(te);Qr(pe,{onWorkspaceClick:M}),n(te),c(y,A)};$(b,y=>{t(d)&&y(z)})}n(ee),n(P);var O=v(P,2),h=s(O);h.__click=[vn,i];var X=s(h);{var j=y=>{Za(y,{class:"h-4 w-4"})},ue=y=>{ja(y,{class:"h-4 w-4"})};$(X,y=>{t(i)==="lofi"?y(j):y(ue,!1)})}n(h),n(O),n(N),n(g);var fe=v(g,2);{var he=y=>{var A=bn();A.__click=M,A.__keydown=[hn,M],c(y,A)};$(fe,y=>{t(u)&&y(he)})}var $e=v(fe,2);{var He=y=>{var A=pn(),te=s(A),pe=s(te),Te=s(pe);n(pe);var Me=v(pe,2),U=s(Me);Re(U,{class:"h-4 w-4"}),n(Me);var S=v(Me,2);S.__click=R;var oe=s(S);Mt(oe,{class:"h-4 w-4"}),n(S),n(te),n(A),se(()=>{xe(pe,"href",F),xe(Te,"src",Nt),xe(Me,"href",_)}),c(y,A)};$($e,y=>{t(l)&&y(He)})}var je=v($e,2);{var tt=y=>{var A=gn(),te=s(A),pe=s(te);Re(pe,{class:"h-5 w-5"}),n(te);var Te=v(te,2);Te.__click=[dn,u];var Me=s(Te);Oa(Me,{class:"h-5 w-5"}),n(Te),n(A),se(()=>xe(te,"href",_)),c(y,A)};$(je,y=>{t(u)||y(tt)})}var qe=v(je,2),at=s(qe);le(at,()=>e.children??G),n(qe),n(k);var rt=v(k,2);ln(rt,{}),se(()=>{Ie(g,1,`
		bg-base-200 transition-all duration-300 ease-in-out
		${t(l)?"hidden lg:block lg:w-0":"hidden lg:block lg:w-80"}
		${t(u)?"fixed inset-y-0 left-0 z-40 block! w-80":"lg:relative"}
	`),Ie(N,1,`flex h-full flex-col ${t(l)?"lg:overflow-hidden":""}`),Ie(w,1,`flex h-15 items-center justify-between p-2 ${t(l)?"":"min-w-80"}`),xe(W,"href",F),xe(ae,"src",Nt),xe(I,"href",_),xe(B,"aria-label",t(l)?"Open sidebar":"Close sidebar"),Ie(P,1,`flex-1 overflow-hidden ${t(l)?"":"min-w-80"}`),Ie(D,1,xa(["flex-shrink-0 overflow-y-auto",{"max-h-4/10":t(d)}]))}),c(r,x),Y()}ze(["click","keydown"]);export{Tn as component,In as universal};
