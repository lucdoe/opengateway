// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightThemeNova from 'starlight-theme-nova';
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
      plugins: [starlightThemeNova()],
      social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/lucdoe/opengateway' }],
      // Page paths are relative to site/, as ../docs/auth.md.
      editLink: { baseUrl: 'https://github.com/lucdoe/opengateway/edit/main/site/' },
      sidebar: [
        { label: 'Guides', items: ['configuration', 'plugins', 'auth', 'operations'] },
        { label: 'Internals', items: ['architecture'] },
      ],
    }),
  ],
});
