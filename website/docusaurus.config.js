'use strict';

const organizationName = 'tommy351';
const projectName = 'pullup';
const githubUrl = `https://github.com/${organizationName}/${projectName}`;

module.exports = {
  title: 'Pullup',
  tagline: 'The tagline of my site',
  url: 'https://pullup.dev',
  baseUrl: '/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/favicon.ico',
  organizationName,
  projectName,
  themeConfig: {
    navbar: {
      title: 'Pullup',
      logo: {
        alt: 'Pullup Logo',
        src: 'img/logo.svg',
      },
      items: [
        {
          to: 'docs/',
          activeBasePath: 'docs',
          label: 'Docs',
          position: 'left',
        },
        {to: 'blog', label: 'Blog', position: 'left'},
        {
          href: githubUrl,
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      copyright: `Copyright © ${new Date().getFullYear()} Tommy Chen. Built with Docusaurus.`,
    },
  },
  presets: [
    [
      '@docusaurus/preset-classic',
      {
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl: githubUrl + '/edit/master/website/',
        },
        blog: {
          showReadingTime: true,
          editUrl: githubUrl + '/edit/master/website/blog/',
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      },
    ],
  ],
};
