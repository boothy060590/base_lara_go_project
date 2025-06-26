<template>
  <div class="w-100">
    <div class="form-group with-icon input-group" :class="{ 'mb-0': hasError }">
      <span class="input-icon">
        <font-awesome-icon icon="lock" />
      </span>
      <Field
        :name="name"
        :modelValue="modelValue"
        @update:modelValue="$emit('update:modelValue', $event)"
        :type="showPassword ? 'text' : 'password'"
        class="form-control"
        :autocomplete="autocomplete"
        :placeholder="placeholder"
      />
      <button
        type="button"
        class="password-toggle-btn"
        @click="togglePassword"
        :aria-label="showPassword ? 'Hide password' : 'Show password'"
      >
        <font-awesome-icon
          :icon="showPassword ? 'eye-slash' : 'eye'"
          class="password-toggle-icon"
        />
      </button>
    </div>
    <ErrorMessage :name="name" class="error-feedback" />
  </div>
</template>

<script>
import { Field, ErrorMessage } from "vee-validate";
import { FontAwesomeIcon } from "@/plugins/font-awesome";

export default {
  name: "PasswordFormField",
  components: { Field, ErrorMessage, FontAwesomeIcon },
  props: {
    name: { type: String, required: true },
    placeholder: { type: String, default: "" },
    autocomplete: { type: String, default: "" },
    modelValue: { type: String, default: "" },
  },
  emits: ["update:modelValue"],
  data() {
    return {
      showPassword: false,
    };
  },
  computed: {
    hasError() {
      return this.$parent?.errors?.[this.name] || false;
    },
  },
  methods: {
    togglePassword() {
      this.showPassword = !this.showPassword;
    },
  },
};
</script>
