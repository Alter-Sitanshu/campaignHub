import axios from "axios";

export const api = axios.create({
    baseURL: "https://frogmedia.onrender.com/api/v1",
    withCredentials: true,
});

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
            "type": "success",
            "status": response.status,
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
            "entity": data.entity,
        };
    } catch(err) {
        if (err.response) {
            return { type: "error", status: err.response.status };
        }
        
        return { type: "error", status: 400 };
    }
}