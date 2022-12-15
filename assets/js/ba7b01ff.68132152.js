"use strict";(self.webpackChunkkusk_gateway_docs_2=self.webpackChunkkusk_gateway_docs_2||[]).push([[6193],{3905:(e,t,n)=>{n.d(t,{Zo:()=>u,kt:()=>m});var a=n(67294);function r(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function o(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);t&&(a=a.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,a)}return n}function l(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?o(Object(n),!0).forEach((function(t){r(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):o(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function i(e,t){if(null==e)return{};var n,a,r=function(e,t){if(null==e)return{};var n,a,r={},o=Object.keys(e);for(a=0;a<o.length;a++)n=o[a],t.indexOf(n)>=0||(r[n]=e[n]);return r}(e,t);if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(a=0;a<o.length;a++)n=o[a],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(r[n]=e[n])}return r}var s=a.createContext({}),p=function(e){var t=a.useContext(s),n=t;return e&&(n="function"==typeof e?e(t):l(l({},t),e)),n},u=function(e){var t=p(e.components);return a.createElement(s.Provider,{value:t},e.children)},c={inlineCode:"code",wrapper:function(e){var t=e.children;return a.createElement(a.Fragment,{},t)}},d=a.forwardRef((function(e,t){var n=e.components,r=e.mdxType,o=e.originalType,s=e.parentName,u=i(e,["components","mdxType","originalType","parentName"]),d=p(n),m=r,y=d["".concat(s,".").concat(m)]||d[m]||c[m]||o;return n?a.createElement(y,l(l({ref:t},u),{},{components:n})):a.createElement(y,l({ref:t},u))}));function m(e,t){var n=arguments,r=t&&t.mdxType;if("string"==typeof e||r){var o=n.length,l=new Array(o);l[0]=d;var i={};for(var s in t)hasOwnProperty.call(t,s)&&(i[s]=t[s]);i.originalType=e,i.mdxType="string"==typeof e?e:r,l[1]=i;for(var p=2;p<o;p++)l[p]=n[p];return a.createElement.apply(null,l)}return a.createElement.apply(null,n)}d.displayName="MDXCreateElement"},60406:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>s,contentTitle:()=>l,default:()=>c,frontMatter:()=>o,metadata:()=>i,toc:()=>p});var a=n(87462),r=(n(67294),n(3905));const o={},l="OpenAPI Overlays",i={unversionedId:"guides/overlays",id:"guides/overlays",title:"OpenAPI Overlays",description:"Kusk supports code-first approaches, i.e. OpenAPI generated from code annotations, by use of OpenAPI Overlays.",source:"@site/docs/guides/overlays.md",sourceDirName:"guides",slug:"/guides/overlays",permalink:"/guides/overlays",draft:!1,editUrl:"https://github.com/kubeshop/kusk-gateway/docs/guides/overlays.md",tags:[],version:"current",frontMatter:{},sidebar:"tutorialSidebar",previous:{title:"Rate limiting",permalink:"/guides/rate-limit"},next:{title:"Frontend web applications",permalink:"/guides/web-applications"}},s={},p=[{value:"Example",id:"example",level:2}],u={toc:p};function c(e){let{components:t,...n}=e;return(0,r.kt)("wrapper",(0,a.Z)({},u,n,{components:t,mdxType:"MDXLayout"}),(0,r.kt)("h1",{id:"openapi-overlays"},"OpenAPI Overlays"),(0,r.kt)("p",null,"Kusk supports code-first approaches, i.e. OpenAPI generated from code annotations, by use of OpenAPI Overlays."),(0,r.kt)("p",null,(0,r.kt)("a",{parentName:"p",href:"https://github.com/OAI/Overlay-Specification"},"OpenAPI Overlays")," is a new specification that allows you to have an OpenAPI file without any Kusk extensions, and an overlay file containing the extenions you would want to add to your OpenAPI definition. "),(0,r.kt)("p",null,"After merging the overlay with the OpenAPI file, the resulting file is an OpenAPI definition with all the metadata added to it, ready to be deployed by Kusk."),(0,r.kt)("p",null,"This way, teams can generate their OpenAPI from code, and then add the gateway deployment metadata later."),(0,r.kt)("h2",{id:"example"},"Example"),(0,r.kt)("p",null,"Let's start with an OpenAPI definition that does not have any Kusk extensions added to it."),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-yaml",metastring:'file="openapi.yaml"',file:'"openapi.yaml"'},"openapi: 3.0.0\nservers:\n  - url: http://api.mydomain.com\ninfo:\n  title: simple-api\n  version: 0.1.0\npaths:\n  /hello:\n    get:\n      summary: Returns a Hello world to the user\n      responses:\n        '200':\n          description: A simple hello world!\n          content:\n            application/json; charset=utf-8:\n              schema:\n                type: object\n                properties:\n                  message:\n                    type: string\n                required:\n                  - message\n")),(0,r.kt)("p",null,"Now let's create an overlay that adds Kusk mocking policy to our OpenAPI definition. "),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-yaml",metastring:'file="overlay.yaml"',file:'"overlay.yaml"'},'overlays: 1.0.0\nextends: ./openapi.yaml\nactions:\n  - target: "$"\n    update:\n      mocking:\n        enabled: true\n')),(0,r.kt)("p",null,"To apply the overlay to the OpenAPI definition, run: "),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-sh"},"kusk deploy --overlay overlay.yaml\n")),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-sh",metastring:'title="Expected output"',title:'"Expected','output"':!0},"\ud83c\udf89 successfully parsed\n\u2705 initiallizing deployment to fleet kusk-gateway-envoy-fleet\napi.gateway.kusk.io/simple-api created\n")),(0,r.kt)("p",null,"If you want to look at the generated OpenAPI file before deploying it to Kusk, you can run: "),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-sh"},"kusk generate --overlay overlay.yaml\n")),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-yaml",metastring:'title="Expected output"',title:'"Expected','output"':!0},'apiVersion: gateway.kusk.io/v1alpha1\nkind: API\nmetadata:\n  name: simple-api\n  namespace: default\nspec:\n  fleet:\n    name: kusk-gateway-envoy-fleet\n    namespace: kusk-system\n  spec: |\n    openapi: 3.0.0\n    servers:\n    - url: http://api.mydomain.com\n    components: {}\n    info:\n      title: simple-api\n      version: 0.1.0\n    x-kusk:\n      mocking:\n        enabled: true\n    paths:\n      /hello:\n        get:\n          responses:\n            "200":\n              content:\n                application/json; charset=utf-8:\n                  schema:\n                    properties:\n                      message:\n                        type: string\n                    required:\n                    - message\n                    type: object\n              description: A simple hello world!\n          summary: Returns a Hello world to the user\n')),(0,r.kt)("p",null,"Overlay reference:"),(0,r.kt)("ul",null,(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("strong",{parentName:"li"},"target")," - property is a JSONPath selector (currently proposed by OpenAPI initiative is JMESPath)."),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("strong",{parentName:"li"},"extends")," - defines which API spec to \u201cextend\u201d or which API spec should be overlayed. The value must be in form of full path either as relative or absolute path. For example ",(0,r.kt)("inlineCode",{parentName:"li"},"extends: overlay.yaml")," won't work but ",(0,r.kt)("inlineCode",{parentName:"li"},"extends: ./overlay.yaml")," will"),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("strong",{parentName:"li"},"update")," - property should be a valid YAML that will be placed in the target object"),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("strong",{parentName:"li"},"remove")," - property is a boolean - indicates that the selected target should be removed")))}c.isMDXComponent=!0}}]);