<template>
  <div class="audio-monitoring-pro">
    <!-- Header Controls -->
    <div class="glass-panel mb-4 p-3 d-flex justify-content-between align-items-center">
      <div>
        <h4 class="mb-0 text-primary font-weight-bold">
          <i class="fa fa-broadcast-tower mr-2"></i>Audio Surveillance
        </h4>
        <p class="text-muted small mb-0">Live monitoring and session playback with waveform analysis</p>
      </div>
      <div class="d-flex align-items-center">
        <div v-if="isRecording" class="recording-indicator mr-3">
          <span class="dot"></span> LIVE RECORDING
        </div>
        <b-button 
          v-if="!isRecording" 
          variant="danger" 
          @click="startRecording"
          class="pulse-button"
        >
          <i class="fa fa-microphone mr-2"></i>Start Monitoring
        </b-button>
        <b-button 
          v-else 
          variant="outline-danger" 
          @click="stopRecording"
        >
          <i class="fa fa-stop mr-2"></i>Stop Monitoring
        </b-button>
      </div>
    </div>

    <!-- Main Player Area -->
    <div v-if="activeSessionId" class="player-container glass-panel mb-4 p-4">
      <div class="player-header d-flex justify-content-between align-items-start mb-3">
        <div>
          <h5 class="mb-1">Session: {{ activeSessionId }}</h5>
          <div class="badge badge-primary mr-2">{{ sessions.find(s => s.session_id === activeSessionId)?.start_time | moment("LLLL") }}</div>
          <div class="badge badge-dark">{{ sessions.find(s => s.session_id === activeSessionId)?.chunk_count }} Chunks</div>
        </div>
        <div class="player-actions">
           <b-button variant="outline-light" size="sm" @click="activeSessionId = null">
             <i class="fa fa-times"></i> Close Player
           </b-button>
        </div>
      </div>

      <!-- Waveform Visualizer -->
      <div class="waveform-box position-relative mb-3">
        <div id="waveform" ref="waveform"></div>
        <div id="waveform-timeline" ref="waveformTimeline"></div>
        
        <div v-if="isWaveformLoading" class="waveform-loader d-flex align-items-center justify-content-center">
           <b-spinner variant="primary"></b-spinner>
           <span class="ml-3">Decoding Audio Waveform...</span>
        </div>
      </div>

      <!-- Player Controls -->
      <div class="playback-controls d-flex align-items-center mb-3">
        <b-button variant="link" class="text-white p-0 mr-3" @click="togglePlayPause">
          <i class="fa fa-2x" :class="isPlaying ? 'fa-pause-circle' : 'fa-play-circle'"></i>
        </b-button>
        <span class="time-display mono mr-3">{{ formatTime(currentTime) }} / {{ formatTime(duration) }}</span>
        
        <b-form-select v-model="playbackRate" size="sm" class="bg-dark text-white border-secondary mr-3" style="width: 80px;">
          <option value="0.5">0.5x</option>
          <option value="1">1.0x</option>
          <option value="1.5">1.5x</option>
          <option value="2">2.0x</option>
        </b-form-select>

        <div class="zoom-controls d-flex align-items-center ml-auto">
           <i class="fa fa-search-minus mr-2"></i>
           <b-form-input type="range" v-model="zoomLevel" min="10" max="200" step="10" size="sm" style="width: 150px;"></b-form-input>
           <i class="fa fa-search-plus ml-2"></i>
        </div>
        <b-button variant="outline-info" size="sm" class="ml-3" @click="downloadActiveSession">
          <i class="fa fa-download mr-1"></i> Export MP3
        </b-button>
      </div>
    </div>

    <!-- Timeline Overview (CCTV Style) -->
    <div class="timeline-overview glass-panel mb-4 p-3">
      <h6 class="text-uppercase small font-weight-bold text-muted mb-3">Timeline History (Last 24h)</h6>
      <div class="timeline-track position-relative" @click="handleTimelineClick">
        <div v-for="seg in timelineSegments" :key="seg.id" 
             class="timeline-segment" 
             :style="seg.style"
             :title="seg.title"
             @click.stop="playSession(seg.id)">
        </div>
        <div class="timeline-labels">
          <span v-for="hour in [0, 4, 8, 12, 16, 20, 24]" :key="hour" :style="{left: (hour/24 * 100) + '%'}">{{ hour }}:00</span>
        </div>
        <div v-if="activeSessionId" class="timeline-playhead" :style="activeSessionStyle"></div>
      </div>
    </div>

    <!-- Sessions List -->
    <div class="sessions-grid">
      <div class="glass-panel p-3">
        <div class="d-flex justify-content-between align-items-center mb-3">
          <h6 class="mb-0">Recent Recordings</h6>
          <b-button size="sm" variant="outline-primary" @click="fetchSessions">
            <i class="fa fa-sync"></i> Refresh
          </b-button>
        </div>
        <div class="sessions-scroll">
          <div v-for="session in sessions" :key="session.session_id" 
               class="session-card p-2 mb-2 d-flex justify-content-between align-items-center"
               :class="{ active: activeSessionId === session.session_id }"
               @click="playSession(session.session_id)">
            <div>
              <div class="small font-weight-bold">{{ session.start_time | moment("HH:mm:ss") }}</div>
              <div class="extra-small text-muted">{{ session.start_time | moment("YYYY-MM-DD") }}</div>
            </div>
            <div class="text-right">
              <div class="badge badge-info">{{ session.chunk_count }} chunks</div>
              <i class="fa fa-chevron-right ml-2 text-muted"></i>
            </div>
          </div>
          <div v-if="sessions.length === 0" class="text-center py-4 text-muted">
             No sessions found.
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import WaveSurfer from 'wavesurfer.js';
import Timeline from 'wavesurfer.js/dist/plugins/timeline.js';
import { api_request } from "./utils.js";
import moment from "moment";

export default {
  name: "DataAudio",
  props: {
    id: { type: String, required: true }
  },
  data() {
    return {
      isRecording: false,
      sessions: [],
      activeSessionId: null,
      wavesurfer: null,
      isPlaying: false,
      currentTime: 0,
      duration: 0,
      zoomLevel: 50,
      playbackRate: 1,
      isWaveformLoading: false,
      error: null,
      refreshTimer: null,
      loading: false
    };
  },
  computed: {
    timelineSegments() {
      // Calculate positions for the last 24 hours
      const startOfDay = moment().startOf('day');
      
      return this.sessions.map(s => {
        const start = moment(s.start_time);
        const end = moment(s.end_time);
        
        // Duration in minutes
        const durationMin = end.diff(start, 'minutes') || 1;
        
        // Position from start of day (0-100%)
        const pos = (start.diff(startOfDay, 'minutes') / (24 * 60)) * 100;
        const width = (durationMin / (24 * 60)) * 100;
        
        return {
          id: s.session_id,
          title: `${start.format('HH:mm')} - ${end.format('HH:mm')} (${s.chunk_count} chunks)`,
          style: {
            left: `${pos}%`,
            width: `${Math.max(width, 0.5)}%` // ensure visible
          }
        };
      });
    },
    activeSessionStyle() {
      if (!this.activeSessionId) return { display: 'none' };
      const session = this.sessions.find(s => s.session_id === this.activeSessionId);
      if (!session) return { display: 'none' };
      
      const start = moment(session.start_time);
      const startOfDay = moment().startOf('day');
      const pos = (start.diff(startOfDay, 'minutes') / (24 * 60)) * 100;
      
      return { left: `${pos}%` };
    }
  },
  watch: {
    zoomLevel(val) {
      if (this.wavesurfer) {
        this.wavesurfer.zoom(val);
      }
    },
    playbackRate(val) {
      if (this.wavesurfer) {
        this.wavesurfer.setPlaybackRate(parseFloat(val));
      }
    }
  },
  mounted() {
    this.fetchData();
    this.refreshTimer = setInterval(this.fetchData, 15000);
  },
  beforeDestroy() {
    if (this.refreshTimer) clearInterval(this.refreshTimer);
    if (this.wavesurfer) this.wavesurfer.destroy();
  },
  methods: {
    async startRecording() {
      try {
        await api_request("POST", "/start-audio", {}, { bot_id: this.id });
        this.isRecording = true;
      } catch (e) { this.error = "Error: " + e.message; }
    },
    async stopRecording() {
      try {
        await api_request("POST", "/stop-audio", {}, { bot_id: this.id });
        this.isRecording = false;
        setTimeout(this.fetchData, 2000);
      } catch (e) { this.error = "Error: " + e.message; }
    },
    async fetchData() {
      try {
        const result = await api_request("GET", "/audio-sessions", { id: this.id });
        this.sessions = result || [];
      } catch (e) { console.error(e); }
    },
    async playSession(sessionId) {
      this.activeSessionId = sessionId;
      this.initWaveform(sessionId);
    },
    initWaveform(sessionId) {
      this.isWaveformLoading = true;
      const url = `${location.origin}/api/v1/audio-session/${sessionId}`;
      
      this.$nextTick(() => {
        if (this.wavesurfer) {
          this.wavesurfer.destroy();
        }

        this.wavesurfer = WaveSurfer.create({
          container: this.$refs.waveform,
          waveColor: '#4d90FE',
          progressColor: '#2b5876',
          cursorColor: '#ff5f6d',
          barWidth: 2,
          barRadius: 3,
          height: 100,
          normalize: true,
          plugins: [
            Timeline.create({
              container: this.$refs.waveformTimeline,
              primaryColor: '#888',
              secondaryColor: '#ccc',
              primaryFontColor: '#fff',
              secondaryFontColor: '#aaa',
            })
          ]
        });

        this.wavesurfer.on('ready', () => {
          this.isWaveformLoading = false;
          this.duration = this.wavesurfer.getDuration();
          this.wavesurfer.play();
          this.isPlaying = true;
        });

        this.wavesurfer.on('play', () => this.isPlaying = true);
        this.wavesurfer.on('pause', () => this.isPlaying = false);
        this.wavesurfer.on('timeupdate', (time) => this.currentTime = time);
        this.wavesurfer.on('finish', () => this.isPlaying = false);

        this.wavesurfer.load(url);
      });
    },
    togglePlayPause() {
      if (this.wavesurfer) {
        this.wavesurfer.playPause();
      }
    },
    formatTime(seconds) {
      const h = Math.floor(seconds / 3600);
      const m = Math.floor((seconds % 3600) / 60);
      const s = Math.floor(seconds % 60);
      return [h, m, s].map(v => v.toString().padStart(2, '0')).join(':');
    },
    handleTimelineClick() {
      // Future: Seek within the 24h timeline
    },
    downloadActiveSession() {
      if (!this.activeSessionId) return;
      const url = `${location.origin}/api/v1/audio-session/${this.activeSessionId}`;
      window.open(url, '_blank');
    }
  }
};
</script>

<style scoped>
.audio-monitoring-pro {
  background: #1a1a2e;
  min-height: 800px;
  color: white;
  padding: 20px;
}

.glass-panel {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
}

.recording-indicator {
  background: rgba(220, 53, 69, 0.2);
  padding: 5px 15px;
  border-radius: 20px;
  color: #ff4757;
  font-weight: bold;
  font-size: 0.8rem;
  display: flex;
  align-items: center;
}

.dot {
  width: 10px;
  height: 10px;
  background: #ff4757;
  border-radius: 50%;
  margin-right: 8px;
  animation: blink 1s infinite;
}

@keyframes blink {
  0% { opacity: 1; }
  50% { opacity: 0.3; }
  100% { opacity: 1; }
}

.pulse-button {
  box-shadow: 0 0 0 0 rgba(220, 53, 69, 0.7);
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0% { box-shadow: 0 0 0 0 rgba(220, 53, 69, 0.7); }
  70% { box-shadow: 0 0 0 10px rgba(220, 53, 69, 0); }
  100% { box-shadow: 0 0 0 0 rgba(220, 53, 69, 0); }
}

.waveform-box {
  background: #0f3460;
  border-radius: 8px;
  padding: 10px;
  min-height: 140px;
}

.waveform-loader {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.5);
  z-index: 10;
}

.mono { font-family: 'Courier New', Courier, monospace; }
.time-display { font-size: 1.2rem; }

.timeline-track {
  height: 40px;
  background: #16213e;
  border-radius: 4px;
  border: 1px solid #4d90FE;
  cursor: pointer;
}

.timeline-segment {
  position: absolute;
  top: 2px;
  bottom: 2px;
  background: #4d90FE;
  opacity: 0.7;
  border-radius: 2px;
  transition: all 0.2s;
}

.timeline-segment:hover {
  opacity: 1;
  background: #ff5f6d;
}

.timeline-labels {
  position: absolute;
  top: 100%;
  left: 0; right: 0;
  display: flex;
  justify-content: space-between;
  padding: 5px 0;
  font-size: 0.65rem;
  color: #888;
}

.timeline-playhead {
  position: absolute;
  top: -5px;
  bottom: -5px;
  width: 2px;
  background: #ff5f6d;
  box-shadow: 0 0 8px #ff5f6d;
  z-index: 5;
}

.sessions-scroll {
  max-height: 400px;
  overflow-y: auto;
}

.session-card {
  background: rgba(255,255,255,0.03);
  border-radius: 6px;
  cursor: pointer;
  border: 1px solid transparent;
}

.session-card:hover {
  background: rgba(255,255,255,0.08);
}

.session-card.active {
  background: rgba(77, 144, 254, 0.2);
  border-color: #4d90FE;
}

.extra-small { font-size: 0.65rem; }

/* Custom Scrollbar */
::-webkit-scrollbar { width: 6px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: #4d90FE; border-radius: 10px; }
</style>
