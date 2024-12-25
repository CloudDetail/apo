import { Checkbox, Flex, Form, Input, InputNumber, Modal, Radio, Select, Tag, Tooltip } from 'antd'
import TextArea from 'antd/es/input/TextArea'
import _ from 'lodash'
import React, { useEffect, useState } from 'react'
import { IoIosRemoveCircleOutline, IoMdAddCircleOutline } from 'react-icons/io'
import { useSelector } from 'react-redux'
import { addRuleApi, updateRuleApi } from 'core/api/alerts'
import { showToast } from 'src/core/utils/toast'
import ALertConditionCom from './ALertConditionCom'
import { useTranslation } from 'react-i18next' // 引入i18n
import TranslationCom from 'src/oss/components/TranslationCom'

function isValidKeyValue(str) {
  // 定义正则表达式，key 以字母或下划线开头，后面可以跟字母、数字或下划线
  const regex = /^[a-zA-Z_][a-zA-Z0-9_]*=.*$/

  return regex.test(str)
}
function isValidKey(str) {
  // 定义正则表达式，确保开头是字母或下划线
  const regex = /^[a-zA-Z_].*$/

  return regex.test(str)
}
function splitTimeUnit(timeStr) {
  const regex = /^(\d+)([a-zA-Z]+)$/ // 正则表达式，用来匹配数字和字母
  const match = timeStr.match(regex)

  if (match) {
    const time = match[1] // 数字部分
    const unit = match[2] // 单位部分
    return { time, unit }
  } else {
    return null
    // throw new Error('Invalid time format')
  }
}
export default function ModifyAlertRuleModal({
  modalVisible,
  ruleInfo = null,
  closeModal,
  refresh,
}) {
  const [form] = Form.useForm()
  const [expr, setExpr] = useState(null)
  const [forUnit, setForUnit] = useState('s')
  const [keepUnit, setKeepUnit] = useState('s')
  const options = [
    {
      value: 's',
      label: <TranslationCom text="ModifyAlertRuleModal.option.secondText" url={'oss/alert'} />,
    },
    {
      value: 'm',
      label: <TranslationCom text="ModifyAlertRuleModal.option.minuteText" url={'oss/alert'} />,
    },
    {
      value: 'h',
      label: <TranslationCom text="ModifyAlertRuleModal.option.hourText" url={'oss/alert'} />,
    },
    {
      value: 'd',
      label: <TranslationCom text="ModifyAlertRuleModal.option.dayText" url={'oss/alert'} />,
    },
  ]
  const { groupLabelSelectOptions, groupLabel } = useSelector((state) => state.groupLabelReducer)
  const { t } = useTranslation('oss/alert') // 使用i18n
  const labelsTag = ({ value }) => {
    return isValidKeyValue(value) ? (
      <Tag closeIcon>{value}</Tag>
    ) : (
      <Tooltip title={t('ModifyAlertRuleModal.invalidLabelFormat')}>
        <Tag closeIcon color={'error'}>
          {value}
        </Tag>
      </Tooltip>
    )
  }
  const labelsValidator = (_, value) => {
    let pass = true
    for (const item of value ?? []) {
      if (!isValidKeyValue(item)) {
        pass = false
        break // 跳出循环
      }
    }
    if (!pass) {
      return Promise.reject(new Error(t('ModifyAlertRuleModal.invalidLabelFormat')))
    }
    return Promise.resolve()
  }
  // 表单格式转为需要存储的格式
  const formatterLabelsToMap = (labels = []) => {
    let result = {}
    labels.forEach((label) => {
      const [key, value] = label.split('=')
      result[key] = value
    })
    return result
  }
  // 转为表单需要的格式
  const formatterLabelsToList = (labels = {}) => {
    return Object.entries(labels).map(([key, value]) => key + '=' + value)
  }
  // 表单格式转为需要存储的格式
  const formatterAnnotations = (annotations) => {
    let result = {}
    annotations.forEach((label) => {
      if (label) {
        result[label.key] = label.value
      }
    })
    return result
  }
  // 转为表单需要的格式
  const formatterAnnotationsToFormState = (annotations) => {
    return Object.entries(annotations).reduce((acc, [key, value]) => {
      // 直接跳过 description 项
      if (key !== 'description') {
        acc.push({ key, value })
      }
      return acc
    }, [])
  }
  const changeGroupLabel = (key, value) => {
    const labelsList = _.cloneDeep(form.getFieldValue('labels') ?? [])
    const labelValue = key + '=' + value
    const index = labelsList.findIndex((item) => item.startsWith(key + '='))

    // 如果找到了匹配项，则进行替换
    if (index !== -1) {
      labelsList[index] = labelValue
    } else {
      labelsList.push(labelValue)
    }
    form.setFieldValue('labels', labelsList)
  }
  const updateRule = (alertRule) => {
    let api = addRuleApi
    let params = {
      alertRule,
    }
    if (ruleInfo) {
      api = updateRuleApi
      params.oldGroup = ruleInfo.group
      params.oldAlert = ruleInfo.alert
    }
    api(params).then(() => {
      showToast({
        title: t('ModifyAlertRuleModal.saveSuccess'),
        color: 'success',
      })
      closeModal()
      refresh()
    })
  }
  const saveRule = () => {
    form
      .validateFields({ validateOnly: true })
      .then(() => {
        const formState = form.getFieldsValue(true)
        const alertRule = {
          group: formState.group.label,
          alert: formState.alert,
          expr: expr,
        }
        if (formState.for) {
          alertRule.for = formState.for + forUnit
          // alertRule.keepFiringFor = null
        }
        if (formState.labels) {
          alertRule.labels = formatterLabelsToMap(formState.labels)
        }
        alertRule.annotations = formatterAnnotations(formState.annotations ?? [])

        if (formState.description) {
          alertRule.annotations['description'] = formState.description
        }
        updateRule(alertRule)
      })
      .catch((error) => console.log(error))
  }
  useEffect(() => {
    // console.log(ruleInfo)
    if (ruleInfo && modalVisible) {
      setExpr(ruleInfo.expr)
      const { time, unit } = splitTimeUnit(ruleInfo.for)
      if (unit) {
        setForUnit(unit)
      }
      const groupKey = Object.keys(groupLabel).find((key) => groupLabel[key] === ruleInfo.group)

      form.setFieldsValue({
        group: {
          key: groupKey,
          value: groupKey,
          label: ruleInfo.group,
        },
        alert: ruleInfo.alert,
        for: time,
        labels: formatterLabelsToList(ruleInfo.labels),
        annotations: formatterAnnotationsToFormState(ruleInfo.annotations ?? {}),
        description: ruleInfo.annotations?.description,
        severity: ruleInfo.labels?.severity,
      })
    } else {
      form.resetFields()
      setExpr()
      setForUnit('s')
    }
  }, [modalVisible, ruleInfo])
  const updateDescription = (e, value) => {
    let description = _.cloneDeep(form.getFieldValue('description') ?? '')
    const isExists = description.includes(value)
    if (e.target.checked) {
      if (!isExists) {
        // 如果不包含，则在最后添加该 value
        description = description ? description + '\n' + value : value
      }
    } else {
      // 检查有无 有则删去
      if (isExists) {
        // 如果包含，则删除该 value
        description = description
          .split('\n')
          .filter((line) => line.trim() !== value)
          .join('\n')
      }
    }

    // 更新表单的字段值
    form.setFieldsValue({
      description: description,
    })
  }
  return (
    <>
      <Modal
        title={t('ModifyAlertRuleModal.title')}
        open={modalVisible}
        onCancel={closeModal}
        destroyOnClose
        centered
        okText={t('ModifyAlertRuleModal.save')}
        cancelText={t('ModifyAlertRuleModal.cancel')}
        maskClosable={false}
        onOk={saveRule}
        width={1000}
        bodyStyle={{ maxHeight: '80vh', overflowY: 'auto', overflowX: 'hidden' }}
      >
        <Form
          layout={'vertical'}
          form={form}
          preserve={false}
          initialValues={{
            description: 'VALUE = {{ $value }}\nLABELS = {{ $labels }}',
          }}
        >
          <Form.Item
            label={t('ModifyAlertRuleModal.groupName')}
            name="group"
            required
            rules={[
              {
                validator: async (_, value) => {
                  if (!value)
                    return Promise.reject(
                      new Error(
                        t('ModifyAlertRuleModal.groupName') +
                          ' ' +
                          t('ModifyAlertRuleModal.invalidLabelFormat'),
                      ),
                    )
                },
              },
            ]}
          >
            <Select
              options={groupLabelSelectOptions}
              labelInValue
              placeholder={t('ModifyAlertRuleModal.groupName')}
              onChange={(value) => changeGroupLabel('group', value?.key)}
            />
          </Form.Item>
          <Form.Item label={t('ModifyAlertRuleModal.alertRuleName')} name="alert" required>
            <Input placeholder={t('ModifyAlertRuleModal.alertRuleName')} />
          </Form.Item>
          <Form.Item label={t('ModifyAlertRuleModal.alertCondition')} name="condition">
            <ALertConditionCom expr={expr} setExpr={setExpr} />
          </Form.Item>
          <Form.Item
            label={t('ModifyAlertRuleModal.alertLevel')}
            name="severity"
            layout="horizontal"
          >
            <Radio.Group onChange={(e) => changeGroupLabel('severity', e.target.value)}>
              <Radio value="critical">{t('ModifyAlertRuleModal.critical')}</Radio>
              <Radio value="warning">{t('ModifyAlertRuleModal.warning')}</Radio>
              <Radio value="info">{t('ModifyAlertRuleModal.info')}</Radio>
            </Radio.Group>
          </Form.Item>
          {/* <Form.Item label="查询语句" name="expr">
            <div className="border-1 border-solid rounded  border-[#424242] hover:border-[#3c89e8]  focus:border-[#3c89e8] ">
              <MonacoEditorWrapper defaultValue={expr} handleEditorChange={setExpr} />
            </div>
          </Form.Item> */}
          <Form.Item label={t('ModifyAlertRuleModal.duration')} name="for" required>
            <InputNumber
              addonAfter={
                <Select
                  value={forUnit}
                  style={{ width: 120 }}
                  options={options}
                  onChange={setForUnit}
                ></Select>
              }
              placeholder=""
            />
          </Form.Item>
          <Form.Item label={t('ModifyAlertRuleModal.alertMessage')} name="description">
            <TextArea placeholder={t('ModifyAlertRuleModal.alertMessage')} rows={2} />
          </Form.Item>
          {!ruleInfo && (
            <div className="flex mb-4">
              <Checkbox
                onChange={(e) => updateDescription(e, 'VALUE = {{ $value }}')}
                defaultChecked
              >
                {t('ModifyAlertRuleModal.includeMetricValue')}
              </Checkbox>
              <Checkbox
                onChange={(e) => updateDescription(e, 'LABELS = {{ $labels }}')}
                defaultChecked
              >
                {t('ModifyAlertRuleModal.includeMetricLabels')}
              </Checkbox>
            </div>
          )}
          <Form.Item
            label={t('ModifyAlertRuleModal.additionalLabels')}
            name="labels"
            rules={[{ validator: labelsValidator }]}
          >
            <Select
              mode="tags"
              style={{ width: '100%' }}
              tagRender={labelsTag}
              open={false}
              options={null}
              suffixIcon={null}
              placeholder={t('ModifyAlertRuleModal.additionalLabels')}
            />
          </Form.Item>
          {/* <Input placeholder="input placeholder" value={annotations} /> */}
          <Form.List name="annotations">
            {(fields, { add, remove }, { errors }) => (
              <>
                <Form.Item
                  label={
                    <>
                      {t('ModifyAlertRuleModal.additionalInfo')}{' '}
                      <IoMdAddCircleOutline onClick={() => add()} size={20} className="mx-2" />
                    </>
                  }
                >
                  {fields.map((field, index) => (
                    <Form.Item required={false} key={field.key} className="m-0">
                      <Flex gap="middle">
                        <Form.Item
                          layout="horizontal"
                          {...field}
                          validateTrigger={['onChange', 'onBlur']}
                          className="w-1/3"
                          name={[field.name, 'key']}
                          rules={[
                            {
                              validator: async (_, value) => {
                                if (!isValidKey(value)) {
                                  return Promise.reject(
                                    new Error(t('ModifyAlertRuleModal.invalidKey')),
                                  )
                                }
                              },
                            },
                          ]}
                        >
                          <Input placeholder={t('ModifyAlertRuleModal.key')} />
                        </Form.Item>
                        <Form.Item
                          // layout="horizontal"
                          className="w-2/3 block test"
                          name={[field.name, 'value']}
                        >
                          <TextArea placeholder={t('ModifyAlertRuleModal.value')} rows={2} />
                        </Form.Item>

                        <IoIosRemoveCircleOutline
                          size={20}
                          className="mt-1"
                          onClick={() => remove(field.name)}
                        />
                      </Flex>
                    </Form.Item>
                  ))}
                </Form.Item>
              </>
            )}
          </Form.List>
          {/* </Form.Item> */}
        </Form>
      </Modal>
    </>
  )
}
