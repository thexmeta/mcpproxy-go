<template>
  <div class="min-h-screen flex items-center justify-center bg-base-200">
    <div class="card w-96 bg-base-100 shadow-xl">
      <div class="card-body items-center text-center">
        <h1 class="card-title text-2xl font-bold">MCPProxy Server</h1>
        <p class="text-base-content/70 mb-4">Sign in to access your MCP tools</p>
        
        <!-- Error message -->
        <div v-if="errorMessage" class="alert alert-error w-full mb-4">
          <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span class="text-sm">{{ errorMessage }}</span>
        </div>
        
        <div class="divider"></div>
        <button
          class="btn btn-primary w-full"
          @click="handleLogin"
          :disabled="loading || configError"
        >
          <span v-if="loading" class="loading loading-spinner"></span>
          <span v-else>Sign in with {{ providerName }}</span>
        </button>
        
        <!-- Configuration help -->
        <div v-if="configError" class="alert alert-warning w-full mt-4">
          <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <div class="text-xs text-left">
            <p class="font-bold mb-1">Administrator Action Required:</p>
            <p class="font-mono">Set teams.oauth.client_id and teams.oauth.client_secret in config</p>
          </div>
        </div>
        
        <p class="text-sm text-base-content/50 mt-4">
          Powered by MCPProxy
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { authApi } from '@/services/auth-api'

const authStore = useAuthStore()

// Provider name will come from status API in future; for now show generic
const providerName = 'your organization'

const loading = ref(false)
const errorMessage = ref<string | null>(null)
const configError = ref(false)

// Check OAuth configuration on mount
onMounted(async () => {
  loading.value = true
  try {
    const oauthConfigError = await authApi.checkOAuthConfig()
    if (oauthConfigError) {
      errorMessage.value = oauthConfigError
      configError.value = true
    }
  } catch (e) {
    console.error('Failed to check OAuth config:', e)
  } finally {
    loading.value = false
  }
})

function handleLogin() {
  if (configError.value) {
    return
  }
  authStore.login()
}
</script>
