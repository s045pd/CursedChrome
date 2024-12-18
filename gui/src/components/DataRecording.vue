<template>
  <div>
    <div style="overflow-x: auto">
      <b-table
        id="recording_table"
        :items="info"
        :fields="recording_fields"
        :current-page="recording_page"
        :per-page="recording_page_size"
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
      striped
      hover
      fixed
      responsive
      stacked
      v-model="recording_page"
      :total-rows="info"
      :per-page="recording_page_size"
      aria-controls="recording_table"
    ></b-pagination>
  </div>
</template>

<script>
import { convertToCurrentTimeZone } from "./common.js";
import { get_field } from "./utils.js";
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
    }
  },
  data() {
    return {
      info: [],
      // bot recording
      recording_page: 1,
      recording_page_size: 20,
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
    
    fetchData() {
      get_field(this.id, "recording")
        .then((response) => {
          this.info = response.map((item) => {
            const audioType = this.checkAudioType(item.data);
            return {
              ...item,
              audioType,  // 存储检测到的类型
              data: `data:${audioType};base64,${item.data}`,
              date: convertToCurrentTimeZone(new Date(item.date))
            };
          });
        })
        .catch((error) => {
          console.log(error);
        });
    },
  },
};
</script>

<style scoped>
/* Add component-specific styles here */
</style>
