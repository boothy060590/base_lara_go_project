<template>
  <div class="w-100">
    <div class="form-group with-icon input-group" :class="{ 'mb-0': hasError }">
      <span class="input-icon">
        <font-awesome-icon icon="envelope" />
      </span>
      <Field
        :name="name"
        :modelValue="modelValue"
        @update:modelValue="$emit('update:modelValue', $event)"
        type="email"
        class="form-control"
        :autocomplete="autocomplete"
        :placeholder="placeholder"
      />
    </div>
    <ErrorMessage :name="name" class="error-feedback" />
  </div>
</template>

<script>
import { Field, ErrorMessage } from "vee-validate";
import { FontAwesomeIcon } from "@/plugins/font-awesome";

export default {
  name: "EmailFormField",
  components: { Field, ErrorMessage, FontAwesomeIcon },
  props: {
    name: { type: String, required: true },
    placeholder: { type: String, default: "" },
    autocomplete: { type: String, default: "" },
    modelValue: { type: String, default: "" },
  },
  emits: ["update:modelValue"],
  computed: {
    hasError() {
      return this.$parent?.errors?.[this.name] || false;
    },
  },
};
</script>
