import"../chunks/DsnmJJEf.js";import{w as ea,be as ta,Q as aa,bf as ra,bg as na,e as sa,u as oa,bh as ia,bi as la,bj as ca,bk as It,ak as da,q as va,a2 as G,p as Q,d as J,f as O,an as ce,g as d,c as Y,ar as Ue,k as p,s as u,l as s,m as n,i as t,ao as he,bl as ut,a8 as k,t as oe,O as pe,at as ze,j as B,a as ua,av as fa,ad as Tt,o as _a,h as ha}from"../chunks/TV9P7kG6.js";import{s as ve,r as ue,p as ba,i as w,b as Ft,a as Et}from"../chunks/DJGld0fm.js";import{I as fe,e as Pe,i as Pt,r as Ye,b as et,X as Oe,a as Te,s as ye,g as pa,f as ga,h as ma,j as xa}from"../chunks/CnB9Cxot.js";import{a as wa,T as ka,S as ya,d as qe}from"../chunks/Czs1io1a.js";import{g as $a}from"../chunks/B35UuHvl.js";import{r as tt}from"../chunks/3m3tUzhU.js";import{w as le,P as Ca,L as Na,B as Ma,p as Sa}from"../chunks/DHeSRFc6.js";import{p as Ia}from"../chunks/BVTW_ri_.js";const Ta=()=>performance.now(),Ne={tick:r=>requestAnimationFrame(r),now:()=>Ta(),tasks:new Set};function Wt(){const r=Ne.now();Ne.tasks.forEach(e=>{e.c(r)||(Ne.tasks.delete(e),e.f())}),Ne.tasks.size!==0&&Ne.tick(Wt)}function Fa(r){let e;return Ne.tasks.size===0&&Ne.tick(Wt),{promise:new Promise(a=>{Ne.tasks.add(e={c:r,f:a})}),abort(){Ne.tasks.delete(e)}}}function Qe(r,e){It(()=>{r.dispatchEvent(new CustomEvent(e))})}function Ea(r){if(r==="float")return"cssFloat";if(r==="offset")return"cssOffset";if(r.startsWith("--"))return r;const e=r.split("-");return e.length===1?e[0]:e[0]+e.slice(1).map(a=>a[0].toUpperCase()+a.slice(1)).join("")}function $t(r){const e={},a=r.split(";");for(const i of a){const[l,f]=i.split(":");if(!l||f===void 0)break;const o=Ea(l.trim());e[o]=f.trim()}return e}const Pa=r=>r;function Ct(r,e,a,i){var l=(r&la)!==0,f=(r&ca)!==0,o=l&&f,c=(r&ia)!==0,S=o?"both":l?"in":"out",_,C=e.inert,R=e.style.overflow,y,j;function H(){return It(()=>_??=a()(e,i?.()??{},{direction:S}))}var L={is_global:c,in(){if(e.inert=C,!l){j?.abort(),j?.reset?.();return}f||y?.abort(),Qe(e,"introstart"),y=vt(e,H(),j,1,()=>{Qe(e,"introend"),y?.abort(),y=_=void 0,e.style.overflow=R})},out(I){if(!f){I?.(),_=void 0;return}e.inert=!0,Qe(e,"outrostart"),j=vt(e,H(),y,0,()=>{Qe(e,"outroend"),I?.()})},stop:()=>{y?.abort(),j?.abort()}},M=ea;if((M.transitions??=[]).push(L),l&&ta){var m=c;if(!m){for(var b=M.parent;b&&(b.f&aa)!==0;)for(;(b=b.parent)&&(b.f&ra)===0;);m=!b||(b.f&na)!==0}m&&sa(()=>{oa(()=>L.in())})}}function vt(r,e,a,i,l){var f=i===1;if(da(e)){var o,c=!1;return va(()=>{if(!c){var M=e({direction:f?"in":"out"});o=vt(r,M,a,i,l)}}),{abort:()=>{c=!0,o?.abort()},deactivate:()=>o.deactivate(),reset:()=>o.reset(),t:()=>o.t()}}if(a?.deactivate(),!e?.duration)return l(),{abort:G,deactivate:G,reset:G,t:()=>i};const{delay:S=0,css:_,tick:C,easing:R=Pa}=e;var y=[];if(f&&a===void 0&&(C&&C(0,1),_)){var j=$t(_(0,1));y.push(j,j)}var H=()=>1-i,L=r.animate(y,{duration:S,fill:"forwards"});return L.onfinish=()=>{L.cancel();var M=a?.t()??1-i;a?.abort();var m=i-M,b=e.duration*Math.abs(m),I=[];if(b>0){var P=!1;if(_)for(var q=Math.ceil(b/16.666666666666668),W=0;W<=q;W+=1){var T=M+m*R(W/q),ne=$t(_(T,1-T));I.push(ne),P||=ne.overflow==="hidden"}P&&(r.style.overflow="hidden"),H=()=>{var K=L.currentTime;return M+m*R(K/b)},C&&Fa(()=>{if(L.playState!=="running")return!1;var K=H();return C(K,1-K),!0})}L=r.animate(I,{duration:b,fill:"forwards"}),L.onfinish=()=>{H=()=>i,C?.(i,1-i),l()}},{abort:()=>{L&&(L.cancel(),L.effect=null,L.onfinish=G)},deactivate:()=>{l=G},reset:()=>{i===0&&C?.(1,0)},t:()=>H()}}const Wa=!1,Tn=Object.freeze(Object.defineProperty({__proto__:null,ssr:Wa},Symbol.toStringTag,{value:"Module"})),Da="data:image/svg+xml,%3c?xml%20version='1.0'%20encoding='UTF-8'?%3e%3csvg%20id='Layer_1'%20data-name='Layer%201'%20xmlns='http://www.w3.org/2000/svg'%20viewBox='0%200%20512%20512'%3e%3cdefs%3e%3cstyle%3e%20.cls-1%20{%20fill:%20%23fff;%20}%20%3c/style%3e%3c/defs%3e%3cpath%20class='cls-1'%20d='M348.33,145.39h-14.33v8.86c0,13.77-5.47,23.43-15.21,33.17-9.74,9.74-22.94,8.11-36.71,8.11h-55.69c-13.77,0-23.43,1.62-33.17-8.11-9.74-9.74-15.21-19.4-15.21-33.17v-8.86h-14.32c-24.89,0-44.77,19.88-44.77,44.77v122.85c0,24.88,19.89,44.77,44.77,44.77h184.65c24.88,0,44.77-19.88,44.77-44.77v-122.85c0-24.89-19.88-44.77-44.77-44.77ZM247.33,284.01c0,5.59-4.53,10.13-10.13,10.13h-39.35c-5.6,0-10.13-4.54-10.13-10.13v-40.62c0-5.59,4.53-10.13,10.13-10.13h39.35c5.6,0,10.13,4.54,10.13,10.13v40.62ZM324.29,284.01c0,5.59-4.54,10.13-10.14,10.13h-39.34c-5.59,0-10.14-4.54-10.14-10.13v-40.62c0-5.59,4.54-10.13,10.14-10.13h39.34c5.6,0,10.14,4.54,10.14,10.13v40.62Z'/%3e%3cpath%20d='M330.95,408.78h-48.98c-3.94,0-7.72,1.57-10.51,4.36-2.79,2.79-4.36,6.57-4.36,10.52v48.04c0,18.94,15.36,34.3,34.3,34.3,11.06,0,21.41-5.47,27.65-14.6,6.87-10.06,16.4-24.01,25.67-37.59,6.02-8.82,6.66-20.24,1.68-29.68-4.98-9.44-14.78-15.35-25.45-15.35Z'/%3e%3cpath%20d='M470.99,212.16c-8.36,0-15.23,6.88-15.23,15.23v18.41h-29.34v-55.65c0-42.92-35.17-78.09-78.09-78.09h-9.02l14.22-53.17c11.11-3.46,18.75-13.99,18.76-25.85,0-14.84-11.95-27.04-26.48-27.04h0c-14.53,0-26.48,12.21-26.48,27.04.03,5.8,1.88,11.44,5.28,16.09l-16.82,62.94h-103.55l-16.82-62.94c3.4-4.64,5.26-10.29,5.28-16.09,0-14.84-11.95-27.04-26.48-27.04h0c-14.53,0-26.48,12.21-26.48,27.04,0,11.86,7.65,22.39,18.76,25.85l14.22,53.17h-9.02c-42.92,0-78.09,35.17-78.09,78.09v55.65h-29.34v-18.41c0-8.36-6.88-15.23-15.23-15.23s-15.23,6.88-15.23,15.23v68.81c.12,8.28,6.96,15.02,15.23,15.02s15.11-6.74,15.23-15.02v-19.94h29.34v36.74c0,42.92,35.18,78.09,78.09,78.09h184.65c42.92,0,78.09-35.17,78.09-78.09v-36.74h29.34v19.94c.12,8.28,6.96,15.02,15.23,15.02s15.11-6.74,15.23-15.02v-68.81c0-8.36-6.88-15.23-15.23-15.23ZM393.09,313c0,24.88-19.88,44.77-44.77,44.77h-184.65c-24.89,0-44.77-19.88-44.77-44.77v-122.85c0-24.89,19.89-44.77,44.77-44.77h14.32v8.86c0,13.77,5.47,23.43,15.21,33.17,9.74,9.74,19.4,8.11,33.17,8.11h55.69c13.77,0,26.97,1.62,36.71-8.11,9.74-9.74,15.21-19.4,15.21-33.17v-8.86h14.33c24.88,0,44.77,19.88,44.77,44.77v122.85Z'/%3e%3cpath%20d='M227.28,94.04c1.8,2.37,4.61,3.77,7.59,3.77s5.79-1.4,7.59-3.77l14.21-18.74,13.63,17.98c1.8,2.37,4.61,3.77,7.59,3.77s5.79-1.4,7.59-3.77l17.23-22.74c3.15-4.16,2.33-10.18-1.83-13.34-1.66-1.26-3.68-1.94-5.76-1.94-2.98,0-5.79,1.39-7.59,3.77l-9.65,12.72-13.63-17.98c-1.8-2.37-4.61-3.76-7.59-3.76h0c-2.98,0-5.79,1.39-7.59,3.76l-14.21,18.75-10.39-13.71c-3.16-4.16-9.18-4.98-13.34-1.83-4.16,3.15-4.99,9.17-1.84,13.34l17.98,23.71Z'/%3e%3cpath%20d='M230.04,408.78h-48.98c-10.67,0-20.47,5.91-25.45,15.35-4.98,9.44-4.34,20.86,1.68,29.68,9.27,13.58,18.8,27.53,25.67,37.59,6.24,9.13,16.59,14.6,27.64,14.6h0c18.94,0,34.3-15.36,34.3-34.3v-48.04c0-3.94-1.57-7.73-4.36-10.52s-6.57-4.36-10.52-4.36Z'/%3e%3cpath%20d='M197.84,233.26h39.35c5.59,0,10.13,4.54,10.13,10.13v40.62c0,5.59-4.54,10.13-10.13,10.13h-39.35c-5.59,0-10.13-4.54-10.13-10.13v-40.62c0-5.59,4.54-10.13,10.13-10.13Z'/%3e%3cpath%20d='M314.15,233.26h-39.34c-5.59,0-10.14,4.54-10.14,10.13v40.62c0,5.59,4.54,10.13,10.14,10.13h39.34c5.6,0,10.14-4.54,10.14-10.13v-40.62c0-5.59-4.54-10.13-10.14-10.13Z'/%3e%3c/svg%3e",Nt=""+new URL("../assets/nanobot.Bn3X0Wtr.svg",import.meta.url).href;function at(r,e){Q(e,!0);/**
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
 */let a=ue(e,["$$slots","$$events","$$legacy"]);const i=[["path",{d:"M20 6 9 17l-5-5"}]];fe(r,ve({name:"check"},()=>a,{get iconNode(){return i},children:(l,f)=>{var o=J(),c=O(o);ce(c,()=>e.children??G),d(l,o)},$$slots:{default:!0}})),Y()}function Ae(r,e){Q(e,!0);/**
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
 */let a=ue(e,["$$slots","$$events","$$legacy"]);const i=[["path",{d:"m9 18 6-6-6-6"}]];fe(r,ve({name:"chevron-right"},()=>a,{get iconNode(){return i},children:(l,f)=>{var o=J(),c=O(o);ce(c,()=>e.children??G),d(l,o)},$$slots:{default:!0}})),Y()}function Ra(r,e){Q(e,!0);/**
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
 */let a=ue(e,["$$slots","$$events","$$legacy"]);const i=[["circle",{cx:"12",cy:"12",r:"10"}],["line",{x1:"12",x2:"12",y1:"8",y2:"12"}],["line",{x1:"12",x2:"12.01",y1:"16",y2:"16"}]];fe(r,ve({name:"circle-alert"},()=>a,{get iconNode(){return i},children:(l,f)=>{var o=J(),c=O(o);ce(c,()=>e.children??G),d(l,o)},$$slots:{default:!0}})),Y()}function Aa(r,e){Q(e,!0);/**
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
 */let a=ue(e,["$$slots","$$events","$$legacy"]);const i=[["path",{d:"M21.801 10A10 10 0 1 1 17 3.335"}],["path",{d:"m9 11 3 3L22 4"}]];fe(r,ve({name:"circle-check-big"},()=>a,{get iconNode(){return i},children:(l,f)=>{var o=J(),c=O(o);ce(c,()=>e.children??G),d(l,o)},$$slots:{default:!0}})),Y()}function ft(r,e){Q(e,!0);/**
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
 */let a=ue(e,["$$slots","$$events","$$legacy"]);const i=[["circle",{cx:"12",cy:"12",r:"1"}],["circle",{cx:"12",cy:"5",r:"1"}],["circle",{cx:"12",cy:"19",r:"1"}]];fe(r,ve({name:"ellipsis-vertical"},()=>a,{get iconNode(){return i},children:(l,f)=>{var o=J(),c=O(o);ce(c,()=>e.children??G),d(l,o)},$$slots:{default:!0}})),Y()}function La(r,e){Q(e,!0);/**
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
 */let a=ue(e,["$$slots","$$events","$$legacy"]);const i=[["path",{d:"M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"}],["path",{d:"M14 2v4a2 2 0 0 0 2 2h4"}],["path",{d:"M10 9H8"}],["path",{d:"M16 13H8"}],["path",{d:"M16 17H8"}]];fe(r,ve({name:"file-text"},()=>a,{get iconNode(){return i},children:(l,f)=>{var o=J(),c=O(o);ce(c,()=>e.children??G),d(l,o)},$$slots:{default:!0}})),Y()}function za(r,e){Q(e,!0);/**
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
 */let a=ue(e,["$$slots","$$events","$$legacy"]);const i=[["path",{d:"M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"}],["path",{d:"M14 2v4a2 2 0 0 0 2 2h4"}]];fe(r,ve({name:"file"},()=>a,{get iconNode(){return i},children:(l,f)=>{var o=J(),c=O(o);ce(c,()=>e.children??G),d(l,o)},$$slots:{default:!0}})),Y()}function Dt(r,e){Q(e,!0);/**
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
 */let a=ue(e,["$$slots","$$events","$$legacy"]);const i=[["path",{d:"m6 14 1.5-2.9A2 2 0 0 1 9.24 10H20a2 2 0 0 1 1.94 2.5l-1.54 6a2 2 0 0 1-1.95 1.5H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h3.9a2 2 0 0 1 1.69.9l.81 1.2a2 2 0 0 0 1.67.9H18a2 2 0 0 1 2 2v2"}]];fe(r,ve({name:"folder-open"},()=>a,{get iconNode(){return i},children:(l,f)=>{var o=J(),c=O(o);ce(c,()=>e.children??G),d(l,o)},$$slots:{default:!0}})),Y()}function Rt(r,e){Q(e,!0);/**
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
 */let a=ue(e,["$$slots","$$events","$$legacy"]);const i=[["path",{d:"M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"}]];fe(r,ve({name:"folder"},()=>a,{get iconNode(){return i},children:(l,f)=>{var o=J(),c=O(o);ce(c,()=>e.children??G),d(l,o)},$$slots:{default:!0}})),Y()}function Oa(r,e){Q(e,!0);/**
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
 */let a=ue(e,["$$slots","$$events","$$legacy"]);const i=[["circle",{cx:"12",cy:"12",r:"10"}],["path",{d:"M12 16v-4"}],["path",{d:"M12 8h.01"}]];fe(r,ve({name:"info"},()=>a,{get iconNode(){return i},children:(l,f)=>{var o=J(),c=O(o);ce(c,()=>e.children??G),d(l,o)},$$slots:{default:!0}})),Y()}function Ua(r,e){Q(e,!0);/**
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
 */let a=ue(e,["$$slots","$$events","$$legacy"]);const i=[["path",{d:"M4 12h16"}],["path",{d:"M4 18h16"}],["path",{d:"M4 6h16"}]];fe(r,ve({name:"menu"},()=>a,{get iconNode(){return i},children:(l,f)=>{var o=J(),c=O(o);ce(c,()=>e.children??G),d(l,o)},$$slots:{default:!0}})),Y()}function Za(r,e){Q(e,!0);/**
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
 */let a=ue(e,["$$slots","$$events","$$legacy"]);const i=[["path",{d:"M22 17a2 2 0 0 1-2 2H6.828a2 2 0 0 0-1.414.586l-2.202 2.202A.71.71 0 0 1 2 21.286V5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2z"}]];fe(r,ve({name:"message-square"},()=>a,{get iconNode(){return i},children:(l,f)=>{var o=J(),c=O(o);ce(c,()=>e.children??G),d(l,o)},$$slots:{default:!0}})),Y()}function Ha(r,e){Q(e,!0);/**
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
 */let a=ue(e,["$$slots","$$events","$$legacy"]);const i=[["path",{d:"M20.985 12.486a9 9 0 1 1-9.473-9.472c.405-.022.617.46.402.803a6 6 0 0 0 8.268 8.268c.344-.215.825-.004.803.401"}]];fe(r,ve({name:"moon"},()=>a,{get iconNode(){return i},children:(l,f)=>{var o=J(),c=O(o);ce(c,()=>e.children??G),d(l,o)},$$slots:{default:!0}})),Y()}function ja(r,e){Q(e,!0);/**
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
 */let a=ue(e,["$$slots","$$events","$$legacy"]);const i=[["rect",{width:"18",height:"18",x:"3",y:"3",rx:"2"}],["path",{d:"M9 3v18"}],["path",{d:"m16 15-3-3 3-3"}]];fe(r,ve({name:"panel-left-close"},()=>a,{get iconNode(){return i},children:(l,f)=>{var o=J(),c=O(o);ce(c,()=>e.children??G),d(l,o)},$$slots:{default:!0}})),Y()}function Mt(r,e){Q(e,!0);/**
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
 */let a=ue(e,["$$slots","$$events","$$legacy"]);const i=[["rect",{width:"18",height:"18",x:"3",y:"3",rx:"2"}],["path",{d:"M9 3v18"}],["path",{d:"m14 9 3 3-3 3"}]];fe(r,ve({name:"panel-left-open"},()=>a,{get iconNode(){return i},children:(l,f)=>{var o=J(),c=O(o);ce(c,()=>e.children??G),d(l,o)},$$slots:{default:!0}})),Y()}function Le(r,e){Q(e,!0);/**
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
 */let a=ue(e,["$$slots","$$events","$$legacy"]);const i=[["path",{d:"M12 3H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"}],["path",{d:"M18.375 2.625a1 1 0 0 1 3 3l-9.013 9.014a2 2 0 0 1-.853.505l-2.873.84a.5.5 0 0 1-.62-.62l.84-2.873a2 2 0 0 1 .506-.852z"}]];fe(r,ve({name:"square-pen"},()=>a,{get iconNode(){return i},children:(l,f)=>{var o=J(),c=O(o);ce(c,()=>e.children??G),d(l,o)},$$slots:{default:!0}})),Y()}function qa(r,e){Q(e,!0);/**
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
 */let a=ue(e,["$$slots","$$events","$$legacy"]);const i=[["circle",{cx:"12",cy:"12",r:"4"}],["path",{d:"M12 2v2"}],["path",{d:"M12 20v2"}],["path",{d:"m4.93 4.93 1.41 1.41"}],["path",{d:"m17.66 17.66 1.41 1.41"}],["path",{d:"M2 12h2"}],["path",{d:"M20 12h2"}],["path",{d:"m6.34 17.66-1.41 1.41"}],["path",{d:"m19.07 4.93-1.41 1.41"}]];fe(r,ve({name:"sun"},()=>a,{get iconNode(){return i},children:(l,f)=>{var o=J(),c=O(o);ce(c,()=>e.children??G),d(l,o)},$$slots:{default:!0}})),Y()}function _t(r,e){Q(e,!0);/**
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
 */let a=ue(e,["$$slots","$$events","$$legacy"]);const i=[["path",{d:"M10 11v6"}],["path",{d:"M14 11v6"}],["path",{d:"M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6"}],["path",{d:"M3 6h18"}],["path",{d:"M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"}]];fe(r,ve({name:"trash-2"},()=>a,{get iconNode(){return i},children:(l,f)=>{var o=J(),c=O(o);ce(c,()=>e.children??G),d(l,o)},$$slots:{default:!0}})),Y()}function Va(r,e,a){r.key==="Enter"?e():r.key==="Escape"&&a()}var Ba=p('<div class="flex items-center border-b border-base-200 p-3"><div class="flex-1"><div class="flex items-center justify-between gap-2"><div class="flex min-w-0 flex-1 items-center gap-2"><div class="h-5 w-48 skeleton"></div></div> <div class="h-4 w-8 skeleton"></div></div></div> <div class="w-8"></div></div>'),Ka=(r,e,a)=>e(t(a).id),Ja=r=>r.stopPropagation(),Xa=p('<input type="text" class="input input-sm min-w-0 flex-1"/>'),Ga=p('<h3 class="truncate text-sm font-medium"> </h3>'),Qa=p('<span class="flex-shrink-0 text-xs text-base-content/50"> </span>'),Ya=p('<div class="flex items-center gap-1 px-2"><button class="btn btn-ghost btn-xs" aria-label="Cancel editing"><!></button> <button class="btn text-success btn-ghost btn-xs hover:bg-success/20" aria-label="Save changes"><!></button></div>'),er=(r,e,a)=>e(t(a).id,t(a).title),tr=(r,e,a)=>e(t(a).id),ar=p('<div class="dropdown dropdown-end opacity-0 transition-opacity group-hover:opacity-100"><div tabindex="0" role="button" class="btn btn-square btn-ghost btn-sm"><!></div> <ul class="dropdown-content menu z-[1] w-32 rounded-box border bg-base-100 p-2 shadow"><li><button class="text-sm"><!> Rename</button></li> <li><button class="text-sm text-error"><!> Delete</button></li></ul></div>'),rr=p('<div class="group flex items-center border-b border-base-200 hover:bg-base-100"><button class="flex-1 truncate p-3 text-left transition-colors focus:outline-none"><div class="flex items-center justify-between gap-2"><div class="flex min-w-0 flex-1 items-center gap-2"><!></div> <!></div></button> <!> <!></div>'),nr=p('<div class="flex h-full flex-col"><div class="flex-shrink-0 p-2"><h2 class="font-semibold text-base-content/60">Conversations</h2></div> <div class="flex-1 overflow-y-auto"><!></div></div>');function sr(r,e){Q(e,!0);let a=ba(e,"isLoading",3,!1),i=he(null),l=he("");function f(M){$a(tt(`/c/${M}`)),e.onThreadClick?.()}function o(M){const b=new Date().getTime()-new Date(M).getTime(),I=Math.floor(b/(1e3*60)),P=Math.floor(b/(1e3*60*60)),q=Math.floor(b/(1e3*60*60*24));return I<1?"now":I<60?`${I}m`:P<24?`${P}h`:`${q}d`}function c(M,m){k(i,M,!0),k(l,m||"",!0)}function S(){t(i)&&t(l).trim()&&(e.onRename(t(i),t(l).trim()),k(i,null),k(l,""))}function _(){k(i,null),k(l,"")}function C(M){e.onDelete(M)}var R=nr(),y=u(s(R),2),j=s(y);{var H=M=>{var m=J(),b=O(m);Pe(b,16,()=>Array(5).fill(null),Pt,(I,P)=>{var q=Ba();d(I,q)}),d(M,m)},L=M=>{var m=J(),b=O(m);Pe(b,17,()=>e.threads,I=>I.id,(I,P)=>{var q=rr(),W=s(q);W.__click=[Ka,f,P];var T=s(W),ne=s(T),K=s(ne);{var _e=D=>{var z=Xa();Ye(z),z.__keydown=[Va,S,_],z.__click=[Ja],ut("focus",z,U=>U.target.select()),et(z,()=>t(l),U=>k(l,U)),d(D,z)},me=D=>{var z=Ga(),U=s(z,!0);n(z),oe(()=>pe(U,t(P).title||"Untitled")),d(D,z)};w(K,D=>{t(i)===t(P).id?D(_e):D(me,!1)})}n(ne);var xe=u(ne,2);{var $e=D=>{var z=Qa(),U=s(z,!0);n(z),oe(be=>pe(U,be),[()=>o(t(P).created)]),d(D,z)};w(xe,D=>{t(i)!==t(P).id&&D($e)})}n(T),n(W);var we=u(W,2);{var h=D=>{var z=Ya(),U=s(z);U.__click=_;var be=s(U);Oe(be,{class:"h-3 w-3"}),n(U);var g=u(U,2);g.__click=S;var v=s(g);at(v,{class:"h-3 w-3"}),n(g),n(z),d(D,z)};w(we,D=>{t(i)===t(P).id&&D(h)})}var F=u(we,2);{var ee=D=>{var z=ar(),U=s(z),be=s(U);ft(be,{class:"h-4 w-4"}),n(U);var g=u(U,2),v=s(g),E=s(v);E.__click=[er,c,P];var V=s(E);Le(V,{class:"h-4 w-4"}),ze(),n(E),n(v);var ae=u(v,2),se=s(ae);se.__click=[tr,C,P];var de=s(se);_t(de,{class:"h-4 w-4"}),ze(),n(se),n(ae),n(g),n(z),d(D,z)};w(F,D=>{t(i)!==t(P).id&&D(ee)})}n(q),d(I,q)}),d(M,m)};w(j,M=>{a()?M(H):M(L,!1)})}n(y),n(R),d(r,R),Y()}Ue(["click","keydown"]);function or(r,e,a){r.key==="Enter"?e():r.key==="Escape"&&a()}var ir=p('<div class="px-3 py-2 text-xs text-base-content/40 italic">No items</div>'),lr=(r,e)=>e.onItemClick?.(),cr=r=>r.stopPropagation(),dr=p('<input type="text" class="input input-sm min-w-0 flex-1"/>'),vr=p('<span class="badge badge-xs badge-success">Done</span>'),ur=p('<span class="truncate text-sm"> </span> <!>',1),fr=p('<span class="flex-shrink-0 text-xs text-base-content/50"> </span>'),_r=p('<div class="flex items-center gap-1 px-2"><button class="btn btn-ghost btn-xs" aria-label="Cancel editing"><!></button> <button class="btn text-success btn-ghost btn-xs hover:bg-success/20" aria-label="Save changes"><!></button></div>'),hr=(r,e,a)=>e(t(a).id,t(a).title),br=(r,e,a)=>e(t(a).id),pr=p('<div class="dropdown dropdown-end opacity-0 transition-opacity group-hover:opacity-100"><div tabindex="0" role="button" class="btn btn-square btn-ghost btn-sm"><!></div> <ul class="dropdown-content menu z-[1] w-32 rounded-box border bg-base-100 p-2 shadow"><li><button class="text-sm"><!> Rename</button></li> <li><button class="text-sm text-error"><!> Delete</button></li></ul></div>'),gr=p('<div><a><div class="flex items-center justify-between gap-2"><div class="flex min-w-0 flex-1 items-center gap-2"><!></div> <!></div></a> <!> <!></div>');function dt(r,e){Q(e,!0);const a=()=>Et(Ia,"$page",i),[i,l]=Ft(),f=B(()=>a().params.taskId);let o=he(null),c=he("");function S(b){const P=new Date().getTime()-new Date(b).getTime(),q=Math.floor(P/(1e3*60)),W=Math.floor(P/(1e3*60*60)),T=Math.floor(P/(1e3*60*60*24));return q<1?"now":q<60?`${q}m`:W<24?`${W}h`:`${T}d`}function _(b,I){k(o,b,!0),k(c,I||"",!0)}function C(){t(o)&&t(c).trim()&&(e.onRename(t(o),t(c).trim()),k(o,null),k(c,""))}function R(){k(o,null),k(c,"")}function y(b){e.onDelete(b)}function j(b){return b.type==="task"?`/w/${e.workspaceId}/t/${b.id}`:"#"}var H=J(),L=O(H);{var M=b=>{var I=ir();d(b,I)},m=b=>{var I=J(),P=O(I);Pe(P,17,()=>e.items,q=>q.id,(q,W)=>{const T=B(()=>t(W).type==="task"&&t(W).id===t(f)),ne=B(()=>j(t(W)));var K=gr(),_e=s(K);_e.__click=[lr,e];var me=s(_e),xe=s(me),$e=s(xe);{var we=g=>{var v=dr();Ye(v),v.__keydown=[or,C,R],v.__click=[cr],ut("focus",v,E=>E.target.select()),et(v,()=>t(c),E=>k(c,E)),d(g,v)},h=g=>{var v=ur(),E=O(v),V=s(E,!0);n(E);var ae=u(E,2);{var se=de=>{var ke=vr();d(de,ke)};w(ae,de=>{t(W).status==="completed"&&de(se)})}oe(()=>pe(V,t(W).title||"Untitled")),d(g,v)};w($e,g=>{t(o)===t(W).id?g(we):g(h,!1)})}n(xe);var F=u(xe,2);{var ee=g=>{var v=fr(),E=s(v,!0);n(v),oe(V=>pe(E,V),[()=>S(t(W).created)]),d(g,v)};w(F,g=>{t(o)!==t(W).id&&g(ee)})}n(me),n(_e);var D=u(_e,2);{var z=g=>{var v=_r(),E=s(v);E.__click=R;var V=s(E);Oe(V,{class:"h-3 w-3"}),n(E);var ae=u(E,2);ae.__click=C;var se=s(ae);at(se,{class:"h-3 w-3"}),n(ae),n(v),d(g,v)};w(D,g=>{t(o)===t(W).id&&g(z)})}var U=u(D,2);{var be=g=>{var v=pr(),E=s(v),V=s(E);ft(V,{class:"h-4 w-4"}),n(E);var ae=u(E,2),se=s(ae),de=s(se);de.__click=[hr,_,W];var ke=s(de);Le(ke,{class:"h-4 w-4"}),ze(),n(de),n(se);var Fe=u(se,2),Me=s(Fe);Me.__click=[br,y,W];var Ze=s(Me);_t(Ze,{class:"h-4 w-4"}),ze(),n(Me),n(Fe),n(ae),n(v),d(g,v)};w(U,g=>{t(o)!==t(W).id&&g(be)})}n(K),oe(g=>{Te(K,1,`group flex items-center border-b border-base-200 ${t(T)?"bg-primary/10":"hover:bg-base-100"}`),ye(_e,"href",g),Te(_e,1,`flex-1 truncate py-2 pr-3 pl-6 text-left transition-colors focus:outline-none ${t(T)?"font-semibold":""}`)},[()=>tt(t(ne))]),d(q,K)}),d(b,I)};w(L,b=>{e.items.length===0?b(M):b(m,!1)})}d(r,H),Y(),l()}Ue(["click","keydown"]);var mr=(r,e,a)=>e(t(a)),xr=p('<div class="pl-4"><!></div>'),wr=p('<button class="flex w-full items-center gap-1 py-1 pr-2 pl-6 text-left text-sm hover:bg-base-100"><!> <!> <span class="truncate"> </span></button> <!>',1),kr=(r,e,a)=>e(t(a)),yr=p('<button class="flex w-full items-center gap-1 py-1 pr-2 pl-9 text-left text-sm hover:bg-base-100"><!> <span class="truncate"> </span></button>'),$r=p("<div><!></div>");function At(r,e){Q(e,!0);function a(f){f.isDirectory?e.itemStore.toggleFilePath(f.path):e.onFileClick&&e.onFileClick(f)}var i=J(),l=O(i);Pe(l,17,()=>e.nodes,f=>f.path,(f,o)=>{var c=$r(),S=s(c);{var _=R=>{const y=B(()=>e.itemStore.isFilePathExpanded(t(o).path));var j=wr(),H=O(j);H.__click=[mr,a,o];var L=s(H);{let T=B(()=>t(y)?"rotate-90":"");Ae(L,{get class(){return`h-3 w-3 flex-shrink-0 transition-transform ${t(T)??""}`}})}var M=u(L,2);{var m=T=>{Dt(T,{class:"h-3.5 w-3.5 flex-shrink-0 text-warning"})},b=T=>{Rt(T,{class:"h-3.5 w-3.5 flex-shrink-0 text-warning"})};w(M,T=>{t(y)?T(m):T(b,!1)})}var I=u(M,2),P=s(I,!0);n(I),n(H);var q=u(H,2);{var W=T=>{var ne=xr(),K=s(ne);At(K,{get nodes(){return t(o).children},get itemStore(){return e.itemStore},get onFileClick(){return e.onFileClick}}),n(ne),d(T,ne)};w(q,T=>{t(y)&&t(o).children&&T(W)})}oe(()=>pe(P,t(o).name)),d(R,j)},C=R=>{var y=yr();y.__click=[kr,a,o];var j=s(y);za(j,{class:"h-3.5 w-3.5 flex-shrink-0"});var H=u(j,2),L=s(H,!0);n(H),n(y),oe(()=>pe(L,t(o).name)),d(R,y)};w(S,R=>{t(o).isDirectory?R(_):R(C,!1)})}n(c),d(f,c)}),d(r,i),Y()}Ue(["click"]);function Cr(r,e,a){r.key==="Enter"?e():r.key==="Escape"&&a()}function Nr(r,e,a){k(e,!0),k(a,"")}function Mr(r,e,a){r.key==="Enter"?e():r.key==="Escape"&&a()}var Sr=p('<div class="border-b border-base-200 p-3"><div class="h-5 w-48 skeleton"></div></div>'),Ir=p('<div class="flex items-center gap-2 border-b border-base-200 bg-base-100 p-3"><input type="text" placeholder="Workspace name..." class="input input-sm flex-1"/> <button class="btn btn-ghost btn-xs" aria-label="Cancel"><!></button> <button class="btn text-success btn-ghost btn-xs hover:bg-success/20" aria-label="Create"><!></button></div>'),Tr=(r,e,a)=>e(t(a).id),Fr=(r,e,a)=>e(t(a).id),Er=r=>r.stopPropagation(),Pr=p('<input type="text" class="input input-sm min-w-0 flex-1"/>'),Wr=p('<span class="truncate text-sm font-medium"> </span>'),Dr=p('<div class="flex items-center gap-1 px-2"><button class="btn btn-ghost btn-xs" aria-label="Cancel editing"><!></button> <button class="btn text-success btn-ghost btn-xs hover:bg-success/20" aria-label="Save changes"><!></button></div>'),Rr=(r,e,a)=>e(t(a).id,t(a).name),Ar=(r,e,a)=>e(t(a).id),Lr=p('<div class="dropdown dropdown-end opacity-0 transition-opacity group-hover:opacity-100"><div tabindex="0" role="button" class="btn btn-square btn-ghost btn-sm"><!></div> <ul class="dropdown-content menu z-[1] w-32 rounded-box border bg-base-100 p-2 shadow"><li><button class="text-sm"><!> Rename</button></li> <li><button class="text-sm text-error"><!> Delete</button></li></ul></div>'),zr=(r,e,a)=>e(t(a).id,"task"),Or=p('<span class="badge badge-xs"> </span>'),Ur=(r,e,a)=>e(t(a).id,"agent"),Zr=p('<span class="badge badge-xs"> </span>'),Hr=(r,e,a)=>e(t(a).id,"conversation"),jr=p('<span class="badge badge-xs"> </span>'),qr=(r,e,a)=>e(t(a).id,"files"),Vr=p('<span class="badge badge-xs"> </span>'),Br=p('<div class="max-h-64 overflow-y-auto"><!></div>'),Kr=p('<div class="bg-base-50"><div class="border-t border-base-200/50"><button class="flex w-full items-center gap-2 py-2 pr-3 pl-4 text-left text-sm hover:bg-base-100"><!> <!> <span class="flex-1">Tasks</span> <!></button> <!></div> <div class="border-t border-base-200/50"><button class="flex w-full items-center gap-2 py-2 pr-3 pl-4 text-left text-sm hover:bg-base-100"><!> <!> <span class="flex-1">Agents</span> <!></button> <!></div> <div class="border-t border-base-200/50"><button class="flex w-full items-center gap-2 py-2 pr-3 pl-4 text-left text-sm hover:bg-base-100"><!> <!> <span class="flex-1">Conversations</span> <!></button> <!></div> <div class="border-t border-base-200/50"><button class="flex w-full items-center gap-2 py-2 pr-3 pl-4 text-left text-sm hover:bg-base-100"><!> <!> <span class="flex-1">Files</span> <!></button> <!></div></div>'),Jr=p('<div class="border-b border-base-200"><div class="group flex items-center hover:bg-base-100"><button class="btn px-2 btn-ghost btn-xs"><!></button> <button class="flex-1 truncate py-2 text-left transition-colors focus:outline-none"><div class="flex items-center gap-2"><!> <!></div></button> <!> <!></div> <!></div>'),Xr=p('<div class="p-4 text-center text-sm text-base-content/40">No workspaces yet. Click the + button to create one.</div>'),Gr=p("<!> <!> <!>",1),Qr=p('<div class="flex h-full flex-col"><div class="flex flex-shrink-0 items-center justify-between p-2"><h2 class="font-semibold text-base-content/60">Workspaces</h2> <button class="btn btn-ghost btn-xs" aria-label="Create new workspace"><!></button></div> <div class="flex-1 overflow-y-auto"><!></div></div>');function Yr(r,e){Q(e,!0);const a=()=>Et(Sa,"$page",i),[i,l]=Ft(),f=B(()=>a().params.workspaceId),o=B(()=>a().params.taskId);ua(()=>{if(t(f)&&t(o)){le.isWorkspaceExpanded(t(f))||le.toggleWorkspace(t(f));const h=le.getItemStore(t(f));h.isSectionExpanded("task")||h.toggleSection("task")}});let c=he(null),S=he(""),_=he(!1),C=he("");function R(h){le.toggleWorkspace(h)}function y(h,F){le.getItemStore(h).toggleSection(F)}function j(h,F){k(c,h,!0),k(S,F||"",!0)}async function H(){if(t(c)&&t(S).trim())try{await le.updateWorkspace(t(c),{name:t(S).trim()}),k(c,null),k(S,"")}catch(h){console.error("Failed to rename workspace:",h)}}function L(){k(c,null),k(S,"")}async function M(h){try{await le.deleteWorkspace(h)}catch(F){console.error("Failed to delete workspace:",F)}}async function m(){if(t(C).trim())try{await le.createWorkspace(t(C).trim()),k(_,!1),k(C,"")}catch(h){console.error("Failed to create workspace:",h)}}function b(){k(_,!1),k(C,"")}async function I(h,F,ee){try{await le.getItemStore(h).updateItem(F,{title:ee})}catch(D){console.error("Failed to rename item:",D)}}async function P(h,F){try{await le.getItemStore(h).deleteItem(F)}catch(ee){console.error("Failed to delete item:",ee)}}function q(){e.onWorkspaceClick?.()}function W(h){console.log("File clicked:",h),e.onWorkspaceClick?.()}var T=Qr(),ne=s(T),K=u(s(ne),2);K.__click=[Nr,_,C];var _e=s(K);Ca(_e,{class:"h-4 w-4"}),n(K),n(ne);var me=u(ne,2),xe=s(me);{var $e=h=>{var F=J(),ee=O(F);Pe(ee,16,()=>Array(3).fill(null),Pt,(D,z)=>{var U=Sr();d(D,U)}),d(h,F)},we=h=>{var F=Gr(),ee=O(F);{var D=g=>{var v=Ir(),E=s(v);Ye(E),E.__keydown=[Mr,m,b],fa(E,!0);var V=u(E,2);V.__click=b;var ae=s(V);Oe(ae,{class:"h-3 w-3"}),n(V);var se=u(V,2);se.__click=m;var de=s(se);at(de,{class:"h-3 w-3"}),n(se),n(v),et(E,()=>t(C),ke=>k(C,ke)),d(g,v)};w(ee,g=>{t(_)&&g(D)})}var z=u(ee,2);Pe(z,17,()=>le.workspaces,g=>g.id,(g,v)=>{var E=Jr(),V=s(E),ae=s(V);ae.__click=[Tr,R,v];var se=s(ae);{let Z=B(()=>le.isWorkspaceExpanded(t(v).id)?"rotate-90":"");Ae(se,{get class(){return`h-3 w-3 transition-transform ${t(Z)??""}`}})}n(ae);var de=u(ae,2);de.__click=[Fr,R,v];var ke=s(de),Fe=s(ke);{var Me=Z=>{{let $=B(()=>t(v).color||"#888");Dt(Z,{class:"h-4 w-4",get style(){return`color: ${t($)??""}`}})}},Ze=Z=>{{let $=B(()=>t(v).color||"#888");Rt(Z,{class:"h-4 w-4",get style(){return`color: ${t($)??""}`}})}};w(Fe,Z=>{le.isWorkspaceExpanded(t(v).id)?Z(Me):Z(Ze,!1)})}var Ve=u(Fe,2);{var rt=Z=>{var $=Pr();Ye($),$.__keydown=[Cr,H,L],$.__click=[Er],ut("focus",$,ie=>ie.target.select()),et($,()=>t(S),ie=>k(S,ie)),d(Z,$)},nt=Z=>{var $=Wr(),ie=s($,!0);n($),oe(()=>pe(ie,t(v).name)),d(Z,$)};w(Ve,Z=>{t(c)===t(v).id?Z(rt):Z(nt,!1)})}n(ke),n(de);var x=u(de,2);{var A=Z=>{var $=Dr(),ie=s($);ie.__click=L;var We=s(ie);Oe(We,{class:"h-3 w-3"}),n(ie);var Ce=u(ie,2);Ce.__click=H;var Ie=s(Ce);at(Ie,{class:"h-3 w-3"}),n(Ce),n($),d(Z,$)};w(x,Z=>{t(c)===t(v).id&&Z(A)})}var te=u(x,2);{var ge=Z=>{var $=Lr(),ie=s($),We=s(ie);ft(We,{class:"h-4 w-4"}),n(ie);var Ce=u(ie,2),Ie=s(Ce),De=s(Ie);De.__click=[Rr,j,v];var Be=s(De);Le(Be,{class:"h-4 w-4"}),ze(),n(De),n(Ie);var He=u(Ie,2),Re=s(He);Re.__click=[Ar,M,v];var st=s(Re);_t(st,{class:"h-4 w-4"}),ze(),n(Re),n(He),n(Ce),n($),d(Z,$)};w(te,Z=>{t(c)!==t(v).id&&Z(ge)})}n(V);var Ee=u(V,2);{var Se=Z=>{const $=B(()=>le.getItemStore(t(v).id)),ie=B(()=>t($).getItemCount("task")),We=B(()=>t($).isSectionExpanded("task")),Ce=B(()=>t($).getItemCount("agent")),Ie=B(()=>t($).isSectionExpanded("agent")),De=B(()=>t($).getItemCount("conversation")),Be=B(()=>t($).isSectionExpanded("conversation")),He=B(()=>t($).getFileCount()),Re=B(()=>t($).isSectionExpanded("files")),st=B(()=>t($).buildFileTree());var ot=Kr(),it=s(ot),Ke=s(it);Ke.__click=[zr,y,v];var ht=s(Ke);{let N=B(()=>t(We)?"rotate-90":"");Ae(ht,{get class(){return`h-3 w-3 transition-transform ${t(N)??""}`}})}var bt=u(ht,2);Na(bt,{class:"h-3.5 w-3.5"});var Lt=u(bt,4);{var zt=N=>{var X=Or(),re=s(X,!0);n(X),oe(()=>pe(re,t(ie))),d(N,X)};w(Lt,N=>{t(ie)>0&&N(zt)})}n(Ke);var Ot=u(Ke,2);{var Ut=N=>{{let X=B(()=>t($).getItems("task"));dt(N,{get items(){return t(X)},get workspaceId(){return t(v).id},onRename:(re,je)=>I(t(v).id,re,je),onDelete:re=>P(t(v).id,re),onItemClick:q})}};w(Ot,N=>{t(We)&&N(Ut)})}n(it);var lt=u(it,2),Je=s(lt);Je.__click=[Ur,y,v];var pt=s(Je);{let N=B(()=>t(Ie)?"rotate-90":"");Ae(pt,{get class(){return`h-3 w-3 transition-transform ${t(N)??""}`}})}var gt=u(pt,2);Ma(gt,{class:"h-3.5 w-3.5"});var Zt=u(gt,4);{var Ht=N=>{var X=Zr(),re=s(X,!0);n(X),oe(()=>pe(re,t(Ce))),d(N,X)};w(Zt,N=>{t(Ce)>0&&N(Ht)})}n(Je);var jt=u(Je,2);{var qt=N=>{{let X=B(()=>t($).getItems("agent"));dt(N,{get items(){return t(X)},get workspaceId(){return t(v).id},onRename:(re,je)=>I(t(v).id,re,je),onDelete:re=>P(t(v).id,re),onItemClick:q})}};w(jt,N=>{t(Ie)&&N(qt)})}n(lt);var ct=u(lt,2),Xe=s(ct);Xe.__click=[Hr,y,v];var mt=s(Xe);{let N=B(()=>t(Be)?"rotate-90":"");Ae(mt,{get class(){return`h-3 w-3 transition-transform ${t(N)??""}`}})}var xt=u(mt,2);Za(xt,{class:"h-3.5 w-3.5"});var Vt=u(xt,4);{var Bt=N=>{var X=jr(),re=s(X,!0);n(X),oe(()=>pe(re,t(De))),d(N,X)};w(Vt,N=>{t(De)>0&&N(Bt)})}n(Xe);var Kt=u(Xe,2);{var Jt=N=>{{let X=B(()=>t($).getItems("conversation"));dt(N,{get items(){return t(X)},get workspaceId(){return t(v).id},onRename:(re,je)=>I(t(v).id,re,je),onDelete:re=>P(t(v).id,re),onItemClick:q})}};w(Kt,N=>{t(Be)&&N(Jt)})}n(ct);var wt=u(ct,2),Ge=s(wt);Ge.__click=[qr,y,v];var kt=s(Ge);{let N=B(()=>t(Re)?"rotate-90":"");Ae(kt,{get class(){return`h-3 w-3 transition-transform ${t(N)??""}`}})}var yt=u(kt,2);La(yt,{class:"h-3.5 w-3.5"});var Xt=u(yt,4);{var Gt=N=>{var X=Vr(),re=s(X,!0);n(X),oe(()=>pe(re,t(He))),d(N,X)};w(Xt,N=>{t(He)>0&&N(Gt)})}n(Ge);var Qt=u(Ge,2);{var Yt=N=>{var X=Br(),re=s(X);At(re,{get nodes(){return t(st)},get itemStore(){return t($)},onFileClick:W}),n(X),d(N,X)};w(Qt,N=>{t(Re)&&N(Yt)})}n(wt),n(ot),d(Z,ot)};w(Ee,Z=>{le.isWorkspaceExpanded(t(v).id)&&Z(Se)})}n(E),oe(Z=>ye(ae,"aria-label",Z),[()=>le.isWorkspaceExpanded(t(v).id)?"Collapse workspace":"Expand workspace"]),d(g,E)});var U=u(z,2);{var be=g=>{var v=Xr();d(g,v)};w(U,g=>{le.workspaces.length===0&&!t(_)&&g(be)})}d(h,F)};w(xe,h=>{le.isLoading?h($e):h(we,!1)})}n(me),n(T),d(r,T),Y(),l()}Ue(["click","keydown"]);function en(r){const e=r-1;return e*e*e+1}function St(r,{delay:e=0,duration:a=400,easing:i=en,axis:l="y"}={}){const f=getComputedStyle(r),o=+f.opacity,c=l==="y"?"height":"width",S=parseFloat(f[c]),_=l==="y"?["top","bottom"]:["left","right"],C=_.map(m=>`${m[0].toUpperCase()}${m.slice(1)}`),R=parseFloat(f[`padding${C[0]}`]),y=parseFloat(f[`padding${C[1]}`]),j=parseFloat(f[`margin${C[0]}`]),H=parseFloat(f[`margin${C[1]}`]),L=parseFloat(f[`border${C[0]}Width`]),M=parseFloat(f[`border${C[1]}Width`]);return{delay:e,duration:a,easing:i,css:m=>`overflow: hidden;opacity: ${Math.min(m*20,1)*o};${c}: ${m*S}px;padding-${_[0]}: ${m*R}px;padding-${_[1]}: ${m*y}px;margin-${_[0]}: ${m*j}px;margin-${_[1]}: ${m*H}px;border-${_[0]}-width: ${m*L}px;border-${_[1]}-width: ${m*M}px;min-${c}: 0`}}var tn=p('<div class="mt-1 text-xs break-all opacity-80"> </div>'),an=(r,e,a)=>e(t(a)),rn=(r,e,a)=>e(t(a).id),nn=p('<div class="absolute -top-8 right-1 rounded bg-success px-2 py-1 text-xs text-success-content opacity-100 shadow-lg transition-opacity duration-500">Copied!</div>'),sn=p('<div class="mt-2 h-1 overflow-hidden rounded bg-black/10"><div class="h-full animate-pulse bg-current opacity-60"></div></div>'),on=p('<div><div class="flex items-start gap-3"><div class="flex-shrink-0"><!></div> <div class="min-w-0 flex-1"><div class="text-sm font-medium break-all"> </div> <!></div></div> <div class="absolute top-1 right-1 flex gap-1 rounded p-1 opacity-0 backdrop-blur-sm transition-opacity group-hover:opacity-100"><button type="button" class="btn btn-ghost btn-xs" title="Copy notification"><!></button> <button class="btn btn-ghost btn-xs" aria-label="Close notification"><!></button></div> <!></div> <!>',1),ln=p('<div class="fixed right-4 bottom-4 z-50 flex w-80 flex-col gap-3"></div>');function cn(r,e){Q(e,!0);let a=he(null);const i=pa();function l(S){const _="alert shadow-lg border";switch(S){case"success":return`${_} alert-success`;case"error":return`${_} alert-error`;case"warning":return`${_} alert-warning`;case"info":return`${_} alert-info`;default:return`${_}`}}function f(S){i.remove(S)}async function o(S){const _=S.message?`${S.title}
${S.message}`:S.title;await navigator.clipboard.writeText(_),k(a,S.id,!0),setTimeout(()=>{k(a,null)},2e3)}var c=ln();Pe(c,21,()=>i.notifications,S=>S.id,(S,_)=>{var C=on(),R=O(C),y=s(R),j=s(y),H=s(j);{var L=h=>{Aa(h,{class:"h-5 w-5"})},M=h=>{var F=J(),ee=O(F);{var D=U=>{Ra(U,{class:"h-5 w-5"})},z=U=>{var be=J(),g=O(be);{var v=V=>{ka(V,{class:"h-5 w-5"})},E=V=>{Oa(V,{class:"h-5 w-5"})};w(g,V=>{t(_).type==="warning"?V(v):V(E,!1)},!0)}d(U,be)};w(ee,U=>{t(_).type==="error"?U(D):U(z,!1)},!0)}d(h,F)};w(H,h=>{t(_).type==="success"?h(L):h(M,!1)})}n(j);var m=u(j,2),b=s(m),I=s(b,!0);n(b);var P=u(b,2);{var q=h=>{var F=tn(),ee=s(F,!0);n(F),oe(()=>pe(ee,t(_).message)),d(h,F)};w(P,h=>{t(_).message&&h(q)})}n(m),n(y);var W=u(y,2),T=s(W);T.__click=[an,o,_];var ne=s(T);wa(ne,{class:"h-3 w-3"}),n(T);var K=u(T,2);K.__click=[rn,f,_];var _e=s(K);Oe(_e,{class:"h-3 w-3"}),n(K),n(W);var me=u(W,2);{var xe=h=>{var F=nn();d(h,F)};w(me,h=>{t(a)===t(_).id&&h(xe)})}n(R);var $e=u(R,2);{var we=h=>{var F=sn(),ee=s(F);n(F),oe(()=>ga(ee,`animation: shrink ${t(_).duration??""}ms linear forwards;`)),d(h,F)};w($e,h=>{t(_).autoClose&&t(_).duration&&t(_).duration>0&&h(we)})}oe(h=>{Te(R,1,`${h??""} group relative`),pe(I,t(_).title)},[()=>l(t(_).type)]),Ct(1,R,()=>St,()=>({duration:300})),Ct(2,R,()=>St,()=>({duration:200})),d(S,C)}),n(c),d(r,c),Y()}Ue(["click"]);class dn{#e=he(Tt([]));get notifications(){return t(this.#e)}set notifications(e){k(this.#e,e,!0)}add(e){const a=crypto.randomUUID(),i={...e,id:a,timestamp:new ya,autoClose:typeof e.autoClose=="boolean"?e.autoClose:e.type!=="error",duration:e.duration||(e.type==="error"?0:5e3)};return this.notifications.push(i),i.autoClose&&i.duration&&i.duration>0&&setTimeout(()=>{this.remove(a)},i.duration),a}remove(e){this.notifications=this.notifications.filter(a=>a.id!==e)}clear(){this.notifications=[]}success(e,a,i){return this.add({type:"success",title:e,message:a,duration:i})}error(e,a){return this.add({type:"error",title:e,message:a,autoClose:!1})}warning(e,a,i){return this.add({type:"warning",title:e,message:a,duration:i})}info(e,a,i){return this.add({type:"info",title:e,message:a,duration:i})}}function vn(r,e){k(e,!t(e))}function un(r,e){k(e,t(e)==="lofi"?"black":"lofi",!0),document.documentElement.setAttribute("data-theme",t(e)),localStorage.setItem("theme",t(e))}var fn=p('<link rel="icon"/>'),_n=(r,e,a)=>{window.innerWidth>=1024?e():a()},hn=p('<div class="divider my-0"></div> <div class="flex-1 overflow-hidden"><!></div>',1),bn=(r,e)=>r.key==="Enter"||r.key===" "?e():null,pn=p('<div class="fixed inset-0 z-30 bg-black/50 lg:hidden" role="button" tabindex="0"></div>'),gn=p('<div class="absolute top-0 left-0 z-10 hidden h-15 items-center bg-transparent p-2 lg:flex"><div class="flex items-center gap-2"><a class="flex items-center gap-2 text-xl font-bold hover:opacity-80"><img alt="Nanobot" class="h-12"/></a> <a class="btn p-1 btn-ghost btn-sm" aria-label="New thread"><!></a> <button class="btn p-1 btn-ghost btn-sm" aria-label="Open sidebar"><!></button></div></div>'),mn=p('<div class="absolute top-4 left-4 z-50 flex gap-2 lg:hidden"><a class="btn border border-base-300 bg-base-100/80 btn-ghost backdrop-blur-sm btn-sm" aria-label="New thread"><!></a> <button class="btn border border-base-300 bg-base-100/80 btn-ghost backdrop-blur-sm btn-sm" aria-label="Open sidebar"><!></button></div>'),xn=p('<div class="relative flex h-dvh"><div><div><div><a class="flex items-center gap-2 text-xl font-bold hover:opacity-80"><img alt="Nanobot" class="h-12"/></a> <div class="flex items-center gap-1"><a class="btn p-1 btn-ghost btn-sm" aria-label="New thread"><!></a> <button class="btn p-1 btn-ghost btn-sm"><span class="hidden lg:inline"><!></span> <span class="lg:hidden"><!></span></button></div></div> <div><div class="flex h-full flex-col"><div><!></div> <!></div></div> <div class="absolute bottom-4 left-4 z-50"><button class="btn btn-circle border-base-300 bg-base-100 shadow-lg btn-sm" aria-label="Toggle theme"><!></button></div></div></div> <!> <!> <!> <div class="h-dvh flex-1"><!></div></div> <!>',1);function Fn(r,e){Q(e,!0);let a=he(Tt([])),i=he(!0),l=he(!1),f=he(!1),o=he("lofi"),c=he(!1);const S=tt("/"),_=tt("/"),C=new dn;ma(C),_a(async()=>{if(window.innerWidth>=1024){const A=localStorage.getItem("sidebar-collapsed");A!==null&&k(l,JSON.parse(A),!0)}{const A=localStorage.getItem("theme");if(A)k(o,A,!0);else{const te=window.matchMedia("(prefers-color-scheme: dark)").matches;k(o,te?"black":"lofi",!0)}document.documentElement.setAttribute("data-theme",t(o))}console.log("Capabilities:",await qe.capabilities()),k(c,!!(await qe.capabilities()).workspace?.supported);const[x]=await Promise.all([qe.getThreads(),le.load()]);k(a,x,!0),k(i,!1)});function R(){window.innerWidth>=1024&&(k(l,!t(l)),localStorage.setItem("sidebar-collapsed",JSON.stringify(t(l))))}function y(){k(f,!1)}async function j(x,A){try{await qe.renameThread(x,A);const te=t(a).findIndex(ge=>ge.id===x);te!==-1&&(t(a)[te].title=A),C.success("Thread Renamed",`Successfully renamed to "${A}"`)}catch(te){C.error("Rename Failed","Unable to rename the thread. Please try again."),console.error("Failed to rename thread:",te)}}async function H(x){try{await qe.deleteThread(x);const A=t(a).find(te=>te.id===x);k(a,t(a).filter(te=>te.id!==x),!0),C.success("Thread Deleted",`Deleted "${A?.title||"thread"}"`)}catch(A){C.error("Delete Failed","Unable to delete the thread. Please try again."),console.error("Failed to delete thread:",A)}}var L=xn();ha(x=>{var A=fn();oe(()=>ye(A,"href",Da)),d(x,A)});var M=O(L),m=s(M),b=s(m),I=s(b),P=s(I),q=s(P);n(P);var W=u(P,2),T=s(W),ne=s(T);Le(ne,{class:"h-5 w-5"}),n(T);var K=u(T,2);K.__click=[_n,R,y];var _e=s(K),me=s(_e);{var xe=x=>{Mt(x,{class:"h-5 w-5"})},$e=x=>{ja(x,{class:"h-5 w-5"})};w(me,x=>{t(l)?x(xe):x($e,!1)})}n(_e);var we=u(_e,2),h=s(we);Oe(h,{class:"h-5 w-5"}),n(we),n(K),n(W),n(I);var F=u(I,2),ee=s(F),D=s(ee),z=s(D);sr(z,{get threads(){return t(a)},onRename:j,onDelete:H,get isLoading(){return t(i)},onThreadClick:y}),n(D);var U=u(D,2);{var be=x=>{var A=hn(),te=u(O(A),2),ge=s(te);Yr(ge,{onWorkspaceClick:y}),n(te),d(x,A)};w(U,x=>{t(c)&&x(be)})}n(ee),n(F);var g=u(F,2),v=s(g);v.__click=[un,o];var E=s(v);{var V=x=>{Ha(x,{class:"h-4 w-4"})},ae=x=>{qa(x,{class:"h-4 w-4"})};w(E,x=>{t(o)==="lofi"?x(V):x(ae,!1)})}n(v),n(g),n(b),n(m);var se=u(m,2);{var de=x=>{var A=pn();A.__click=y,A.__keydown=[bn,y],d(x,A)};w(se,x=>{t(f)&&x(de)})}var ke=u(se,2);{var Fe=x=>{var A=gn(),te=s(A),ge=s(te),Ee=s(ge);n(ge);var Se=u(ge,2),Z=s(Se);Le(Z,{class:"h-4 w-4"}),n(Se);var $=u(Se,2);$.__click=R;var ie=s($);Mt(ie,{class:"h-4 w-4"}),n($),n(te),n(A),oe(()=>{ye(ge,"href",S),ye(Ee,"src",Nt),ye(Se,"href",_)}),d(x,A)};w(ke,x=>{t(l)&&x(Fe)})}var Me=u(ke,2);{var Ze=x=>{var A=mn(),te=s(A),ge=s(te);Le(ge,{class:"h-5 w-5"}),n(te);var Ee=u(te,2);Ee.__click=[vn,f];var Se=s(Ee);Ua(Se,{class:"h-5 w-5"}),n(Ee),n(A),oe(()=>ye(te,"href",_)),d(x,A)};w(Me,x=>{t(f)||x(Ze)})}var Ve=u(Me,2),rt=s(Ve);ce(rt,()=>e.children??G),n(Ve),n(M);var nt=u(M,2);cn(nt,{}),oe(()=>{Te(m,1,`
		bg-base-200 transition-all duration-300 ease-in-out
		${t(l)?"hidden lg:block lg:w-0":"hidden lg:block lg:w-80"}
		${t(f)?"fixed inset-y-0 left-0 z-40 block! w-80":"lg:relative"}
	`),Te(b,1,`flex h-full flex-col ${t(l)?"lg:overflow-hidden":""}`),Te(I,1,`flex h-15 items-center justify-between p-2 ${t(l)?"":"min-w-80"}`),ye(P,"href",S),ye(q,"src",Nt),ye(T,"href",_),ye(K,"aria-label",t(l)?"Open sidebar":"Close sidebar"),Te(F,1,`flex-1 overflow-hidden ${t(l)?"":"min-w-80"}`),Te(D,1,xa(["flex-shrink-0 overflow-y-auto",{"max-h-4/10":t(c)}]))}),d(r,L),Y()}Ue(["click","keydown"]);export{Fn as component,Tn as universal};
