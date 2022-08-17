"use strict";(self.webpackChunkkusk_gateway_docs_2=self.webpackChunkkusk_gateway_docs_2||[]).push([[449],{3905:(e,t,n)=>{n.d(t,{Zo:()=>c,kt:()=>h});var a=n(7294);function r(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function o(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);t&&(a=a.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,a)}return n}function i(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?o(Object(n),!0).forEach((function(t){r(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):o(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function l(e,t){if(null==e)return{};var n,a,r=function(e,t){if(null==e)return{};var n,a,r={},o=Object.keys(e);for(a=0;a<o.length;a++)n=o[a],t.indexOf(n)>=0||(r[n]=e[n]);return r}(e,t);if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(a=0;a<o.length;a++)n=o[a],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(r[n]=e[n])}return r}var p=a.createContext({}),s=function(e){var t=a.useContext(p),n=t;return e&&(n="function"==typeof e?e(t):i(i({},t),e)),n},c=function(e){var t=s(e.components);return a.createElement(p.Provider,{value:t},e.children)},d={inlineCode:"code",wrapper:function(e){var t=e.children;return a.createElement(a.Fragment,{},t)}},u=a.forwardRef((function(e,t){var n=e.components,r=e.mdxType,o=e.originalType,p=e.parentName,c=l(e,["components","mdxType","originalType","parentName"]),u=s(n),h=r,m=u["".concat(p,".").concat(h)]||u[h]||d[h]||o;return n?a.createElement(m,i(i({ref:t},c),{},{components:n})):a.createElement(m,i({ref:t},c))}));function h(e,t){var n=arguments,r=t&&t.mdxType;if("string"==typeof e||r){var o=n.length,i=new Array(o);i[0]=u;var l={};for(var p in t)hasOwnProperty.call(t,p)&&(l[p]=t[p]);l.originalType=e,l.mdxType="string"==typeof e?e:r,i[1]=l;for(var s=2;s<o;s++)i[s]=n[s];return a.createElement.apply(null,i)}return a.createElement.apply(null,n)}u.displayName="MDXCreateElement"},2456:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>p,contentTitle:()=>i,default:()=>d,frontMatter:()=>o,metadata:()=>l,toc:()=>s});var a=n(7462),r=(n(7294),n(3905));const o={},i="Connect an upstream service",l={unversionedId:"getting-started/connect-a-service-to-the-api",id:"getting-started/connect-a-service-to-the-api",title:"Connect an upstream service",description:"Once you have created an API and mocked its responses, you are ready to implement the services and connect them to the API.",source:"@site/docs/getting-started/connect-a-service-to-the-api.md",sourceDirName:"getting-started",slug:"/getting-started/connect-a-service-to-the-api",permalink:"/getting-started/connect-a-service-to-the-api",draft:!1,editUrl:"https://github.com/kubeshop/kusk-gateway/docs/getting-started/connect-a-service-to-the-api.md",tags:[],version:"current",frontMatter:{},sidebar:"tutorialSidebar",previous:{title:"Deploy an API",permalink:"/getting-started/deploy-an-api"},next:{title:"The Kusk OpenAPI Extension",permalink:"/guides/working-with-extension"}},p={},s=[{value:"<strong>1. Deploy a Service</strong>",id:"1-deploy-a-service",level:2},{value:"<strong>2. Update the API Manifest to Connect the Service to the Gateway</strong>",id:"2-update-the-api-manifest-to-connect-the-service-to-the-gateway",level:2},{value:"<strong>3. Apply the Changes</strong>",id:"3-apply-the-changes",level:2},{value:"<strong>4. Test the API</strong>",id:"4-test-the-api",level:2},{value:"Next Steps",id:"next-steps",level:2}],c={toc:s};function d(e){let{components:t,...n}=e;return(0,r.kt)("wrapper",(0,a.Z)({},c,n,{components:t,mdxType:"MDXLayout"}),(0,r.kt)("h1",{id:"connect-an-upstream-service"},"Connect an upstream service"),(0,r.kt)("p",null,"Once you have ",(0,r.kt)("a",{parentName:"p",href:"/getting-started/deploy-an-api"},"created an API")," and mocked its responses, you are ready to implement the services and connect them to the API.\nThis section explains how you would connect your services to Kusk-gateway. "),(0,r.kt)("h2",{id:"1-deploy-a-service"},(0,r.kt)("strong",{parentName:"h2"},"1. Deploy a Service")),(0,r.kt)("p",null,"Let's deploy a hello-world Deployment. Create ",(0,r.kt)("inlineCode",{parentName:"p"},"deployment.yaml")," file:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-sh"},'apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: hello-world\nspec:\n  selector:\n    matchLabels:\n      app: hello-world\n  template:\n    metadata:\n      labels:\n        app: hello-world\n    spec:\n      containers:\n      - name: hello-world\n        image: aabedraba/kusk-hello-world:1.0\n        resources:\n          limits:\n            memory: "128Mi"\n            cpu: "500m"\n        ports:\n        - containerPort: 8080\n---\napiVersion: v1\nkind: Service\nmetadata:\n  name: hello-world-svc\nspec:\n  selector:\n    app: hello-world\n  ports:\n  - port: 8080\n    targetPort: 8080\n')),(0,r.kt)("p",null,"And apply it with: "),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-sh"},"kubectl apply -f deployment.yaml\n")),(0,r.kt)("h2",{id:"2-update-the-api-manifest-to-connect-the-service-to-the-gateway"},(0,r.kt)("strong",{parentName:"h2"},"2. Update the API Manifest to Connect the Service to the Gateway")),(0,r.kt)("p",null,"Once you have finished implementing and deploying the service, you will need to stop the mocking of the API endpoint and connect the service to the gateway. "),(0,r.kt)("p",null,"Stop the API mocking by deleting the ",(0,r.kt)("inlineCode",{parentName:"p"},"mocking")," section from the ",(0,r.kt)("inlineCode",{parentName:"p"},"openapi.yaml")," file: "),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-diff"},"...\n- mocking: \n-  enabled: true\n...\n")),(0,r.kt)("p",null,"Add the ",(0,r.kt)("inlineCode",{parentName:"p"},"upstream")," policy to the top of the ",(0,r.kt)("inlineCode",{parentName:"p"},"x-kusk")," section of the ",(0,r.kt)("inlineCode",{parentName:"p"},"openapi.yaml")," file, which contains the details of the service we just created:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-yaml"},"x-kusk:\n upstream:\n  service:\n    name: hello-world-svc\n    namespace: default\n    port: 8080\n")),(0,r.kt)("h2",{id:"3-apply-the-changes"},(0,r.kt)("strong",{parentName:"h2"},"3. Apply the Changes")),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"kubectl apply -f api.yaml\n")),(0,r.kt)("h2",{id:"4-test-the-api"},(0,r.kt)("strong",{parentName:"h2"},"4. Test the API")),(0,r.kt)("p",null,"Get the External IP of Kusk-gateway as indicated in ",(0,r.kt)("a",{parentName:"p",href:"./installation/#get-the-gateways-external-ip"},"installing Kusk-gateway section")," and run:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"$ curl 104.198.194.37/hello\nHello from an implemented service!\n")),(0,r.kt)("p",null,"Now you have successfully deployed an API! "),(0,r.kt)("h2",{id:"next-steps"},"Next Steps"),(0,r.kt)("p",null,'The approach from this "Getting Started" section of the documentation follows a ',(0,r.kt)("a",{parentName:"p",href:"https://kubeshop.io/blog/from-design-first-to-automated-deployment-with-openapi"},"design-first")," approach where you deployed the API first, mocked the API to and later implemented the services and connected them to the API."),(0,r.kt)("p",null,"Check out the ",(0,r.kt)("a",{parentName:"p",href:"/guides/working-with-extension"},"available OpenAPI extensions")," to see all the features that you can enable in your gateway through OpenAPI. And, if you want, connect with us on ",(0,r.kt)("a",{parentName:"p",href:"https://discord.gg/6zupCZFQbe"},"Discord")," to tell us about your experience!"))}d.isMDXComponent=!0}}]);