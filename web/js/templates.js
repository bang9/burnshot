// Polyfill for roundRect (Safari < 16)
if (!CanvasRenderingContext2D.prototype.roundRect) {
  CanvasRenderingContext2D.prototype.roundRect = function(x, y, w, h, radii) {
    const r = typeof radii === 'number' ? radii : (Array.isArray(radii) ? radii[0] : 0);
    this.moveTo(x + r, y);
    this.arcTo(x + w, y, x + w, y + h, r);
    this.arcTo(x + w, y + h, x, y + h, r);
    this.arcTo(x, y + h, x, y, r);
    this.arcTo(x, y, x + w, y, r);
    this.closePath();
  };
}

function formatTokens(n) {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
  if (n >= 1_000) return (n / 1_000).toFixed(0) + 'K';
  return n.toString();
}

function formatNumber(n) {
  return n.toLocaleString();
}

function formatTime(ts, tz) {
  try {
    return new Date(ts * 1000).toLocaleTimeString('en-US', {
      hour: '2-digit', minute: '2-digit', hour12: false, timeZone: tz
    });
  } catch {
    return new Date(ts * 1000).toLocaleTimeString('en-US', {
      hour: '2-digit', minute: '2-digit', hour12: false
    });
  }
}

function formatDate(date) {
  return date.replace(/-/g, '.');
}

// Template 1: Minimal
const minimal = {
  name: 'Minimal',
  render(ctx, w, h, data) {
    // Top: date + time
    ctx.font = '600 11px monospace';
    ctx.fillStyle = 'rgba(255,255,255,0.7)';
    ctx.fillText(formatDate(data.date), 16, 24);
    ctx.textAlign = 'right';
    ctx.fillText(formatTime(data.ts, data.tz), w - 16, 24);
    ctx.textAlign = 'left';

    // Bottom stats
    const y = h - 20;
    ctx.font = '700 28px system-ui';
    ctx.fillStyle = '#fff';
    ctx.fillText(formatNumber(data.tokens.total), 20, y - 60);

    ctx.font = '500 11px monospace';
    ctx.fillStyle = 'rgba(255,255,255,0.5)';
    ctx.fillText(`tokens · ${data.period.from}–${data.period.to}`, 20, y - 44);

    ctx.font = '600 18px system-ui';
    ctx.fillStyle = '#fff';
    ctx.fillText(`$${data.cost.toFixed(2)}`, 20, y - 20);
    ctx.fillText(`${data.sessions.total}`, 100, y - 20);

    ctx.font = '400 9px monospace';
    ctx.fillStyle = 'rgba(255,255,255,0.4)';
    ctx.fillText('COST', 20, y - 8);
    ctx.fillText('SESSIONS', 100, y - 8);

    ctx.font = '500 10px monospace';
    ctx.fillStyle = 'rgba(255,255,255,0.3)';
    ctx.fillText('BURNSHOT', 20, y);
  }
};

// Template 2: HUD
const hud = {
  name: 'HUD',
  render(ctx, w, h, data) {
    const green = '#00ff88';

    // Top bar
    ctx.font = '600 9px monospace';
    ctx.fillStyle = green;
    ctx.fillText('BURNSHOT', 16, 20);
    ctx.font = '400 11px monospace';
    ctx.fillStyle = 'rgba(255,255,255,0.6)';
    ctx.fillText(`${formatDate(data.date)} · ${formatTime(data.ts, data.tz)}`, 16, 34);

    ctx.textAlign = 'right';
    ctx.font = '400 9px monospace';
    ctx.fillStyle = 'rgba(0,255,136,0.6)';
    ctx.fillText(`${data.period.from}–${data.period.to}`, w - 16, 20);
    ctx.textAlign = 'left';

    // Bottom HUD panel
    const panelH = 90;
    const panelY = h - panelH - 16;
    ctx.fillStyle = 'rgba(0,0,0,0.3)';
    ctx.strokeStyle = 'rgba(0,255,136,0.2)';
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.roundRect(12, panelY, w - 24, panelH, 8);
    ctx.fill();
    ctx.stroke();

    const col1 = 24;
    const col2 = w / 2 + 4;
    let row = panelY + 20;

    ctx.font = '600 9px monospace';
    ctx.fillStyle = 'rgba(0,255,136,0.6)';
    ctx.fillText('TOKENS', col1, row);
    ctx.fillText('COST', col2, row);
    row += 18;
    ctx.font = '700 20px monospace';
    ctx.fillStyle = '#fff';
    ctx.fillText(formatTokens(data.tokens.total), col1, row);
    ctx.fillText(`$${data.cost.toFixed(2)}`, col2, row);

    row += 20;
    ctx.font = '600 9px monospace';
    ctx.fillStyle = 'rgba(0,255,136,0.6)';
    ctx.fillText('SESSIONS', col1, row);
    ctx.fillText('CLI', col2, row);
    row += 14;
    ctx.font = '600 14px monospace';
    ctx.fillStyle = '#fff';
    ctx.fillText(`${data.sessions.total}`, col1, row);
    ctx.fillText(`CC ${data.sessions.claude} · CX ${data.sessions.codex}`, col2, row);
  }
};

// Template 3: Nike-Style Bold
const bold = {
  name: 'Bold',
  render(ctx, w, h, data) {
    // Bottom gradient
    const grad = ctx.createLinearGradient(0, h * 0.5, 0, h);
    grad.addColorStop(0, 'rgba(0,0,0,0)');
    grad.addColorStop(1, 'rgba(0,0,0,0.85)');
    ctx.fillStyle = grad;
    ctx.fillRect(0, h * 0.5, w, h * 0.5);

    const baseY = h - 20;

    // Big number
    ctx.font = '900 42px system-ui';
    ctx.fillStyle = '#fff';
    ctx.fillText(formatTokens(data.tokens.total), 20, baseY - 80);

    ctx.font = '700 12px system-ui';
    ctx.fillStyle = '#ff6b35';
    ctx.letterSpacing = '4px';
    ctx.fillText('TOKENS BURNED', 20, baseY - 62);
    ctx.letterSpacing = '0px';

    // Stats row
    const stats = [
      [`$${Math.floor(data.cost)}`, 'COST'],
      [`${data.sessions.total}`, 'SESSIONS'],
    ];
    let x = 20;
    for (const [val, label] of stats) {
      ctx.font = '800 22px system-ui';
      ctx.fillStyle = '#fff';
      ctx.fillText(val, x, baseY - 30);
      ctx.font = '500 9px system-ui';
      ctx.fillStyle = 'rgba(255,255,255,0.5)';
      ctx.fillText(label, x, baseY - 18);
      x += 80;
    }

    // Bottom bar
    ctx.font = '700 10px system-ui';
    ctx.fillStyle = 'rgba(255,255,255,0.4)';
    ctx.fillText('BURNSHOT', 20, baseY);
    ctx.textAlign = 'right';
    ctx.font = '400 10px monospace';
    ctx.fillText(`${formatDate(data.date)} · ${formatTime(data.ts, data.tz)}`, w - 20, baseY);
    ctx.textAlign = 'left';
  }
};

// Template 4: Glass Card
const glass = {
  name: 'Glass',
  render(ctx, w, h, data) {
    const cardW = w - 40;
    const cardH = 140;
    const cardX = 20;
    const cardY = h - cardH - 20;

    // Glass background
    ctx.fillStyle = 'rgba(255,255,255,0.08)';
    ctx.strokeStyle = 'rgba(255,255,255,0.1)';
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.roundRect(cardX, cardY, cardW, cardH, 16);
    ctx.fill();
    ctx.stroke();

    let y = cardY + 24;

    // Header
    ctx.font = '600 11px system-ui';
    ctx.fillStyle = '#fff';
    ctx.fillText('BURNSHOT', cardX + 20, y);
    ctx.textAlign = 'right';
    ctx.font = '400 11px monospace';
    ctx.fillStyle = 'rgba(255,255,255,0.5)';
    ctx.fillText(formatDate(data.date).slice(5), cardX + cardW - 20, y);
    ctx.textAlign = 'left';

    y += 24;
    ctx.font = '700 32px system-ui';
    ctx.fillStyle = '#fff';
    ctx.fillText(formatNumber(data.tokens.total), cardX + 20, y);

    y += 16;
    ctx.font = '400 10px monospace';
    ctx.fillStyle = 'rgba(255,255,255,0.4)';
    ctx.fillText(`tokens · ${data.period.from}–${data.period.to}`, cardX + 20, y);

    // Divider
    y += 14;
    ctx.strokeStyle = 'rgba(255,255,255,0.1)';
    ctx.beginPath();
    ctx.moveTo(cardX + 20, y);
    ctx.lineTo(cardX + cardW - 20, y);
    ctx.stroke();

    // Stats row
    y += 22;
    const colW = (cardW - 40) / 3;
    const cols = [
      [`$${data.cost.toFixed(2)}`, 'COST'],
      [`${data.sessions.total}`, 'SESSIONS'],
      [`CC ${data.sessions.claude} · CX ${data.sessions.codex}`, 'TOOLS'],
    ];
    for (let i = 0; i < cols.length; i++) {
      const cx = cardX + 20 + colW * i + colW / 2;
      ctx.textAlign = 'center';
      ctx.font = '700 16px system-ui';
      ctx.fillStyle = '#fff';
      ctx.fillText(cols[i][0], cx, y);
      ctx.font = '400 8px monospace';
      ctx.fillStyle = 'rgba(255,255,255,0.4)';
      ctx.fillText(cols[i][1], cx, y + 14);
    }
    ctx.textAlign = 'left';
  }
};

// Template 5: VS (Claude vs Codex breakdown)
const versus = {
  name: 'VS',
  render(ctx, w, h, data) {
    const claudeColor = '#d97757'; // Claude signature orange
    const codexColor = '#10a37f';  // OpenAI signature teal
    const by = data.tokens.by_source || {};
    const claudeTokens = by.claude || 0;
    const codexTokens = by.codex || 0;
    const total = claudeTokens + codexTokens || 1;
    const claudePct = Math.round((claudeTokens / total) * 100);
    const codexPct = 100 - claudePct;

    // Bottom gradient
    const grad = ctx.createLinearGradient(0, h * 0.35, 0, h);
    grad.addColorStop(0, 'rgba(0,0,0,0)');
    grad.addColorStop(1, 'rgba(0,0,0,0.9)');
    ctx.fillStyle = grad;
    ctx.fillRect(0, h * 0.35, w, h * 0.65);

    const pad = 20;
    const barY = h - 150;

    // Title
    ctx.font = '700 10px monospace';
    ctx.fillStyle = 'rgba(255,255,255,0.4)';
    ctx.fillText('BURNSHOT', pad, barY - 60);
    ctx.textAlign = 'right';
    ctx.font = '400 10px monospace';
    ctx.fillText(`${formatDate(data.date)} · ${formatTime(data.ts, data.tz)}`, w - pad, barY - 60);
    ctx.textAlign = 'left';

    // Total
    ctx.font = '800 28px system-ui';
    ctx.fillStyle = '#fff';
    const totalText = formatTokens(total);
    const totalTextWidth = ctx.measureText(totalText).width;
    ctx.fillText(totalText, pad, barY - 30);
    ctx.font = '400 10px monospace';
    ctx.fillStyle = 'rgba(255,255,255,0.4)';
    ctx.fillText(`tokens · $${data.cost.toFixed(2)}`, pad + totalTextWidth + 8, barY - 30);

    // Percent bar
    const barW = w - pad * 2;
    const barH = 20;
    const claudeW = Math.max(barW * (claudePct / 100), claudePct > 0 ? 4 : 0);
    const codexW = barW - claudeW;

    // Claude bar (left)
    if (claudeW > 0) {
      ctx.fillStyle = claudeColor;
      ctx.beginPath();
      ctx.roundRect(pad, barY, claudeW, barH, claudeW >= barW ? 6 : [6, 0, 0, 6]);
      ctx.fill();
    }

    // Codex bar (right)
    if (codexW > 0) {
      ctx.fillStyle = codexColor;
      ctx.beginPath();
      ctx.roundRect(pad + claudeW, barY, codexW, barH, codexW >= barW ? 6 : [0, 6, 6, 0]);
      ctx.fill();
    }

    // Percent labels on bar
    ctx.font = '700 11px system-ui';
    if (claudePct > 15) {
      ctx.fillStyle = '#fff';
      ctx.fillText(`${claudePct}%`, pad + 6, barY + 14);
    }
    if (codexPct > 15) {
      ctx.fillStyle = '#fff';
      ctx.textAlign = 'right';
      ctx.fillText(`${codexPct}%`, w - pad - 6, barY + 14);
      ctx.textAlign = 'left';
    }

    // VS detail rows
    const detailY = barY + barH + 20;
    const halfW = (w - pad * 2) / 2;

    // Claude side (left)
    ctx.fillStyle = claudeColor;
    ctx.font = '700 9px monospace';
    ctx.fillText('CLAUDE', pad, detailY);
    ctx.fillStyle = '#fff';
    ctx.font = '700 20px system-ui';
    ctx.fillText(formatTokens(claudeTokens), pad, detailY + 22);
    ctx.font = '400 10px monospace';
    ctx.fillStyle = 'rgba(255,255,255,0.5)';
    ctx.fillText(`${data.sessions.claude} sessions`, pad, detailY + 38);

    // Codex side (right)
    ctx.textAlign = 'right';
    ctx.fillStyle = codexColor;
    ctx.font = '700 9px monospace';
    ctx.fillText('CODEX', w - pad, detailY);
    ctx.fillStyle = '#fff';
    ctx.font = '700 20px system-ui';
    ctx.fillText(formatTokens(codexTokens), w - pad, detailY + 22);
    ctx.font = '400 10px monospace';
    ctx.fillStyle = 'rgba(255,255,255,0.5)';
    ctx.fillText(`${data.sessions.codex} sessions`, w - pad, detailY + 38);
    ctx.textAlign = 'left';

    // VS divider
    ctx.textAlign = 'center';
    ctx.font = '800 14px system-ui';
    ctx.fillStyle = 'rgba(255,255,255,0.2)';
    ctx.fillText('VS', w / 2, detailY + 22);
    ctx.textAlign = 'left';

    // Period at bottom
    ctx.font = '400 9px monospace';
    ctx.fillStyle = 'rgba(255,255,255,0.3)';
    ctx.fillText(`${data.period.from}–${data.period.to}`, pad, h - 16);
  }
};

export const TEMPLATES = [minimal, hud, bold, glass, versus];
