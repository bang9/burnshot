import { trackEvent } from './firebase.js';

export function initSwipe(container, templates, setTemplate, dotsContainer) {
  let current = 0;

  // Create dots
  templates.forEach((_, i) => {
    const dot = document.createElement('div');
    dot.className = 'dot' + (i === 0 ? ' active' : '');
    dotsContainer.appendChild(dot);
  });

  function updateDots() {
    const dots = dotsContainer.querySelectorAll('.dot');
    dots.forEach((d, i) => d.classList.toggle('active', i === current));
  }

  function switchTo(index) {
    current = index;
    setTemplate(current);
    updateDots();
    trackEvent('template_switch', { template: templates[current].name });
  }

  // Touch swipe
  let startX = 0;
  container.addEventListener('touchstart', (e) => {
    startX = e.touches[0].clientX;
  }, { passive: true });

  container.addEventListener('touchend', (e) => {
    const diff = e.changedTouches[0].clientX - startX;
    if (Math.abs(diff) < 50) return;

    if (diff < 0 && current < templates.length - 1) {
      switchTo(current + 1);
    } else if (diff > 0 && current > 0) {
      switchTo(current - 1);
    }
  }, { passive: true });

  // Keyboard fallback (for desktop testing)
  document.addEventListener('keydown', (e) => {
    if (e.key === 'ArrowRight' && current < templates.length - 1) {
      switchTo(current + 1);
    } else if (e.key === 'ArrowLeft' && current > 0) {
      switchTo(current - 1);
    }
  });
}
