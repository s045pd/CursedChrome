<template>
  <div>
    <div style="overflow-x: auto">
      <b-input-group size="sm">
        <b-form-input
          id="cookies-filter-input"
          v-model="cookies_search_word"
          type="search"
          placeholder="Search domain or name or value"
        ></b-form-input>

        <b-input-group-append>
          <b-button
            :disabled="!cookies_search_word"
            @click="cookies_search_word = ''"
            variant="outline-secondary"
            >Clear</b-button
          >
          <b-button
            variant="primary"
            @click="copyToClipboard"
            title="Copy all visible cookies to clipboard"
          >
            <i class="fa fa-copy"></i> Copy
          </b-button>
        </b-input-group-append>
      </b-input-group>

      <div class="my-2 text-muted small">
        Showing {{ filteredInfo.length > 0 || cookies_search_word ? filteredInfo.length : info.length }} cookies
      </div>

      <b-table
        id="cookies_table"
        :items="info"
        :fields="cookies_fields"
        :current-page="cookies_page"
        :per-page="cookies_page_size"
        :filter="cookies_search_word"
        :filter-included-fields="cookies_filterOn"
        small
        hover
        :tbody-tr-class="cookies_row_class"
        @filtered="on_cookies_filtered"
        responsive="sm"
        thead-class="bg-light"
      >
        <template #cell(value)="data">
          <div class="text-truncate" style="max-width: 200px;" :title="data.value">
            {{ data.value }}
          </div>
        </template>
        <template #cell(domain)="data">
          <span class="font-weight-bold">{{ data.value }}</span>
        </template>
      </b-table>
    </div>
    <b-pagination
      striped
      hover
      fixed
      responsive
      stacked
      v-model="cookies_page"
      :total-rows="info.length"
      :per-page="cookies_page_size"
      aria-controls="cookies_table"
    ></b-pagination>
  </div>
</template>

<script>
import { copy_toast } from "./common.js";
import { get_field } from "./utils.js";
export default {
  name: "DataCookies",
  props: {
    id: {
      type: String,
      required: true,
    },
    // enableClipboard: {
    //   type: Boolean,
    //   default: false,
    // },
  },
  data() {
    return {
      info: [],
      filteredInfo: [],
      // boot cookies
      cookies_search_word: "",
      cookies_filterOn: ["domain", "name", "value"],
      cookies_page: 1,
      cookies_page_size: 20,
      cookies_fields: [
        { key: "domain", label: "domain", sortable: true },
        { key: "name", label: "name", sortable: true },
        {
          key: "hostOnly",
          label: "hostOnly",
          sortable: true,
          formatter: (value) => {
            return value === true ? "✅" : "🚫";
          },
        },
        {
          key: "httpOnly",
          label: "httpOnly",
          sortable: true,
          formatter: (value) => {
            return value === true ? "✅" : "🚫";
          },
        },

        { key: "path", label: "path", sortable: true },
        { key: "sameSite", label: "sameSite", sortable: true },
        {
          key: "secure",
          label: "secure",
          sortable: true,
          formatter: (value) => {
            return value === true ? "✅" : "🚫";
          },
        },
        {
          key: "session",
          label: "session",
          sortable: true,
          formatter: (value) => {
            return value === true ? "✅" : "🚫";
          },
        },
        { key: "storeId", label: "storeId", sortable: true },
        { key: "value", label: "value", sortable: true },
      ],
    };
  },
  mounted() {
    this.fetchData();
  },
  methods: {
    async copyToClipboard() {
      const target = (this.cookies_search_word || this.filteredInfo.length > 0) ? this.filteredInfo : this.info;
      const textToCopy = JSON.stringify(target, null, 2);
      
      try {
        await navigator.clipboard.writeText(textToCopy);
        this.copy_toast();
      } catch (err) {
        console.error('Failed to copy text: ', err);
      }
    },
    copy_toast() {
      copy_toast();
    },
    fetchData() {
      get_field(this.id, "cookies")
        .then((response) => {
          this.info = response;
        })
        .catch((error) => {
          console.log(error);
        });
    },
    cookies_row_class(item, type) {
      try {
        if (!item || type !== "row" || !item.expirationDate) {
          return;
        }

        const cookie_expiration_date = new Date(item.expirationDate);
        const current_time = new Date();
        const expiration_warning_time = 1000 * 60 * 60 * 1; // 1 hour

        if (current_time - cookie_expiration_date > expiration_warning_time) {
          return "table-success";
        } else if (current_time - cookie_expiration_date > 0) {
          return "table-warning";
        } else {
          return "table-danger";
        }
      } catch (e) {
        console.log(e);
      }
    },
    on_cookies_filtered(filtered_items) {
      this.cookies_page = 1;
      this.filteredInfo = filtered_items;
    },
  },
};
</script>

<style scoped>
.text-truncate {
  display: inline-block;
  cursor: help;
}
#cookies_table >>> td {
  vertical-align: middle;
  white-space: nowrap;
}
#cookies_table >>> th {
  border-top: none;
}
.bg-light {
  background-color: #f8f9fa !important;
}
</style>
