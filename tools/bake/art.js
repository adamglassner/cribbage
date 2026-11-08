'use strict';
/* =====================================================================
   CRIBBAGE - A Lighthouse Card Game (the Keeper's Cottage, Seal Harbor Light)
   All card art and props are drawn here in code (original artwork, (c) Ironite).
   ===================================================================== */
const Art = (() => {
  const CW = 360, CH = 504;              // card canvas size (HD; the game scales it down)
  const INK = '#1c1a1e';
  const RED = '#b22a22', BLACK = '#16202c', GOLD = '#a8843a', GOLD2 = '#e8cf8a';
  const IVORY1 = '#fdfaf2', IVORY2 = '#efe7d4';
  const SERIF = '"Lora", Georgia, serif';
  const DISP = '"Playfair Display", Georgia, serif';
  const TAU = Math.PI * 2;
  const NAVY = '#1d3350', NAVY2 = '#0f1e33';

  // Suit palettes for the court figures: lighthouse red, buoy amber, harbour navy, spruce green.
  const PAL = {
    H: { coat: '#a82a22', dark: '#621712', accent: '#f1e6cc', ink: RED, wall: '#f2dcd4' },
    D: { coat: '#d9a12a', dark: '#8a5e12', accent: '#f6efe0', ink: RED, wall: '#f4e6c4' },
    S: { coat: '#1d3350', dark: '#0f1e33', accent: '#e8c35a', ink: BLACK, wall: '#dde4ec' },
    C: { coat: '#2f5a44', dark: '#183426', accent: '#e8c35a', ink: BLACK, wall: '#dce8dc' }
  };

  /* ---------- small helpers ---------- */
  function rng(seed) { let s = seed >>> 0 || 1; return () => (s = (s * 1664525 + 1013904223) >>> 0) / 4294967296; }
  function rrect(c, x, y, w, h, r) { c.beginPath(); c.moveTo(x + r, y); c.arcTo(x + w, y, x + w, y + h, r); c.arcTo(x + w, y + h, x, y + h, r); c.arcTo(x, y + h, x, y, r); c.arcTo(x, y, x + w, y, r); c.closePath(); }
  function canvas(w, h) { const c = document.createElement('canvas'); c.width = w; c.height = h; return c; }
  const ink = c => { c.strokeStyle = INK; c.lineWidth = 2.6; c.lineJoin = 'round'; c.lineCap = 'round'; };

  /* ---------- suit shapes (drawn, so they look the same on every PC) ---------- */
  function heartPath(c, x, y, s) {
    c.moveTo(x, y + s * 0.45);
    c.bezierCurveTo(x - s * 0.62, y + s * 0.02, x - s * 0.52, y - s * 0.55, x, y - s * 0.2);
    c.bezierCurveTo(x + s * 0.52, y - s * 0.55, x + s * 0.62, y + s * 0.02, x, y + s * 0.45);
  }
  function suitPath(c, s, x, y, size) {
    if (s === 'H') heartPath(c, x, y, size);
    else if (s === 'D') { c.moveTo(x, y - size * 0.5); c.quadraticCurveTo(x + size * 0.18, y - size * 0.18, x + size * 0.4, y); c.quadraticCurveTo(x + size * 0.18, y + size * 0.18, x, y + size * 0.5); c.quadraticCurveTo(x - size * 0.18, y + size * 0.18, x - size * 0.4, y); c.quadraticCurveTo(x - size * 0.18, y - size * 0.18, x, y - size * 0.5); }
    else if (s === 'S') {
      c.save(); c.translate(x, y - size * 0.08); c.rotate(Math.PI); heartPath(c, 0, 0, size * 0.9); c.restore();
      c.moveTo(x, y + size * 0.05); c.quadraticCurveTo(x - size * 0.05, y + size * 0.38, x - size * 0.24, y + size * 0.5); c.lineTo(x + size * 0.24, y + size * 0.5); c.quadraticCurveTo(x + size * 0.05, y + size * 0.38, x, y + size * 0.05);
    } else {
      const r = size * 0.205;
      c.moveTo(x + r, y - size * 0.24); c.arc(x, y - size * 0.24, r, 0, TAU); c.moveTo(x - size * 0.22 + r, y + size * 0.06); c.arc(x - size * 0.22, y + size * 0.06, r, 0, TAU);
      c.moveTo(x + size * 0.22 + r, y + size * 0.06); c.arc(x + size * 0.22, y + size * 0.06, r, 0, TAU);
      c.moveTo(x + size * 0.12, y); c.arc(x, y, size * 0.12, 0, TAU);
      c.moveTo(x, y + size * 0.02); c.quadraticCurveTo(x - size * 0.04, y + size * 0.38, x - size * 0.22, y + size * 0.5); c.lineTo(x + size * 0.22, y + size * 0.5); c.quadraticCurveTo(x + size * 0.04, y + size * 0.38, x, y + size * 0.02);
    }
  }
  function suit(c, s, x, y, size, color, shine = true) {
    c.save(); c.fillStyle = color; c.beginPath(); suitPath(c, s, x, y, size); c.fill();
    if (shine && size > 30) { // a little lacquer highlight
      c.clip(); c.fillStyle = 'rgba(255,255,255,.16)'; c.beginPath(); c.ellipse(x - size * 0.16, y - size * 0.16, size * 0.12, size * 0.2, -0.6, 0, TAU); c.fill();
    }
    c.restore();
  }
  // a four-point compass star
  function star(c, x, y, r, color) {
    c.save(); c.fillStyle = color; c.beginPath();
    for (let i = 0; i < 8; i++) { const a = i * Math.PI / 4 - Math.PI / 2, rr = i % 2 ? r * 0.32 : r; c.lineTo(x + Math.cos(a) * rr, y + Math.sin(a) * rr); }
    c.closePath(); c.fill(); c.restore();
  }
  // a twisted rope, drawn along an ellipse
  function rope(c, cx, cy, rx, ry, w, col, dark) {
    c.save(); c.strokeStyle = col; c.lineWidth = w; c.beginPath(); c.ellipse(cx, cy, rx, ry, 0, 0, TAU); c.stroke();
    c.strokeStyle = dark; c.lineWidth = Math.max(1, w * 0.22);
    const n = Math.round((rx + ry) * 0.42);
    for (let i = 0; i < n; i++) { const a = i / n * TAU, x = cx + Math.cos(a) * rx, y = cy + Math.sin(a) * ry, nx = Math.cos(a), ny = Math.sin(a) * rx / ry, l = Math.hypot(nx, ny);
      c.beginPath(); c.moveTo(x - nx / l * w * 0.45 - ny / l * w * 0.3, y - ny / l * w * 0.45 + nx / l * w * 0.3); c.lineTo(x + nx / l * w * 0.45 + ny / l * w * 0.3, y + ny / l * w * 0.45 - nx / l * w * 0.3); c.stroke(); }
    c.restore();
  }
  function ropeRect(c, x, y, w, h, r, lw, col, dark) {
    c.save(); c.strokeStyle = col; c.lineWidth = lw; rrect(c, x, y, w, h, r); c.stroke();
    c.setLineDash([lw * 0.5, lw * 0.6]); c.strokeStyle = dark; c.lineWidth = lw * 0.55; c.lineDashOffset = 0; rrect(c, x, y, w, h, r); c.stroke(); c.restore();
  }
  // a lighthouse: white tower, red bands, black lantern room
  function lighthouse(c, x, base, h, opt = {}) {
    const w0 = h * 0.2, w1 = h * 0.13;
    c.save(); c.translate(x, base);
    if (opt.beam) { const g = c.createRadialGradient(0, -h * 0.9, 2, 0, -h * 0.9, h * 1.6); g.addColorStop(0, 'rgba(255,240,180,.75)'); g.addColorStop(1, 'rgba(255,240,180,0)'); c.fillStyle = g;
      for (const sx of [-1, 1]) { c.beginPath(); c.moveTo(0, -h * 0.9); c.lineTo(sx * h * 1.8, -h * 1.15); c.lineTo(sx * h * 1.8, -h * 0.68); c.closePath(); c.fill(); } }
    const body = () => { c.beginPath(); c.moveTo(-w0 / 2, 0); c.lineTo(-w1 / 2, -h * 0.78); c.lineTo(w1 / 2, -h * 0.78); c.lineTo(w0 / 2, 0); c.closePath(); };
    if (opt.mono) { c.fillStyle = opt.mono; body(); c.fill(); c.fillRect(-w1 * 0.7, -h * 0.8, w1 * 1.4, h * 0.03); c.fillRect(-w1 * 0.4, -h * 0.95, w1 * 0.8, h * 0.15); c.beginPath(); c.moveTo(-w1 * 0.5, -h * 0.95); c.lineTo(0, -h); c.lineTo(w1 * 0.5, -h * 0.95); c.fill(); c.restore(); return; }
    c.lineWidth = Math.max(1.5, h * 0.012); c.strokeStyle = INK;
    c.fillStyle = '#f8f4ea'; body(); c.fill();
    c.save(); body(); c.clip(); c.fillStyle = '#b22a22'; for (const k of [0.16, 0.44]) c.fillRect(-w0, -h * (k + 0.14), w0 * 2, h * 0.14); c.fillStyle = 'rgba(0,0,0,.12)'; c.fillRect(0, -h, w0, h); c.restore();
    body(); c.stroke();
    c.fillStyle = '#1c1a1e'; c.fillRect(-w1 * 0.75, -h * 0.81, w1 * 1.5, h * 0.035);
    c.fillStyle = opt.lit === false ? '#3a4a5a' : '#ffe9a0'; c.fillRect(-w1 * 0.42, -h * 0.94, w1 * 0.84, h * 0.13); c.strokeRect(-w1 * 0.42, -h * 0.94, w1 * 0.84, h * 0.13);
    c.fillStyle = '#1c1a1e'; c.beginPath(); c.moveTo(-w1 * 0.55, -h * 0.94); c.lineTo(0, -h * 1.02); c.lineTo(w1 * 0.55, -h * 0.94); c.closePath(); c.fill();
    c.fillStyle = '#2a2a2a'; rrect(c, -w0 * 0.12, -h * 0.12, w0 * 0.24, h * 0.12, w0 * 0.1); c.fill();
    c.restore();
  }

  /* ---------- paper ---------- */
  function paper(c, w, h, seed, r = 22) {
    const g = c.createLinearGradient(0, 0, w, h); g.addColorStop(0, IVORY1); g.addColorStop(1, IVORY2);
    rrect(c, 1, 1, w - 2, h - 2, r); c.fillStyle = g; c.fill();
    c.save(); c.clip();
    const R = rng(seed);
    for (let i = 0; i < 900; i++) { c.fillStyle = `rgba(90,80,60,${R() * 0.045})`; c.fillRect(R() * w, R() * h, 1 + R() * 2, 1 + R() * 2); }
    const v = c.createRadialGradient(w / 2, h / 2, h * 0.3, w / 2, h / 2, h * 0.75); v.addColorStop(0, 'rgba(0,0,0,0)'); v.addColorStop(1, 'rgba(90,70,40,.13)');
    c.fillStyle = v; c.fillRect(0, 0, w, h); c.restore();
    rrect(c, 1, 1, w - 2, h - 2, r); c.strokeStyle = 'rgba(50,45,40,.45)'; c.lineWidth = 2; c.stroke();
  }
  // a ship's-chart border: a thin double line with a compass star in each corner
  function border(c, w, h, color) {
    c.save(); c.strokeStyle = color; c.globalAlpha = 0.85;
    c.lineWidth = 2; rrect(c, 14, 14, w - 28, h - 28, 14); c.stroke();
    c.lineWidth = 0.8; rrect(c, 19, 19, w - 38, h - 38, 10); c.stroke();
    for (const [x, y] of [[19, 19], [w - 19, 19], [19, h - 19], [w - 19, h - 19]]) star(c, x, y, 9, color);
    c.restore();
  }
  function index(c, rank, s, color, w, h) {
    const draw = () => {
      c.fillStyle = color; c.textAlign = 'center'; c.textBaseline = 'alphabetic';
      c.font = `700 ${rank === '10' ? 44 : 52}px ${DISP}`;
      c.fillText(rank, 46, 74);
      suit(c, s, 46, 102, 32, color, false);
    };
    draw(); c.save(); c.translate(w, h); c.rotate(Math.PI); draw(); c.restore();
  }

  /* ---------- the harbour folk (court figures) ----------
     Drawn around (0,0) = bottom centre of the bust. Head is ~54px wide at scale 1. */
  function shoulderX(y) {
    if (y <= -64) return 22;
    const bx = t => { const u = 1 - t; return [u * u * u * 22 + 3 * u * u * t * 54 + 3 * u * t * t * 78 + t * t * t * 80, u * u * u * -64 + 3 * u * u * t * -60 + 3 * u * t * t * -42]; };
    let lo = 0, hi = 1; for (let i = 0; i < 30; i++) { const m = (lo + hi) / 2; if (bx(m)[1] < y) lo = m; else hi = m; } return bx(lo)[0];
  }
  function shoulders(c) { c.beginPath(); c.moveTo(-80, 0); c.bezierCurveTo(-78, -42, -54, -60, -22, -64); c.lineTo(22, -64); c.bezierCurveTo(54, -60, 78, -42, 80, 0); c.closePath(); }
  function bust(c, o) {
    const skin = o.skin || '#f2d4b6', hair = o.hair || '#4a2e1c', hy = -118;
    c.save(); ink(c);
    c.fillStyle = hair;
    if (o.hairStyle === 'bun') { c.beginPath(); c.arc(0, hy - 36, 15, 0, TAU); c.fill(); c.stroke(); }
    if (o.hairStyle === 'waves') { c.beginPath(); c.moveTo(-30, hy - 10); c.bezierCurveTo(-44, hy + 20, -40, hy + 52, -26, hy + 60); c.lineTo(26, hy + 60); c.bezierCurveTo(40, hy + 52, 44, hy + 20, 30, hy - 10); c.closePath(); c.fill(); c.stroke(); }
    if (o.hairStyle === 'bob') { c.beginPath(); c.moveTo(-31, hy - 8); c.bezierCurveTo(-38, hy + 14, -36, hy + 28, -30, hy + 32); c.lineTo(30, hy + 32); c.bezierCurveTo(36, hy + 28, 38, hy + 14, 31, hy - 8); c.closePath(); c.fill(); c.stroke(); }
    // the top
    const top = (o.top === 'stripes' || o.top === 'chef' || o.top === 'baker') ? '#fbf8f0' : o.coat;
    shoulders(c); c.fillStyle = top; c.fill();
    if (o.top === 'stripes') { c.fillStyle = o.stripe || '#1f2f55'; for (let y = -58; y < 2; y += 13) { const y0 = y, y1 = Math.min(y + 6, 0), x0 = shoulderX(y0), x1 = shoulderX(y1); c.beginPath(); c.moveTo(-x0, y0); c.lineTo(x0, y0); c.lineTo(x1, y1); c.lineTo(-x1, y1); c.closePath(); c.fill(); } }
    if (o.top === 'dress') { c.save(); shoulders(c); c.clip(); c.fillStyle = 'rgba(255,255,255,.12)'; for (let x = -80; x < 80; x += 14) { c.beginPath(); c.moveTo(x, 0); c.quadraticCurveTo(x + 6, -40, x + 2, -70); c.lineTo(x + 6, -70); c.quadraticCurveTo(x + 10, -40, x + 4, 0); c.fill(); } c.restore(); }
    shoulders(c); c.stroke();
    // neck
    c.fillStyle = skin; c.beginPath(); c.moveTo(-11, -90); c.lineTo(-12, -61); c.lineTo(12, -61); c.lineTo(11, -90); c.closePath(); c.fill();
    c.fillStyle = 'rgba(120,60,40,.18)'; c.beginPath(); c.moveTo(-11, -88); c.quadraticCurveTo(0, -80, 11, -88); c.lineTo(11, -84); c.quadraticCurveTo(0, -76, -11, -84); c.closePath(); c.fill();
    if (o.top === 'dress') { // a sweetheart neckline
      c.fillStyle = skin; c.beginPath(); c.moveTo(-26, -64); c.quadraticCurveTo(-14, -42, 0, -50); c.quadraticCurveTo(14, -42, 26, -64); c.closePath(); c.fill();
      c.beginPath(); c.moveTo(-26, -64); c.quadraticCurveTo(-14, -42, 0, -50); c.quadraticCurveTo(14, -42, 26, -64); c.stroke();
    }
    if (o.top === 'stripes') { c.beginPath(); c.moveTo(-23, -64); c.quadraticCurveTo(0, -55, 23, -64); c.stroke(); }
    if (o.top === 'blouse') { c.fillStyle = o.accent; for (const sx of [-1, 1]) { c.beginPath(); c.moveTo(sx * 2, -62); c.lineTo(sx * 20, -66); c.lineTo(sx * 16, -50); c.closePath(); c.fill(); c.stroke(); } }
    if (o.top === 'turtleneck') { c.fillStyle = o.coat; rrect(c, -15, -82, 30, 22, 6); c.fill(); c.stroke(); c.strokeStyle = o.dark; c.lineWidth = 1.2; for (let x = -10; x <= 10; x += 5) { c.beginPath(); c.moveTo(x, -80); c.lineTo(x, -62); c.stroke(); } ink(c); }
    if (o.top === 'suit' || o.top === 'waistcoat') {
      c.fillStyle = '#fbf8f0'; c.beginPath(); c.moveTo(-18, -64); c.lineTo(0, -14); c.lineTo(18, -64); c.closePath(); c.fill(); c.stroke();
      if (o.top === 'waistcoat') { c.fillStyle = o.vest || '#1e1e24'; for (const sx of [-1, 1]) { c.beginPath(); c.moveTo(sx * 10, -48); c.lineTo(sx * 2, -8); c.lineTo(sx * 2, 0); c.lineTo(sx * 40, 0); c.lineTo(sx * 34, -50); c.closePath(); c.fill(); c.stroke(); } c.fillStyle = GOLD; for (const y of [-30, -18, -6]) { c.beginPath(); c.arc(6, y, 2.4, 0, TAU); c.fill(); } }
      c.fillStyle = o.dark; for (const sx of [-1, 1]) { c.beginPath(); c.moveTo(sx * 20, -64); c.lineTo(sx * (o.top === 'waistcoat' ? 34 : 3), -14); c.lineTo(sx * 46, -34); c.lineTo(sx * 30, -54); c.closePath(); c.fill(); c.stroke(); }
      c.fillStyle = '#fbf8f0'; for (const sx of [-1, 1]) { c.beginPath(); c.moveTo(sx * 1, -62); c.lineTo(sx * 13, -66); c.lineTo(sx * 10, -54); c.closePath(); c.fill(); c.stroke(); }
    }
    if (o.top === 'cardigan') {
      c.fillStyle = o.accent; c.beginPath(); c.moveTo(-18, -64); c.lineTo(0, -26); c.lineTo(18, -64); c.closePath(); c.fill(); c.stroke();
      c.fillStyle = o.dark; for (const y of [-18, -6]) { c.beginPath(); c.arc(0, y, 3, 0, TAU); c.fill(); }
    }
    if (o.top === 'chef' || o.top === 'baker') {
      c.fillStyle = '#d8d4cc'; for (const y of [-46, -30, -14]) for (const x of [-14, 14]) { c.beginPath(); c.arc(x, y, 3.2, 0, TAU); c.fill(); c.stroke(); }
      if (o.top === 'baker') { c.fillStyle = '#e9dcc0'; c.beginPath(); c.moveTo(-34, 0); c.lineTo(-30, -38); c.lineTo(30, -38); c.lineTo(34, 0); c.closePath(); c.fill(); c.stroke(); c.fillStyle = 'rgba(255,255,255,.7)'; c.beginPath(); c.arc(-14, -24, 4, 0, TAU); c.arc(10, -14, 3, 0, TAU); c.fill(); }
    }
    // ---- harbour clothes ----
    if (o.top === 'peacoat') {   // double-breasted wool with brass buttons
      c.fillStyle = o.shirt || '#fbf8f0'; c.beginPath(); c.moveTo(-16, -64); c.lineTo(0, -30); c.lineTo(16, -64); c.closePath(); c.fill(); c.stroke();
      c.fillStyle = o.dark; for (const sx of [-1, 1]) { c.beginPath(); c.moveTo(sx * 18, -64); c.lineTo(sx * 2, -28); c.lineTo(sx * 30, -36); c.lineTo(sx * 40, -58); c.closePath(); c.fill(); c.stroke(); }
      c.fillStyle = GOLD2; for (const y of [-22, -8]) for (const x of [-14, 14]) { c.beginPath(); c.arc(x, y, 3.2, 0, TAU); c.fill(); c.lineWidth = 1.2; c.stroke(); } ink(c);
    }
    if (o.top === 'oilskin') {   // a fisherman's slicker
      c.strokeStyle = 'rgba(0,0,0,.25)'; c.lineWidth = 1.6; for (const x of [-40, 40]) { c.beginPath(); c.moveTo(x * 0.6, -56); c.quadraticCurveTo(x * 0.8, -30, x * 0.9, 0); c.stroke(); } ink(c);
      c.fillStyle = o.dark; c.beginPath(); c.moveTo(-24, -66); c.lineTo(-6, -48); c.lineTo(0, -60); c.lineTo(6, -48); c.lineTo(24, -66); c.lineTo(14, -70); c.lineTo(0, -62); c.lineTo(-14, -70); c.closePath(); c.fill(); c.stroke();
      c.fillStyle = '#2a2a2a'; for (const y of [-38, -22, -6]) { c.fillRect(-3, y - 2, 6, 4); }
      c.fillStyle = 'rgba(255,255,255,.25)'; c.beginPath(); c.ellipse(-44, -30, 6, 18, 0.4, 0, TAU); c.fill();
    }
    if (o.top === 'sweater') {   // a cable-knit fisherman's jumper
      c.save(); shoulders(c); c.clip(); c.strokeStyle = o.dark; c.lineWidth = 1.6;
      for (const x0 of [-36, -12, 12, 36]) { for (let y = -60; y < 4; y += 9) { c.beginPath(); c.moveTo(x0 - 5, y); c.quadraticCurveTo(x0, y + 4, x0 + 5, y); c.moveTo(x0 - 5, y + 4.5); c.quadraticCurveTo(x0, y + 8.5, x0 + 5, y + 4.5); c.stroke(); } }
      c.restore(); ink(c);
      c.fillStyle = o.coat; rrect(c, -16, -80, 32, 20, 7); c.fill(); c.stroke();
      c.strokeStyle = o.dark; c.lineWidth = 1.2; for (let x = -11; x <= 11; x += 5.5) { c.beginPath(); c.moveTo(x, -78); c.lineTo(x, -62); c.stroke(); } ink(c);
    }
    if (o.top === 'apron') {     // a blouse with a pinafore over it
      c.fillStyle = '#fbf8f0'; c.beginPath(); c.moveTo(-30, 0); c.lineTo(-24, -50); c.lineTo(24, -50); c.lineTo(30, 0); c.closePath(); c.fill(); c.stroke();
      c.strokeStyle = o.dark; c.lineWidth = 1; for (let x = -22; x <= 22; x += 6) { c.beginPath(); c.moveTo(x, -48); c.lineTo(x * 1.15, -2); c.stroke(); } ink(c);
      c.beginPath(); c.moveTo(-24, -50); c.lineTo(-18, -64); c.moveTo(24, -50); c.lineTo(18, -64); c.stroke();
      c.fillStyle = o.accent || '#fbf8f0'; c.beginPath(); c.moveTo(-18, -64); c.quadraticCurveTo(0, -52, 18, -64); c.lineTo(10, -68); c.quadraticCurveTo(0, -62, -10, -68); c.closePath(); c.fill(); c.stroke();
    }
    if (o.top === 'shirt') {     // a work shirt with braces
      c.save(); shoulders(c); c.clip(); c.fillStyle = 'rgba(0,0,0,.12)'; for (let x = -76; x < 80; x += 12) c.fillRect(x, -64, 5, 64); c.restore();
      c.fillStyle = '#fbf8f0'; for (const sx of [-1, 1]) { c.beginPath(); c.moveTo(sx * 2, -62); c.lineTo(sx * 18, -66); c.lineTo(sx * 14, -52); c.closePath(); c.fill(); c.stroke(); }
      c.strokeStyle = o.braces || '#3a2a1e'; c.lineWidth = 6; for (const sx of [-1, 1]) { c.beginPath(); c.moveTo(sx * 30, -60); c.lineTo(sx * 22, 0); c.stroke(); } ink(c);
    }
    if (o.neck === 'shawl') {
      c.fillStyle = o.neckColor; c.beginPath(); c.moveTo(-60, -40); c.quadraticCurveTo(-30, -74, 0, -66); c.quadraticCurveTo(30, -74, 60, -40); c.lineTo(20, -10); c.quadraticCurveTo(0, -40, -20, -10); c.closePath(); c.fill(); c.stroke();
      c.fillStyle = GOLD2; c.beginPath(); c.arc(0, -48, 5, 0, TAU); c.fill(); c.stroke();
    }
    // neckwear
    if (o.neck === 'scarf') {
      c.fillStyle = o.neckColor; c.beginPath(); c.moveTo(-17, -78); c.quadraticCurveTo(0, -66, 17, -78); c.lineTo(19, -66); c.quadraticCurveTo(0, -54, -19, -66); c.closePath(); c.fill(); c.stroke();
      c.beginPath(); c.moveTo(8, -64); c.quadraticCurveTo(20, -46, 14, -26); c.lineTo(24, -30); c.quadraticCurveTo(26, -48, 14, -62); c.closePath(); c.fill(); c.stroke();
      c.beginPath(); c.ellipse(9, -63, 6, 5, 0, 0, TAU); c.fill(); c.stroke();
    }
    if (o.neck === 'pearls') { c.fillStyle = '#fffaf0'; c.lineWidth = 1.2; for (let i = 0; i < 11; i++) { const a = Math.PI * (0.12 + i * 0.076); c.beginPath(); c.arc(Math.cos(a) * 22, -70 + Math.sin(a) * 16, 3, 0, TAU); c.fill(); c.stroke(); } ink(c); }
    if (o.neck === 'bowtie') { c.fillStyle = o.neckColor; c.beginPath(); c.moveTo(0, -60); c.lineTo(-13, -67); c.lineTo(-13, -52); c.closePath(); c.moveTo(0, -60); c.lineTo(13, -67); c.lineTo(13, -52); c.closePath(); c.fill(); c.stroke(); rrect(c, -3.5, -63.5, 7, 7, 2); c.fill(); c.stroke(); }
    if (o.neck === 'kerchief') { c.fillStyle = o.neckColor; c.beginPath(); c.moveTo(-16, -66); c.quadraticCurveTo(0, -58, 16, -66); c.lineTo(0, -44); c.closePath(); c.fill(); c.stroke(); }
    if (o.neck === 'choker') { c.fillStyle = '#18141a'; c.fillRect(-11, -84, 22, 5); c.fillStyle = '#9a7ab8'; c.beginPath(); c.arc(0, -79, 3.5, 0, TAU); c.fill(); c.stroke(); }
    // head
    c.fillStyle = skin; c.beginPath(); c.ellipse(-27, hy + 2, 5, 8, 0, 0, TAU); c.ellipse(27, hy + 2, 5, 8, 0, 0, TAU); c.fill(); c.stroke();
    if (o.earrings) { c.fillStyle = o.earrings; for (const sx of [-1, 1]) { c.beginPath(); c.arc(sx * 28, hy + 13, 3.4, 0, TAU); c.fill(); } }
    c.fillStyle = skin; c.beginPath(); c.ellipse(0, hy, 27, 33, 0, 0, TAU); c.fill(); c.stroke();
    // front hair
    c.fillStyle = hair;
    if (o.hairStyle === 'short') { c.beginPath(); c.ellipse(0, hy - 14, 28.5, 22, 0, Math.PI, 0); c.quadraticCurveTo(14, hy - 22, -4, hy - 18); c.quadraticCurveTo(-18, hy - 14, -28.5, hy - 6); c.closePath(); c.fill(); c.stroke(); }
    if (o.hairStyle === 'side') { c.beginPath(); c.moveTo(-28.5, hy - 2); c.bezierCurveTo(-34, hy - 36, -6, hy - 44, 6, hy - 38); c.bezierCurveTo(24, hy - 40, 34, hy - 24, 28.5, hy - 4); c.quadraticCurveTo(24, hy - 22, 8, hy - 24); c.quadraticCurveTo(-2, hy - 26, -10, hy - 20); c.quadraticCurveTo(-20, hy - 14, -28.5, hy - 2); c.closePath(); c.fill(); c.stroke(); c.strokeStyle = 'rgba(255,255,255,.22)'; c.lineWidth = 1.4; c.beginPath(); c.moveTo(-12, hy - 32); c.quadraticCurveTo(4, hy - 38, 18, hy - 32); c.stroke(); ink(c); }
    if (o.hairStyle === 'bob') { c.beginPath(); c.moveTo(-29, hy + 14); c.bezierCurveTo(-36, hy - 34, 36, hy - 34, 29, hy + 14); c.lineTo(26, hy - 6); c.quadraticCurveTo(18, hy - 20, 4, hy - 16); c.quadraticCurveTo(-10, hy - 20, -20, hy - 14); c.quadraticCurveTo(-26, hy - 8, -29, hy + 14); c.closePath(); c.fill(); c.stroke(); }
    if (o.hairStyle === 'bun' || o.hairStyle === 'waves') { c.beginPath(); c.moveTo(-29, hy + 6); c.bezierCurveTo(-34, hy - 30, -14, hy - 40, 0, hy - 36); c.bezierCurveTo(14, hy - 40, 34, hy - 30, 29, hy + 6); c.bezierCurveTo(22, hy - 10, 14, hy - 24, 0, hy - 22); c.bezierCurveTo(-14, hy - 24, -22, hy - 10, -29, hy + 6); c.closePath(); c.fill(); c.stroke(); }
    if (o.beard) { c.fillStyle = o.beardColor || hair; c.beginPath(); c.moveTo(-25, hy + 4); c.quadraticCurveTo(-24, hy + 42, 0, hy + 48); c.quadraticCurveTo(24, hy + 42, 25, hy + 4); c.quadraticCurveTo(14, hy + 22, 0, hy + 20); c.quadraticCurveTo(-14, hy + 22, -25, hy + 4); c.fill(); c.stroke(); }
    // face
    c.fillStyle = 'rgba(214,110,100,.28)'; c.beginPath(); c.arc(-15, hy + 9, 6.5, 0, TAU); c.arc(15, hy + 9, 6.5, 0, TAU); c.fill();
    c.fillStyle = INK; const ey = o.lady ? 3.4 : 3;
    c.beginPath(); c.ellipse(-10, hy - 3, 2.6, ey, 0, 0, TAU); c.ellipse(10, hy - 3, 2.6, ey, 0, 0, TAU); c.fill();
    if (o.lady) { c.lineWidth = 1.4; for (const sx of [-1, 1]) { c.beginPath(); c.moveTo(sx * 12.5, hy - 5); c.lineTo(sx * 15.5, hy - 7); c.stroke(); } }
    c.lineWidth = 2.2; c.strokeStyle = o.brow || hair; const st = o.stern ? 2 : 0, st1 = o.stern ? 0 : 1;
    c.beginPath(); c.moveTo(-15, hy - 11 + st); c.lineTo(-5, hy - 12 - st1); c.moveTo(15, hy - 11 + st); c.lineTo(5, hy - 12 - st1); c.stroke();
    c.strokeStyle = INK; c.lineWidth = 1.8; c.beginPath(); c.moveTo(1, hy); c.quadraticCurveTo(-3, hy + 9, 1, hy + 11); c.stroke();
    if (o.mole) { c.fillStyle = INK; c.beginPath(); c.arc(-13, hy + 16, 1.5, 0, TAU); c.fill(); }
    if (o.lady) { c.fillStyle = o.lip || '#b3283a'; c.beginPath(); if (o.smile) { c.moveTo(-7, hy + 18); c.quadraticCurveTo(0, hy + 26, 7, hy + 18); c.quadraticCurveTo(0, hy + 21, -7, hy + 18); } else c.ellipse(0, hy + 20, 6, o.stern ? 1.8 : 2.8, 0, 0, TAU); c.fill(); }
    else if (!o.mustache || o.mustache === 'pencil') { c.beginPath(); if (o.smile) { c.moveTo(-7, hy + 18); c.quadraticCurveTo(0, hy + 24, 7, hy + 18); } else { c.moveTo(-6, hy + 20); c.quadraticCurveTo(0, hy + 22, 6, hy + 20); } c.stroke(); }
    const mc = o.mustacheColor || o.brow || hair;
    if (o.mustache === 'pencil') { c.strokeStyle = mc; c.lineWidth = 2.2; c.beginPath(); c.moveTo(-10, hy + 14); c.quadraticCurveTo(0, hy + 11, 10, hy + 14); c.stroke(); }
    if (o.mustache === 'handlebar') { c.fillStyle = mc; c.strokeStyle = INK; c.lineWidth = 1.6; c.beginPath(); c.moveTo(0, hy + 13); c.bezierCurveTo(-8, hy + 10, -18, hy + 18, -24, hy + 10); c.bezierCurveTo(-28, hy + 6, -26, hy + 2, -22, hy + 4); c.bezierCurveTo(-22, hy + 10, -16, hy + 20, 0, hy + 17); c.bezierCurveTo(16, hy + 20, 22, hy + 10, 22, hy + 4); c.bezierCurveTo(26, hy + 2, 28, hy + 6, 24, hy + 10); c.bezierCurveTo(18, hy + 18, 8, hy + 10, 0, hy + 13); c.fill(); c.stroke(); }
    if (o.mustache === 'walrus') { c.fillStyle = mc; c.strokeStyle = INK; c.lineWidth = 1.6; c.beginPath(); c.moveTo(0, hy + 11); c.bezierCurveTo(-10, hy + 8, -18, hy + 12, -18, hy + 22); c.quadraticCurveTo(-8, hy + 18, 0, hy + 20); c.quadraticCurveTo(8, hy + 18, 18, hy + 22); c.bezierCurveTo(18, hy + 12, 10, hy + 8, 0, hy + 11); c.fill(); c.stroke(); }
    ink(c);
    if (o.glasses === 'round') { c.strokeStyle = '#3a2a1e'; c.lineWidth = 2; c.beginPath(); c.arc(-10, hy - 3, 7.5, 0, TAU); c.moveTo(17.5, hy - 3); c.arc(10, hy - 3, 7.5, 0, TAU); c.moveTo(-2.5, hy - 4); c.lineTo(2.5, hy - 4); c.stroke(); ink(c); }
    // hats
    if (o.hat === 'peaked') {    // a keeper's or captain's cap
      c.fillStyle = o.hatTop || o.hatColor; c.beginPath(); c.moveTo(-31, hy - 22); c.bezierCurveTo(-40, hy - 50, 40, hy - 50, 31, hy - 22); c.closePath(); c.fill(); c.stroke();
      c.fillStyle = o.hatColor; rrect(c, -30, hy - 28, 60, 12, 3); c.fill(); c.stroke();
      c.fillStyle = '#18141a'; c.beginPath(); c.moveTo(-28, hy - 17); c.quadraticCurveTo(0, hy - 6, 28, hy - 17); c.quadraticCurveTo(0, hy - 13, -28, hy - 17); c.fill(); c.stroke();
      c.fillStyle = GOLD2; c.beginPath(); c.ellipse(0, hy - 30, 7, 6, 0, 0, TAU); c.fill(); c.lineWidth = 1.4; c.stroke();
      c.strokeStyle = '#7a5a1a'; c.lineWidth = 1.4; c.beginPath(); c.moveTo(0, hy - 34); c.lineTo(0, hy - 26); c.moveTo(-3, hy - 31); c.lineTo(3, hy - 31); c.stroke(); ink(c);
    }
    if (o.hat === 'souwester') { // the yellow storm hat, brim longer at the back
      c.fillStyle = o.hatColor; c.beginPath(); c.moveTo(-44, hy - 6); c.quadraticCurveTo(-36, hy - 22, -26, hy - 22); c.bezierCurveTo(-26, hy - 54, 26, hy - 54, 26, hy - 22); c.quadraticCurveTo(36, hy - 22, 44, hy - 6); c.quadraticCurveTo(0, hy - 18, -44, hy - 6); c.closePath(); c.fill(); c.stroke();
      c.strokeStyle = 'rgba(0,0,0,.25)'; c.lineWidth = 1.4; c.beginPath(); c.moveTo(-26, hy - 24); c.quadraticCurveTo(0, hy - 30, 26, hy - 24); c.stroke(); ink(c);
    }
    if (o.hat === 'watchcap') {  // a rolled knit cap
      c.fillStyle = o.hatColor; c.beginPath(); c.moveTo(-30, hy - 14); c.bezierCurveTo(-32, hy - 52, 32, hy - 52, 30, hy - 14); c.closePath(); c.fill(); c.stroke();
      rrect(c, -31, hy - 22, 62, 12, 5); c.fill(); c.stroke();
      c.strokeStyle = 'rgba(0,0,0,.22)'; c.lineWidth = 1.2; for (let x = -24; x <= 24; x += 6) { c.beginPath(); c.moveTo(x, hy - 24); c.lineTo(x * 0.8, hy - 44); c.stroke(); } ink(c);
    }
    if (o.hat === 'sunhat') {    // a straw garden hat with a ribbon
      c.fillStyle = '#e8cf8a'; c.beginPath(); c.ellipse(0, hy - 22, 52, 12, 0, 0, TAU); c.fill(); c.stroke();
      c.beginPath(); c.moveTo(-26, hy - 22); c.bezierCurveTo(-28, hy - 52, 28, hy - 52, 26, hy - 22); c.closePath(); c.fill(); c.stroke();
      c.strokeStyle = o.band || '#3a7a9a'; c.lineWidth = 5; c.beginPath(); c.moveTo(-26, hy - 27); c.quadraticCurveTo(0, hy - 33, 26, hy - 27); c.stroke(); ink(c);
      c.strokeStyle = 'rgba(120,90,40,.35)'; c.lineWidth = 1; for (let r = 30; r < 52; r += 6) { c.beginPath(); c.ellipse(0, hy - 22, r, r * 0.23, 0, 0.1, Math.PI - 0.1); c.stroke(); } ink(c);
    }
    if (o.hat === 'beret') { c.save(); c.translate(4, hy - 30); c.rotate(-0.14); c.fillStyle = o.hatColor; c.beginPath(); c.ellipse(0, 0, 36, 13, 0, 0, TAU); c.fill(); c.stroke(); c.beginPath(); c.moveTo(-2, -12); c.lineTo(0, -20); c.lineTo(3, -12); c.closePath(); c.fill(); c.stroke(); c.restore(); }
    if (o.hat === 'cloche') {
      c.fillStyle = o.hatColor; c.beginPath(); c.moveTo(-33, hy - 4); c.bezierCurveTo(-36, hy - 50, 36, hy - 50, 33, hy - 4); c.quadraticCurveTo(0, hy - 16, -33, hy - 4); c.closePath(); c.fill(); c.stroke();
      c.strokeStyle = o.band || 'rgba(255,255,255,.4)'; c.lineWidth = 4; c.beginPath(); c.moveTo(-31, hy - 14); c.quadraticCurveTo(0, hy - 26, 31, hy - 14); c.stroke(); ink(c);
      if (o.flowers) { for (const [x, y, col] of [[20, hy - 22, '#e8b4b8'], [27, hy - 17, '#f0d36a'], [21, hy - 13, '#e8b4b8']]) { c.fillStyle = col; c.beginPath(); c.arc(x, y, 4.5, 0, TAU); c.fill(); c.stroke(); } }
      if (o.veil) { c.save(); c.globalAlpha = 0.55; c.strokeStyle = '#18141a'; c.lineWidth = 0.9; c.beginPath(); c.moveTo(-31, hy - 10); c.quadraticCurveTo(0, hy + 14, 31, hy - 10); c.stroke(); for (let x = -28; x <= 28; x += 5) { c.beginPath(); c.moveTo(x, hy - 12); c.lineTo(x * 0.8, hy + 6 - Math.abs(x) * 0.3); c.stroke(); } for (let y = hy - 8; y < hy + 6; y += 5) { c.beginPath(); c.moveTo(-28 + (y - hy) * 0.3, y); c.lineTo(28 - (y - hy) * 0.3, y); c.stroke(); } c.restore(); ink(c); }
    }
    if (o.hat === 'cap') { c.fillStyle = o.hatColor; c.beginPath(); c.moveTo(-30, hy - 14); c.bezierCurveTo(-30, hy - 42, 26, hy - 46, 32, hy - 22); c.quadraticCurveTo(4, hy - 26, -30, hy - 14); c.closePath(); c.fill(); c.stroke(); c.beginPath(); c.moveTo(4, hy - 24); c.quadraticCurveTo(34, hy - 28, 40, hy - 16); c.quadraticCurveTo(24, hy - 12, 4, hy - 18); c.closePath(); c.fill(); c.stroke(); }
    if (o.hat === 'toque' || o.hat === 'bakercap') {
      c.fillStyle = '#fdfcf8'; rrect(c, -24, hy - 46, 48, 16, 3); c.fill(); c.stroke();
      const puffs = o.hat === 'toque' ? [[-18, hy - 60, 15], [18, hy - 60, 15], [0, hy - 70, 18]] : [[-10, hy - 50, 12], [10, hy - 50, 12]];
      for (const [x, y, r] of puffs) { c.beginPath(); c.arc(x, y, r, 0, TAU); c.fill(); c.stroke(); }
      c.fillRect(-22, o.hat === 'toque' ? hy - 58 : hy - 50, 44, o.hat === 'toque' ? 14 : 6);
    }
    if (o.hat === 'feathers') {   // a cabaret headband with a plume
      c.strokeStyle = o.band || '#e8c35a'; c.lineWidth = 4; c.beginPath(); c.moveTo(-29, hy - 14); c.quadraticCurveTo(0, hy - 30, 29, hy - 14); c.stroke(); ink(c);
      c.fillStyle = '#e8c35a'; c.beginPath(); c.arc(16, hy - 22, 5, 0, TAU); c.fill(); c.stroke();
      c.fillStyle = o.plume || '#f6efe2'; c.beginPath(); c.moveTo(16, hy - 26); c.bezierCurveTo(24, hy - 70, 52, hy - 74, 48, hy - 50); c.bezierCurveTo(40, hy - 58, 30, hy - 44, 16, hy - 26); c.fill(); c.stroke();
    }
    c.restore();
  }
  function hand(c, x, y, skin) { c.save(); ink(c); c.fillStyle = skin || '#f2d4b6'; c.beginPath(); c.ellipse(x, y, 9, 8, 0, 0, TAU); c.fill(); c.stroke(); c.restore(); }

  // who sits on each court card
  const COURT = {
    K: {
      H: { name: 'The Keeper', o: { top: 'peacoat', coat: '#1d3350', dark: '#0f1e33', shirt: '#f1e6cc', hat: 'peaked', hatColor: '#1d3350', hatTop: '#1d3350', hair: '#c8c4c0', beard: true, beardColor: '#e8e4de', brow: '#b8b4b0', skin: '#eec3a2', smile: true }, prop: 'lantern' },
      D: { name: 'The Lobsterman', o: { top: 'oilskin', coat: '#e0a81e', dark: '#9a6e0e', hat: 'souwester', hatColor: '#e8b425', hair: '#7a5a3a', beard: true, beardColor: '#8a6a4a', skin: '#d9a07a' }, prop: 'lobster' },
      S: { name: 'The Captain', o: { top: 'peacoat', coat: '#1d3350', dark: '#0f1e33', hat: 'peaked', hatColor: '#1d3350', hatTop: '#f6f2e8', hair: '#5a5a60', mustache: 'walrus', mustacheColor: '#9a9a9e', brow: '#6a6a70', skin: '#f1d0b5', stern: true }, prop: 'spyglass' },
      C: { name: 'The Boatwright', o: { top: 'sweater', coat: '#2f5a44', dark: '#183426', hat: 'watchcap', hatColor: '#8e2a22', hair: '#5a3a22', beard: true, skin: '#c98f66', smile: true }, prop: 'mallet' }
    },
    Q: {
      H: { name: 'The Pie Baker', o: { lady: true, top: 'apron', coat: '#a82a22', dark: '#c8b8a0', accent: '#fbf8f0', hairStyle: 'bun', hair: '#8a4a22', neck: 'pearls', lip: '#b3283a', skin: '#f6dbc4', smile: true }, prop: 'pie' },
      D: { name: 'The Quilter', o: { lady: true, top: 'cardigan', coat: '#d9a12a', dark: '#7a5212', accent: '#fbf8f0', hairStyle: 'bun', hair: '#d8d4d0', glasses: 'round', lip: '#b05050', skin: '#eec8a8', smile: true }, prop: 'yarn' },
      S: { name: "The Captain's Lady", o: { lady: true, top: 'dress', coat: '#1d3350', dark: '#0f1e33', hairStyle: 'waves', hair: '#2a1e18', neck: 'shawl', neckColor: '#5a6a80', lip: '#8e2a3a', skin: '#f1d0b5', earrings: '#e8c35a', stern: true }, prop: 'bottleship' },
      C: { name: 'The Gardener', o: { lady: true, top: 'blouse', coat: '#3a7a5a', dark: '#183426', accent: '#fbf8f0', hairStyle: 'bob', hair: '#b5651d', hat: 'sunhat', band: '#3a6a9a', lip: '#c04a4a', skin: '#d6a27a', smile: true }, prop: 'hydrangea' }
    },
    J: {
      H: { name: 'The Deckhand', o: { top: 'stripes', stripe: '#a82a22', hat: 'watchcap', hatColor: '#1d3350', hair: '#c88a3a', hairStyle: 'short', skin: '#f2d4b6', smile: true }, prop: 'rope' },
      D: { name: 'The Clam Digger', o: { top: 'shirt', coat: '#d9a12a', dark: '#8a5e12', braces: '#4a3020', hat: 'cap', hatColor: '#6a5a3a', hair: '#3a2418', hairStyle: 'short', skin: '#a8704a', smile: true }, prop: 'rake' },
      S: { name: 'The Ferryman', o: { top: 'sweater', coat: '#1d3350', dark: '#0f1e33', hat: 'cap', hatColor: '#2a2a32', hair: '#4a2e1c', hairStyle: 'short', mustache: 'pencil', skin: '#e3b48e' }, prop: 'oar' },
      C: { name: 'The Fiddler', o: { top: 'waistcoat', coat: '#2f5a44', dark: '#183426', vest: '#5a3a22', neck: 'kerchief', neckColor: '#a82a22', hairStyle: 'side', hair: '#6b3f22', skin: '#f1d0b5', smile: true }, prop: 'fiddle' }
    }
  };

  // what each court figure holds
  function courtProp(c, prop, o) {
    c.save(); ink(c); c.lineWidth = 2.4;
    const sk = o.skin;
    switch (prop) {
      case 'lantern': {
        c.save(); c.translate(56, -60);
        const g = c.createRadialGradient(0, 0, 2, 0, 0, 40); g.addColorStop(0, 'rgba(255,230,140,.7)'); g.addColorStop(1, 'rgba(255,230,140,0)'); c.fillStyle = g; c.beginPath(); c.arc(0, 0, 40, 0, TAU); c.fill();
        ink(c); c.lineWidth = 2;
        c.beginPath(); c.arc(0, -30, 9, Math.PI, 0); c.stroke();
        c.fillStyle = '#c9a24c'; rrect(c, -14, -24, 28, 7, 2); c.fill(); c.stroke(); rrect(c, -15, 16, 30, 8, 2); c.fill(); c.stroke();
        c.fillStyle = 'rgba(255,240,190,.9)'; c.beginPath(); c.moveTo(-10, -17); c.quadraticCurveTo(-17, 0, -10, 16); c.lineTo(10, 16); c.quadraticCurveTo(17, 0, 10, -17); c.closePath(); c.fill(); c.stroke();
        c.fillStyle = '#f2902a'; c.beginPath(); c.moveTo(0, -8); c.quadraticCurveTo(5, 2, 0, 6); c.quadraticCurveTo(-5, 2, 0, -8); c.fill();
        c.restore(); hand(c, 56, -88, sk); break;
      }
      case 'lobster': {
        c.save(); c.translate(54, -58); c.rotate(-0.2);
        c.fillStyle = '#c4301e'; ink(c); c.lineWidth = 2;
        c.beginPath(); c.ellipse(0, 0, 10, 20, 0, 0, TAU); c.fill(); c.stroke();
        for (let i = 0; i < 4; i++) { c.beginPath(); c.ellipse(0, 22 + i * 8, 8 - i, 5, 0, 0, TAU); c.fill(); c.stroke(); }
        c.beginPath(); c.moveTo(-6, 52); c.lineTo(-12, 62); c.lineTo(12, 62); c.lineTo(6, 52); c.closePath(); c.fill(); c.stroke();
        for (const sx of [-1, 1]) { c.beginPath(); c.moveTo(sx * 6, -16); c.quadraticCurveTo(sx * 18, -30, sx * 16, -42); c.stroke();
          c.beginPath(); c.ellipse(sx * 17, -50, 8, 13, sx * 0.3, 0, TAU); c.fill(); c.stroke(); c.beginPath(); c.moveTo(sx * 17, -62); c.lineTo(sx * 17, -48); c.stroke(); }
        c.strokeStyle = '#c4301e'; c.lineWidth = 1.2; for (const sx of [-1, 1]) { c.beginPath(); c.moveTo(sx * 3, -19); c.quadraticCurveTo(sx * 30, -60, sx * 10, -80); c.stroke(); }
        c.restore(); hand(c, 52, -34, sk); break;
      }
      case 'spyglass':
        c.save(); c.translate(52, -40); c.rotate(-1.15);
        c.fillStyle = '#c9a24c'; rrect(c, -6, -8, 30, 16, 3); c.fill(); c.stroke();
        c.fillStyle = '#5a3a22'; rrect(c, 24, -7, 26, 14, 3); c.fill(); c.stroke();
        c.fillStyle = '#c9a24c'; rrect(c, 50, -6, 22, 12, 3); c.fill(); c.stroke();
        c.fillStyle = '#9ab8d0'; c.beginPath(); c.ellipse(72, 0, 3, 6, 0, 0, TAU); c.fill(); c.stroke();
        c.restore(); hand(c, 52, -36, sk); break;
      case 'mallet':
        c.strokeStyle = '#9a6a3a'; c.lineWidth = 7; c.beginPath(); c.moveTo(54, -14); c.lineTo(62, -86); c.stroke();
        ink(c); c.fillStyle = '#c08a50'; c.save(); c.translate(63, -94); c.rotate(0.11); rrect(c, -22, -12, 44, 24, 5); c.fill(); c.stroke();
        c.strokeStyle = 'rgba(80,50,20,.4)'; c.lineWidth = 1.2; for (let x = -16; x < 20; x += 8) { c.beginPath(); c.moveTo(x, -10); c.lineTo(x, 10); c.stroke(); } c.restore();
        hand(c, 56, -32, sk); break;
      case 'pie':
        c.save(); c.translate(-46, -46);
        c.fillStyle = '#e6ded0'; c.beginPath(); c.ellipse(0, 6, 36, 12, 0, 0, TAU); c.fill(); c.stroke();
        c.fillStyle = '#d9a35a'; c.beginPath(); c.ellipse(0, 0, 30, 11, 0, 0, TAU); c.fill(); c.stroke();
        c.fillStyle = '#3a2a5a'; c.save(); c.beginPath(); c.ellipse(0, 0, 25, 8, 0, 0, TAU); c.clip(); c.fillRect(-26, -9, 52, 18);
        c.strokeStyle = '#e0b06a'; c.lineWidth = 3; for (let x = -24; x < 26; x += 10) { c.beginPath(); c.moveTo(x, -9); c.lineTo(x + 6, 9); c.stroke(); c.beginPath(); c.moveTo(x + 6, -9); c.lineTo(x, 9); c.stroke(); } c.restore();
        c.strokeStyle = 'rgba(255,255,255,.6)'; c.lineWidth = 1.6; for (const x of [-8, 6]) { c.beginPath(); c.moveTo(x, -14); c.quadraticCurveTo(x + 4, -24, x, -32); c.stroke(); }
        c.restore(); hand(c, -76, -40, sk); hand(c, -16, -40, sk); break;
      case 'yarn':
        c.save(); c.translate(-48, -34);
        c.fillStyle = '#a82a22'; c.beginPath(); c.arc(0, 0, 18, 0, TAU); c.fill(); c.stroke();
        c.strokeStyle = '#6a1712'; c.lineWidth = 1.4; for (let i = 0; i < 5; i++) { c.beginPath(); c.ellipse(0, 0, 17 - i * 3, 8 + i, i * 0.6, 0, TAU); c.stroke(); }
        ink(c); c.strokeStyle = '#c9a24c'; c.lineWidth = 3; c.beginPath(); c.moveTo(-14, -30); c.lineTo(20, 14); c.moveTo(14, -32); c.lineTo(-18, 16); c.stroke();
        c.restore();
        c.strokeStyle = '#a82a22'; c.lineWidth = 1.6; c.beginPath(); c.moveTo(-30, -24); c.bezierCurveTo(0, 0, 30, -10, 46, -20); c.stroke();
        hand(c, 46, -22, sk); break;
      case 'bottleship':
        c.save(); c.translate(-50, -44); c.rotate(-0.15);
        c.fillStyle = 'rgba(190,220,210,.55)'; c.beginPath(); c.moveTo(-26, -12); c.lineTo(18, -12); c.quadraticCurveTo(26, -12, 28, -4); c.lineTo(36, -4); c.lineTo(36, 4); c.lineTo(28, 4); c.quadraticCurveTo(26, 12, 18, 12); c.lineTo(-26, 12); c.quadraticCurveTo(-32, 0, -26, -12); c.fill(); c.stroke();
        c.fillStyle = '#8a5a30'; c.fillRect(36, -3, 6, 6);
        c.fillStyle = '#5a3a22'; c.beginPath(); c.moveTo(-18, 6); c.lineTo(14, 6); c.lineTo(10, 10); c.lineTo(-14, 10); c.closePath(); c.fill();
        c.fillStyle = '#fbf8f0'; c.lineWidth = 1; c.beginPath(); c.moveTo(-6, 5); c.lineTo(-6, -9); c.lineTo(4, 5); c.closePath(); c.fill(); c.stroke(); c.beginPath(); c.moveTo(-8, 5); c.lineTo(-8, -6); c.lineTo(-16, 5); c.closePath(); c.fill(); c.stroke();
        c.restore(); hand(c, -50, -30, sk); break;
      case 'hydrangea':
        c.strokeStyle = '#3f6a3a'; c.lineWidth = 3; c.beginPath(); c.moveTo(46, -10); c.lineTo(52, -56); c.stroke();
        for (const [x, y, r] of [[38, -64, 14], [60, -66, 14], [50, -80, 15]]) { const R = rng(x * 7 + y); for (let i = 0; i < 16; i++) { const a = R() * TAU, d = R() * r; c.fillStyle = ['#6a8ad0', '#8a7ac8', '#a0b4e8', '#7a9ae0'][i % 4]; c.beginPath(); c.arc(x + Math.cos(a) * d, y + Math.sin(a) * d, 4.2, 0, TAU); c.fill(); c.lineWidth = 0.8; c.stroke(); } }
        c.fillStyle = '#4f8a4a'; c.lineWidth = 1.5; c.beginPath(); c.ellipse(64, -46, 10, 5, -0.5, 0, TAU); c.fill(); c.stroke();
        hand(c, 48, -22, sk); break;
      case 'rope':
        c.save(); c.translate(-50, -36);
        for (let i = 0; i < 4; i++) rope(c, 0, 0, 24 - i * 5, 18 - i * 4, 5, '#d8c08a', '#9a7a4a');
        c.restore(); hand(c, -50, -54, sk); break;
      case 'rake':
        c.strokeStyle = '#9a6a3a'; c.lineWidth = 5; c.beginPath(); c.moveTo(56, -10); c.lineTo(60, -110); c.stroke();
        ink(c); c.fillStyle = '#7a7a80'; c.fillRect(44, -116, 34, 6); c.strokeRect(44, -116, 34, 6);
        c.strokeStyle = '#7a7a80'; c.lineWidth = 2.4; for (let x = 46; x <= 76; x += 6) { c.beginPath(); c.moveTo(x, -110); c.quadraticCurveTo(x + 2, -96, x - 2, -92); c.stroke(); }
        c.save(); c.translate(-46, -26); ink(c); c.fillStyle = '#8a8a90'; c.beginPath(); c.moveTo(-18, -20); c.lineTo(18, -20); c.lineTo(14, 16); c.lineTo(-14, 16); c.closePath(); c.fill(); c.stroke();
        c.fillStyle = '#e8e0d0'; for (const [x, y] of [[-8, -22], [4, -24], [12, -20], [-2, -18]]) { c.beginPath(); c.ellipse(x, y, 6, 4, 0.3, 0, TAU); c.fill(); c.lineWidth = 1.2; c.stroke(); }
        c.restore(); hand(c, 58, -40, sk); break;
      case 'oar':
        c.save(); c.translate(56, -60); c.rotate(0.12);
        c.fillStyle = '#c08a50'; ink(c); c.fillRect(-3, -40, 6, 100); c.strokeRect(-3, -40, 6, 100);
        c.beginPath(); c.moveTo(-3, -40); c.quadraticCurveTo(-12, -60, -9, -82); c.quadraticCurveTo(0, -90, 9, -82); c.quadraticCurveTo(12, -60, 3, -40); c.closePath(); c.fill(); c.stroke();
        c.fillStyle = '#a82a22'; c.fillRect(-9, -84, 18, 8);
        c.restore(); hand(c, 56, -40, sk); break;
      case 'fiddle':
        c.save(); c.translate(-44, -48); c.rotate(-0.55);
        c.fillStyle = '#b5651d'; ink(c); c.lineWidth = 2;
        c.beginPath(); c.moveTo(0, -30); c.bezierCurveTo(16, -30, 18, -14, 10, -8); c.bezierCurveTo(20, 0, 20, 28, 0, 30); c.bezierCurveTo(-20, 28, -20, 0, -10, -8); c.bezierCurveTo(-18, -14, -16, -30, 0, -30); c.fill(); c.stroke();
        c.fillStyle = '#3a2418'; c.fillRect(-3, -60, 6, 34); c.beginPath(); c.arc(0, -63, 5, 0, TAU); c.fill();
        c.strokeStyle = '#2a1a10'; c.lineWidth = 1.4; for (const sx of [-1, 1]) { c.beginPath(); c.moveTo(sx * 6, -4); c.quadraticCurveTo(sx * 9, 2, sx * 5, 8); c.stroke(); }
        c.strokeStyle = '#e8e0d0'; c.lineWidth = 0.8; for (const x of [-2, -0.7, 0.7, 2]) { c.beginPath(); c.moveTo(x, -58); c.lineTo(x, 20); c.stroke(); }
        c.restore();
        c.strokeStyle = '#5a3a22'; c.lineWidth = 2.4; c.beginPath(); c.moveTo(-90, -60); c.lineTo(20, -30); c.stroke();
        hand(c, 20, -30, sk); hand(c, -60, -78, sk); break;
    }
    c.restore();
  }

  // painted beadboard behind a court figure
  function wallpaper(c, x, y, w, h, s) {
    const p = PAL[s];
    c.fillStyle = p.wall; c.fillRect(x, y, w, h);
    c.fillStyle = 'rgba(0,0,0,.06)'; for (let xx = x; xx < x + w; xx += 18) c.fillRect(xx, y, 2, h);
    c.fillStyle = 'rgba(255,255,255,.35)'; for (let xx = x + 3; xx < x + w; xx += 18) c.fillRect(xx, y, 3, h);
    c.fillStyle = s === 'H' || s === 'D' ? 'rgba(178,42,34,.08)' : 'rgba(22,32,44,.09)';
    for (let yy = y + 14; yy < y + h; yy += 32) for (let xx = x + 17 + ((yy / 32) % 2) * 12; xx < x + w; xx += 26) star(c, xx, yy, 6, c.fillStyle);
  }

  function courtCard(c, rank, s) {
    const color = PAL[s].ink, w = CW, h = CH;
    const fx = 76, fy = 64, fw = w - 152, fh = h - 128;
    const who = COURT[rank][s];
    c.save(); rrect(c, fx, fy, fw, fh, 12); c.clip();
    wallpaper(c, fx, fy, fw, fh, s);
    const glow = c.createRadialGradient(fx + fw / 2, fy + fh * 0.42, 10, fx + fw / 2, fy + fh * 0.42, fh * 0.6);
    glow.addColorStop(0, 'rgba(255,250,232,.8)'); glow.addColorStop(1, 'rgba(255,248,225,0)'); c.fillStyle = glow; c.fillRect(fx, fy, fw, fh);
    // a round porthole window behind the head, with the sea in it
    { const px = fx + fw / 2, py = fy + 96, pr = fw / 2 - 30;
      c.save(); c.beginPath(); c.arc(px, py, pr, 0, TAU); c.clip();
      const sky = c.createLinearGradient(0, py - pr, 0, py + pr); sky.addColorStop(0, '#cfe0ea'); sky.addColorStop(0.55, '#e8eef0'); sky.addColorStop(0.56, '#5a86a8'); sky.addColorStop(1, '#2a4a6a');
      c.fillStyle = sky; c.fillRect(px - pr, py - pr, pr * 2, pr * 2);
      if (rank === 'K' && s === 'H') lighthouse(c, px + pr * 0.45, py + pr * 0.12, pr * 0.9);
      c.restore();
      c.strokeStyle = '#a8843a'; c.lineWidth = 6; c.beginPath(); c.arc(px, py, pr, 0, TAU); c.stroke();
      c.strokeStyle = '#6a5020'; c.lineWidth = 1.4; c.beginPath(); c.arc(px, py, pr + 3, 0, TAU); c.stroke();
      c.fillStyle = '#6a5020'; for (let i = 0; i < 8; i++) { const a = i / 8 * TAU; c.beginPath(); c.arc(px + Math.cos(a) * pr, py + Math.sin(a) * pr, 2, 0, TAU); c.fill(); }
    }
    const fit = 1.48, o = who.o;
    c.save(); c.translate(fx + fw / 2, fy + fh + 2); c.scale(fit, fit);
    bust(c, o); courtProp(c, who.prop, o);
    c.restore();
    // the name plate
    c.fillStyle = 'rgba(255,250,238,.92)'; rrect(c, fx + 14, fy + fh - 30, fw - 28, 22, 6); c.fill();
    c.strokeStyle = GOLD; c.lineWidth = 1.2; c.stroke();
    c.fillStyle = color; c.font = `italic 700 ${who.name.length > 15 ? 12 : 13}px ${DISP}`; c.textAlign = 'center'; c.fillText(who.name, fx + fw / 2, fy + fh - 14);
    c.restore();
    c.strokeStyle = GOLD; c.lineWidth = 3; rrect(c, fx, fy, fw, fh, 12); c.stroke();
    c.lineWidth = 1; rrect(c, fx + 5, fy + 5, fw - 10, fh - 10, 8); c.stroke();
    suit(c, s, fx + 22, fy + 26, 26, color, false);
    suit(c, s, fx + fw - 22, fy + 26, 26, color, false);
  }

  const PIPS = {
    2: [[.5, 0], [.5, 1]], 3: [[.5, 0], [.5, .5], [.5, 1]], 4: [[0, 0], [1, 0], [0, 1], [1, 1]],
    5: [[0, 0], [1, 0], [.5, .5], [0, 1], [1, 1]], 6: [[0, 0], [1, 0], [0, .5], [1, .5], [0, 1], [1, 1]],
    7: [[0, 0], [1, 0], [.5, .25], [0, .5], [1, .5], [0, 1], [1, 1]],
    8: [[0, 0], [1, 0], [.5, .25], [0, .5], [1, .5], [.5, .75], [0, 1], [1, 1]],
    9: [[0, 0], [1, 0], [0, 1 / 3], [1, 1 / 3], [.5, .5], [0, 2 / 3], [1, 2 / 3], [0, 1], [1, 1]],
    10: [[0, 0], [1, 0], [.5, 1 / 6], [0, 1 / 3], [1, 1 / 3], [0, 2 / 3], [1, 2 / 3], [.5, 5 / 6], [0, 1], [1, 1]]
  };
  function pipCard(c, n, s) {
    const color = PAL[s].ink, x0 = 114, x1 = CW - 114, y0 = 104, y1 = CH - 104;
    for (const [px, py] of PIPS[n]) {
      const x = x0 + (x1 - x0) * px, y = y0 + (y1 - y0) * py;
      if (py > 0.5) { c.save(); c.translate(x, y); c.rotate(Math.PI); suit(c, s, 0, 0, 56, color); c.restore(); }
      else suit(c, s, x, y, 56, color);
    }
  }
  function aceCard(c, s) {
    const color = PAL[s].ink, cx = CW / 2, cy = CH / 2;
    // a coil of rope round a compass rose
    rope(c, cx, cy, 104, 132, 7, '#d8c08a', '#9a7a4a');
    c.save(); c.globalAlpha = 0.18; for (let i = 0; i < 4; i++) { c.save(); c.translate(cx, cy); c.rotate(i * Math.PI / 4); star(c, 0, 0, 96, GOLD); c.restore(); } c.restore();
    suit(c, s, cx, cy, 132, color);
  }

  /* ---------- the card back: harbour navy, a rope border and the light ---------- */
  function backCard() {
    const cv = canvas(CW, CH), c = cv.getContext('2d');
    rrect(c, 1, 1, CW - 2, CH - 2, 22); c.fillStyle = '#f6efe0'; c.fill();
    rrect(c, 12, 12, CW - 24, CH - 24, 14); const g = c.createLinearGradient(0, 0, 0, CH); g.addColorStop(0, '#244068'); g.addColorStop(1, NAVY2); c.fillStyle = g; c.fill();
    c.save(); c.clip();
    // a net of little compass stars
    c.strokeStyle = 'rgba(232,207,138,.16)'; c.lineWidth = 1.2;
    for (let k = -CH; k < CW + CH; k += 30) { c.beginPath(); c.moveTo(k, 0); c.lineTo(k + CH, CH); c.moveTo(k, CH); c.lineTo(k + CH, 0); c.stroke(); }
    for (let y = 12; y < CH; y += 30) for (let x = 12 + ((y - 12) / 30 % 2) * 15; x < CW; x += 30) star(c, x, y + 15, 5, 'rgba(232,207,138,.35)');
    c.restore();
    ropeRect(c, 28, 28, CW - 56, CH - 56, 12, 6, '#d8c08a', '#8a6a3a');
    const cx = CW / 2, cy = CH / 2;
    c.save(); c.beginPath(); c.ellipse(cx, cy, 94, 124, 0, 0, TAU); c.clip();
    const sky = c.createLinearGradient(0, cy - 124, 0, cy + 124); sky.addColorStop(0, '#0e1a30'); sky.addColorStop(0.62, '#3a4a78'); sky.addColorStop(0.63, '#1a2c48'); sky.addColorStop(1, '#0c1626'); c.fillStyle = sky; c.fillRect(cx - 100, cy - 130, 200, 260);
    const R = rng(11); for (let i = 0; i < 40; i++) { c.fillStyle = `rgba(255,250,230,${R() * 0.8})`; c.beginPath(); c.arc(cx - 94 + R() * 188, cy - 124 + R() * 140, R() * 1.4, 0, TAU); c.fill(); }
    c.fillStyle = '#2a2a30'; c.beginPath(); c.moveTo(cx - 100, cy + 50); c.quadraticCurveTo(cx - 30, cy + 26, cx + 10, cy + 44); c.quadraticCurveTo(cx + 60, cy + 30, cx + 100, cy + 50); c.lineTo(cx + 100, cy + 130); c.lineTo(cx - 100, cy + 130); c.closePath(); c.fill();
    lighthouse(c, cx - 10, cy + 40, 150, { beam: true });
    c.strokeStyle = 'rgba(255,240,180,.35)'; c.lineWidth = 1; for (let i = 0; i < 6; i++) { c.beginPath(); c.moveTo(cx - 90 + i * 30, cy + 66 + i % 2 * 6); c.lineTo(cx - 70 + i * 30, cy + 66 + i % 2 * 6); c.stroke(); }
    c.restore();
    c.strokeStyle = '#e8c35a'; c.lineWidth = 3; c.beginPath(); c.ellipse(cx, cy, 94, 124, 0, 0, TAU); c.stroke();
    c.lineWidth = 1; c.beginPath(); c.ellipse(cx, cy, 100, 130, 0, 0, TAU); c.stroke();
    c.fillStyle = '#f2d68e'; c.font = `italic 700 22px ${DISP}`; c.textAlign = 'center'; c.fillText('Cribbage', cx, CH - 52);
    c.font = `700 12px ${DISP}`; c.fillText('SEAL HARBOR LIGHT', cx, 70);
    rrect(c, 1, 1, CW - 2, CH - 2, 22); c.strokeStyle = 'rgba(50,45,40,.45)'; c.lineWidth = 2; c.stroke();
    return cv;
  }

  /* ---------- public: a card face (r: 1 = ace ... 13 = king) ---------- */
  const RANK = { 1: 'A', 11: 'J', 12: 'Q', 13: 'K' };
  function face(card) {
    const cv = canvas(CW, CH), c = cv.getContext('2d');
    paper(c, CW, CH, card.r * 7 + 'SHDC'.indexOf(card.s) * 131);
    const color = PAL[card.s].ink;
    border(c, CW, CH, card.s === 'H' || card.s === 'D' ? 'rgba(178,42,34,.7)' : GOLD);
    const label = RANK[card.r] || String(card.r);
    index(c, label, card.s, color, CW, CH);
    if (card.r === 1) aceCard(c, card.s);
    else if (card.r > 10) courtCard(c, label, card.s);
    else pipCard(c, card.r, card.s);
    return cv;
  }

  /* ---------- props on the cottage table, seen from above ---------- */
  function mug(size) {   // a speckled enamel mug of cocoa with marshmallows
    const cv = canvas(size, size), c = cv.getContext('2d'), k = size / 220; c.scale(k, k);
    c.fillStyle = 'rgba(0,0,0,.35)'; c.beginPath(); c.ellipse(116, 122, 82, 80, 0, 0, TAU); c.fill();
    c.strokeStyle = INK; c.lineWidth = 3;
    // handle
    c.fillStyle = '#2a5a8a'; c.beginPath(); c.ellipse(178, 108, 26, 16, 0.2, 0, TAU); c.fill(); c.stroke();
    c.fillStyle = '#6a4a2a'; c.beginPath(); c.ellipse(180, 108, 13, 6, 0.2, 0, TAU); c.fill();
    // body (rim), speckled blue enamel
    let g = c.createRadialGradient(96, 92, 20, 108, 108, 76); g.addColorStop(0, '#4a82b8'); g.addColorStop(1, '#1f4a78');
    c.fillStyle = g; c.beginPath(); c.arc(108, 108, 72, 0, TAU); c.fill(); c.stroke();
    const R = rng(21); c.fillStyle = 'rgba(255,255,255,.7)'; for (let i = 0; i < 90; i++) { const a = R() * TAU, d = 62 + R() * 9; c.beginPath(); c.arc(108 + Math.cos(a) * d, 108 + Math.sin(a) * d, 0.8 + R() * 1.2, 0, TAU); c.fill(); }
    c.strokeStyle = '#1a1a20'; c.lineWidth = 4; c.beginPath(); c.arc(108, 108, 70, 0, TAU); c.stroke();
    c.fillStyle = '#f2ece0'; c.beginPath(); c.arc(108, 108, 62, 0, TAU); c.fill();
    g = c.createRadialGradient(100, 100, 6, 108, 108, 56); g.addColorStop(0, '#9a6232'); g.addColorStop(0.8, '#6a3c1a'); g.addColorStop(1, '#4a2810');
    c.fillStyle = g; c.beginPath(); c.arc(108, 108, 56, 0, TAU); c.fill();
    // marshmallows
    for (const [x, y, r] of [[90, 92, 0.3], [118, 86, -0.4], [128, 114, 0.6], [100, 124, 0.1], [108, 104, -0.9]]) {
      c.save(); c.translate(x, y); c.rotate(r); c.fillStyle = '#fbf8f2'; c.strokeStyle = 'rgba(120,90,60,.6)'; c.lineWidth = 1.6; rrect(c, -10, -9, 20, 18, 5); c.fill(); c.stroke();
      c.fillStyle = 'rgba(220,170,110,.4)'; c.fillRect(-10, 3, 20, 6); c.restore();
    }
    c.fillStyle = 'rgba(255,255,255,.28)'; c.beginPath(); c.ellipse(80, 66, 18, 6, -0.7, 0, TAU); c.fill();
    return cv;
  }
  function lantern(size) {   // the keeper's brass hurricane lantern, lit
    const cv = canvas(size, size), c = cv.getContext('2d'), k = size / 220; c.scale(k, k);
    let g = c.createRadialGradient(110, 112, 6, 110, 112, 108); g.addColorStop(0, 'rgba(255,214,120,.55)'); g.addColorStop(1, 'rgba(255,200,100,0)'); c.fillStyle = g; c.fillRect(0, 0, 220, 220);
    c.fillStyle = 'rgba(0,0,0,.35)'; c.beginPath(); c.ellipse(118, 200, 62, 14, 0, 0, TAU); c.fill();
    c.strokeStyle = INK; c.lineWidth = 3; c.lineJoin = 'round';
    // bail handle
    c.strokeStyle = '#7a5a1a'; c.lineWidth = 4; c.beginPath(); c.arc(110, 44, 30, Math.PI * 1.05, Math.PI * 1.95); c.stroke();
    c.strokeStyle = INK; c.lineWidth = 3;
    const brass = (x, y, w, h) => { const gg = c.createLinearGradient(x, 0, x + w, 0); gg.addColorStop(0, '#8a6a22'); gg.addColorStop(0.35, '#f0d27a'); gg.addColorStop(0.7, '#c9a24c'); gg.addColorStop(1, '#6a4e14'); c.fillStyle = gg; rrect(c, x, y, w, h, 5); c.fill(); c.stroke(); };
    brass(84, 30, 52, 14); brass(70, 44, 80, 16);
    // the glass globe
    g = c.createRadialGradient(108, 108, 4, 110, 110, 50); g.addColorStop(0, 'rgba(255,248,220,.95)'); g.addColorStop(0.5, 'rgba(255,226,150,.75)'); g.addColorStop(1, 'rgba(230,190,110,.55)');
    c.fillStyle = g; c.beginPath(); c.moveTo(80, 60); c.bezierCurveTo(52, 90, 52, 140, 82, 164); c.lineTo(138, 164); c.bezierCurveTo(168, 140, 168, 90, 140, 60); c.closePath(); c.fill(); c.stroke();
    c.strokeStyle = 'rgba(122,90,26,.9)'; c.lineWidth = 3; for (const x of [76, 144]) { c.beginPath(); c.moveTo(x + (x < 110 ? 6 : -6), 62); c.lineTo(x, 164); c.stroke(); }
    // the flame
    c.fillStyle = '#f2902a'; c.beginPath(); c.moveTo(110, 96); c.bezierCurveTo(124, 118, 120, 138, 110, 140); c.bezierCurveTo(100, 138, 96, 118, 110, 96); c.fill();
    c.fillStyle = '#fff2b0'; c.beginPath(); c.moveTo(110, 112); c.bezierCurveTo(116, 124, 115, 136, 110, 137); c.bezierCurveTo(105, 136, 104, 124, 110, 112); c.fill();
    c.fillStyle = 'rgba(255,255,255,.45)'; c.beginPath(); c.ellipse(86, 104, 6, 22, 0.15, 0, TAU); c.fill();
    c.strokeStyle = INK; c.lineWidth = 3; brass(68, 162, 84, 18); brass(60, 178, 100, 20);
    return cv;
  }

  /* ---------- the tabletop: scrubbed pine planks ---------- */
  function wood(W, H) {
    const cv = canvas(W, H), c = cv.getContext('2d'), img = c.createImageData(W, H), d = img.data;
    const P = new Uint8Array(512); { const R = rng(1852); const p = [...Array(256).keys()]; for (let i = 255; i > 0; i--) { const j = Math.floor(R() * (i + 1)); [p[i], p[j]] = [p[j], p[i]]; } for (let i = 0; i < 512; i++) P[i] = p[i & 255]; }
    const fade = t => t * t * t * (t * (t * 6 - 15) + 10), lerp = (a, b, t) => a + (b - a) * t;
    const grad = (h, x, y) => { const v = h & 3; return (v & 1 ? -x : x) + (v & 2 ? -y : y); };
    function noise(x, y) { const X = Math.floor(x) & 255, Y = Math.floor(y) & 255; x -= Math.floor(x); y -= Math.floor(y); const u = fade(x), v = fade(y); const a = P[X] + Y, b = P[X + 1] + Y; return lerp(lerp(grad(P[a], x, y), grad(P[b], x - 1, y), u), lerp(grad(P[a + 1], x, y - 1), grad(P[b + 1], x - 1, y - 1), u), v); }
    const plank = 150, R = rng(77), tones = [], knots = [];
    for (let i = 0; i < Math.ceil(H / plank) + 1; i++) { tones.push(0.85 + R() * 0.3); for (let k = 0; k < 2; k++) knots.push([R() * W, i * plank + 30 + R() * (plank - 60), 10 + R() * 14]); }
    for (let y = 0; y < H; y++) {
      const pi = Math.floor(y / plank), py = y - pi * plank, tone = tones[pi];
      for (let x = 0; x < W; x++) {
        let ny = y + noise(x / 420, y / 90 + pi * 9) * 9;
        for (const [kx, ky, kr] of knots) { const dx = (x - kx) / (kr * 3.2), dy = (y - ky) / kr, dd = dx * dx + dy * dy; if (dd < 6) ny += (ky - y) * 0.9 * Math.exp(-dd); }
        const grain = 0.5 + 0.5 * Math.sin(ny / 3.4 + noise(x / 160, ny / 40) * 1.6);
        const fine = noise(x / 5, y / 1.4) * 0.5;
        let k = 0.78 + 0.16 * grain + 0.05 * fine;
        for (const [kx, ky, kr] of knots) { const dx = (x - kx) / (kr * 1.1), dy = (y - ky) / (kr * 0.7), dd = Math.sqrt(dx * dx + dy * dy); if (dd < 1) k -= 0.35 * (1 - dd); else if (dd < 1.2) k -= 0.15; }
        if (py < 3) k *= 0.45; else if (py < 5) k *= 0.8;
        k *= tone;
        const i = (y * W + x) * 4; d[i] = Math.min(255, 214 * k); d[i + 1] = Math.min(255, 168 * k); d[i + 2] = Math.min(255, 112 * k); d[i + 3] = 255;
      }
    }
    c.putImageData(img, 0, 0);
    return cv;
  }

  /* ---------- a cribbage board, for the icon and store art ---------- */
  function board(c, x, y, w, h, pegs) {
    c.save(); c.translate(x, y);
    c.fillStyle = 'rgba(0,0,0,.4)'; rrect(c, 6, 10, w, h, h * 0.22); c.fill();
    const g = c.createLinearGradient(0, 0, 0, h); g.addColorStop(0, '#c08a50'); g.addColorStop(1, '#7a4a20'); c.fillStyle = g; rrect(c, 0, 0, w, h, h * 0.22); c.fill();
    c.strokeStyle = 'rgba(60,30,10,.35)'; c.lineWidth = Math.max(1, h * 0.012); for (let i = 1; i < 8; i++) { c.beginPath(); c.moveTo(0, h * i / 8); c.bezierCurveTo(w * 0.3, h * i / 8 + h * 0.03, w * 0.6, h * i / 8 - h * 0.03, w, h * i / 8); c.stroke(); }
    c.strokeStyle = '#3a2410'; c.lineWidth = Math.max(1.5, h * 0.02); rrect(c, 0, 0, w, h, h * 0.22); c.stroke();
    const rows = 3, cols = 30, r = Math.min(w / cols * 0.2, h / 10 * 0.5), holes = [];
    for (let row = 0; row < rows; row++) for (let col = 0; col < cols; col++) {
      const hx = w * 0.07 + (w * 0.86) * (col + Math.floor(col / 5) * 0.5) / (cols + 2.5), hy = h * (0.25 + row * 0.25);
      holes.push([hx, hy]); c.fillStyle = '#2a1608'; c.beginPath(); c.arc(hx, hy, r, 0, TAU); c.fill();
    }
    for (const [row, col, color] of pegs || []) {
      const [hx, hy] = holes[row * cols + col]; c.fillStyle = 'rgba(0,0,0,.4)'; c.beginPath(); c.arc(hx + r * 0.6, hy + r * 0.8, r * 1.5, 0, TAU); c.fill();
      c.fillStyle = color; c.beginPath(); c.arc(hx, hy, r * 1.5, 0, TAU); c.fill(); c.strokeStyle = 'rgba(0,0,0,.5)'; c.lineWidth = r * 0.3; c.stroke();
      c.fillStyle = 'rgba(255,255,255,.45)'; c.beginPath(); c.arc(hx - r * 0.5, hy - r * 0.5, r * 0.5, 0, TAU); c.fill();
    }
    c.restore();
  }

  /* ---------- the icon: a board with pegs and a five ---------- */
  function icon(size) {
    const cv = canvas(size, size), c = cv.getContext('2d'), s = size / 512;
    c.scale(s, s);
    rrect(c, 16, 16, 480, 480, 96); const g = c.createLinearGradient(0, 16, 0, 496); g.addColorStop(0, '#2a4a74'); g.addColorStop(1, NAVY2); c.fillStyle = g; c.fill();
    c.save(); rrect(c, 16, 16, 480, 480, 96); c.clip();
    for (let y = 16; y < 512; y += 64) for (let x = 16 + ((y - 16) / 64 % 2) * 32; x < 512; x += 64) star(c, x, y + 32, 12, 'rgba(232,207,138,.14)');
    c.restore();
    c.strokeStyle = '#e8c35a'; c.lineWidth = 12; rrect(c, 40, 40, 432, 432, 76); c.stroke();
    c.save(); c.translate(256, 300); c.rotate(-0.32); board(c, -200, -60, 400, 120, [[0, 22, '#d8402e'], [0, 17, '#d8402e'], [1, 26, '#3f86d0'], [1, 12, '#3f86d0']]); c.restore();
    c.save(); c.translate(300, 196); c.rotate(0.18);
    rrect(c, -78, -110, 156, 220, 18); c.fillStyle = '#fffaf0'; c.fill(); c.strokeStyle = '#1c1a1e'; c.lineWidth = 6; c.stroke();
    c.fillStyle = RED; c.font = `700 64px ${DISP}`; c.textAlign = 'center'; c.fillText('5', -40, -46);
    suit(c, 'H', 14, 20, 104, '#c0282a'); c.restore();
    return cv;
  }

  return { face, back: backCard, mug, lantern, wood, icon, board, lighthouse, star, rope, heartPath, suit, CW, CH, PAL, DISP, SERIF, COURT };
})();
