// Config keys in tables, like Server.CircuitBreaker.ConsecutiveFailures, are
// long words with no place to wrap, so their column squeezes the others. This
// adds a <wbr> after each dot in inline code inside table cells: the key wraps
// at a dot when it has to, and copies unchanged.

export function breakKeys() {
  return (tree) => {
    const visit = (node, inCell) => {
      const cell = inCell || node.tagName === 'td' || node.tagName === 'th';
      if (cell && node.tagName === 'code') {
        node.children = node.children.flatMap((child) =>
          child.type === 'text' ? splitAtDots(child.value) : [child],
        );
        return;
      }
      node.children?.forEach((child) => visit(child, cell));
    };
    visit(tree, false);
  };
}

function splitAtDots(text) {
  return text.split(/(?<=\.)/).flatMap((part, i) =>
    i === 0 ? [{ type: 'text', value: part }] : [{ type: 'element', tagName: 'wbr', properties: {}, children: [] }, { type: 'text', value: part }],
  );
}
