import { Button, Checkbox, Drawer, Form, Input, Modal, Select, Tooltip } from 'antd'
import {
  ArrowDownOutlined,
  ArrowUpOutlined,
  DeleteOutlined,
  PlusOutlined,
  SaveOutlined,
} from '@ant-design/icons'
import { diffWorkflowUpdate, toWorkflowEditValues } from '../client/workCapabilityApi.ts'
import type { WorkflowEditValues, WorkflowStep, WorkflowUpdateInput, WorkCapability } from '../client/types.ts'
import styles from './index.module.scss'

interface CapabilityEditDrawerProps {
  open: boolean
  capability: WorkCapability
  saving: boolean
  onClose: () => void
  onSave: (input: WorkflowUpdateInput) => Promise<void>
}

type EditValues = WorkflowEditValues

export default function CapabilityEditDrawer({
  open,
  capability,
  saving,
  onClose,
  onSave,
}: CapabilityEditDrawerProps) {
  const [form] = Form.useForm<EditValues>()
  const initialValues = toWorkflowEditValues(capability)

  const requestClose = () => {
    if (saving) return
    const update = diffWorkflowUpdate(initialValues, form.getFieldsValue(true))
    if (!Object.keys(update).length) {
      onClose()
      return
    }
    Modal.confirm({
      title: '放弃未保存的修改？',
      content: '任务名称、目标、提示、输入项和步骤的修改将不会保留。',
      okText: '放弃修改',
      cancelText: '继续编辑',
      okButtonProps: { danger: true },
      onOk: onClose,
    })
  }

  const submit = async (values: EditValues) => {
    const update = diffWorkflowUpdate(initialValues, values)
    if (!Object.keys(update).length) {
      onClose()
      return
    }
    try {
      await onSave(update)
    } catch {
      // The detail page keeps the drawer open and reports the server error.
    }
  }

  return (
    <Drawer
      open={open}
      width={680}
      title="编辑任务"
      destroyOnClose
      maskClosable={!saving}
      closable={!saving}
      onClose={requestClose}
      footer={(
        <div className={styles.drawerFooter}>
          <Button disabled={saving} onClick={requestClose}>取消</Button>
          <Button
            type="primary"
            icon={<SaveOutlined />}
            loading={saving}
            onClick={() => form.submit()}
          >
            保存修改
          </Button>
        </div>
      )}
    >
      <Form
        form={form}
        className={styles.editForm}
        layout="vertical"
        disabled={saving}
        initialValues={initialValues}
        clearOnDestroy
        requiredMark={false}
        onFinish={(values) => void submit(values)}
      >
        <Form.Item
          name="name"
          label="任务名称"
          rules={[
            { required: true, whitespace: true, message: '请输入任务名称' },
            { min: 2, max: 120, message: '任务名称需要 2 至 120 个字符' },
          ]}
        >
          <Input maxLength={120} showCount placeholder="输入便于识别的任务名称" />
        </Form.Item>

        <Form.Item
          name="goal"
          label="任务目标"
          rules={[
            { required: true, whitespace: true, message: '请输入任务目标' },
            { min: 2, max: 800, message: '任务目标需要 2 至 800 个字符' },
          ]}
        >
          <Input.TextArea
            autoSize={{ minRows: 3, maxRows: 8 }}
            maxLength={800}
            showCount
            placeholder="描述该任务要解决的问题和预期结果"
          />
        </Form.Item>

        <Form.Item
          name="examples"
          label="典型提问"
          rules={[{
            validator: async (_rule, values: unknown) => {
              const examples = Array.isArray(values) ? values.map((item) => String(item).trim()) : []
              if (examples.length > 12) throw new Error('典型提问最多 12 项')
              if (examples.some((item) => !item || item.length > 500)) {
                throw new Error('每项典型提问需要 1 至 500 个字符')
              }
            },
          }]}
        >
          <Select
            mode="tags"
            maxCount={12}
            maxTagCount="responsive"
            tokenSeparators={['，']}
            placeholder="输入可触发该任务的典型提问后按回车"
          />
        </Form.Item>

        <Form.List name="inputs">
          {(fields, { add, remove }) => (
            <div className={styles.editStepList}>
              <div className={styles.editSectionHeader}>
                <span>输入提示</span>
                <Button
                  size="small"
                  icon={<PlusOutlined />}
                  disabled={fields.length >= 12}
                  onClick={() => add({ name: '', required: false, source: 'conversation_or_prompt' })}
                >
                  添加输入
                </Button>
              </div>
              {fields.map(({ key, ...field }, index) => (
                <div className={styles.editInputItem} key={key}>
                  <Form.Item
                    {...field}
                    className={styles.editStepField}
                    name={[field.name, 'name']}
                    rules={[
                      { required: true, whitespace: true, message: '请输入提示名称' },
                      { max: 100, message: '提示名称最多 100 个字符' },
                    ]}
                  >
                    <Input maxLength={100} placeholder={`输入项 ${index + 1}，例如：服务名称`} />
                  </Form.Item>
                  <Form.Item name={[field.name, 'required']} valuePropName="checked" className={styles.editInputRequired}>
                    <Checkbox>必填</Checkbox>
                  </Form.Item>
                  <Form.Item name={[field.name, 'source']} hidden><PreservedValue /></Form.Item>
                  <Tooltip title="删除输入项">
                    <Button
                      type="text"
                      danger
                      aria-label={`删除输入项 ${index + 1}`}
                      icon={<DeleteOutlined />}
                      onClick={() => remove(field.name)}
                    />
                  </Tooltip>
                </div>
              ))}
            </div>
          )}
        </Form.List>

        <Form.List
          name="steps"
          rules={[{
            validator: async (_rule, steps: unknown[]) => {
              if (!Array.isArray(steps) || steps.length < 2 || steps.length > 6) {
                throw new Error('稳定步骤需要保留 2 至 6 项')
              }
            },
          }]}
        >
          {(fields, { add, remove, move }, { errors }) => (
            <div className={styles.editStepList}>
              <div className={styles.editSectionHeader}>
                <span>稳定步骤</span>
                <Button
                  size="small"
                  icon={<PlusOutlined />}
                  disabled={fields.length >= 6}
                  onClick={() => add(newWorkflowStep(form.getFieldValue('steps') || []))}
                >
                  添加步骤
                </Button>
              </div>

              {fields.map(({ key, ...field }, index) => (
                <div className={styles.editStepItem} key={key}>
                  <span className={styles.stepNumber}>{index + 1}</span>
                  <div className={styles.editStepFields}>
                    <Form.Item
                      {...field}
                      className={styles.editStepField}
                      name={[field.name, 'purpose']}
                      rules={[
                        { required: true, whitespace: true, message: '请输入步骤内容' },
                        { max: 500, message: '步骤内容最多 500 个字符' },
                      ]}
                    >
                      <Input.TextArea
                        autoSize={{ minRows: 1, maxRows: 4 }}
                        maxLength={500}
                        placeholder={`步骤 ${index + 1}`}
                      />
                    </Form.Item>
                    <Form.Item
                      className={styles.editStepField}
                      name={[field.name, 'expectedOutput']}
                      rules={[
                        { required: true, whitespace: true, message: '请输入预期输出' },
                        { max: 500, message: '预期输出最多 500 个字符' },
                      ]}
                    >
                      <Input.TextArea
                        autoSize={{ minRows: 1, maxRows: 3 }}
                        maxLength={500}
                        placeholder="该步骤应产生的结果"
                      />
                    </Form.Item>
                    <Form.Item name={[field.name, 'id']} hidden><PreservedValue /></Form.Item>
                    <Form.Item name={[field.name, 'providerRef']} hidden><PreservedValue /></Form.Item>
                    <Form.Item name={[field.name, 'capabilityRef']} hidden><PreservedValue /></Form.Item>
                    <Form.Item name={[field.name, 'inputMap']} hidden><PreservedValue /></Form.Item>
                    <Form.Item name={[field.name, 'effects']} hidden><PreservedValue /></Form.Item>
                  </div>
                  <div className={styles.editStepActions}>
                    <Tooltip title="上移">
                      <Button
                        type="text"
                        size="small"
                        aria-label={`上移步骤 ${index + 1}`}
                        icon={<ArrowUpOutlined />}
                        disabled={index === 0}
                        onClick={() => move(index, index - 1)}
                      />
                    </Tooltip>
                    <Tooltip title="下移">
                      <Button
                        type="text"
                        size="small"
                        aria-label={`下移步骤 ${index + 1}`}
                        icon={<ArrowDownOutlined />}
                        disabled={index === fields.length - 1}
                        onClick={() => move(index, index + 1)}
                      />
                    </Tooltip>
                    <Tooltip title={fields.length <= 2 ? '至少保留两个步骤' : '删除步骤'}>
                      <Button
                        type="text"
                        size="small"
                        danger
                        aria-label={`删除步骤 ${index + 1}`}
                        icon={<DeleteOutlined />}
                        disabled={fields.length <= 2}
                        onClick={() => remove(field.name)}
                      />
                    </Tooltip>
                  </div>
                </div>
              ))}
              <Form.ErrorList errors={errors} />
            </div>
          )}
        </Form.List>
      </Form>
    </Drawer>
  )
}

function PreservedValue() {
  return null
}

function newWorkflowStep(steps: WorkflowStep[]): WorkflowStep {
  const existing = new Set((steps || []).map((step) => String(step?.id || '')))
  const entropy = typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
    ? crypto.randomUUID().replace(/-/g, '')
    : `${Date.now()}_${steps.length + 1}`
  let id = `apo_step_${entropy}`.slice(0, 100)
  let suffix = 2
  while (existing.has(id)) {
    id = `${id.slice(0, 96)}_${suffix}`
    suffix += 1
  }
  return {
    id,
    purpose: '',
    providerRef: null,
    capabilityRef: null,
    inputMap: {},
    expectedOutput: '',
    effects: [],
  }
}
