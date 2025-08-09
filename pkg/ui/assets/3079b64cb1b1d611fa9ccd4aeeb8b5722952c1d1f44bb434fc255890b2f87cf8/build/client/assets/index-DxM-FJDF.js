import{r as a,j as c}from"./chunk-QMGIS6GS-DfTvyqh6.js";function S(t,e,n){const r=a.useRef(typeof window<"u"&&"BroadcastChannel"in window?new BroadcastChannel(`${t}-channel`):null);return w(r,"message",e),w(r,"messageerror",n),a.useCallback(s=>{var i;(i=r==null?void 0:r.current)==null||i.postMessage(s)},[])}function w(t,e,n=()=>{}){a.useEffect(()=>{const r=t.current;if(r)return r.addEventListener(e,n),()=>r.removeEventListener(e,n)},[e,n])}function x(t){const e=document.createElement("style");e.appendChild(document.createTextNode(`* {
       -webkit-transition: none !important;
       -moz-transition: none !important;
       -o-transition: none !important;
       -ms-transition: none !important;
       transition: none !important;
    }`)),document.head.appendChild(e),t(),setTimeout(()=>{window.getComputedStyle(e).transition,document.head.removeChild(e)},100)}function k({disableTransitions:t=!1}={}){return a.useCallback(e=>{t?x(()=>{e()}):e()},[t])}var C=(t=>(t.DARK="dark",t.LIGHT="light",t))(C||{}),A=Object.values(C),g=a.createContext(void 0);g.displayName="ThemeContext";var p="(prefers-color-scheme: light)",T=()=>window.matchMedia(p).matches?"light":"dark",h=typeof window<"u"?window.matchMedia(p):null;function L({children:t,specifiedTheme:e,themeAction:n,disableTransitionOnThemeChange:r=!1}){const s=k({disableTransitions:r}),[i,l]=a.useState(()=>e?A.includes(e)?e:null:typeof window!="object"?null:T()),[m,u]=a.useState(e?"USER":"SYSTEM"),f=S("remix-themes",o=>{s(()=>{console.log("broadcastThemeChange",r),l(o.data.theme),u(o.data.definedBy)})});a.useEffect(()=>{if(m==="USER")return()=>{};const o=d=>{s(()=>{l(d.matches?"light":"dark")})};return h==null||h.addEventListener("change",o),()=>h==null?void 0:h.removeEventListener("change",o)},[s,m]);const y=a.useCallback(o=>{const d=typeof o=="function"?o(i):o;if(d===null){const v=T();s(()=>{l(v),u("SYSTEM"),f({theme:v,definedBy:"SYSTEM"})}),fetch(`${n}`,{method:"POST",body:JSON.stringify({theme:null})})}else s(()=>{l(d),u("USER")}),f({theme:d,definedBy:"USER"}),fetch(`${n}`,{method:"POST",body:JSON.stringify({theme:d})})},[f,s,i,n]),E=a.useMemo(()=>[i,y,{definedBy:m}],[i,y,m]);return c.jsx(g.Provider,{value:E,children:t})}var M=String.raw`
(() => {
  const theme = window.matchMedia(${JSON.stringify(p)}).matches
    ? 'light'
    : 'dark';
  
  const cl = document.documentElement.classList;
  const dataAttr = document.documentElement.dataset.theme;

  if (dataAttr != null) {
    const themeAlreadyApplied = dataAttr === 'light' || dataAttr === 'dark';
    if (!themeAlreadyApplied) {
      document.documentElement.dataset.theme = theme;
    }
  } else {
    const themeAlreadyApplied = cl.contains('light') || cl.contains('dark');
    if (!themeAlreadyApplied) {
      cl.add(theme);
    }
  }
  
  const meta = document.querySelector('meta[name=color-scheme]');
  if (meta) {
    if (theme === 'dark') {
      meta.content = 'dark light';
    } else if (theme === 'light') {
      meta.content = 'light dark';
    }
  }
})();
`;function j({ssrTheme:t,nonce:e}){const[n]=b();return c.jsxs(c.Fragment,{children:[c.jsx("meta",{name:"color-scheme",content:n==="light"?"light dark":"dark light"}),t?null:c.jsx("script",{dangerouslySetInnerHTML:{__html:M},nonce:e,suppressHydrationWarning:!0})]})}function b(){const t=a.useContext(g);if(t===void 0)throw new Error("useTheme must be used within a ThemeProvider");return t}export{j as P,L as T,C as a,b as u};
