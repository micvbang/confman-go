import axios from "axios";
import configs from "@/configs.js";

const servicePathConfigs = configs.BACKEND_API_BASE_URL + "/service_paths";

export class ConfmanClient {
  constructor() { }

  async getServicePathConfigs() {
    console.log(servicePathConfigs)
    const result = await axios.get(`${servicePathConfigs}?recursive=true`);
    if (result.status !== 200) {
      console.log("failed to retrieve service paths", result);
      return [];
    }

    return result.data;
  }
}
