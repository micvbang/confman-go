import axios from "axios";
import configs from "@/configs.js";

const servicePathConfigs = configs.BACKEND_API_BASE_URL + "/service_paths";
const servicePathConfigKey = configs.BACKEND_API_BASE_URL + "/service_paths/key";
const servicePathConfigKeys = configs.BACKEND_API_BASE_URL + "/service_paths/keys";

export class ConfmanClient {
  constructor() {}

  async getServicePathConfigs() {
    const result = await axios.get(`${servicePathConfigs}?recursive=true`);
    if (result.status !== 200) {
      console.log("failed to retrieve service paths", result);
      return [];
    }

    return result.data;
  }

  async deleteServicePathKeys(servicePath, keys) {
    const data = {
      path: servicePath,
      keys: keys,
    };
    const result = await axios.delete(`${servicePathConfigKeys}`, { data: data });
    if (result.status !== 200) {
      console.log("failed to delete service paths", result);
      return false;
    }

    return true;
  }

  async writeServicePathKey(servicePath, key, value) {
    const data = {
      path: servicePath,
      key: key,
      value: value,
    };
    const result = await axios.put(`${servicePathConfigKey}`, data);
    if (result.status !== 200) {
      console.log("failed to delete service paths", result);
      return false;
    }

    return true;
  }

  async readServicePathKey(servicePath, key) {
    const result = await axios.get(`${servicePathConfigKey}?service-path=${servicePath}&key=${key}`);
    if (result.status !== 200) {
      console.log("failed to delete service paths", result);
      return "";
    }

    return result.data;
  }
}
