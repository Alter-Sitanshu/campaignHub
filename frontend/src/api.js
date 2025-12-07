import axios from "axios";

const api = axios.create({
    baseURL: "http://localhost:8080/api/v1",
    withCredentials: true,
});

export const signup = (data, entity) => {
    let output = "";
    if (!entity || entity.trim().length === 0) {
        return "error";
    } else if (entity === "users"){
        api.post("/users/signup", data)
        .then(response => {
            console.log('Signup success:', response.data);
            output = response.data.id;
        })
        .catch(err => {
            console.log(err);
            output = "error";
        });
    } else if (entity === "brands") {
        api.post("/brands/signup", data)
        .then(response => {
            console.log('Signup success:', response.data);
            output = response.data.id;
        })
        .catch(err => {
            console.log(err);
            output = "error";
        });
    } else {
        return "error";
    }

    return output;
}
