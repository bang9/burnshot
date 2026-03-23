// Target zoom to approximate native camera 1x field of view
const TARGET_ZOOM = 2;

let softwareZoom = 1;

export async function initCamera(videoEl) {
  const stream = await navigator.mediaDevices.getUserMedia({
    video: {
      facingMode: { ideal: 'environment' },
      width: { ideal: 1080 },
      height: { ideal: 1080 },
    },
    audio: false,
  });
  videoEl.srcObject = stream;
  await videoEl.play();

  // Try hardware zoom first
  const track = stream.getVideoTracks()[0];
  const caps = track.getCapabilities?.();
  if (caps?.zoom) {
    const zoom = Math.min(TARGET_ZOOM, caps.zoom.max);
    await track.applyConstraints({ advanced: [{ zoom }] });
  } else {
    // Fallback: software crop
    softwareZoom = TARGET_ZOOM;
    videoEl.style.transform = `scale(${softwareZoom})`;
  }
}

export function getCropFactor() {
  return softwareZoom;
}
