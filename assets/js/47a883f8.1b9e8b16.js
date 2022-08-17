"use strict";(self.webpackChunkkusk_gateway_docs_2=self.webpackChunkkusk_gateway_docs_2||[]).push([[796],{3905:(e,t,n)=>{n.d(t,{Zo:()=>m,kt:()=>c});var a=n(7294);function r(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function l(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);t&&(a=a.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,a)}return n}function o(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?l(Object(n),!0).forEach((function(t){r(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):l(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function s(e,t){if(null==e)return{};var n,a,r=function(e,t){if(null==e)return{};var n,a,r={},l=Object.keys(e);for(a=0;a<l.length;a++)n=l[a],t.indexOf(n)>=0||(r[n]=e[n]);return r}(e,t);if(Object.getOwnPropertySymbols){var l=Object.getOwnPropertySymbols(e);for(a=0;a<l.length;a++)n=l[a],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(r[n]=e[n])}return r}var i=a.createContext({}),p=function(e){var t=a.useContext(i),n=t;return e&&(n="function"==typeof e?e(t):o(o({},t),e)),n},m=function(e){var t=p(e.components);return a.createElement(i.Provider,{value:t},e.children)},d={inlineCode:"code",wrapper:function(e){var t=e.children;return a.createElement(a.Fragment,{},t)}},u=a.forwardRef((function(e,t){var n=e.components,r=e.mdxType,l=e.originalType,i=e.parentName,m=s(e,["components","mdxType","originalType","parentName"]),u=p(n),c=r,k=u["".concat(i,".").concat(c)]||u[c]||d[c]||l;return n?a.createElement(k,o(o({ref:t},m),{},{components:n})):a.createElement(k,o({ref:t},m))}));function c(e,t){var n=arguments,r=t&&t.mdxType;if("string"==typeof e||r){var l=n.length,o=new Array(l);o[0]=u;var s={};for(var i in t)hasOwnProperty.call(t,i)&&(s[i]=t[i]);s.originalType=e,s.mdxType="string"==typeof e?e:r,o[1]=s;for(var p=2;p<l;p++)o[p]=n[p];return a.createElement.apply(null,o)}return a.createElement.apply(null,n)}u.displayName="MDXCreateElement"},146:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>i,contentTitle:()=>o,default:()=>d,frontMatter:()=>l,metadata:()=>s,toc:()=>p});var a=n(7462),r=(n(7294),n(3905));const l={},o="Installing Kusk Gateway with the Kusk CLI",s={unversionedId:"cli/install-cmd",id:"cli/install-cmd",title:"Installing Kusk Gateway with the Kusk CLI",description:"The install command will install Kusk Gateway and all its components with a single command.",source:"@site/docs/cli/install-cmd.md",sourceDirName:"cli",slug:"/cli/install-cmd",permalink:"/cli/install-cmd",draft:!1,editUrl:"https://github.com/kubeshop/kusk-gateway/docs/cli/install-cmd.md",tags:[],version:"current",frontMatter:{},sidebar:"tutorialSidebar",previous:{title:"Kusk CLI",permalink:"/cli/overview"},next:{title:"Mock your API locally using Kusk",permalink:"/cli/mock-cmd"}},i={},p=[{value:"<strong>Kusk Gateway Components</strong>",id:"kusk-gateway-components",level:3},{value:"<strong>Examples</strong>",id:"examples",level:4},{value:"<strong>Arguments</strong>",id:"arguments",level:4},{value:"<strong>Environment Variables</strong>",id:"environment-variables",level:4}],m={toc:p};function d(e){let{components:t,...n}=e;return(0,r.kt)("wrapper",(0,a.Z)({},m,n,{components:t,mdxType:"MDXLayout"}),(0,r.kt)("h1",{id:"installing-kusk-gateway-with-the-kusk-cli"},"Installing Kusk Gateway with the Kusk CLI"),(0,r.kt)("p",null,"The ",(0,r.kt)("inlineCode",{parentName:"p"},"install")," command will install Kusk Gateway and all its components with a single command.\nKusk uses Helm to do this, so you will need to have ",(0,r.kt)("a",{parentName:"p",href:"https://helm.sh/docs/intro/install/"},"Helm installed"),"."),(0,r.kt)("h3",{id:"kusk-gateway-components"},(0,r.kt)("strong",{parentName:"h3"},"Kusk Gateway Components")),(0,r.kt)("ul",null,(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("strong",{parentName:"li"},"Kusk Gateway Manager")," - Responsible for updating and rolling out the Envoy configuration to your Envoy Fleets as you deploy APIs and Static Routes."),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("strong",{parentName:"li"},"Envoy Fleet")," - Responsible for exposing and routing to your APIs and frontends."),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("strong",{parentName:"li"},"Kusk Gateway API")," - REST API, which is exposed by Kusk Gateway and allows you to programmatically query which APIs, Static Routes and Envoy Fleets are deployed."),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("strong",{parentName:"li"},"Kusk Gateway Dashboard")," - A web UI for Kusk Gateway where you can deploy APIs and see which APIs, StaticRoutes and Envoy Fleets are deployed.")),(0,r.kt)("h4",{id:"examples"},(0,r.kt)("strong",{parentName:"h4"},"Examples")),(0,r.kt)("p",null,"The default ",(0,r.kt)("inlineCode",{parentName:"p"},"kusk install")," command will install Kusk Gateway, a public (for your APIs) and private (for the Kusk dashboard and API)\nenvoy-fleet, api, and dashboard in the kusk-system namespace using Helm and using the current kubeconfig context."),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-shell"},"$ kusk install\nadding the kubeshop helm repository\ndone\nfetching the latest charts\ndone\ninstalling Kusk Gateway\ndone\ninstalling Envoy Fleet\ndone\ninstalling Kusk API\ndone\ninstalling Kusk Dashboard\ndone\n\nTo access the dashboard, port forward to the envoy-fleet service that exposes it:\n        $ kubectl port-forward -n            kusk-system svc/kusk-gateway-private-envoy-fleet 8080:80\n\nThen go http://localhost:8080/\n")),(0,r.kt)("p",null,"The following command will create a Helm release named with ",(0,r.kt)("strong",{parentName:"p"},"--name")," in the namespace specified by ",(0,r.kt)("strong",{parentName:"p"},"--namespace"),"."),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-shell"},"$ kusk install --name=my-release --namespace=my-namespace\n...\n")),(0,r.kt)("p",null,"The following command will install Kusk Gateway, but not the dashboard, api, or envoy-fleet."),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-shell"},"$ kusk install --no-dashboard --no-api --no-envoy-fleet\n...\n")),(0,r.kt)("h4",{id:"arguments"},(0,r.kt)("strong",{parentName:"h4"},"Arguments")),(0,r.kt)("table",null,(0,r.kt)("thead",{parentName:"table"},(0,r.kt)("tr",{parentName:"thead"},(0,r.kt)("th",{parentName:"tr",align:"left"},"Flag"),(0,r.kt)("th",{parentName:"tr",align:"left"},"Description"),(0,r.kt)("th",{parentName:"tr",align:"center"},"Required?"))),(0,r.kt)("tbody",{parentName:"table"},(0,r.kt)("tr",{parentName:"tbody"},(0,r.kt)("td",{parentName:"tr",align:"left"},(0,r.kt)("inlineCode",{parentName:"td"},"--name")),(0,r.kt)("td",{parentName:"tr",align:"left"},"The prefix of the name to give to the helm releases for each of the kusk gateway components (default: kusk-gateway)."),(0,r.kt)("td",{parentName:"tr",align:"center"},"\u274c")),(0,r.kt)("tr",{parentName:"tbody"},(0,r.kt)("td",{parentName:"tr",align:"left"},(0,r.kt)("inlineCode",{parentName:"td"},"--namespace"),"/",(0,r.kt)("inlineCode",{parentName:"td"},"-n")),(0,r.kt)("td",{parentName:"tr",align:"left"},"The namespace to install kusk gateway into. Will create the namespace if it doesn't exist (default: kusk-system)."),(0,r.kt)("td",{parentName:"tr",align:"center"},"\u274c")),(0,r.kt)("tr",{parentName:"tbody"},(0,r.kt)("td",{parentName:"tr",align:"left"},(0,r.kt)("inlineCode",{parentName:"td"},"--no-dashboard")),(0,r.kt)("td",{parentName:"tr",align:"left"},"When set, will not install the kusk gateway dashboard."),(0,r.kt)("td",{parentName:"tr",align:"center"},"\u274c")),(0,r.kt)("tr",{parentName:"tbody"},(0,r.kt)("td",{parentName:"tr",align:"left"},(0,r.kt)("inlineCode",{parentName:"td"},"--no-api")),(0,r.kt)("td",{parentName:"tr",align:"left"},"When set, will not install the kusk gateway api. implies --no-dashboard."),(0,r.kt)("td",{parentName:"tr",align:"center"},"\u274c")),(0,r.kt)("tr",{parentName:"tbody"},(0,r.kt)("td",{parentName:"tr",align:"left"},(0,r.kt)("inlineCode",{parentName:"td"},"--no-envoy-fleet")),(0,r.kt)("td",{parentName:"tr",align:"left"},"When set, will not install any envoy fleets."),(0,r.kt)("td",{parentName:"tr",align:"center"},"\u274c")))),(0,r.kt)("h4",{id:"environment-variables"},(0,r.kt)("strong",{parentName:"h4"},"Environment Variables")),(0,r.kt)("p",null,"To disable analytics set following environment variable:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"export ANALYTICS_ENABLED=false\n")),(0,r.kt)("p",null,"or run "),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"ANALYTICS_ENABLED=false kusk install\n")))}d.isMDXComponent=!0}}]);