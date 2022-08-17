"use strict";(self.webpackChunkkusk_gateway_docs_2=self.webpackChunkkusk_gateway_docs_2||[]).push([[775],{3905:(e,t,n)=>{n.d(t,{Zo:()=>u,kt:()=>d});var r=n(7294);function a(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function o(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,r)}return n}function l(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?o(Object(n),!0).forEach((function(t){a(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):o(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function i(e,t){if(null==e)return{};var n,r,a=function(e,t){if(null==e)return{};var n,r,a={},o=Object.keys(e);for(r=0;r<o.length;r++)n=o[r],t.indexOf(n)>=0||(a[n]=e[n]);return a}(e,t);if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(r=0;r<o.length;r++)n=o[r],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(a[n]=e[n])}return a}var c=r.createContext({}),s=function(e){var t=r.useContext(c),n=t;return e&&(n="function"==typeof e?e(t):l(l({},t),e)),n},u=function(e){var t=s(e.components);return r.createElement(c.Provider,{value:t},e.children)},p={inlineCode:"code",wrapper:function(e){var t=e.children;return r.createElement(r.Fragment,{},t)}},m=r.forwardRef((function(e,t){var n=e.components,a=e.mdxType,o=e.originalType,c=e.parentName,u=i(e,["components","mdxType","originalType","parentName"]),m=s(n),d=a,f=m["".concat(c,".").concat(d)]||m[d]||p[d]||o;return n?r.createElement(f,l(l({ref:t},u),{},{components:n})):r.createElement(f,l({ref:t},u))}));function d(e,t){var n=arguments,a=t&&t.mdxType;if("string"==typeof e||a){var o=n.length,l=new Array(o);l[0]=m;var i={};for(var c in t)hasOwnProperty.call(t,c)&&(i[c]=t[c]);i.originalType=e,i.mdxType="string"==typeof e?e:a,l[1]=i;for(var s=2;s<o;s++)l[s]=n[s];return r.createElement.apply(null,l)}return r.createElement.apply(null,n)}m.displayName="MDXCreateElement"},5520:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>c,contentTitle:()=>l,default:()=>p,frontMatter:()=>o,metadata:()=>i,toc:()=>s});var r=n(7462),a=(n(7294),n(3905));const o={},l="Privacy Policy",i={unversionedId:"privacy",id:"privacy",title:"Privacy Policy",description:"With the aim to improve the user experience, Kusk collects anonymous usage data.",source:"@site/docs/privacy.md",sourceDirName:".",slug:"/privacy",permalink:"/privacy",draft:!1,editUrl:"https://github.com/kubeshop/kusk-gateway/docs/privacy.md",tags:[],version:"current",frontMatter:{},sidebar:"tutorialSidebar",previous:{title:"How to Develop Kusk Gateway",permalink:"/contributing"}},c={},s=[{value:"<strong>What We Collect</strong>",id:"what-we-collect",level:2},{value:"How to opt out",id:"how-to-opt-out",level:2},{value:"Helm Chart",id:"helm-chart",level:3},{value:"Kusk CLI",id:"kusk-cli",level:3}],u={toc:s};function p(e){let{components:t,...n}=e;return(0,a.kt)("wrapper",(0,r.Z)({},u,n,{components:t,mdxType:"MDXLayout"}),(0,a.kt)("h1",{id:"privacy-policy"},"Privacy Policy"),(0,a.kt)("p",null,"With the aim to improve the user experience, Kusk collects anonymous usage data."),(0,a.kt)("p",null,"You may ",(0,a.kt)("a",{parentName:"p",href:"#how-to-opt-out"},"opt-out")," if you'd prefer not to share any information."),(0,a.kt)("p",null,"The data collected is always anonymous, not traceable to the source, and only used in aggregate form. "),(0,a.kt)("p",null,"Telemetry collects and scrambles information about the host when the API server is bootstrapped for the first time. "),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-json"},'{\n   "anonymousId": "37c7dd3d2f0cd7eca8fdc5b606577278bf2a65e5da42fd4b809cfdf103583a98",\n   "context": {\n     "library": {\n       "name": "analytics-go",\n       "version": "3.0.0"\n     }\n   },\n   "event": "kusk-cli",\n   "integrations": {},\n   "messageId": "c785d086-2d85-4d7a-9468-1da350822c95",\n   "originalTimestamp": "2022-07-15T11:42:41.213006+08:00",\n   "properties": {\n     "event": "dashboard"\n   },\n   "receivedAt": "2022-07-15T03:42:42.691Z",\n   "sentAt": "2022-07-15T03:42:41.215Z",\n   "timestamp": "2022-07-15T03:42:42.689Z",\n   "type": "track",\n   "userId": "37c7dd3d2f0cd7eca8fdc5b606577278bf2a65e5da42fd4b809cfdf103583a98",\n   "writeKey": "1t8VoI1wfqa43n0pYU01VZU2ZVDJKcQh"\n }\n')),(0,a.kt)("h2",{id:"what-we-collect"},(0,a.kt)("strong",{parentName:"h2"},"What We Collect")),(0,a.kt)("p",null,"The telemetry data we use in our metrics is limited to:"),(0,a.kt)("ul",null,(0,a.kt)("li",{parentName:"ul"},"The number of CLI installations."),(0,a.kt)("li",{parentName:"ul"},"The number of unique CLI usages in a day."),(0,a.kt)("li",{parentName:"ul"},"The number of installations to a cluster."),(0,a.kt)("li",{parentName:"ul"},"The number of unique active cluster installations."),(0,a.kt)("li",{parentName:"ul"},"The number of people who disable telemetry."),(0,a.kt)("li",{parentName:"ul"},"The number of unique sessions in the UI."),(0,a.kt)("li",{parentName:"ul"},"The number of API, StaticRoute and EnvoyFleet creations.")),(0,a.kt)("h2",{id:"how-to-opt-out"},"How to opt out"),(0,a.kt)("h3",{id:"helm-chart"},"Helm Chart"),(0,a.kt)("p",null,"To disable sending the anonymous analytics, provide the ",(0,a.kt)("inlineCode",{parentName:"p"},"analytics.enable: false")," override during Helm chart installation or upgrade. See the ",(0,a.kt)("a",{href:"https://github.com/kubeshop/helm-charts/blob/main/charts/kusk-gateway/values.yaml",target:"_blank"},"Helm chart parameters")," for more details about Helm chart configuration."),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},"helm upgrade kusk-gateway kubeshop/kusk-gateway \\\n--install --namespace --create-namespace \\\n--set analytics.enabled=false \\\n...\n")),(0,a.kt)("h3",{id:"kusk-cli"},"Kusk CLI"),(0,a.kt)("p",null,"Set the following environment variable when running kusk commands"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},"export ANALYTICS_ENABLED=false\n")),(0,a.kt)("p",null,"or"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},"ANALYTICS_ENABLED=false kusk install\n")))}p.isMDXComponent=!0}}]);