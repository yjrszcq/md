import axios, { AxiosRequestConfig } from "axios";
import { ElMessage } from "element-plus";
import { host, context, basicTokenKey } from "@/config";
import Token from "@/store/token";
import sha256 from "crypto-js/sha256";

interface ResponseResult<T> {
  code: number;
  message: string;
  data: T;
}

// 刷新token中
let refreshing = false;

// 数据接口axios实例
const dataInstance = axios.create({
  baseURL: host + context + "/data",
  timeout: 20000,
});

// 授权接口axios实例
const authInstance = axios.create({
  baseURL: host + context + "/token",
  timeout: 10000,
});

// 开放接口axios实例
const openInstance = axios.create({
  baseURL: host + context + "/open",
  timeout: 20000,
});

/**
 * 数据接口request拦截器
 */
dataInstance.interceptors.request.use(
  (config) => {
    // 设置请求头认证信息
    if (!config.headers.Authorization) {
      config.headers.Authorization = "Bearer " + Token.getAccessToken();
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

/**
 * 授权接口request拦截器
 */
authInstance.interceptors.request.use(
  (config) => {
    // 设置请求头认证信息
    if (!config.headers.Authorization) {
      let token = sha256(basicTokenKey + Math.floor(new Date().getTime() / 600000)).toString();
      config.headers.Authorization = "Basic " + token;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

/**
 * 数据接口respone拦截器
 */
dataInstance.interceptors.response.use(
  (response) => {
    // 请求结果码
    if (response.status === 200) {
      // 业务结果码
      if (response.data.code === 200) {
        return response;
      } else if (response.data.code === 401) {
        // 认证失败
        return new Promise((resolve, reject) => {
          if (!refreshing) {
            // 未刷新token时才进行刷新
            refreshing = true;
            authInstance({ method: "post", url: "/refresh", data: { refreshToken: Token.getRefreshToken() } })
              .then((tokenResult) => {
                refreshing = false;
                // 刷新token成功，保存新的token
                Token.setToken(tokenResult.data.data);
                // 重设头信息中的token
                response.config.headers.Authorization = "Bearer " + Token.getAccessToken();
                // 重新进行上一次的请求
                dataInstance(response.config)
                  .then((dataResult) => {
                    resolve(dataResult);
                  })
                  .catch((dataError) => {
                    reject(dataError);
                  });
              })
              .catch((err) => {
                refreshing = false;
                // Refresh token failed, clear token and reload page
                Token.removeToken();
                reject(err);
              });
          } else {
            // Token refresh already in progress, wait and retry the original request
            const interval = setInterval(() => {
              if (!refreshing) {
                clearInterval(interval);
                // Reset Authorization header with new token
                response.config.headers.Authorization = "Bearer " + Token.getAccessToken();
                // Retry the original request
                dataInstance(response.config)
                  .then((dataResult) => {
                    resolve(dataResult);
                  })
                  .catch((dataError) => {
                    reject(dataError);
                  });
              }
            }, 500);
          }
        });
      }
      ElMessage.error(response.data.message ? response.data.message : "服务器错误");
    } else {
      ElMessage.error(response.statusText ? response.statusText : "连接超时");
    }
    return Promise.reject(response);
  },
  (error) => {
    ElMessage.error("连接超时");
    return Promise.reject(error);
  }
);

/**
 * 授权接口respone拦截器
 */
authInstance.interceptors.response.use(
  (response) => {
    // 请求结果码
    if (response.status === 200) {
      // 业务结果码
      if (response.data.code === 200) {
        return response;
      }
      ElMessage.error(response.data.message ? response.data.message : "服务器错误");
    } else {
      ElMessage.error(response.statusText ? response.statusText : "连接超时");
    }
    return Promise.reject(response);
  },
  (error) => {
    ElMessage.error("连接超时");
    return Promise.reject(error);
  }
);

/**
 * 开放接口respone拦截器
 */
openInstance.interceptors.response.use(
  (response) => {
    // 请求结果码
    if (response.status === 200) {
      // 业务结果码
      if (response.data.code === 200) {
        return response;
      }
      ElMessage.error(response.data.message ? response.data.message : "服务器错误");
    } else {
      ElMessage.error(response.statusText ? response.statusText : "连接超时");
    }
    return Promise.reject(response);
  },
  (error) => {
    ElMessage.error("连接超时");
    return Promise.reject(error);
  }
);

/**
 * 调用数据接口
 * @param config
 * @returns
 */
export default <T>(config: AxiosRequestConfig) => {
  return new Promise<ResponseResult<T>>((resolve, reject) => {
    dataInstance
      .request<ResponseResult<T>>(config)
      .then((res) => {
        resolve(res.data);
      })
      .catch((err) => {
        reject(err.data ? err.data : err);
      });
  });
};

/**
 * 调用授权接口
 * @param config
 * @returns
 */
export const authRequest = <T>(config: AxiosRequestConfig) => {
  return new Promise<ResponseResult<T>>((resolve, reject) => {
    authInstance
      .request<ResponseResult<T>>(config)
      .then((res) => {
        resolve(res.data);
      })
      .catch((err) => {
        reject(err.data ? err.data : err);
      });
  });
};

/**
 * 调用开放接口
 * @param config
 * @returns
 */
export const openRequest = <T>(config: AxiosRequestConfig) => {
  return new Promise<ResponseResult<T>>((resolve, reject) => {
    openInstance
      .request<ResponseResult<T>>(config)
      .then((res) => {
        resolve(res.data);
      })
      .catch((err) => {
        reject(err.data ? err.data : err);
      });
  });
};
