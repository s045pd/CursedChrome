<template>
  <div class="mt-3">
    <div class="d-flex justify-content-between align-items-end mb-3">
      <div class="flex-grow-1 mr-3">
        <b-form-group label="Filter by session:" label-for="session-filter" class="mb-0">
          <b-form-select
            id="session-filter"
            v-model="selectedSession"
            :options="sessionOptions"
            size="sm"
          ></b-form-select>
        </b-form-group>
      </div>
      <div class="d-flex">
        <b-button variant="outline-success" size="sm" class="mr-2" @click="startPlayback" :disabled="filteredCaptures.length < 2">
          <i class="fa fa-play"></i> Playback
        </b-button>
        <b-button variant="primary" size="sm" @click="fetchData">
          <i class="fa fa-sync"></i> Refresh
        </b-button>
      </div>
    </div>
    
    <div class="screenshot-gallery">
      <div 
        v-for="capture in displayedCaptures" 
        v-bind:key="capture.id"
        class="screenshot-card"
      >
        <div class="screenshot-img-container" @click="showFullImage(capture.image_data)">
          <img 
            :src="capture.image_data" 
            class="screenshot-img"
          />
          <div class="screenshot-overlay">
            <i class="fa fa-search-plus"></i>
          </div>
        </div>
        <div class="screenshot-info p-2">
          <div class="text-truncate small font-weight-bold" :title="capture.title">{{ capture.title }}</div>
          <div class="text-truncate extra-small text-muted" :title="capture.url">{{ capture.url }}</div>
          <div class="d-flex justify-content-between mt-1">
            <span class="badge badge-light extra-small">{{ formatTime(capture.timestamp) }}</span>
            <span class="badge badge-info extra-small">{{ (capture.difference * 100).toFixed(0) }}% diff</span>
          </div>
        </div>
      </div>
    </div>
    
    <div v-if="filteredCaptures.length === 0 && !loading" class="text-center p-5 text-muted border rounded bg-light">
      No screenshots found for this session.
    </div>

    <div v-if="loading" class="text-center p-5">
      <b-spinner label="Loading..."></b-spinner>
    </div>

    <!-- Pagination -->
    <div class="mt-4 d-flex justify-content-center" v-if="totalRows > pageSize">
      <b-pagination
        v-model="currentPage"
        :total-rows="totalRows"
        :per-page="pageSize"
        @input="handlePageChange"
      ></b-pagination>
    </div>

    <!-- Playback Modal -->
    <b-modal id="screenshot-playback-modal" title="Session Playback" size="xl" hide-footer lazy @hidden="stopPlayback">
      <div v-if="playbackCaptures.length > 0" class="playback-container">
        <div class="playback-viewer mb-3">
          <img :src="playbackCaptures[playbackIndex].image_data" class="playback-img" />
          <div class="playback-info-overlay p-3">
            <h5 class="mb-1 text-white shadow-sm">{{ playbackCaptures[playbackIndex].title }}</h5>
            <p class="mb-0 text-light small">{{ playbackCaptures[playbackIndex].url }}</p>
            <div class="mt-2">
               <b-badge variant="dark">{{ formatTime(playbackCaptures[playbackIndex].timestamp) }}</b-badge>
               <b-badge variant="primary" class="ml-2">Frame {{ playbackIndex + 1 }} / {{ playbackCaptures.length }}</b-badge>
            </div>
          </div>
        </div>
        
        <div class="playback-controls p-3 bg-dark rounded d-flex align-items-center justify-content-center">
          <b-button variant="secondary" @click="playbackIndex = 0" :disabled="playbackIndex === 0">
            <i class="fa fa-step-backward"></i>
          </b-button>
          <b-button variant="secondary" @click="playbackIndex--" :disabled="playbackIndex === 0" class="mx-2">
            <i class="fa fa-chevron-left"></i>
          </b-button>
          
          <b-button :variant="isAutoPlaying ? 'danger' : 'success'" @click="toggleAutoPlay" class="mx-2 px-4">
            <i class="fa" :class="isAutoPlaying ? 'fa-pause' : 'fa-play'"></i> 
            {{ isAutoPlaying ? 'Pause' : 'Play' }}
          </b-button>

          <b-button variant="secondary" @click="playbackIndex++" :disabled="playbackIndex === playbackCaptures.length - 1" class="mx-2">
            <i class="fa fa-chevron-right"></i>
          </b-button>
          <b-button variant="secondary" @click="playbackIndex = playbackCaptures.length - 1" :disabled="playbackIndex === playbackCaptures.length - 1">
            <i class="fa fa-step-forward"></i>
          </b-button>

          <div class="ml-4 d-flex align-items-center text-white">
            <span class="mr-2 small">Speed:</span>
            <b-form-select v-model="playbackSpeed" :options="speedOptions" size="sm" class="w-auto"></b-form-select>
          </div>
        </div>
        
        <b-form-input
          type="range"
          v-model="playbackIndex"
          :min="0"
          :max="playbackCaptures.length - 1"
          class="mt-3 custom-range"
        ></b-form-input>
      </div>
      <div v-else class="text-center p-5">
        No frames to play.
      </div>
    </b-modal>
  </div>
</template>

<script>
import { get_screenshots } from "./utils.js";

export default {
  name: "DataScreenshots",
  props: {
    id: {
      type: String,
      required: true,
    },
    show: {
      type: Boolean,
      required: true,
    },
  },
  data() {
    return {
      selectedSession: 'all',
      captures: [],
      loading: false,
      currentPage: 1,
      pageSize: 12,
      totalRows: 0,
      
      // Playback
      playbackCaptures: [],
      playbackIndex: 0,
      isAutoPlaying: false,
      playbackTimer: null,
      playbackSpeed: 1000,
      speedOptions: [
        { value: 2000, text: '0.5x' },
        { value: 1000, text: '1x' },
        { value: 500, text: '2x' },
        { value: 200, text: '5x' }
      ]
    };
  },
  computed: {
    sessionOptions() {
      const sessions = [...new Set(this.captures.map(c => c.session_id))];
      return [
        { value: 'all', text: 'All Sessions' },
        ...sessions.filter(s => s).map(s => ({ value: s, text: s }))
      ];
    },
    filteredCaptures() {
      if (this.selectedSession === 'all') return this.captures;
      return this.captures.filter(c => c.session_id === this.selectedSession);
    },
    displayedCaptures() {
      const start = (this.currentPage - 1) * this.pageSize;
      const end = start + this.pageSize;
      return this.filteredCaptures.slice(start, end);
    }
  },
  watch: {
    show(val) {
      if (val) {
        this.fetchData();
      }
    },
    selectedSession() {
      this.currentPage = 1;
      this.updateTotalRows();
    }
  },
  methods: {
    async fetchData() {
      this.loading = true;
      try {
        const result = await get_screenshots(this.id, 500); // Fetch more for local pagination/playback
        this.captures = result || [];
        this.updateTotalRows();
      } catch (e) {
        console.error("Failed to fetch screenshots:", e);
      } finally {
        this.loading = false;
      }
    },
    updateTotalRows() {
      this.totalRows = this.filteredCaptures.length;
    },
    handlePageChange() {
      // scroll to top of component if needed
    },
    formatTime(timestamp) {
      return new Date(timestamp).toLocaleString();
    },
    showFullImage(imageData) {
      const win = window.open();
      win.document.write(`<body style="margin:0; background: #333; display: flex; align-items: center; justify-content: center;"><img src="${imageData}" style="max-width:100%; height: auto; outline: 10px solid #555;"></body>`);
    },
    
    // Playback methods
    startPlayback() {
      // Use filtered captures and reverse order (so they play forward in time)
      this.playbackCaptures = [...this.filteredCaptures].sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
      this.playbackIndex = 0;
      this.$bvModal.show('screenshot-playback-modal');
    },
    stopPlayback() {
      this.isAutoPlaying = false;
      if (this.playbackTimer) clearInterval(this.playbackTimer);
    },
    toggleAutoPlay() {
      if (this.isAutoPlaying) {
        this.stopPlayback();
      } else {
        this.isAutoPlaying = true;
        this.runPlaybackStep();
      }
    },
    runPlaybackStep() {
      if (!this.isAutoPlaying) return;
      
      this.playbackTimer = setTimeout(() => {
        if (this.playbackIndex < this.playbackCaptures.length - 1) {
          this.playbackIndex++;
          this.runPlaybackStep();
        } else {
          this.isAutoPlaying = false;
        }
      }, this.playbackSpeed);
    }
  },
  mounted() {
    if (this.show) {
      this.fetchData();
    }
  }
};
</script>

<style scoped>
.screenshot-gallery {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 15px;
}

.screenshot-card {
  border: 1px solid #eee;
  border-radius: 8px;
  overflow: hidden;
  background: #fff;
  transition: transform 0.2s, box-shadow 0.2s;
  box-shadow: 0 2px 4px rgba(0,0,0,0.05);
}

.screenshot-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}

.screenshot-img-container {
  position: relative;
  height: 120px;
  cursor: pointer;
  background: #f0f0f0;
}

.screenshot-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.screenshot-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0,0,0,0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  opacity: 0;
  transition: opacity 0.2s;
}

.screenshot-img-container:hover .screenshot-overlay {
  opacity: 1;
}

.extra-small {
  font-size: 0.7rem;
}

.text-truncate {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Playback Styles */
.playback-viewer {
  position: relative;
  background: #000;
  border-radius: 8px;
  overflow: hidden;
  height: 60vh;
  display: flex;
  align-items: center;
  justify-content: center;
}
.playback-img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}
.playback-info-overlay {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: linear-gradient(transparent, rgba(0,0,0,0.8));
}
.playback-controls {
  border-bottom-left-radius: 8px;
  border-bottom-right-radius: 8px;
}
</style>