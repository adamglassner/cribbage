'use strict';
/* =====================================================================
   CRIBBAGE - the cottage music and sound effects, synthesized with WebAudio.
   Original compositions (c) Ironite:
     "The Keeper's Reel"     - a fiddle reel over guitar
     "Fog on the Bay"        - a slow concertina air
     "The Lobster Pot Jig"   - a jig for fiddle and concertina
   bake.html renders each with an OfflineAudioContext into the game's files.
   ===================================================================== */
const Sound = (() => {
  const mtof = m => 440 * Math.pow(2, (m - 69) / 12);
  let noiseBuf = null;
  function noise(ctx) {
    if (noiseBuf && noiseBuf.sampleRate === ctx.sampleRate) return noiseBuf;
    const n = ctx.sampleRate * 2, b = ctx.createBuffer(1, n, ctx.sampleRate), d = b.getChannelData(0);
    let s = 12345; for (let i = 0; i < n; i++) { s = (s * 1103515245 + 12345) & 0x7fffffff; d[i] = s / 0x3fffffff - 1; }
    return (noiseBuf = b);
  }
  function env(g, t, a, d, sus, r, peak, len) {   // ADSR on a gain param
    g.setValueAtTime(0, t); g.linearRampToValueAtTime(peak, t + a);
    g.linearRampToValueAtTime(peak * sus, t + a + d);
    g.setValueAtTime(peak * sus, Math.max(t + a + d, t + len));
    g.linearRampToValueAtTime(0, Math.max(t + a + d, t + len) + r);
  }
  function reverb(ctx, secs = 1.8, wet = 0.22) {
    const len = Math.floor(ctx.sampleRate * secs), ir = ctx.createBuffer(2, len, ctx.sampleRate);
    for (let ch = 0; ch < 2; ch++) { const d = ir.getChannelData(ch); let s = 99 + ch; for (let i = 0; i < len; i++) { s = (s * 1103515245 + 12345) & 0x7fffffff; d[i] = (s / 0x3fffffff - 1) * Math.pow(1 - i / len, 3); } }
    const conv = ctx.createConvolver(); conv.buffer = ir;
    const inp = ctx.createGain(), dry = ctx.createGain(), w = ctx.createGain();
    dry.gain.value = 1; w.gain.value = wet;
    inp.connect(dry); inp.connect(conv); conv.connect(w);
    const out = ctx.createGain(); dry.connect(out); w.connect(out);
    return { inp, out };
  }

  /* ---------------- instruments ---------------- */
  // the concertina: three reeds, two of them tuned a hair apart
  function accordion(ctx, dest, t, m, dur, vel, bright = 1, spread = 9) {
    const out = ctx.createGain(); env(out.gain, t, 0.025, 0.08, 0.85, 0.07, vel * 0.2, dur);
    const lp = ctx.createBiquadFilter(); lp.type = 'lowpass'; lp.frequency.value = 2400 * bright; lp.Q.value = 0.7;
    const trem = ctx.createGain(); trem.gain.value = 1;
    const lfo = ctx.createOscillator(); lfo.frequency.value = 5.2; const lg = ctx.createGain(); lg.gain.value = 0.08; lfo.connect(lg); lg.connect(trem.gain);
    for (const [det, typ, g] of [[-spread, 'sawtooth', 0.5], [spread, 'sawtooth', 0.5], [0, 'square', 0.25]]) {
      const o = ctx.createOscillator(); o.type = typ; o.frequency.value = mtof(m); o.detune.value = det;
      const og = ctx.createGain(); og.gain.value = g; o.connect(og); og.connect(lp);
      o.start(t); o.stop(t + dur + 0.2);
    }
    lp.connect(trem); trem.connect(out); out.connect(dest);
    lfo.start(t); lfo.stop(t + dur + 0.2);
  }
  // a plucked string (guitar, double bass): a bright saw that darkens as it rings
  function pluck(ctx, dest, t, m, dur, vel, bright = 3200, decay = 0.6) {
    const o = ctx.createOscillator(); o.type = 'sawtooth'; o.frequency.value = mtof(m);
    const o2 = ctx.createOscillator(); o2.type = 'triangle'; o2.frequency.value = mtof(m);
    const lp = ctx.createBiquadFilter(); lp.type = 'lowpass'; lp.Q.value = 1;
    lp.frequency.setValueAtTime(bright, t); lp.frequency.exponentialRampToValueAtTime(Math.max(200, bright * 0.18), t + decay);
    const g = ctx.createGain(); g.gain.setValueAtTime(0, t); g.gain.linearRampToValueAtTime(vel * 0.22, t + 0.004);
    g.gain.exponentialRampToValueAtTime(0.0005, t + Math.max(dur, 0.05) + decay);
    const g2 = ctx.createGain(); g2.gain.value = 0.6; o2.connect(g2); g2.connect(lp);
    o.connect(lp); lp.connect(g); g.connect(dest);
    o.start(t); o.stop(t + dur + decay + 0.05); o2.start(t); o2.stop(t + dur + decay + 0.05);
  }
  function bass(ctx, dest, t, m, dur, vel) {
    const o = ctx.createOscillator(); o.type = 'triangle'; o.frequency.value = mtof(m);
    const s = ctx.createOscillator(); s.type = 'sine'; s.frequency.value = mtof(m);
    const g = ctx.createGain(); g.gain.setValueAtTime(0, t); g.gain.linearRampToValueAtTime(vel * 0.42, t + 0.01);
    g.gain.exponentialRampToValueAtTime(vel * 0.12, t + dur * 0.6); g.gain.linearRampToValueAtTime(0, t + dur + 0.05);
    const lp = ctx.createBiquadFilter(); lp.type = 'lowpass'; lp.frequency.value = 900;
    o.connect(lp); s.connect(lp); lp.connect(g); g.connect(dest);
    o.start(t); o.stop(t + dur + 0.1); s.start(t); s.stop(t + dur + 0.1);
  }
  // a violin: a bowed saw with vibrato that blooms in
  function violin(ctx, dest, t, m, dur, vel) {
    const o = ctx.createOscillator(); o.type = 'sawtooth'; o.frequency.value = mtof(m);
    const vib = ctx.createOscillator(); vib.frequency.value = 5.6; const vg = ctx.createGain();
    vg.gain.setValueAtTime(0, t); vg.gain.linearRampToValueAtTime(16, t + Math.min(0.35, dur));
    vib.connect(vg); vg.connect(o.detune);
    const bp = ctx.createBiquadFilter(); bp.type = 'lowpass'; bp.frequency.value = 3600; bp.Q.value = 0.8;
    const pk = ctx.createBiquadFilter(); pk.type = 'peaking'; pk.frequency.value = 2600; pk.gain.value = 6; pk.Q.value = 1.4;
    const g = ctx.createGain(); env(g.gain, t, 0.05, 0.1, 0.8, 0.12, vel * 0.16, dur);
    o.connect(bp); bp.connect(pk); pk.connect(g); g.connect(dest);
    o.start(t); o.stop(t + dur + 0.3); vib.start(t); vib.stop(t + dur + 0.3);
  }
  // a piano: a few decaying partials
  function piano(ctx, dest, t, m, dur, vel) {
    const f = mtof(m), ring = Math.min(3.2, 0.6 + dur * 1.2) * (m < 60 ? 1.3 : 1);
    const out = ctx.createGain(); out.gain.value = vel * 0.2; out.connect(dest);
    for (const [k, a, det] of [[1, 0.62, 0], [1, 0.3, 4], [2, 0.22, 0], [3, 0.1, 0], [4, 0.05, 0], [5.02, 0.03, 0]]) {
      const o = ctx.createOscillator(); o.type = 'sine'; o.frequency.value = f * k; o.detune.value = det;
      const g = ctx.createGain(); g.gain.setValueAtTime(0, t); g.gain.linearRampToValueAtTime(a, t + 0.006);
      g.gain.exponentialRampToValueAtTime(0.0008, t + ring / Math.sqrt(k));
      o.connect(g); g.connect(out); o.start(t); o.stop(t + ring + 0.05);
    }
  }
  function pad(ctx, dest, t, m, dur, vel) {
    for (const det of [-7, 7]) {
      const o = ctx.createOscillator(); o.type = 'sawtooth'; o.frequency.value = mtof(m); o.detune.value = det;
      const lp = ctx.createBiquadFilter(); lp.type = 'lowpass'; lp.frequency.value = 900;
      const g = ctx.createGain(); env(g.gain, t, 0.5, 0.3, 0.8, 0.8, vel * 0.04, dur);
      o.connect(lp); lp.connect(g); g.connect(dest); o.start(t); o.stop(t + dur + 1);
    }
  }
  function brush(ctx, dest, t, vel) {
    const s = ctx.createBufferSource(); s.buffer = noise(ctx);
    const hp = ctx.createBiquadFilter(); hp.type = 'bandpass'; hp.frequency.value = 5000; hp.Q.value = 0.6;
    const g = ctx.createGain(); g.gain.setValueAtTime(vel * 0.12, t); g.gain.exponentialRampToValueAtTime(0.001, t + 0.14);
    s.connect(hp); hp.connect(g); g.connect(dest); s.start(t, Math.random() * 1.5); s.stop(t + 0.16);
  }

  /* ---------------- the tunes ---------------- */
  const CH = {   // chord tones (mid register) and the bass root
    D: [[62, 66, 69], 50], G: [[62, 67, 71], 43], A: [[61, 64, 69], 45], A7: [[61, 64, 67], 45], Em: [[59, 64, 67], 52],
    Bm: [[59, 62, 66], 47], C: [[60, 64, 67], 48], Am: [[60, 64, 69], 45], B7: [[59, 63, 66], 47], D7: [[60, 62, 66], 50]
  };
  const n = (s) => s.replace(/\|/g, ' ').split(' ').filter(Boolean).map(x => { const [m, d] = x.split(':'); return [+m, d === undefined ? 0.5 : +d]; });

  // 1. The Keeper's Reel: 4/4, D major, AABB, fiddle over boom-chuck guitar
  const REEL = {
    name: "The Keeper's Reel", bpm: 196, beats: 4, kind: 'reel',
    A: { ch: 'D D G A D G A D'.split(' '),
      mel: n('62 66 69 66 74 69 66 69 | 62 66 69 74 78 76 74 73 | 71 67 71 74 79 74 71 74 | 73 69 73 76 81 76 73 69 | 74 73 74 76 78 76 74 69 | 71 74 71 67 64 67 71 74 | 73 71 69 67 66 64 62 64 | 66:1 69:1 62:2') },
    B: { ch: 'G D A D G D A D'.split(' '),
      mel: n('79 78 79 81 83 81 79 78 | 78 74 78 81 86 81 78 74 | 76 73 76 81 85 81 76 73 | 74 78 81 78 74:2 | 79 83 79 74 71 74 79 83 | 81 78 74 78 81 86 81 78 | 76 78 76 73 69 71 73 76 | 74:1 69:1 74:2') },
    form: ['A', 'A', 'B', 'B', 'A'], end: 'D'
  };
  // 2. Fog on the Bay: a slow 3/4 air in E minor, concertina then fiddle
  const AIR = {
    name: 'Fog on the Bay', bpm: 92, beats: 3, kind: 'air',
    A: { ch: 'Em Em C D G G Am B7 Em Em C D G D Em Em'.split(' '),
      mel: n('64:2 67:1 | 71:2 69:.5 67:.5 | 64:1.5 67:.5 72:1 | 69:3 | 71:2 74:1 | 79:2 78:.5 76:.5 | 76:1.5 74:.5 72:1 | 71:3 | 64:2 67:1 | 71:2 74:1 | 76:1.5 74:.5 72:1 | 74:1 72:1 69:1 | 71:2 67:1 | 69:1.5 66:.5 62:1 | 64:3 | 64:2 0:1') },
    form: ['A', 'A'], end: 'Em'
  };
  // 3. The Lobster Pot Jig: 6/8 in G (counted in eighths), AABB
  const JIG = {
    name: 'The Lobster Pot Jig', bpm: 312, beats: 6, kind: 'jig',
    A: { ch: 'G G C D G Em D G'.split(' '),
      mel: n('67:1 71:1 74:1 79:1 74:1 71:1 | 72:1 71:1 69:1 71:2 67:1 | 72:1 76:1 79:1 76:1 72:1 76:1 | 74:1 78:1 81:1 78:2 74:1 | 71:1 74:1 79:1 74:1 71:1 67:1 | 64:1 67:1 71:1 76:2 71:1 | 78:1 76:1 74:1 72:1 71:1 69:1 | 67:3 67:2 0:1') },
    B: { ch: 'C G C D G C D G'.split(' '),
      mel: n('76:1 79:1 76:1 72:1 76:1 79:1 | 79:1 78:1 79:1 74:2 71:1 | 76:1 79:1 81:1 79:1 76:1 72:1 | 74:2 78:1 81:2 78:1 | 79:1 81:1 83:1 81:1 79:1 74:1 | 76:1 79:1 76:1 72:1 69:1 72:1 | 74:1 78:1 76:1 74:1 72:1 69:1 | 71:1 69:1 67:4') },
    form: ['A', 'A', 'B', 'B'], end: 'G'
  };
  const TUNES = [REEL, AIR, JIG];

  // buildQueue turns a tune into timed note events.
  function buildQueue(tune) {
    const beat = 60 / tune.bpm, ev = [];
    let t = 0.1;
    for (let pass = 0; pass < tune.form.length; pass++) {
      const sec = tune[tune.form[pass]];
      // melody: the fiddle leads the reel and jig; the concertina sings the air first time round
      const lead = tune.kind === 'air' ? (pass === 0 ? 'con' : 'violin') : 'violin';
      let mt = 0;
      for (const [m, d] of sec.mel) {
        if (m > 0) {
          const st = mt, dd = Math.max(0.08, d * beat * (tune.kind === 'air' ? 0.96 : 0.86));
          ev.push({ i: lead, t: t + st * beat, m, d: dd, v: tune.kind === 'air' ? 0.8 : 0.9 });
          if (tune.kind === 'jig' && tune.form[pass] === 'B') ev.push({ i: 'con', t: t + st * beat, m: m - 12, d: dd, v: 0.4 });
          if (tune.kind === 'air' && pass === 1) ev.push({ i: 'con', t: t + st * beat, m: m - 12, d: dd, v: 0.3 });
        }
        mt += d;
      }
      // accompaniment
      sec.ch.forEach((name, k) => {
        const [tones, root] = CH[name], b0 = t + k * tune.beats * beat;
        if (tune.kind === 'reel') {   // boom-chuck: bass on 1 and 3, a strum on 2 and 4
          for (let bb = 0; bb < 4; bb++) {
            if (bb % 2 === 0) ev.push({ i: 'bass', t: b0 + bb * beat, m: bb === 0 ? root : root + 7, d: beat * 0.8, v: 0.9 });
            else tones.forEach((tn, j) => ev.push({ i: 'gtr', t: b0 + bb * beat + j * 0.012, m: tn, d: beat * 0.3, v: 0.5, br: 3600 }));
          }
          if (pass >= 2) ev.push({ i: 'con', t: b0, m: tones[0], d: tune.beats * beat * 0.95, v: 0.18, b: 0.6 });
        } else if (tune.kind === 'air') {   // the guitar picks a slow arpeggio
          const arp = [root, tones[0], tones[1], tones[2], tones[1], tones[0]];
          arp.forEach((m, j) => ev.push({ i: 'gtr', t: b0 + j * beat * 0.5, m, d: beat * 0.6, v: j === 0 ? 0.55 : 0.36, br: 2400, dec: 1.1 }));
        } else {   // jig: bass on 1 and 4, guitar on 3 and 6
          ev.push({ i: 'bass', t: b0, m: root, d: beat * 1.8, v: 0.9 });
          ev.push({ i: 'bass', t: b0 + 3 * beat, m: root + 7, d: beat * 1.8, v: 0.8 });
          for (const bb of [2, 5]) tones.forEach((tn, j) => ev.push({ i: 'gtr', t: b0 + bb * beat + j * 0.01, m: tn, d: beat * 0.6, v: 0.45, br: 3400 }));
        }
      });
      t += sec.ch.length * tune.beats * beat;
    }
    // a final chord
    const [tones, root] = CH[tune.end], hold = tune.kind === 'jig' ? beat * 6 : beat * 3;
    ev.push({ i: 'bass', t, m: root, d: hold, v: 1 });
    for (const tn of tones) ev.push({ i: 'gtr', t, m: tn, d: hold, v: 0.5, br: 3000, dec: 1.4 });
    for (const tn of tones) ev.push({ i: 'con', t, m: tn + 12, d: hold, v: 0.35 });
    ev.push({ i: tune.kind === 'air' ? 'con' : 'violin', t, m: tones[0] + 12, d: hold, v: 0.7 });
    return { ev, len: t + hold + 1.4 };
  }

  function renderTune(ctx, dest, tune) {
    const air = tune.kind === 'air';
    const rv = reverb(ctx, air ? 2.6 : 1.5, air ? 0.3 : 0.18); rv.out.connect(dest);
    const bus = ctx.createGain(); bus.gain.value = 0.8; bus.connect(rv.inp);
    const q = buildQueue(tune);
    for (const e of q.ev) {
      if (e.i === 'con') accordion(ctx, bus, e.t, e.m, e.d, e.v, (e.b || 1) * 0.85, 3);
      else if (e.i === 'bass') bass(ctx, bus, e.t, e.m, e.d, e.v);
      else if (e.i === 'gtr') pluck(ctx, bus, e.t, e.m, e.d, e.v, e.br || 3000, e.dec || 0.3);
      else if (e.i === 'violin') violin(ctx, bus, e.t, e.m, e.d, e.v);
    }
    return q.len;
  }

  /* ---------------- sound effects ---------------- */
  function flick(ctx, dest, t, vel = 1, f = 2600, len = 0.06) {   // one card snapping through the air
    const s = ctx.createBufferSource(); s.buffer = noise(ctx);
    const bp = ctx.createBiquadFilter(); bp.type = 'bandpass'; bp.frequency.setValueAtTime(f, t); bp.frequency.exponentialRampToValueAtTime(f * 0.5, t + len); bp.Q.value = 1.2;
    const g = ctx.createGain(); g.gain.setValueAtTime(0, t); g.gain.linearRampToValueAtTime(vel * 0.6, t + 0.004); g.gain.exponentialRampToValueAtTime(0.001, t + len);
    s.connect(bp); bp.connect(g); g.connect(dest); s.start(t, (t * 7.3) % 1.5); s.stop(t + len + 0.02);
  }
  function thump(ctx, dest, t, vel = 1) {   // a card landing on the pine table
    flick(ctx, dest, t, vel * 0.8, 1600, 0.05);
    const o = ctx.createOscillator(); o.frequency.setValueAtTime(160, t); o.frequency.exponentialRampToValueAtTime(70, t + 0.07);
    const g = ctx.createGain(); g.gain.setValueAtTime(vel * 0.28, t); g.gain.exponentialRampToValueAtTime(0.001, t + 0.09);
    o.connect(g); g.connect(dest); o.start(t); o.stop(t + 0.1);
  }
  function wood(ctx, dest, t, f = 900, vel = 1, len = 0.09) {   // a hollow wooden tok
    for (const [k, a] of [[1, 1], [2.3, 0.4], [3.9, 0.15]]) {
      const o = ctx.createOscillator(); o.type = 'sine'; o.frequency.setValueAtTime(f * k * 1.04, t); o.frequency.exponentialRampToValueAtTime(f * k, t + 0.02);
      const g = ctx.createGain(); g.gain.setValueAtTime(0, t); g.gain.linearRampToValueAtTime(vel * 0.32 * a, t + 0.002); g.gain.exponentialRampToValueAtTime(0.0005, t + len / k);
      o.connect(g); g.connect(dest); o.start(t); o.stop(t + len + 0.02);
    }
    flick(ctx, dest, t, vel * 0.3, 3800, 0.012);
  }
  function bell(ctx, dest, t, m, vel = 1, len = 1.2) {   // a brass ship's bell
    for (const [k, a] of [[1, 1], [2.0, 0.5], [2.76, 0.3], [5.4, 0.12]]) {
      const o = ctx.createOscillator(); o.type = 'sine'; o.frequency.value = mtof(m) * k;
      const g = ctx.createGain(); g.gain.setValueAtTime(0, t); g.gain.linearRampToValueAtTime(vel * 0.16 * a, t + 0.003); g.gain.exponentialRampToValueAtTime(0.0005, t + len / Math.sqrt(k));
      o.connect(g); g.connect(dest); o.start(t); o.stop(t + len);
    }
  }
  const con = (ctx, d, t, m, dur, v, b = 1) => accordion(ctx, d, t, m, dur, v, b * 0.9, 3);
  const fx = {
    shuffle(ctx, d) { let t = 0.02; for (let i = 0; i < 26; i++) { flick(ctx, d, t, 0.45 + 0.2 * Math.sin(i), 3000 + (i % 5) * 200, 0.035); t += 0.018 + (i % 3) * 0.004; } thump(ctx, d, t + 0.05, 0.8); return t + 0.3; },
    deal(ctx, d) { flick(ctx, d, 0.01, 0.8, 2800, 0.07); return 0.15; },
    play(ctx, d) { flick(ctx, d, 0.01, 0.5, 2300, 0.05); thump(ctx, d, 0.05, 0.9); return 0.2; },
    flip(ctx, d) { flick(ctx, d, 0.01, 0.7, 3400, 0.05); flick(ctx, d, 0.07, 0.6, 2200, 0.06); thump(ctx, d, 0.12, 0.7); return 0.3; },
    hover(ctx, d) { const o = ctx.createOscillator(); o.frequency.value = 2000; const g = ctx.createGain(); g.gain.setValueAtTime(0.03, 0.005); g.gain.exponentialRampToValueAtTime(0.0005, 0.04); o.connect(g); g.connect(d); o.start(0.005); o.stop(0.05); return 0.06; },
    tick(ctx, d) { wood(ctx, d, 0.005, 1700, 0.35, 0.05); return 0.08; },
    pass(ctx, d) { for (let i = 0; i < 2; i++) flick(ctx, d, 0.02 + i * 0.11, 0.8, 2600 - i * 200, 0.09); thump(ctx, d, 0.28, 0.6); return 0.45; },
    slide(ctx, d) { const s = ctx.createBufferSource(); s.buffer = noise(ctx); const bp = ctx.createBiquadFilter(); bp.type = 'bandpass'; bp.frequency.setValueAtTime(1100, 0.01); bp.frequency.linearRampToValueAtTime(2400, 0.2); bp.Q.value = 0.8; const g = ctx.createGain(); g.gain.setValueAtTime(0, 0.01); g.gain.linearRampToValueAtTime(0.25, 0.06); g.gain.exponentialRampToValueAtTime(0.001, 0.24); s.connect(bp); bp.connect(g); g.connect(d); s.start(0.01); s.stop(0.26); return 0.3; },
    sweep(ctx, d) { const s = ctx.createBufferSource(); s.buffer = noise(ctx); const bp = ctx.createBiquadFilter(); bp.type = 'bandpass'; bp.frequency.setValueAtTime(700, 0.01); bp.frequency.exponentialRampToValueAtTime(2600, 0.28); bp.Q.value = 0.7; const g = ctx.createGain(); g.gain.setValueAtTime(0, 0.01); g.gain.linearRampToValueAtTime(0.3, 0.1); g.gain.exponentialRampToValueAtTime(0.001, 0.34); s.connect(bp); bp.connect(g); g.connect(d); s.start(0.01); s.stop(0.36); return 0.4; },
    knock(ctx, d) { wood(ctx, d, 0.01, 240, 1, 0.14); wood(ctx, d, 0.16, 230, 0.9, 0.14); return 0.4; },   // "Go!" - two raps on the table
    peg(ctx, d) { wood(ctx, d, 0.01, 1150, 0.9, 0.08); wood(ctx, d, 0.035, 620, 0.5, 0.12); return 0.2; },
    count(ctx, d) { wood(ctx, d, 0.005, 1400, 0.6, 0.07); bell(ctx, d, 0.01, 91, 0.25, 0.4); return 0.4; },
    fifteen(ctx, d) { bell(ctx, d, 0.01, 81, 0.7, 1.2); bell(ctx, d, 0.16, 86, 0.6, 1.4); return 1.5; },
    run(ctx, d) { [74, 76, 78, 79, 81].forEach((m, i) => pluck(ctx, d, 0.01 + i * 0.065, m, 0.18, 0.6, 4600, 0.5)); bell(ctx, d, 0.35, 86, 0.5, 1.2); return 1.5; },
    royal(ctx, d) { for (let i = 0; i < 3; i++) bell(ctx, d, 0.01 + i * 0.22, 79, 0.8, 1.6); for (const m of [67, 71, 74, 79]) con(ctx, d, 0.66, m, 0.9, 0.45, 1.2); return 2.2; },   // three strokes of the ship's bell
    good(ctx, d) { con(ctx, d, 0.01, 74, 0.16, 0.6, 1.1); con(ctx, d, 0.17, 79, 0.42, 0.6, 1.1); bell(ctx, d, 0.17, 91, 0.4, 1.1); return 1.2; },
    big(ctx, d) {   // a concertina flourish up to a full chord and the bell
      [67, 71, 74, 79, 83].forEach((m, i) => con(ctx, d, 0.01 + i * 0.08, m, 0.14, 0.6, 1.2));
      for (const m of [55, 67, 71, 74, 79, 86]) con(ctx, d, 0.42, m, 1.0, 0.45, 1.2);
      bell(ctx, d, 0.42, 91, 0.7, 1.8); bell(ctx, d, 0.62, 98, 0.5, 1.6); return 2.4;
    },
    nineteen(ctx, d) { [[64, 0.22], [63, 0.22], [62, 0.22], [61, 0.6]].reduce((t, [m, dur]) => { con(ctx, d, t, m, dur * 0.9, 0.7, 0.6); con(ctx, d, t, m - 12, dur * 0.9, 0.4, 0.5); return t + dur; }, 0.02); return 1.6; },   // a sad "wah-wah"
    win(ctx, d) {   // a reel's last bars on fiddle and concertina
      const seq = [[74, .12], [78, .12], [81, .12], [86, .45], [85, .12], [86, .7]];
      let t = 0.02; for (const [m, dur] of seq) { violin(ctx, d, t, m, dur * 0.95, 0.9); con(ctx, d, t, m - 12, dur * 0.95, 0.4, 1.1); t += dur; }
      for (const m of [50, 62, 66, 69, 74]) con(ctx, d, t - 0.7, m, 0.8, 0.45);
      bell(ctx, d, t - 0.7, 93, 0.6, 1.8);
      return t + 1.3;
    },
    lose(ctx, d) {
      const seq = [[71, .3], [69, .3], [67, .3], [64, .9]];
      let t = 0.02; for (const [m, dur] of seq) { con(ctx, d, t, m, dur * 0.95, 0.6, 0.7); t += dur; }
      for (const m of [52, 59, 64, 67]) con(ctx, d, t - 0.9, m, 0.9, 0.35, 0.6);
      return t + 0.8;
    }
  };

  return { fx, TUNES, renderTune, buildQueue, reverb };
})();
