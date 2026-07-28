const assert = require('node:assert/strict');
const { describe, it } = require('node:test');

const { normalizeSvgContent } = require('./svg-content');

describe('normalizeSvgContent', () => {
  it('normalizeSvgContent when comment delimiters are present then rejects content', () => {
    // Arrange
    const payloads = [
      '<!-- harmless comment --><path d="M0 0"/>',
      '<!<!-- -->--><script>alert(1)</script>',
      '<path d="M0 0"/><!--',
    ];

    // Act / Assert
    for (const payload of payloads) {
      assert.throws(() => normalizeSvgContent(payload), /comments are not allowed/);
    }
  });

  it('normalizeSvgContent when markup is comment-free then normalizes whitespace', () => {
    // Arrange
    const content = '\n  <path d="M0 0" />\n  <circle cx="1" />\n';

    // Act
    const result = normalizeSvgContent(content);

    // Assert
    assert.equal(result, '<path d="M0 0"/> <circle cx="1"/>');
  });
});
