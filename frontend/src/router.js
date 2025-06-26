import { createWebHistory, createRouter } from "vue-router";
import Login from "./Pages/auth/login/Login.vue";
import Home from "./Pages/home/Home.vue";
import { useAuthStore } from "./store/auth";

const Customer = () => import("./Pages/home/customer/Customer.vue");
const Admin = () => import("./Pages/home/admin/Admin.vue");
const Register = () => import("./Pages/auth/register/Register.vue");

const routes = [
  {
    path: "/",
    name: "home",
    component: Home,
  },
  {
    path: "/login",
    name: "login",
    component: Login,
  },
  {
    path: "/customer",
    name: "customer",
    component: Customer,
    meta: { requiresAuth: true, role: "customer" },
  },
  {
    path: "/admin",
    name: "admin",
    component: Admin,
    meta: { requiresAuth: true, role: "admin" },
  },
  {
  },
  {
    path: "/register",
    name: "register",
    component: Register,
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to, from, next) => {
  const auth = useAuthStore();

  // If user is authenticated and trying to access home/login/register, redirect to their dashboard
  if (
    auth.isAuthenticated &&
    auth.roles.length > 0 &&
    ["home", "login", "register"].includes(to.name)
  ) {
    const role = auth.roles[0]; // Take the first role
    return next({ name: role });
  }

  if (to.meta.requiresAuth) {
    if (!auth.isAuthenticated) {
      return next({ name: "home" });
    }
    if (to.meta.role && !auth.roles.includes(to.meta.role)) {
      // Not authorized for this role
      return next({ name: "home" });
    }
  }
  next();
});

export default router;
