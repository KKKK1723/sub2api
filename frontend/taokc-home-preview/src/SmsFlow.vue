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
import Icon from '../../src/components/icons/Icon.vue'
defineProps({ copy: { type: Object, required: true } })
</script>

<style>
.sms-flow {
  position: relative;
  margin: 0;
  padding: 27px 30px 30px;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: var(--page);
}
.sms-flow figcaption {
  display: flex;
  align-items: center;
  gap: 9px;
  color: var(--muted);
  font-size: 12px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--line);
}
.sms-flow-tag {
  margin-left: auto;
  font:
    10px ui-monospace,
    monospace;
  color: var(--amber);
}
.sms-flow-platforms {
  display: flex;
  align-items: center;
  gap: 25px;
  padding-block: 22px 26px;
}
.sms-flow-platforms > span {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 550;
}
.sms-flow-platforms img {
  width: 22px;
  height: 22px;
  object-fit: contain;
}
.sms-flow-nodes {
  display: grid;
  gap: 30px;
}
.sms-flow-nodes > li {
  position: relative;
  display: grid;
  grid-template-columns: 42px minmax(0, 1fr);
  gap: 17px;
}
.sms-flow-nodes > li:not(:last-child)::after {
  position: absolute;
  content: '';
  top: 47px;
  bottom: -23px;
  left: 20px;
  width: 1px;
  background: var(--line);
}
.sms-flow-nodes > li:not(:first-child)::before {
  position: absolute;
  content: '';
  top: -13px;
  left: 17px;
  width: 7px;
  height: 7px;
  border-right: 1px solid var(--accent);
  border-bottom: 1px solid var(--accent);
  transform: rotate(45deg);
}
.sms-node-icon {
  display: grid;
  place-items: center;
  width: 42px;
  height: 42px;
  border: 1px solid var(--line);
  border-radius: 7px;
  background: var(--surface);
  color: var(--accent);
}
.sms-node-icon.amber {
  color: var(--amber);
}
.sms-node-icon.blue {
  color: var(--blue);
}
.sms-node-number {
  float: right;
  margin-left: 10px;
  color: var(--subtle);
  font:
    10px/22px ui-monospace,
    monospace;
}
.sms-flow-nodes h3 {
  font-size: 14px;
  font-weight: 600;
  line-height: 1.6;
}
.sms-flow-nodes p {
  font-size: 11px;
  line-height: 1.8;
  color: var(--muted);
  margin-top: 5px;
}
.sms-code-placeholder {
  display: flex;
  gap: 7px;
  margin-top: 11px;
}
.sms-code-placeholder > span {
  display: grid;
  place-items: center;
  width: 28px;
  height: 32px;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--surface);
}
.sms-code-placeholder i {
  width: 7px;
  height: 2px;
  background: var(--blue);
  opacity: 0.7;
}
@media (max-width: 850px) {
  .sms-flow {
    padding: 23px;
  }
  .sms-flow-nodes > li {
    gap: 12px;
  }
  .sms-code-placeholder {
    gap: 5px;
  }
  .sms-code-placeholder > span {
    width: 23px;
    height: 29px;
  }
}
@media (max-width: 600px) {
  .sms-flow {
    padding: 24px;
  }
  .sms-flow-nodes > li {
    gap: 15px;
  }
}
@media (max-width: 360px) {
  .sms-flow {
    padding: 20px;
  }
  .sms-flow-nodes > li {
    gap: 12px;
    grid-template-columns: 36px minmax(0, 1fr);
  }
  .sms-node-icon {
    width: 36px;
    height: 36px;
  }
  .sms-flow-nodes > li:not(:last-child)::after {
    left: 17px;
    top: 41px;
  }
  .sms-flow-nodes > li:not(:first-child)::before {
    left: 14px;
  }
  .sms-flow-nodes h3 {
    font-size: 12px;
  }
  .sms-code-placeholder > span {
    width: 22px;
    height: 28px;
  }
}
</style>
