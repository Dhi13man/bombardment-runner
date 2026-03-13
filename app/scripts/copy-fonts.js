const fs = require('fs');
const path = require('path');

const FONTS_DIR = path.join(__dirname, '..', 'src', 'app', 'ui', 'static', 'fonts');

const FONT_MAP = [
  {
    src: 'node_modules/@fontsource/ibm-plex-sans/files/ibm-plex-sans-latin-400-normal.woff2',
    dest: 'ibm-plex-sans-latin-400.woff2',
  },
  {
    src: 'node_modules/@fontsource/ibm-plex-sans/files/ibm-plex-sans-latin-500-normal.woff2',
    dest: 'ibm-plex-sans-latin-500.woff2',
  },
  {
    src: 'node_modules/@fontsource/ibm-plex-sans/files/ibm-plex-sans-latin-600-normal.woff2',
    dest: 'ibm-plex-sans-latin-600.woff2',
  },
  {
    src: 'node_modules/@fontsource/jetbrains-mono/files/jetbrains-mono-latin-400-normal.woff2',
    dest: 'jetbrains-mono-latin-400.woff2',
  },
  {
    src: 'node_modules/@fontsource/jetbrains-mono/files/jetbrains-mono-latin-500-normal.woff2',
    dest: 'jetbrains-mono-latin-500.woff2',
  },
];

if (!fs.existsSync(FONTS_DIR)) {
  fs.mkdirSync(FONTS_DIR, { recursive: true });
}

for (const { src, dest } of FONT_MAP) {
  const srcPath = path.join(__dirname, '..', src);
  const destPath = path.join(FONTS_DIR, dest);

  if (!fs.existsSync(srcPath)) {
    console.error(`Font source not found: ${srcPath}`);
    console.error('Run "npm install" first to download font packages.');
    process.exit(1);
  }

  fs.copyFileSync(srcPath, destPath);
  const sizeKB = (fs.statSync(destPath).size / 1024).toFixed(1);
  console.log(`Copied ${dest} (${sizeKB} KB)`);
}

console.log(`All fonts copied to ${FONTS_DIR}`);
