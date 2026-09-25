<template>
  <figure class="sms-flow" :aria-label="copy.title">
    <figcaption>
      <Icon name="chatBubble" size="sm" /><span>{{ copy.title }}</span
      ><span class="sms-flow-tag">SMS</span>
    </figcaption>
    <div class="sms-flow-platforms">
      <span
        ><img src="/brands/claude.svg" alt="" class="claude-logo" />Claude</span
      ><span
        ><img src="/brands/openai.svg" alt="" class="mono-logo" />ChatGPT</span
      >
    </div>
    <ol class="sms-flow-nodes">
      <li>
        <span class="sms-node-icon amber"><Icon name="key" size="md" /></span>
        <div>
          <span class="sms-node-number">01</span>
          <h3>{{ copy.number }}</h3>
          <p>{{ copy.numberNote }}</p>
        </div>
      </li>
      <li>
        <span class="sms-node-icon blue"
          ><Icon name="chatBubble" size="md"
        /></span>
        <div>
          <span class="sms-node-number">02</span>
          <h3>{{ copy.code }}</h3>
          <div class="sms-code-placeholder" aria-hidden="true">
            <span v-for="digit in 6" :key="digit"><i></i></span>
          </div>
        </div>
      </li>
      <li>
        <span class="sms-node-icon mint"><Icon name="shield" size="md" /></span>
        <div>
          <span class="sms-node-number">03</span>
          <h3>{{ copy.complete }}</h3>
          <p>{{ copy.completeNote }}</p>
        </div>
      </li>
    </ol>
  </figure>
</template>

<script setup>
import Icon from '@/components/icons/Icon.vue'
defineProps({ copy: { type: Object, required: true } })
</script>

