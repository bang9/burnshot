import { TEMPLATES } from './templates.js';

let currentTemplate = 0;
let animFrameId = null;
let currentData = null;

export function setTemplate(index) {
  currentTemplate = index;
}

export function startOverlay(canvas, video, data) {
  currentData = data;
  const ctx = canvas.getContext('2d');

  function draw() {
    const w = canvas.clientWidth;
    const h = canvas.clientHeight;
    if (canvas.width !== w * 2 || canvas.height !== h * 2) {
      canvas.width = w * 2;
      canvas.height = h * 2;
    }
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    ctx.save();
    ctx.scale(2, 2); // retina

    const template = TEMPLATES[currentTemplate];
    template.render(ctx, w, h, currentData);

    ctx.restore();
    animFrameId = requestAnimationFrame(draw);
  }

  draw();
}

export function stopOverlay() {
  if (animFrameId) cancelAnimationFrame(animFrameId);
}

export function getCurrentTemplate() {
  return currentTemplate;
}
