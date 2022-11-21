"use strict";(self.webpackChunkkusk_gateway_docs_2=self.webpackChunkkusk_gateway_docs_2||[]).push([[8099],{3905:(e,t,n)=>{n.d(t,{Zo:()=>c,kt:()=>d});var r=n(67294);function a(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function s(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,r)}return n}function i(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?s(Object(n),!0).forEach((function(t){a(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):s(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function o(e,t){if(null==e)return{};var n,r,a=function(e,t){if(null==e)return{};var n,r,a={},s=Object.keys(e);for(r=0;r<s.length;r++)n=s[r],t.indexOf(n)>=0||(a[n]=e[n]);return a}(e,t);if(Object.getOwnPropertySymbols){var s=Object.getOwnPropertySymbols(e);for(r=0;r<s.length;r++)n=s[r],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(a[n]=e[n])}return a}var u=r.createContext({}),l=function(e){var t=r.useContext(u),n=t;return e&&(n="function"==typeof e?e(t):i(i({},t),e)),n},c=function(e){var t=l(e.components);return r.createElement(u.Provider,{value:t},e.children)},p={inlineCode:"code",wrapper:function(e){var t=e.children;return r.createElement(r.Fragment,{},t)}},g=r.forwardRef((function(e,t){var n=e.components,a=e.mdxType,s=e.originalType,u=e.parentName,c=o(e,["components","mdxType","originalType","parentName"]),g=l(n),d=a,k=g["".concat(u,".").concat(d)]||g[d]||p[d]||s;return n?r.createElement(k,i(i({ref:t},c),{},{components:n})):r.createElement(k,i({ref:t},c))}));function d(e,t){var n=arguments,a=t&&t.mdxType;if("string"==typeof e||a){var s=n.length,i=new Array(s);i[0]=g;var o={};for(var u in t)hasOwnProperty.call(t,u)&&(o[u]=t[u]);o.originalType=e,o.mdxType="string"==typeof e?e:a,i[1]=o;for(var l=2;l<s;l++)i[l]=n[l];return r.createElement.apply(null,i)}return r.createElement.apply(null,n)}g.displayName="MDXCreateElement"},33751:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>u,contentTitle:()=>i,default:()=>p,frontMatter:()=>s,metadata:()=>o,toc:()=>l});var r=n(87462),a=(n(67294),n(3905));const s={},i="2. Launch a Kubernetes Cluster",o={unversionedId:"getting-started/launch-a-kubernetes-cluster",id:"getting-started/launch-a-kubernetes-cluster",title:"2. Launch a Kubernetes Cluster",description:"Kusk needs to be installed in a Kubernetes cluster to serve its traffic.",source:"@site/docs/getting-started/launch-a-kubernetes-cluster.md",sourceDirName:"getting-started",slug:"/getting-started/launch-a-kubernetes-cluster",permalink:"/getting-started/launch-a-kubernetes-cluster",draft:!1,editUrl:"https://github.com/kubeshop/kusk-gateway/docs/getting-started/launch-a-kubernetes-cluster.md",tags:[],version:"current",frontMatter:{},sidebar:"tutorialSidebar",previous:{title:"1. Install Kusk CLI",permalink:"/getting-started/install-kusk-cli"},next:{title:"3. Install Kusk Gateway",permalink:"/getting-started/install-kusk-gateway"}},u={},l=[],c={toc:l};function p(e){let{components:t,...n}=e;return(0,a.kt)("wrapper",(0,r.Z)({},c,n,{components:t,mdxType:"MDXLayout"}),(0,a.kt)("h1",{id:"2-launch-a-kubernetes-cluster"},"2. Launch a Kubernetes Cluster"),(0,a.kt)("p",null,"Kusk needs to be installed in a Kubernetes cluster to serve its traffic."),(0,a.kt)("p",null,"You can start a local Kubernetes cluster or connect to a remote cluster. In this tutorial, you'll find instructions to start a local cluster using ",(0,a.kt)("a",{parentName:"p",href:"https://minikube.sigs.k8s.io/docs/"},"Minikube"),", which will help you get started with Kusk."),(0,a.kt)("p",null,"For more information on the different options for running a Kubernetes cluster locally or remotely, ",(0,a.kt)("a",{parentName:"p",href:"https://docs.tilt.dev/choosing_clusters.html"},"check this great resource")," which contains a vast comparison list."),(0,a.kt)("p",null,"Install Minikube"),(0,a.kt)("p",null,"Use the installation guide from Minikube ",(0,a.kt)("a",{parentName:"p",href:"https://minikube.sigs.k8s.io/docs/start/"},"here"),". "),(0,a.kt)("p",null,"Start your Minikube cluster"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-sh"},"minikube start\n")),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-sh",metastring:'title="Expected output:"',title:'"Expected','output:"':!0},'\ud83d\ude04  minikube v1.28.0 on Darwin 13.0 (arm64)\n\u2728  Automatically selected the docker driver\n\ud83d\udccc  Using Docker Desktop driver with root privileges\n\ud83d\udc4d  Starting control plane node minikube in cluster minikube\n\ud83d\ude9c  Pulling base image ...\n    > gcr.io/k8s-minikube/kicbase:  0 B [_______________________] ?% ? p/s 1m5s\n\ud83d\udd25  Creating docker container (CPUs=2, Memory=7802MB) ...\n\ud83d\udc33  Preparing Kubernetes v1.25.3 on Docker 20.10.20 ...\n    \u25aa Generating certificates and keys ...\n    \u25aa Booting up control plane ...\n    \u25aa Configuring RBAC rules ...\n\ud83d\udd0e  Verifying Kubernetes components...\n    \u25aa Using image gcr.io/k8s-minikube/storage-provisioner:v5\n\ud83c\udf1f  Enabled addons: storage-provisioner, default-storageclass\n\ud83c\udfc4  Done! kubectl is now configured to use "minikube" cluster and "default" namespace by default\n')))}p.isMDXComponent=!0}}]);