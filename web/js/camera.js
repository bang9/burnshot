let videoElement = null;
let videoTrack = null;
let hwZoomRange = null; // { min, max } if hardware zoom supported
let currentZoom = 1;
let usingHardwareZoom = false;

export async function initCamera(videoEl) {
  videoElement = videoEl;
  const stream = await navigator.mediaDevices.getUserMedia({
    video: {
      facingMode: { ideal: 'environment' },
      width: { ideal: 1080 },
      height: { ideal: 1080 },
    },
    audio: false,
  });
  videoEl.srcObject = stream;
  await new Promise((resolve, reject) => {
    videoEl.addEventListener('loadedmetadata', () => {
      videoEl.play().then(resolve).catch(reject);
    }, { once: true });
  });

  // Detect hardware zoom support
  videoTrack = stream.getVideoTracks()[0];
  try {
    const caps = videoTrack.getCapabilities?.();
    if (caps?.zoom) {
      hwZoomRange = { min: caps.zoom.min, max: caps.zoom.max };
    }
  } catch { /* no hardware zoom */ }
}

export async function setZoom(level) {
  currentZoom = level;
  usingHardwareZoom = false;

  if (hwZoomRange) {
    const clamped = Math.min(Math.max(level, hwZoomRange.min), hwZoomRange.max);
    try {
      await videoTrack.applyConstraints({ advanced: [{ zoom: clamped }] });
      videoElement.style.transform = '';
      usingHardwareZoom = true;
      return;
    } catch { /* fallback to software */ }
  }

  // Software zoom: only scale >= 1 (sub-1x requires hardware)
  const swLevel = Math.max(level, 1);
  videoElement.style.transform = swLevel === 1 ? '' : `scale(${swLevel})`;
}

export function hasHardwareZoom() {
  return hwZoomRange !== null;
}

export function getCropFactor() {
  if (usingHardwareZoom) return 1;
  return Math.max(currentZoom, 1);
}
