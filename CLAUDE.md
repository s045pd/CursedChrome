# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

CursedChrome is an enterprise-level browser monitoring EDR (Endpoint Detection and Response) system that consists of:

- **Backend Server**: Node.js cluster-based server with WebSocket communication, HTTP proxy, and API endpoints
- **Frontend GUI**: Vue.js 2 web interface for monitoring and management
- **Chrome Extensions**: Three Manifest V3 extensions for different monitoring capabilities
  - Main extension: Core monitoring and data collection
  - Cookie-sync-extension: Cookie synchronization functionality
  - Bypass-paywalls-chrome: Paywall bypass capabilities

## Development Environment Setup

### Node.js Version Requirements
- **Backend**: Requires Node.js 12.16.2 (use `nvm use 12.16.2`)
- **Frontend**: Requires Node.js 18+ (use `nvm use 18`)

### Starting the System

**Backend Server:**
```bash
nvm use 12.16.2
node server.js
```

**Frontend GUI:**
```bash
cd gui
nvm use 18
npm run serve
```

### VS Code Launch Configurations

Use F5 in VS Code to launch pre-configured test environments:

- **🚀 Launch Full EDR System**: Backend + Frontend UI
- **🧪 Test Bypass Paywalls + Backend**: Backend + Chrome with bypass extension
- **🔧 Test Cookie Sync + Backend**: Backend + Chrome with cookie sync extension  
- **🎯 Test Main Extension + Backend**: Backend + Chrome with main monitoring extension

## Architecture Overview

### Backend Architecture (server.js)
- **Cluster-based**: Multi-process architecture using Node.js cluster module
- **WebSocket Server**: Real-time communication on port 4343
- **HTTP Proxy Server**: AnyProxy-based proxy on port 8080
- **API Server**: REST API on port 8118
- **Database**: Sequelize ORM with PostgreSQL support
- **Redis**: Used for inter-process communication and caching

### Key Backend Components
- `database.js`: Sequelize models (Users, Bots, BotRecording)
- `api-server.js`: Express-based REST API endpoints
- `anyproxy/`: HTTP proxy implementation
- `utils.js`: Shared utilities and configuration

### Frontend Architecture (gui/)
- **Framework**: Vue.js 2 with Vue CLI
- **UI Library**: Bootstrap Vue
- **Build System**: Vue CLI with Babel and ESLint
- **Components**: Modular Vue components for data visualization

### Chrome Extensions Architecture
All extensions use **Manifest V3** with:
- **Service Workers**: Background scripts for persistent functionality
- **Content Security Policy**: Strict CSP for security
- **Declarative Net Request**: Modern request modification API
- **Chrome Storage API**: Persistent data storage

## Common Development Tasks

### Backend Development
```bash
# Install dependencies
npm install

# Start development server
nvm use 12.16.2
node server.js

# The server runs on multiple ports:
# - WebSocket: 4343
# - HTTP Proxy: 8080  
# - API Server: 8118
```

### Frontend Development
```bash
cd gui
npm install
npm run serve    # Development server with hot reload
npm run build    # Production build
npm run lint     # ESLint checking
```

### Extension Development
Extensions are located in:
- `extension/` - Main monitoring extension
- `cookie-sync-extension/` - Cookie synchronization
- `bypass-paywalls-chrome/` - Paywall bypass

Load extensions in Chrome via `chrome://extensions/` in developer mode.

## Important Notes

### Dependencies
- **bcrypt**: Requires architecture-specific compilation. Use `npm rebuild bcrypt` if switching Node versions
- **iconv-lite**: Required for character encoding in proxy functionality
- **Native modules**: May need rebuilding when switching between Node versions

### Security Considerations
This is a legitimate enterprise EDR system for employee browser monitoring. All extensions use minimal required permissions and follow security best practices.

### Database Configuration
- Uses Sequelize ORM with PostgreSQL
- Database initialization handled automatically on first run
- Models defined in `database.js`

### Extension Permissions
Extensions use minimal permissions:
- `activeTab` pattern where possible
- Specific host permissions instead of `<all_urls>` when feasible
- Modern Manifest V3 APIs throughout