/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.
 */

// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  // By default, Docusaurus generates a sidebar from the docs folder structure
  // tutorialSidebar: [{type: 'autogenerated', dirName: '.'}],

  // But you can create a sidebar manually

  tutorialSidebar: [
    {
      type: "category",
      label: "Overview",
      link: {
        type: "doc",
        id: "index",
      },
      items: [
        "overview/when-can-you-use-kusk",
      ],
    },
    {
      type: "category",
      label: "Getting Started",
      items: [
        "getting-started/install-kusk-cli",
        "getting-started/launch-a-kubernetes-cluster",
        "getting-started/install-kusk-gateway",
        "getting-started/mock-an-api",
        "getting-started/connecting-an-application",
      ],
    },
    {
      type: "doc",
      id: "extension",
      label: "Kusk OpenAPI Extension",
    },
    {
      type: "category",
      label: "Guides",
      items: [
        {
          type: "doc",
          id: "guides/working-with-extension",
          label: "Working with the extensions",
        },
        {
          type: "category",
          label: "Authentication",
          items: [
            "guides/authentication/custom-auth-upstream",
            "guides/authentication/oauth2",
            "guides/authentication/cloudentity",
            "guides/authentication/jwt",
          ],
        },
        {
          type: "doc",
          id: "guides/cors",
          label: "CORS",
        },
        "guides/traffic_splitting",
        "guides/mocking",
        "guides/validation",
        "guides/cache",
        "guides/timeouts",
        "guides/routing",
        "guides/rate-limit",
        "guides/overlays",
        "guides/web-applications",
        "guides/cert-manager",
        "guides/troubleshooting",
        "guides/observability",
        {
          type: "category",
          label: "Security",
          items: [
            "guides/security/42crunch",
          ],
        },
      ],

    },
    {
      type: "category",
      label: "Reference",
      items: [
        {
          type: "category",
          label: "Kusk CLI",
          items: [
            "reference/cli/overview",
            "reference/cli/install-cmd",
            "reference/cli/deploy-cmd",
            "reference/cli/mock-cmd",
            "reference/cli/generate-cmd",
            "reference/cli/dashboard-cmd",
          ],
        },
        {
          type: "category",
          label: "Kusk Dashboard",
          items: [
            "reference/dashboard/overview",
            "reference/dashboard/deploying-apis",
            "reference/dashboard/inspecting-apis",
          ],
        },
        {
          type: "link",
          label: "Kusk API playground",
          href: "/docs/reference/kusk-api-server",
        },
        {
          type: "category",
          label: "Kusk Kubernetes Resources",
          items: [
            "reference/customresources/overview",
            "reference/customresources/api",
            "reference/customresources/envoyfleet",
            "reference/customresources/staticroute",
          ],
        },
      ],
    },
    {
      type: "category",
      label: "Quick Links",
      items: [
        "quick-links/install",
        "quick-links/upgrade",
        "quick-links/helm-install",
      ],
    },
    {
      type: "doc",
      label: "Contributing",
      id: "contributing",
    },
    {
      type: "doc",
      id: "privacy",
      label: "Privacy",
    },
  ],
};

module.exports = sidebars;
