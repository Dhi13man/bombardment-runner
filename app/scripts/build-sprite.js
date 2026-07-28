const fs = require('fs');
const path = require('path');

const { normalizeSvgContent } = require('./svg-content');

const ICONS_SRC = path.join(__dirname, '..', 'node_modules', 'lucide-static', 'icons');
const SPRITE_OUT = path.join(__dirname, '..', 'src', 'app', 'ui', 'static', 'icons', 'sprite.svg');

// Only include the icons the UI actually uses (mapped from FontAwesome)
const ICON_LIST = [
  'rocket',
  'file-input',
  'sliders-horizontal',
  'target',
  'check-circle-2',
  'arrow-left',
  'arrow-right',
  'code',
  'link',
  'tags',
  'clock',
  'heart-pulse',
  'shield',
  'file-code',
  'hourglass',
  'timer',
  'plug',
  'scale',
  'upload',
  'plus',
  'trash-2',
  'refresh-cw',
  'alert-triangle',
  'info',
  'folder',
  'layers',
  'settings',
  'list-checks',
  'history',
  'inbox',
  'x',
  'chevron-down',
  'sun',
  'moon',
  'menu',
  'panel-left',
  'check',
  'search',
  'database',
];

const outDir = path.dirname(SPRITE_OUT);
if (!fs.existsSync(outDir)) {
  fs.mkdirSync(outDir, { recursive: true });
}

const symbols = [];
let missing = 0;

for (const name of ICON_LIST) {
  const svgPath = path.join(ICONS_SRC, `${name}.svg`);
  if (!fs.existsSync(svgPath)) {
    console.error(`Missing icon: ${name} (${svgPath})`);
    missing++;
    continue;
  }

  let svg = fs.readFileSync(svgPath, 'utf8');

  // Extract inner content (paths, circles, lines, etc.) from the <svg> wrapper
  const innerMatch = svg.match(/<svg[^>]*>([\s\S]*?)<\/svg>/);
  if (!innerMatch) {
    console.error(`Could not parse SVG: ${name}`);
    missing++;
    continue;
  }

  const inner = normalizeSvgContent(innerMatch[1]);

  symbols.push(
    `<symbol id="icon-${name}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">${inner}</symbol>`
  );
}

const sprite = `<svg xmlns="http://www.w3.org/2000/svg" style="display:none">${symbols.join('')}</svg>`;

fs.writeFileSync(SPRITE_OUT, sprite);
const sizeKB = (fs.statSync(SPRITE_OUT).size / 1024).toFixed(1);

console.log(`Sprite built: ${ICON_LIST.length - missing}/${ICON_LIST.length} icons, ${sizeKB} KB`);
if (missing > 0) {
  console.error(`${missing} icon(s) missing, check ICON_LIST names against lucide-static`);
  process.exit(1);
}
