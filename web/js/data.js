export function parseData() {
  const hash = window.location.hash;
  if (!hash || !hash.startsWith('#data=')) return null;

  try {
    const encoded = hash.slice(6);
    // URL-safe base64 decode (no padding)
    const base64 = encoded.replace(/-/g, '+').replace(/_/g, '/');
    const json = atob(base64);
    const data = JSON.parse(json);

    if (!data.v || !data.tk) return null;
    if (data.v > 1) return { error: 'version_mismatch' };
    // Expand short keys to readable names for templates
    return {
      v: data.v,
      ts: data.ts,
      tz: data.tz,
      date: data.d,
      period: { from: data.p.f, to: data.p.t },
      tokens: { input: data.tk.i, output: data.tk.o, total: data.tk.t, by_source: data.tk.s },
      cost: data.c,
      sessions: { total: data.ss.t, claude: data.ss.c, codex: data.ss.x },
    };
  } catch {
    return null;
  }
}
