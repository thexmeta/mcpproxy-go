<template>
  <dialog :open="show" class="modal" @close="handleClose">
    <div class="modal-box max-w-2xl">
      <h3 class="text-lg font-bold mb-4">
        Edit Tool: {{ toolName }}
      </h3>

      <form @submit.prevent="handleSave" class="space-y-4">
        <!-- Original Name (read-only) -->
        <div class="form-control">
          <label class="label">
            <span class="label-text font-medium">Original Tool Name</span>
          </label>
          <input
            :value="toolName"
            type="text"
            class="input input-bordered bg-base-200"
            disabled
          />
          <label class="label">
            <span class="label-text-alt text-base-content/60">
              This is the tool name from the MCP server
            </span>
          </label>
        </div>

        <!-- Custom Name (optional) -->
        <div class="form-control">
          <label class="label">
            <span class="label-text font-medium">Custom Tool Name (optional)</span>
          </label>
          <input
            v-model="formData.customName"
            type="text"
            class="input input-bordered"
            placeholder="Leave empty to use original name"
          />
          <label class="label">
            <span class="label-text-alt text-base-content/60">
              Override the display name for this tool
            </span>
          </label>
        </div>

        <!-- Custom Description (optional) -->
        <div class="form-control">
          <label class="label">
            <span class="label-text font-medium">Custom Description (optional)</span>
          </label>
          <textarea
            v-model="formData.customDescription"
            class="textarea textarea-bordered min-h-[100px]"
            placeholder="Leave empty to use original description"
          ></textarea>
          <label class="label">
            <span class="label-text-alt text-base-content/60">
              Override the tool description
            </span>
          </label>
        </div>

        <!-- Enable/Disable Toggle -->
        <div class="form-control">
          <label class="label cursor-pointer justify-start gap-4">
            <span class="label-text font-medium">Enabled</span>
            <input
              v-model="formData.enabled"
              type="checkbox"
              class="toggle toggle-primary"
            />
          </label>
          <label class="label">
            <span class="label-text-alt text-base-content/60">
              Disabled tools are hidden from AI agents
            </span>
          </label>
        </div>

        <!-- Action Buttons -->
        <div class="modal-action mt-6">
          <button
            type="button"
            class="btn"
            @click="handleClose"
          >
            Cancel
          </button>
          <button
            type="button"
            class="btn btn-ghost"
            :disabled="!hasChanges"
            @click="handleReset"
          >
            Reset to Defaults
          </button>
          <button
            type="submit"
            class="btn btn-primary"
            :disabled="saving"
          >
            <span v-if="saving" class="loading loading-spinner"></span>
            Save Changes
          </button>
        </div>
      </form>
    </div>
    <form method="dialog" class="modal-backdrop" @click="handleClose">
      <button>close</button>
    </form>
  </dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { ToolPreference } from '@/types/api'

const props = defineProps<{
  show: boolean
  serverName: string
  toolName: string
  preference?: ToolPreference | null
  originalDescription?: string
}>()

const emit = defineEmits<{
  close: []
  save: [preference: { enabled: boolean; custom_name?: string; custom_description?: string }]
  reset: []
}>()

const saving = ref(false)

const formData = ref({
  enabled: true,
  customName: '',
  customDescription: '',
})

const hasChanges = computed(() => {
  if (!props.preference) {
    return formData.value.customName !== '' || formData.value.customDescription !== '' || !formData.value.enabled
  }
  return (
    formData.value.enabled !== props.preference.enabled ||
    formData.value.customName !== (props.preference.custom_name || '') ||
    formData.value.customDescription !== (props.preference.custom_description || '')
  )
})

watch(
  () => props.show,
  (newShow) => {
    if (newShow && props.preference) {
      formData.value = {
        enabled: props.preference.enabled,
        customName: props.preference.custom_name || '',
        customDescription: props.preference.custom_description || '',
      }
    } else if (newShow) {
      // No existing preference, use defaults
      formData.value = {
        enabled: true,
        customName: '',
        customDescription: '',
      }
    }
  },
  { immediate: true }
)

function handleClose() {
  emit('close')
}

async function handleSave() {
  saving.value = true
  try {
    emit('save', {
      enabled: formData.value.enabled,
      custom_name: formData.value.customName || undefined,
      custom_description: formData.value.customDescription || undefined,
    })
  } finally {
    saving.value = false
  }
}

function handleReset() {
  emit('reset')
}
</script>

<style scoped>
.modal {
  &::backdrop {
    background: rgba(0, 0, 0, 0.5);
  }
}
</style>
