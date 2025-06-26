<template>
  <div class="page-background home-container">
    <div class="container">
      <div class="row justify-content-center align-items-center min-vh-100">
        <div class="col-md-8 col-lg-6">
          <div class="card shadow-lg border-0">
            <div class="card-body p-5">
              <!-- Header -->
              <div class="text-center mb-5">
                <div class="mb-4">
                  <font-awesome-icon
                    :icon="['fas', 'home']"
                    class="text-primary"
                    size="3x"
                  />
                </div>
                <h1 class="display-4 fw-bold text-primary mb-4">
                  Welcome to Base Laravel Go Project
                </h1>
                <p class="lead text-muted mb-5">
                  Your modern Laravel-inspired Go API with Vue.js frontend
                </p>
              </div>

              <!-- Action Buttons -->
              <div class="row g-3">
                <div class="col-12">
                  <router-link
                    to="/login"
                    class="btn btn-primary btn-lg w-100 d-flex align-items-center justify-content-center gap-2"
                  >
                    <font-awesome-icon :icon="['fas', 'sign-in-alt']" />
                    Sign In
                  </router-link>
                </div>
                <div class="col-12">
                  <router-link
                    to="/register"
                    class="btn btn-outline-primary btn-lg w-100 d-flex align-items-center justify-content-center gap-2"
                  >
                    <font-awesome-icon :icon="['fas', 'user-plus']" />
                    Create Account
                  </router-link>
                </div>
              </div>

              <!-- Features -->
              <section class="why-choose-us py-5 bg-light">
                <div class="container text-center">
                  <h2 class="mb-5">Why Use This Starter?</h2>
                  <div class="row justify-content-center">
                    <div
                      class="col-12 col-md-6 mb-4 d-flex flex-column align-items-center"
                      v-for="feature in features"
                      :key="feature.label"
                    >
                      <font-awesome-icon :icon="feature.icon" :class="feature.color + ' mb-3'" size="3x" />
                      <div class="h5 text-center">{{ feature.label }}</div>
                    </div>
                  </div>
                </div>
              </section>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { useAuthStore } from "@/store/auth";

export default {
  name: "Home",
  data() {
    return {
      features: [
        { icon: ['fas', 'rocket'], color: 'text-primary', label: 'Rapid Development' },
        { icon: ['fas', 'shield-alt'], color: 'text-success', label: 'Modern Security' },
        { icon: ['fas', 'bolt'], color: 'text-warning', label: 'Scalable Architecture' },
        { icon: ['fas', 'puzzle-piece'], color: 'text-info', label: 'Modular & Extensible' }
      ]
    }
  },
  // mounted() {
  //   // Check if user is already authenticated and redirect accordingly
  //   const auth = useAuthStore();
  //   if (auth.isAuthenticated && auth.roles.length > 0) {
  //     // Redirect to appropriate role-based dashboard
  //     const role = auth.roles[0]; // Take the first role
  //     this.$router.push(`/${role}`);
  //   }
  // },
};
</script>
