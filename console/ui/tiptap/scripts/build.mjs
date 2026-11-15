import path from 'node:path';
import url from 'node:url';
import { build } from 'esbuild';

const ROOT = path.resolve(path.dirname(url.fileURLToPath(import.meta.url)), '..');
const OUT = path.resolve(ROOT, '..', 'economy', 'vendor', 'tiptap.js');

await build({
  entryPoints: [path.join(ROOT, 'src', 'entry.js')],
  bundle: true,
  minify: true,
  format: 'iife',
  target: ['es2019'],
  legalComments: 'eof',
  outfile: OUT,
  logLevel: 'info'
});
