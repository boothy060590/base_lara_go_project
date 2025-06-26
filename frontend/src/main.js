import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import { FontAwesomeIcon } from "@/plugins/font-awesome";
import { createPinia } from "pinia";
import "bootstrap/dist/css/bootstrap.min.css";
import "@/assets/scss/utilities.scss";
import "@/assets/scss/form.scss";
import "@/assets/scss/card.scss";
import "@/assets/scss/button.scss";
import config from '@/config';

// Debug environment variables
console.log('Environment variables:', {
  VITE_API_URL: import.meta.env.VITE_API_URL,
  VITE_APP_NAME: import.meta.env.VITE_APP_NAME,
  VITE_VIRTUAL_HOST: import.meta.env.VITE_VIRTUAL_HOST,
  VITE_VIRTUAL_PORT: import.meta.env.VITE_VIRTUAL_PORT
});

document.title = config.appName;

const pinia = createPinia();

try {
  const app = createApp(App);
  app.use(router);
  app.use(pinia);
  app.component("font-awesome-icon", FontAwesomeIcon);
  app.mount("#app");
  console.log('Vue app mounted successfully');
} catch (error) {
  console.error('Error mounting Vue app:', error);
}
