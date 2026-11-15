import React, { useEffect, useRef } from 'react';

const P = 'https://images.higgs.ai/?default=1&output=webp&url=https%3A%2F%2Fd8j0ntlcm91z4.cloudfront.net%2Fuser_38xzZboKViGWJOttwIXH07lWA1P%2Fhf_';
const SHIP = P + '20261031_024714_39df675e-5060-4528-87dc-9d0c274815aa.png&w=1600&q=90';
const CAPT = P + '20261031_024703_a0298b6a-0117-47f1-8462-4be671d0407e.png&w=2048&q=90';
const PLANTS = P + '20261031_024704_3ccf1ac2-bec7-4fdb-bfce-ac19eca8db28.png&w=2048&q=90';

const SHEET = { w: 2048, h: 1360 };
const BOX = {
  capt:  [CAPT, 34, 153, 1246, 1074],
  chest: [CAPT, 1475, 830, 446, 392],
  palm:  [PLANTS, 19, 32, 910, 1284],
  bushL: [PLANTS, 515, 820, 860, 496],
  bushR: [PLANTS, 1528, 908, 499, 405]
};

const LAY = {
  land: {
    w: 1600, h: 900,
    wave: { y0: 748, y1: 272, gap0: 1, gap1: .6 },
    ship: { x0: 1130, w0: 710, x1: 1345, w1: 176 },
    palm: { x: -92, y: -112, w: 620 }, capt: { x: 120, y: 318, w: 540 }, chest: { x: 640, y: 628, w: 200 },
    bushL: { x: 96, y: 716, w: 700 }, bushR: { x: 1370, y: 740, w: 300 }
  },
  port: {
    w: 800, h: 1500,
    wave: { y0: 1300, y1: 430, gap0: 1, gap1: .7 },
    ship: { x0: 420, w0: 620, x1: 640, w1: 150 },
    palm: { x: -150, y: -60, w: 560 }, capt: { x: 20, y: 440, w: 600 }, chest: { x: 560, y: 830, w: 190 },
    bushL: { x: -40, y: 1330, w: 520 }, bushR: { x: 570, y: 1340, w: 260 }
  }
};

const BACK = [
  { dy: -100, a: 20, k: 1.5, v: .9, ph: 0, fill: 'var(--band)' },
  { dy: -52, a: 23, k: 1.9, v: 1.15, ph: 1.7, fill: 'var(--paper)' }
];
const FRONT = [
  { dy: 0, a: 25, k: 1.7, v: 1.3, ph: 3.1, fill: 'var(--band2)' },
  { dy: 26, a: 18, k: 2.4, v: 1, ph: 4.4, span: [.06, .44] },
  { dy: 48, a: 22, k: 1.45, v: 1.4, ph: .6, fill: 'var(--paper)' },
  { dy: 76, a: 15, k: 2.1, v: 1.2, ph: 5.1, span: [.52, .97] },
  { dy: 104, a: 24, k: 1.8, v: 1.25, ph: 2.2, fill: 'var(--paper)' },
  { dy: 130, a: 14, k: 2.7, v: 1, ph: 1.1, span: [.0, .62] },
  { dy: 158, a: 20, k: 1.6, v: 1.1, ph: 3.8 }
];

export default function SaltwakeHero({ onExploreShop, onOpenCrew }) {
  const containerRef = useRef(null);

  useEffect(() => {
    const frame = containerRef.current.querySelector('#frame');
    const scene = containerRef.current.querySelector('#scene');
    const wBack = containerRef.current.querySelector('#wBack');
    const wFront = containerRef.current.querySelector('#wFront');
    const shipEl = containerRef.current.querySelector('#ship');
    const cloudA = containerRef.current.querySelector('#cloudA');
    const t1El = containerRef.current.querySelector('#t1');
    const t2El = containerRef.current.querySelector('#t2');
    const h2El = containerRef.current.querySelector('#h2');
    const p2El = containerRef.current.querySelector('#p2');
    const btn2El = containerRef.current.querySelector('#btn2');
    const logoEl = containerRef.current.querySelector('.logo');
    const shadeCanvas = containerRef.current.querySelector('#shade');
    const curEl = containerRef.current.querySelector('#cur');
    const pingEl = containerRef.current.querySelector('#ping');

    shipEl.style.backgroundImage = `url("${SHIP}")`;
    for (const [id, [url, x, y, w, h]] of Object.entries(BOX)) {
      const el = containerRef.current.querySelector('#' + id);
      if (!el) continue;
      el.style.backgroundImage = `url("${url}")`;
      const bgW = (SHEET.w / w) * 100;
      const bgH = (SHEET.h / h) * 100;
      const posX = (x / (SHEET.w - w)) * 100;
      const posY = (y / (SHEET.h - h)) * 100;
      el.style.backgroundSize = `${bgW}% ${bgH}%`;
      el.style.backgroundPosition = `${posX}% ${posY}%`;
    }

    const clamp = (v, min = 0, max = 1) => Math.min(Math.max(v, min), max);
    const mix = (a, b, t) => a + (b - a) * t;
    const seg = (p, a, b) => clamp((p - a) / (b - a));
    const io = (k) => (k < 0.5 ? 4 * k * k * k : 1 - Math.pow(-2 * k + 2, 3) / 2);
    const out = (k) => 1 - Math.pow(1 - k, 3);
    const back = (k) => 1 + 2.4 * Math.pow(k - 1, 3) + 1.4 * Math.pow(k - 1, 2);

    let W = 0, H = 0, portrait = false, L = LAY.land, S = 1, DW = 0, DH = 0, DX = 0, DY = 0;
    const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

    // Shade canvas init
    const ctx = shadeCanvas.getContext('2d');
    ctx.save();
    ctx.scale(0.5, 0.5);
    ctx.fillStyle = '#ffffff';
    ctx.fillRect(0, 0, 128, 72);
    ctx.strokeStyle = 'rgba(120,114,96,.9)';
    ctx.lineCap = 'round';
    const curves = [
      [-10, 26, 38, -8, 8],
      [-8, 46, 58, -2, 5],
      [-10, 62, 74, 6, 4],
      [14, 78, 96, 16, 4],
      [40, 82, 118, 22, 3],
      [78, 80, 140, 34, 5],
      [-12, 74, 22, 56, 6]
    ];
    for (const [x0, y0, x1, y1, lw] of curves) {
      ctx.lineWidth = lw;
      ctx.beginPath();
      ctx.moveTo(x0, y0);
      ctx.quadraticCurveTo((x0 + x1) / 2 + 6, (y0 + y1) / 2 + 8, x1, y1);
      ctx.stroke();
    }
    ctx.restore();

    function resize() {
      if (!frame) return;
      W = frame.clientWidth;
      H = frame.clientHeight;
      portrait = H > W;
      document.documentElement.classList.toggle('portrait', portrait);
      L = portrait ? LAY.port : LAY.land;
      S = Math.min(W / L.w, H / L.h);
      DW = W / S;
      DH = H / S;
      DX = (DW - L.w) / 2;
      DY = (DH - L.h) / 2;
      scene.style.transform = `translate(${DX * S}px,${DY * S}px) scale(${S})`;

      for (const [id, [_, __, ___, w, h]] of Object.entries(BOX)) {
        const el = containerRef.current.querySelector('#' + id);
        const cfg = L[id];
        if (!el || !cfg) continue;
        const elH = cfg.w * (h / w);
        el.style.width = `${cfg.w}px`;
        el.style.height = `${elH}px`;
        el.style.left = `${cfg.x}px`;
        el.style.top = `${cfg.y}px`;
      }

      wBack.setAttribute('width', DW);
      wBack.setAttribute('height', DH);
      wFront.setAttribute('width', DW);
      wFront.setAttribute('height', DH);
    }

    window.addEventListener('resize', resize);
    resize();

    function createWavePaths(svg, layers) {
      svg.innerHTML = '';
      return layers.map(l => {
        const p = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        p.setAttribute('stroke', 'var(--line)');
        p.setAttribute('stroke-width', '3.2');
        p.setAttribute('stroke-linecap', 'round');
        p.setAttribute('fill', l.fill || 'none');
        svg.appendChild(p);
        return { layer: l, path: p };
      });
    }
    const backPaths = createWavePaths(wBack, BACK);
    const frontPaths = createWavePaths(wFront, FRONT);

    function yAt(l, x, t, p) {
      const wave = L.wave;
      const base = mix(wave.y0, wave.y1, p) + l.dy * mix(wave.gap0, wave.gap1, p);
      const a = l.a * mix(1, 0.72, p);
      const f = (2 * Math.PI / 1100) * l.k * (1 - 0.35 * p);
      return base + a * Math.sin(f * x + t * l.v + l.ph) + a * 0.45 * Math.sin(f * 2.3 * x - t * l.v * 0.7 + l.ph * 1.7);
    }

    let p = 0, pFrom = 0, pTo = 0, pT0 = -99999;
    const t0 = performance.now();
    let wheelLock = 0;

    function go(to) {
      if (to === pTo) return;
      pFrom = p;
      pTo = to;
      pT0 = performance.now();
    }

    const clickHandler = (e) => {
      const target = e.target.closest('[data-go]');
      if (target) {
        e.preventDefault();
        go(parseInt(target.getAttribute('data-go'), 10));
      }
    };
    frame.addEventListener('click', clickHandler);

    const wheelHandler = (e) => {
      const now = performance.now();
      if (now < wheelLock) return;
      if (e.deltaY > 8) {
        go(1);
        wheelLock = now + 1200;
      } else if (e.deltaY < -8) {
        go(0);
        wheelLock = now + 1200;
      }
    };
    window.addEventListener('wheel', wheelHandler, { passive: true });

    let mouseX = -100, mouseY = -100, curX = -100, curY = -100;
    let isPressed = false, isHoveringLink = false;
    let pingStartTime = -99999, pingX = 0, pingY = 0;

    if (window.matchMedia('(hover:hover) and (pointer:fine)').matches) {
      document.documentElement.classList.add('fine');
      window.addEventListener('pointermove', (e) => {
        mouseX = e.clientX;
        mouseY = e.clientY;
        isHoveringLink = !!e.target.closest('a, button, [data-go], .pill');
      });
      window.addEventListener('pointerdown', (e) => {
        isPressed = true;
        pingX = e.clientX;
        pingY = e.clientY;
        pingStartTime = performance.now();
      });
      window.addEventListener('pointerup', () => { isPressed = false; });
    }

    let animId;
    let lastNow = performance.now();

    function render(now) {
      const dt = Math.min((now - lastNow) / 1000, 0.05);
      lastNow = now;
      const t = reduced ? 0 : (now - t0) / 1000;

      const DUR = reduced ? 1 : 2100;
      const k = seg(now, pT0, pT0 + DUR);
      p = mix(pFrom, pTo, io(k));

      const xStart = -DX - 20;
      const xEnd = L.w + DX + 20;
      const step = 22;
      const bottomH = L.h + DY + 40;

      for (const { layer, path } of backPaths) {
        let x0 = xStart, x1 = xEnd;
        if (layer.span) {
          const fullW = xEnd - xStart;
          x0 = xStart + fullW * layer.span[0];
          x1 = xStart + fullW * layer.span[1];
        }
        let d = `M ${x0.toFixed(1)} ${yAt(layer, x0, t, p).toFixed(1)}`;
        for (let x = x0 + step; x < x1; x += step) {
          d += ` L ${x.toFixed(1)} ${yAt(layer, x, t, p).toFixed(1)}`;
        }
        d += ` L ${x1.toFixed(1)} ${yAt(layer, x1, t, p).toFixed(1)}`;
        if (layer.fill) {
          d += ` L ${x1.toFixed(1)} ${bottomH.toFixed(1)} L ${x0.toFixed(1)} ${bottomH.toFixed(1)} Z`;
        }
        path.setAttribute('d', d);
      }

      for (const { layer, path } of frontPaths) {
        let x0 = xStart, x1 = xEnd;
        if (layer.span) {
          const fullW = xEnd - xStart;
          x0 = xStart + fullW * layer.span[0];
          x1 = xStart + fullW * layer.span[1];
        }
        let d = `M ${x0.toFixed(1)} ${yAt(layer, x0, t, p).toFixed(1)}`;
        for (let x = x0 + step; x < x1; x += step) {
          d += ` L ${x.toFixed(1)} ${yAt(layer, x, t, p).toFixed(1)}`;
        }
        d += ` L ${x1.toFixed(1)} ${yAt(layer, x1, t, p).toFixed(1)}`;
        if (layer.fill) {
          d += ` L ${x1.toFixed(1)} ${bottomH.toFixed(1)} L ${x0.toFixed(1)} ${bottomH.toFixed(1)} Z`;
        }
        path.setAttribute('d', d);
      }

      const c = L.ship;
      const sw = mix(c.w0, c.w1, p);
      const sx = mix(c.x0, c.x1, p);
      const sh = sw * 2002 / 1961;
      const wy = yAt(FRONT[0], sx, t, p);
      const top = wy + sw * 0.05 - sh;
      const left = sx - sw / 2;
      const slope = (yAt(FRONT[0], sx + 90, t, p) - yAt(FRONT[0], sx - 90, t, p)) / 180;
      const rot = Math.atan(slope) * 57.3 * 0.75;
      shipEl.style.width = `${sw}px`;
      shipEl.style.height = `${sh}px`;
      shipEl.style.left = `${left}px`;
      shipEl.style.top = `${top}px`;
      shipEl.style.transform = `rotate(${rot}deg)`;

      const cloudLeft = ((L.w * 0.72 - t * 16) % (L.w + 700)) - 400;
      const cloudTop = portrait ? 110 : 150;
      cloudA.style.left = `${cloudLeft}px`;
      cloudA.style.top = `${cloudTop}px`;
      cloudA.style.opacity = (1 - seg(p, 0.3, 0.7)).toFixed(3);

      const intro = out(seg(now, t0 + 200, t0 + 1100));
      const t1Opacity = intro * (1 - seg(p, 0.62, 0.7));
      const t1Y = (1 - intro) * 26 - p * 40;
      t1El.style.opacity = t1Opacity.toFixed(3);
      t1El.style.transform = `translateY(${t1Y.toFixed(1)}px)`;
      t1El.style.pointerEvents = p > 0.1 ? 'none' : 'auto';

      logoEl.style.opacity = (1 - seg(p, 0.12, 0.3)).toFixed(3);

      const e = io(seg(p, 0.04, 0.92));
      const palmEl = containerRef.current.querySelector('#palm');
      const captEl = containerRef.current.querySelector('#capt');
      const chestEl = containerRef.current.querySelector('#chest');
      const bushLEl = containerRef.current.querySelector('#bushL');
      const bushREl = containerRef.current.querySelector('#bushR');

      palmEl.style.transform = `translate(${mix(-L.palm.w - 80, 0, e).toFixed(1)}px,${mix(-170, 0, e).toFixed(1)}px) rotate(${(Math.sin(t * 0.8) * 0.7).toFixed(2)}deg)`;
      captEl.style.transform = `translateX(${mix(-L.capt.w - 160, 0, e).toFixed(1)}px) rotate(${(Math.sin(t * 1.1) * 0.5 * p).toFixed(2)}deg)`;
      chestEl.style.transform = `translate(${mix(-L.chest.w - 720, 0, e).toFixed(1)}px,${mix(120, 0, e).toFixed(1)}px)`;

      const b = out(seg(p, 0.5, 1));
      bushLEl.style.transform = `translateY(${mix(340, 0, b).toFixed(1)}px)`;
      bushREl.style.transform = `translateY(${mix(300, 0, b).toFixed(1)}px)`;

      const hk = seg(p, 0.42, 0.82);
      h2El.style.opacity = (hk * hk).toFixed(3);
      h2El.style.transform = `translateY(${mix(70, 0, out(hk)).toFixed(1)}px)`;

      const pk = seg(p, 0.76, 0.92);
      p2El.style.opacity = pk.toFixed(3);
      p2El.style.transform = `translateY(${mix(16, 0, pk).toFixed(1)}px)`;

      const bk = seg(p, 0.86, 1);
      btn2El.style.opacity = bk.toFixed(3);
      btn2El.style.transform = `scale(${mix(0.7, 1, back(bk)).toFixed(3)})`;
      t2El.style.pointerEvents = p > 0.9 ? 'auto' : 'none';

      shadeCanvas.style.transform = `translate(${(Math.sin(t * 0.35) * 2.2).toFixed(2)}%, ${(Math.cos(t * 0.27) * 1.6).toFixed(2)}%) rotate(${(Math.sin(t * 0.22) * 0.8).toFixed(2)}deg)`;

      if (curEl && curEl.style.display !== 'none') {
        const f = 1 - Math.exp(-dt * 22);
        curX += (mouseX - curX) * f;
        curY += (mouseY - curY) * f;
        const curScale = isPressed ? 0.7 : (isHoveringLink ? 1.35 : 1);
        curEl.style.transform = `translate(${curX.toFixed(1)}px, ${curY.toFixed(1)}px) scale(${curScale})`;

        if (now - pingStartTime < 650) {
          const pingK = (now - pingStartTime) / 650;
          pingEl.style.display = 'block';
          pingEl.style.transform = `translate(${pingX}px, ${pingY}px) scale(${(1 + out(pingK) * 1.9).toFixed(3)})`;
          pingEl.style.opacity = (1 - pingK).toFixed(3);
        } else {
          pingEl.style.display = 'none';
        }
      }

      animId = requestAnimationFrame(render);
    }

    animId = requestAnimationFrame(render);

    return () => {
      cancelAnimationFrame(animId);
      window.removeEventListener('resize', resize);
      window.removeEventListener('wheel', wheelHandler);
      frame.removeEventListener('click', clickHandler);
    };
  }, []);

  return (
    <div ref={containerRef} className="saltwake-root">
      <main className="frame" id="frame">
        <div className="logo">
          <svg viewBox="0 0 26 30" fill="none" stroke="currentColor" strokeWidth="2.4" strokeLinecap="round">
            <circle cx="13" cy="4.5" r="2.8"/>
            <path d="M13 7.5V28M7 12h12M2.5 19c.6 5.500 5.500 9 10.500 9s9.900-3.500 10.500-9"/>
            <path d="m1 21 1.500-2.500L5 20.500M25 21l-1.500-2.500-2.500 2"/>
          </svg>
          Saltwake
        </div>

        <nav>
          <a href="#" data-go="0">Voyages
            <svg viewBox="0 0 17 21" fill="none" stroke="currentColor" strokeWidth="2.6" strokeLinecap="round">
              <circle cx="8.5" cy="3" r="2"/><path d="M8.5 5v14M4 8h9M1 13c.4 3.8 3.8 6.5 7.5 6.5s7.1-2.7 7.5-6.5"/>
            </svg>
          </a>
          <a href="#" data-go="1">Crew
            <svg viewBox="0 0 17 21" fill="none" stroke="currentColor" strokeWidth="2.6" strokeLinecap="round">
              <circle cx="8.5" cy="3" r="2"/><path d="M8.5 5v14M4 8h9M1 13c.4 3.8 3.8 6.5 7.5 6.5s7.1-2.7 7.5-6.5"/>
            </svg>
          </a>
          <a href="#" onClick={(e) => { e.preventDefault(); onExploreShop(); }}>Store
            <svg viewBox="0 0 17 21" fill="none" stroke="currentColor" strokeWidth="2.6" strokeLinecap="round">
              <circle cx="8.5" cy="3" r="2"/><path d="M8.5 5v14M4 8h9M1 13c.4 3.8 3.8 6.5 7.5 6.5s7.1-2.7 7.5-6.5"/>
            </svg>
          </a>
          <a href="#" onClick={(e) => { e.preventDefault(); onOpenCrew(); }}>Account
            <svg viewBox="0 0 17 21" fill="none" stroke="currentColor" strokeWidth="2.6" strokeLinecap="round">
              <circle cx="8.5" cy="3" r="2"/><path d="M8.5 5v14M4 8h9M1 13c.4 3.8 3.8 6.5 7.5 6.5s7.1-2.7 7.5-6.5"/>
            </svg>
          </a>
        </nav>

        <div id="scene">
          <div className="spr" id="cloudA">
            <svg viewBox="0 0 300 80" width="300" height="80">
              <path fill="#d3cebe" d="M0 79c14-3 26-3 38-10-4-16 10-30 30-26 6-22 42-34 62-16 16-12 42-6 50 14 10-2 18 2 22 8 24 22 66 26 98 30z"/>
            </svg>
          </div>

          <svg className="w" id="wBack"></svg>

          <div className="t1" id="t1">
            <div className="eb">Port Royal</div>
            <h1>Tales of<br/>the open sea</h1>
            <a className="pill" href="#" data-go="1">Set sail</a>
          </div>

          <div className="spr" id="ship"></div>

          <svg className="w" id="wFront"></svg>

          <div className="spr" id="palm"></div>
          <div className="spr" id="capt"></div>
          <div className="spr" id="chest"></div>
          <div className="spr" id="bushL"></div>
          <div className="spr" id="bushR"></div>

          <div className="t2" id="t2">
            <h2 id="h2">Rules of<br/>the crew</h2>
            <p id="p2">Every sailor signed the articles<br/>before the anchor came up</p>
            <span className="btnw" id="btn2">
              <i className="tick a"></i><i className="tick b"></i><i className="tick c"></i>
              <button className="pill" onClick={onExploreShop}>Enter Bazaars</button>
            </span>
          </div>
        </div>

        <canvas id="shade" width="64" height="36"></canvas>
      </main>

      <div id="ping"></div>
      <div id="cur"></div>
    </div>
  );
}
