<template>
  <div class="w-100">
    <div class="form-group with-icon input-group" :class="{ 'mb-0': hasError }">
      <span v-if="icon" class="input-icon">
        <font-awesome-icon :icon="icon" />
      </span>
      <Field
        :name="name"
        :modelValue="modelValue"
        @update:modelValue="$emit('update:modelValue', $event)"
        :type="type"
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
  name: "TextFormField",
  components: { Field, ErrorMessage, FontAwesomeIcon },
  props: {
    name: { type: String, required: true },
    icon: { type: [String, Array], default: null },
    placeholder: { type: String, default: "" },
    autocomplete: { type: String, default: "" },
    modelValue: { type: String, default: "" },
    type: { type: String, default: "text" },
  },
  emits: ["update:modelValue"],
  computed: {
    hasError() {
      return this.$parent?.errors?.[this.name] || false;
    },
  },
};
</script>
