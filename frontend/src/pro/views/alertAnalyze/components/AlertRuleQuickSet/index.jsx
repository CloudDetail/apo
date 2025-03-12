import React, { useEffect, useState } from 'react'
import { Modal, Form, Input, Select, Card, Checkbox, Flex } from 'antd'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { getMetricsApi, addMutationRule } from 'src/pro/api/alertAnalyze'
import { useSelector } from 'react-redux'
import { getStep } from 'src/core/utils/step'
import { showToast } from 'src/core/utils/toast'
import { useTranslation } from 'react-i18next'
const { TextArea } = Input
const { Option } = Select
const i18nKey = 'alertRuleQuickSet'
const AlertRuleQuickSet = ({ modalVisibility, closeModal }) => {
  const { t } = useTranslation('pro/alertAnalyze')
  //获取时间信息
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const step = getStep(startTime, endTime)

  //保存group信息
  const [group, setGroup] = useState('')
  const [groupUpper, setGroupUpper] = useState('')
  const [groupLower, setGroupLower] = useState('')

  //获取表单实例
  const [form] = Form.useForm()

  //告警等待时间单位
  const [unit, setUnit] = useState('s')

  //指标列表
  const [metricsList, setMetricsList] = useState([])

  //metrics数据处理
  const handleMetricsList = (metrics) => {
    const newMetricsList = []

    metrics.forEach((metric) => {
      // 查找是否存在该 group
      const existingGroup = newMetricsList.find((group) => group.title === metric.group)

      if (existingGroup) {
        // 如果存在，直接添加到 options 中
        existingGroup.options.push({
          label: (
            <span>
              {metric.name} <span style={{ color: '#fc6f03' }}>{metric.describe}</span>
            </span>
          ),
          value: `${metric.name}+${metric.describe}+${metric.customMetric}+${metric.group}`,
        })
      } else {
        // 如果不存在，创建新的 group
        newMetricsList.push({
          label: <span>{metric.group}</span>,
          title: metric.group,
          options: [
            {
              label: (
                <span>
                  {metric.name} <span style={{ color: '#fc6f03' }}>{metric.describe}</span>
                </span>
              ),
              value: `${metric.name}+${metric.describe}+${metric.customMetric}+${metric.group}`,
            },
          ],
        })
      }
    })

    return newMetricsList
  }

  //点击保存按钮处理逻辑
  const handleOk = async () => {
    const uploadData = await form.validateFields()
    uploadData.for = uploadData.for + unit
    uploadData.mutationCheckPQL = {
      customMetric: uploadData.mutationCheckPQL,
      group: group,
    }
    uploadData.mutationLowerLimit = uploadData.mutationLowerLimit
      ? {
          customMetric: uploadData.mutationLowerLimit,
          group: groupLower,
        }
      : undefined
    uploadData.mutationUpperLimit = uploadData.mutationUpperLimit
      ? {
          customMetric: uploadData.mutationUpperLimit,
          group: groupUpper,
        }
      : undefined
    uploadData.startTime = startTime
    uploadData.endTime = endTime
    uploadData.step = step
    const result = await addMutationRule(uploadData)
    if (result == 'ok') {
      showToast({
        title: t('successAddConfig'),
        color: 'success',
      })
      closeModal()
    }
  }

  //获取快速指标列表
  const getMetrics = async () => {
    try {
      let { metrics } = await getMetricsApi()
      metrics = handleMetricsList(metrics)
      setMetricsList(metrics)
    } catch (err) {
      console.error('Error fetching metrics:', err)
    }
  }

  //处理指标选择改变
  const handleChange = (value) => {
    console.log(value)
    const customMetric = value.split('+')[2]
    setGroup(value.split('+')[3])
    form.setFieldValue('mutationCheckPQL', customMetric)
  }

  //处理上界指标改变
  const handleChangeUpper = (value) => {
    const customMetric = value.split('+')[2]
    setGroupUpper(value.split('+')[3])
    form.setFieldValue('mutationUpperLimit', customMetric)
  }

  //处理下界指标改变
  const handleChangeLower = (value) => {
    const customMetric = value.split('+')[2]
    setGroupLower(value.split('+')[3])
    form.setFieldValue('mutationLowerLimit', customMetric)
  }

  //时间单位选择
  const selectAfter = (
    // <Form.Item
    //     name={"unit"}
    //     style={{
    //         height:"7px"
    //     }}
    // >
    <Select style={{ width: '120px' }} value={unit} onChange={(value) => setUnit(value)}>
      <Option value="s">{t(`${i18nKey}.timeSeconds`)}</Option>
      <Option value="m">{t(`${i18nKey}.timeMinutes`)}</Option>
      <Option value="h">{t(`${i18nKey}.timeHours`)}</Option>
      <Option value="d">{t(`${i18nKey}.timeDays`)}</Option>
    </Select>
    // </Form.Item>
  )
  //初始化
  useEffect(() => {
    form.resetFields()
    getMetrics()
  }, [modalVisibility])

  return (
    <Modal
      open={modalVisibility}
      onCancel={closeModal}
      maskClosable={false}
      onOk={handleOk}
      width={800}
      okText={t(`${i18nKey}.save`)}
      cancelText={t(`${i18nKey}.cancel`)}
      title={t(`${i18nKey}.alertConfig`)}
      centered
      destroyOnClose
    >
      <div className="modal-container">
        <Form form={form} className="form-container">
          <div className="form-header">
            <div className="form-left">
              <Form.Item
                label={t(`${i18nKey}.mutationName`)}
                name="name"
                rules={[
                  {
                    required: true,
                    message: t(`${i18nKey}.enterMutationName`),
                  },
                ]}
              >
                <Input />
              </Form.Item>
            </div>
            <Form.Item
              label={t(`${i18nKey}.alertTime`)}
              name="for"
              rules={[
                {
                  required: true,
                  message: t(`${i18nKey}.enterAlertTime`),
                },
              ]}
              className="form-right"
            >
              <Input addonAfter={selectAfter} />
            </Form.Item>
          </div>
          <Card title={t(`${i18nKey}.detectionStatement`)} className="detection-card">
            <Form.Item
              layout="vertical"
              label={
                <p className="form-label">
                  <span className="required">*</span> {t(`${i18nKey}.alertMetric`)}
                </p>
              }
            >
              <Select
                maxCount={1}
                placeholder={t(`${i18nKey}.selectMetric`)}
                options={metricsList}
                style={{ width: '100%', marginBottom: '20px' }}
                onSelect={handleChange}
              />
            </Form.Item>
            <Form.Item
              name="mutationCheckPQL"
              layout="vertical"
              className="pt-4 pb-5"
              rules={[
                {
                  required: true,
                  message: t(`${i18nKey}.selectQuickMetric`),
                },
              ]}
            >
              <TextArea />
            </Form.Item>
            <Form.Item
              layout="vertical"
              label={
                <p className="form-label">
                  {t(`${i18nKey}.upperThreshold`)}
                  <span className="text-xs text-gray-400">
                    {' '}
                    {t(`${i18nKey}.upperThresholdHint`)}
                  </span>
                </p>
              }
            >
              <Select
                maxCount={1}
                placeholder={t(`${i18nKey}.selectMetric`)}
                options={metricsList}
                style={{ width: '100%', marginBottom: '20px' }}
                onSelect={handleChangeUpper}
              />
            </Form.Item>
            <Form.Item name="mutationUpperLimit" layout="vertical" className="pt-4 pb-5">
              <TextArea />
            </Form.Item>
            <Form.Item
              layout="vertical"
              label={
                <p className="form-label">
                  {t(`${i18nKey}.lowerThreshold`)}
                  <span className="text-xs text-gray-400">
                    {' '}
                    {t(`${i18nKey}.lowerThresholdHint`)}
                  </span>
                </p>
              }
            >
              <Select
                maxCount={1}
                placeholder={t(`${i18nKey}.selectMetric`)}
                options={metricsList}
                style={{ width: '100%', marginBottom: '20px' }}
                onSelect={handleChangeLower}
              />
            </Form.Item>
            <Form.Item name="mutationLowerLimit" layout="vertical" className="pt-4 pb-5">
              <TextArea />
            </Form.Item>
          </Card>
          <Form.Item
            name="synchronizeToAlertRules"
            valuePropName="checked"
            style={{ marginTop: '10px' }}
          >
            <Checkbox>{t(`${i18nKey}.syncAlertRules`)}</Checkbox>
          </Form.Item>
        </Form>
      </div>
    </Modal>
  )
}

export default AlertRuleQuickSet
