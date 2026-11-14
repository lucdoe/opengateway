// The docs link to each other the way GitHub renders them, as
// configuration.md#store. On the site every page is a folder, so those
// become ../configuration/#store. Links into the repository, such as
// diagrams-as-code, go to GitHub.

const repo = 'https://github.com/lucdoe/opengateway/tree/main/docs/';

export function relativeLinks() {
  return (tree) => {
    const visit = (node) => {
      if (node.type === 'link' && typeof node.url === 'string') {
        node.url = rewrite(node.url);
      }
      node.children?.forEach(visit);
    };
    visit(tree);
  };
}

function rewrite(url) {
  if (/^[a-z]+:|^#|^\//i.test(url)) return url; // absolute, an anchor, or rooted
  const page = url.match(/^([\w-]+)\.md(#.*)?$/);
  if (page) return `../${page[1].toLowerCase() === 'readme' ? '' : page[1] + '/'}${page[2] ?? ''}`;
  return repo + url;
}
