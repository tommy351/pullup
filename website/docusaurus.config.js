"use strict";

const organizationName = "tommy351";
const projectName = "pullup";
const githubUrl = `https://github.com/${organizationName}/${projectName}`;
const currentVersion = "1.0";

module.exports = {
  title: "Pullup",
  tagline: "Update Kubernetes resources by webhooks.",
  url: "https://pullup.dev",
  baseUrl: "/",
  onBrokenLinks: "throw",
  onBrokenMarkdownLinks: "warn",
  favicon: "img/favicon.ico",
  organizationName,
  projectName,
  themeConfig: {
    navbar: {
      title: "Pullup",
      logo: {
        alt: "Pullup Logo",
        src: "img/logo.svg",
      },
      items: [
        {
          type: "doc",
          label: "Docs",
          position: "left",
          docId: "installation",
        },
        {
          type: "docsVersionDropdown",
          position: "right",
          dropdownActiveClassDisabled: true,
        },
        {
          href: githubUrl,
          label: "GitHub",
          position: "right",
        },
      ],
    },
    footer: {
      style: "dark",
      copyright: `Copyright © ${new Date().getFullYear()} Tommy Chen. Built with Docusaurus.`,
    },
    gtag: {
      trackingID: "G-BH4SDZ77Q0",
      anonymizeIP: true,
    },
  },
  presets: [
    [
      "@docusaurus/preset-classic",
      {
        docs: {
          editUrl: githubUrl + "/edit/master/website/",
          lastVersion: "current",
          versions: {
            current: {
              label: currentVersion,
              path: currentVersion,
            },
          },
        },
        blog: {
          showReadingTime: true,
          editUrl: githubUrl + "/edit/master/website/blog/",
        },
        theme: {
          customCss: require.resolve("./src/css/custom.scss"),
        },
      },
    ],
  ],
  plugins: ["docusaurus-plugin-sass"],
};
