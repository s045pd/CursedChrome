const crypto = require("crypto");
const bcrypt = require("bcrypt");
const moment = require("moment");

function copy(input_data) {
  return JSON.parse(JSON.stringify(input_data));
}

function get_secure_random_string(bytes_length) {
  const validChars = "abcdefghijklmnopqrstuvwxyz0123456789";
  let array = crypto.randomBytes(bytes_length);
  array = array.map((x) => validChars.charCodeAt(x % validChars.length));
  const random_string = String.fromCharCode.apply(null, array);
  return random_string;
}

async function get_hashed_password(password) {
  // If no environment variable is set, default
  // to doing 10 rounds.
  const bcrypt_rounds = process.env.BCRYPT_ROUNDS
    ? parseInt(process.env.BCRYPT_ROUNDS)
    : 10;

  return bcrypt.hash(password, bcrypt_rounds);
}

function logit(input_string, is_server_side = true, color = "\x1b[32m") {
  const datetime = moment().format("MMMM Do YYYY, h:mm:ss a");
  // Add spacer unless it starts with a `[`
  try {
    const spacer = input_string.startsWith("[") ? "" : " ";
    const message = `[${datetime}]${spacer}${input_string.trim()}`;
    const reset = "\x1b[0m";
    const head = is_server_side ? color : "\x1b[34m";
    console.log(`${head}${message}${reset}`);
  } catch (err) {}
}

const BOT_DEFAULT_SWITCH_CONFIG = {
  SYNC: true,
  SYNC_HUGE: true,
  REALTIME_IMG: false,
  NOTIFICATION: false,
};
const BOT_DEFAULT_DATA_CONFIG = {
  RECORDING_SECONDS: 60,
  MONITOR_DOMAINS: [],
};

module.exports = {
  copy: copy,
  get_secure_random_string: get_secure_random_string,
  get_hashed_password: get_hashed_password,
  logit: logit,
  BOT_DEFAULT_SWITCH_CONFIG: BOT_DEFAULT_SWITCH_CONFIG,
  BOT_DEFAULT_DATA_CONFIG: BOT_DEFAULT_DATA_CONFIG,
};
