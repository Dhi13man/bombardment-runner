function normalizeSvgContent(content) {
  if (content.includes('<!--') || content.includes('-->')) {
    throw new Error('SVG comments are not allowed');
  }

  return content
    .trim()
    .replace(/\s+/g, ' ')
    .replace(/ ?(\/?>)/g, '$1');
}

module.exports = { normalizeSvgContent };
