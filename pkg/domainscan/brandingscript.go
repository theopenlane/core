package domainscan

// brandStyleProbeScript returns computed styles directly through CDP evaluation.
const brandStyleProbeScript = `(() => {
  const hex = (r, g, b) => '#' + [r, g, b].map(n => n.toString(16).padStart(2, '0')).join('');
  // a 1x1 canvas resolves any css color syntax, including oklch and color(), to srgb bytes plus
  // alpha, which is what tailwind v4 and other modern frameworks expose as computed values
  const canvas = document.createElement('canvas');
  canvas.width = canvas.height = 1;
  const ctx = canvas.getContext('2d', { willReadFrequently: true });
  const toHex = c => {
    if (!c || c === 'transparent' || c === 'none') return '';
    const m = /^rgba?\((\d+),\s*(\d+),\s*(\d+)(?:,\s*([\d.]+))?\)$/.exec(c);
    if (m) {
      if (m[4] !== undefined && parseFloat(m[4]) === 0) return '';
      return hex(+m[1], +m[2], +m[3]);
    }
    if (!ctx) return '';
    try {
      ctx.clearRect(0, 0, 1, 1);
      ctx.fillStyle = '#010203';
      ctx.fillStyle = c;
      if (ctx.fillStyle === '#010203') return '';
      ctx.fillRect(0, 0, 1, 1);
      const d = ctx.getImageData(0, 0, 1, 1).data;
      return d[3] === 0 ? '' : hex(d[0], d[1], d[2]);
    } catch (e) { return ''; }
  };
  const rgb = h => [parseInt(h.slice(1, 3), 16), parseInt(h.slice(3, 5), 16), parseInt(h.slice(5, 7), 16)];
  // chroma is the absolute channel spread, so dark navies and pale tints score low even when
  // their relative saturation is high
  const chroma = h => { const c = rgb(h); return Math.max(...c) - Math.min(...c); };
  const lum = h => { const [r, g, b] = rgb(h); return (0.2126 * r + 0.7152 * g + 0.0722 * b) / 255; };
  const isNoise = el => !!el.closest('[id*="cookie" i],[class*="cookie" i],[id*="consent" i],[class*="consent" i],[id*="iubenda" i],[class*="iubenda" i],[id*="banner" i],[class*="banner" i]');
  const cs = el => getComputedStyle(el);
  const MIN_CHROMA = 48, MIN_LUM = 0.12, MAX_LUM = 0.92;
  const isAccentLike = h => chroma(h) >= MIN_CHROMA && lum(h) >= MIN_LUM && lum(h) <= MAX_LUM;

  // effective background walks html, body, and large early descendants until a painted color is found
  let background = '';
  const bgCandidates = [document.documentElement, document.body, ...Array.from(document.body.children).slice(0, 6), document.querySelector('main')].filter(Boolean);
  for (const el of bgCandidates) {
    const h = toHex(cs(el).backgroundColor);
    if (h) { background = h; break; }
  }

  const bodyStyle = cs(document.body);
  const mainText = document.querySelector('main p') || document.querySelector('p') || document.querySelector('main') || document.body;
  let foreground = toHex(cs(mainText || document.body).color) || toHex(bodyStyle.color);

  // gather accent candidates from interactive elements, weighted so filled button backgrounds
  // outrank text and border colors, then ranked by chroma and weighted frequency
  const tally = new Map();
  const add = (h, w) => { if (h && isAccentLike(h)) tally.set(h, (tally.get(h) || 0) + w); };
  const interactive = Array.from(document.querySelectorAll('button, a, [role="button"], [class*="btn" i], [class*="button" i], [class*="cta" i]')).filter(el => !isNoise(el));
  for (const el of interactive.slice(0, 200)) {
    const st = cs(el);
    add(toHex(st.backgroundColor), 3);
    add(toHex(st.color), 1);
    add(toHex(st.borderColor), 1);
  }
  for (const el of Array.from(document.querySelectorAll('h1, h2, span, strong, svg')).slice(0, 200)) {
    if (isNoise(el)) continue;
    const st = cs(el);
    add(toHex(st.color), 1);
    add(toHex(st.fill), 1);
  }

  // css variables are bucketed by name; only canonical brand token names may feed the accent
  // pool, so utility tokens like --border-accent or --color-icon-primary do not masquerade as brand
  const root = cs(document.documentElement);
  const vars = { accent: {}, foreground: {}, background: {} };
  const brandToken = /^--(color-)?(brand|primary|accent)(-(button|cta|500|600|default))?$/i;
  const brandPriority = ['--color-primary-button', '--color-primary', '--primary', '--color-brand', '--brand', '--color-brand-500', '--color-brand-600', '--color-accent', '--accent'];
  const visit = rules => {
    for (const rule of rules) {
      // descend into @layer, @media, and @supports blocks where frameworks put their tokens
      if (rule.cssRules && rule.cssRules.length) visit(Array.from(rule.cssRules));
      if (!rule.style || !rule.selectorText || !/(^|,)\s*(:root|html|body|:host)\s*($|,)/.test(rule.selectorText)) continue;
      for (const name of Array.from(rule.style)) {
        if (!name.startsWith('--')) continue;
        const raw = root.getPropertyValue(name).trim();
        const h = /^#[0-9a-f]{6}$/i.test(raw) ? raw.toLowerCase() : toHex(raw);
        if (!h) continue;
        if (/text|foreground|fg/i.test(name)) vars.foreground[name] = h;
        else if (/background|bg/i.test(name)) vars.background[name] = h;
        else if (brandToken.test(name)) { vars.accent[name] = h; add(h, 6); }
      }
    }
  };
  for (const sheet of Array.from(document.styleSheets)) {
    try { visit(Array.from(sheet.cssRules || [])); } catch (e) { continue; }
  }

  // primary comes straight from the first canonical brand token that reads as an accent color
  let primary = '';
  for (const name of brandPriority) {
    const h = vars.accent[name];
    if (h && isAccentLike(h)) { primary = h; break; }
  }

  const accent_candidates = Array.from(tally.entries())
    .sort((a, b) => (chroma(b[0]) * Math.log(1 + b[1])) - (chroma(a[0]) * Math.log(1 + a[1])))
    .slice(0, 5)
    .map(([h]) => h);

  // secondary surfaces come from the header or nav and the first card-like element
  const nav = document.querySelector('header, nav');
  const card = Array.from(document.querySelectorAll('[class*="card" i], section, article')).find(el => !isNoise(el) && toHex(cs(el).backgroundColor) && toHex(cs(el).backgroundColor) !== background);
  const navBackground = nav ? toHex(cs(nav).backgroundColor) : '';
  const secondarySurface = navBackground && navBackground !== background ? nav : card;
  const secondary_background = secondarySurface ? toHex(cs(secondarySurface).backgroundColor) : '';
  const secondary_foreground = secondarySurface ? toHex(cs(secondarySurface).color) : '';

  // strip next.js style generated family names like __Inter_abc123 and quotes
  const font = (bodyStyle.fontFamily || '').split(',')[0].trim().replace(/^["']|["']$/g, '').replace(/^__([A-Za-z0-9]+?)_[0-9a-f]+$/, '$1').replace(/_/g, ' ');

  const logoImage = Array.from(document.querySelectorAll('header img, nav img, [role="banner"] img')).find(el => {
    if (!el.getClientRects().length || el.closest('[aria-hidden="true"]')) return false;
    const st = cs(el);
    if (st.visibility !== 'visible' || st.display === 'none' || Number(st.opacity) === 0) return false;
    const link = el.closest('a');
    const homeLink = link && new URL(link.href, location.href).origin === location.origin &&
      new URL(link.href, location.href).pathname.replace(/\/$/, '') === location.pathname.replace(/\/$/, '');
    return /logo/i.test([el.alt, el.className, el.src].join(' ')) || homeLink;
  });
  let logo_url = logoImage ? logoImage.currentSrc || logoImage.src : '';
  if (logo_url && (!/^https?:/.test(logo_url) || /favicon|\.ico(?:[?#]|$)/i.test(logo_url))) logo_url = '';

  return {
    logo_url,
    primary_color: primary || accent_candidates[0] || '',
    accent_color: accent_candidates[0] || '',
    background_color: background,
    foreground_color: foreground,
    secondary_background_color: secondary_background,
    secondary_foreground_color: secondary_foreground,
    font,
  };
})();`
