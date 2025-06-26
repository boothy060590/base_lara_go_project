import { defineStore } from "pinia";
import axios from "axios";
import router from "../router";
import authApi from "@/helpers/api/auth/authApi";

export const useAuthStore = defineStore("auth", {
  state: () => ({
    token: localStorage.getItem("token") || "",
    user: null,
    roles: [],
    loading: false,
    error: "",
    registrationSuccess: false,
    registrationMessage: "",
    registeredEmail: "",
  }),
  getters: {
    isAuthenticated: (state) => !!state.token,
    primaryRole: (state) => {
      return rolePriority.find(role => state.roles.includes(role)) || (state.roles[0] || "");
    },
  },
  actions: {
    async login({ email, password }) {
      this.loading = true;
      this.error = "";
      try {
        const res = await authApi.login({ email, password });
        this.token = res.data.token;
        this.roles = Array.isArray(res.data.roles) ? res.data.roles : [res.data.role].filter(Boolean);
        localStorage.setItem("token", this.token);
        axios.defaults.headers.common["Authorization"] = `Bearer ${this.token}`;
        await router.push(this.primaryRole ? `/${this.primaryRole}` : "/");
      } catch (err) {
        this.error = err.response?.data?.error || "Login failed";
      } finally {
        this.loading = false;
      }
    },
    async register(userData) {
      this.loading = true;
      this.error = "";
      try {
        const formattedPhone = (userData.mobile_number || "").replace(/\s+/g, "");
        await authApi.register({ ...userData, mobile_number: formattedPhone });
        this.setRegistrationSuccess("Registration successful! You can now log in.", userData.email);
        return true;
      } catch (err) {
        this.error = err.response?.data?.error || "Registration failed";
        return false;
      } finally {
        this.loading = false;
      }
    },
    async logout() {
      this.token = "";
      this.user = null;
      this.roles = [];
      localStorage.removeItem("token");
      delete axios.defaults.headers.common["Authorization"];
      await router.push("/login");
    },
    setUser(user) {
      this.user = user;
      this.roles = user.roles || [];
    },
    initialize() {
      if (this.token) {
        axios.defaults.headers.common["Authorization"] = `Bearer ${this.token}`;
      }
    },
    setRegistrationSuccess(message, email) {
      this.registrationSuccess = true;
      this.registrationMessage = message;
      this.registeredEmail = email;
    },
    clearRegistrationSuccess() {
      this.registrationSuccess = false;
      this.registrationMessage = "";
      this.registeredEmail = "";
    },
  },
});
