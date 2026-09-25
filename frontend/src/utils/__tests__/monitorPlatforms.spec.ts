import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import type { UserMonitorView } from '@/api/channelMonitor'
import { groupMonitorPlatforms } from '@/utils/monitorPlatforms'
import MonitorStatusBoard from '@/components/user/monitor/MonitorStatusBoard.vue'

const item = (id: number, provider: string, status = 'operational') => ({
  id, provider, primary_status: status, name: `group-${id}`, group_name: `group-${id}`,
  primary_model: 'test', primary_latency_ms: 100, primary_ping_latency_ms: null,
  availability_7d: 99, extra_models: [], timeline: []
}) as UserMonitorView

describe('按平台展示监控', () => {
  it('保留故障分组和平台内原排序，隐藏空平台，保留未知平台', () => {
    const sections = groupMonitorPlatforms([item(5, 'anthropic'), item(3, 'deepseek', 'failed'), item(2, 'openai'), item(1, 'deepseek'), item(4, 'future')])
    expect(sections.map(s => s.value)).toEqual(['openai', 'anthropic', 'deepseek', 'future'])
    expect(sections[2]?.items.map(i => i.id)).toEqual([3, 1])
    expect(sections[2]?.items[0]?.primary_status).toBe('failed')
    expect(groupMonitorPlatforms([])).toEqual([])
  })

  it('平台切换只显示对应分组，切换视图和刷新时保留选择，平台消失时回退', async () => {
    const wrapper = mount(MonitorStatusBoard, {
      props: { items: [item(1, 'openai'), item(2, 'deepseek', 'failed')], window: '7d', countdownSeconds: 180,
        intervalSeconds: 180, loading: false, detailCache: {}, isAdmin: true, probeCache: {}, probeLoading: {} },
      global: { plugins: [createI18n({ legacy: false, locale: 'zh', messages: { zh: {} } })], stubs: { ProviderIcon: true, Icon: true } }
    })
    expect(wrapper.findAll('.status-board__platforms button')).toHaveLength(2)
    expect(wrapper.findAll('.platform-section')).toHaveLength(1)
    expect(wrapper.get('.platform-section').text()).toContain('group-1')
    expect(wrapper.get('.platform-section').text()).not.toContain('group-2')
    await wrapper.findAll('.status-board__platforms button')[1]!.trigger('click')
    expect(wrapper.get('.platform-section').text()).toContain('group-2')
    expect(wrapper.get('.platform-section').text()).not.toContain('group-1')
    expect(wrapper.text()).toContain('3 min')
    await wrapper.findAll('button').find(b => b.text() === '探针状态')!.trigger('click')
    expect(wrapper.findAll('.probe-group')).toHaveLength(1)
    expect(wrapper.get('.probe-group').text()).toContain('group-2')
    await wrapper.setProps({ items: [item(1, 'openai'), item(2, 'deepseek', 'failed')] })
    expect(wrapper.get('.probe-group').text()).toContain('group-2')
    await wrapper.findAll('button').find(b => b.text() === '分组状态')!.trigger('click')
    expect(wrapper.get('.platform-section').text()).toContain('group-2')
    await wrapper.setProps({ items: [item(1, 'openai')] })
    expect(wrapper.get('.platform-section').text()).toContain('group-1')
    await wrapper.setProps({ items: [] })
    expect(wrapper.find('.status-board__platforms').exists()).toBe(false)
    expect(wrapper.text()).toContain('暂无可展示的分组')
  })
})
