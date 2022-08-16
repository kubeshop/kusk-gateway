"use strict";(self.webpackChunkkusk_gateway_docs_2=self.webpackChunkkusk_gateway_docs_2||[]).push([[41],{3905:(e,t,n)=>{n.d(t,{Zo:()=>u,kt:()=>d});var r=n(7294);function a(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function o(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,r)}return n}function i(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?o(Object(n),!0).forEach((function(t){a(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):o(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function s(e,t){if(null==e)return{};var n,r,a=function(e,t){if(null==e)return{};var n,r,a={},o=Object.keys(e);for(r=0;r<o.length;r++)n=o[r],t.indexOf(n)>=0||(a[n]=e[n]);return a}(e,t);if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(r=0;r<o.length;r++)n=o[r],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(a[n]=e[n])}return a}var l=r.createContext({}),p=function(e){var t=r.useContext(l),n=t;return e&&(n="function"==typeof e?e(t):i(i({},t),e)),n},u=function(e){var t=p(e.components);return r.createElement(l.Provider,{value:t},e.children)},c={inlineCode:"code",wrapper:function(e){var t=e.children;return r.createElement(r.Fragment,{},t)}},k=r.forwardRef((function(e,t){var n=e.components,a=e.mdxType,o=e.originalType,l=e.parentName,u=s(e,["components","mdxType","originalType","parentName"]),k=p(n),d=a,g=k["".concat(l,".").concat(d)]||k[d]||c[d]||o;return n?r.createElement(g,i(i({ref:t},u),{},{components:n})):r.createElement(g,i({ref:t},u))}));function d(e,t){var n=arguments,a=t&&t.mdxType;if("string"==typeof e||a){var o=n.length,i=new Array(o);i[0]=k;var s={};for(var l in t)hasOwnProperty.call(t,l)&&(s[l]=t[l]);s.originalType=e,s.mdxType="string"==typeof e?e:a,i[1]=s;for(var p=2;p<o;p++)i[p]=n[p];return r.createElement.apply(null,i)}return r.createElement.apply(null,n)}k.displayName="MDXCreateElement"},7250:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>l,contentTitle:()=>i,default:()=>c,frontMatter:()=>o,metadata:()=>s,toc:()=>p});var r=n(7462),a=(n(7294),n(3905));const o={},i="The Kusk OpenAPI Extension",s={unversionedId:"guides/working-with-extension",id:"guides/working-with-extension",title:"The Kusk OpenAPI Extension",description:"Kusk Gateway comes with an x-kusk OpenAPI extension to allow",source:"@site/docs/guides/working-with-extension.md",sourceDirName:"guides",slug:"/guides/working-with-extension",permalink:"/kusk-gateway/docs/guides/working-with-extension",draft:!1,editUrl:"https://github.com/kubeshop/kusk-gateway/docs/guides/working-with-extension.md",tags:[],version:"current",frontMatter:{},sidebar:"tutorialSidebar",previous:{title:"Connect an upstream service",permalink:"/kusk-gateway/docs/getting-started/connect-a-service-to-the-api"},next:{title:"CORS Settings",permalink:"/kusk-gateway/docs/guides/cors"}},l={},p=[{value:"<strong>Properties Overview</strong>",id:"properties-overview",level:2},{value:"<strong>Property Overriding/Inheritance</strong>",id:"property-overridinginheritance",level:2}],u={toc:p};function c(e){let{components:t,...n}=e;return(0,a.kt)("wrapper",(0,r.Z)({},u,n,{components:t,mdxType:"MDXLayout"}),(0,a.kt)("h1",{id:"the-kusk-openapi-extension"},"The Kusk OpenAPI Extension"),(0,a.kt)("p",null,"Kusk Gateway comes with an ",(0,a.kt)("inlineCode",{parentName:"p"},"x-kusk")," ",(0,a.kt)("a",{parentName:"p",href:"https://swagger.io/specification/#specification-extensions"},"OpenAPI extension")," to allow\nan OpenAPI definition to be the source of truth for both operational and functional aspects of your APIs."),(0,a.kt)("p",null,"The ",(0,a.kt)("a",{parentName:"p",href:"../reference/extension"},"extension reference")," describes all available properties and the following guides are\navailable to help you make the most of them:"),(0,a.kt)("ul",null,(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"/kusk-gateway/docs/guides/mocking"},"Mocking")," - How to mock all or parts of your API."),(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"/kusk-gateway/docs/guides/validation"},"Validation")," - How work with automatic request validation."),(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"/kusk-gateway/docs/guides/cors"},"CORS")," - How to specify CORS settings."),(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"/kusk-gateway/docs/guides/routing"},"Routing")," - How to configure routing of API requests."),(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"/kusk-gateway/docs/guides/timeouts"},"Timeouts")," - How to set request timeouts.")),(0,a.kt)("h2",{id:"properties-overview"},(0,a.kt)("strong",{parentName:"h2"},"Properties Overview")),(0,a.kt)("p",null,(0,a.kt)("inlineCode",{parentName:"p"},"x-kusk")," extension can be applied at (not exclusively):"),(0,a.kt)("ol",null,(0,a.kt)("li",{parentName:"ol"},"Top level of an OpenAPI definition:")),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-yaml"},'  openapi: 3.0.2\n  info:\n    title: Swagger Petstore - OpenAPI 3.0\n  x-kusk:\n    hosts:\n    - "example.org"\n    disabled: false\n    cors:\n      ...\n')),(0,a.kt)("ol",{start:2},(0,a.kt)("li",{parentName:"ol"},"Path level:")),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-yaml"},"openapi: 3.0.2\ninfo:\n  title: Swagger Petstore - OpenAPI 3.0\npaths:\n  /pet:\n    x-kusk:\n      disabled: true # disables all /pet endpoints\n    post:\n      ...\n")),(0,a.kt)("ol",{start:3},(0,a.kt)("li",{parentName:"ol"},"Method (operation) level:")),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-yaml"},"  openapi: 3.0.2\n  info:\n    title: Swagger Petstore - OpenAPI 3.0\n  paths:\n    /pet:\n      post:\n        x-kusk:\n          upstream: # routes the POST /pet endpoint to a Kubernetes service\n            service:\n              namespace: default\n              name: petstore\n              port: 8000\n        ...\n")),(0,a.kt)("h2",{id:"property-overridinginheritance"},(0,a.kt)("strong",{parentName:"h2"},"Property Overriding/Inheritance")),(0,a.kt)("p",null,"The ",(0,a.kt)("inlineCode",{parentName:"p"},"x-kusk")," extension at the operation level takes precedence, or overrides, what is specified at the path level, including the ",(0,a.kt)("inlineCode",{parentName:"p"},"disabled")," option.\nLikewise, the path level settings override what is specified at the global level."),(0,a.kt)("p",null,"If settings aren't specified at a path or operation level, they will be inherited from the layer above, (Operation > Path > Global)."))}c.isMDXComponent=!0}}]);