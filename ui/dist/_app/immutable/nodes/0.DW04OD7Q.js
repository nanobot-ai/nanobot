import"../chunks/DsnmJJEf.js";import{V as Re,b0 as De,E as Ae,b1 as Ee,b2 as Le,g as Se,i as Oe,b3 as Ie,b4 as Ue,b5 as qe,b6 as $e,a0 as ze,q as je,aY as R,p as I,ad as O,a as A,aX as j,b as h,c as U,aP as ke,f as D,s as F,d as f,r as u,n as l,ab as Q,b7 as We,T as S,t as Y,e as se,b8 as _e,b9 as pe,ba as me,ae as Be,R as Te,aa as Ve,bb as Xe}from"../chunks/C1840Ik7.js";import{I as X,e as ve,i as He,r as Ke,b as Ge,X as Ne,c as Ye,s as Je,a as Qe,d as Ze,f as de,g as ge}from"../chunks/CCgKi2tH.js";import{s as H,r as K,p as et,i as q}from"../chunks/M7r-35Aa.js";import{g as tt}from"../chunks/DdHERLS9.js";import{i as at}from"../chunks/BunFOKHu.js";const rt=()=>performance.now(),z={tick:r=>requestAnimationFrame(r),now:()=>rt(),tasks:new Set};function Ce(){const r=z.now();z.tasks.forEach(e=>{e.c(r)||(z.tasks.delete(e),e.f())}),z.tasks.size!==0&&z.tick(Ce)}function st(r){let e;return z.tasks.size===0&&z.tick(Ce),{promise:new Promise(t=>{z.tasks.add(e={c:r,f:t})}),abort(){z.tasks.delete(e)}}}function re(r,e){$e(()=>{r.dispatchEvent(new CustomEvent(e))})}function it(r){if(r==="float")return"cssFloat";if(r==="offset")return"cssOffset";if(r.startsWith("--"))return r;const e=r.split("-");return e.length===1?e[0]:e[0]+e.slice(1).map(t=>t[0].toUpperCase()+t.slice(1)).join("")}function be(r){const e={},t=r.split(";");for(const s of t){const[i,n]=s.split(":");if(!i||n===void 0)break;const a=it(i.trim());e[a]=n.trim()}return e}const nt=r=>r;function ye(r,e,t,s){var i=(r&Ue)!==0,n=(r&qe)!==0,a=i&&n,c=(r&Ie)!==0,N=a?"both":i?"in":"out",_,m=e.inert,P=e.style.overflow,y,x;function T(){return $e(()=>_??=t()(e,s?.()??{},{direction:N}))}var b={is_global:c,in(){if(e.inert=m,!i){x?.abort(),x?.reset?.();return}n||y?.abort(),re(e,"introstart"),y=ue(e,T(),x,1,()=>{re(e,"introend"),y?.abort(),y=_=void 0,e.style.overflow=P})},out(w){if(!n){w?.(),_=void 0;return}e.inert=!0,re(e,"outrostart"),x=ue(e,T(),y,0,()=>{re(e,"outroend"),w?.()})},stop:()=>{y?.abort(),x?.abort()}},d=Re;if((d.transitions??=[]).push(b),i&&De){var o=c;if(!o){for(var v=d.parent;v&&(v.f&Ae)!==0;)for(;(v=v.parent)&&(v.f&Ee)===0;);o=!v||(v.f&Le)!==0}o&&Se(()=>{Oe(()=>b.in())})}}function ue(r,e,t,s,i){var n=s===1;if(ze(e)){var a,c=!1;return je(()=>{if(!c){var d=e({direction:n?"in":"out"});a=ue(r,d,t,s,i)}}),{abort:()=>{c=!0,a?.abort()},deactivate:()=>a.deactivate(),reset:()=>a.reset(),t:()=>a.t()}}if(t?.deactivate(),!e?.duration)return i(),{abort:R,deactivate:R,reset:R,t:()=>s};const{delay:N=0,css:_,tick:m,easing:P=nt}=e;var y=[];if(n&&t===void 0&&(m&&m(0,1),_)){var x=be(_(0,1));y.push(x,x)}var T=()=>1-s,b=r.animate(y,{duration:N,fill:"forwards"});return b.onfinish=()=>{b.cancel();var d=t?.t()??1-s;t?.abort();var o=s-d,v=e.duration*Math.abs(o),w=[];if(v>0){var $=!1;if(_)for(var E=Math.ceil(v/16.666666666666668),p=0;p<=E;p+=1){var C=d+o*P(p/E),L=be(_(C,1-C));w.push(L),$||=L.overflow==="hidden"}$&&(r.style.overflow="hidden"),T=()=>{var W=b.currentTime;return d+o*P(W/v)},m&&st(()=>{if(b.playState!=="running")return!1;var W=T();return m(W,1-W),!0})}b=r.animate(w,{duration:v,fill:"forwards"}),b.onfinish=()=>{T=()=>s,m?.(s,1-s),i()}},{abort:()=>{b&&(b.cancel(),b.effect=null,b.onfinish=R)},deactivate:()=>{i=R},reset:()=>{s===0&&m?.(1,0)},t:()=>T()}}const ot=!1,Kt=Object.freeze(Object.defineProperty({__proto__:null,ssr:ot},Symbol.toStringTag,{value:"Module"})),lt="data:image/svg+xml,%3csvg%20xmlns='http://www.w3.org/2000/svg'%20width='107'%20height='128'%20viewBox='0%200%20107%20128'%3e%3ctitle%3esvelte-logo%3c/title%3e%3cpath%20d='M94.157%2022.819c-10.4-14.885-30.94-19.297-45.792-9.835L22.282%2029.608A29.92%2029.92%200%200%200%208.764%2049.65a31.5%2031.5%200%200%200%203.108%2020.231%2030%2030%200%200%200-4.477%2011.183%2031.9%2031.9%200%200%200%205.448%2024.116c10.402%2014.887%2030.942%2019.297%2045.791%209.835l26.083-16.624A29.92%2029.92%200%200%200%2098.235%2078.35a31.53%2031.53%200%200%200-3.105-20.232%2030%2030%200%200%200%204.474-11.182%2031.88%2031.88%200%200%200-5.447-24.116'%20style='fill:%23ff3e00'/%3e%3cpath%20d='M45.817%20106.582a20.72%2020.72%200%200%201-22.237-8.243%2019.17%2019.17%200%200%201-3.277-14.503%2018%2018%200%200%201%20.624-2.435l.49-1.498%201.337.981a33.6%2033.6%200%200%200%2010.203%205.098l.97.294-.09.968a5.85%205.85%200%200%200%201.052%203.878%206.24%206.24%200%200%200%206.695%202.485%205.8%205.8%200%200%200%201.603-.704L69.27%2076.28a5.43%205.43%200%200%200%202.45-3.631%205.8%205.8%200%200%200-.987-4.371%206.24%206.24%200%200%200-6.698-2.487%205.7%205.7%200%200%200-1.6.704l-9.953%206.345a19%2019%200%200%201-5.296%202.326%2020.72%2020.72%200%200%201-22.237-8.243%2019.17%2019.17%200%200%201-3.277-14.502%2017.99%2017.99%200%200%201%208.13-12.052l26.081-16.623a19%2019%200%200%201%205.3-2.329%2020.72%2020.72%200%200%201%2022.237%208.243%2019.17%2019.17%200%200%201%203.277%2014.503%2018%2018%200%200%201-.624%202.435l-.49%201.498-1.337-.98a33.6%2033.6%200%200%200-10.203-5.1l-.97-.294.09-.968a5.86%205.86%200%200%200-1.052-3.878%206.24%206.24%200%200%200-6.696-2.485%205.8%205.8%200%200%200-1.602.704L37.73%2051.72a5.42%205.42%200%200%200-2.449%203.63%205.79%205.79%200%200%200%20.986%204.372%206.24%206.24%200%200%200%206.698%202.486%205.8%205.8%200%200%200%201.602-.704l9.952-6.342a19%2019%200%200%201%205.295-2.328%2020.72%2020.72%200%200%201%2022.237%208.242%2019.17%2019.17%200%200%201%203.277%2014.503%2018%2018%200%200%201-8.13%2012.053l-26.081%2016.622a19%2019%200%200%201-5.3%202.328'%20style='fill:%23fff'/%3e%3c/svg%3e",ct=""+new URL("../assets/nanobot.DvXbFPon.svg",import.meta.url).href;function dt(r,e){I(e,!0);/**
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
 */let t=K(e,["$$slots","$$events","$$legacy"]);const s=[["path",{d:"M20 6 9 17l-5-5"}]];X(r,H({name:"check"},()=>t,{get iconNode(){return s},children:(i,n)=>{var a=O(),c=A(a);j(c,()=>e.children??R),h(i,a)},$$slots:{default:!0}})),U()}function vt(r,e){I(e,!0);/**
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
 */let t=K(e,["$$slots","$$events","$$legacy"]);const s=[["circle",{cx:"12",cy:"12",r:"10"}],["line",{x1:"12",x2:"12",y1:"8",y2:"12"}],["line",{x1:"12",x2:"12.01",y1:"16",y2:"16"}]];X(r,H({name:"circle-alert"},()=>t,{get iconNode(){return s},children:(i,n)=>{var a=O(),c=A(a);j(c,()=>e.children??R),h(i,a)},$$slots:{default:!0}})),U()}function ut(r,e){I(e,!0);/**
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
 */let t=K(e,["$$slots","$$events","$$legacy"]);const s=[["path",{d:"M21.801 10A10 10 0 1 1 17 3.335"}],["path",{d:"m9 11 3 3L22 4"}]];X(r,H({name:"circle-check-big"},()=>t,{get iconNode(){return s},children:(i,n)=>{var a=O(),c=A(a);j(c,()=>e.children??R),h(i,a)},$$slots:{default:!0}})),U()}function ft(r,e){I(e,!0);/**
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
 */let t=K(e,["$$slots","$$events","$$legacy"]);const s=[["circle",{cx:"12",cy:"12",r:"1"}],["circle",{cx:"12",cy:"5",r:"1"}],["circle",{cx:"12",cy:"19",r:"1"}]];X(r,H({name:"ellipsis-vertical"},()=>t,{get iconNode(){return s},children:(i,n)=>{var a=O(),c=A(a);j(c,()=>e.children??R),h(i,a)},$$slots:{default:!0}})),U()}function ht(r,e){I(e,!0);/**
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
 */let t=K(e,["$$slots","$$events","$$legacy"]);const s=[["circle",{cx:"12",cy:"12",r:"10"}],["path",{d:"M12 16v-4"}],["path",{d:"M12 8h.01"}]];X(r,H({name:"info"},()=>t,{get iconNode(){return s},children:(i,n)=>{var a=O(),c=A(a);j(c,()=>e.children??R),h(i,a)},$$slots:{default:!0}})),U()}function _t(r,e){I(e,!0);/**
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
 */let t=K(e,["$$slots","$$events","$$legacy"]);const s=[["path",{d:"M12 3H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"}],["path",{d:"M18.375 2.625a1 1 0 0 1 3 3l-9.013 9.014a2 2 0 0 1-.853.505l-2.873.84a.5.5 0 0 1-.62-.62l.84-2.873a2 2 0 0 1 .506-.852z"}]];X(r,H({name:"square-pen"},()=>t,{get iconNode(){return s},children:(i,n)=>{var a=O(),c=A(a);j(c,()=>e.children??R),h(i,a)},$$slots:{default:!0}})),U()}function pt(r,e){I(e,!0);/**
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
 */let t=K(e,["$$slots","$$events","$$legacy"]);const s=[["path",{d:"M10 11v6"}],["path",{d:"M14 11v6"}],["path",{d:"M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6"}],["path",{d:"M3 6h18"}],["path",{d:"M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"}]];X(r,H({name:"trash-2"},()=>t,{get iconNode(){return s},children:(i,n)=>{var a=O(),c=A(a);j(c,()=>e.children??R),h(i,a)},$$slots:{default:!0}})),U()}function mt(r,e){I(e,!0);/**
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
 */let t=K(e,["$$slots","$$events","$$legacy"]);const s=[["path",{d:"m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"}],["path",{d:"M12 9v4"}],["path",{d:"M12 17h.01"}]];X(r,H({name:"triangle-alert"},()=>t,{get iconNode(){return s},children:(i,n)=>{var a=O(),c=A(a);j(c,()=>e.children??R),h(i,a)},$$slots:{default:!0}})),U()}function gt(r,e,t){r.key==="Enter"?e():r.key==="Escape"&&t()}var bt=D('<div class="flex items-center border-b border-base-200 p-3"><div class="flex-1"><div class="flex items-center justify-between gap-2"><div class="flex min-w-0 flex-1 items-center gap-2"><div class="h-5 w-48 skeleton"></div></div> <div class="h-4 w-8 skeleton"></div></div></div> <div class="w-8"></div></div>'),yt=(r,e,t)=>e(l(t).id),wt=r=>r.stopPropagation(),xt=D('<input type="text" class="input input-sm min-w-0 flex-1"/>'),$t=D('<h3 class="truncate text-sm font-medium"> </h3>'),kt=D('<span class="flex-shrink-0 text-xs text-base-content/50"> </span>'),Tt=D('<div class="flex items-center gap-1 px-2"><button class="btn btn-ghost btn-xs" aria-label="Cancel editing"><!></button> <button class="btn text-success btn-ghost btn-xs hover:bg-success/20" aria-label="Save changes"><!></button></div>'),Nt=(r,e,t)=>e(l(t).id,l(t).title),Ct=(r,e,t)=>e(l(t).id),Mt=D('<div class="dropdown dropdown-end"><div tabindex="0" role="button" class="btn btn-square btn-ghost btn-sm"><!></div> <ul class="dropdown-content menu z-[1] w-32 rounded-box border bg-base-100 p-2 shadow"><li><button class="text-sm"><!> Rename</button></li> <li><button class="text-sm text-error"><!> Delete</button></li></ul></div>'),Ft=D('<div class="flex items-center border-b border-base-200 hover:bg-base-100"><button class="flex-1 p-3 text-left transition-colors focus:outline-none"><div class="flex items-center justify-between gap-2"><div class="flex min-w-0 flex-1 items-center gap-2"><!></div> <!></div></button> <!> <!></div>'),Pt=D('<div class="flex h-full flex-col"><div class="flex-shrink-0 border-b border-base-300 p-4"><h2 class="text-lg font-semibold">Chats</h2></div> <div class="flex-1 overflow-y-auto"><!></div></div>');function Rt(r,e){I(e,!0);let t=et(e,"isLoading",3,!1),s=Q(null),i=Q("");function n(d){tt(`/c/${d}`)}function a(d){const v=new Date().getTime()-new Date(d).getTime(),w=Math.floor(v/(1e3*60)),$=Math.floor(v/(1e3*60*60)),E=Math.floor(v/(1e3*60*60*24));return w<1?"now":w<60?`${w}m`:$<24?`${$}h`:`${E}d`}function c(d,o){S(s,d,!0),S(i,o,!0)}function N(){l(s)&&l(i).trim()&&(e.onRename(l(s),l(i).trim()),S(s,null),S(i,""))}function _(){S(s,null),S(i,"")}function m(d){e.onDelete(d)}var P=Pt(),y=F(f(P),2),x=f(y);{var T=d=>{var o=O(),v=A(o);ve(v,16,()=>Array(5).fill(null),He,(w,$)=>{var E=bt();h(w,E)}),h(d,o)},b=d=>{var o=O(),v=A(o);ve(v,17,()=>e.threads,w=>w.id,(w,$)=>{var E=Ft(),p=f(E);p.__click=[yt,n,$];var C=f(p),L=f(C),W=f(L);{var ie=k=>{var g=xt();Ke(g),g.__keydown=[gt,N,_],g.__click=[wt],We("focus",g,M=>M.target.select()),Ge(g,()=>l(i),M=>S(i,M)),h(k,g)},B=k=>{var g=$t(),M=f(g,!0);u(g),Y(()=>se(M,l($).title)),h(k,g)};q(W,k=>{l(s)===l($).id?k(ie):k(B,!1)})}u(L);var te=F(L,2);{var ne=k=>{var g=kt(),M=f(g,!0);u(g),Y(Z=>se(M,Z),[()=>a(l($).created)]),h(k,g)};q(te,k=>{l(s)!==l($).id&&k(ne)})}u(C),u(p);var ae=F(p,2);{var oe=k=>{var g=Tt(),M=f(g);M.__click=_;var Z=f(M);Ne(Z,{class:"h-3 w-3"}),u(M);var J=F(M,2);J.__click=N;var ee=f(J);dt(ee,{class:"h-3 w-3"}),u(J),u(g),h(k,g)};q(ae,k=>{l(s)===l($).id&&k(oe)})}var V=F(ae,2);{var Me=k=>{var g=Mt(),M=f(g),Z=f(M);ft(Z,{class:"h-4 w-4"}),u(M);var J=F(M,2),ee=f(J),le=f(ee);le.__click=[Nt,c,$];var Fe=f(le);_t(Fe,{class:"h-4 w-4"}),_e(),u(le),u(ee);var he=F(ee,2),ce=f(he);ce.__click=[Ct,m,$];var Pe=f(ce);pt(Pe,{class:"h-4 w-4"}),_e(),u(ce),u(he),u(J),u(g),h(k,g)};q(V,k=>{l(s)!==l($).id&&k(Me)})}u(E),h(w,E)}),h(d,o)};q(x,d=>{t()?d(T):d(b,!1)})}u(y),u(P),h(r,P),U()}ke(["click","keydown"]);function Dt(r){const e=r-1;return e*e*e+1}function we(r,{delay:e=0,duration:t=400,easing:s=Dt,axis:i="y"}={}){const n=getComputedStyle(r),a=+n.opacity,c=i==="y"?"height":"width",N=parseFloat(n[c]),_=i==="y"?["top","bottom"]:["left","right"],m=_.map(o=>`${o[0].toUpperCase()}${o.slice(1)}`),P=parseFloat(n[`padding${m[0]}`]),y=parseFloat(n[`padding${m[1]}`]),x=parseFloat(n[`margin${m[0]}`]),T=parseFloat(n[`margin${m[1]}`]),b=parseFloat(n[`border${m[0]}Width`]),d=parseFloat(n[`border${m[1]}Width`]);return{delay:e,duration:t,easing:s,css:o=>`overflow: hidden;opacity: ${Math.min(o*20,1)*a};${c}: ${o*N}px;padding-${_[0]}: ${o*P}px;padding-${_[1]}: ${o*y}px;margin-${_[0]}: ${o*x}px;margin-${_[1]}: ${o*T}px;border-${_[0]}-width: ${o*b}px;border-${_[1]}-width: ${o*d}px;min-${c}: 0`}}var xe=!1;class fe extends Date{#e=Q(super.getTime());#t=new Map;#a=pe;constructor(...e){super(...e),xe||this.#r()}#r(){xe=!0;var e=fe.prototype,t=Date.prototype,s=Object.getOwnPropertyNames(t);for(const i of s)(i.startsWith("get")||i.startsWith("to")||i==="valueOf")&&(e[i]=function(...n){if(n.length>0)return l(this.#e),t[i].apply(this,n);var a=this.#t.get(i);if(a===void 0){const c=pe;me(this.#a),a=Be(()=>(l(this.#e),t[i].apply(this,n))),this.#t.set(i,a),me(c)}return l(a)}),i.startsWith("set")&&(e[i]=function(...n){var a=t[i].apply(this,n);return S(this.#e,t.getTime.call(this)),a})}}class At{#e=Q(Te([]));get notifications(){return l(this.#e)}set notifications(e){S(this.#e,e,!0)}add(e){const t=crypto.randomUUID(),s={...e,id:t,timestamp:new fe,autoClose:typeof e.autoClose=="boolean"?e.autoClose:e.type!=="error",duration:e.duration||(e.type==="error"?0:5e3)};return this.notifications.push(s),s.autoClose&&s.duration&&s.duration>0&&setTimeout(()=>{this.remove(t)},s.duration),t}remove(e){this.notifications=this.notifications.filter(t=>t.id!==e)}clear(){this.notifications=[]}success(e,t,s){return this.add({type:"success",title:e,message:t,duration:s})}error(e,t){return this.add({type:"error",title:e,message:t,autoClose:!1})}warning(e,t,s){return this.add({type:"warning",title:e,message:t,duration:s})}info(e,t,s){return this.add({type:"info",title:e,message:t,duration:s})}}const G=new At;var Et=D('<div class="mt-1 text-xs opacity-80"> </div>'),Lt=(r,e,t)=>e(l(t).id),St=D('<div class="mt-2 h-1 overflow-hidden rounded bg-black/10"><div class="h-full animate-pulse bg-current opacity-60"></div></div>'),Ot=D('<div><div class="flex items-start gap-3"><div class="flex-shrink-0"><!></div> <div class="min-w-0 flex-1"><div class="text-sm font-medium"> </div> <!></div> <button class="btn btn-square flex-shrink-0 btn-ghost btn-xs" aria-label="Close notification"><!></button></div> <!></div>'),It=D('<div class="fixed right-4 bottom-4 z-50 flex w-80 flex-col gap-3"></div>');function Ut(r,e){I(e,!1);function t(n){const a="alert shadow-lg border";switch(n){case"success":return`${a} alert-success`;case"error":return`${a} alert-error`;case"warning":return`${a} alert-warning`;case"info":return`${a} alert-info`;default:return`${a}`}}function s(n){G.remove(n)}at();var i=It();ve(i,5,()=>G.notifications,n=>n.id,(n,a)=>{var c=Ot(),N=f(c),_=f(N),m=f(_);{var P=p=>{ut(p,{class:"h-5 w-5"})},y=p=>{var C=O(),L=A(C);{var W=B=>{vt(B,{class:"h-5 w-5"})},ie=B=>{var te=O(),ne=A(te);{var ae=V=>{mt(V,{class:"h-5 w-5"})},oe=V=>{ht(V,{class:"h-5 w-5"})};q(ne,V=>{l(a).type==="warning"?V(ae):V(oe,!1)},!0)}h(B,te)};q(L,B=>{l(a).type==="error"?B(W):B(ie,!1)},!0)}h(p,C)};q(m,p=>{l(a).type==="success"?p(P):p(y,!1)})}u(_);var x=F(_,2),T=f(x),b=f(T,!0);u(T);var d=F(T,2);{var o=p=>{var C=Et(),L=f(C,!0);u(C),Y(()=>se(L,l(a).message)),h(p,C)};q(d,p=>{l(a).message&&p(o)})}u(x);var v=F(x,2);v.__click=[Lt,s,a];var w=f(v);Ne(w,{class:"h-4 w-4"}),u(v),u(N);var $=F(N,2);{var E=p=>{var C=St(),L=f(C);u(C),Y(()=>Qe(L,`animation: shrink ${l(a).duration??""}ms linear forwards;`)),h(p,C)};q($,p=>{l(a).autoClose&&l(a).duration&&l(a).duration>0&&p(E)})}u(c),Y(p=>{Je(c,1,p),se(b,l(a).title)},[()=>Ye(t(l(a).type))]),ye(1,c,()=>we,()=>({duration:300})),ye(2,c,()=>we,()=>({duration:200})),h(n,c)}),u(i),h(r,i),U()}ke(["click"]);var qt=D('<link rel="icon"/>'),zt=D('<div id="sidebar" class="drawer lg:drawer-open"><input id="my-drawer" type="checkbox" class="drawer-toggle"/> <div class="drawer-content h-dvh"><!></div> <div class="drawer-side"><label for="my-drawer" aria-label="close sidebar" class="drawer-overlay"></label> <div class="min-h-full w-80 bg-base-200 p-4"><a href="/" class="flex items-center gap-2 text-xl font-bold hover:opacity-80"><img alt="Nanobot" class="h-12"/></a> <!></div></div></div> <!>',1);function Gt(r,e){I(e,!0);let t=Q(Te([])),s=Q(!0);Ze(G),Ve(async()=>{S(t,await de.getThreads(),!0),S(s,!1)});async function i(d,o){try{await de.renameThread(d,o);const v=l(t).findIndex(w=>w.id===d);v!==-1&&(l(t)[v].title=o),G.success("Thread Renamed",`Successfully renamed to "${o}"`)}catch(v){G.error("Rename Failed","Unable to rename the thread. Please try again."),console.error("Failed to rename thread:",v)}}async function n(d){try{await de.deleteThread(d);const o=l(t).find(v=>v.id===d);S(t,l(t).filter(v=>v.id!==d),!0),G.success("Thread Deleted",`Deleted "${o?.title||"thread"}"`)}catch(o){G.error("Delete Failed","Unable to delete the thread. Please try again."),console.error("Failed to delete thread:",o)}}var a=zt();Xe(d=>{var o=qt();Y(()=>ge(o,"href",lt)),h(d,o)});var c=A(a),N=F(f(c),2),_=f(N);j(_,()=>e.children??R),u(N);var m=F(N,2),P=F(f(m),2),y=f(P),x=f(y);u(y);var T=F(y,2);Rt(T,{get threads(){return l(t)},onRename:i,onDelete:n,get isLoading(){return l(s)}}),u(P),u(m),u(c);var b=F(c,2);Ut(b,{}),Y(()=>ge(x,"src",ct)),h(r,a),U()}export{Gt as component,Kt as universal};
