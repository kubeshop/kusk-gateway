"use strict";(self.webpackChunkkusk_gateway_docs_2=self.webpackChunkkusk_gateway_docs_2||[]).push([[442],{3905:(e,t,n)=>{n.d(t,{Zo:()=>d,kt:()=>g});var i=n(7294);function a(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function s(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var i=Object.getOwnPropertySymbols(e);t&&(i=i.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,i)}return n}function o(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?s(Object(n),!0).forEach((function(t){a(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):s(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function r(e,t){if(null==e)return{};var n,i,a=function(e,t){if(null==e)return{};var n,i,a={},s=Object.keys(e);for(i=0;i<s.length;i++)n=s[i],t.indexOf(n)>=0||(a[n]=e[n]);return a}(e,t);if(Object.getOwnPropertySymbols){var s=Object.getOwnPropertySymbols(e);for(i=0;i<s.length;i++)n=s[i],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(a[n]=e[n])}return a}var l=i.createContext({}),p=function(e){var t=i.useContext(l),n=t;return e&&(n="function"==typeof e?e(t):o(o({},t),e)),n},d=function(e){var t=p(e.components);return i.createElement(l.Provider,{value:t},e.children)},c={inlineCode:"code",wrapper:function(e){var t=e.children;return i.createElement(i.Fragment,{},t)}},u=i.forwardRef((function(e,t){var n=e.components,a=e.mdxType,s=e.originalType,l=e.parentName,d=r(e,["components","mdxType","originalType","parentName"]),u=p(n),g=a,f=u["".concat(l,".").concat(g)]||u[g]||c[g]||s;return n?i.createElement(f,o(o({ref:t},d),{},{components:n})):i.createElement(f,o({ref:t},d))}));function g(e,t){var n=arguments,a=t&&t.mdxType;if("string"==typeof e||a){var s=n.length,o=new Array(s);o[0]=u;var r={};for(var l in t)hasOwnProperty.call(t,l)&&(r[l]=t[l]);r.originalType=e,r.mdxType="string"==typeof e?e:a,o[1]=r;for(var p=2;p<s;p++)o[p]=n[p];return i.createElement.apply(null,o)}return i.createElement.apply(null,n)}u.displayName="MDXCreateElement"},2619:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>l,contentTitle:()=>o,default:()=>c,frontMatter:()=>s,metadata:()=>r,toc:()=>p});var i=n(7462),a=(n(7294),n(3905));const s={},o="Inspecting Deployed APIs",r={unversionedId:"dashboard/inspecting-apis",id:"dashboard/inspecting-apis",title:"Inspecting Deployed APIs",description:"Selecting a deployed API in the dashboard opens a corresponding details panel to the right containing 3 tabs:",source:"@site/docs/dashboard/inspecting-apis.md",sourceDirName:"dashboard",slug:"/dashboard/inspecting-apis",permalink:"/kusk-gateway/docs/dashboard/inspecting-apis",draft:!1,editUrl:"https://github.com/kubeshop/kusk-gateway/docs/dashboard/inspecting-apis.md",tags:[],version:"current",frontMatter:{},sidebar:"tutorialSidebar",previous:{title:"Deploying APIs",permalink:"/kusk-gateway/docs/dashboard/deploying-apis"},next:{title:"OpenAPI Extension Reference",permalink:"/kusk-gateway/docs/reference/extension"}},l={},p=[{value:"<strong>API Definition Tab</strong>",id:"api-definition-tab",level:2},{value:"<strong>Kusk Extensions Tab</strong>",id:"kusk-extensions-tab",level:2},{value:"<strong>Public API Definition Tab</strong>",id:"public-api-definition-tab",level:2}],d={toc:p};function c(e){let{components:t,...s}=e;return(0,a.kt)("wrapper",(0,i.Z)({},d,s,{components:t,mdxType:"MDXLayout"}),(0,a.kt)("h1",{id:"inspecting-deployed-apis"},"Inspecting Deployed APIs"),(0,a.kt)("p",null,"Selecting a deployed API in the dashboard opens a corresponding details panel to the right containing 3 tabs:"),(0,a.kt)("ul",null,(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("strong",{parentName:"li"},"API Definition"),": Shows an extended Swagger UI for the OpenAPI that was deployed to Kusk Gateway. "),(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("strong",{parentName:"li"},"Kusk Extensions"),": Shows an overview of all ",(0,a.kt)("inlineCode",{parentName:"li"},"x-kusk")," extensions in the deployed OpenAPI definition."),(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("strong",{parentName:"li"},"Public API Definition"),": Shows Swagger UI for the OpenAPI definition that would be exposed to consumers.")),(0,a.kt)("h2",{id:"api-definition-tab"},(0,a.kt)("strong",{parentName:"h2"},"API Definition Tab")),(0,a.kt)("p",null,"The API definition tab shows a Swagger UI for the deployed API definition - together with a table of contents at the\ntop, making it easy to navigate to individual operations."),(0,a.kt)("p",null,(0,a.kt)("img",{alt:"img.png",src:n(1309).Z,width:"1433",height:"1019"})),(0,a.kt)("p",null,"An indicator is shown next to any level in the Table of Contents if there is a ",(0,a.kt)("inlineCode",{parentName:"p"},"x-kusk")," extension defined.\nClicking it will open the corresponding extension in the Kusk Extensions tab (see below)."),(0,a.kt)("p",null,(0,a.kt)("img",{alt:"img.png",src:n(8372).Z,width:"690",height:"376"})),(0,a.kt)("h2",{id:"kusk-extensions-tab"},(0,a.kt)("strong",{parentName:"h2"},"Kusk Extensions Tab")),(0,a.kt)("p",null,"The Kusk Extensions tab contains a tree view showing all ",(0,a.kt)("inlineCode",{parentName:"p"},"x-kusk")," extensions that have been specified in the\ndeployed OpenAPI definition - making it easy to understand how the API has been configured for Kusk Gateway."),(0,a.kt)("p",null,(0,a.kt)("img",{alt:"img_1.png",src:n(8335).Z,width:"1435",height:"700"})),(0,a.kt)("h2",{id:"public-api-definition-tab"},(0,a.kt)("strong",{parentName:"h2"},"Public API Definition Tab")),(0,a.kt)("p",null,'The Public API Definition tab contains the "post-processed" OpenAPI definition as you would provide publicly to consumers of your API. This differs from the deployed OpenAPI definition in the following ways:'),(0,a.kt)("ul",null,(0,a.kt)("li",{parentName:"ul"},"All ",(0,a.kt)("inlineCode",{parentName:"li"},"x-kusk")," extensions have been removed."),(0,a.kt)("li",{parentName:"ul"},"All disabled operations have been removed - see ",(0,a.kt)("a",{parentName:"li",href:"/kusk-gateway/docs/guides/routing#disabling-operations"},"Disabling Operations"),".")),(0,a.kt)("p",null,"A Table of Contents is available as in the API Definition tab. "),(0,a.kt)("p",null,(0,a.kt)("img",{alt:"img_2.png",src:n(8097).Z,width:"1436",height:"1192"})),(0,a.kt)("p",null,"This tab includes the possibility to specify server(s) to be used when executing requests through the integrated Swagger UI:"),(0,a.kt)("p",null,(0,a.kt)("img",{alt:"img_1.png",src:n(5226).Z,width:"742",height:"780"})),(0,a.kt)("p",null,"Specifying the server used by the dashboard itself allows us to execute requests against the Dashboard API. For example,\nto get a list of APIs (as seen in the dashboard), we can execute the ",(0,a.kt)("inlineCode",{parentName:"p"},"GET /apis")," operation."),(0,a.kt)("p",null,(0,a.kt)("img",{alt:"img_2.png",src:n(4066).Z,width:"1417",height:"1448"})))}c.isMDXComponent=!0},1309:(e,t,n)=>{n.d(t,{Z:()=>i});const i=n.p+"assets/images/api-definition-tab-9b3726ae8fb43f610b2ef75eaf46182f.png"},4066:(e,t,n)=>{n.d(t,{Z:()=>i});const i=n.p+"assets/images/executing-requests-f6cc70185cf8a6016b28e0e573033775.png"},8372:(e,t,n)=>{n.d(t,{Z:()=>i});const i=n.p+"assets/images/kusk-extension-icon-58883d1f49b62dbe4226eb91febe97a5.png"},8335:(e,t,n)=>{n.d(t,{Z:()=>i});const i=n.p+"assets/images/kusk-extensions-tab-29a54ee6e087a799f03f2ea9a9f461dd.png"},8097:(e,t,n)=>{n.d(t,{Z:()=>i});const i=n.p+"assets/images/public-api-definition-tab-af50074fd24dc7c60432c27308c49659.png"},5226:(e,t,n)=>{n.d(t,{Z:()=>i});const i=n.p+"assets/images/servers-input-4867266451ab44b52af367863422756e.png"}}]);