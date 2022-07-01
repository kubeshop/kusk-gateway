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
      type: "doc",
      id: "index",
      label: "Introduction",
    },
    {
      type: "category",
      label: "Getting Started",
      items: [
        "getting-started/installation",
        "getting-started/deploy-an-api",
        "getting-started/connect-a-service-to-the-api",
      ],
    },
    {
      type: "category",
      label: "Guides",
      items: [
        "guides/working-with-extension",
        {
          type: "doc",
          id: "guides/cors",
          label: "CORS",
        },
        {
          type: "doc",
          id: "guides/mocking",
          label: "Mocking",
        },
        {
          type: "doc",
          id: "guides/validation",
          label: "Validation",
        },
        {
          type: "doc",
          id: "guides/cache",
          label: "Caching",
        },
        {
          type: "doc",
          id: "guides/routing",
          label: "Routing",
        },
        "guides/timeouts",
        "guides/routing",
        "guides/rate-limit",
        "guides/basic-auth",
        "guides/oauth2",
        "guides/cert-manager",
        "guides/troubleshooting",
        "guides/observability",
      ],
    },
    {
      type: "category",
      label: "Kusk CLI",
      items: [
        {
          type: "doc",
          id: "cli/overview",
          label: "Installation",
        },
        "cli/install-cmd",
        "cli/mock-cmd",
        "cli/generate-cmd",
        "cli/dashboard-cmd",
      ],
    },
    {
      type: "category",
      label: "Kusk Dashboard",
      items: [
        "dashboard/overview",
        "dashboard/deploying-apis",
        "dashboard/inspecting-apis",
      ],
    },
    {
      type: "category",
      label: "Reference",
      items: [
        "reference/extension",
        {
          type: "link",
          label: "Dashboard API",
          href: "/docs/reference/kusk-api-server",
        },
        {
          type: "category",
          label: "Kusk Resources",
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
