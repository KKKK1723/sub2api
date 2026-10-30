<template>
  <div class="card">
    <div
      class="flex items-start justify-between gap-4 border-b border-gray-100 px-6 py-4 dark:border-dark-700"
    >
      <div>
        <h3 class="text-base font-medium text-gray-900 dark:text-white">
          {{ t("admin.settings.accountAlertNotify.title") }}
        </h3>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          {{ t("admin.settings.accountAlertNotify.description") }}
        </p>
      </div>
      <button
        type="button"
        class="btn btn-primary btn-sm shrink-0"
        :disabled="loading || loadFailed || saving"
        @click="save"
      >
        {{
          saving
            ? t("admin.settings.accountAlertNotify.saving")
            : t("admin.settings.accountAlertNotify.save")
        }}
      </button>
    </div>

    <div
      v-if="loading"
      class="flex items-center gap-2 px-6 py-6 text-sm text-gray-500 dark:text-gray-400"
    >
      <span
        class="h-4 w-4 animate-spin rounded-full border-b-2 border-primary-600"
      ></span>
      {{ t("common.loading") }}
    </div>

    <div
      v-else-if="loadFailed"
      class="flex items-center gap-3 px-6 py-6 text-sm text-red-600 dark:text-red-400"
    >
      {{ t("admin.settings.accountAlertNotify.loadFailed") }}
      <button type="button" class="btn btn-secondary btn-sm" @click="load">
        {{ t("common.refresh") }}
      </button>
    </div>

    <div v-else class="divide-y divide-gray-100 px-6 dark:divide-dark-700">
      <!-- 收件邮箱 -->
      <div class="py-5">
        <label
          class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
        >
          {{ t("admin.settings.accountAlertNotify.emails") }}
        </label>
        <p class="mb-3 text-xs text-gray-500 dark:text-gray-400">
          {{ t("admin.settings.accountAlertNotify.emailsHint") }}
        </p>
        <div class="space-y-2">
          <div
            v-for="(entry, index) in form.emails"
            :key="index"
            class="flex items-center gap-2"
          >
            <label
              class="relative inline-flex items-center cursor-pointer shrink-0"
            >
              <input
                type="checkbox"
                :checked="!entry.disabled"
                @change="entry.disabled = !entry.disabled"
                class="sr-only peer"
              />
              <div
                class="w-9 h-5 bg-gray-200 peer-focus:outline-none rounded-full peer dark:bg-gray-600 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all dark:after:border-gray-500 peer-checked:bg-primary-600"
              ></div>
            </label>
            <input
              v-model="entry.email"
              type="email"
              class="input flex-1"
              :placeholder="t('admin.settings.accountAlertNotify.emailPlaceholder')"
            />
            <button
              type="button"
              class="btn btn-secondary px-2"
              @click="removeEmail(index)"
            >
              <Icon name="x" size="xs" class="h-4 w-4" />
            </button>
          </div>
          <button
            v-if="form.emails.length < MAX_EMAILS"
            type="button"
            class="btn btn-secondary btn-sm"
            @click="addEmail"
          >
            + {{ t("admin.settings.accountAlertNotify.addEmail") }}
          </button>
        </div>
      </div>

      <!-- 账号出错通知 -->
      <div class="flex items-center justify-between gap-4 py-5">
        <div>
          <label
            class="mb-0 block text-sm font-medium text-gray-700 dark:text-gray-300"
          >
            {{ t("admin.settings.accountAlertNotify.accountError") }}
          </label>
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
            {{ t("admin.settings.accountAlertNotify.accountErrorHint") }}
          </p>
        </div>
        <Toggle v-model="form.account_error_enabled" />
      </div>

      <!-- 分组无可用账号通知 -->
      <div class="py-5">
        <div class="flex items-center justify-between gap-4">
          <div>
            <label
              class="mb-0 block text-sm font-medium text-gray-700 dark:text-gray-300"
            >
              {{ t("admin.settings.accountAlertNotify.groupUnavailable") }}
            </label>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              {{ t("admin.settings.accountAlertNotify.groupUnavailableHint") }}
            </p>
          </div>
          <Toggle v-model="form.group_unavailable_enabled" />
        </div>
        <div
          v-if="form.group_unavailable_enabled"
          class="mt-4 space-y-4 rounded-lg bg-gray-50 p-4 dark:bg-dark-700/40"
        >
          <div>
            <label
              class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
            >
              {{ t("admin.settings.accountAlertNotify.groupDelay") }}
            </label>
            <div class="flex items-center gap-2">
              <input
                v-model.number="form.group_unavailable_delay_minutes"
                type="number"
                min="0"
                max="60"
                step="1"
                class="input w-28"
              />
              <span class="text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.accountAlertNotify.groupDelayUnit") }}
              </span>
            </div>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              {{ t("admin.settings.accountAlertNotify.groupDelayHint") }}
            </p>
          </div>
          <div class="flex items-center justify-between gap-4">
            <div>
              <label
                class="mb-0 block text-sm font-medium text-gray-700 dark:text-gray-300"
              >
                {{ t("admin.settings.accountAlertNotify.groupRecovered") }}
              </label>
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.accountAlertNotify.groupRecoveredHint") }}
              </p>
            </div>
            <Toggle v-model="form.group_recovered_enabled" />
          </div>
        </div>
        <div
          class="mt-3 rounded-lg bg-blue-50 p-3 text-xs text-blue-700 dark:bg-blue-900/20 dark:text-blue-300"
        >
          {{ t("admin.settings.accountAlertNotify.groupScopeInfo") }}
        </div>
      </div>

      <!-- 上游余额不足通知 -->
      <div class="py-5">
        <div class="flex items-center justify-between gap-4">
          <div>
            <label
              class="mb-0 block text-sm font-medium text-gray-700 dark:text-gray-300"
            >
              {{ t("admin.settings.accountAlertNotify.balanceLow") }}
            </label>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              {{ t("admin.settings.accountAlertNotify.balanceLowHint") }}
            </p>
          </div>
          <Toggle v-model="form.upstream_balance_low_enabled" />
        </div>
        <div
          v-if="form.upstream_balance_low_enabled"
          class="mt-4 rounded-lg bg-gray-50 p-4 dark:bg-dark-700/40"
        >
          <label
            class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
          >
            {{ t("admin.settings.accountAlertNotify.balanceThreshold") }}
          </label>
          <div class="flex items-center gap-2">
            <input
              v-model.number="form.upstream_balance_low_threshold"
              type="number"
              min="0"
              step="0.01"
              class="input w-36"
            />
            <span class="text-sm text-gray-500 dark:text-gray-400">
              {{ t("admin.settings.accountAlertNotify.balanceThresholdUnit") }}
            </span>
          </div>
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
            {{ t("admin.settings.accountAlertNotify.balanceThresholdHint") }}
          </p>
        </div>
        <div
          class="mt-3 rounded-lg bg-blue-50 p-3 text-xs text-blue-700 dark:bg-blue-900/20 dark:text-blue-300"
        >
          {{ t("admin.settings.accountAlertNotify.balanceScopeInfo") }}
        </div>
      </div>

      <!-- 发送测试通知 -->
      <div class="flex items-center justify-between gap-4 py-5">
        <div>
          <label
            class="mb-0 block text-sm font-medium text-gray-700 dark:text-gray-300"
          >
            {{ t("admin.settings.accountAlertNotify.test") }}
          </label>
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
            {{ t("admin.settings.accountAlertNotify.testHint") }}
          </p>
        </div>
        <button
          type="button"
          class="btn btn-secondary btn-sm shrink-0"
          :disabled="testing"
          @click="sendTest"
        >
          {{
            testing
              ? t("admin.settings.accountAlertNotify.testing")
              : t("admin.settings.accountAlertNotify.test")
          }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import { adminAPI } from "@/api";
import type { AccountAlertNotifySettings } from "@/api/admin/settings";
import Icon from "@/components/icons/Icon.vue";
import Toggle from "@/components/common/Toggle.vue";
import { useAppStore } from "@/stores";
import { extractApiErrorMessage } from "@/utils/apiError";

const { t } = useI18n();
const appStore = useAppStore();

// 与后端收件人上限保持一致
const MAX_EMAILS = 10;
const EMAIL_PATTERN = /^[^\s@<>]+@[^\s@<>]+\.[^\s@<>]+$/;

const loading = ref(true);
const loadFailed = ref(false);
const saving = ref(false);
const testing = ref(false);

const form = reactive<AccountAlertNotifySettings>({
  emails: [],
  account_error_enabled: true,
  group_unavailable_enabled: true,
  group_unavailable_delay_minutes: 3,
  group_recovered_enabled: true,
  upstream_balance_low_enabled: true,
  upstream_balance_low_threshold: 10,
});

function applySettings(settings: AccountAlertNotifySettings) {
  form.emails = (settings.emails || []).map((entry) => ({ ...entry }));
  form.account_error_enabled = settings.account_error_enabled;
  form.group_unavailable_enabled = settings.group_unavailable_enabled;
  form.group_unavailable_delay_minutes = settings.group_unavailable_delay_minutes;
  form.group_recovered_enabled = settings.group_recovered_enabled;
  form.upstream_balance_low_enabled = settings.upstream_balance_low_enabled;
  form.upstream_balance_low_threshold = settings.upstream_balance_low_threshold;
}

async function load() {
  loading.value = true;
  loadFailed.value = false;
  try {
    applySettings(await adminAPI.settings.getAccountAlertNotifySettings());
  } catch (error: unknown) {
    loadFailed.value = true;
    appStore.showError(
      extractApiErrorMessage(
        error,
        t("admin.settings.accountAlertNotify.loadFailed"),
      ),
    );
  } finally {
    loading.value = false;
  }
}

function addEmail() {
  if (form.emails.length >= MAX_EMAILS) {
    return;
  }
  form.emails.push({ email: "", disabled: false, verified: true });
}

function removeEmail(index: number) {
  form.emails.splice(index, 1);
}

function trimmedEmails(): string[] {
  return form.emails.map((entry) => entry.email.trim()).filter(Boolean);
}

function validateEmails(emails: string[]): boolean {
  const invalid = emails.find((email) => !EMAIL_PATTERN.test(email));
  if (invalid) {
    appStore.showError(
      t("admin.settings.accountAlertNotify.invalidEmail", { email: invalid }),
    );
    return false;
  }
  return true;
}

function validate(): boolean {
  if (!validateEmails(trimmedEmails())) {
    return false;
  }
  const delay = Number(form.group_unavailable_delay_minutes);
  if (!Number.isInteger(delay) || delay < 0 || delay > 60) {
    appStore.showError(t("admin.settings.accountAlertNotify.invalidDelay"));
    return false;
  }
  const threshold = Number(form.upstream_balance_low_threshold);
  if (!Number.isFinite(threshold) || threshold < 0) {
    appStore.showError(t("admin.settings.accountAlertNotify.invalidThreshold"));
    return false;
  }
  return true;
}

async function save() {
  if (!validate()) {
    return;
  }
  saving.value = true;
  try {
    const updated = await adminAPI.settings.updateAccountAlertNotifySettings({
      emails: form.emails
        .map((entry) => ({
          email: entry.email.trim(),
          disabled: entry.disabled,
          verified: true,
        }))
        .filter((entry) => entry.email),
      account_error_enabled: form.account_error_enabled,
      group_unavailable_enabled: form.group_unavailable_enabled,
      group_unavailable_delay_minutes: Number(
        form.group_unavailable_delay_minutes,
      ),
      group_recovered_enabled: form.group_recovered_enabled,
      upstream_balance_low_enabled: form.upstream_balance_low_enabled,
      upstream_balance_low_threshold: Number(
        form.upstream_balance_low_threshold,
      ),
    });
    applySettings(updated);
    appStore.showSuccess(t("admin.settings.accountAlertNotify.saved"));
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t("common.error")));
  } finally {
    saving.value = false;
  }
}

async function sendTest() {
  // 只发给当前启用的邮箱，不要求先保存
  const emails = form.emails
    .filter((entry) => !entry.disabled)
    .map((entry) => entry.email.trim())
    .filter(Boolean);
  if (emails.length === 0) {
    appStore.showError(t("admin.settings.accountAlertNotify.testNoEmail"));
    return;
  }
  if (!validateEmails(emails)) {
    return;
  }
  testing.value = true;
  try {
    const result = await adminAPI.settings.sendAccountAlertNotifyTest(emails);
    appStore.showSuccess(
      t("admin.settings.accountAlertNotify.testSuccess", {
        sent: result.sent,
        total: result.total,
      }),
    );
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t("common.error")));
  } finally {
    testing.value = false;
  }
}

onMounted(load);
</script>
