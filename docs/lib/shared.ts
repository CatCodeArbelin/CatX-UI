export const appName = 'CatX-UI';
export const appTagline = 'Maintained downstream fork of 3x-ui for operating Xray';

export const docsRoute = '/docs';
export const docsImageRoute = '/og/docs';
export const docsContentRoute = '/llms.mdx/docs';

// The CatX product repository — used for the navbar GitHub link,
// build-time star/release stats, and install commands.
export const productRepo = {
  user: 'CatCodeArbelin',
  repo: 'CatX-UI',
  branch: 'main',
};

// Where these inherited/CatX docs live — used for "Edit on GitHub" links.
export const gitConfig = {
  user: 'CatCodeArbelin',
  repo: 'CatX-UI',
  branch: 'main',
  docsDir: 'docs/content/docs',
};

export const productRepoUrl = `https://github.com/${productRepo.user}/${productRepo.repo}`;

// The repository URL is a safe metadata fallback when no independently
// published CatX documentation origin has been configured.
export const siteUrl = process.env.NEXT_PUBLIC_SITE_URL || productRepoUrl;
