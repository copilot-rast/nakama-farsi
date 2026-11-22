import fs from 'node:fs';
import path from 'node:path';

const ROOT = path.resolve(__dirname, '..');
const EXCLUDE_FILE = path.join(ROOT, '..', 'build', 'plugins.exclude');
const STARGAME_MARKS = /\/stargame\/|lib\/game'|GAME_URL|STAR100|HALF50|STARFAN|\bstar_[a-z]|star-vip|stargame_tictactoe|stargame\.(banner|golden_rush)|stargame_(start|finish)|daily_login|round_finished|golden_star/;

export function excludedPlugins(): string[] {
  if (!fs.existsSync(EXCLUDE_FILE)) return [];
  return fs.readFileSync(EXCLUDE_FILE, 'utf8').split(/\s+/).filter((n) => /^[a-z][a-z0-9_]*$/.test(n));
}

export function stargameBuilt(): boolean {
  return !excludedPlugins().includes('stargame');
}

export function sampleSpecIgnores(): RegExp[] {
  if (stargameBuilt()) return [];
  const out: RegExp[] = [];
  for (const dir of ['console', 'game']) {
    const base = path.join(ROOT, 'specs', dir);
    if (!fs.existsSync(base)) continue;
    for (const f of fs.readdirSync(base)) {
      if (!f.endsWith('.spec.ts')) continue;
      if (STARGAME_MARKS.test(fs.readFileSync(path.join(base, f), 'utf8'))) out.push(new RegExp('(^|/)' + dir + '/' + f.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') + '$'));
    }
  }
  return out;
}
