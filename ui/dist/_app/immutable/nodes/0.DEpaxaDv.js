import"../chunks/DsnmJJEf.js";import{Z as Xe,b2 as He,F as Je,b3 as Ke,b4 as Ve,e as Ge,u as Ye,b5 as Qe,b6 as et,b7 as tt,b8 as ze,a5 as at,q as rt,a_ as P,p as W,d as z,f as O,aZ as B,g as h,c as j,aR as Ce,k as R,s as w,l as c,m as l,i as o,af as ae,b9 as st,Y as T,t as ne,B as ye,ba as Pe,bb as Oe,bc as Re,j as nt,X as Ue,o as ot,h as it}from"../chunks/DwQzyupF.js";import{s as J,r as K,p as lt,i as Z}from"../chunks/rhY9p7tb.js";import{I as V,e as Ne,i as ct,r as dt,b as vt,X as Se,a as ut,s as pe,c as ft,T as ht,d as _t,f as ke,h as re}from"../chunks/DjLVYbZw.js";import{g as pt}from"../chunks/BoQToNOB.js";import{r as Me}from"../chunks/BFqM2P5_.js";const bt=()=>performance.now(),se={tick:a=>requestAnimationFrame(a),now:()=>bt(),tasks:new Set};function We(){const a=se.now();se.tasks.forEach(e=>{e.c(a)||(se.tasks.delete(e),e.f())}),se.tasks.size!==0&&se.tick(We)}function mt(a){let e;return se.tasks.size===0&&se.tick(We),{promise:new Promise(t=>{se.tasks.add(e={c:a,f:t})}),abort(){se.tasks.delete(e)}}}function ge(a,e){ze(()=>{a.dispatchEvent(new CustomEvent(e))})}function gt(a){if(a==="float")return"cssFloat";if(a==="offset")return"cssOffset";if(a.startsWith("--"))return a;const e=a.split("-");return e.length===1?e[0]:e[0]+e.slice(1).map(t=>t[0].toUpperCase()+t.slice(1)).join("")}function De(a){const e={},t=a.split(";");for(const s of t){const[r,d]=s.split(":");if(!r||d===void 0)break;const n=gt(r.trim());e[n]=d.trim()}return e}const wt=a=>a;function Ee(a,e,t,s){var r=(a&et)!==0,d=(a&tt)!==0,n=r&&d,i=(a&Qe)!==0,f=n?"both":r?"in":"out",y,b=e.inert,I=e.style.overflow,C,D;function E(){return ze(()=>y??=t()(e,s?.()??{},{direction:f}))}var x={is_global:i,in(){if(e.inert=b,!r){D?.abort(),D?.reset?.();return}d||C?.abort(),ge(e,"introstart"),C=Te(e,E(),D,1,()=>{ge(e,"introend"),C?.abort(),C=y=void 0,e.style.overflow=I})},out(S){if(!d){S?.(),y=void 0;return}e.inert=!0,ge(e,"outrostart"),D=Te(e,E(),C,0,()=>{ge(e,"outroend"),S?.()})},stop:()=>{C?.abort(),D?.abort()}},_=Xe;if((_.transitions??=[]).push(x),r&&He){var u=i;if(!u){for(var g=_.parent;g&&(g.f&Je)!==0;)for(;(g=g.parent)&&(g.f&Ke)===0;);u=!g||(g.f&Ve)!==0}u&&Ge(()=>{Ye(()=>x.in())})}}function Te(a,e,t,s,r){var d=s===1;if(at(e)){var n,i=!1;return rt(()=>{if(!i){var _=e({direction:d?"in":"out"});n=Te(a,_,t,s,r)}}),{abort:()=>{i=!0,n?.abort()},deactivate:()=>n.deactivate(),reset:()=>n.reset(),t:()=>n.t()}}if(t?.deactivate(),!e?.duration)return r(),{abort:P,deactivate:P,reset:P,t:()=>s};const{delay:f=0,css:y,tick:b,easing:I=wt}=e;var C=[];if(d&&t===void 0&&(b&&b(0,1),y)){var D=De(y(0,1));C.push(D,D)}var E=()=>1-s,x=a.animate(C,{duration:f,fill:"forwards"});return x.onfinish=()=>{x.cancel();var _=t?.t()??1-s;t?.abort();var u=s-_,g=e.duration*Math.abs(u),S=[];if(g>0){var N=!1;if(y)for(var F=Math.ceil(g/16.666666666666668),U=0;U<=F;U+=1){var X=_+u*I(U/F),q=De(y(X,1-X));S.push(q),N||=q.overflow==="hidden"}N&&(a.style.overflow="hidden"),E=()=>{var G=x.currentTime;return _+u*I(G/g)},b&&mt(()=>{if(x.playState!=="running")return!1;var G=E();return b(G,1-G),!0})}x=a.animate(S,{duration:g,fill:"forwards"}),x.onfinish=()=>{E=()=>s,b?.(s,1-s),r()}},{abort:()=>{x&&(x.cancel(),x.effect=null,x.onfinish=P)},deactivate:()=>{r=P},reset:()=>{s===0&&b?.(1,0)},t:()=>E()}}const yt=!1,ma=Object.freeze(Object.defineProperty({__proto__:null,ssr:yt},Symbol.toStringTag,{value:"Module"})),xt="data:image/svg+xml,%3c?xml%20version='1.0'%20encoding='UTF-8'?%3e%3csvg%20id='Layer_1'%20data-name='Layer%201'%20xmlns='http://www.w3.org/2000/svg'%20viewBox='0%200%20512%20512'%3e%3cdefs%3e%3cstyle%3e%20.cls-1%20{%20fill:%20%23fff;%20}%20%3c/style%3e%3c/defs%3e%3cpath%20class='cls-1'%20d='M348.33,145.39h-14.33v8.86c0,13.77-5.47,23.43-15.21,33.17-9.74,9.74-22.94,8.11-36.71,8.11h-55.69c-13.77,0-23.43,1.62-33.17-8.11-9.74-9.74-15.21-19.4-15.21-33.17v-8.86h-14.32c-24.89,0-44.77,19.88-44.77,44.77v122.85c0,24.88,19.89,44.77,44.77,44.77h184.65c24.88,0,44.77-19.88,44.77-44.77v-122.85c0-24.89-19.88-44.77-44.77-44.77ZM247.33,284.01c0,5.59-4.53,10.13-10.13,10.13h-39.35c-5.6,0-10.13-4.54-10.13-10.13v-40.62c0-5.59,4.53-10.13,10.13-10.13h39.35c5.6,0,10.13,4.54,10.13,10.13v40.62ZM324.29,284.01c0,5.59-4.54,10.13-10.14,10.13h-39.34c-5.59,0-10.14-4.54-10.14-10.13v-40.62c0-5.59,4.54-10.13,10.14-10.13h39.34c5.6,0,10.14,4.54,10.14,10.13v40.62Z'/%3e%3cpath%20d='M330.95,408.78h-48.98c-3.94,0-7.72,1.57-10.51,4.36-2.79,2.79-4.36,6.57-4.36,10.52v48.04c0,18.94,15.36,34.3,34.3,34.3,11.06,0,21.41-5.47,27.65-14.6,6.87-10.06,16.4-24.01,25.67-37.59,6.02-8.82,6.66-20.24,1.68-29.68-4.98-9.44-14.78-15.35-25.45-15.35Z'/%3e%3cpath%20d='M470.99,212.16c-8.36,0-15.23,6.88-15.23,15.23v18.41h-29.34v-55.65c0-42.92-35.17-78.09-78.09-78.09h-9.02l14.22-53.17c11.11-3.46,18.75-13.99,18.76-25.85,0-14.84-11.95-27.04-26.48-27.04h0c-14.53,0-26.48,12.21-26.48,27.04.03,5.8,1.88,11.44,5.28,16.09l-16.82,62.94h-103.55l-16.82-62.94c3.4-4.64,5.26-10.29,5.28-16.09,0-14.84-11.95-27.04-26.48-27.04h0c-14.53,0-26.48,12.21-26.48,27.04,0,11.86,7.65,22.39,18.76,25.85l14.22,53.17h-9.02c-42.92,0-78.09,35.17-78.09,78.09v55.65h-29.34v-18.41c0-8.36-6.88-15.23-15.23-15.23s-15.23,6.88-15.23,15.23v68.81c.12,8.28,6.96,15.02,15.23,15.02s15.11-6.74,15.23-15.02v-19.94h29.34v36.74c0,42.92,35.18,78.09,78.09,78.09h184.65c42.92,0,78.09-35.17,78.09-78.09v-36.74h29.34v19.94c.12,8.28,6.96,15.02,15.23,15.02s15.11-6.74,15.23-15.02v-68.81c0-8.36-6.88-15.23-15.23-15.23ZM393.09,313c0,24.88-19.88,44.77-44.77,44.77h-184.65c-24.89,0-44.77-19.88-44.77-44.77v-122.85c0-24.89,19.89-44.77,44.77-44.77h14.32v8.86c0,13.77,5.47,23.43,15.21,33.17,9.74,9.74,19.4,8.11,33.17,8.11h55.69c13.77,0,26.97,1.62,36.71-8.11,9.74-9.74,15.21-19.4,15.21-33.17v-8.86h14.33c24.88,0,44.77,19.88,44.77,44.77v122.85Z'/%3e%3cpath%20d='M227.28,94.04c1.8,2.37,4.61,3.77,7.59,3.77s5.79-1.4,7.59-3.77l14.21-18.74,13.63,17.98c1.8,2.37,4.61,3.77,7.59,3.77s5.79-1.4,7.59-3.77l17.23-22.74c3.15-4.16,2.33-10.18-1.83-13.34-1.66-1.26-3.68-1.94-5.76-1.94-2.98,0-5.79,1.39-7.59,3.77l-9.65,12.72-13.63-17.98c-1.8-2.37-4.61-3.76-7.59-3.76h0c-2.98,0-5.79,1.39-7.59,3.76l-14.21,18.75-10.39-13.71c-3.16-4.16-9.18-4.98-13.34-1.83-4.16,3.15-4.99,9.17-1.84,13.34l17.98,23.71Z'/%3e%3cpath%20d='M230.04,408.78h-48.98c-10.67,0-20.47,5.91-25.45,15.35-4.98,9.44-4.34,20.86,1.68,29.68,9.27,13.58,18.8,27.53,25.67,37.59,6.24,9.13,16.59,14.6,27.64,14.6h0c18.94,0,34.3-15.36,34.3-34.3v-48.04c0-3.94-1.57-7.73-4.36-10.52s-6.57-4.36-10.52-4.36Z'/%3e%3cpath%20d='M197.84,233.26h39.35c5.59,0,10.13,4.54,10.13,10.13v40.62c0,5.59-4.54,10.13-10.13,10.13h-39.35c-5.59,0-10.13-4.54-10.13-10.13v-40.62c0-5.59,4.54-10.13,10.13-10.13Z'/%3e%3cpath%20d='M314.15,233.26h-39.34c-5.59,0-10.14,4.54-10.14,10.13v40.62c0,5.59,4.54,10.13,10.14,10.13h39.34c5.6,0,10.14-4.54,10.14-10.13v-40.62c0-5.59-4.54-10.13-10.14-10.13Z'/%3e%3c/svg%3e",Ie=""+new URL("../assets/nanobot.Bn3X0Wtr.svg",import.meta.url).href;function $t(a,e){W(e,!0);/**
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
 */let t=K(e,["$$slots","$$events","$$legacy"]);const s=[["path",{d:"M20 6 9 17l-5-5"}]];V(a,J({name:"check"},()=>t,{get iconNode(){return s},children:(r,d)=>{var n=z(),i=O(n);B(i,()=>e.children??P),h(r,n)},$$slots:{default:!0}})),j()}function kt(a,e){W(e,!0);/**
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
 */let t=K(e,["$$slots","$$events","$$legacy"]);const s=[["circle",{cx:"12",cy:"12",r:"10"}],["line",{x1:"12",x2:"12",y1:"8",y2:"12"}],["line",{x1:"12",x2:"12.01",y1:"16",y2:"16"}]];V(a,J({name:"circle-alert"},()=>t,{get iconNode(){return s},children:(r,d)=>{var n=z(),i=O(n);B(i,()=>e.children??P),h(r,n)},$$slots:{default:!0}})),j()}function Nt(a,e){W(e,!0);/**
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
 */let t=K(e,["$$slots","$$events","$$legacy"]);const s=[["path",{d:"M21.801 10A10 10 0 1 1 17 3.335"}],["path",{d:"m9 11 3 3L22 4"}]];V(a,J({name:"circle-check-big"},()=>t,{get iconNode(){return s},children:(r,d)=>{var n=z(),i=O(n);B(i,()=>e.children??P),h(r,n)},$$slots:{default:!0}})),j()}function Mt(a,e){W(e,!0);/**
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
 */let t=K(e,["$$slots","$$events","$$legacy"]);const s=[["circle",{cx:"12",cy:"12",r:"1"}],["circle",{cx:"12",cy:"5",r:"1"}],["circle",{cx:"12",cy:"19",r:"1"}]];V(a,J({name:"ellipsis-vertical"},()=>t,{get iconNode(){return s},children:(r,d)=>{var n=z(),i=O(n);B(i,()=>e.children??P),h(r,n)},$$slots:{default:!0}})),j()}function Tt(a,e){W(e,!0);/**
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
 */let t=K(e,["$$slots","$$events","$$legacy"]);const s=[["circle",{cx:"12",cy:"12",r:"10"}],["path",{d:"M12 16v-4"}],["path",{d:"M12 8h.01"}]];V(a,J({name:"info"},()=>t,{get iconNode(){return s},children:(r,d)=>{var n=z(),i=O(n);B(i,()=>e.children??P),h(r,n)},$$slots:{default:!0}})),j()}function Ct(a,e){W(e,!0);/**
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
 */let t=K(e,["$$slots","$$events","$$legacy"]);const s=[["path",{d:"M4 12h16"}],["path",{d:"M4 18h16"}],["path",{d:"M4 6h16"}]];V(a,J({name:"menu"},()=>t,{get iconNode(){return s},children:(r,d)=>{var n=z(),i=O(n);B(i,()=>e.children??P),h(r,n)},$$slots:{default:!0}})),j()}function St(a,e){W(e,!0);/**
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
 */let t=K(e,["$$slots","$$events","$$legacy"]);const s=[["path",{d:"M20.985 12.486a9 9 0 1 1-9.473-9.472c.405-.022.617.46.402.803a6 6 0 0 0 8.268 8.268c.344-.215.825-.004.803.401"}]];V(a,J({name:"moon"},()=>t,{get iconNode(){return s},children:(r,d)=>{var n=z(),i=O(n);B(i,()=>e.children??P),h(r,n)},$$slots:{default:!0}})),j()}function Ft(a,e){W(e,!0);/**
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
 */let t=K(e,["$$slots","$$events","$$legacy"]);const s=[["rect",{width:"18",height:"18",x:"3",y:"3",rx:"2"}],["path",{d:"M9 3v18"}],["path",{d:"m16 15-3-3 3-3"}]];V(a,J({name:"panel-left-close"},()=>t,{get iconNode(){return s},children:(r,d)=>{var n=z(),i=O(n);B(i,()=>e.children??P),h(r,n)},$$slots:{default:!0}})),j()}function Ae(a,e){W(e,!0);/**
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
 */let t=K(e,["$$slots","$$events","$$legacy"]);const s=[["rect",{width:"18",height:"18",x:"3",y:"3",rx:"2"}],["path",{d:"M9 3v18"}],["path",{d:"m14 9 3 3-3 3"}]];V(a,J({name:"panel-left-open"},()=>t,{get iconNode(){return s},children:(r,d)=>{var n=z(),i=O(n);B(i,()=>e.children??P),h(r,n)},$$slots:{default:!0}})),j()}function we(a,e){W(e,!0);/**
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
 */let t=K(e,["$$slots","$$events","$$legacy"]);const s=[["path",{d:"M12 3H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"}],["path",{d:"M18.375 2.625a1 1 0 0 1 3 3l-9.013 9.014a2 2 0 0 1-.853.505l-2.873.84a.5.5 0 0 1-.62-.62l.84-2.873a2 2 0 0 1 .506-.852z"}]];V(a,J({name:"square-pen"},()=>t,{get iconNode(){return s},children:(r,d)=>{var n=z(),i=O(n);B(i,()=>e.children??P),h(r,n)},$$slots:{default:!0}})),j()}function Pt(a,e){W(e,!0);/**
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
 */let t=K(e,["$$slots","$$events","$$legacy"]);const s=[["circle",{cx:"12",cy:"12",r:"4"}],["path",{d:"M12 2v2"}],["path",{d:"M12 20v2"}],["path",{d:"m4.93 4.93 1.41 1.41"}],["path",{d:"m17.66 17.66 1.41 1.41"}],["path",{d:"M2 12h2"}],["path",{d:"M20 12h2"}],["path",{d:"m6.34 17.66-1.41 1.41"}],["path",{d:"m19.07 4.93-1.41 1.41"}]];V(a,J({name:"sun"},()=>t,{get iconNode(){return s},children:(r,d)=>{var n=z(),i=O(n);B(i,()=>e.children??P),h(r,n)},$$slots:{default:!0}})),j()}function Ot(a,e){W(e,!0);/**
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
 */let t=K(e,["$$slots","$$events","$$legacy"]);const s=[["path",{d:"M10 11v6"}],["path",{d:"M14 11v6"}],["path",{d:"M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6"}],["path",{d:"M3 6h18"}],["path",{d:"M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"}]];V(a,J({name:"trash-2"},()=>t,{get iconNode(){return s},children:(r,d)=>{var n=z(),i=O(n);B(i,()=>e.children??P),h(r,n)},$$slots:{default:!0}})),j()}function Rt(a,e,t){a.key==="Enter"?e():a.key==="Escape"&&t()}var Dt=R('<div class="flex items-center border-b border-base-200 p-3"><div class="flex-1"><div class="flex items-center justify-between gap-2"><div class="flex min-w-0 flex-1 items-center gap-2"><div class="h-5 w-48 skeleton"></div></div> <div class="h-4 w-8 skeleton"></div></div></div> <div class="w-8"></div></div>'),Et=(a,e,t)=>e(o(t).id),It=a=>a.stopPropagation(),At=R('<input type="text" class="input input-sm min-w-0 flex-1"/>'),Lt=R('<h3 class="truncate text-sm font-medium"> </h3>'),Zt=R('<span class="flex-shrink-0 text-xs text-base-content/50"> </span>'),zt=R('<div class="flex items-center gap-1 px-2"><button class="btn btn-ghost btn-xs" aria-label="Cancel editing"><!></button> <button class="btn text-success btn-ghost btn-xs hover:bg-success/20" aria-label="Save changes"><!></button></div>'),Ut=(a,e,t)=>e(o(t).id,o(t).title),Wt=(a,e,t)=>e(o(t).id),jt=R('<div class="dropdown dropdown-end opacity-0 transition-opacity group-hover:opacity-100"><div tabindex="0" role="button" class="btn btn-square btn-ghost btn-sm"><!></div> <ul class="dropdown-content menu z-[1] w-32 rounded-box border bg-base-100 p-2 shadow"><li><button class="text-sm"><!> Rename</button></li> <li><button class="text-sm text-error"><!> Delete</button></li></ul></div>'),qt=R('<div class="group flex items-center border-b border-base-200 hover:bg-base-100"><button class="flex-1 truncate p-3 text-left transition-colors focus:outline-none"><div class="flex items-center justify-between gap-2"><div class="flex min-w-0 flex-1 items-center gap-2"><!></div> <!></div></button> <!> <!></div>'),Bt=R('<div class="flex h-full flex-col"><div class="flex-shrink-0 p-2"><h2 class="font-semibold text-base-content/60">Conversations</h2></div> <div class="flex-1 overflow-y-auto"><!></div></div>');function Xt(a,e){W(e,!0);let t=lt(e,"isLoading",3,!1),s=ae(null),r=ae("");function d(_){pt(Me(`/c/${_}`)),e.onThreadClick?.()}function n(_){const g=new Date().getTime()-new Date(_).getTime(),S=Math.floor(g/(1e3*60)),N=Math.floor(g/(1e3*60*60)),F=Math.floor(g/(1e3*60*60*24));return S<1?"now":S<60?`${S}m`:N<24?`${N}h`:`${F}d`}function i(_,u){T(s,_,!0),T(r,u||"",!0)}function f(){o(s)&&o(r).trim()&&(e.onRename(o(s),o(r).trim()),T(s,null),T(r,""))}function y(){T(s,null),T(r,"")}function b(_){e.onDelete(_)}var I=Bt(),C=w(c(I),2),D=c(C);{var E=_=>{var u=z(),g=O(u);Ne(g,16,()=>Array(5).fill(null),ct,(S,N)=>{var F=Dt();h(S,F)}),h(_,u)},x=_=>{var u=z(),g=O(u);Ne(g,17,()=>e.threads,S=>S.id,(S,N)=>{var F=qt(),U=c(F);U.__click=[Et,d,N];var X=c(U),q=c(X),G=c(q);{var ue=k=>{var p=At();dt(p),p.__keydown=[Rt,f,y],p.__click=[It],st("focus",p,M=>M.target.select()),vt(p,()=>o(r),M=>T(r,M)),h(k,p)},fe=k=>{var p=Lt(),M=c(p,!0);l(p),ne(()=>ye(M,o(N).title||"Untitled")),h(k,p)};Z(G,k=>{o(s)===o(N).id?k(ue):k(fe,!1)})}l(q);var ce=w(q,2);{var he=k=>{var p=Zt(),M=c(p,!0);l(p),ne(Q=>ye(M,Q),[()=>n(o(N).created)]),h(k,p)};Z(ce,k=>{o(s)!==o(N).id&&k(he)})}l(X),l(U);var m=w(U,2);{var A=k=>{var p=zt(),M=c(p);M.__click=y;var Q=c(M);Se(Q,{class:"h-3 w-3"}),l(M);var ee=w(M,2);ee.__click=f;var te=c(ee);$t(te,{class:"h-3 w-3"}),l(ee),l(p),h(k,p)};Z(m,k=>{o(s)===o(N).id&&k(A)})}var Y=w(m,2);{var ie=k=>{var p=jt(),M=c(p),Q=c(M);Mt(Q,{class:"h-4 w-4"}),l(M);var ee=w(M,2),te=c(ee),H=c(te);H.__click=[Ut,i,N];var be=c(H);we(be,{class:"h-4 w-4"}),Pe(),l(H),l(te);var me=w(te,2),de=c(me);de.__click=[Wt,b,N];var xe=c(de);Ot(xe,{class:"h-4 w-4"}),Pe(),l(de),l(me),l(ee),l(p),h(k,p)};Z(Y,k=>{o(s)!==o(N).id&&k(ie)})}l(F),h(S,F)}),h(_,u)};Z(D,_=>{t()?_(E):_(x,!1)})}l(C),l(I),h(a,I),j()}Ce(["click","keydown"]);function Ht(a){const e=a-1;return e*e*e+1}function Le(a,{delay:e=0,duration:t=400,easing:s=Ht,axis:r="y"}={}){const d=getComputedStyle(a),n=+d.opacity,i=r==="y"?"height":"width",f=parseFloat(d[i]),y=r==="y"?["top","bottom"]:["left","right"],b=y.map(u=>`${u[0].toUpperCase()}${u.slice(1)}`),I=parseFloat(d[`padding${b[0]}`]),C=parseFloat(d[`padding${b[1]}`]),D=parseFloat(d[`margin${b[0]}`]),E=parseFloat(d[`margin${b[1]}`]),x=parseFloat(d[`border${b[0]}Width`]),_=parseFloat(d[`border${b[1]}Width`]);return{delay:e,duration:t,easing:s,css:u=>`overflow: hidden;opacity: ${Math.min(u*20,1)*n};${i}: ${u*f}px;padding-${y[0]}: ${u*I}px;padding-${y[1]}: ${u*C}px;margin-${y[0]}: ${u*D}px;margin-${y[1]}: ${u*E}px;border-${y[0]}-width: ${u*x}px;border-${y[1]}-width: ${u*_}px;min-${i}: 0`}}var Ze=!1;class Fe extends Date{#e=ae(super.getTime());#t=new Map;#a=Oe;constructor(...e){super(...e),Ze||this.#r()}#r(){Ze=!0;var e=Fe.prototype,t=Date.prototype,s=Object.getOwnPropertyNames(t);for(const r of s)(r.startsWith("get")||r.startsWith("to")||r==="valueOf")&&(e[r]=function(...d){if(d.length>0)return o(this.#e),t[r].apply(this,d);var n=this.#t.get(r);if(n===void 0){const i=Oe;Re(this.#a),n=nt(()=>(o(this.#e),t[r].apply(this,d))),this.#t.set(r,n),Re(i)}return o(n)}),r.startsWith("set")&&(e[r]=function(...d){var n=t[r].apply(this,d);return T(this.#e,t.getTime.call(this)),n})}}class Jt{#e=ae(Ue([]));get notifications(){return o(this.#e)}set notifications(e){T(this.#e,e,!0)}add(e){const t=crypto.randomUUID(),s={...e,id:t,timestamp:new Fe,autoClose:typeof e.autoClose=="boolean"?e.autoClose:e.type!=="error",duration:e.duration||(e.type==="error"?0:5e3)};return this.notifications.push(s),s.autoClose&&s.duration&&s.duration>0&&setTimeout(()=>{this.remove(t)},s.duration),t}remove(e){this.notifications=this.notifications.filter(t=>t.id!==e)}clear(){this.notifications=[]}success(e,t,s){return this.add({type:"success",title:e,message:t,duration:s})}error(e,t){return this.add({type:"error",title:e,message:t,autoClose:!1})}warning(e,t,s){return this.add({type:"warning",title:e,message:t,duration:s})}info(e,t,s){return this.add({type:"info",title:e,message:t,duration:s})}}const le=new Jt;var Kt=R('<div class="mt-1 text-xs break-all opacity-80"> </div>'),Vt=(a,e,t)=>e(o(t)),Gt=(a,e,t)=>e(o(t).id),Yt=R('<div class="absolute -top-8 right-1 rounded bg-success px-2 py-1 text-xs text-success-content opacity-100 shadow-lg transition-opacity duration-500">Copied!</div>'),Qt=R('<div class="mt-2 h-1 overflow-hidden rounded bg-black/10"><div class="h-full animate-pulse bg-current opacity-60"></div></div>'),ea=R('<div><div class="flex items-start gap-3"><div class="flex-shrink-0"><!></div> <div class="min-w-0 flex-1"><div class="text-sm font-medium break-all"> </div> <!></div></div> <div class="absolute top-1 right-1 flex gap-1 rounded p-1 opacity-0 backdrop-blur-sm transition-opacity group-hover:opacity-100"><button type="button" class="btn btn-ghost btn-xs" title="Copy notification"><!></button> <button class="btn btn-ghost btn-xs" aria-label="Close notification"><!></button></div> <!></div> <!>',1),ta=R('<div class="fixed right-4 bottom-4 z-50 flex w-80 flex-col gap-3"></div>');function aa(a,e){W(e,!0);let t=ae(null);function s(i){const f="alert shadow-lg border";switch(i){case"success":return`${f} alert-success`;case"error":return`${f} alert-error`;case"warning":return`${f} alert-warning`;case"info":return`${f} alert-info`;default:return`${f}`}}function r(i){le.remove(i)}async function d(i){const f=i.message?`${i.title}
${i.message}`:i.title;await navigator.clipboard.writeText(f),T(t,i.id,!0),setTimeout(()=>{T(t,null)},2e3)}var n=ta();Ne(n,21,()=>le.notifications,i=>i.id,(i,f)=>{var y=ea(),b=O(y),I=c(b),C=c(I),D=c(C);{var E=m=>{Nt(m,{class:"h-5 w-5"})},x=m=>{var A=z(),Y=O(A);{var ie=p=>{kt(p,{class:"h-5 w-5"})},k=p=>{var M=z(),Q=O(M);{var ee=H=>{ht(H,{class:"h-5 w-5"})},te=H=>{Tt(H,{class:"h-5 w-5"})};Z(Q,H=>{o(f).type==="warning"?H(ee):H(te,!1)},!0)}h(p,M)};Z(Y,p=>{o(f).type==="error"?p(ie):p(k,!1)},!0)}h(m,A)};Z(D,m=>{o(f).type==="success"?m(E):m(x,!1)})}l(C);var _=w(C,2),u=c(_),g=c(u,!0);l(u);var S=w(u,2);{var N=m=>{var A=Kt(),Y=c(A,!0);l(A),ne(()=>ye(Y,o(f).message)),h(m,A)};Z(S,m=>{o(f).message&&m(N)})}l(_),l(I);var F=w(I,2),U=c(F);U.__click=[Vt,d,f];var X=c(U);ut(X,{class:"h-3 w-3"}),l(U);var q=w(U,2);q.__click=[Gt,r,f];var G=c(q);Se(G,{class:"h-3 w-3"}),l(q),l(F);var ue=w(F,2);{var fe=m=>{var A=Yt();h(m,A)};Z(ue,m=>{o(t)===o(f).id&&m(fe)})}l(b);var ce=w(b,2);{var he=m=>{var A=Qt(),Y=c(A);l(A),ne(()=>ft(Y,`animation: shrink ${o(f).duration??""}ms linear forwards;`)),h(m,A)};Z(ce,m=>{o(f).autoClose&&o(f).duration&&o(f).duration>0&&m(he)})}ne(m=>{pe(b,1,`${m??""} group relative`),ye(g,o(f).title)},[()=>s(o(f).type)]),Ee(1,b,()=>Le,()=>({duration:300})),Ee(2,b,()=>Le,()=>({duration:200})),h(i,y)}),l(n),h(a,n),j()}Ce(["click"]);function ra(a,e){T(e,!o(e))}function sa(a,e){T(e,o(e)==="lofi"?"black":"lofi",!0),document.documentElement.setAttribute("data-theme",o(e)),localStorage.setItem("theme",o(e))}var na=R('<link rel="icon"/>'),oa=(a,e,t)=>{window.innerWidth>=1024?e():t()},ia=(a,e)=>a.key==="Enter"||a.key===" "?e():null,la=R('<div class="fixed inset-0 z-30 bg-black/50 lg:hidden" role="button" tabindex="0"></div>'),ca=R('<div class="absolute top-0 left-0 z-10 hidden h-15 items-center bg-transparent p-2 lg:flex"><div class="flex items-center gap-2"><a class="flex items-center gap-2 text-xl font-bold hover:opacity-80"><img alt="Nanobot" class="h-12"/></a> <a class="btn p-1 btn-ghost btn-sm" aria-label="New thread"><!></a> <button class="btn p-1 btn-ghost btn-sm" aria-label="Open sidebar"><!></button></div></div>'),da=R('<div class="absolute top-4 left-4 z-50 flex gap-2 lg:hidden"><a class="btn border border-base-300 bg-base-100/80 btn-ghost backdrop-blur-sm btn-sm" aria-label="New thread"><!></a> <button class="btn border border-base-300 bg-base-100/80 btn-ghost backdrop-blur-sm btn-sm" aria-label="Open sidebar"><!></button></div>'),va=R('<div class="relative flex h-dvh"><div><div><div><a class="flex items-center gap-2 text-xl font-bold hover:opacity-80"><img alt="Nanobot" class="h-12"/></a> <div class="flex items-center gap-1"><a class="btn p-1 btn-ghost btn-sm" aria-label="New thread"><!></a> <button class="btn p-1 btn-ghost btn-sm"><span class="hidden lg:inline"><!></span> <span class="lg:hidden"><!></span></button></div></div> <div><!></div> <div class="absolute bottom-4 left-4"><button class="btn btn-circle border-base-300 bg-base-100 shadow-lg btn-sm" aria-label="Toggle theme"><!></button></div></div></div> <!> <!> <!> <div class="h-dvh flex-1"><!></div></div> <!>',1);function ga(a,e){W(e,!0);let t=ae(Ue([])),s=ae(!0),r=ae(!1),d=ae(!1),n=ae("lofi");const i=Me("/"),f=Me("/");_t(le),ot(async()=>{if(window.innerWidth>=1024){const v=localStorage.getItem("sidebar-collapsed");v!==null&&T(r,JSON.parse(v),!0)}{const v=localStorage.getItem("theme");if(v)T(n,v,!0);else{const $=window.matchMedia("(prefers-color-scheme: dark)").matches;T(n,$?"black":"lofi",!0)}document.documentElement.setAttribute("data-theme",o(n))}T(t,await ke.getThreads(),!0),T(s,!1)});function y(){window.innerWidth>=1024&&(T(r,!o(r)),localStorage.setItem("sidebar-collapsed",JSON.stringify(o(r))))}function b(){T(d,!1)}async function I(v,$){try{await ke.renameThread(v,$);const L=o(t).findIndex(oe=>oe.id===v);L!==-1&&(o(t)[L].title=$),le.success("Thread Renamed",`Successfully renamed to "${$}"`)}catch(L){le.error("Rename Failed","Unable to rename the thread. Please try again."),console.error("Failed to rename thread:",L)}}async function C(v){try{await ke.deleteThread(v);const $=o(t).find(L=>L.id===v);T(t,o(t).filter(L=>L.id!==v),!0),le.success("Thread Deleted",`Deleted "${$?.title||"thread"}"`)}catch($){le.error("Delete Failed","Unable to delete the thread. Please try again."),console.error("Failed to delete thread:",$)}}var D=va();it(v=>{var $=na();ne(()=>re($,"href",xt)),h(v,$)});var E=O(D),x=c(E),_=c(x),u=c(_),g=c(u),S=c(g);l(g);var N=w(g,2),F=c(N),U=c(F);we(U,{class:"h-5 w-5"}),l(F);var X=w(F,2);X.__click=[oa,y,b];var q=c(X),G=c(q);{var ue=v=>{Ae(v,{class:"h-5 w-5"})},fe=v=>{Ft(v,{class:"h-5 w-5"})};Z(G,v=>{o(r)?v(ue):v(fe,!1)})}l(q);var ce=w(q,2),he=c(ce);Se(he,{class:"h-5 w-5"}),l(ce),l(X),l(N),l(u);var m=w(u,2),A=c(m);Xt(A,{get threads(){return o(t)},onRename:I,onDelete:C,get isLoading(){return o(s)},onThreadClick:b}),l(m);var Y=w(m,2),ie=c(Y);ie.__click=[sa,n];var k=c(ie);{var p=v=>{St(v,{class:"h-4 w-4"})},M=v=>{Pt(v,{class:"h-4 w-4"})};Z(k,v=>{o(n)==="lofi"?v(p):v(M,!1)})}l(ie),l(Y),l(_),l(x);var Q=w(x,2);{var ee=v=>{var $=la();$.__click=b,$.__keydown=[ia,b],h(v,$)};Z(Q,v=>{o(d)&&v(ee)})}var te=w(Q,2);{var H=v=>{var $=ca(),L=c($),oe=c(L),_e=c(oe);l(oe);var ve=w(oe,2),qe=c(ve);we(qe,{class:"h-4 w-4"}),l(ve);var $e=w(ve,2);$e.__click=y;var Be=c($e);Ae(Be,{class:"h-4 w-4"}),l($e),l(L),l($),ne(()=>{re(oe,"href",i),re(_e,"src",Ie),re(ve,"href",f)}),h(v,$)};Z(te,v=>{o(r)&&v(H)})}var be=w(te,2);{var me=v=>{var $=da(),L=c($),oe=c(L);we(oe,{class:"h-5 w-5"}),l(L);var _e=w(L,2);_e.__click=[ra,d];var ve=c(_e);Ct(ve,{class:"h-5 w-5"}),l(_e),l($),ne(()=>re(L,"href",f)),h(v,$)};Z(be,v=>{o(d)||v(me)})}var de=w(be,2),xe=c(de);B(xe,()=>e.children??P),l(de),l(E);var je=w(E,2);aa(je,{}),ne(()=>{pe(x,1,`
		bg-base-200 transition-all duration-300 ease-in-out
		${o(r)?"hidden lg:block lg:w-0":"hidden lg:block lg:w-80"}
		${o(d)?"fixed inset-y-0 left-0 z-40 block! w-80":"lg:relative"}
	`),pe(_,1,`flex h-full flex-col ${o(r)?"lg:overflow-hidden":""}`),pe(u,1,`flex h-15 items-center justify-between p-2 ${o(r)?"":"min-w-80"}`),re(g,"href",i),re(S,"src",Ie),re(F,"href",f),re(X,"aria-label",o(r)?"Open sidebar":"Close sidebar"),pe(m,1,`flex-1 overflow-hidden ${o(r)?"":"min-w-80"}`)}),h(a,D),j()}Ce(["click","keydown"]);export{ga as component,ma as universal};
