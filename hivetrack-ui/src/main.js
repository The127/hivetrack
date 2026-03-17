import { createApp } from "vue";
import { createRouter, createWebHistory } from "vue-router";
import { VueQueryPlugin } from "@tanstack/vue-query";
import App from "./App.vue";
import "./style.css";

// Router — routes are added as views are built
const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      component: () => import("./views/HomeView.vue"),
    },
  ],
});

const app = createApp(App);
app.use(router);
app.use(VueQueryPlugin);
app.mount("#app");
