import { trackEvent } from './firebase.js';
import { getCropFactor } from './camera.js';

export function initCapture(button, video, overlayCanvas, compositeCanvas) {
  const ctx = compositeCanvas.getContext('2d');
  const SIZE = 1080;

  button.addEventListener('click', () => {
    trackEvent('photo_captured');

    // Draw camera frame (center-cropped to 1:1, with zoom crop factor)
    const vw = video.videoWidth;
    const vh = video.videoHeight;
    const side = Math.min(vw, vh) / getCropFactor();
    const sx = (vw - side) / 2;
    const sy = (vh - side) / 2;
    ctx.drawImage(video, sx, sy, side, side, 0, 0, SIZE, SIZE);

    // Draw overlay on top
    ctx.drawImage(overlayCanvas, 0, 0, SIZE, SIZE);

    // Convert to blob and trigger save/share
    compositeCanvas.toBlob(async (blob) => {
      const file = new File([blob], 'burnshot.png', { type: 'image/png' });

      // Try Web Share API first (mobile)
      if (navigator.canShare?.({ files: [file] })) {
        try {
          await navigator.share({ files: [file] });
          trackEvent('photo_shared', { method: 'web_share' });
          return;
        } catch {
          // User cancelled or share failed — fall through to download
        }
      }

      // Fallback: download
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'burnshot.png';
      a.click();
      URL.revokeObjectURL(url);
      trackEvent('photo_shared', { method: 'download' });
    }, 'image/png');
  });
}
