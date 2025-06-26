<template>
  <div class="w-100">
    <div class="form-group phone-align" :class="{ 'mb-0': hasError }">
      <VueTelInput
        :modelValue="modelValue"
        @update:modelValue="$emit('update:modelValue', $event)"
        :defaultCountry="defaultCountry"
        :preferredCountries="preferredCountries"
        :showDialCode="showDialCode"
        :showFlags="showFlags"
        :mode="mode"
        :placeholder="placeholder"
        class="form-control"
        :id="id"
        autocomplete="tel"
      />
    </div>
    <span v-if="error" class="error-feedback">{{ error }}</span>
  </div>
</template>

<script>
import { VueTelInput } from "vue-tel-input";
import "vue-tel-input/dist/vue-tel-input.css";

export default {
  name: "TelephoneFormField",
  components: { VueTelInput },
  props: {
    modelValue: { type: String, default: "" },
    placeholder: { type: String, default: "" },
    id: { type: String, default: "" },
    error: { type: String, default: "" },
    defaultCountry: { type: String, default: "gb" },
    preferredCountries: { type: Array, default: () => ["gb"] },
    showDialCode: { type: Boolean, default: true },
    showFlags: { type: Boolean, default: true },
    mode: { type: String, default: "international" },
  },
  emits: ["update:modelValue"],
  computed: {
    hasError() {
      return !!this.error;
    },
  },
};
</script>
