<template>            <div class="auth-field">
              <label for="auth-email">{{ c.email }}</label>
              <div
                class="auth-input-wrap"
                :class="{ 'has-error': errors.email }"
              >
                <Icon name="mail" size="sm" class="auth-field-icon" />
                <input
                  id="auth-email"
                  :disabled="disabled"
                  v-model="email"
                  type="email"
                  required
                  autocomplete="email"
                  autocapitalize="none"
                  :spellcheck="false"
                  :placeholder="
                    isRegister ? (emailPlaceholder ?? c.registerEmailPlaceholder) : c.emailPlaceholder
                  "
                  :aria-invalid="!!errors.email"
                  :aria-describedby="
                    errors.email
                      ? 'email-error'
                      : isRegister
                        ? 'email-hint'
                        : undefined
                  "
                  @input="clearField('email')"
                  @blur="$emit('email-blur')"
                />
              </div>
              <p
                v-if="errors.email"
                id="email-error"
                class="auth-field-error"
                role="alert"
              >
                <Icon name="exclamationCircle" size="xs" />{{ errors.email }}
              </p>
              <p v-else-if="isRegister" id="email-hint" class="auth-field-hint">
                {{ emailHint ?? c.emailHint }}
              </p>
            </div>

            <div class="auth-field">
              <div class="auth-field-label">
                <label for="auth-password">{{
                  isRegister ? c.newPassword : c.password
                }}</label
                ><button
                  v-if="!isRegister"
                  type="button"
                  class="auth-inline-link"
                  @click="$emit('recover')"
                >
                  {{ c.forgot }}
                </button>
              </div>
              <div
                class="auth-input-wrap"
                :class="{ 'has-error': errors.password }"
              >
                <Icon name="lock" size="sm" class="auth-field-icon" />
                <input
                  id="auth-password"
                  :disabled="disabled"
                  v-model="password"
                  :type="showPassword ? 'text' : 'password'"
                  required
                  minlength="6"
                  :autocomplete="
                    isRegister ? 'new-password' : 'current-password'
                  "
                  :placeholder="
                    isRegister
                      ? c.newPasswordPlaceholder
                      : c.passwordPlaceholder
                  "
                  :aria-invalid="!!errors.password"
                  :aria-describedby="
                    errors.password
                      ? 'password-error'
                      : isRegister
                        ? 'password-hint'
                        : undefined
                  "
                  @input="clearField('password')"
                  @keydown="capsLock = $event.getModifierState('CapsLock')"
                  @keyup="capsLock = $event.getModifierState('CapsLock')"
                  @blur="capsLock = false"
                />
                <button
                  class="icon-button auth-password-toggle"
                  type="button"
                  :title="showPassword ? c.hidePassword : c.showPassword"
                  :aria-label="showPassword ? c.hidePassword : c.showPassword"
                  :aria-pressed="showPassword"
                  @click="showPassword = !showPassword"
                >
                  <Icon :name="showPassword ? 'eyeOff' : 'eye'" size="sm" />
                </button>
              </div>
              <p
                v-if="errors.password"
                id="password-error"
                class="auth-field-error"
                role="alert"
              >
                <Icon name="exclamationCircle" size="xs" />{{
                  errors.password
                }}
              </p>
              <p
                v-else-if="isRegister"
                id="password-hint"
                class="auth-field-hint"
              >
                {{ c.passwordHint }}
              </p>
              <p v-if="capsLock" class="auth-caps-warning" role="status">
                {{ c.capsLock }}
              </p>
            </div>

</template>
<script setup lang="ts">
import { ref } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import { authCopy } from './auth-content'
defineProps<{ c: typeof authCopy.zh; isRegister?: boolean; disabled?: boolean; errors: { email: string; password: string }; emailHint?: string; emailPlaceholder?: string }>()
const emit = defineEmits<{ recover: []; 'email-blur': []; clear: [field: 'email' | 'password'] }>()
const email = defineModel<string>('email', { required: true })
const password = defineModel<string>('password', { required: true })
const showPassword = ref(false), capsLock = ref(false)
function clearField(field: 'email' | 'password') { emit('clear', field) }
</script>
