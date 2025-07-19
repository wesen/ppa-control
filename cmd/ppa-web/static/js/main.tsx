import { render } from 'preact';
import { App } from './components/App';

// Import Bootstrap JavaScript for interactive components
import 'bootstrap/dist/js/bootstrap.bundle.min.js';

// Render the app
const root = document.getElementById('app');
if (root) {
    render(<App />, root);
} else {
    console.error('Root element not found');
}

// Add any global event listeners or initialization here
document.addEventListener('DOMContentLoaded', () => {
    console.log('PPA Packet Analysis Dashboard loaded');
});

// Handle any global errors
window.addEventListener('error', (event) => {
    console.error('Global error:', event.error);
});

window.addEventListener('unhandledrejection', (event) => {
    console.error('Unhandled promise rejection:', event.reason);
});
