// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import mermaid from 'astro-mermaid';
import { relativeLinks } from './src/relative-links.mjs';

export default defineConfig({
  site: 'https://lucdoe.github.io',
  base: '/opengateway',
  markdown: {
    remarkPlugins: [relativeLinks],
  },
  integrations: [
    mermaid({ autoTheme: true }),
    starlight({
      title: 'OpenGateway',
      description: 'A lightweight API gateway in Go, configured in one YAML file.',
      logo: { src: './src/assets/logo.svg' },
      customCss: ['./src/styles/opengateway.css'],
      components: {
        SiteTitle: './src/components/SiteTitle.astro',
        ThemeSelect: './src/components/ThemeSelect.astro',
      },
      head: [
        { tag: 'link', attrs: { rel: 'preconnect', href: 'https://fonts.googleapis.com' } },
        { tag: 'link', attrs: { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: true } },
        {
          tag: 'link',
          attrs: {
            rel: 'stylesheet',
            href: 'https://fonts.googleapis.com/css2?family=Bricolage+Grotesque:opsz,wght@12..96,600..700&family=Instrument+Sans:wght@400..700&family=JetBrains+Mono:wght@400;600&display=swap',
          },
        },
      ],
      expressiveCode: {
        themes: ['github-dark-default', 'github-light-default'],
        styleOverrides: {
          borderRadius: '8px',
          borderColor: 'var(--og-line)',
          codeBackground: 'var(--og-paper-sunken)',
          codeFontFamily: 'var(--__sl-font-mono)',
          frames: {
            editorTabBarBackground: 'var(--og-paper-raised)',
            editorActiveTabBackground: 'var(--og-paper-sunken)',
            editorActiveTabIndicatorTopColor: 'var(--og-barrier)',
            terminalTitlebarBackground: 'var(--og-paper-raised)',
            terminalBackground: 'var(--og-paper-sunken)',
            frameBoxShadowCssValue: 'none',
          },
        },
      },
      social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/lucdoe/opengateway' }],
      // Page paths are relative to site/, as ../docs/auth.md.
      editLink: { baseUrl: 'https://github.com/lucdoe/opengateway/edit/main/site/' },
      sidebar: [
        { label: 'Guides', items: ['configuration', 'plugins', 'auth', 'operations'] },
        { label: 'Internals', items: ['architecture', 'benchmark'] },
      ],
    }),
  ],
});
