// See https://docusaurus.io/docs/site-config for all the possible site configuration options.

// List of projects/orgs using your project for the users page.
const users = [];

const siteConfig = {
  title: 'ghz', // Title for your website.
  tagline: 'gRPC benchmarking and load testing tool',
  url: 'https://ghz.sh', // Your website URL
  baseUrl: '/', // Base URL for your project */

  // Used for publishing and more
  projectName: 'ghz',
  organizationName: 'bojand',
  // For top-level user or org sites, the organization is still the same.
  // e.g., for the https://JoelMarcey.github.io site, it would be set like...
  //   organizationName: 'JoelMarcey'

  // For no header links in the top nav bar -> headerLinks: [],
  headerLinks: [
    { doc: 'intro', label: 'CLI'},
    { doc: 'web/intro', label: 'Web'},
    { href: "https://godoc.org/github.com/bojand/ghz", label: "GoDoc" },
    { href: "https://github.com/bojand/ghz", label: "GitHub" }
  ],

  // If you have users set above, you add it here:
  users,

  /* path to images for header/footer */
  headerIcon: 'img/green_fwd2.svg',
  footerIcon: 'img/green_fwd2.svg',
  favicon: 'img/favicon.ico',

  /* Colors for website */
  colors: {
    primaryColor: '#3CBCBC',
    secondaryColor: '#205C3B',
  },

  // This copyright info is used in /core/Footer.js and blog RSS/Atom feeds.
  copyright: `Copyright © ${new Date().getFullYear()} Bojan D.`,

  highlight: {
    // Highlight.js theme to use for syntax highlighting in code blocks.
    theme: 'default',
  },

  // Add custom scripts here that would be placed in <script> tags.
  scripts: ['https://buttons.github.io/buttons.js'],

  // On page navigation for the current documentation page.
  // onPageNav: 'separate',

  // No .html extensions for paths.
  cleanUrl: true,

  // Open Graph and Twitter card images.
  ogImage: 'img/green_fwd2.svg',
  twitterImage: 'img/green_fwd2.svg',

  // Show documentation's last contributor's name.
  // enableUpdateBy: true,

  // Show documentation's last update time.
  // enableUpdateTime: true,

  // You may provide arbitrary config keys to be used as needed by your
  // template. For example, if you need your repo's URL...
  repoUrl: 'https://github.com/bojand/ghz',
};

module.exports = siteConfig;
