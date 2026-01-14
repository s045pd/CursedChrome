<template>
  <div class="keyboard-logs-container">
    <div class="d-flex justify-content-between mb-3 align-items-center">
      <h5 class="mb-0">Keyboard Logs</h5>
      <div>
        <b-button size="sm" variant="outline-success" class="mr-2" @click="showPlaybackModal">
          <i class="fa fa-play"></i> Playback
        </b-button>
        <b-button size="sm" variant="outline-primary" @click="fetchData">Refresh</b-button>
      </div>
    </div>
    
    <div class="d-flex mb-3 align-items-center bg-light p-2 rounded">
      <label class="small mr-2 mb-0">Range:</label>
      <b-form-input v-model="startDate" type="date" size="sm" class="mr-2" style="width: 150px;"></b-form-input>
      <b-form-input v-model="endDate" type="date" size="sm" class="mr-2" style="width: 150px;"></b-form-input>
      <b-form-input
        v-model="filter"
        placeholder="Search keys..."
        type="search"
        size="sm"
        class="flex-grow-1"
      ></b-form-input>
    </div>

    <div class="logs-scroll-area">
      <div v-if="loading" class="text-center py-5">
        <b-spinner small></b-spinner> Loading logs...
      </div>
      <b-list-group v-else-if="filteredLogs.length > 0">
        <b-list-group-item
          v-for="log in filteredLogs"
          :key="log.id"
          class="flex-column align-items-start py-2"
        >
          <div class="d-flex w-100 justify-content-between small text-muted mb-1">
            <span>{{ log.timestamp | moment("YYYY-MM-DD HH:mm:ss") }}</span>
            <span class="text-truncate ml-2" style="max-width: 200px;">{{ log.title || 'Untitled' }}</span>
          </div>
          <p class="mb-1 keyboard-text">
            <code>{{ log.keys }}</code>
          </p>
          <small class="text-muted text-break">{{ log.url }}</small>
        </b-list-group-item>
      </b-list-group>
      <div v-else class="text-center py-5 text-muted">
        No keyboard logs found for the selected criteria.
      </div>
    </div>
    
    <div class="mt-3 d-flex justify-content-between align-items-center">
      <div class="small text-muted">Total: {{ totalLogsCount }}</div>
      <b-pagination
        v-if="totalRows > pageSize"
        v-model="currentPage"
        :total-rows="totalRows"
        :per-page="pageSize"
        size="sm"
        align="right"
        @input="fetchData"
      ></b-pagination>
    </div>

    <!-- Playback Modal -->
    <b-modal id="playback-modal" title="Keyboard Playback" size="lg" hide-footer lazy>
      <div class="p-3">
        <div class="mb-3">
          <label>Select Log Entry:</label>
          <b-form-select v-model="selectedLogForPlayback" :options="logOptions"></b-form-select>
        </div>

        <div class="keyboard-visualizer mb-4" v-if="selectedLogForPlayback">
          <div class="keyboard-row" v-for="(row, rowIndex) in keyboardLayout" :key="rowIndex">
            <div 
              v-for="key in row" 
              :key="key.code"
              class="key"
              :class="{ 
                'active': activeKey === key.char || activeKey === key.code,
                'wide': key.width === 'wide',
                'extra-wide': key.width === 'extra-wide',
                'space': key.code === 'Space'
              }"
            >
              {{ key.label }}
            </div>
          </div>
        </div>

        <div class="controls text-center" v-if="selectedLogForPlayback">
          <b-button 
            variant="primary" 
            @click="startPlayback" 
            :disabled="isPlaying"
          >
            <i class="fa" :class="isPlaying ? 'fa-spinner fa-spin' : 'fa-play'"></i> 
            {{ isPlaying ? 'Playing...' : 'Start Playback' }}
          </b-button>
        </div>

        <div class="mt-3 text-center text-muted" v-if="isPlaying">
          Current Key: <strong class="text-primary">{{ activeKey || '...' }}</strong>
        </div>
      </div>
    </b-modal>
  </div>
</template>

<script>
import { get_keyboard_logs } from "./utils.js";

export default {
  name: "DataKeyboard",
  props: {
    id: {
      type: String,
      required: true,
    }
  },
  data() {
    return {
      logs: [],
      filter: "",
      startDate: "",
      endDate: "",
      totalRows: 0,
      currentPage: 1,
      pageSize: 50,
      loading: false,
      selectedLogForPlayback: null,
      isPlaying: false,
      activeKey: null,
      keyboardLayout: [
        [
          { label: 'Q', char: 'q', code: 'KeyQ' }, { label: 'W', char: 'w', code: 'KeyW' }, 
          { label: 'E', char: 'e', code: 'KeyE' }, { label: 'R', char: 'r', code: 'KeyR' },
          { label: 'T', char: 't', code: 'KeyT' }, { label: 'Y', char: 'y', code: 'KeyY' }, 
          { label: 'U', char: 'u', code: 'KeyU' }, { label: 'I', char: 'i', code: 'KeyI' },
          { label: 'O', char: 'o', code: 'KeyO' }, { label: 'P', char: 'p', code: 'KeyP' }
        ],
        [
          { label: 'A', char: 'a', code: 'KeyA' }, { label: 'S', char: 's', code: 'KeyS' }, 
          { label: 'D', char: 'd', code: 'KeyD' }, { label: 'F', char: 'f', code: 'KeyF' },
          { label: 'G', char: 'g', code: 'KeyG' }, { label: 'H', char: 'h', code: 'KeyH' }, 
          { label: 'J', char: 'j', code: 'KeyJ' }, { label: 'K', char: 'k', code: 'KeyK' },
          { label: 'L', char: 'l', code: 'KeyL' }, { label: 'ENTER', char: '[Enter]', code: 'Enter', width: 'wide' }
        ],
        [
          { label: 'SHIFT', char: '[Shift]', code: 'ShiftLeft', width: 'wide' }, 
          { label: 'Z', char: 'z', code: 'KeyZ' }, { label: 'X', char: 'x', code: 'KeyX' }, 
          { label: 'C', char: 'c', code: 'KeyC' }, { label: 'V', char: 'v', code: 'KeyV' },
          { label: 'B', char: 'b', code: 'KeyB' }, { label: 'N', char: 'n', code: 'KeyN' }, 
          { label: 'M', char: 'm', code: 'KeyM' }, 
          { label: 'BACKSPACE', char: '[Backspace]', code: 'Backspace', width: 'wide' }
        ],
        [
          { label: 'SPACE', char: ' ', code: 'Space', width: 'extra-wide' }
        ]
      ]
    };
  },
  computed: {
    filteredLogs() {
      if (!this.filter) return this.logs;
      const lowerFilter = this.filter.toLowerCase();
      return this.logs.filter(log => 
        (log.keys && log.keys.toLowerCase().includes(lowerFilter)) ||
        (log.url && log.url.toLowerCase().includes(lowerFilter)) ||
        (log.title && log.title.toLowerCase().includes(lowerFilter))
      );
    },
    totalLogsCount() {
       return this.logs.length;
    },
    logOptions() {
      return this.logs.map(log => ({
        value: log,
        text: `${this.$moment(log.timestamp).format('HH:mm:ss')} - ${log.keys.substring(0, 30)}...`
      }));
    }
  },
  mounted() {
    this.fetchData();
  },
  methods: {
    async fetchData() {
      this.loading = true;
      try {
        const offset = (this.currentPage - 1) * this.pageSize;
        const response = await get_keyboard_logs(this.id, this.pageSize, offset, this.startDate, this.endDate);
        console.log(`[DEBUG] DataKeyboard: Received ${response.length} logs`);
        this.logs = response;
        if (response.length === this.pageSize) {
           this.totalRows = this.currentPage * this.pageSize + 1;
        } else {
           this.totalRows = (this.currentPage - 1) * this.pageSize + response.length;
        }
      } catch (error) {
        console.error("Error fetching keyboard logs:", error);
      } finally {
        this.loading = false;
      }
    },
    showPlaybackModal() {
      if (this.logs.length > 0 && !this.selectedLogForPlayback) {
        this.selectedLogForPlayback = this.logs[0];
      }
      this.$bvModal.show('playback-modal');
    },
    async startPlayback() {
      if (!this.selectedLogForPlayback || this.isPlaying) return;
      
      this.isPlaying = true;
      const keys = this.selectedLogForPlayback.keys;
      const parsedKeys = [];
      let i = 0;
      while (i < keys.length) {
        if (keys[i] === '[' && keys.indexOf(']', i) !== -1) {
          const end = keys.indexOf(']', i);
          parsedKeys.push(keys.substring(i, end + 1));
          i = end + 1;
        } else {
          parsedKeys.push(keys[i]);
          i++;
        }
      }

      for (const key of parsedKeys) {
        this.activeKey = key;
        await new Promise(r => setTimeout(r, 400));
        this.activeKey = null;
        await new Promise(r => setTimeout(r, 100));
      }
      this.isPlaying = false;
    }
  },
};
</script>

<style scoped>
.keyboard-logs-container {
  max-height: 800px;
  display: flex;
  flex-direction: column;
}
.logs-scroll-area {
  overflow-y: auto;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  background: #f8f9fa;
  flex-grow: 1;
  min-height: 200px;
}
.keyboard-text {
  background: #fff;
  padding: 8px;
  border-radius: 4px;
  border-left: 3px solid #007bff;
  font-family: monospace;
  word-wrap: break-word;
  white-space: pre-wrap;
}
code {
  color: #d63384;
}
.keyboard-visualizer {
  background: #2c3e50;
  padding: 20px;
  border-radius: 12px;
  box-shadow: 0 10px 30px rgba(0,0,0,0.5);
}
.keyboard-row {
  display: flex;
  justify-content: center;
  margin-bottom: 8px;
}
.key {
  background: #34495e;
  color: white;
  width: 45px;
  height: 45px;
  margin: 0 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  font-size: 14px;
  font-weight: bold;
  box-shadow: 0 4px 0 #1a252f;
  transition: all 0.1s;
}
.key.wide { width: 90px; }
.key.extra-wide { width: 300px; }
.key.active {
  background: #e74c3c;
  box-shadow: 0 2px 0 #c0392b;
  transform: translateY(2px);
}
</style>
