import { initializeApp } from 'https://www.gstatic.com/firebasejs/11.4.0/firebase-app.js';
import { getAnalytics, logEvent } from 'https://www.gstatic.com/firebasejs/11.4.0/firebase-analytics.js';

const firebaseConfig = {
  apiKey: "AIzaSyArT5vhfAJl5EeYAUrrXwagnie5XJAAJBc",
  authDomain: "ai-tools-bang9.firebaseapp.com",
  projectId: "ai-tools-bang9",
  storageBucket: "ai-tools-bang9.firebasestorage.app",
  messagingSenderId: "955089984529",
  appId: "1:955089984529:web:70f71bf44ddf7f0ea2e5b0",
  measurementId: "G-6BCN3CW732"
};

const app = initializeApp(firebaseConfig);
const analytics = getAnalytics(app);

// Re-export for use by other modules
export function trackEvent(name, params) {
  logEvent(analytics, name, params);
}

// Auto-track page view
trackEvent('page_view', { page: window.location.pathname });
