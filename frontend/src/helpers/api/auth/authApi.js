import api from "@/helpers/api/api.js";

export const register = ({
  password,
  password_confirmation,
  first_name,
  last_name,
  email,
  mobile_number,
}) => {
  return api.post("/v1/auth/register", {
    password,
    password_confirmation,
    first_name,
    last_name,
    email,
    mobile_number,
  });
};

export const login = ({ email, password }) => {
  return api.post("/v1/auth/login", {
    email,
    password,
  });
};

export default {
  register,
  login,
};
