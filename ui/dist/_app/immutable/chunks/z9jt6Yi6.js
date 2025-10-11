import{L as ye,C as x,ai as Le,J as W,D as w,aj as ke,E as pe,i as C,a0 as Xe,G as Ze,H as xe,I as Ee,K as U,R as P,ak as Qe,al as et,M as $,O as tt,am as Q,an as ee,N as D,ao as Ae,ae as st,ap as Ce,aq as Fe,ar as Ue,P as me,Q as De,as as rt,at as ce,q as at,au as R,av as it,aw as nt,ax as ot,ay as lt,az as ut,Z as $e,aA as ct,F as ft,aB as dt,aC as ht,aD as vt,aE as Se,e as be,aF as gt,aG as pt,aH as _t,aI as Re,aJ as yt,aK as mt,aL as bt,aM as It,aN as wt,aO as Tt,aP as Et,aQ as At,aR as Ct,aS as St,aT as Nt,U as Mt,aU as Ot,aV as Pt,aW as Lt,ag as kt,u as Ft,r as Ut,aX as Dt,p as te,aY as $t,aZ as se,g as z,c as re,l as Rt,s as jt,a_ as ae,m as Ht,d as ie,f as ne,j as Vt,a$ as Jt,b0 as Wt,b1 as Yt,Y as N,af as M,X as F}from"./DwQzyupF.js";import"./DsnmJJEf.js";import{p as H,r as oe,s as Ie}from"./rhY9p7tb.js";function zt(e,t){return t}function Bt(e,t,s){for(var r=e.items,a=[],i=t.length,n=0;n<i;n++)ot(t[n].e,a,!0);var o=i>0&&a.length===0&&s!==null;if(o){var v=s.parentNode;lt(v),v.append(s),r.clear(),O(e,t[0].prev,t[i-1].next)}ut(a,()=>{for(var l=0;l<i;l++){var y=t[l];o||(r.delete(y.k),O(e,y.prev,y.next)),R(y.e,!o)}})}function Kt(e,t,s,r,a,i=null){var n=e,o={flags:t,items:new Map,first:null},v=(t&Le)!==0;if(v){var l=e;n=w?W(ke(l)):l.appendChild(ye())}w&&pe();var y=null,g=!1,d=new Map,T=Xe(()=>{var _=s();return Ue(_)?_:_==null?[]:Fe(_)}),u,h;function m(){qt(h,u,o,d,n,a,t,r,s),i!==null&&(u.length===0?y?me(y):y=$(()=>i(n)):y!==null&&De(y,()=>{y=null}))}x(()=>{h??=$e,u=C(T);var _=u.length;if(g&&_===0)return;g=_===0;let E=!1;if(w){var p=Ze(n)===xe;p!==(_===0)&&(n=Ee(),W(n),U(!1),E=!0)}if(w){for(var b=null,I,c=0;c<_;c++){if(P.nodeType===Qe&&P.data===et){n=P,E=!0,U(!1);break}var f=u[c],A=r(f,c);I=_e(P,o,b,null,f,A,c,a,t,s),o.items.set(A,I),b=I}_>0&&W(Ee())}if(w)_===0&&i&&(y=$(()=>i(n)));else if(tt()){var L=new Set,B=D;for(c=0;c<_;c+=1){f=u[c],A=r(f,c);var k=o.items.get(A)??d.get(A);k?(t&(Q|ee))!==0&&je(k,f,c,t):(I=_e(null,o,null,null,f,A,c,a,t,s,!0),d.set(A,I)),L.add(A)}for(const[S,K]of o.items)L.has(S)||B.skipped_effects.add(K.e);B.add_callback(m)}else m();E&&U(!0),C(T)}),w&&(n=P)}function qt(e,t,s,r,a,i,n,o,v){var l=(n&it)!==0,y=(n&(Q|ee))!==0,g=t.length,d=s.items,T=s.first,u=T,h,m=null,_,E=[],p=[],b,I,c,f;if(l)for(f=0;f<g;f+=1)b=t[f],I=o(b,f),c=d.get(I),c!==void 0&&(c.a?.measure(),(_??=new Set).add(c));for(f=0;f<g;f+=1){if(b=t[f],I=o(b,f),c=d.get(I),c===void 0){var A=r.get(I);if(A!==void 0){r.delete(I),d.set(I,A);var L=m?m.next:u;O(s,m,A),O(s,A,L),fe(A,L,a),m=A}else{var B=u?u.e.nodes_start:a;m=_e(B,s,m,m===null?s.first:m.next,b,I,f,i,n,v)}d.set(I,m),E=[],p=[],u=m.next;continue}if(y&&je(c,b,f,n),(c.e.f&ce)!==0&&(me(c.e),l&&(c.a?.unfix(),(_??=new Set).delete(c))),c!==u){if(h!==void 0&&h.has(c)){if(E.length<p.length){var k=p[0],S;m=k.prev;var K=E[0],le=E[E.length-1];for(S=0;S<E.length;S+=1)fe(E[S],k,a);for(S=0;S<p.length;S+=1)h.delete(p[S]);O(s,K.prev,le.next),O(s,m,K),O(s,le,k),u=k,m=le,f-=1,E=[],p=[]}else h.delete(c),fe(c,u,a),O(s,c.prev,c.next),O(s,c,m===null?s.first:m.next),O(s,m,c),m=c;continue}for(E=[],p=[];u!==null&&u.k!==I;)(u.e.f&ce)===0&&(h??=new Set).add(u),p.push(u),u=u.next;if(u===null)continue;c=u}E.push(c),m=c,u=c.next}if(u!==null||h!==void 0){for(var j=h===void 0?[]:Fe(h);u!==null;)(u.e.f&ce)===0&&j.push(u),u=u.next;var ue=j.length;if(ue>0){var qe=(n&Le)!==0&&g===0?a:null;if(l){for(f=0;f<ue;f+=1)j[f].a?.measure();for(f=0;f<ue;f+=1)j[f].a?.fix()}Bt(s,j,qe)}}l&&at(()=>{if(_!==void 0)for(c of _)c.a?.apply()}),e.first=s.first&&s.first.e,e.last=m&&m.e;for(var Ge of r.values())R(Ge.e);r.clear()}function je(e,t,s,r){(r&Q)!==0&&Ae(e.v,t),(r&ee)!==0?Ae(e.i,s):e.i=s}function _e(e,t,s,r,a,i,n,o,v,l,y){var g=(v&Q)!==0,d=(v&rt)===0,T=g?d?st(a,!1,!1):Ce(a):a,u=(v&ee)===0?n:Ce(n),h={i:u,v:T,k:i,a:null,e:null,prev:s,next:r};try{if(e===null){var m=document.createDocumentFragment();m.append(e=ye())}return h.e=$(()=>o(e,T,u,l),w),h.e.prev=s&&s.e,h.e.next=r&&r.e,s===null?y||(t.first=h):(s.next=h,s.e.next=h.e),r!==null&&(r.prev=h,r.e.prev=h.e),h}finally{}}function fe(e,t,s){for(var r=e.next?e.next.e.nodes_start:s,a=t?t.e.nodes_start:s,i=e.e.nodes_start;i!==null&&i!==r;){var n=nt(i);a.before(i),i=n}}function O(e,t,s){t===null?e.first=s:(t.next=s,t.e.next=s&&s.e),s!==null&&(s.prev=t,s.e.prev=t&&t.e)}function Gt(e,t,s,r,a,i){let n=w;w&&pe();var o,v,l=null;w&&P.nodeType===ct&&(l=P,pe());var y=w?P:e,g;x(()=>{const d=t()||null;var T=dt;d!==o&&(g&&(d===null?De(g,()=>{g=null,v=null}):d===v?me(g):(R(g),Se(!1))),d&&d!==v&&(g=$(()=>{if(l=w?l:document.createElementNS(T,d),ht(l,l),r){w&&vt(d)&&l.append(document.createComment(""));var u=w?ke(l):l.appendChild(ye());w&&(u===null?U(!1):W(u)),r(l,u)}$e.nodes_end=l,y.before(l)})),o=d,o&&(v=o),Se(!0))},ft),n&&(U(!0),W(y))}function Xt(e,t){var s=void 0,r;x(()=>{s!==(s=t())&&(r&&(R(r),r=null),s&&(r=$(()=>{be(()=>s(e))})))})}function He(e){var t,s,r="";if(typeof e=="string"||typeof e=="number")r+=e;else if(typeof e=="object")if(Array.isArray(e)){var a=e.length;for(t=0;t<a;t++)e[t]&&(s=He(e[t]))&&(r&&(r+=" "),r+=s)}else for(s in e)e[s]&&(r&&(r+=" "),r+=s);return r}function Zt(){for(var e,t,s=0,r="",a=arguments.length;s<a;s++)(e=arguments[s])&&(t=He(e))&&(r&&(r+=" "),r+=t);return r}function xt(e){return typeof e=="object"?Zt(e):e??""}const Ne=[...` 	
\r\f \v\uFEFF`];function Qt(e,t,s){var r=e==null?"":""+e;if(s){for(var a in s)if(s[a])r=r?r+" "+a:a;else if(r.length)for(var i=a.length,n=0;(n=r.indexOf(a,n))>=0;){var o=n+i;(n===0||Ne.includes(r[n-1]))&&(o===r.length||Ne.includes(r[o]))?r=(n===0?"":r.substring(0,n))+r.substring(o+1):n=o}}return r===""?null:r}function Me(e,t=!1){var s=t?" !important;":";",r="";for(var a in e){var i=e[a];i!=null&&i!==""&&(r+=" "+a+": "+i+s)}return r}function de(e){return e[0]!=="-"||e[1]!=="-"?e.toLowerCase():e}function es(e,t){if(t){var s="",r,a;if(Array.isArray(t)?(r=t[0],a=t[1]):r=t,e){e=String(e).replaceAll(/\s*\/\*.*?\*\/\s*/g,"").trim();var i=!1,n=0,o=!1,v=[];r&&v.push(...Object.keys(r).map(de)),a&&v.push(...Object.keys(a).map(de));var l=0,y=-1;const h=e.length;for(var g=0;g<h;g++){var d=e[g];if(o?d==="/"&&e[g-1]==="*"&&(o=!1):i?i===d&&(i=!1):d==="/"&&e[g+1]==="*"?o=!0:d==='"'||d==="'"?i=d:d==="("?n++:d===")"&&n--,!o&&i===!1&&n===0){if(d===":"&&y===-1)y=g;else if(d===";"||g===h-1){if(y!==-1){var T=de(e.substring(l,y).trim());if(!v.includes(T)){d!==";"&&g++;var u=e.substring(l,g).trim();s+=" "+u+";"}}l=g+1,y=-1}}}}return r&&(s+=Me(r)),a&&(s+=Me(a,!0)),s=s.trim(),s===""?null:s}return e==null?null:String(e)}function ts(e,t,s,r,a,i){var n=e.__className;if(w||n!==s||n===void 0){var o=Qt(s,r,i);(!w||o!==e.getAttribute("class"))&&(o==null?e.removeAttribute("class"):t?e.className=o:e.setAttribute("class",o)),e.__className=s}else if(i&&a!==i)for(var v in i){var l=!!i[v];(a==null||l!==!!a[v])&&e.classList.toggle(v,l)}return i}function he(e,t={},s,r){for(var a in s){var i=s[a];t[a]!==i&&(s[a]==null?e.style.removeProperty(a):e.style.setProperty(a,i,r))}}function ss(e,t,s,r){var a=e.__style;if(w||a!==t){var i=es(t,r);(!w||i!==e.getAttribute("style"))&&(i==null?e.removeAttribute("style"):e.style.cssText=i),e.__style=t}else r&&(Array.isArray(r)?(he(e,s?.[0],r[0]),he(e,s?.[1],r[1],"important")):he(e,s,r));return r}function X(e,t,s=!1){if(e.multiple){if(t==null)return;if(!Ue(t))return gt();for(var r of e.options)r.selected=t.includes(Y(r));return}for(r of e.options){var a=Y(r);if(pt(a,t)){r.selected=!0;return}}(!s||t!==void 0)&&(e.selectedIndex=-1)}function Ve(e){var t=new MutationObserver(()=>{X(e,e.__value)});t.observe(e,{childList:!0,subtree:!0,attributes:!0,attributeFilter:["value"]}),_t(()=>{t.disconnect()})}function ds(e,t,s=t){var r=!0;Re(e,"change",a=>{var i=a?"[selected]":":checked",n;if(e.multiple)n=[].map.call(e.querySelectorAll(i),Y);else{var o=e.querySelector(i)??e.querySelector("option:not([disabled])");n=o&&Y(o)}s(n)}),be(()=>{var a=t();if(X(e,a,r),r&&a===void 0){var i=e.querySelector(":checked");i!==null&&(a=Y(i),s(a))}e.__value=a,r=!1}),Ve(e)}function Y(e){return"__value"in e?e.__value:e.value}const V=Symbol("class"),J=Symbol("style"),Je=Symbol("is custom element"),We=Symbol("is html");function hs(e){if(w){var t=!1,s=()=>{if(!t){if(t=!0,e.hasAttribute("value")){var r=e.value;Z(e,"value",null),e.value=r}if(e.hasAttribute("checked")){var a=e.checked;Z(e,"checked",null),e.checked=a}}};e.__on_r=s,Pt(s),Lt()}}function vs(e,t){var s=we(e);s.checked!==(s.checked=t??void 0)&&(e.checked=t)}function rs(e,t){t?e.hasAttribute("selected")||e.setAttribute("selected",""):e.removeAttribute("selected")}function Z(e,t,s,r){var a=we(e);w&&(a[t]=e.getAttribute(t),t==="src"||t==="srcset"||t==="href"&&e.nodeName==="LINK")||a[t]!==(a[t]=s)&&(t==="loading"&&(e[mt]=s),s==null?e.removeAttribute(t):typeof s!="string"&&Ye(e).includes(t)?e[t]=s:e.setAttribute(t,s))}function as(e,t,s,r,a=!1){var i=we(e),n=i[Je],o=!i[We];let v=w&&n;v&&U(!1);var l=t||{},y=e.tagName==="OPTION";for(var g in t)g in s||(s[g]=null);s.class?s.class=xt(s.class):s[V]&&(s.class=null),s[J]&&(s.style??=null);var d=Ye(e);for(const p in s){let b=s[p];if(y&&p==="value"&&b==null){e.value=e.__value="",l[p]=b;continue}if(p==="class"){var T=e.namespaceURI==="http://www.w3.org/1999/xhtml";ts(e,T,b,r,t?.[V],s[V]),l[p]=b,l[V]=s[V];continue}if(p==="style"){ss(e,b,t?.[J],s[J]),l[p]=b,l[J]=s[J];continue}var u=l[p];if(!(b===u&&!(b===void 0&&e.hasAttribute(p)))){l[p]=b;var h=p[0]+p[1];if(h!=="$$")if(h==="on"){const I={},c="$$"+p;let f=p.slice(2);var m=Ot(f);if(Et(f)&&(f=f.slice(0,-7),I.capture=!0),!m&&u){if(b!=null)continue;e.removeEventListener(f,l[c],I),l[c]=null}if(b!=null)if(m)e[`__${f}`]=b,Ct([f]);else{let A=function(L){l[p].call(this,L)};l[c]=At(f,e,A,I)}else m&&(e[`__${f}`]=void 0)}else if(p==="style")Z(e,p,b);else if(p==="autofocus")St(e,!!b);else if(!n&&(p==="__value"||p==="value"&&b!=null))e.value=e.__value=b;else if(p==="selected"&&y)rs(e,b);else{var _=p;o||(_=Nt(_));var E=_==="defaultValue"||_==="defaultChecked";if(b==null&&!n&&!E)if(i[p]=null,_==="value"||_==="checked"){let I=e;const c=t===void 0;if(_==="value"){let f=I.defaultValue;I.removeAttribute(_),I.defaultValue=f,I.value=I.__value=c?f:null}else{let f=I.defaultChecked;I.removeAttribute(_),I.defaultChecked=f,I.checked=c?f:!1}}else e.removeAttribute(p);else E||d.includes(_)&&(n||typeof b!="string")?(e[_]=b,_ in i&&(i[_]=Mt)):typeof b!="function"&&Z(e,_,b)}}}return v&&U(!0),l}function Oe(e,t,s=[],r=[],a,i=!1){yt(s,r,n=>{var o=void 0,v={},l=e.nodeName==="SELECT",y=!1;if(x(()=>{var d=t(...n.map(C)),T=as(e,o,d,a,i);y&&l&&"value"in d&&X(e,d.value);for(let h of Object.getOwnPropertySymbols(v))d[h]||R(v[h]);for(let h of Object.getOwnPropertySymbols(d)){var u=d[h];h.description===wt&&(!o||u!==o[h])&&(v[h]&&R(v[h]),v[h]=$(()=>Xt(e,()=>u))),T[h]=u}o=T}),l){var g=e;be(()=>{X(g,o.value,!0),Ve(g)})}y=!0})}function we(e){return e.__attributes??={[Je]:e.nodeName.includes("-"),[We]:e.namespaceURI===bt}}var Pe=new Map;function Ye(e){var t=e.getAttribute("is")||e.nodeName,s=Pe.get(t);if(s)return s;Pe.set(t,s=[]);for(var r,a=e,i=Element.prototype;i!==a;){r=Tt(a);for(var n in r)r[n].set&&s.push(n);a=It(a)}return s}function gs(e,t,s=t){var r=new WeakSet;Re(e,"input",async a=>{var i=a?e.defaultValue:e.value;if(i=ve(e)?ge(i):i,s(i),D!==null&&r.add(D),await kt(),i!==(i=t())){var n=e.selectionStart,o=e.selectionEnd;e.value=i??"",o!==null&&(e.selectionStart=n,e.selectionEnd=Math.min(o,e.value.length))}}),(w&&e.defaultValue!==e.value||Ft(t)==null&&e.value)&&(s(ve(e)?ge(e.value):e.value),D!==null&&r.add(D)),Ut(()=>{var a=t();if(e===document.activeElement){var i=Dt??D;if(r.has(i))return}ve(e)&&a===ge(e.value)||e.type==="date"&&!a&&!e.value||a!==e.value&&(e.value=a??"")})}function ve(e){var t=e.type;return t==="number"||t==="range"}function ge(e){return e===""?null:+e}/**
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
 */const is={xmlns:"http://www.w3.org/2000/svg",width:24,height:24,viewBox:"0 0 24 24",fill:"none",stroke:"currentColor","stroke-width":2,"stroke-linecap":"round","stroke-linejoin":"round"};var ns=$t("<svg><!><!></svg>");function Te(e,t){te(t,!0);const s=H(t,"color",3,"currentColor"),r=H(t,"size",3,24),a=H(t,"strokeWidth",3,2),i=H(t,"absoluteStrokeWidth",3,!1),n=H(t,"iconNode",19,()=>[]),o=oe(t,["$$slots","$$events","$$legacy","name","color","size","strokeWidth","absoluteStrokeWidth","iconNode","children"]);var v=ns();Oe(v,g=>({...is,...o,width:r(),height:r(),stroke:s(),"stroke-width":g,class:["lucide-icon lucide",t.name&&`lucide-${t.name}`,t.class]}),[()=>i()?Number(a())*24/Number(r()):a()]);var l=Rt(v);Kt(l,17,n,zt,(g,d)=>{var T=Vt(()=>Jt(C(d),2));let u=()=>C(T)[0],h=()=>C(T)[1];var m=ie(),_=ne(m);Gt(_,u,!0,(E,p)=>{Oe(E,()=>({...h()}))}),z(g,m)});var y=jt(l);se(y,()=>t.children??ae),Ht(v),z(e,v),re()}function ps(e,t){te(t,!0);/**
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
 */let s=oe(t,["$$slots","$$events","$$legacy"]);const r=[["rect",{width:"14",height:"14",x:"8",y:"8",rx:"2",ry:"2"}],["path",{d:"M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"}]];Te(e,Ie({name:"copy"},()=>s,{get iconNode(){return r},children:(a,i)=>{var n=ie(),o=ne(n);se(o,()=>t.children??ae),z(a,n)},$$slots:{default:!0}})),re()}function _s(e,t){te(t,!0);/**
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
 */let s=oe(t,["$$slots","$$events","$$legacy"]);const r=[["path",{d:"m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"}],["path",{d:"M12 9v4"}],["path",{d:"M12 17h.01"}]];Te(e,Ie({name:"triangle-alert"},()=>s,{get iconNode(){return r},children:(a,i)=>{var n=ie(),o=ne(n);se(o,()=>t.children??ae),z(a,n)},$$slots:{default:!0}})),re()}function ys(e,t){te(t,!0);/**
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
 */let s=oe(t,["$$slots","$$events","$$legacy"]);const r=[["path",{d:"M18 6 6 18"}],["path",{d:"m6 6 12 12"}]];Te(e,Ie({name:"x"},()=>s,{get iconNode(){return r},children:(a,i)=>{var n=ie(),o=ne(n);se(o,()=>t.children??ae),z(a,n)},$$slots:{default:!0}})),re()}const ze=Symbol("notifications");function ms(e){Wt(ze,e)}function Be(){return Yt(ze)}class os{baseUrl;fetcher;constructor(t="",s){this.baseUrl=t,this.fetcher=s?.fetcher||fetch}async reply(t,s,r){const a=await this.fetcher(`${this.baseUrl}/mcp/ui`,{method:"POST",headers:{"Content-Type":"application/json",...r?.sessionId&&{"Mcp-Session-Id":r.sessionId}},body:JSON.stringify({jsonrpc:"2.0",id:t,result:s})});if(!(a.status==204||a.status==202)){if(!a.ok){const i=await a.text();throw G(`response: ${a.status}: ${a.statusText}: ${i}`),new Error(i)}try{const i=await a.json();i.error?.message&&G(i.error.message)}catch(i){console.debug("Error parsing JSON:",i)}}}async exchange(t,s,r){const a=await this.fetcher(`${this.baseUrl}/mcp/ui`,{method:"POST",headers:{"Content-Type":"application/json",...r?.sessionId&&{"Mcp-Session-Id":r.sessionId}},body:JSON.stringify({id:crypto.randomUUID(),jsonrpc:"2.0",method:t,params:s})});let i=null;try{const n=await a.json();if(n.error?.message)G(n.error.message),i=new Error(n.error.message);else return n.result}catch(n){G(n),i=n}throw i}async callMCPTool(t,s){const a=await(await this.fetcher(`${this.baseUrl}/mcp/ui`,{method:"POST",headers:{"Content-Type":"application/json",...s?.sessionId&&{"Mcp-Session-Id":s.sessionId}},signal:s?.abort?.signal,body:JSON.stringify({id:crypto.randomUUID(),jsonrpc:"2.0",method:"tools/call",params:{name:t,arguments:s?.payload||{},...s?.async&&{_meta:{"ai.nanobot.async":!0,progressToken:s?.progressToken}}}})})).json();if(a.error?.message){try{Be().error("API Error",a.error.message)}catch{console.error("MCP Tool Error:",a.error.message)}throw new Error(a.error.message)}return s?.parseResponse?s.parseResponse(a.result):a.result?.structuredContent?a.result.structuredContent:{}}async deleteThread(t){await this.callMCPTool("delete_chat",{payload:{chatId:t}})}async renameThread(t,s){return await this.callMCPTool("update_chat",{payload:{chatId:t,title:s}})}async listAgents(t){return await this.callMCPTool("list_agents",t)}async getThreads(){return(await this.callMCPTool("list_chats")).chats}async createThread(){return await this.callMCPTool("create_chat")}async createResource(t,s,r,a){return await this.callMCPTool("create_resource",{payload:{blob:r,mimeType:s,name:t},sessionId:a?.sessionId,abort:a?.abort,parseResponse:i=>i.content?.[0]?.type==="resource_link"?{uri:i.content[0].uri}:{uri:""}})}async sendMessage(t){return await this.callMCPTool("chat_ui",{payload:{prompt:t.message,attachments:t.attachments?.map(r=>({url:r.uri,mimeType:r.mimeType}))},sessionId:t.threadId,progressToken:t.id,async:!0}),{message:{id:t.id,role:"user",created:Ke(),items:[{id:t.id+"_0",type:"text",text:t.message}]}}}subscribe(t,s,r){const a=new EventSource(`${this.baseUrl}/api/events/${t}`);a.onmessage=i=>{const n=JSON.parse(i.data);s({type:"message",message:n})};for(const i of r?.events??[])a.addEventListener(i,n=>{const o=parseInt(n.lastEventId);s({id:o||n.lastEventId,type:i,data:JSON.parse(n.data)})});return a.onerror=i=>{s({type:"error",error:String(i)}),console.error("EventSource failed:",i),a.close()},a.onopen=()=>{console.log("EventSource connected for thread:",t)},()=>a.close()}}function q(e,t){let s=!1;return t.id&&(e=e.map(r=>r.id===t.id?(s=!0,t):r)),s||(e=[...e,t]),e}const ls=new os;class bs{#e;get messages(){return C(this.#e)}set messages(t){N(this.#e,t,!0)}#t;get history(){return C(this.#t)}set history(t){N(this.#t,t,!0)}#s;get isLoading(){return C(this.#s)}set isLoading(t){N(this.#s,t,!0)}#r;get elicitations(){return C(this.#r)}set elicitations(t){N(this.#r,t,!0)}#a;get prompts(){return C(this.#a)}set prompts(t){N(this.#a,t,!0)}#i;get resources(){return C(this.#i)}set resources(t){N(this.#i,t,!0)}#n;get chatId(){return C(this.#n)}set chatId(t){N(this.#n,t,!0)}#o;get agent(){return C(this.#o)}set agent(t){N(this.#o,t,!0)}#l;get uploadedFiles(){return C(this.#l)}set uploadedFiles(t){N(this.#l,t,!0)}#u;get uploadingFiles(){return C(this.#u)}set uploadingFiles(t){N(this.#u,t,!0)}api;closer=()=>{};onChatDone=[];constructor(t){this.api=t?.api||ls,this.#e=M(F([])),this.#t=M(),this.#s=M(!1),this.#r=M(F([])),this.#a=M(F([])),this.#i=M(F([])),this.#n=M(""),this.#o=M(F({})),this.#l=M(F([])),this.#u=M(F([])),this.setChatId(t?.chatId)}close=()=>{this.closer(),this.setChatId("")};setChatId=async t=>{t!==this.chatId&&(this.messages=[],this.prompts=[],this.elicitations=[],this.history=void 0,this.isLoading=!1,this.uploadedFiles=[],this.uploadingFiles=[],t&&(this.chatId=t,this.subscribe(t)),this.listResources().then(s=>{s&&s.resources&&(this.resources=s.resources)}),this.listPrompts().then(s=>{s&&s.prompts&&(this.prompts=s.prompts)}),await this.reloadAgent())};reloadAgent=async()=>{const t=await this.api.listAgents({sessionId:this.chatId});t.agents?.length>0&&(this.agent=t.agents[0])};listPrompts=async()=>await this.api.exchange("prompts/list",{},{sessionId:this.chatId});listResources=async()=>await this.api.exchange("resources/list",{},{sessionId:this.chatId});subscribe(t){this.closer(),t&&(this.closer=this.api.subscribe(t,s=>{if(s.type=="message"&&s.message?.id)this.history?this.history=q(this.history,s.message):this.messages=q(this.messages,s.message);else if(s.type=="history-start")this.history=[];else if(s.type=="history-end")this.messages=this.history||[],this.history=void 0;else if(s.type=="chat-in-progress")this.isLoading=!0;else if(s.type=="chat-done"){this.isLoading=!1;for(const r of this.onChatDone)r();this.onChatDone=[]}else s.type=="elicitation/create"&&(this.elicitations=[...this.elicitations,{id:s.id,...s.data}]);console.debug("Received event:",s)},{events:["history-start","history-end","chat-in-progress","chat-done","elicitation/create"]}))}replyToElicitation=async(t,s)=>{await this.api.reply(t.id,s,{sessionId:this.chatId}),this.elicitations=this.elicitations.filter(r=>r.id!==t.id)};newChat=async()=>{const t=await this.api.createThread();await this.setChatId(t.id)};sendMessage=async(t,s)=>{if(!(!t.trim()||this.isLoading)){this.isLoading=!0,this.chatId||await this.newChat();try{const r=await this.api.sendMessage({id:crypto.randomUUID(),threadId:this.chatId,message:t,attachments:[...this.uploadedFiles,...s||[]]});return this.uploadedFiles=[],this.messages=q(this.messages,r.message),new Promise(a=>{this.onChatDone.push(()=>{this.isLoading=!1;const i=this.messages.findIndex(n=>n.id===r.message.id);i!==-1&&i<=this.messages.length?a({message:this.messages[i+1]}):a()})})}catch(r){this.isLoading=!1,this.messages=q(this.messages,{id:crypto.randomUUID(),role:"assistant",created:Ke(),items:[{id:crypto.randomUUID(),type:"text",text:`Sorry, I couldn't send your message. Please try again. Error: ${r}`}]})}}};cancelUpload=t=>{this.uploadingFiles=this.uploadingFiles.filter(s=>s.id!==t?!0:(s.controller&&s.controller.abort(),!1)),this.uploadedFiles=this.uploadedFiles.filter(s=>s.id!==t)};uploadFile=async(t,s)=>{if(!this.chatId){const i=await this.api.createThread();await this.setChatId(i.id)}const r=crypto.randomUUID(),a=s?.controller||new AbortController;this.uploadingFiles.push({file:t,id:r,controller:a});try{const i=await this.doUploadFile(t,a);return this.uploadedFiles.push({file:t,uri:i.uri,id:r,mimeType:i.mimeType}),i}finally{this.uploadingFiles=this.uploadingFiles.filter(i=>i.id!==r)}};doUploadFile=async(t,s)=>{const r=new FileReader;r.readAsDataURL(t),await new Promise((i,n)=>{r.onloadend=i,r.onerror=n});const a=r.result.split(",")[1];if(!this.chatId)throw new Error("Chat ID not set");return await this.api.createResource(t.name,t.type,a,{description:t.name,sessionId:this.chatId,abort:s})}}function Ke(){return new Date().toISOString()}function G(e){try{Be().error("API Error",e?.toString())}catch{console.error("MCP Tool Error:",e)}console.error("Error:",e)}export{bs as C,Te as I,_s as T,ys as X,ps as a,gs as b,ss as c,ms as d,Kt as e,ls as f,Be as g,Z as h,zt as i,xt as j,ds as k,vs as l,hs as r,ts as s};
