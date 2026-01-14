# CursedChrome - Enterprise Browser Monitoring & EDR System

CursedChrome is a professional, enterprise-level browser monitoring and Endpoint Detection and Response (EDR) system. It provides real-time visibility into browser activities across the network, including keyboard logging, screen capture, and audio monitoring.

## 🚀 Quick Start

### Prerequisites

- **Backend**: Node.js 12.16.2 (use `nvm use 12.16.2`)
- **Frontend**: Node.js 18+ (use `nvm use 18`)
- **Database**: PostgreSQL (Auto-initialized on first run)

### Running the System

1. **Start Backend Server:**
   ```bash
   nvm use 12.16.2
   npm install
   node server.js
   ```

2. **Start Frontend Management UI:**
   ```bash
   cd gui
   nvm use 18
   npm install
   npm run serve
   ```
   The UI will be available at `http://localhost:8080/`.

## 📦 System Architecture

### Components
- **Backend Server**: Multi-process Node.js server handling WebSockets (4343), HTTP Proxy (8080), and REST API (8118).
- **Management UI**: Modern Vue.js dashboard for real-time monitoring.
- **Chrome Extensions**: Manifest V3 compliant implants for data collection.

### Monitoring Features
- **Keyboard Monitor**: Real-time keystroke logging with visual playback.
- **Screen Capture**: Automated screenshot capturing with change detection.
- **Audio Surveillance**: Live 60s chunked audio recording with waveform visualization.
- **Activity Tracker**: Detailed monitoring of user interactions and tab history.

## 🧪 Testing in Development (VS Code)

Use the pre-configured VS Code Launch settings (`F5`):
- **Launch Full EDR System**: Backend + Frontend UI.
- **Test Main Extension**: Backend + Chrome with monitoring implant.
- **Test Special Features**: Dedicated configs for Cookie Sync and Paywall Bypass.

## 🛡️ Security & Privacy
This system is designed for legitimate enterprise monitoring. All extensions are Manifest V3 compliant, using minimal required permissions and secure communication protocols.

---
*Created for Advanced Agentic Coding - DeepMind Team*
