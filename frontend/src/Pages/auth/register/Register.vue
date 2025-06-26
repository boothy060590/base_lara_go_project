<template>
  <div class="page-background register-container">
    <div class="container">
      <div class="row justify-content-center align-items-center min-vh-100">
        <div class="col-md-8 col-lg-6">
          <div class="card shadow-lg border-0">
            <div class="card-body px-5">
              <!-- Error Banner at the top -->
              <div v-if="error" class="alert alert-danger mb-4" role="alert">
                {{ error }}
              </div>
              <!-- Header -->
              <div class="text-center mb-5">
                <div class="mb-4">
                  <font-awesome-icon
                    :icon="['fas', 'user-plus']"
                    class="text-primary"
                    size="3x"
                  />
                </div>
                <h1 class="h2 text-dark mb-2">Create Your Account</h1>
                <p class="text-muted">Register to access all our services</p>
              </div>
              <Form
                @submit="handleRegister"
                :validation-schema="schema"
                v-slot="{ submitCount, meta }"
              >
                <TextFormField
                  name="first_name"
                  icon="user"
                  placeholder="First Name"
                  autocomplete="given-name"
                  v-model="user.first_name"
                />
                <TextFormField
                  name="last_name"
                  icon="user"
                  placeholder="Last Name"
                  autocomplete="family-name"
                  v-model="user.last_name"
                />
                <EmailFormField
                  placeholder="Email"
                  autocomplete="email"
                  v-model="user.email"
                  :name="'email'"
                />
                <TelephoneFormField
                  v-model="user.mobile_number"
                  placeholder="+441234567890"
                  id="mobile_number"
                  :defaultCountry="getDefaultCountry()"
                  :preferredCountries="[UK_COUNTRY_CODE]"
                  :showDialCode="true"
                  :showFlags="true"
                  :mode="'international'"
                  :error="
                    submitCount > 0 && !isValidPhone()
                      ? 'Mobile number is invalid!'
                      : ''
                  "
                />
                <PasswordFormField
                  name="password"
                  placeholder="Password"
                  autocomplete="new-password"
                  v-model="user.password"
                />
                <PasswordFormField
                  name="password_confirmation"
                  placeholder="Confirm Password"
                  autocomplete="new-password"
                  v-model="user.password_confirmation"
                />
                <div class="form-group">
                  <button
                    class="btn btn-primary btn-block mt-5 w-100"
                    :disabled="loading || !canSubmit(meta)"
                  >
                    <span
                      v-show="loading"
                      class="spinner-border spinner-border-sm"
                    ></span>
                    <span>Register</span>
                  </button>
                </div>
              </Form>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { Form } from "vee-validate";
import { useAuthStore } from "@/store/auth";
import { mapActions } from "pinia";
import { FontAwesomeIcon } from "@/plugins/font-awesome";
import "vue-tel-input/dist/vue-tel-input.css";
import TextFormField from "@/components/form/TextFormField.vue";
import TelephoneFormField from "@/components/form/TelephoneFormField.vue";
import EmailFormField from "@/components/form/EmailFormField.vue";
import PasswordFormField from "@/components/form/PasswordFormField.vue";
import { registerSchema } from "@/form_validators/register_validator.js";

export default {
  name: "Register",
  components: { Form, TelephoneFormField, EmailFormField, PasswordFormField, FontAwesomeIcon, TextFormField },
  data() {
    return {
      schema: registerSchema,
      user: { password: "", password_confirmation: "", first_name: "", last_name: "", email: "", mobile_number: "" },
      loading: false,
      error: "",
      UK_COUNTRY_CODE: "gb",
    };
  },
  computed: {
    auth() {
      return useAuthStore();
    }
  },
  methods: {
    ...mapActions(useAuthStore, ['register']),
    getDefaultCountry() {
      const country = navigator.language?.split("-")?.[1]?.toLowerCase();
      return (!country || country.length !== 2) ? this.UK_COUNTRY_CODE : country;
    },
    async handleRegister(values, { resetForm }) {
      try {
        this.loading = true;
        const success = await this.register(this.user);
        if (success) {
          this.user = { password: "", password_confirmation: "", first_name: "", last_name: "", email: "", mobile_number: "" };
          resetForm();
          await this.$router.push("/login");
        }
      } catch (err) {
        this.error = err?.message || "Registration failed";
      } finally {
        this.loading = false;
      }
    },
    isValidPhone() {
      const phone = (this.user.mobile_number || "").replace(/\s+/g, "");
      return /^\+[1-9]\d{1,14}$/.test(phone);
    },
    canSubmit(meta) {
      return meta.valid && this.isValidPhone();
    },
  },
};
</script>

<style src="./Register.scss" lang="scss" scoped></style>
