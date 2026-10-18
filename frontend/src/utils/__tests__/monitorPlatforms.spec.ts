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

  it('默认显示全部平台，筛选只显示对应分组且指标随之变化，切换视图和刷新时保留选择，平台消失时回退全部', async () => {
    const wrapper = mount(MonitorStatusBoard, {
      props: { items: [item(1, 'openai'), item(2, 'deepseek', 'failed')], window: '7d', countdownSeconds: 300,
        intervalSeconds: 300, loading: false, detailCache: {}, isAdmin: true, probeCache: {}, probeLoading: {} },
      global: { plugins: [createI18n({ legacy: false, locale: 'zh', messages: { zh: {} } })], stubs: { ProviderIcon: true, Icon: true } }
    })
    // 测试环境使用 vue-i18n 运行时版本，不编译文案，t() 直接返回 key
    const sectionsText = () => wrapper.findAll('.board-section').map(section => section.text()).join('\n')
    // 「全部平台」+ 两个平台
    expect(wrapper.findAll('.board-platforms button')).toHaveLength(3)
    expect(wrapper.findAll('.board-section')).toHaveLength(2)
    expect(wrapper.text()).toContain('channelStatus.board.overall.outage')
    expect(wrapper.text()).toContain('300s')
    expect(wrapper.text()).toContain('channelStatus.board.overallLabel')
    // 指标卡跟随平台筛选：只看 OpenAI 时整体正常，只看 DeepSeek 时整体中断
    await wrapper.findAll('.board-platforms button')[1]!.trigger('click')
    expect(wrapper.text()).toContain('channelStatus.board.overall.operational')
    expect(wrapper.text()).toContain('channelStatus.board.platformOverallLabel')
    await wrapper.findAll('.board-platforms button')[2]!.trigger('click')
    expect(wrapper.text()).toContain('channelStatus.board.overall.outage')
    expect(wrapper.findAll('.board-section')).toHaveLength(1)
    expect(sectionsText()).toContain('group-2')
    expect(sectionsText()).not.toContain('group-1')
    await wrapper.findAll('button').find(b => b.text() === 'channelStatus.board.views.probes')!.trigger('click')
    expect(wrapper.findAll('.board-probe-group')).toHaveLength(1)
    expect(wrapper.get('.board-probe-group').text()).toContain('group-2')
    await wrapper.setProps({ items: [item(1, 'openai'), item(2, 'deepseek', 'failed')] })
    expect(wrapper.get('.board-probe-group').text()).toContain('group-2')
    await wrapper.findAll('button').find(b => b.text() === 'channelStatus.board.views.groups')!.trigger('click')
    expect(sectionsText()).toContain('group-2')
    await wrapper.setProps({ items: [item(1, 'openai')] })
    expect(wrapper.findAll('.board-section')).toHaveLength(1)
    expect(sectionsText()).toContain('group-1')
    expect(wrapper.text()).toContain('channelStatus.board.overall.operational')
    await wrapper.setProps({ items: [] })
    expect(wrapper.find('.board-platforms').exists()).toBe(false)
    expect(wrapper.text()).toContain('channelStatus.board.empty')
  })
})
