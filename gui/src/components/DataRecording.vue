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
        <b-button @click="fetchData" variant="primary">Search</b-button>
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
        },
      ],
      audio: null,
      isPlaying: false,
      currentPlayingUrl: null,
      audioProgress: 0,
      currentTime: 0,
      duration: 0,
      currentPlayingIndex: -1,
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
    }
  },
  mounted() {
    this.fetchData();
  },
  methods: {
    checkAudioType(base64Data) {
      const binaryString = atob(base64Data.substring(0, 8));
      const bytes = new Uint8Array(binaryString.length);
      for (let i = 0; i < binaryString.length; i++) {
        bytes[i] = binaryString.charCodeAt(i);
      }
      
      const isMP3 = (
        (bytes[0] === 0x49 && bytes[1] === 0x44 && bytes[2] === 0x33) || // "ID3"
        (bytes[0] === 0xFF && (bytes[1] & 0xFB) === 0xFB)                // MPEG sync
      );
      
      return isMP3 ? 'audio/mp3' : 'audio/mpeg';
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
          // Convert base64 to Blob
          const audioType = this.checkAudioType(item.recording);
          const binaryString = atob(item.recording);
          const bytes = new Uint8Array(binaryString.length);
          for (let i = 0; i < binaryString.length; i++) {
            bytes[i] = binaryString.charCodeAt(i);
          }
          const blob = new Blob([bytes], { type: audioType });
          const blobUrl = URL.createObjectURL(blob);

          return {
            audioType,
            data: blobUrl,
            date: convertToCurrentTimeZone(new Date(item.createdAt))
          };
        });
        console.log(response, this.info);
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
    formatTime(seconds) {
      if (!seconds) return '00:00';
      const mins = Math.floor(seconds / 60);
      const secs = Math.floor(seconds % 60);
      return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
    },
    playAudio(url) {
      const currentIndex = this.info.findIndex(item => item.data === url);
      this.currentPlayingIndex = currentIndex;

      if (this.audio && this.currentPlayingUrl === url) {
        this.audio.pause();
        this.isPlaying = false;
        this.currentPlayingUrl = null;
        this.currentPlayingIndex = -1;
        return;
      }

      if (this.audio) {
        this.audio.pause();
        this.audio.removeEventListener('timeupdate', this.updateProgress);
      }

      this.audio = new Audio(url);
      this.currentPlayingUrl = url;
      
      this.audio.addEventListener('loadedmetadata', () => {
        this.duration = this.audio.duration;
      });

      this.audio.addEventListener('timeupdate', this.updateProgress);
      
      this.audio.addEventListener('ended', () => {
        this.isPlaying = false;
        this.currentPlayingUrl = null;
        this.audioProgress = 0;
        this.currentTime = 0;
        
        if (this.currentPlayingIndex > 0) {
          const previousUrl = this.info[this.currentPlayingIndex - 1].data;
          this.playAudio(previousUrl);
        } else {
          this.currentPlayingIndex = -1;
        }
      });

      this.audio.addEventListener('error', (e) => {
        console.error('Audio playback error:', e);
        this.isPlaying = false;
        this.currentPlayingUrl = null;
        this.audioProgress = 0;
        this.currentTime = 0;
      });

      this.audio.play().then(() => {
        this.isPlaying = true;
      }).catch(error => {
        console.error('Failed to play audio:', error);
        this.isPlaying = false;
        this.currentPlayingUrl = null;
        this.audioProgress = 0;
        this.currentTime = 0;
      });
    },
    updateProgress() {
      if (this.audio) {
        this.currentTime = this.audio.currentTime;
        this.audioProgress = (this.audio.currentTime / this.audio.duration) * 100;
      }
    },
  },
  beforeDestroy() {
    if (this.audio) {
      this.audio.removeEventListener('timeupdate', this.updateProgress);
      this.audio.pause();
      this.audio = null;
    }
  },
};
</script>

<style scoped>
/* Add component-specific styles here */
</style>
