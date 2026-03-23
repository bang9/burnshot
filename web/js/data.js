export function parseData() {
  const hash = window.location.hash;
  if (!hash || !hash.startsWith('#data=')) return null;

  try {
    const encoded = hash.slice(6);
    // URL-safe base64 decode (no padding)
    const base64 = encoded.replace(/-/g, '+').replace(/_/g, '/');
    const json = atob(base64);
    const data = JSON.parse(json);

    if (!data.v || !data.tokens) return null;
    return data;
  } catch {
    return null;
  }
}
