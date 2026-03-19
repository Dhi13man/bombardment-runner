import type { ParserStrategy } from '../types/api';

/** Strip BOM from atob-decoded string (UTF-8: EF BB BF as 3 chars, UTF-16: FEFF as 1 char). */
function stripBom(raw: string): string {
  if (raw.charCodeAt(0) === 0xEF && raw.charCodeAt(1) === 0xBB && raw.charCodeAt(2) === 0xBF) return raw.slice(3);
  if (raw.charCodeAt(0) === 0xFEFF) return raw.slice(1);
  return raw;
}

/** Peek first non-whitespace char of base64 content to detect JSON vs NDJSON. */
export function detectJsonMode(b64: string): 'JSON' | 'NDJSON' {
  try {
    const raw = stripBom(atob(b64.slice(0, 4096)));
    return raw.trimStart()[0] === '[' ? 'JSON' : 'NDJSON';
  } catch { return 'JSON'; }
}

/** Detect CSV delimiter from first 5 lines of base64 content. */
export function detectDelimiter(b64: string): string {
  try {
    const raw = stripBom(atob(b64.slice(0, 4096)));
    const lines = raw.split('\n').slice(0, 5).filter(l => l.trim());
    if (!lines.length) return ',';
    let best = ',', bestScore = 0;
    for (const d of [',', '\t', '|', ';']) {
      const counts = lines.map(l => l.split(d).length - 1);
      const f = counts[0];
      if (f > 0 && counts.every(c => c === f) && f > bestScore) { bestScore = f; best = d; }
    }
    return best;
  } catch { return ','; }
}

/** Detect parser strategy from file extension. */
export function detectStrategyFromExt(ext: string | undefined): ParserStrategy | null {
  switch (ext) {
    case 'csv': return 'CSV';
    case 'tsv': return 'CSV';
    case 'json': return 'JSON';
    case 'jsonl': case 'ndjson': return 'NDJSON';
    case 'xlsx': return 'EXCEL';
    case 'parquet': return 'PARQUET';
    default: return null;
  }
}
