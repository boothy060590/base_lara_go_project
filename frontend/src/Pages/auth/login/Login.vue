<script>
import { Form } from "vee-validate";
import { useAuthStore } from "@/store/auth";
import { mapActions } from "pinia";
import { FontAwesomeIcon } from "@/plugins/font-awesome";
import EmailFormField from "@/components/form/EmailFormField.vue";
import PasswordFormField from "@/components/form/PasswordFormField.vue";
import { loginSchema } from "@/form_validators/login_validator.js";

export default {
  name: "Login",
  components: {
    Form,
    FontAwesomeIcon,
    EmailFormField,
    PasswordFormField,
  },
  data() {
    return {
      schema: loginSchema,
      user: {
        email: "",
        password: "",
      },
    };
  },
  computed: {
    auth() {
      return useAuthStore();
    }
  },
  created() {
    if (this.loggedIn) {
      // this.$router.push("/profile");
    }

    // Check for registration success from store
    if (this.auth.registrationSuccess && this.auth.registeredEmail) {
      // Pre-fill email if available
      this.user.email = this.auth.registeredEmail;
    }
  },
  methods: {
    ...mapActions(useAuthStore, ['login']),
    async handleLogin() {
      await this.login({
        email: this.user.email,
        password: this.user.password,
      });
    },
  },
  mounted() {
    if (this.auth.registrationSuccess && this.auth.registeredEmail) {
      this.user.email = this.auth.registeredEmail;
      setTimeout(() => {
        this.auth.clearRegistrationSuccess();
      }, 3000); // Show banner for 3 seconds
    }
  },
};
</script>

<template>
  <div class="page-background login-container">
    <div class="container">
      <div class="row justify-content-center align-items-center min-vh-100">
        <div class="col-md-8 col-lg-6">
          <div class="card shadow-lg border-0">
            <div class="card-body px-5">
              <!-- Error Banner at the top -->
              <div
                v-if="auth.error"
                class="alert alert-danger mb-4"
                role="alert"
              >
                {{ auth.error }}
              </div>
              <!-- Success Banner at the top -->
              <div
                v-if="auth.registrationSuccess"
                class="alert alert-success mb-4"
                role="alert"
              >
                {{ auth.registrationMessage }}
              </div>

              <!-- Header -->
              <div class="text-center mb-5">
                <div class="mb-4">
                  <font-awesome-icon
                    :icon="['fas', 'user']"
                    class="text-primary"
                    size="3x"
                  />
                </div>
                <h1 class="h2 text-dark mb-2">Sign In</h1>
                <p class="text-muted">Access your account</p>
              </div>

              <Form
                @submit="handleLogin"
                :validation-schema="schema"
                v-slot="{ submitCount }"
              >
                <EmailFormField
                  name="email"
                  placeholder="Email"
                  autocomplete="email"
                  v-model="user.email"
                />
                <PasswordFormField
                  name="password"
                  placeholder="Password"
                  autocomplete="current-password"
                  v-model="user.password"
                />
                <div class="form-group">
                  <button
                    class="btn btn-primary btn-block w-100 mt-4"
                    :disabled="auth.loading"
                  >
                    <span
                      v-show="auth.loading"
                      class="spinner-border spinner-border-sm"
                    ></span>
                    <span>Login</span>
                  </button>
                </div>
              </Form>

              <div class="mt-3 text-center">
                <router-link to="/register"
                  >Don't have an account? Register</router-link
                >
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style src="./login.scss" lang="scss" scoped></style>
