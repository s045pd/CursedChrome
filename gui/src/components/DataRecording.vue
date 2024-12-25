<template>
  <div>
    <b-row class="mb-3">
      <b-col md="6">
        <b-form-group label="日期范围">
          <b-form-datepicker
            v-model="dateRange.startDate"
            :max="dateRange.endDate"
            placeholder="开始日期"
            class="mb-2"
          ></b-form-datepicker>
          <b-form-datepicker
            v-model="dateRange.endDate"
            :min="dateRange.startDate"
            placeholder="结束日期"
          ></b-form-datepicker>
        </b-form-group>
      </b-col>
      <b-col md="6" class="d-flex align-items-end">
        <b-button @click="fetchData" variant="primary">查询</b-button>
        <b-button @click="downloadData" variant="primary">下载</b-button>
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
          <audio controls="controls" autobuffer="autobuffer">
            <source :src="data.item.data" :type="data.item.audioType" />
          </audio>
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
        this.info = response.recordings.map(item => ({
          audioType: this.checkAudioType(item.recording),
          data: `data:${this.checkAudioType(item.recording)};base64,${item.recording}`,
          date: convertToCurrentTimeZone(new Date(item.createdAt))
        }));
        console.log(response,this.info);
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
    }
  },
};
</script>

<style scoped>
/* Add component-specific styles here */
</style>
