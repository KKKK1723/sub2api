import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import MonitorOrderDialog from './MonitorOrderDialog.vue'

const { getOrder, updateOrder, showError, showSuccess } = vi.hoisted(() => ({
  getOrder: vi.fn(), updateOrder: vi.fn(), showError: vi.fn(), showSuccess: vi.fn(),
}))
vi.mock('@/api/admin/channelMonitor', () => ({ getOrder, updateOrder }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError, showSuccess }) }))
const initial = [
  { id: 7, name: 'monitor1', group_name: 'OpenAI官key1', enabled: true },
  { id: 3, name: 'Pro', group_name: '', enabled: false },
]
function mountDialog() {
  return mount(MonitorOrderDialog, {
    props: { show: true },
    global: { stubs: {
      BaseDialog: defineComponent({ template: '<div><slot /><slot name="footer" /></div>' }),
      VueDraggable: defineComponent({ template: '<div><slot /></div>' }),
    } },
  })
}
describe('MonitorOrderDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getOrder.mockResolvedValue(initial.map(item => ({ ...item })))
    updateOrder.mockResolvedValue(undefined)
  })
  it('keeps disabled groups and saves the moved IDs with the original snapshot', async () => {
    const wrapper = mountDialog()
    await flushPromises()
    expect(wrapper.text()).toContain('已停用')
    await wrapper.get('[aria-label="上移 Pro"]').trigger('click')
    expect(wrapper.findAll('[data-monitor-id]').map(row => row.attributes('data-monitor-id'))).toEqual(['3', '7'])
    await wrapper.get('.btn-primary').trigger('click')
    await flushPromises()
    expect(updateOrder).toHaveBeenCalledWith([3, 7], [7, 3])
    expect(wrapper.emitted('saved')).toHaveLength(1)
    expect(wrapper.emitted('close')).toHaveLength(1)
  })
  it('does not persist a cancelled edit', async () => {
    const wrapper = mountDialog()
    await flushPromises()
    await wrapper.get('[aria-label="上移 Pro"]').trigger('click')
    await wrapper.get('.btn-secondary').trigger('click')
    expect(updateOrder).not.toHaveBeenCalled()
    expect(wrapper.emitted('close')).toHaveLength(1)
  })
  it('keeps the editor open when saving fails', async () => {
    updateOrder.mockRejectedValue(new Error('conflict'))
    const wrapper = mountDialog()
    await flushPromises()
    await wrapper.get('[aria-label="上移 Pro"]').trigger('click')
    await wrapper.get('.btn-primary').trigger('click')
    await flushPromises()
    expect(showError).toHaveBeenCalled()
    expect(wrapper.emitted('saved')).toBeUndefined()
    expect(wrapper.emitted('close')).toBeUndefined()
  })
  it('disables saving when loading fails', async () => {
    getOrder.mockRejectedValue(new Error('offline'))
    const wrapper = mountDialog()
    await flushPromises()
    expect(wrapper.find('[role="alert"]').exists()).toBe(true)
    expect(wrapper.get('.btn-primary').attributes('disabled')).toBeDefined()
  })
})
