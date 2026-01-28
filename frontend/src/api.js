import axios from "axios";
import { getLogoutHandler } from "./AuthContext";
import { useNavigate } from "react-router-dom";

export const api = axios.create({
    baseURL: "https://frogmedia.onrender.com/api/v1",
    // baseURL: "http://localhost:8080/api/v1",
    withCredentials: true,
});

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401 || error.response?.status === 403) {
      try {
        let logout = getLogoutHandler();
        logout?.(); // clear tokens, user state, etc.
      } finally {
        window.location.href = "/auth/sign_in";
      }
    }

    return Promise.reject(error);
  }
)

export const signup = async (data, entity) => {
    if (!entity || entity.trim().length === 0) {
        return {
            "type": "error",
            "status": 400,
        };
    }
    try {
        let response;
        if (entity === "users") {
            response = await api.post("/users/signup", data);
        } else if (entity === "brands") {
            response = await api.post("/brands/signup", data)
        } else {
            return {
                "type": "error",
                "status": 400
            };
        }
        if (response.status != 201) {
            return {
                "type": "error",
                "status": response.status
            };
        };
        return {
            type: "success",
            status: response.status,
            id: response.data.data,
        };
    } catch {
        return {
            "type": "error",
            "status": 500
        };
    }
}

export const signin = async (data) => {
    const email = data.email;
    if (!email || email.trim().length === 0) {
        return {
            "type": "error",
            "status": 400,
        };
    }
    try {
        let response;
        response = await api.post("/login", data);
        if (response.status != 200) {
            return {
                "type": "error",
                "status": response.status
            };
        };
        return {
            "type": "success",
            "status": response.status,
            "id": response.data.data.id,
            "username": response.data.data.username,
            "email": response.data.data.email,
            "account_exists": response.data.data.account_exists,
            "entity": data.entity,
        };
    } catch(err) {
        if (err.response) {
            return { type: "error", status: err.response.status };
        }
        
        return { type: "error", status: 400 };
    }
}

export const oauthCallbackSignIn = async (payload) => {
    // payload: { email, first_name, last_name, entity }
    if (!payload || !payload.email || !payload.entity) {
        return { type: "error", status: 400 };
    }
    try {
        const response = await api.post('/oauth/callback', payload);
        if (response.status !== 200) {
            if (response.status === 401) {
                // user is not a registered user yet | redirect to sign up
                window.location.href = "/auth/users/sign_up";
                return;
            }
            return { type: 'error', status: response.status };
        }
        return {
            type: 'success',
            status: response.status,
            id: response.data.data.id,
            username: response.data.data.username,
            email: response.data.data.email,
            account_exists: response.data.data.account_exists,
        };
    } catch (err) {
        if (err.response) return { type: 'error', status: err.response.status };
        return { type: 'error', status: 500 };
    }
}