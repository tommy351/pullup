"use strict";

const organizationName = "tommy351";
const projectName = "pullup";
const githubUrl = `https://github.com/${organizationName}/${projectName}`;
const currentVersion = "1.0";

module.exports = {
  title: "Pullup",
  tagline: "The tagline of my site",
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
          to: `docs/${currentVersion}`,
          activeBasePath: "docs",
          label: "Docs",
          position: "left",
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
          customCss: require.resolve("./src/css/custom.css"),
        },
      },
    ],
  ],
};
