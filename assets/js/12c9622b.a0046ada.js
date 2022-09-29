"use strict";(self.webpackChunkkusk_gateway_docs_2=self.webpackChunkkusk_gateway_docs_2||[]).push([[412],{3905:(e,r,t)=>{t.d(r,{Zo:()=>l,kt:()=>m});var n=t(67294);function o(e,r,t){return r in e?Object.defineProperty(e,r,{value:t,enumerable:!0,configurable:!0,writable:!0}):e[r]=t,e}function a(e,r){var t=Object.keys(e);if(Object.getOwnPropertySymbols){var n=Object.getOwnPropertySymbols(e);r&&(n=n.filter((function(r){return Object.getOwnPropertyDescriptor(e,r).enumerable}))),t.push.apply(t,n)}return t}function s(e){for(var r=1;r<arguments.length;r++){var t=null!=arguments[r]?arguments[r]:{};r%2?a(Object(t),!0).forEach((function(r){o(e,r,t[r])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(t)):a(Object(t)).forEach((function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(t,r))}))}return e}function c(e,r){if(null==e)return{};var t,n,o=function(e,r){if(null==e)return{};var t,n,o={},a=Object.keys(e);for(n=0;n<a.length;n++)t=a[n],r.indexOf(t)>=0||(o[t]=e[t]);return o}(e,r);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);for(n=0;n<a.length;n++)t=a[n],r.indexOf(t)>=0||Object.prototype.propertyIsEnumerable.call(e,t)&&(o[t]=e[t])}return o}var u=n.createContext({}),i=function(e){var r=n.useContext(u),t=r;return e&&(t="function"==typeof e?e(r):s(s({},r),e)),t},l=function(e){var r=i(e.components);return n.createElement(u.Provider,{value:r},e.children)},p={inlineCode:"code",wrapper:function(e){var r=e.children;return n.createElement(n.Fragment,{},r)}},f=n.forwardRef((function(e,r){var t=e.components,o=e.mdxType,a=e.originalType,u=e.parentName,l=c(e,["components","mdxType","originalType","parentName"]),f=i(t),m=o,y=f["".concat(u,".").concat(m)]||f[m]||p[m]||a;return t?n.createElement(y,s(s({ref:r},l),{},{components:t})):n.createElement(y,s({ref:r},l))}));function m(e,r){var t=arguments,o=r&&r.mdxType;if("string"==typeof e||o){var a=t.length,s=new Array(a);s[0]=f;var c={};for(var u in r)hasOwnProperty.call(r,u)&&(c[u]=r[u]);c.originalType=e,c.mdxType="string"==typeof e?e:o,s[1]=c;for(var i=2;i<a;i++)s[i]=t[i];return n.createElement.apply(null,s)}return n.createElement.apply(null,t)}f.displayName="MDXCreateElement"},45637:(e,r,t)=>{t.r(r),t.d(r,{assets:()=>u,contentTitle:()=>s,default:()=>p,frontMatter:()=>a,metadata:()=>c,toc:()=>i});var n=t(87462),o=(t(67294),t(3905));const a={},s="Kusk Custom Resources",c={unversionedId:"reference/customresources/overview",id:"reference/customresources/overview",title:"Kusk Custom Resources",description:"Kusk Gateway defines a number of Kubernetes CRDs for managing its configuration. These are installed as part of the",source:"@site/docs/reference/customresources/overview.md",sourceDirName:"reference/customresources",slug:"/reference/customresources/overview",permalink:"/reference/customresources/overview",draft:!1,editUrl:"https://github.com/kubeshop/kusk-gateway/docs/reference/customresources/overview.md",tags:[],version:"current",frontMatter:{},sidebar:"tutorialSidebar",previous:{title:"Inspecting Deployed APIs",permalink:"/reference/dashboard/inspecting-apis"},next:{title:"API",permalink:"/reference/customresources/api"}},u={},i=[],l={toc:i};function p(e){let{components:r,...t}=e;return(0,o.kt)("wrapper",(0,n.Z)({},l,t,{components:r,mdxType:"MDXLayout"}),(0,o.kt)("h1",{id:"kusk-custom-resources"},"Kusk Custom Resources"),(0,o.kt)("p",null,"Kusk Gateway defines a number of Kubernetes CRDs for managing its configuration. These are installed as part of the\nKusk Gateway installation process."),(0,o.kt)("p",null,"Kusk Gateway uses the following CRDs:"),(0,o.kt)("ul",null,(0,o.kt)("li",{parentName:"ul"},(0,o.kt)("a",{parentName:"li",href:"/reference/customresources/envoyfleet"},"Envoy Fleet")," - For managing Envoy deployments."),(0,o.kt)("li",{parentName:"ul"},(0,o.kt)("a",{parentName:"li",href:"/reference/customresources/api"},"API")," - For using an OpenAPI definition to configure Gateway behaviour."),(0,o.kt)("li",{parentName:"ul"},(0,o.kt)("a",{parentName:"li",href:"/reference/customresources/staticroute"},"Static Route")," - For exposing static content through Kusk Gateway.")))}p.isMDXComponent=!0}}]);