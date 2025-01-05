<template>
  <div>
    <b-row class="mb-3">
      <b-col md="6">
        <b-form-group label="Date Range">
          <b-form-datepicker
            v-model="dateRange.startDate"
            :max="dateRange.endDate"
            placeholder="Start Date"
            class="mb-2"
          ></b-form-datepicker>
          <b-form-datepicker
            v-model="dateRange.endDate"
            :min="dateRange.startDate"
            placeholder="End Date"
          ></b-form-datepicker>
        </b-form-group>
      </b-col>
      <b-col md="6" class="d-flex align-items-end">
        <div class="mr-3">
          <b-form-checkbox v-model="useMp3Local">
            Use MP3 Format
          </b-form-checkbox>
        </div>
        <b-button @click="fetchData" variant="primary" class="mr-2">Search</b-button>
        <b-button @click="downloadData" variant="primary">Download</b-button>
      </b-col>
    </b-row>

    <div style="overflow-x: auto">
      <b-table
        id="recording_table"
        :items="info"
        :fields="recording_fields"
        :total-rows="totalRows"
        small
        hover
      >
        <template #cell(date)="data">
          {{ data.value }}
        </template>

        <template #cell(data)="data">
          <div>
            <b-button 
              variant="outline-primary" 
              size="sm"
              @click="playAudio(data.item.data)"
              :disabled="isPlaying && currentPlayingUrl === data.item.data"
              class="mb-1"
            >
              <i class="fas" :class="isPlaying && currentPlayingUrl === data.item.data ? 'fa-pause' : 'fa-play'"></i>
              {{ isPlaying && currentPlayingUrl === data.item.data ? 'Playing' : 'Play' }}
            </b-button>
            
            <div v-if="currentPlayingUrl === data.item.data" class="mt-1" style="width: 200px">
              <b-progress 
                :value="audioProgress" 
                :max="100" 
                height="4px"
                class="mb-1"
              ></b-progress>
              <small class="text-muted">
                {{ formatTime(currentTime) }} / {{ formatTime(duration) }}
              </small>
              <div :id="'waveform-' + currentPlayingIndex"></div>
            </div>
          </div>
        </template>
      </b-table>
    </div>
    <b-pagination
      v-model="recording_page"
      :total-rows="totalRows"
      :per-page="recording_page_size"
      aria-controls="recording_table"
    ></b-pagination>
  </div>
</template>

<script>
import { convertToCurrentTimeZone } from "./common.js";
import { get_recordings,api_file_request } from "./utils.js";
export default {
  name: "DataRecording",
  props: {
    id: {
      type: String,
      required: true,
    },
    useMp3: {
      type: Boolean,
      default: false
    },
   
  },
  data() {
    return {
      info: [],
      totalRows: 0,
      recording_page: 1,
      recording_page_size: 20,
      dateRange: {
        startDate: null,
        endDate: null
      },
      recording_fields: [
        {
          key: "date",
          label: "日期",
        },
        {
          key: "data",
          label: "数据",
        }
      ],
      audio: null,
      isPlaying: false,
      currentPlayingUrl: null,
      audioProgress: 0,
      currentTime: 0,
      duration: 0,
      currentPlayingIndex: -1,
      audioContext: null,
      useMp3Local: false
    };
  },
  computed: {
    audioType() {
      return this.useMp3 ? 'audio/mp3' : 'audio/mpeg'
    }
  },
  watch: {
    recording_page() {
      this.fetchData();
    },
    useMp3Local(newVal) {
      this.$emit('update:useMp3', newVal);
      this.fetchData();
    },
    useMp3: {
      immediate: true,
      handler(newVal) {
        this.useMp3Local = newVal;
      }
    }
  },
  mounted() {
    this.fetchData();
  },
  methods: {
    // Audio playback control methods
    playAudio(url) {
      const currentIndex = this.info.findIndex(item => item.data === url);
      this.currentPlayingIndex = currentIndex;

      // If clicking the currently playing audio, pause it
      if (this.audio && this.currentPlayingUrl === url) {
        this.pauseAudio();
        return;
      }

      // Clean up existing audio if any
      this.cleanupExistingAudio();

      // Initialize and play new audio
      this.initializeAudio(url);
    },

    pauseAudio() {
      if (this.audio) {
        this.audio.pause();
        this.resetPlaybackState();
      }
    },

    cleanupExistingAudio() {
      if (this.audio) {
        this.audio.pause();
        
        // Remove specific timeupdate handler if it exists
        if (this.audio._timeUpdateHandler) {
          this.audio.removeEventListener('timeupdate', this.audio._timeUpdateHandler);
          delete this.audio._timeUpdateHandler;
        }
        
        // Remove waveform container
        const containerId = `waveform-${this.currentPlayingIndex}`;
        const container = document.getElementById(containerId);
        if (container) {
          while (container.firstChild) {
            container.removeChild(container.firstChild);
          }
        }
      }
    },

    async initializeAudio(url) {
      try {
        // Resume AudioContext if it was suspended
        if (this.audioContext && this.audioContext.state === 'suspended') {
          await this.audioContext.resume();
        }

        // Initialize Web Audio API context
        if (!this.audioContext) {
          this.audioContext = new (window.AudioContext || window.webkitAudioContext)();
        }

        // Fetch the audio data - modified for base64
        const base64Data = url.split(',')[1];
        const binaryString = atob(base64Data);
        const bytes = new Uint8Array(binaryString.length);
        for (let i = 0; i < binaryString.length; i++) {
          bytes[i] = binaryString.charCodeAt(i);
        }
        
        // Decode the audio data
        const audioBuffer = await this.audioContext.decodeAudioData(bytes.buffer);
        
        // Create audio element for playback
        this.audio = new Audio(url);  // 直接使用 data URL
        this.currentPlayingUrl = url;
        
        // Setup event listeners
        this.setupAudioEventListeners();
        
        // Wait for next tick to ensure container is rendered
        await this.$nextTick();
        
        // Draw static waveform
        const containerId = `waveform-${this.currentPlayingIndex}`;
        await this.drawStaticWaveform(audioBuffer, containerId);
        
        // Start playback
        await this.audio.play();
        this.isPlaying = true;
      } catch (error) {
        console.error('Failed to initialize audio:', error);
        this.resetPlaybackState();
      }
    },

    async drawStaticWaveform(audioBuffer, containerId) {
      const container = document.getElementById(containerId);
      if (!container) {
        console.error('Container not found:', containerId);
        return;
      }

      // Create wrapper div for canvas and progress indicator
      const wrapper = document.createElement('div');
      wrapper.style.position = 'relative';
      wrapper.style.width = '200px';
      wrapper.style.height = '60px';
      wrapper.style.cursor = 'pointer';

      // Create canvas
      const canvas = document.createElement('canvas');
      canvas.width = 200;
      canvas.height = 60;
      canvas.style.display = 'block';
      canvas.style.marginTop = '10px';
      canvas.style.backgroundColor = '#f8f9fa';

      // Create progress indicator
      const progressBar = document.createElement('div');
      progressBar.style.position = 'absolute';
      progressBar.style.top = '0';
      progressBar.style.left = '0';
      progressBar.style.width = '2px';
      progressBar.style.height = '100%';
      progressBar.style.backgroundColor = 'red';
      progressBar.style.pointerEvents = 'none';

      // Add elements to wrapper
      wrapper.appendChild(canvas);
      wrapper.appendChild(progressBar);

      const ctx = canvas.getContext('2d');

      // Draw static waveform once
      const channelData = audioBuffer.getChannelData(0);
      const step = Math.ceil(channelData.length / canvas.width);
      
      ctx.fillStyle = '#f8f9fa';
      ctx.fillRect(0, 0, canvas.width, canvas.height);
      
      ctx.beginPath();
      ctx.lineWidth = 2;
      ctx.strokeStyle = '#007bff';
      
      for (let i = 0; i < canvas.width; i++) {
        const startIndex = i * step;
        const endIndex = startIndex + step > channelData.length ? channelData.length : startIndex + step;
        
        let min = 1.0;
        let max = -1.0;
        
        for (let j = startIndex; j < endIndex; j++) {
          const value = channelData[j];
          if (value < min) min = value;
          if (value > max) max = value;
        }
        
        const x = i;
        const yMin = ((1 + min) * canvas.height) / 2;
        const yMax = ((1 + max) * canvas.height) / 2;
        
        ctx.moveTo(x, yMin);
        ctx.lineTo(x, yMax);
      }
      
      ctx.stroke();

      // Clear container
      while (container.firstChild) {
        container.removeChild(container.firstChild);
      }
      
      // Add wrapper to container
      container.appendChild(wrapper);

      // Add event listeners for seeking
      let isDragging = false;

      const updateProgress = (e) => {
        const rect = wrapper.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const progress = Math.max(0, Math.min(1, x / rect.width));
        progressBar.style.left = `${progress * 100}%`;
        
        if (this.audio) {
          this.audio.currentTime = this.audio.duration * progress;
        }
      };

      wrapper.addEventListener('mousedown', (e) => {
        isDragging = true;
        updateProgress(e);
      });

      document.addEventListener('mousemove', (e) => {
        if (isDragging) {
          updateProgress(e);
        }
      });

      document.addEventListener('mouseup', () => {
        isDragging = false;
      });

      // Update progress bar position during playback
      const timeUpdateHandler = () => {
        if (!isDragging && this.audio && this.audio.duration) {
          const progress = this.audio.currentTime / this.audio.duration;
          progressBar.style.left = `${progress * 100}%`;
        }
      };

      this.audio.addEventListener('timeupdate', timeUpdateHandler);

      // Store the event handler reference for cleanup
      this.audio._timeUpdateHandler = timeUpdateHandler;
    },

    setupAudioEventListeners() {
      // Duration metadata loaded
      this.audio.addEventListener('loadedmetadata', () => {
        this.duration = this.audio.duration;
      });

      // Progress update
      this.audio.addEventListener('timeupdate', this.updateProgress);
      
      // Playback ended
      this.audio.addEventListener('ended', this.handlePlaybackEnded);

      // Error handling
      this.audio.addEventListener('error', this.handlePlaybackError);
    },

    handlePlaybackEnded() {
      this.resetPlaybackState();
      
      if (this.currentPlayingIndex > 0) {
        // Play previous track in current page
        const previousUrl = this.info[this.currentPlayingIndex - 1].data;
        this.playAudio(previousUrl);
      } else if (this.recording_page > 1) {
        // If at top of page and not first page, go to previous page
        this.recording_page -= 1;
        this.$nextTick(() => {
          // Wait for data to load before playing last track
          setTimeout(() => {
            if (this.info.length > 0) {
              const lastUrl = this.info[this.info.length - 1].data;
              this.playAudio(lastUrl);
            }
          }, 500);
        });
      } else {
        this.currentPlayingIndex = -1;
      }
    },

    handlePlaybackError(e) {
      console.error('Audio playback error:', e);
      this.resetPlaybackState();
    },

    resetPlaybackState() {
      this.isPlaying = false;
      this.currentPlayingUrl = null;
      this.audioProgress = 0;
      this.currentTime = 0;
    },

    updateProgress() {
      if (this.audio) {
        this.currentTime = this.audio.currentTime;
        this.audioProgress = (this.audio.currentTime / this.audio.duration) * 100;
      }
    },

    // Audio format detection
    checkAudioType(base64Data) {
      const binaryString = atob(base64Data.substring(0, 8));
      const bytes = new Uint8Array(binaryString.length);
      for (let i = 0; i < binaryString.length; i++) {
        bytes[i] = binaryString.charCodeAt(i);
      }
      
      const isMP3 = (
        (bytes[0] === 0x49 && bytes[1] === 0x44 && bytes[2] === 0x33) || // "ID3" header
        (bytes[0] === 0xFF && (bytes[1] & 0xFB) === 0xFB)                // MPEG sync frame
      );
      
      return isMP3 ? 'audio/mp3' : 'audio/mpeg';
    },

    // Time formatting helper
    formatTime(seconds) {
      if (!seconds) return '00:00';
      const mins = Math.floor(seconds / 60);
      const secs = Math.floor(seconds % 60);
      return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
    },

    async fetchData() {
      try {
        const response = await get_recordings({
          id: this.id,
          page: this.recording_page,
          pageSize: this.recording_page_size,
          startDate: this.dateRange.startDate,
          endDate: this.dateRange.endDate,
        });
        this.info = response.recordings.map(item => {
          // 直接使用 base64 数据
          const audioData = `data:${this.useMp3 ? 'audio/mp3' : 'audio/mpeg'};base64,${item.recording}`;
          return {
            data: audioData,
            date: convertToCurrentTimeZone(new Date(item.createdAt))
          };
        });
        this.totalRows = response.pagination.total;
      } catch (error) {
        console.error('Error fetching recordings:', error);
      }
    },
    async downloadData() {
      await api_file_request("GET", "/recordings", null, {
        id: this.id,
        download: true,
        startDate: this.dateRange.startDate,
        endDate: this.dateRange.endDate
      })
    },
  },
  beforeDestroy() {
    this.cleanupExistingAudio();
    this.audio = null;
    
    // Cleanup audio context
    if (this.audioContext) {
      this.audioContext.close();
      this.audioContext = null;
    }
  },
};
</script>

<style scoped>
/* Add component-specific styles here */
</style>
