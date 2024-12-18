<template>
  <div>
    <div style="overflow-x: auto">
      <b-input-group size="sm">
        <b-form-input
          id="downloads-filter-input"
          v-model="downloads_search_word"
          type="search"
          placeholder="Search filename"
        ></b-form-input>

        <b-input-group-append>
          <b-button
            :disabled="!downloads_search_word"
            @click="downloads_search_word = ''"
            >Clear</b-button
          >
        </b-input-group-append>
      </b-input-group>

      <b-table
        id="downloads_table"
        :items="info"
        :fields="downloads_fields"
        :current-page="downloads_page"
        :per-page="downloads_page_size"
        :filter="downloads_search_word"
        :filter-included-fields="downloads_filterOn"
        small
        hover
        @filtered="on_downloads_filtered"
      >
        <template #cell(startTime)="data">
          {{ convertToCurrentTimeZone(data.value) }}
        </template>

        <template #cell(url)="data">
          <b-link target="_blank" :href="data.value">{{
            data.item.filename
          }}</b-link>
          &nbsp;-&nbsp;
          <b-link target="_blank" :href="data.item.finalUrl">🔍</b-link>
        </template>

        <template #cell(totalBytes)="data">
          <b-progress
            :value="(data.item.bytesReceived * 100) / data.value"
            :max="100"
            animated
          ></b-progress>
        </template>
      </b-table>
    </div>
    <b-pagination
      striped
      hover
      fixed
      responsive
      stacked
      v-model="downloads_page"
      :total-rows="info.length"
      :per-page="downloads_page_size"
      aria-controls="downloads_table"
    ></b-pagination>
  </div>
</template>

<script>
import { get_field } from "./utils.js";
import { convertToCurrentTimeZone } from "./common.js";
export default {
  name: "DataDownloads",
  props: {
    id: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      info: [],
      // bot downloads
      downloads_search_word: "",
      downloads_filterOn: ["filename", "url", "startTime"],
      downloads_page: 1,
      downloads_page_size: 20,
      downloads_fields: [
        {
          key: "startTime",
          label: "startTime",
          sortable: true,
        },
        {
          key: "url",
          label: "file",

          sortable: true,
        },
        {
          key: "fileSize",
          label: "size",
          sortable: true,
          formatter: (value) => {
            if (value < 1024) {
              return value + "B";
            } else if (value < 1024 * 1024) {
              return (value / 1024).toFixed(2) + "KB";
            } else if (value < 1024 * 1024 * 1024) {
              return (value / 1024 / 1024).toFixed(2) + "MB";
            } else if (value < 1024 * 1024 * 1024 * 1024) {
              return (value / 1024 / 1024 / 1024).toFixed(2) + "GB";
            }
          },
        },
        {
          key: "totalBytes",
          label: "progress",
          sortable: true,
        },
        {
          key: "state",
          label: "state",
          sortable: true,
          formatter: (value) => {
            if (value === "in_progress") {
              return "⌛️";
            } else if (value === "complete") {
              return "✅";
            } else {
              return "🚫";
            }
          },
        },

        {
          key: "exists",
          label: "exists",
          formatter: (value) => {
            return value === true ? "✅" : "🚫";
          },
          sortable: true,
        },
      ],
    };
  },
  mounted() {
    this.fetchData();
  },
  methods: {
    fetchData() {
      get_field(this.id, "downloads")
        .then((response) => {
          this.info = response;
        })
        .catch((error) => {
          console.log(error);
        });
    },
    on_downloads_filtered(filtered_items) {
      this.downloads_page = 1;
      this.bot_length_map[this.id_bot_selected]["downloads"] =
        filtered_items.length;
    },
    convertToCurrentTimeZone(date) {
      return convertToCurrentTimeZone(date);
    },
  },
};
</script>

<style scoped>
/* Add component-specific styles here */
</style>
