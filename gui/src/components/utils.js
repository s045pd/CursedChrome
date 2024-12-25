const BASE_API_PATH = `${location.origin.toString()}/api/v1`;

// make a request to the api for a file
export const api_file_request = async (method, path, body, params) => {
  var request_options = {
    method: method,
    credentials: "include",
    params: params,
    mode: "cors",
    cache: "no-cache",
    responseType: "blob",
    headers: {
      "Content-Type": "application/json",
    },
    redirect: "follow",
  };

  if (body) {
    request_options.body = JSON.stringify(body);
  }

  window.app.loading = true;

  try {
    const res = await fetch(`${BASE_API_PATH}${path}`, request_options);
    const blob = await res.blob();
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.setAttribute("download", "record.mp3");
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  } catch (e) {
    window.app.loading = false;
    throw e;
  }
  window.app.loading = false;
};

// make a request to the api
export const api_request = async (method, path, params, body) => {
  var request_options = {
    method: method,
    credentials: "include",
    mode: "cors",
    cache: "no-cache",
    headers: {
      "Content-Type": "application/json",
    },
    redirect: "follow",
  };
  if (params && Object.keys(params).length > 0) {
    path += "?";
    path += Object.keys(params)
      .map((key) => `${key}=${params[key]}`)
      .join("&");
  }
  if (body && Object.keys(body).length > 0) {
    request_options.body = JSON.stringify(body);
  }

  window.app.loading = true;

  try {
    var response = await fetch(`${BASE_API_PATH}${path}`, request_options);
  } catch (e) {
    window.app.loading = false;
    throw e;
  }
  window.app.loading = false;

  const response_body = await response.json();

  if (!response_body.success) {
    return Promise.reject({
      error: response_body.error,
      code: response_body.code,
    });
  }

  return response_body.result;
};

// get the list of all the fields
export const get_field = async (id, field) => {
  return await api_request("GET", "/fields", { id: id, field: field });
};

export const get_recordings = async ({
  id,
  page,
  pageSize,
  startDate,
  endDate,
} = {}) => {
  return await api_request("GET", "/recordings", {
    id,
    page,
    pageSize,
    startDate,
    endDate,
  });
};

// download the certificate authority
export const download_ca = async () => {
  window.location = `${BASE_API_PATH}/download_ca`;
};
